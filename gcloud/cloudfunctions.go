package gcloud

import (
	"context"
	"fmt"
	"github.com/alexandre-normand/rogerchallenger"
	"github.com/spf13/cast"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
	"net/http"
	"os"
)

// Secret names
const (
	slackTokenKey         = "slackToken"
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

var rc *rogerchallenger.RogerChallenger

func init() {
	projectID := os.Getenv(projectIDEnv)
	region := os.Getenv(regionEnv)

	slackToken, slackSigningSecret, fitbitClientID, fitbitClientSecret, err := loadSecrets()
	if err != nil {
		panic(fmt.Sprintf("Failed to load Roger Challenger secrets: %s", err.Error()))
	}

	storer, err := NewDatastorer(projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize datastore: %s", err.Error()))
	}

	taskScheduler, err := NewTaskScheduler(projectID, region, challengeUpdatesQueue)
	if err != nil {
		log.Fatal(err.Error())
	}

	slackClient := slack.New(slackToken, cast.ToBool(slack.OptionDebug(debugEnv)))

	roger, err := rogerchallenger.New(inferBaseURL(os.Getenv(projectIDEnv), os.Getenv(regionEnv)), fitbitClientID, fitbitClientSecret, OptionSlackVerifier(os.Getenv(slackSigningSecretEnvKey)), OptionStorer(storer), OptionUserInfoFinder(slackClient), OptionMessenger(slackClient), OptionChannelInfoFinder(slackClient), OptionTaskScheduler(taskScheduler))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Roger Challenger: %s", err.Error()))
	}

	rc = roger
}

func inferBaseURL(projectID string, region string) (baseURL string) {
	return fmt.Sprintf("https://%s-%s.cloudfunctions.net", projectID, region)
}

func loadSecrets() (slackToken, slackSigningSecret, fitbitClientID, fitbitCLientSecret string, err error) {
	ctx := context.Background()
	ss := secretmanager.NewService(ctx)
	psvs := secretmanager.NewProjectsSecretsVersionsService(ss)
	slackToken, err = getSecret(psvs, os.Getenv(projectIDEnv), slackTokenKey)
	if err != nil {
		return err
	}

	slackSigningSecret, err = getSecret(psvs, os.Getenv(projectIDEnv), signingSecretKey)
	if err != nil {
		return err
	}

	fitbitClientID, err = getSecret(psvs, os.Getenv(projectIDEnv), fitbitClientIDKey)
	if err != nil {
		return err
	}

	fitbitClientSecret, err = getSecret(psvs, os.Getenv(projectIDEnv), fitbitClientSecretKey)
	if err != nil {
		return err
	}

	return slackToken, slackSigningSecret, fitbitClientID, fitbitClientSecret, nil
}

func getSecret(psvs *secretmanager.ProjectsSecretsVersionsService, projectID string, key string) (value string, err error) {
	request := pss.Access(key)
	resp, err := request.Do()
	if err != nil {
		return "", err
	}

	return resp.Payload.Data, nil
}

func formatSecretName(projectID string, key string) (qualifiedName string) {
	fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, key)
}
