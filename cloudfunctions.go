package rogerchallenger

import (
	"fmt"
	"net/http"
	"os"
)

// Secret names
const (
	slackTokenKey         = "slackToken"
	slackClientIDKey      = "slackClientID"
	slackClientSecretKey  = "slackClientSecret"
	signingSecretKey      = "slackSigningSecret"
	fitbitClientIDKey     = "fitbitClientID"
	fitbitClientSecretKey = "fitbitClientSecret"
)

// GCP Environment Variables
const (
	projectIDEnv = "GCP_PROJECT"
	regionEnv    = "FUNCTION_REGION"
	debugEnv     = "DEBUG"
)

// Cloud Tasks Queues
const (
	challengeUpdatesQueue = "challenge-updates"
)

var rc *RogerChallenger

func init() {
	projectID := os.Getenv(projectIDEnv)
	region := os.Getenv(regionEnv)

	slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret, err := loadSecrets(projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Roger Challenger secrets: %s", err.Error()))
	}

	storer, err := NewDatastorer(projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize datastore: %s", err.Error()))
	}

	taskScheduler, err := NewTaskScheduler(projectID, region, challengeUpdatesQueue)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Cloud Tasks Client: %s", err.Error()))
	}

	tokenManager := &MultiTenantTokenManager{}
	router, err := NewMultiTenantRouter(projectID, storer, tokenManager, tokenManager)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Roger Challenger: %s", err.Error()))
	}

	roger, err := New(inferBaseURL(projectID, region), fitbitClientID, fitbitClientSecret, slackClientID, slackClientSecret, OptionSlackVerifier(slackSigningSecret), OptionStorer(storer), OptionTeamRouter(router), OptionTaskScheduler(taskScheduler))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Roger Challenger: %s", err.Error()))
	}

	rc = roger
}

func inferBaseURL(projectID string, region string) (baseURL string) {
	return fmt.Sprintf("https://%s-%s.cloudfunctions.net", region, projectID)
}

func LinkAccount(w http.ResponseWriter, r *http.Request) {
	rc.StartFitbitOauthFlow(w, r)
}

func Challenge(w http.ResponseWriter, r *http.Request) {
	rc.Challenge(w, r)
}

func UpdateChallenge(w http.ResponseWriter, r *http.Request) {
	rc.UpdateChallenge(w, r)
}

func HandleFitbitAuth(w http.ResponseWriter, r *http.Request) {
	rc.HandleFitbitAuth(w, r)
}

func Standings(w http.ResponseWriter, r *http.Request) {
	rc.Standings(w, r)
}

func InvokeSlackAuth(w http.ResponseWriter, r *http.Request) {
	rc.InvokeSlackAuth(w, r)
}

func HandleSlackAuth(w http.ResponseWriter, r *http.Request) {
	rc.HandleSlackAuth(w, r)
}
