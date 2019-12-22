package rogerchallenger

import (
	"context"
	"encoding/base64"
	"fmt"
	"google.golang.org/api/googleapi"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
	"net/http"
	"strings"
)

func loadSecrets(projectID string) (slackToken, slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret string, err error) {
	ctx := context.Background()
	ss, err := secretmanager.NewService(ctx)
	if err != nil {
		return "", "", "", "", "", "", err
	}

	psvs := secretmanager.NewProjectsSecretsVersionsService(ss)
	// Ignore error on slackToken retrieval as it might not exist
	// if the app has never been installed yet
	slackToken, _ = getSecret(psvs, projectID, slackTokenKey)

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
	request := psvs.Access(formatLatestVersionSecretName(projectID, key))
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

func saveSecret(projectID string, key string, value string) (err error) {
	ctx := context.Background()
	ss, err := secretmanager.NewService(ctx)
	if err != nil {
		return err
	}

	psvs := secretmanager.NewProjectsSecretsVersionsService(ss)
	pss := secretmanager.NewProjectsSecretsService(ss)

	_, err = getSecret(psvs, projectID, key)
	if apiError, ok := err.(*googleapi.Error); ok && apiError.Code == http.StatusNotFound {
		// Create the secret
		call := pss.Create(formatSecretName(projectID, key), &secretmanager.Secret{Replication: &secretmanager.Replication{Automatic: new(secretmanager.Automatic)}})
		_, err := call.Do()
		if err != nil {
			return err
		}
	}

	call := pss.AddVersion(formatSecretName(projectID, key), &secretmanager.AddSecretVersionRequest{Payload: &secretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString([]byte(value))}})
	_, err = call.Do()
	return err
}

func formatSecretName(projectID string, key string) (qualifiedName string) {
	return fmt.Sprintf("projects/%s/secrets/%s", projectID, key)
}

func formatLatestVersionSecretName(projectID string, key string) (qualifiedName string) {
	return fmt.Sprintf("%s/versions/latest", formatSecretName(projectID, key))
}
