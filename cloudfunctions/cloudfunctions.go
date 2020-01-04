package cloudfunctions

import (
	"fmt"
	"github.com/alexandre-normand/stepcurry"
	"github.com/spf13/cast"
	"net/http"
	"os"
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

var rc *stepcurry.RogerChallenger

func init() {
	projectID := os.Getenv(projectIDEnv)
	region := os.Getenv(regionEnv)

	appID, slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret, err := loadSecrets(projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Roger Challenger secrets: %s", err.Error()))
	}

	storer, err := stepcurry.NewDatastorer(projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize datastore: %s", err.Error()))
	}

	taskScheduler, err := stepcurry.NewTaskScheduler(projectID, region, challengeUpdatesQueue)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Cloud Tasks Client: %s", err.Error()))
	}

	tokenManager := NewMultiTenantTokenManager(projectID)
	router, err := stepcurry.NewMultiTenantRouter(projectID, storer, tokenManager, tokenManager, cast.ToBool(os.Getenv(debugEnv)))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Roger Challenger: %s", err.Error()))
	}

	roger, err := stepcurry.New(inferBaseURL(projectID, region), appID, fitbitClientID, fitbitClientSecret, slackClientID, slackClientSecret, stepcurry.OptionSlackVerifier(slackSigningSecret), stepcurry.OptionStorer(storer), stepcurry.OptionTeamRouter(router), stepcurry.OptionTaskScheduler(taskScheduler))
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
