package rogerchallenger

import (
	"context"
	"fmt"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"net/http"
	"os"
)

// RogerChallenger holds state and dependencies for a server instance
type RogerChallenger struct {
	baseURL            string
	fitbitAuthBaseURL  string
	fitbitAPIBaseURL   string
	slackBaseURL       string
	slackClientID      string
	slackClientSecret  string
	fitbitClientID     string
	fitbitClientSecret string
	storer             Datastorer
	verifier           Verifier
	taskScheduler      TaskScheduler
	TeamRouter
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

// UserInfoFinder defines the interface for getting users' info
type UserInfoFinder interface {
	// GetUserInfo fetches a user info by user id. See https://godoc.org/github.com/nlopes/slack#Client.GetUserInfo for more details.
	GetUserInfo(userID string) (userInfo *slack.User, err error)
}

// // OptionUserInfoFinder sets a userInfoFinder as the implementation for a RogerChallenger instance
// func OptionUserInfoFinder(userInfoFinder UserInfoFinder) Option {
// 	return func(rc *RogerChallenger) (err error) {
// 		rc.userInfoFinder = userInfoFinder
// 		return nil
// 	}
// }

// BotIdentificator defines the interface for getting the bot's self id
type BotIdentificator interface {
	// GetBotID returns the bot user ID
	GetBotID() (botUserID string, err error)
}

// FixedBotIdentificator holds a fixed bot user ID for cases where the bot user ID is known
// and doesn't required interaction with the slack web API
type FixedBotIdentificator struct {
	botUserID string
}

func (fbi FixedBotIdentificator) GetBotID() (botUserID string, err error) {
	return fbi.botUserID, nil
}

// SlackAPIBotIdentificator holds a reference to a slack client to get the bot's self user ID using
// the GetBotInfo method
type SlackAPIBotIdentificator struct {
	slackClient *slack.Client
}

func (sabi *SlackAPIBotIdentificator) GetBotID() (botUserID string, err error) {
	bot, err := sabi.slackClient.GetBotInfo("")
	if err != nil {
		return "", err
	}

	return bot.ID, nil
}

// OptionBotIdentificator sets a botIdentificator as the implementation for a RogerChallenger instance
// func OptionBotIdentificator(botIdentificator BotIdentificator) Option {
// 	return func(rc *RogerChallenger) (err error) {
// 		rc.botIdentificator = botIdentificator
// 		return nil
// 	}
// }

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
// func OptionMessenger(messenger Messenger) Option {
// 	return func(rc *RogerChallenger) (err error) {
// 		rc.messenger = messenger
// 		return nil
// 	}
// }

// ChannelInfoFinder defines the interface for finding channel info
type ChannelInfoFinder interface {
	// GetChannelInfo fetches info on a channel. See https://godoc.org/github.com/nlopes/slack#Client.GetChannelInfo for more details
	GetChannelInfo(channelID string) (channel *slack.Channel, err error)
}

// OptionChannelInfoFinder sets a channelInfoFinder as the implementation on RogerChallenger
// func OptionChannelInfoFinder(channelInfoFinder ChannelInfoFinder) Option {
// 	return func(rc *RogerChallenger) (err error) {
// 		rc.channelInfoFinder = channelInfoFinder
// 		return nil
// 	}
// }

// OptionTaskScheduler sets a taskScheduler as the implementation on RogerChallenger
func OptionTaskScheduler(taskScheduler TaskScheduler) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.taskScheduler = taskScheduler
		return nil
	}
}

// OptionTeamRouter sets a teamRouter as the implementation on RogerChallenger
func OptionTeamRouter(teamRouter TeamRouter) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.TeamRouter = teamRouter
		return nil
	}
}

// TeamServices holds references to tenanted services
type TeamServices struct {
	userInfoFinder    UserInfoFinder
	botIdentificator  BotIdentificator
	messenger         Messenger
	channelInfoFinder ChannelInfoFinder
}

// TeamRouter defines the interface for routing to various tenanted services on team ID
type TeamRouter interface {
	Route(teamID string) (svcs TeamServices, err error)
}

type SingleTenantRouter struct {
	services TeamServices
}

func (stRouter *SingleTenantRouter) Route(teamID string) (svcs TeamServices, err error) {
	return stRouter.services, nil
}

func NewSingleTenantRouter(userInfoFinder UserInfoFinder, botIdentificator BotIdentificator, messenger Messenger, channelInfoFinder ChannelInfoFinder) (stRouter *SingleTenantRouter, err error) {
	stRouter = new(SingleTenantRouter)
	stRouter.services = TeamServices{userInfoFinder: userInfoFinder, botIdentificator: botIdentificator, messenger: messenger, channelInfoFinder: channelInfoFinder}

	return stRouter, nil
}

type MultiTenantRouter struct {
	projectID   string
	storer      Datastorer
	tokenLoader TokenLoader
	tokenSaver  TokenSaver
	svcsByTeam  map[string]TeamServices
}

func (mtRouter *MultiTenantRouter) Route(teamID string) (svcs TeamServices, err error) {
	if svcs, ok := mtRouter.svcsByTeam[teamID]; !ok {
		token, err := mtRouter.tokenLoader.Load(mtRouter.projectID, teamID)

		if err != nil {
			return svcs, errors.Wrapf(err, "team [%s] not found", teamID)
		}

		ctx := context.Background()
		var botInfo BotInfo
		k := NewKeyWithNamespace("BotInfo", teamID, "Bot", nil)
		err = rc.storer.Get(ctx, k, &botInfo)
		if err != nil {
			return svcs, errors.Wrapf(err, "Error loading bot info [%s] for team [%s]", botInfo.UserID, teamID)
		}

		slackClient := slack.New(token, slack.OptionDebug(cast.ToBool(os.Getenv(debugEnv))))

		teamSvcs := TeamServices{userInfoFinder: slackClient, botIdentificator: FixedBotIdentificator{botUserID: botInfo.UserID}, messenger: slackClient, channelInfoFinder: slackClient}
		mtRouter.svcsByTeam[teamID] = teamSvcs
	}

	return mtRouter.svcsByTeam[teamID], nil
}

func NewMultiTenantRouter(projectID string, storer Datastorer, tokenLoader TokenLoader, tokenSaver TokenSaver) (mtRouter *MultiTenantRouter, err error) {
	mtRouter = new(MultiTenantRouter)
	mtRouter.projectID = projectID
	mtRouter.storer = storer
	mtRouter.tokenSaver = tokenSaver
	mtRouter.tokenLoader = tokenLoader
	mtRouter.svcsByTeam = make(map[string]TeamServices)

	return mtRouter, nil
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

// OptionSlackBaseURL overrides the Slack base URL
func OptionSlackBaseURL(slackBaseURL string) Option {
	return func(rc *RogerChallenger) (err error) {
		rc.slackBaseURL = slackBaseURL
		return nil
	}
}

// New creates a new instance of RogerChallenger with a baseURL, fitbit client id and secret as well as all of its required
// dependencies via Option
func New(baseURL string, fitbitClientID string, fitbitClientSecret string, slackClientID string, slackClientSecret string, opts ...Option) (rc *RogerChallenger, err error) {
	rc = new(RogerChallenger)

	rc.baseURL = baseURL
	rc.fitbitAuthBaseURL = defaultFitbitAuthBaseURL
	rc.fitbitAPIBaseURL = defaultFitbitAPIBaseURL
	rc.slackBaseURL = defaultSlackBaseURL
	rc.slackClientID = slackClientID
	rc.slackClientSecret = slackClientSecret
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

	if rc.TeamRouter == nil {
		return nil, fmt.Errorf("teamRouter is nil after applying all Options. Did you forget to set one?")
	}

	if rc.taskScheduler == nil {
		return nil, fmt.Errorf("taskScheduler is nil after applying all Options. Did you forget to set one?")
	}

	return rc, nil
}
