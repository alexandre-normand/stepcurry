package rogerchallenger

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"net/http"
)

// RogerChallenger holds state and dependencies for a server instance
type RogerChallenger struct {
	baseURL            string
	fitbitAuthBaseURL  string
	fitbitAPIBaseURL   string
	fitbitClientID     string
	fitbitClientSecret string
	storer             Datastorer
	verifier           Verifier
	userInfoFinder     UserInfoFinder
	messenger          Messenger
	channelInfoFinder  ChannelInfoFinder
	taskScheduler      TaskScheduler
}

// Option is a function that applies an option to a RogerChallenger instance
type Option func(rc *RogerChallenger) (err error)

// Verifier is implemented by any value that has the Verify method
type Verifier interface {
	Verify(header http.Header, body []byte) (err error)
}

// SlackVerifier represents a slack verifier backed by github.com/nlopes/slack
type SlackVerifier struct {
	slackSigningSecret string
}

// OptionVerifier sets a verifier as the implementation on RogerChallenger
func OptionVerifier(verifier Verifier) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.verifier = verifier
		return nil
	}
}

// OptionSlackVerifier sets a nlopes/slack.Client as the implementation of Verifier
func OptionSlackVerifier(slackSigningSecret string) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.verifier = &SlackVerifier{slackSigningSecret: slackSigningSecret}

		return nil
	}
}

// UserInfoFinder defines the interface for finding self and other users' info
type UserInfoFinder interface {
	// GetBotInfo fetches bot info. See https://godoc.org/github.com/nlopes/slack#Client.GetBotInfo for more details
	GetBotInfo(botID string) (bot *slack.Bot, err error)
	// GetUserInfo fetches a user info by user id. See https://godoc.org/github.com/nlopes/slack#Client.GetUserInfo for more details.
	GetUserInfo(userID string) (userInfo *slack.User, err error)
}

// OptionUserInfoFinder sets a userInfoFinder as the implementation for a RogerChallenger instance
func OptionUserInfoFinder(userInfoFinder UserInfoFinder) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.userInfoFinder = userInfoFinder
		return nil
	}
}

// OptionStorer sets a storer as the implementation on RogerChallenger
func OptionStorer(storer Datastorer) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.storer = storer
		return nil
	}
}

// Messenger defines the interface for sending messages
type Messenger interface {
	// PostMessage sends a message using the web api. See https://godoc.org/github.com/nlopes/slack#Client.PostMessage for more details
	PostMessage(channelID string, options ...slack.MsgOption) (channel string, timestamp string, err error)
}

// OptionMessenger sets a messenger as the implementation on RogerChallenger
func OptionMessenger(messenger Messenger) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.messenger = messenger
		return nil
	}
}

// ChannelInfoFinder defines the interface for finding channel info
type ChannelInfoFinder interface {
	// GetChannelInfo fetches info on a channel. See https://godoc.org/github.com/nlopes/slack#Client.GetChannelInfo for more details
	GetChannelInfo(channelID string) (channel *slack.Channel, err error)
}

// OptionChannelInfoFinder sets a channelInfoFinder as the implementation on RogerChallenger
func OptionChannelInfoFinder(channelInfoFinder ChannelInfoFinder) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.channelInfoFinder = channelInfoFinder
		return nil
	}
}

// OptionTaskScheduler sets a taskScheduler as the implementation on RogerChallenger
func OptionTaskScheduler(taskScheduler TaskScheduler) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.taskScheduler = taskScheduler
		return nil
	}
}

// Verify verifies the slack request's authenticity (https://api.slack.com/docs/verifying-requests-from-slack). If the request
// can't be verified or if it fails verification, an error is returned. For a verified valid request, nil is returned
func (v SlackVerifier) Verify(header http.Header, body []byte) (err error) {
	verifier, err := slack.NewSecretsVerifier(header, v.slackSigningSecret)
	if err != nil {
		return errors.Wrap(err, "Error creating slack secrets verifier")
	}

	_, err = verifier.Write(body)
	if err != nil {
		return errors.Wrap(err, "Error writing request body to slack verifier")
	}

	err = verifier.Ensure()
	if err != nil {
		return err
	}

	return nil
}

// OptionFitbitURLs overrides the Fitbit base URLs
func OptionFitbitURLs(fitbitAuthBaseURL string, fitbitAPIBaseUrl string) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.fitbitAuthBaseURL = fitbitAuthBaseURL
		rc.fitbitAPIBaseURL = fitbitAPIBaseUrl
		return nil
	}
}

// New creates a new instance of RogerChallenger with a baseURL, fitbit client id and secret as well as all of its required
// dependencies via Option
func New(baseURL string, fitbitClientID string, fitbitClientSecret string, opts ...Option) (rc *RogerChallenger, err error) {
	rc = new(RogerChallenger)

	rc.baseURL = baseURL
	rc.fitbitAuthBaseURL = defaultFitbitAuthBaseURL
	rc.fitbitAPIBaseURL = defaultFitbitAPIBaseURL
	rc.fitbitClientID = fitbitClientID
	rc.fitbitClientSecret = fitbitClientSecret

	for _, apply := range opts {
		err := apply(rc)
		if err != nil {
			return nil, err
		}
	}

	if rc.storer == nil {
		return nil, fmt.Errorf("storer is nil after applying all Options. Did you forget to set one?")
	}

	if rc.verifier == nil {
		return nil, fmt.Errorf("verifier is nil after applying all Options. Did you forget to set one?")
	}

	if rc.messenger == nil {
		return nil, fmt.Errorf("messenger is nil after applying all Options. Did you forget to set one?")
	}

	if rc.userInfoFinder == nil {
		return nil, fmt.Errorf("userInfoFinder is nil after applying all Options. Did you forget to set one?")
	}

	if rc.channelInfoFinder == nil {
		return nil, fmt.Errorf("channelInfoFinder is nil after applying all Options. Did you forget to set one?")
	}

	if rc.taskScheduler == nil {
		return nil, fmt.Errorf("taskScheduler is nil after applying all Options. Did you forget to set one?")
	}

	return rc, nil
}
