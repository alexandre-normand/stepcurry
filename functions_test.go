package stepcurry

import (
	"cloud.google.com/go/datastore"
	"fmt"
	"github.com/alexandre-normand/stepcurry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStartFitbitOauthFlowInvalidSlackSignature(t *testing.T) {
	body := fmt.Sprintf("token=sometoken&team_id=TEAMID&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%%2Fpoll&text=%%22To%%20do%%20or%%20not%%20to%%20do%%3F%%22%%20%%22Do%%22%%20%%22Not%%20Do%%22&response_url=%s&trigger_id=someTriggerID", "https://slack.com")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	w := httptest.NewRecorder()

	storer := &mocks.Datastorer{}
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	channelInfoFinder := &mocks.ChannelInfoFinder{}
	defer channelInfoFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, channelInfoFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionSlackVerifier("1e13414e22545115a2c62c3b8cd67dfe"), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	sc.StartFitbitOauthFlow(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "Error creating slack secrets verifier: timestamp is too old\n", string(rbody))
}

func TestStartFitbitOauthFlowInvalidSlackRequest(t *testing.T) {
	body := "%gh&%ij"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	channelInfoFinder := &mocks.ChannelInfoFinder{}
	defer channelInfoFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, channelInfoFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	sc.StartFitbitOauthFlow(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Error decoding params from body [%gh&%ij]: invalid URL escape \"%gh\"\n", string(rbody))
}

func TestStartFitbitOauthFlowErrorSavingCsrfToken(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=frans&user_name=frans&command=%2FlinkAccount&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Put", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TEAMID" && k.Name == "frans" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil, fmt.Errorf("failed to persist"))
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	channelInfoFinder := &mocks.ChannelInfoFinder{}
	defer channelInfoFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, channelInfoFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	sc.StartFitbitOauthFlow(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "failed to persist\n", string(rbody))
}

func TestStartFitbitOauthFlowErrorSendingSlackMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("wut"))
	}))
	defer server.Close()

	body := fmt.Sprintf("token=sometoken&team_id=TEAMID&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=frans&user_name=frans&command=%%2FlinkAccount&text=&response_url=%s&trigger_id=someTriggerID", server.URL)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Put", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TEAMID" && k.Name == "frans" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(NewKeyWithNamespace("CsrfToken", "TEAMID", "frans", nil), nil)
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	channelInfoFinder := &mocks.ChannelInfoFinder{}
	defer channelInfoFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, channelInfoFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	sc.StartFitbitOauthFlow(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "wut\n", string(rbody))
}

func TestStartFitbitOauthFlow(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	body := fmt.Sprintf("token=sometoken&team_id=TEAMID&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=frans&user_name=frans&command=%%2FlinkAccount&text=&response_url=%s&trigger_id=someTriggerID", server.URL)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Put", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TEAMID" && k.Name == "frans" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(NewKeyWithNamespace("CsrfToken", "TEAMID", "frans", nil), nil)
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	channelInfoFinder := &mocks.ChannelInfoFinder{}
	defer channelInfoFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, channelInfoFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	sc.StartFitbitOauthFlow(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "", string(rbody))
	assert.Contains(t, slackRequest, "https://www.fitbit.com/oauth2/authorize?response_type=code\\u0026client_id=fitbitClientID\\u0026redirect_uri=https%3A%2F%2Flocalhost%2FHandleFitbitAuth\\u0026scope=activity\\u0026prompt=login_consent\\u0026state=")
}
