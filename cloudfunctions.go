package rogerchallenger

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/nlopes/slack"
	"github.com/spf13/cast"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
	"net/http"
	"os"
	"strings"
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

	slackToken, slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret, err := loadSecrets(projectID)
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

	slackClient := slack.New(slackToken, slack.OptionDebug(cast.ToBool(os.Getenv(debugEnv))))

	roger, err := New(inferBaseURL(projectID, region), fitbitClientID, fitbitClientSecret, slackClientID, slackClientSecret, OptionSlackVerifier(slackSigningSecret), OptionStorer(storer), OptionUserInfoFinder(slackClient), OptionMessenger(slackClient), OptionChannelInfoFinder(slackClient), OptionTaskScheduler(taskScheduler))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Roger Challenger: %s", err.Error()))
	}

	rc = roger
}

func inferBaseURL(projectID string, region string) (baseURL string) {
	return fmt.Sprintf("https://%s-%s.cloudfunctions.net", region, projectID)
}

func loadSecrets(projectID string) (slackToken, slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret string, err error) {
	ctx := context.Background()
	ss, err := secretmanager.NewService(ctx)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	psvs := secretmanager.NewProjectsSecretsVersionsService(ss)
	slackToken, err = getSecret(psvs, projectID, slackTokenKey)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	slackSigningSecret, err = getSecret(psvs, projectID, signingSecretKey)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	slackClientID, err = getSecret(psvs, projectID, slackClientIDKey)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	slackClientSecret, err = getSecret(psvs, projectID, slackClientSecretKey)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	fitbitClientID, err = getSecret(psvs, projectID, fitbitClientIDKey)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	fitbitClientSecret, err = getSecret(psvs, projectID, fitbitClientSecretKey)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	return slackToken, slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret, nil
}

func getSecret(psvs *secretmanager.ProjectsSecretsVersionsService, projectID string, key string) (value string, err error) {
	request := psvs.Access(formatSecretName(projectID, key))
	resp, err := request.Do()
	if err != nil {
		return "", err
	}

	secret, err := base64.StdEncoding.DecodeString(resp.Payload.Data)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(secret)), nil
}

func formatSecretName(projectID string, key string) (qualifiedName string) {
	return fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, key)
}

func LinkAccount(w http.ResponseWriter, r *http.Request) {
	rc.StartFitbitOauthFlow(w, r)
}

func Challenge(w http.ResponseWriter, r *http.Request) {
	rc.StartChallenge(w, r)
}

func UpdateChallenge(w http.ResponseWriter, r *http.Request) {
	rc.UpdateChallenge(w, r)
}

func HandleFitbitAuth(w http.ResponseWriter, r *http.Request) {
	rc.HandleFitbitAuthorized(w, r)
}

func Standings(w http.ResponseWriter, r *http.Request) {
	rc.ChallengeStandings(w, r)
}

func InvokeSlackAuth(w http.ResponseWriter, r *http.Request) {
	rc.InvokeSlackAuth(w, r)
}

func HandleSlackAuth(w http.ResponseWriter, r *http.Request) {
	rc.HandleSlackAuth(w, r)
}
