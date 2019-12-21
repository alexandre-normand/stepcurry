package rogerchallenger

import (
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

var slackScopes = [...]string{"chat:write", "users:read", "users.profile:read", "channels:read", "commands"}

const (
	defaultSlackBaseURL = "https://slack.com"
)

type SlackAccess struct {
	AccessToken string  `json:"access_token"`
	Scope       string  `json:"scope"`
	TeamName    string  `json:"team_name,omitempty"`
	TeamID      string  `json:"team_id,omitempty"`
	Bot         BotInfo `json:"bot,omitempty"`
}

type BotInfo struct {
	UserID      string `json:"bot_user_id,omitempty"`
	AccessToken string `json:"bot_access_token,omitempty"`
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

	slackAccess, err := rc.exchangeSlackAuthCodeForToken(code)
	if err != nil {
		log.Printf("Error getting slack access: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Got slack access: %v\n", slackAccess)
}

func (rc *RogerChallenger) exchangeSlackAuthCodeForToken(code string) (slackAccess SlackAccess, err error) {
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
		return slackAccess, errors.Wrap(err, "error creating slack access token request")
	}

	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", rc.slackClientID, rc.slackClientSecret)))))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return slackAccess, errors.Wrap(err, "error executing slack access token request")
	}

	tokenBody, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return slackAccess, fmt.Errorf("error getting slack access token [%s]: %s", resp.Status, tokenBody)
	}

	log.Printf("token body is [%s]", tokenBody)
	err = json.Unmarshal(tokenBody, &slackAccess)
	if err != nil {
		return slackAccess, errors.Wrap(err, "error decoding slack access response")
	}

	return slackAccess, nil
}
