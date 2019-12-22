package rogerchallenger

import (
	"cloud.google.com/go/datastore"
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// Env variables, all mandatory
const (
	slackTokenEnvKey         = "SLACK_TOKEN"
	slackSigningSecretEnvKey = "SLACK_SIGNING_SECRET"
	fitbitClientIDEnvKey     = "FITBIT_CLIENT_ID"
	fitbitClientSecretEnvKey = "FITBIT_CLIENT_SECRET"
)

// Slack parameters
const (
	textParam        = "text"
	userIDParam      = "user_id"
	channelIDParam   = "channel_id"
	teamIDParam      = "team_id"
	responseURLParam = "response_url"
)

// Server paths
const (
	updateChallengePath = "UpdateChallenge"
	oauthCallbackPath   = "HandleFitbitAuth"
	linkAccountPath     = "LinkAccount"
	startChallengePath  = "Challenge"
)

// Date formats
const (
	fitbitDateFormat    = "2006-01-02"
	challengeDateFormat = "2006-01-02"
)

const (
	version = "1.0.0"
)

var updateBanners = [...]string{":rolled_up_newspaper: _Breaking news_, here are the current steps ranking",
	":loudspeaker: Oh snap, look who's winning the race!",
	":wind_blowing_face::athletic_shoe: _The more you take, the more you leave behind_...here's the latest steps count update",
	":thinking_face: All truly great thoughts are conceived while walking (and you also get to stay competitive in this challenge)",
	":fairy: Walking is a great adventure...and if you do it enough, you might get the top spot in this list"}

var winnerAccouncementBanners = [...]string{":rolled_up_newspaper: We have a winner for yesterday's steps challenge! :tada:"}

var selectionRandom = rand.New(rand.NewSource(time.Now().Unix()))

// ClientAccess holds the data linking a slack user to their fitbit account
type ClientAccess struct {
	SlackUser string `datastore:"slackUser"`
	SlackTeam string `datastore:"slackTeam"`
	FitbitApiAccess
}

// ChallengeID holds the attributes composing a challenge identifier
type ChallengeID struct {
	ChannelID string `datastore:"channelID"`
	TeamID    string `datastore:"teamID"`
	Date      string `datastore:"date"`
}

// StepsChallenge holds state and definition of a steps challenge
// Datastore seems to preserve the timezone attached to a time.Time value
// and correctly load it back but the documentation is not clear on that part
// and seems to differ in other languages so we keep the timezone
// separately here and are explicit about keeping the timezone information
type StepsChallenge struct {
	ChallengeID
	Active       bool        `datastore:"active"`
	CreatorID    string      `datastore:"createdBy,noindex"`
	CreationTime time.Time   `datastore:"creationTime"`
	TimezoneID   string      `datastore:"timezoneID"`
	RankedUsers  []UserSteps `datastore:"rankedUsers,noindex"`
}

// ActionResponse holds data for a response to a slash command or action
type ActionResponse struct {
	ResponseType    string        `json:"response_type,omitempty"`
	Text            string        `json:"text,omitempty"`
	Blocks          []slack.Block `json:"blocks,omitempty"`
	ReplaceOriginal bool          `json:"replace_original"`
}

// StartFitbitOauthFlow handles an incoming slack request in response to the /roger-link slash command
// and generates a URL for a user to initiate the oauth flow with the Fitbit 3rd party API
func (rc *RogerChallenger) StartFitbitOauthFlow(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rc.verifier.Verify(r.Header, body)
	if err != nil {
		log.Printf("Error validating request: %s", err.Error())
		http.Error(w, err.Error(), 403)
		return
	}

	// Parse the slack payload to get the originating context (the user requesting the account linking)
	params, err := parseSlackRequest(string(body))
	if err != nil {
		log.Printf("Error parsing slack request: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := params[userIDParam]
	responseURL := params[responseURLParam]

	csrfBytesArr := make([]byte, 16)
	cryptorand.Read(csrfBytesArr)
	csrfToken := CsrfToken{Csrf: csrfBytesArr}
	authIDState := AuthIdentificationState{SlackUser: userID, SlackTeam: params[teamIDParam], SlackChannel: params[channelIDParam], ResponseURL: responseURL, CsrfToken: csrfToken}
	oauthState, err := json.Marshal(authIDState)
	if err != nil {
		log.Printf("Error generating AuthIdentificationState: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	csrfKey := NewKeyWithNamespace("CsrfToken", authIDState.SlackTeam, authIDState.SlackUser, nil)
	ctx := context.Background()
	_, err = rc.storer.Put(ctx, csrfKey, &csrfToken)
	if err != nil {
		log.Printf("Error persisting csrf token for user [%s]: %s", authIDState.SlackUser, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURI := fmt.Sprintf("%s/%s", rc.baseURL, oauthCallbackPath)
	oauthLink := fmt.Sprintf("<%s/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=activity&prompt=login_consent&state=%s", rc.fitbitAuthBaseURL, rc.fitbitClientID, url.QueryEscape(redirectURI), base64.URLEncoding.EncodeToString(oauthState))
	oauthFlowMsg := fmt.Sprintf("%s|Head over> to Fitbit to login and authorize access to your account.\n\n"+
		"If you consent, _Roger Challenger_ will use this to get your daily activity summary that will be shared in steps challenges you participate in. "+
		"Note that you'll automatically be included in a steps challenge if you link your Fitbit account and are a "+
		"member of a channel where a steps challenge is active.", oauthLink)
	oauthFlowMessage := ActionResponse{ResponseType: "ephemeral", Text: oauthFlowMsg}
	resp, err := req.Post(responseURL, req.BodyJSON(&oauthFlowMessage))
	if err != nil || resp.Response().StatusCode != 200 {
		if err != nil {
			log.Printf("Error sending message: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			log.Printf("Error writing message: %s", resp.String())
			http.Error(w, resp.String(), http.StatusInternalServerError)
		}

		return
	}
}

// parseSlackRequest parses a slack request parameters. Since slack request parameters have a single value,
// the parsed query parameters are assumed to have a single value as well
func parseSlackRequest(requestBody string) (params map[string]string, err error) {
	queryParams, err := url.ParseQuery(string(requestBody))
	if err != nil {
		return params, errors.Wrapf(err, "Error decoding params from body [%s]", requestBody)
	}

	params = make(map[string]string)
	for name, vals := range queryParams {
		params[name] = vals[0]
	}

	return params, nil
}

// Challenge handles an incoming slack request in response to a user invoking /fitbit-challenge
// This is done by
//   1. Persisting a new challenge if one doesn't already exist for the channel/date
//   2. Announcing the challenge on the channel
//   3. Scheduling a first challenge ranking update
func (rc *RogerChallenger) Challenge(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rc.verifier.Verify(r.Header, body)
	if err != nil {
		log.Printf("Error validating request: %s", err.Error())
		http.Error(w, err.Error(), 403)
		return
	}

	// Parse the slack payload to get the originating context (channel, user)
	params, err := parseSlackRequest(string(body))
	if err != nil {
		log.Printf("Error parsing slack request: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	channel := params[channelIDParam]
	teamID := params[teamIDParam]
	userID := params[userIDParam]
	responseURL := params[responseURLParam]

	timezoneID, location, err := rc.getChannelTimezone(channel)
	if err != nil {
		log.Printf("Error getting channel timezone for channel [%s]: %s", channel, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the initial challenge
	creationTime := time.Now().In(location)
	challengeID := ChallengeID{ChannelID: channel, TeamID: teamID, Date: creationTime.Format(challengeDateFormat)}

	// Check if the challenge exists first and return ephemeral message if it does
	ctx := context.Background()
	k := NewKeyWithNamespace("StepsChallenge", teamID, challengeID.Key(), nil)
	var existingChallenge StepsChallenge
	err = rc.storer.Get(ctx, k, &existingChallenge)
	if err == nil && existingChallenge.Active {
		membershipWarnMsg := ActionResponse{ResponseType: "ephemeral", ReplaceOriginal: false, Text: ":warning: There's already an active steps challenge so you know ¯\\_(ツ)_/¯"}
		resp, err := req.Post(responseURL, req.BodyJSON(&membershipWarnMsg))
		if err != nil || resp.Response().StatusCode != 200 {
			if err != nil {
				log.Printf("Error sending message that challenge already exists: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				log.Printf("Error writing message that challenge already exists: %s", resp.String())
				http.Error(w, resp.String(), http.StatusInternalServerError)
			}

			return
		}

		return
	}

	_, _, err = rc.messenger.PostMessage(channel, slack.MsgOptionText(fmt.Sprintf("<@%s> started a steps challenge! Get moving :wind_blowing_face::athletic_shoe:. If you haven't linked your fitbit account already, type `/roger-link` and join in on the challenge.", userID), false))
	if err != nil {
		// TODO: consider an additional layered fallback strategy where we use https://godoc.org/github.com/nlopes/slack#Client.JoinConversation to try and join (that would work for public channels)
		// before falling back to a message with instructions
		if err.Error() == "channel_not_found" || err.Error() == "not_in_channel" {
			bot, err := rc.userInfoFinder.GetBotInfo("")
			if err != nil {
				log.Printf("Error getting bot info to send membership warning message: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			membershipWarnMsg := ActionResponse{ResponseType: "ephemeral", ReplaceOriginal: false, Text: fmt.Sprintf("I can't start a challenge in a channel or conversation I'm not a member of. Add me, <@%s> and try again :bow:", bot.ID)}
			resp, err := req.Post(responseURL, req.BodyJSON(&membershipWarnMsg))
			if err != nil || resp.Response().StatusCode != 200 {
				if err != nil {
					log.Printf("Error sending app not member warning message: %s", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else {
					log.Printf("Error writing app not member warning message: %s", resp.String())
					http.Error(w, resp.String(), http.StatusInternalServerError)
				}

				return
			}
		} else {
			log.Printf("Error sending message: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	stepsChallenge := StepsChallenge{ChallengeID: challengeID, Active: true, CreatorID: userID, CreationTime: creationTime, TimezoneID: timezoneID}

	_, err = rc.storer.Put(ctx, k, &stepsChallenge)
	if err != nil {
		log.Printf("Error persisting challenge for team [%s] and channel [%s]: %s", teamID, channel, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rc.scheduleChallengeUpdate(challengeID, time.Now())
	if err != nil {
		log.Printf("Error scheduling task: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getChannelTimezone finds what should be the "master" timezone for a channel which informs the scheduling of the updates
func (rc *RogerChallenger) getChannelTimezone(channelID string) (timezoneID string, location *time.Location, err error) {
	// TODO: Replace this hardcoded value by logic to determine the winning channel timezone/location. This could be by
	//   1. Getting the timezone from the slack user info and getting the most common one amongst participating fitbit users
	//   2. If that fails because no one has linked their fitbit account yet, default to the channel creator and get their timezone from their user info
	//   3. If all fails, return "America/Los_Angeles" for now
	timezoneID = "America/Los_Angeles"
	location, err = time.LoadLocation(timezoneID)
	return timezoneID, location, err
}

// refreshChallenge gets updated step summaries from the fitbit API for all the fitbit users
// part of a steps challenge and then renders and sends an updated ranking to the slack channel
func (rc *RogerChallenger) refreshChallenge(stepsChallenge StepsChallenge) (err error) {
	rankedUsers, err := rc.getChallengeRankedSteps(stepsChallenge)
	if err != nil {
		return errors.Wrap(err, "error getting activity summaries")
	}

	// Update the state
	ctx := context.Background()
	stepsChallenge.RankedUsers = rankedUsers
	k := NewKeyWithNamespace("StepsChallenge", stepsChallenge.TeamID, stepsChallenge.ChallengeID.Key(), nil)
	_, err = rc.storer.Put(ctx, k, &stepsChallenge)
	if err != nil {
		return errors.Wrapf(err, "error persisting challenge [%s.%s]", stepsChallenge.TeamID, stepsChallenge.ChallengeID.Key())
	}

	renderBlocks := make([]slack.Block, 0)
	renderedRanking := rc.renderStepsRanking(rankedUsers)
	if len(renderedRanking) > 0 {
		bannerText := updateBanners[selectionRandom.Intn(len(updateBanners))]
		renderBlocks = append(renderBlocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", bannerText, false, false), nil, nil))
		renderBlocks = append(renderBlocks, renderedRanking...)

		_, _, err = rc.messenger.PostMessage(stepsChallenge.ChannelID, slack.MsgOptionText(bannerText, false), slack.MsgOptionBlocks(renderBlocks...))
		if err != nil {
			return errors.Wrap(err, "error sending slack message")
		}
	}

	return nil
}

// localizeCreationTime takes a creation time and a timezone id and localizes that time using the matching time.Location
func localizeCreationTime(creationTime time.Time, timezoneID string) (localized time.Time, location *time.Location, err error) {
	location, err = time.LoadLocation(timezoneID)
	if err != nil {
		return time.Now(), nil, err
	}

	return creationTime.In(location), location, nil
}

// wrapUpChallenge posts the winner of a challenge and marks the challenge as inactive
func (rc *RogerChallenger) wrapUpChallenge(stepsChallenge StepsChallenge) (err error) {
	rankedUsers, err := rc.getChallengeRankedSteps(stepsChallenge)
	if err != nil {
		return errors.Wrap(err, "error getting activity summaries")
	}

	// Update the state with the final ranking and mark the challenge as inactive
	stepsChallenge.RankedUsers = rankedUsers
	stepsChallenge.Active = false
	ctx := context.Background()
	k := NewKeyWithNamespace("StepsChallenge", stepsChallenge.TeamID, stepsChallenge.ChallengeID.Key(), nil)
	_, err = rc.storer.Put(ctx, k, &stepsChallenge)
	if err != nil {
		return errors.Wrapf(err, "Error persisting final challenge [%s.%s]", stepsChallenge.TeamID, stepsChallenge.ChallengeID.Key())
	}

	renderBlocks := make([]slack.Block, 0)
	renderedRanking := rc.renderStepsRanking(rankedUsers)
	if len(renderedRanking) > 0 {
		bannerText := winnerAccouncementBanners[selectionRandom.Intn(len(winnerAccouncementBanners))]
		renderBlocks = append(renderBlocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", bannerText, false, false), nil, nil))
		renderBlocks = append(renderBlocks, renderedRanking...)

		_, _, err = rc.messenger.PostMessage(stepsChallenge.ChannelID, slack.MsgOptionText(bannerText, false), slack.MsgOptionBlocks(renderBlocks...))
		if err != nil {
			return errors.Wrap(err, "error sending slack message")
		}
	}

	return nil
}

// renderStepsRanking renders the user steps ranking as slack blocks to me included in a slack message
func (rc *RogerChallenger) renderStepsRanking(rankedUsers []UserSteps) (renderBlocks []slack.Block) {
	renderBlocks = make([]slack.Block, 0)

	if len(rankedUsers) == 0 {
		return renderBlocks
	}

	// TODO create a worker pool and submit work with parallelism of 4
	rank := 1
	for _, us := range rankedUsers {
		userInfo, err := rc.userInfoFinder.GetUserInfo(us.UserID)
		profileImage := ""
		realName := ""
		if err != nil {
			log.Printf("Error getting user info for [%s]: [%s]", us.UserID, err.Error())
		} else {
			profileImage = userInfo.Profile.Image32
			realName = userInfo.Profile.RealName
		}

		rankingText := fmt.Sprintf("<@%s> `%d` :athletic_shoe:", us.UserID, us.Steps)
		if rank == 1 {
			rankingText = fmt.Sprintf("<@%s> `%d` :athletic_shoe: :tornado::rocket:", us.UserID, us.Steps)
		}

		renderBlocks = append(renderBlocks, slack.NewContextBlock("", slack.NewImageBlockElement(profileImage, realName), slack.NewTextBlockObject("mrkdwn", rankingText, false, false)))
		rank++
	}

	return renderBlocks
}

func (rc *RogerChallenger) Standings(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rc.verifier.Verify(r.Header, body)
	if err != nil {
		log.Printf("Error validating request: %s", err.Error())
		http.Error(w, err.Error(), 403)
		return
	}

	// Parse the slack payload to get the originating context (channel, user)
	params, err := parseSlackRequest(string(body))
	if err != nil {
		log.Printf("Error parsing slack request: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	channel := params[channelIDParam]
	teamID := params[teamIDParam]
	responseURL := params[responseURLParam]

	_, location, err := rc.getChannelTimezone(channel)
	if err != nil {
		log.Printf("Error getting channel timezone for channel [%s]: %s", channel, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the initial challenge
	creationTime := time.Now().In(location)
	challengeID := ChallengeID{ChannelID: channel, TeamID: teamID, Date: creationTime.Format(challengeDateFormat)}

	ctx := context.Background()
	var stepsChallenge StepsChallenge
	k := NewKeyWithNamespace("StepsChallenge", teamID, challengeID.Key(), nil)
	err = rc.storer.Get(ctx, k, &stepsChallenge)
	// If the challenge doesn't exist and a message to the requester and return
	if (err != nil && err == datastore.ErrNoSuchEntity) || (err == nil && !stepsChallenge.Active) {
		noChallengeMsg := ActionResponse{ResponseType: "ephemeral", ReplaceOriginal: false, Text: ":warning: There's no active challenge in this channel to report status on. Create one by using `/roger-challenge`"}
		resp, err := req.Post(responseURL, req.BodyJSON(&noChallengeMsg))
		if err != nil || resp.Response().StatusCode != 200 {
			if err != nil {
				log.Printf("Error sending message that challenge doens't exist: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				log.Printf("Error writing message that challenge doesn't exist: %s", resp.String())
				http.Error(w, resp.String(), http.StatusInternalServerError)
			}

			return
		}

		return
	} else if err != nil {
		log.Printf("Error fetching challenge with id [%s.%s]", teamID, challengeID.Key())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = rc.refreshChallenge(stepsChallenge)
	if err != nil {
		log.Printf("Error refreshing challenge status: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateChallenge handles a request to update a challenge. The requests are usually
// coming from tasks scheduled via scheduleChallengeUpdate
func (rc *RogerChallenger) UpdateChallenge(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var challengeID ChallengeID
	err = json.Unmarshal([]byte(body), &challengeID)
	if err != nil {
		log.Printf("Error decoding challenge id from body: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the full existing StepsChallenge
	ctx := context.Background()
	var stepsChallenge StepsChallenge
	k := NewKeyWithNamespace("StepsChallenge", challengeID.TeamID, challengeID.Key(), nil)
	err = rc.storer.Get(ctx, k, &stepsChallenge)
	// If the challenge doesn't exist, return and don't shedule a next update
	if err != nil && err == datastore.ErrNoSuchEntity {
		log.Printf("Challenge not found id [%s.%s]", challengeID.TeamID, challengeID.Key())
		return
	} else if err != nil {
		log.Printf("Error loading existing challenge [%s.%s]: %s", challengeID.TeamID, challengeID.Key(), err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	localizedCreationTime, location, err := localizeCreationTime(stepsChallenge.CreationTime, stepsChallenge.TimezoneID)
	if err != nil {
		log.Printf("Error localizing challenge creation time for challenge [%s.%s]: %s", challengeID.TeamID, challengeID.Key(), err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	endScheduledDayUpdates := getLastDayUpdateTime(localizedCreationTime, location)

	// The final update time is the next day at 8am. We start by setting the time right and then adding one day
	finalChannelUpdateTime := getFinalUpdateTime(localizedCreationTime, location)

	switch now := time.Now(); {
	// We're still in day time during the day of the challenge so we keep posting updates and scheduling hourly refreshes
	case !now.After(endScheduledDayUpdates):
		scheduledUpdate := now.Add(time.Duration(1) + time.Hour)
		log.Printf("Challenge [%s.%s] scheduled for a regular update at [%s]", stepsChallenge.TeamID, stepsChallenge.ChallengeID.Key(), scheduledUpdate)
		err = rc.scheduleChallengeUpdate(challengeID, scheduledUpdate)
		if err != nil {
			log.Printf("Error scheduling next challenge update: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = rc.refreshChallenge(stepsChallenge)
		if err != nil {
			log.Printf("Error refreshing challenge status: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	// We're after the end of day updates before the final update for the winner. Create the task to issue that final update
	case now.After(endScheduledDayUpdates) && now.Before(finalChannelUpdateTime):
		log.Printf("Challenge [%s.%s] scheduled for a final update at [%s]", stepsChallenge.TeamID, stepsChallenge.ChallengeID.Key(), finalChannelUpdateTime)

		err = rc.scheduleChallengeUpdate(challengeID, finalChannelUpdateTime)
		if err != nil {
			log.Printf("Error scheduling next challenge update: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	// We're on or after the scheduled final update time so we mark the challenge as inactive after posting the winnner
	case !now.Before(finalChannelUpdateTime):
		log.Printf("Wrapping up challenge [%s.%s], no more updates scheduled", stepsChallenge.TeamID, stepsChallenge.ChallengeID.Key())
		rc.wrapUpChallenge(stepsChallenge)
	}
}

// Key returns the formatted key for a ChallengeID. Not that this is stricly the actual key value which excludes the namespace
// which is set independently of the key. Logically speaking, the coordinates to a challenge and the ChallengeID is represented
// both in the namespace and this generated Key
func (id ChallengeID) Key() string {
	return fmt.Sprintf("%s:%s", id.ChannelID, id.Date)
}

// getLastDayUpdateTime returns the local time of the last update for the day (the last one before we stop sending notifications because people might be sleeping)
func getLastDayUpdateTime(creationTime time.Time, location *time.Location) (lastDayUpdateTime time.Time) {
	creationYear, creationMonth, creationDay := creationTime.Date()
	return time.Date(creationYear, creationMonth, creationDay, 19, 0, 0, 0, location)
}

// getFinalUpdateTime returns the final challenge update that comes the morning the day after and announces the winner
func getFinalUpdateTime(creationTime time.Time, location *time.Location) (finalUpdateTime time.Time) {
	creationYear, creationMonth, creationDay := creationTime.Date()
	finalUpdateTime = time.Date(creationYear, creationMonth, creationDay, 8, 0, 0, 0, location)
	finalUpdateTime = finalUpdateTime.AddDate(0, 0, 1)

	return
}
