package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/consul/api"
	distributed "github.com/monetha/go-distributed"
	"gitlab.com/p-invent/mosoly-ledger-bridge/log"
)

const (
	// debugMode holds the code for the application debug mode
	debugMode = "debug"
	// releaseMode holds the code for the application release mode
	releaseMode = "release"

	// EnvLive holds the live environment name
	EnvLive = "live"
	// EnvStaging holds the live environment name
	EnvStaging = "staging"
	// EnvLocal holds the dev environment name
	EnvLocal = "local"
)

var (
	// ServiceEnvironment is an environment (local, staging, live)
	ServiceEnvironment string
	// SQLConnectionString is a connection string for DB
	SQLConnectionString = ""
	// AppInDebugMode flag means whether application is started in debug mode ore not
	AppInDebugMode = true
	// AppRootPath virtual path that will be used as root for application
	AppRootPath = ""
	// HTTPPort is HTTP port of web application
	HTTPPort = 8087
	// ConsulConfig is the configuration of consul broker, nil when Consul shouldn't be used
	ConsulConfig *distributed.Config
	// ConsulKeyPrefix is a prefix of all Consul keys.
	ConsulKeyPrefix string
	// EthereumJSONRPCURL is Ethereum JSON RPC URL
	EthereumJSONRPCURL string
	// EthereumPassportFactoryAddress is a passport factory address
	// Mainnet: 0x53b21DC502b163Bcf3bD9a68d5db5e8E6110E1CC
	// Ropsten: 0x35Cb95Db8E6d56D1CF8D5877EB13e9EE74e457F2
	EthereumPassportFactoryAddress string
	// AppMosolyDidAddress is Ethereum main passport address
	AppMosolyDidAddress string
	// AppMosolyOpsAccount is Ethereum private key
	// for deploying passports to EthereumPassportFactoryAddress
	// and submitting facts to AppMosolyDidAddress
	AppMosolyOpsAccount string
	// AppMosolyBackendURL is mosoly backend API URL.
	AppMosolyBackendURL string
	// AppMosolyBackendToken is Bearer authorizarion token for Mosoly API
	AppMosolyBackendToken string
)

// Parse parses application configuration from command line and environment variables
func Parse() {
	const (
		serviceEnvironmentCmdLnName = "service.env"
		serviceEnvironmentEnvName   = "SERVICE_ENV"
		serviceEnvironmentDefault   = ""

		appRootPathCmdLnName = "app.rootpath"
		appRootPathEnvName   = "APP_ROOT_PATH"
		appRootPathDefault   = "/plb/v1"

		appModeCmdLnName = "app.mode"
		appModeEnvName   = "APP_MODE"
		appModeDefault   = debugMode

		dbUserCmdLnName = "db.user"
		dbUserEnvName   = "DB_USER"
		dbUserDefault   = "mosoly"

		dbPassCmdLnName = "db.pass"
		dbPassEnvName   = "DB_PASS"
		dbPassDefault   = "qwerty123456"

		dbHostCmdLnName = "db.host"
		dbHostEnvName   = "DB_HOST"
		dbHostDefault   = "localhost"

		dbPortCmdLnName = "db.port"
		dbPortEnvName   = "DB_PORT"
		dbPortDefault   = "45433"

		dbNameCmdLnName = "db.name"
		dbNameEnvName   = "DB_NAME"
		dbNameDefault   = "ledger_bridge_cache"

		dbConnectTimeoutCmdLnName = "db.timeout"
		dbConnectTimeoutEnvName   = "DB_TIMEOUT"
		dbConnectTimeoutDefault   = "5"

		dbReadTimeoutCmdLnName = "db.read.timeout"
		dbReadTimeoutEnvName   = "DB_READ_TIMEOUT"
		dbReadTimeoutDefault   = "300000" // 5 mins = 300k milliseconds

		dbWriteTimeoutCmdLnName = "db.write.timeout"
		dbWriteTimeoutEnvName   = "DB_WRITE_TIMEOUT"
		dbWriteTimeoutDefault   = "300000" // 5 mins = 300k milliseconds

		httpPortCmdLnName = "http.port"
		httpPortEnvName   = "HTTP_PORT"
		httpPortDefault   = 8087

		noConsulCmdLnName = "noconsul"
		noConsulEnvName   = "NO_CONSUL"
		noConsulDefault   = false

		consulHTTPAddrCmdLnName = "consul.http.addr"
		consulHTTPAddrEnvName   = api.HTTPAddrEnvName
		consulHTTPAddrDefault   = ""

		consulHTTPTokenCmdLnName = "consul.http.token"
		consulHTTPTokenEnvName   = api.HTTPTokenEnvName
		consulHTTPTokenDefault   = ""

		consulKeyPrefixCmdLnName = "consul.key.prefix"
		consulKeyPrefixEnvName   = "CONSUL_KEY_PREFIX"
		consulKeyPrefixDefault   = "mosoly"

		ethereumJSONRPCURLCmdLnName = "ethereum.json.rpc.url"
		ethereumJSONRPCURLEnvName   = "ETHEREUM_JSON_RPC_URL"
		ethereumJSONRPCURLDefault   = ""

		appMosolyOpsAccountCmdLnName = "app.mosoly.ops.account"
		appMosolyOpsAccountEnvName   = "APP_MOSOLY_OPS_ACCOUNT"
		appMosolyOpsAccountDefault   = ""

		ethereumPassportFactoryAddressCmdLnName = "ethereum.passport.factory.address"
		ethereumPassportFactoryAddressEnvName   = "ETHEREUM_PASSPORT_FACTORY_ADDRESS"
		ethereumPassportFactoryAddressDefault   = ""

		appMosolyDidAddressCmdLnName = "app.mosoly.did.address"
		appMosolyDidAddressEnvName   = "APP_MOSOLY_DID_ADDRESS"
		appMosolyDidAddressDefault   = ""

		appMosolyBackendURLCmdLnName = "app.mosoly.backend.url"
		appMosolyBackendURLEnvName   = "APP_MOSOLY_BACKEND_URL"
		appMosolyBackendURLDefault   = ""

		appMosolyBackendTokenCmdLnName = "app.mosoly.backend.token"
		appMosolyBackendTokenEnvName   = "APP_MOSOLY_BACKEND_TOKEN"
		appMosolyBackendTokenDefault   = ""
	)

	flag.StringVar(&ServiceEnvironment, serviceEnvironmentCmdLnName, getEnv(serviceEnvironmentEnvName, serviceEnvironmentDefault),
		"The service environment (can be overridden with the "+serviceEnvironmentEnvName+" environment variable")

	flag.StringVar(&AppRootPath, appRootPathCmdLnName, getEnv(appRootPathEnvName, appRootPathDefault),
		"The application root virtual path (can be overridden with the "+appRootPathEnvName+" environment variable")

	var appMode string
	flag.StringVar(&appMode, appModeCmdLnName, getEnv(appModeEnvName, appModeDefault),
		"The application mode (can be overridden with the "+appModeEnvName+" environment variable")

	var dbUser string
	flag.StringVar(&dbUser, dbUserCmdLnName, getEnv(dbUserEnvName, dbUserDefault),
		"The DB username (can be overridden with the "+dbUserEnvName+" environment variable)")

	var dbPass string
	flag.StringVar(&dbPass, dbPassCmdLnName, getEnv(dbPassEnvName, dbPassDefault),
		"The DB password (can be overridden with the "+dbPassEnvName+" environment variable)")

	var dbHost string
	flag.StringVar(&dbHost, dbHostCmdLnName, getEnv(dbHostEnvName, dbHostDefault),
		"The DB host (can be overridden with the "+dbHostEnvName+" environment variable)")

	var dbPort string
	flag.StringVar(&dbPort, dbPortCmdLnName, getEnv(dbPortEnvName, dbPortDefault),
		"The DB port (can be overridden with the "+dbPortEnvName+" environment variable)")

	var dbName string
	flag.StringVar(&dbName, dbNameCmdLnName, getEnv(dbNameEnvName, dbNameDefault),
		"The DB name (can be overridden with the "+dbNameEnvName+" environment variable)")

	var dbConnectTimeout string
	flag.StringVar(&dbConnectTimeout, dbConnectTimeoutCmdLnName, getEnv(dbConnectTimeoutEnvName, dbConnectTimeoutDefault),
		"The DB connect timeout (can be overridden with the "+dbConnectTimeoutEnvName+" environment variable)")

	var dbReadTimeout string
	flag.StringVar(&dbReadTimeout, dbReadTimeoutCmdLnName, getEnv(dbReadTimeoutEnvName, dbReadTimeoutDefault),
		"The DB read timeout (can be overridden with the "+dbReadTimeoutEnvName+" environment variable)")

	var dbWriteTimeout string
	flag.StringVar(&dbWriteTimeout, dbWriteTimeoutCmdLnName, getEnv(dbWriteTimeoutEnvName, dbWriteTimeoutDefault),
		"The DB write timeout (can be overridden with the "+dbWriteTimeoutEnvName+" environment variable)")

	flag.IntVar(&HTTPPort, httpPortCmdLnName, getEnvInt(httpPortEnvName, httpPortDefault),
		"The HTTP server port (can be overridden with the "+httpPortEnvName+" environment variable)")

	var consulHTTPAddr string
	flag.StringVar(&consulHTTPAddr, consulHTTPAddrCmdLnName, getEnv(consulHTTPAddrEnvName, consulHTTPAddrDefault),
		"The Consul HTTP address (can be overridden with the "+consulHTTPAddrEnvName+" environment variable)")

	var consulHTTPToken string
	flag.StringVar(&consulHTTPToken, consulHTTPTokenCmdLnName, getEnv(consulHTTPTokenEnvName, consulHTTPTokenDefault),
		"The Consul HTTP token (can be overridden with the "+consulHTTPTokenEnvName+" environment variable)")

	flag.StringVar(&ConsulKeyPrefix, consulKeyPrefixCmdLnName, getEnv(consulKeyPrefixEnvName, consulKeyPrefixDefault),
		"The prefix of all Consul keys (can be overridden with the "+consulKeyPrefixEnvName+" environment variable)")

	var noConsul bool
	flag.BoolVar(&noConsul, noConsulCmdLnName, getEnvBool(noConsulEnvName, noConsulDefault),
		"Do not use Consul for distributed tasks, run them locally (can be overridden with the "+noConsulEnvName+" environment variable)")

	flag.StringVar(&EthereumJSONRPCURL, ethereumJSONRPCURLCmdLnName, getEnv(ethereumJSONRPCURLEnvName, ethereumJSONRPCURLDefault),
		"The Ethereum network JSON RPC URL (can be overridden with the "+ethereumJSONRPCURLEnvName+" environment variable)")

	flag.StringVar(&AppMosolyOpsAccount, appMosolyOpsAccountCmdLnName, getEnv(appMosolyOpsAccountEnvName, appMosolyOpsAccountDefault),
		"Ethereum passport fact provider key (can be overridden with the "+appMosolyOpsAccountEnvName+" environment variable)")

	flag.StringVar(&EthereumPassportFactoryAddress, ethereumPassportFactoryAddressCmdLnName, getEnv(ethereumPassportFactoryAddressEnvName, ethereumPassportFactoryAddressDefault),
		"Ethereum passport factory address (can be overridden with the "+ethereumPassportFactoryAddressEnvName+" environment variable)")

	flag.StringVar(&AppMosolyDidAddress, appMosolyDidAddressCmdLnName, getEnv(appMosolyDidAddressEnvName, appMosolyDidAddressDefault),
		"Ethereum main passport address (can be overridden with the "+appMosolyDidAddressEnvName+" environment variable)")

	flag.StringVar(&AppMosolyBackendURL, appMosolyBackendURLCmdLnName, getEnv(appMosolyBackendURLEnvName, appMosolyBackendURLDefault),
		"MTH API URL is mth-api URL (can be overridden with the "+appMosolyBackendURLEnvName+" environment variable")

	flag.StringVar(&AppMosolyBackendToken, appMosolyBackendTokenCmdLnName, getEnv(appMosolyBackendTokenEnvName, appMosolyBackendTokenDefault),
		"The Auth token secret key (can be overridden with the "+appMosolyBackendTokenEnvName+" environment variable)")

	flag.Parse()

	if len(ServiceEnvironment) == 0 {
		printUsageErrorAndExit("either add -" + serviceEnvironmentCmdLnName + " command line parameter or service environment with the " + serviceEnvironmentEnvName + " environment variable")
	}

	if appMode != debugMode && appMode != releaseMode {
		printUsageErrorAndExit("provide valid application mode using "+appModeEnvName+" environment variable, valid values are: %v, %v", debugMode, releaseMode)
	}

	if !noConsul {
		if consulHTTPAddr == "" {
			printUsageErrorAndExit("either add -" + noConsulCmdLnName + " command line parameter or specify Consul HTTP address with the " + consulHTTPAddrEnvName + " environment variable")
		} else {
			ConsulConfig = &distributed.Config{
				Address: consulHTTPAddr,
				Token:   consulHTTPToken,
			}
		}
	}

	if AppMosolyOpsAccount == "" {
		printUsageErrorAndExit("provide ethereum fprivate key with " + appMosolyOpsAccountEnvName + " environment variable")
	}

	if EthereumPassportFactoryAddress == "" {
		printUsageErrorAndExit("provide ethereum passport factory address with " + ethereumPassportFactoryAddressEnvName + " environment variable")
	}

	if AppMosolyDidAddress == "" {
		printUsageErrorAndExit("provide ethereum main passport address with " + appMosolyDidAddressEnvName + " environment variable")
	}

	if AppMosolyBackendURL == "" {
		printUsageErrorAndExit("provide mosoly backend URL with " + appMosolyBackendURLEnvName + " environment variable")
	}

	if AppMosolyBackendToken == "" {
		printUsageErrorAndExit("provide mosoly backend token with " + appMosolyBackendURLEnvName + " environment variable")
	}

	AppInDebugMode = appMode == debugMode
	SQLConnectionString = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s connect_timeout=%s read_timeout=%s write_timeout=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPass, dbConnectTimeout, dbReadTimeout, dbWriteTimeout)
}

func getEnvInt(envName string, defaultValue int) int {
	if value, ok := os.LookupEnv(envName); ok {
		result, err := strconv.Atoi(value)
		if err != nil {
			printUsageErrorAndExit(envName+" environment variable contains invalid integer value: %v", value)
		}
		return result
	}
	return defaultValue
}

func getEnvBool(envName string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(envName); ok {
		result, err := strconv.ParseBool(value)
		if err != nil {
			printUsageErrorAndExit(envName+" environment variable contains invalid boolean value: %v", value)
		}
		return result
	}
	return defaultValue
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func printUsageErrorAndExit(format string, values ...interface{}) {
	log.Warn(fmt.Sprintf("ERROR: %s\n", fmt.Sprintf(format, values...)))
	log.Warn("Available command line options:")
	flag.PrintDefaults()
	os.Exit(1)
}
