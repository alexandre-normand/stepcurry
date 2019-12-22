package rogerchallenger

import (
	"bytes"
	"cloud.google.com/go/datastore"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Default base URLs for Fitbit access
const (
	defaultFitbitAuthBaseURL = "https://www.fitbit.com"
	defaultFitbitAPIBaseURL  = "https://api.fitbit.com"
)

// FitbitApiAcccess holds data for an authenticated fitbit user
type FitbitApiAccess struct {
	Token        string `datastore:"accessToken,noindex" json:"access_token,omitempty"`
	FitbitUser   string `datastore:"fitbitUser,noindex" json:"user_id,omitempty"`
	RefreshToken string `datastore:"refreshToken,noindex" json:"refresh_token,omitempty"`
}

// AuthIdentificationState holds data RogerChallenger requires to reconcile a oauth callback
// with the requesting slack user
type AuthIdentificationState struct {
	SlackUser    string `json:"slackUser"`
	SlackChannel string `json:"slackChannel"`
	SlackTeam    string `json:"slackTeam"`
	ResponseURL  string `json:"responseURL"`
	CsrfToken
}

// CsrfToken holds the Csrf randomly generated value
type CsrfToken struct {
	Csrf []byte `datastore:"csrf,noindex" json:"csrf,omitempty"`
}

// Summary holds steps and floors summary data. This is a subset of the full data
// returned by the API (other fields are ignored).
// See details at https://dev.fitbit.com/build/reference/web-api/activity/#get-daily-activity-summary
type Summary struct {
	Steps  int `json:"steps,omitempty"`
	Floors int `json:"floors,omitempty"`
}

// ActivitySummaryResponse holds data RogerChallenger uses from the Fitbit activity web API. See
// details at https://dev.fitbit.com/build/reference/web-api/activity/#get-daily-activity-summary
type ActivitySummaryResponse struct {
	Goals   Goals   `json:"goals,omitempty"`
	Summary Summary `json:"summary,omitempty"`
}

// Goals holds configured goals that RogerChallenger cares about. See details at
// https://dev.fitbit.com/build/reference/web-api/activity/#get-daily-activity-summary
type Goals struct {
	Steps int `json:"steps,omitempty"`
}

// HandleFitbitAuth receives the oauth callback from Fitbit after a user has logged in and
// consented to the access
func (rc *RogerChallenger) HandleFitbitAuth(w http.ResponseWriter, r *http.Request) {
	codes, ok := r.URL.Query()["code"]
	if !ok {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	code := codes[0]

	stateVal, ok := r.URL.Query()["state"]
	if !ok {
		http.Error(w, "Missing Auth Identification State", http.StatusBadRequest)
		return
	}

	stateBase64 := stateVal[0]
	var authIDState AuthIdentificationState
	rawState, err := base64.URLEncoding.DecodeString(stateBase64)
	if err != nil {
		log.Printf("Error base64 decoding slack Auth Identification State: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(rawState, &authIDState)
	if err != nil {
		log.Printf("Error decoding Auth Identification State json: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var csrfToken CsrfToken
	csrfKey := NewKeyWithNamespace("CsrfToken", authIDState.SlackTeam, authIDState.SlackUser, nil)
	err = rc.storer.Get(ctx, csrfKey, &csrfToken)

	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Printf("CSRF token not found")
			http.Error(w, "CSRF token not found", http.StatusUnauthorized)
			return
		} else {
			log.Printf("Error fetching csrf token for user [%s]: %s", authIDState.SlackUser, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if !bytes.Equal(authIDState.Csrf, csrfToken.Csrf) {
		log.Printf("CSRF token mismatch")
		http.Error(w, "Invalid CSRF Token", http.StatusUnauthorized)
		return
	}

	err = rc.storer.Delete(ctx, csrfKey)
	if err != nil {
		log.Printf("Error deleting up csrf token for user [%s]: %s", authIDState.SlackUser, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiAccess, err := rc.exchangeAuthCodeForApiAccess(code, string(rawState))
	if err != nil {
		log.Printf("Error getting fitbit api access for user [%s]", authIDState.SlackUser)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clientAccess := ClientAccess{SlackUser: authIDState.SlackUser, SlackTeam: authIDState.SlackTeam, FitbitApiAccess: apiAccess}

	k := NewKeyWithNamespace("ClientAccess", authIDState.SlackTeam, authIDState.SlackUser, nil)
	_, err = rc.storer.Put(ctx, k, &clientAccess)
	if err != nil {
		log.Printf("Error persisting fitbit key for user [%s]", authIDState.SlackUser)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oauthCompleteMessage := ActionResponse{ResponseType: "ephemeral", ReplaceOriginal: false, Text: "POW :boom: You've got your Fitbit account linked and ready for some challenges :wind_blowing_face::athletic_shoe:"}
	resp, err := req.Post(authIDState.ResponseURL, req.BodyJSON(&oauthCompleteMessage))
	if err != nil || resp.Response().StatusCode != 200 {
		if err != nil {
			log.Printf("Error sending oauth completion message: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			log.Printf("Error writing oauth completion message: %s", resp.String())
			http.Error(w, resp.String(), http.StatusInternalServerError)
		}

		return
	}

	// We could do a server-side redirect but a client-side redirect clears the Fitbit consent page and looks more "done" to the user so that's the approach we're taking here
	w.Write([]byte(fmt.Sprintf("<html><head><meta http-equiv=\"refresh\" content=\"0;URL=slack://channel?team=%s&id=%s\"></head></html>", authIDState.SlackTeam, authIDState.SlackChannel)))
}

// exchangeAuthCodeForApiAccess runs a query with the Fitbit authentication API to exchange an auth code for an access token
func (rc *RogerChallenger) exchangeAuthCodeForApiAccess(code string, state string) (apiAccess FitbitApiAccess, err error) {
	v := url.Values{}
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")
	v.Set("redirect_uri", fmt.Sprintf("%s/%s", rc.baseURL, oauthCallbackPath))
	v.Set("state", state)
	v.Set("client_id", rc.fitbitClientID)

	body := strings.NewReader(v.Encode())
	tokenURL := fmt.Sprintf("%s/oauth2/token", rc.fitbitAPIBaseURL)

	req, err := http.NewRequest("POST", tokenURL, body)
	if err != nil {
		return apiAccess, errors.Wrap(err, "error creating access token request")
	}

	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", rc.fitbitClientID, rc.fitbitClientSecret)))))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return apiAccess, errors.Wrap(err, "error executing access token request")
	}

	tokenBody, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return apiAccess, fmt.Errorf("error getting access token [%s]: %s", resp.Status, tokenBody)
	}

	err = json.Unmarshal(tokenBody, &apiAccess)
	if err != nil {
		return apiAccess, errors.Wrap(err, "error decoding api access response")
	}

	return apiAccess, nil
}

// UserSteps holds a slack user and its step count
type UserSteps struct {
	UserID string
	Steps  int
}

// byStepCount sorts by the step count
type byStepCount []UserSteps

func (p byStepCount) Len() int { return len(p) }

func (p byStepCount) Less(i, j int) bool {
	return p[i].Steps < p[j].Steps || (p[i].Steps == p[j].Steps && strings.Compare(p[i].UserID, p[j].UserID) > 0)
}

func (p byStepCount) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// getChallengeRankedSteps fetches the updated ranking of all fitbit users participating in a steps challenge
func (rc *RogerChallenger) getChallengeRankedSteps(stepsChallenge StepsChallenge) (rankedUsers []UserSteps, err error) {
	userSteps := make([]UserSteps, 0)

	localizedChallengeDate, _, err := localizeCreationTime(stepsChallenge.CreationTime, stepsChallenge.TimezoneID)
	if err != nil {
		return userSteps, errors.Wrapf(err, "error getting localized time for steps challenge [%s.%s]", stepsChallenge.TeamID, stepsChallenge.ChannelID)
	}

	ctx := context.Background()
	q := datastore.NewQuery("ClientAccess").Namespace(stepsChallenge.TeamID)
	usersIterator := rc.storer.Run(ctx, q)

	fitbitUsers := make(map[string]ClientAccess)
	var ca ClientAccess
	for _, err := usersIterator.Next(&ca); err == nil; _, err = usersIterator.Next(&ca) {
		if err != nil && err != iterator.Done {
			return userSteps, err
		}

		fitbitUsers[ca.SlackUser] = ca
	}

	ch, err := rc.channelInfoFinder.GetChannelInfo(stepsChallenge.ChannelID)
	if err != nil {
		return userSteps, errors.Wrapf(err, "error getting channel members for channel id [%s]", stepsChallenge.ChannelID)
	}

	usersToFetch := make([]string, 0)
	for _, userID := range ch.Members {
		if _, ok := fitbitUsers[userID]; ok {
			usersToFetch = append(usersToFetch, userID)
		}
	}

	// TODO: Create a worker pool and submit the work with a parallelism of 4
	for _, user := range usersToFetch {
		clientAccess := fitbitUsers[user]

		if steps, err := rc.getUserSteps(clientAccess, localizedChallengeDate); err != nil {
			log.Printf("Error reading step count for user [%s]: %s", user, err.Error())
		} else {
			userSteps = append(userSteps, UserSteps{UserID: user, Steps: steps})
		}
	}

	sort.Sort(sort.Reverse(byStepCount(userSteps)))
	return userSteps, nil
}

// getUserSteps retrieves the steps summary for a given fitbit user using its access token
func (rc *RogerChallenger) getUserSteps(clientAccess ClientAccess, date time.Time) (steps int, err error) {
	resp, err := rc.fetchActivitySummaryWithRefresh(clientAccess, date)
	if err != nil {
		return 0, errors.Wrapf(err, "error fetching activity summary for user [%s]", clientAccess.FitbitUser)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, fmt.Errorf("error getting activity summary [%s]: %s", resp.Status, body)
	}

	var activitySummaryResp ActivitySummaryResponse
	err = json.Unmarshal(body, &activitySummaryResp)
	if err != nil {
		return 0, errors.Wrap(err, "error decoding activity summary response")
	}

	return activitySummaryResp.Summary.Steps, nil
}

// fetchActivitySummaryWithRefresh fetches a user's activity summary and handles expired tokens by refreshing the token
// if necessary
func (rc *RogerChallenger) fetchActivitySummaryWithRefresh(clientAccess ClientAccess, date time.Time) (resp *http.Response, err error) {
	resp, err = rc.fetchActivitySummary(clientAccess, date)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		log.Printf("Token expired for user [%s], refreshing...", clientAccess.SlackUser)

		// Refresh token
		apiAccess, err := rc.exchangeRefreshTokenForApiAccess(clientAccess.RefreshToken)
		if err != nil {
			return nil, errors.Wrapf(err, "error refreshing token for user [%s]", clientAccess.SlackUser)
		}

		clientAccess = ClientAccess{SlackUser: clientAccess.SlackUser, SlackTeam: clientAccess.SlackTeam, FitbitApiAccess: apiAccess}

		ctx := context.Background()
		k := NewKeyWithNamespace("ClientAccess", clientAccess.SlackTeam, clientAccess.SlackUser, nil)
		_, err = rc.storer.Put(ctx, k, &clientAccess)
		if err != nil {
			return nil, errors.Wrapf(err, "Error persisting fitbit key for user [%s]", clientAccess.SlackUser)
		}

		// Refetch activity summary
		resp, err = rc.fetchActivitySummary(clientAccess, date)
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching activity summary for user [%s]", clientAccess.SlackUser)
		}
	}

	return resp, nil
}

// fetchActivitySummary fetches a user's activity summary
func (rc *RogerChallenger) fetchActivitySummary(clientAccess ClientAccess, date time.Time) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/1/user/%s/activities/date/%s.json", rc.fitbitAPIBaseURL, clientAccess.FitbitUser, date.Format(fitbitDateFormat)), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating activity summary request")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", clientAccess.Token))

	client := http.Client{Timeout: 3 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading activity summary for fitbit user id [%s]", clientAccess.FitbitUser)
	}

	return resp, nil
}

// exchangeRefreshTokenForApiAccess runs a request against the Fitbit authentication API to exchange a refresh token
// to get a new access token (and updated refresh token)
func (rc *RogerChallenger) exchangeRefreshTokenForApiAccess(refreshToken string) (apiAccess FitbitApiAccess, err error) {
	v := url.Values{}
	v.Set("grant_type", "refresh_token")
	v.Set("refresh_token", refreshToken)

	body := strings.NewReader(v.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth2/token", rc.fitbitAPIBaseURL), body)
	if err != nil {
		return apiAccess, errors.Wrap(err, "error creating refresh token request")
	}

	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", rc.fitbitClientID, rc.fitbitClientSecret)))))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return apiAccess, errors.Wrap(err, "error executing refresh token request")
	}

	tokenBody, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return apiAccess, fmt.Errorf("error getting refresh token [%s]: %s", resp.Status, tokenBody)
	}

	err = json.Unmarshal(tokenBody, &apiAccess)
	if err != nil {
		return apiAccess, errors.Wrap(err, "error decoding api access response")
	}

	return apiAccess, nil
}
