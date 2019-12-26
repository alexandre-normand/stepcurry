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

type TokenLoader interface {
	Load(projectID string, teamID string) (token string, err error)
}

type TokenSaver interface {
	Save(projectID string, teamID string, token string) (err error)
}

type SingleTenantTokenManager struct {
}

func (stManager *SingleTenantTokenManager) Load(projectID string, teamID string) (token string, err error) {
	ctx := context.Background()
	ss, err := secretmanager.NewService(ctx)
	if err != nil {
		return "", err
	}

	psvs := secretmanager.NewProjectsSecretsVersionsService(ss)
	token, err = getSecret(psvs, projectID, slackTokenKey)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (stManager *SingleTenantTokenManager) Save(projectID string, teamID string, token string) (err error) {
	return fmt.Errorf("Operation not supported on SingleTenantTokenManager. Consider using [gcloud beta secrets] to save the single tenant slack token")
}

type MultiTenantTokenManager struct {
}

func (mtManager *MultiTenantTokenManager) Load(projectID string, teamID string) (token string, err error) {
	ctx := context.Background()
	ss, err := secretmanager.NewService(ctx)
	if err != nil {
		return "", err
	}

	psvs := secretmanager.NewProjectsSecretsVersionsService(ss)
	token, err = getSecret(psvs, projectID, formatSecretKeyWithTeamNamespace(teamID, slackTokenKey))

	if err != nil {
		return "", err
	}

	return token, nil
}

func (mtManager *MultiTenantTokenManager) Save(projectID string, teamID string, token string) (err error) {
	ctx := context.Background()
	ss, err := secretmanager.NewService(ctx)
	if err != nil {
		return err
	}

	psvs := secretmanager.NewProjectsSecretsVersionsService(ss)
	pss := secretmanager.NewProjectsSecretsService(ss)

	tokenQualifiedKey := formatSecretKeyWithTeamNamespace(teamID, slackTokenKey)
	fullyQualifiedSecretName := formatSecretName(projectID, tokenQualifiedKey)
	_, err = getSecret(psvs, projectID, tokenQualifiedKey)
	if err != nil {
		if apiError, ok := err.(*googleapi.Error); ok && apiError.Code == http.StatusNotFound {
			// Create the secret
			call := pss.Create(fmt.Sprintf("projects/%s", projectID), &secretmanager.Secret{Labels: map[string]string{"team": strings.ToLower(teamID)}, Replication: &secretmanager.Replication{Automatic: new(secretmanager.Automatic)}})
			call = call.SecretId(tokenQualifiedKey)

			_, err = call.Do()
			if err != nil {
				// Ignore errors for a secret already existing (which can happen if no versions were found)
				if apiError, ok := err.(*googleapi.Error); !ok || apiError.Code != http.StatusConflict {
					return err
				}
			}
		} else {
			return err
		}
	}

	call := pss.AddVersion(fullyQualifiedSecretName, &secretmanager.AddSecretVersionRequest{Payload: &secretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString([]byte(token))}})
	_, err = call.Do()
	return err
}

func loadSecrets(projectID string) (slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret string, err error) {
	ctx := context.Background()
	ss, err := secretmanager.NewService(ctx)
	if err != nil {
		return "", "", "", "", "", err
	}

	psvs := secretmanager.NewProjectsSecretsVersionsService(ss)

	slackSigningSecret, err = getSecret(psvs, projectID, signingSecretKey)
	if err != nil {
		return "", "", "", "", "", err
	}

	slackClientID, err = getSecret(psvs, projectID, slackClientIDKey)
	if err != nil {
		return "", "", "", "", "", err
	}

	slackClientSecret, err = getSecret(psvs, projectID, slackClientSecretKey)
	if err != nil {
		return "", "", "", "", "", err
	}

	fitbitClientID, err = getSecret(psvs, projectID, fitbitClientIDKey)
	if err != nil {
		return "", "", "", "", "", err
	}

	fitbitClientSecret, err = getSecret(psvs, projectID, fitbitClientSecretKey)
	if err != nil {
		return "", "", "", "", "", err
	}

	return slackClientID, slackClientSecret, slackSigningSecret, fitbitClientID, fitbitClientSecret, nil
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

func formatSecretKeyWithTeamNamespace(teamID string, key string) (teamKey string) {
	return fmt.Sprintf("%s-%s", key, teamID)
}

func formatSecretNameWithTeamNamespace(projectID string, teamID string, key string) (qualifiedName string) {
	return fmt.Sprintf("projects/%s/secrets/%s", projectID, formatSecretKeyWithTeamNamespace(teamID, key))
}

func formatSecretName(projectID string, key string) (qualifiedName string) {
	return fmt.Sprintf("projects/%s/secrets/%s", projectID, key)
}

func formatLatestVersionSecretName(projectID string, key string) (qualifiedName string) {
	return fmt.Sprintf("%s/versions/latest", formatSecretName(projectID, key))
}
