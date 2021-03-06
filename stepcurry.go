package stepcurry

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	otel "go.opentelemetry.io/otel/metric/global"
	"net/http"
)

const (
	appName = "step-curry"
)

// StepCurry holds state and dependencies for a server instance
type StepCurry struct {
	baseURL            string
	fitbitAuthBaseURL  string
	fitbitAPIBaseURL   string
	slackBaseURL       string
	slackAppID         string
	slackClientID      string
	slackClientSecret  string
	fitbitClientID     string
	fitbitClientSecret string
	debug              bool
	storer             Datastorer
	verifier           Verifier
	taskScheduler      TaskScheduler
	paths              Paths
	slashCommands      SlashCommands
	meter              metric.Meter
	instruments        *instruments
	TeamRouter
}

// Paths holds the paths to the http handlers
type Paths struct {
	UpdateChallenge    string
	FitbitAuthCallback string
	LinkAccount        string
	StartChallenge     string
	Standings          string
}

// SlashCommands holds the names of the app's slash commands
type SlashCommands struct {
	Link      string
	Challenge string
	Standings string
}

// instruments holds general application metrics
type instruments struct {
	challengeCount            metric.BoundInt64Counter
	accountLinkInitiatedCount metric.BoundInt64Counter
	accountLinkCompletedCount metric.BoundInt64Counter
	updateCount               metric.BoundInt64Counter
	totalStepsRecorded        metric.BoundInt64Counter
	challengeParticipants     metric.BoundInt64ValueRecorder
	challengeSteps            metric.BoundInt64ValueRecorder
	challengeWinnerStepCount  metric.BoundInt64ValueRecorder
}

// Option is a function that applies an option to a StepCurry instance
type Option func(sc *StepCurry) (err error)

// Verifier is implemented by any value that has the Verify method
type Verifier interface {
	Verify(header http.Header, body []byte) (err error)
}

// SlackVerifier represents a slack verifier backed by github.com/slack-go/slack
type SlackVerifier struct {
	slackSigningSecret string
}

// OptionVerifier sets a verifier as the implementation on StepCurry
func OptionVerifier(verifier Verifier) Option {
	return func(sc *StepCurry) (err error) {
		sc.verifier = verifier
		return nil
	}
}

// OptionSlackVerifier sets a slack-go/slack.Client as the implementation of Verifier
func OptionSlackVerifier(slackSigningSecret string) Option {
	return func(sc *StepCurry) (err error) {
		sc.verifier = &SlackVerifier{slackSigningSecret: slackSigningSecret}

		return nil
	}
}

// UserInfoFinder defines the interface for getting users' info
type UserInfoFinder interface {
	// GetUserInfo fetches a user info by user id. See https://godoc.org/github.com/slack-go/slack#Client.GetUserInfo for more details.
	GetUserInfo(userID string) (userInfo *slack.User, err error)
}

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

func NewSlackAPIBotIdentificator(slackClient *slack.Client) (sabi *SlackAPIBotIdentificator) {
	sabi = new(SlackAPIBotIdentificator)
	sabi.slackClient = slackClient

	return sabi
}

// OptionStorer sets a storer as the implementation on StepCurry
func OptionStorer(storer Datastorer) Option {
	return func(sc *StepCurry) (err error) {
		sc.storer = storer
		return nil
	}
}

// Messenger defines the interface for sending messages
type Messenger interface {
	// PostMessage sends a message using the web api. See https://godoc.org/github.com/slack-go/slack#Client.PostMessage for more details
	PostMessage(channelID string, options ...slack.MsgOption) (channel string, timestamp string, err error)
}

// ConversationMemberFinder defines the interface for finding members on a conversation
type ConversationMemberFinder interface {
	// GetUsersInConversation fetches members of a conversation. See https://godoc.org/github.com/slack-go/slack#Client.GetUsersInConversation for more details
	GetUsersInConversation(params *slack.GetUsersInConversationParameters) (members []string, cursor string, err error)
}

// OptionTaskScheduler sets a taskScheduler as the implementation on StepCurry
func OptionTaskScheduler(taskScheduler TaskScheduler) Option {
	return func(sc *StepCurry) (err error) {
		sc.taskScheduler = taskScheduler
		return nil
	}
}

// OptionTeamRouter sets a teamRouter as the implementation on StepCurry
func OptionTeamRouter(teamRouter TeamRouter) Option {
	return func(sc *StepCurry) (err error) {
		sc.TeamRouter = teamRouter
		return nil
	}
}

// TeamServices holds references to tenanted services
type TeamServices struct {
	userInfoFinder           UserInfoFinder
	botIdentificator         BotIdentificator
	messenger                Messenger
	conversationMemberFinder ConversationMemberFinder
}

// TeamRouter defines the interface for routing to various tenanted services on team ID
type TeamRouter interface {
	Route(teamID string) (svcs TeamServices, err error)
	TokenSaver
	TokenLoader
}

type SingleTenantRouter struct {
	services TeamServices
	TokenSaver
	TokenLoader
}

func (stRouter *SingleTenantRouter) Route(teamID string) (svcs TeamServices, err error) {
	return stRouter.services, nil
}

func NewSingleTenantRouter(userInfoFinder UserInfoFinder, botIdentificator BotIdentificator, messenger Messenger, conversationMemberFinder ConversationMemberFinder) (stRouter *SingleTenantRouter, err error) {
	stRouter = new(SingleTenantRouter)
	meter := otel.GetMeterProvider().Meter("github.com/alexandre-normand/stepcurry")
	stRouter.services = TeamServices{userInfoFinder: userInfoFinder, botIdentificator: botIdentificator, messenger: NewMessengerWithTelemetry(messenger, appName, meter), conversationMemberFinder: conversationMemberFinder}

	return stRouter, nil
}

type MultiTenantRouter struct {
	debug      bool
	projectID  string
	storer     Datastorer
	svcsByTeam map[string]TeamServices
	TokenLoader
	TokenSaver
}

func (mtRouter *MultiTenantRouter) Route(teamID string) (svcs TeamServices, err error) {
	if svcs, ok := mtRouter.svcsByTeam[teamID]; !ok {
		token, err := mtRouter.LoadToken(teamID)

		if err != nil {
			return svcs, errors.Wrapf(err, "team [%s] not found", teamID)
		}

		ctx := context.Background()
		var botInfo BotInfo
		k := NewKeyWithNamespace("BotInfo", teamID, "Bot", nil)
		err = mtRouter.storer.Get(ctx, k, &botInfo)
		if err != nil {
			return svcs, errors.Wrapf(err, "Error loading bot info [%s] for team [%s]", botInfo.UserID, teamID)
		}

		slackClient := slack.New(token, slack.OptionDebug(mtRouter.debug))
		meter := otel.GetMeterProvider().Meter("github.com/alexandre-normand/stepcurry")
		teamSvcs := TeamServices{userInfoFinder: slackClient, botIdentificator: FixedBotIdentificator{botUserID: botInfo.UserID}, messenger: NewMessengerWithTelemetry(slackClient, appName, meter), conversationMemberFinder: slackClient}
		mtRouter.svcsByTeam[teamID] = teamSvcs
	}

	return mtRouter.svcsByTeam[teamID], nil
}

func NewMultiTenantRouter(projectID string, storer Datastorer, tokenLoader TokenLoader, tokenSaver TokenSaver, debug bool) (mtRouter *MultiTenantRouter, err error) {
	mtRouter = new(MultiTenantRouter)
	mtRouter.projectID = projectID
	mtRouter.storer = storer
	mtRouter.TokenSaver = tokenSaver
	mtRouter.TokenLoader = tokenLoader
	mtRouter.svcsByTeam = make(map[string]TeamServices)
	mtRouter.debug = debug

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
	return func(sc *StepCurry) (err error) {
		sc.fitbitAuthBaseURL = fitbitAuthBaseURL
		sc.fitbitAPIBaseURL = fitbitAPIBaseUrl
		return nil
	}
}

// OptionSlackBaseURL overrides the Slack base URL
func OptionSlackBaseURL(slackBaseURL string) Option {
	return func(sc *StepCurry) (err error) {
		sc.slackBaseURL = slackBaseURL
		return nil
	}
}

// OptionSlashCommands overrides the Slack command names
func OptionSlashCommands(slashCommands SlashCommands) Option {
	return func(sc *StepCurry) (err error) {
		sc.slashCommands = slashCommands
		return nil
	}
}

// OptionPaths overrides the http handler names
func OptionPaths(paths Paths) Option {
	return func(sc *StepCurry) (err error) {
		sc.paths = paths
		return nil
	}
}

// New creates a new instance of StepCurry with a baseURL, fitbit client id and secret as well as all of its required
// dependencies via Option
func New(baseURL string, slackAppID string, fitbitClientID string, fitbitClientSecret string, slackClientID string, slackClientSecret string, opts ...Option) (sc *StepCurry, err error) {
	sc = new(StepCurry)

	sc.baseURL = baseURL
	sc.slackAppID = slackAppID
	sc.fitbitAuthBaseURL = defaultFitbitAuthBaseURL
	sc.fitbitAPIBaseURL = defaultFitbitAPIBaseURL
	sc.slackBaseURL = defaultSlackBaseURL
	sc.slashCommands = SlashCommands{Link: commandLinkFitbit, Challenge: commandChallenge, Standings: commandStandings}
	sc.paths = Paths{UpdateChallenge: updateChallengePath, FitbitAuthCallback: oauthCallbackPath, LinkAccount: linkAccountPath, StartChallenge: startChallengePath, Standings: standingsPath}
	sc.slackClientID = slackClientID
	sc.slackClientSecret = slackClientSecret
	sc.fitbitClientID = fitbitClientID
	sc.fitbitClientSecret = fitbitClientSecret

	for _, apply := range opts {
		err := apply(sc)
		if err != nil {
			return nil, err
		}
	}

	if sc.storer == nil {
		return nil, fmt.Errorf("storer is nil after applying all Options. Did you forget to set one?")
	}

	if sc.verifier == nil {
		return nil, fmt.Errorf("verifier is nil after applying all Options. Did you forget to set one?")
	}

	if sc.TeamRouter == nil {
		return nil, fmt.Errorf("teamRouter is nil after applying all Options. Did you forget to set one?")
	}

	if sc.taskScheduler == nil {
		return nil, fmt.Errorf("taskScheduler is nil after applying all Options. Did you forget to set one?")
	}

	sc.meter = otel.GetMeterProvider().Meter("github.com/alexandre-normand/stepcurry")
	sc.instruments = newInstruments(sc.meter)

	sc.verifier = NewVerifierWithTelemetry(sc.verifier, appName, sc.meter)
	sc.storer = NewDatastorerWithTelemetry(sc.storer, appName, sc.meter)
	sc.taskScheduler = NewTaskSchedulerWithTelemetry(sc.taskScheduler, appName, sc.meter)

	return sc, nil
}

// newInstruments creates a new set of general application metrics
func newInstruments(meter metric.Meter) *instruments {
	defaultLabels := label.String("name", appName)
	mt := metric.Must(meter)

	challengeCounter := mt.NewInt64Counter("challengeCount")
	accountLinkInitiatedCounter := mt.NewInt64Counter("accountLinkInitiatedCount")
	accountLinkCompletedCounter := mt.NewInt64Counter("accountLinkCompletedCount")
	updateCounter := mt.NewInt64Counter("updateCount")
	totalStepsRecordedCounter := mt.NewInt64Counter("totalStepsRecorded")
	challengeParticipantsMeasure := mt.NewInt64ValueRecorder("challengeParticipants")
	challengeStepsMeasure := mt.NewInt64ValueRecorder("challengeSteps")
	challengeWinnerStepCountMeasure := mt.NewInt64ValueRecorder("challengeWinningStepCount")

	return &instruments{
		challengeCount:            challengeCounter.Bind(defaultLabels),
		accountLinkInitiatedCount: accountLinkInitiatedCounter.Bind(defaultLabels),
		accountLinkCompletedCount: accountLinkCompletedCounter.Bind(defaultLabels),
		updateCount:               updateCounter.Bind(defaultLabels),
		totalStepsRecorded:        totalStepsRecordedCounter.Bind(defaultLabels),
		challengeParticipants:     challengeParticipantsMeasure.Bind(defaultLabels),
		challengeSteps:            challengeStepsMeasure.Bind(defaultLabels),
		challengeWinnerStepCount:  challengeWinnerStepCountMeasure.Bind(defaultLabels),
	}
}
