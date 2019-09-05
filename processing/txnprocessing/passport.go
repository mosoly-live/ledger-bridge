package txnprocessing

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"

	"gitlab.com/p-invent/mosoly-ledger-bridge/models/dbmodels"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mosolyapi"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/monetha/go-verifiable-data/deployer"
	"github.com/monetha/go-verifiable-data/eth"
	"github.com/monetha/go-verifiable-data/eth/backend/ethclient"
	"github.com/monetha/go-verifiable-data/facts"
	"gitlab.com/p-invent/mosoly-ledger-bridge/config"
)

// FactProviderContext keeps session data
type FactProviderContext struct {
	context  context.Context
	address  common.Address
	reader   *facts.Reader
	provider *facts.Provider
}

const (
	// Key max. 32 bytes allowed
	factKeyProject  = "project"
	mentorKeySuffix = "_mentorees"

	projectSchemaURL   = "http://portal.mosoly.live/project.json"
	mentoreesSchemaURL = "http://portal.mosoly.live/mentorees.json"
	userSchemaURL      = "http://portal.mosoly.live/user.json"
)

func deployPassports(ctx context.Context, projects []*dbmodels.Project, factProviderSession *eth.Session) (map[int]common.Address, error) {
	passportFactoryAddress := common.HexToAddress(config.EthereumPassportFactoryAddress)

	projectsPassports := make(map[int]common.Address)

	for _, project := range projects {
		if project.PassportAddress != "" {
			continue
		}

		passportAddress, err := deployer.New(factProviderSession).
			DeployPassport(ctx, passportFactoryAddress)
		if err != nil {
			return projectsPassports, err
		}

		log.Println("syncToBlockchain: deployPassports - new address: ", passportAddress.String())

		projectsPassports[project.ID] = passportAddress

		// will be used then for writing facts
		project.PassportAddress = passportAddress.String()
	}

	return projectsPassports, nil
}

func (t *TxnProcessing) savePassportAddresses(passportAddresses map[int]common.Address) error {
	if len(passportAddresses) == 0 {
		return nil
	}

	db := t.db
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin user insert/update transaction: %v", err)
	}
	defer tx.Rollback()

	for projectID, passportAddress := range passportAddresses {
		_, err := tx.Exec(tx.Rebind(`
			UPDATE project_data SET
				passport_address = ?
			WHERE id = ?`), passportAddress.String(), projectID,
		)
		if err != nil {
			return fmt.Errorf("failed to update project passport address: %v", err)
		}

		log.Println("syncToBlockchain: savePassportAddresses - new address saved to db: ", passportAddress.String())
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit savePassportAddresses transaction: %v", err)
	}

	return nil
}

func (t *TxnProcessing) updateProjectsFacts(projects []*dbmodels.Project, providerContext FactProviderContext) error {
	factKeyProjectBytes, err := getFactKeyBytes(factKeyProject)
	if err != nil {
		return err
	}

	for _, project := range projects {
		passportAddress := common.HexToAddress(project.PassportAddress)

		// Write only updated fact
		projectFact := &mosolyapi.BlockchainProjectFact{}
		if err := readFact(factKeyProjectBytes, passportAddress, providerContext, projectFact); err != nil {
			log.Println(err)
		}

		factToWrite := getProjectFact(project, projectFact)
		if factToWrite == nil {
			continue
		}

		hash, err := writeFact(factKeyProjectBytes, passportAddress, providerContext, factToWrite)

		if err != nil {
			log.Println(err)
		}

		trxID, err := t.createTxnData(hash.String())
		if err != nil {
			log.Println(err)
			return err
		}

		err = t.updateProjectFactTransaction(project.ID, trxID)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func getUserFact(user *dbmodels.User, userFact *mosolyapi.BlockchainUserFact) *mosolyapi.BlockchainFact {
	mentors := make([]string, 0)

	for _, mentor := range user.Mentors {
		mentors = append(mentors, mentor.Account)
	}

	sort.Strings(mentors)

	newFact := mosolyapi.UserFact{
		InviteURLHash: user.InviteURLHash,
		Account:       user.Account,
		Validated:     user.Validated,
		Mentors:       mentors,
	}

	if reflect.DeepEqual(newFact, userFact.Payload) {
		return nil
	}

	return &mosolyapi.BlockchainFact{
		Schema:  userSchemaURL,
		Payload: newFact,
	}
}

func getMentorFact(user *dbmodels.User, mentorFact *mosolyapi.BlockchainMentorFact) *mosolyapi.BlockchainFact {
	var newFact = make(mosolyapi.MentorFact, 0)
	for _, mentoree := range user.Mentorees {
		newFact = append(newFact, mentoree.Account)
	}

	if len(newFact) == 0 && len(mentorFact.Payload) == 0 {
		return nil
	}

	sort.Strings(newFact)

	if reflect.DeepEqual(newFact, mentorFact.Payload) {
		return nil
	}

	return &mosolyapi.BlockchainFact{
		Schema:  mentoreesSchemaURL,
		Payload: newFact,
	}
}

func getProjectFact(project *dbmodels.Project, projectFact *mosolyapi.BlockchainProjectFact) *mosolyapi.BlockchainFact {
	fact := mosolyapi.ProjectFact{
		Name: project.Name,
	}

	if reflect.DeepEqual(fact, projectFact.Payload) {
		return nil
	}

	return &mosolyapi.BlockchainFact{
		Schema:  projectSchemaURL,
		Payload: fact,
	}
}

func getBytesFromHexAddress(hexAddr string) (factKeyBytes [32]byte, err error) {
	hexBytes := common.FromHex(hexAddr)
	copy(factKeyBytes[:], hexBytes)

	return factKeyBytes, nil
}

func getMentorFactKeyBytes(mentorAddress string) [32]byte {
	key := fmt.Sprintf("%s%s", mentorAddress, mentorKeySuffix)
	bytes := crypto.Keccak256([]byte(key))

	var factKeyBytes [32]byte
	copy(factKeyBytes[:], bytes)

	return factKeyBytes
}

func (t *TxnProcessing) updateMentorsFacts(users []*dbmodels.User, providerContext FactProviderContext) error {
	passportAddress := common.HexToAddress(config.AppMosolyDidAddress)

	for _, user := range users {
		if user.Mentorees == nil || len(user.Mentorees) == 0 {
			continue
		}

		factKeyBytes := getMentorFactKeyBytes(user.Account)

		// Write only updated fact
		mentorFact := &mosolyapi.BlockchainMentorFact{}
		if err := readFact(factKeyBytes, passportAddress, providerContext, mentorFact); err != nil {
			log.Println(err)
		}

		factToWrite := getMentorFact(user, mentorFact)
		if factToWrite == nil {
			continue
		}

		hash, err := writeFact(factKeyBytes, passportAddress, providerContext, factToWrite)

		if err != nil {
			log.Println(err)
		}

		trxID, err := t.createTxnData(hash.String())
		if err != nil {
			log.Println(err)
			return err
		}

		err = t.updateMentorFactTransaction(user.ID, trxID)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (t *TxnProcessing) updateUsersFacts(users []*dbmodels.User, providerContext FactProviderContext) error {
	passportAddress := common.HexToAddress(config.AppMosolyDidAddress)

	for _, user := range users {
		factKeyUserBytes, err := getBytesFromHexAddress(user.Account)
		if err != nil {
			log.Println(err)
			continue
		}

		// Write only updated fact
		userFact := &mosolyapi.BlockchainUserFact{}
		if err := readFact(factKeyUserBytes, passportAddress, providerContext, userFact); err != nil {
			log.Println(err)
		}

		factToWrite := getUserFact(user, userFact)
		if factToWrite == nil {
			continue
		}

		hash, err := writeFact(factKeyUserBytes, passportAddress, providerContext, factToWrite)

		if err != nil {
			log.Println(err)
		}

		trxID, err := t.createTxnData(hash.String())
		if err != nil {
			log.Println(err)
			return err
		}

		err = t.updateUserFactTransaction(user.ID, trxID)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (t *TxnProcessing) syncToBlockchain(ctx context.Context, projects []*dbmodels.Project, users []*dbmodels.User) error {
	var (
		backendURL         = config.EthereumJSONRPCURL
		factProviderKeyHex = config.AppMosolyOpsAccount

		factProviderKey     *ecdsa.PrivateKey
		factProviderAddress common.Address
		client              *ethclient.Client
		err                 error
	)

	if factProviderKey, err = crypto.HexToECDSA(factProviderKeyHex); err != nil {
		log.Println("syncToBlockchain: wrong fact provider key.", err)
		return err
	}

	if client, err = ethclient.Dial(backendURL); err != nil {
		log.Println("syncToBlockchain: could not create eth client.", err)
		return err
	}

	//e, err := cmdutils.NewEth(ctx, c.BackendURL, "", nil)
	e := eth.New(client, ethlog.Warn)

	// creating owner session and checking balance
	factProviderSession := e.NewSession(factProviderKey)

	newPassportAddresses, err := deployPassports(ctx, projects, factProviderSession)
	if err != nil {
		log.Println("syncToBlockchain: deployPassports error: ", err)
		return err
	}

	err = t.savePassportAddresses(newPassportAddresses)
	if err != nil {
		log.Println("syncToBlockchain: savePassportAddresses error: ", err)
		return err
	}

	factProvider := facts.NewProvider(factProviderSession)

	factReader := facts.NewReader(e)

	factProviderAddress = bind.NewKeyedTransactor(factProviderKey).From

	providerContext := FactProviderContext{
		address:  factProviderAddress,
		context:  ctx,
		provider: factProvider,
		reader:   factReader,
	}

	err = t.updateProjectsFacts(projects, providerContext)
	if err != nil {
		log.Println("syncToBlockchain: updateProjectsFacts error: ", err)
		return err
	}

	err = t.updateMentorsFacts(users, providerContext)
	if err != nil {
		log.Println("syncToBlockchain: updateMentorsFacts error: ", err)
		return err
	}

	err = t.updateUsersFacts(users, providerContext)
	if err != nil {
		log.Println("syncToBlockchain: updateUsersFacts error: ", err)
		return err
	}

	return nil
}

func getFactKeyBytes(factKeyStr string) (factKey [32]byte, err error) {
	if factKeyBytes := []byte(factKeyStr); len(factKeyBytes) <= 32 {
		copy(factKey[:], factKeyBytes)
		return
	}

	err = fmt.Errorf("getFactKeyBytes: the factKey '%s' string should fit into 32 bytes", factKeyStr)
	return
}

func readFact(factKey [32]byte, passportAddress common.Address, ctx FactProviderContext, factObject interface{}) error {
	var resultBytes []byte
	var err error

	if resultBytes, err = ctx.reader.ReadTxData(ctx.context, passportAddress, ctx.address, factKey); err != nil {
		if err != ethereum.NotFound {
			return fmt.Errorf("syncToBlockchain: ReadTxData failed: %s", err)
		}
	}

	if len(resultBytes) > 0 {
		if err := json.Unmarshal(resultBytes, factObject); err != nil {
			return fmt.Errorf("readFact: can't unmarshal tx data: %s", err)
		}
	}

	return nil
}

func writeFact(factKey [32]byte, passportAddress common.Address, ctx FactProviderContext, factObject interface{}) (*common.Hash, error) {
	factBytes, _ := json.Marshal(factObject)

	hash, err := ctx.provider.WriteTxData(ctx.context, passportAddress, factKey, factBytes)

	if err != nil {
		return nil, fmt.Errorf("writeFact: WriteTxData  failed: %s", err)
	}

	return &hash, nil
}
