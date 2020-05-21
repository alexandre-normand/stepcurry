package cloudfunctions

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alexandre-normand/stepcurry"
	"github.com/spf13/cast"
)

// GCP Environment Variables
const (
	projectIDEnv = "PROJECT_ID"
	regionEnv    = "DEPLOY_REGION"
	debugEnv     = "DEBUG"
)

// Cloud Tasks Queues
const (
	challengeUpdatesQueue = "challenge-updates"
)

var sc *stepcurry.StepCurry

func init() {
	projectID := os.Getenv(projectIDEnv)
	region := os.Getenv(regionEnv)

	appID, slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret, err := loadSecrets(projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Step Curry secrets: %s", err.Error()))
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
		panic(fmt.Sprintf("Failed to initialize Step Curry: %s", err.Error()))
	}

	step, err := stepcurry.New(inferBaseURL(projectID, region), appID, fitbitClientID, fitbitClientSecret, slackClientID, slackClientSecret, stepcurry.OptionSlackVerifier(slackSigningSecret), stepcurry.OptionStorer(storer), stepcurry.OptionTeamRouter(router), stepcurry.OptionTaskScheduler(taskScheduler))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Step Curry: %s", err.Error()))
	}

	sc = step
}

func inferBaseURL(projectID string, region string) (baseURL string) {
	return fmt.Sprintf("https://%s-%s.cloudfunctions.net", region, projectID)
}

func LinkAccount(w http.ResponseWriter, r *http.Request) {
	stepcurry.Handler(sc.StartFitbitOauthFlow).ServeHTTP(w, r)
}

func Challenge(w http.ResponseWriter, r *http.Request) {
	stepcurry.Handler(sc.Challenge).ServeHTTP(w, r)
}

func UpdateChallenge(w http.ResponseWriter, r *http.Request) {
	stepcurry.Handler(sc.UpdateChallenge).ServeHTTP(w, r)
}

func HandleFitbitAuth(w http.ResponseWriter, r *http.Request) {
	stepcurry.Handler(sc.HandleFitbitAuth).ServeHTTP(w, r)
}

func Standings(w http.ResponseWriter, r *http.Request) {
	stepcurry.Handler(sc.Standings).ServeHTTP(w, r)
}

func InvokeSlackAuth(w http.ResponseWriter, r *http.Request) {
	sc.InvokeSlackAuth(w, r)
}

func HandleSlackAuth(w http.ResponseWriter, r *http.Request) {
	stepcurry.Handler(sc.HandleSlackAuth).ServeHTTP(w, r)
}
