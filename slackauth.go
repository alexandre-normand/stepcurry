package stepcurry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var slackScopes = [...]string{"chat:write", "users:read", "users.profile:read", "channels:read", "groups:read", "im:read", "mpim:read", "commands"}

const (
	defaultSlackBaseURL = "https://slack.com"
)

type SlackAuthResponse struct {
	Ok          bool       `json:"ok,omitempty"`
	AppID       string     `json:"app_id,omitempty"`
	AuthedUser  AuthedUser `json:"authed_user,omitempty"`
	Scope       string     `json:"scope,omitempty"`
	TokenType   string     `json:"token_type,omitempty"`
	AccessToken string     `json:"access_token,omitempty"`
	BotUserID   string     `json:"bot_user_id,omitempty"`
	Team        TeamInfo   `json:"team,omitempty"`
	Enterprise  string     `json:"enterprise,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type AuthedUser struct {
	ID string `json:"id,omitempty"`
}

type TeamInfo struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (rc *RogerChallenger) InvokeSlackAuth(w http.ResponseWriter, r *http.Request) {
	redirectURI := fmt.Sprintf("%s/%s", rc.baseURL, "HandleSlackAuth")
	slackAuthURL := fmt.Sprintf("%s/oauth/v2/authorize?client_id=%s&redirect_uri=%s&scope=%s", rc.slackBaseURL, rc.slackClientID, redirectURI, strings.Join(slackScopes[:], ","))
	http.Redirect(w, r, slackAuthURL, http.StatusFound)
}

func (rc *RogerChallenger) HandleSlackAuth(w http.ResponseWriter, r *http.Request) {
	codes, ok := r.URL.Query()["code"]
	if !ok {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	code := codes[0]

	authResp, err := rc.exchangeSlackAuthCodeForToken(code)
	if err != nil {
		log.Printf("Error getting slack access: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rc.SaveToken(authResp.Team.ID, authResp.AccessToken)
	if err != nil {
		log.Printf("Error saving slack token: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	botInfo := BotInfo{UserID: authResp.BotUserID}
	k := NewKeyWithNamespace("BotInfo", authResp.Team.ID, "Bot", nil)
	_, err = rc.storer.Put(ctx, k, &botInfo)
	if err != nil {
		log.Printf("Error persisting bot info [%s] for team [%s]: %s", botInfo.UserID, authResp.Team.ID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("<html><head><meta http-equiv=\"refresh\" content=\"0;URL=https://slack.com/app_redirect?app=%s&team=%s\"></head></html>", rc.slackAppID, authResp.Team.ID)))
}

func (rc *RogerChallenger) exchangeSlackAuthCodeForToken(code string) (authResp SlackAuthResponse, err error) {
	redirectURI := fmt.Sprintf("%s/%s", rc.baseURL, "HandleSlackAuth")

	v := url.Values{}
	v.Set("client_id", rc.slackClientID)
	v.Set("client_secret", rc.slackClientSecret)
	v.Set("code", code)
	v.Set("redirect_uri", redirectURI)

	body := strings.NewReader(v.Encode())
	tokenURL := fmt.Sprintf("%s/api/oauth.v2.access", rc.slackBaseURL)

	req, err := http.NewRequest("POST", tokenURL, body)
	if err != nil {
		return authResp, errors.Wrap(err, "error creating slack access token request")
	}

	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", rc.slackClientID, rc.slackClientSecret)))))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return authResp, errors.Wrap(err, "error executing slack access token request")
	}

	tokenBody, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return authResp, fmt.Errorf("error getting slack access token [%s]: %s", resp.Status, tokenBody)
	}

	err = json.Unmarshal(tokenBody, &authResp)
	if err != nil {
		return authResp, errors.Wrap(err, "error decoding slack auth response")
	}

	if !authResp.Ok {
		return authResp, errors.New(authResp.Error)
	}

	return authResp, nil
}
