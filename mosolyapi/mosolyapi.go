package mosolyapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/p-invent/mosoly-ledger-bridge/config"

	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/http/rest"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/http/rest/responses"
)

// Client is an Mosoly API client.
type Client struct {
	*rest.Client
}

// NewClient creates a new Mosoly API client.
func NewClient(httpClient *http.Client, apiBaseURL string) (*Client, error) {
	return &Client{
		rest.NewClient(apiBaseURL).WithClient(httpClient),
	}, nil
}

// BlockchainFact struct describing initial structure for any fact written to blockchain
type BlockchainFact struct {
	Schema  string      `json:"schema"`
	Payload interface{} `json:"payload"`
}

// BlockchainProjectFact is a wrapper for project
type BlockchainProjectFact struct {
	BlockchainFact
	Payload ProjectFact `json:"payload"`
}

// ProjectFact is stored on the Blockchain
type ProjectFact struct {
	Name string `json:"name"`
}

// Project is Mosoly API project
type Project struct {
	ProjectFact
	ID        int       `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BlockchainMentorFact is a wrapper for mentor fact
type BlockchainMentorFact struct {
	BlockchainFact
	Payload MentorFact `json:"payload"`
}

// MentorFact is an array of facts for mentor
type MentorFact []string

// BlockchainUserFact is a wrapper for every fact that'd be stored on the Blockchain
type BlockchainUserFact struct {
	BlockchainFact
	Payload UserFact `json:"payload"`
}

// UserFact is a fact about the user that stored on the Blockchain
type UserFact struct {
	InviteURLHash string   `json:"inviteUrlHash"`
	Account       string   `json:"account"`
	Validated     bool     `json:"validated"`
	Mentors       []string `json:"mentors"`
}

// User is Mosoly API user
type User struct {
	ID            int        `json:"id"`
	InviteURLHash string     `json:"inviteUrlHash"`
	Account       string     `json:"account"`
	Validated     bool       `json:"validated"`
	JoinedAt      time.Time  `json:"joinedAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	Mentorees     []Mentoree `json:"mentorees"`
	Mentors       []Mentor   `json:"mentors"`
}

// Mentoree a Mosoly API User that is being mentored
type Mentoree struct {
	UserID              int       `json:"userId"`
	CustomNameForMentor string    `json:"customNameForMentor"`
	Account             string    `json:"account"`
	Validated           bool      `json:"validated"`
	MentorshipStarted   time.Time `json:"mentorshipStarted"`
}

// Mentor is a Mosoly API User that is mentoring some user
type Mentor struct {
	UserID            int       `json:"userId"`
	Account           string    `json:"account"`
	MentorshipStarted time.Time `json:"mentorshipStarted"`
}

// GetUserUpdates gets user and mentors updates from the Mosoly API
func (c *Client) GetUserUpdates(ctx context.Context, since time.Time) ([]User, error) {
	var errResp *responses.Error
	var resp []User

	path := fmt.Sprintf("/users?since=%d", since.UTC().Unix())

	_, err := c.NewEndpoint(ctx).
		Get(path).
		WithBearerAuth(config.AppMosolyBackendToken).
		SendAndParse(&resp, &errResp)
	if err != nil {
		return nil, errorf("failed to make request: %v", err)
	}
	return resp, nil
}

// GetProjectUpdates return project updates
func (c *Client) GetProjectUpdates(ctx context.Context, since time.Time) ([]Project, error) {
	// TODO: Implement according https://portal.mosoli.live/api specification
	return []Project{}, nil
}

func errorf(msg string, args ...interface{}) error {
	return fmt.Errorf("mosolyapi: "+msg, args...)
}
