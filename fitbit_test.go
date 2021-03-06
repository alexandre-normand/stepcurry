package stepcurry

import (
	"cloud.google.com/go/datastore"
	"encoding/base64"
	"encoding/json"
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

func TestHandleFitbitAuthCallbackUrlParsing(t *testing.T) {
	tests := map[string]struct {
		callbackURL    string
		expectedResult int
		expectedError  string
		expectedMsg    string
	}{
		"MissingCodeParam": {
			callbackURL:    "/" + oauthCallbackPath,
			expectedResult: http.StatusBadRequest,
			expectedError:  "Missing authorization code",
			expectedMsg:    "",
		},
		"MissingStateParam": {
			callbackURL:    "/" + oauthCallbackPath + "?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4",
			expectedResult: http.StatusBadRequest,
			expectedError:  "Missing Auth Identification State",
			expectedMsg:    "",
		},
		"InvalidBase64InStateParam": {
			callbackURL:    "/" + oauthCallbackPath + "?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=@Invalid$",
			expectedResult: http.StatusBadRequest,
			expectedError:  "illegal base64 data at input byte 0",
			expectedMsg:    "Error base64 decoding slack Auth Identification State",
		},
		"InvalidStateParamPayload": {
			callbackURL:    "/" + oauthCallbackPath + "?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=dGVzdGFzZHNhZHNh",
			expectedResult: http.StatusBadRequest,
			expectedError:  "invalid character 'e' in literal true (expecting 'r')",
			expectedMsg:    "Error decoding Auth Identification State json",
		},
	}

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.callbackURL, strings.NewReader(""))
			w := httptest.NewRecorder()
			err = sc.HandleFitbitAuth(w, r)

			require.Error(t, err)
			require.IsType(t, new(httpError), err)

			herr := err.(*httpError)
			assert.EqualError(t, herr.err, tc.expectedError)
			assert.Equal(t, tc.expectedMsg, herr.message)
			assert.Equal(t, tc.expectedResult, herr.code)
		})
	}
}

func TestHandleFitbitAuthCallbackWithErrorLoadingCsrfToken(t *testing.T) {
	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: "ignored", CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(datastore.ErrInvalidEntityType)
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	err = sc.HandleFitbitAuth(w, r)

	require.Error(t, err)
	require.IsType(t, new(httpError), err)

	herr := err.(*httpError)
	assert.EqualError(t, herr.err, "datastore: invalid entity type")
	assert.Equal(t, "Error fetching csrf token for user [UCODE]", herr.message)
	assert.Equal(t, http.StatusInternalServerError, herr.code)
}

func TestHandleFitbitAuthCallbackWithMissingCsrfToken(t *testing.T) {
	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: "ignored", CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(datastore.ErrNoSuchEntity)
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	err = sc.HandleFitbitAuth(w, r)

	require.Error(t, err)
	require.IsType(t, new(httpError), err)

	herr := err.(*httpError)
	assert.EqualError(t, herr.err, "datastore: no such entity")
	assert.Equal(t, "CSRF token not found", herr.message)
	assert.Equal(t, http.StatusUnauthorized, herr.code)
}

func TestHandleFitbitAuthCallbackWithUnexpectedCsrfToken(t *testing.T) {
	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: "ignored", CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	// This is taking a shortcut and returning a success without trying to modify the value struct like the real datastore would
	// This works because we're looking for a mismatch and the empty value won't match
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil)
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	err = sc.HandleFitbitAuth(w, r)

	require.Error(t, err)
	require.IsType(t, new(httpError), err)

	herr := err.(*httpError)
	assert.EqualError(t, herr.err, "CSRF token mismatch")
	assert.Equal(t, "", herr.message)
	assert.Equal(t, http.StatusUnauthorized, herr.code)
}

func TestHandleFitbitAuthCallbackWithErrorDeletingToken(t *testing.T) {
	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: "ignored", CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		returnVal := args.Get(2).(*CsrfToken)
		returnVal.Csrf = []byte("csrf")
	})
	storer.On("Delete", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(fmt.Errorf("backend error"))
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	err = sc.HandleFitbitAuth(w, r)

	require.Error(t, err)
	require.IsType(t, new(httpError), err)

	herr := err.(*httpError)
	assert.EqualError(t, herr.err, "backend error")
	assert.Equal(t, "Error deleting up csrf token for user [UCODE]", herr.message)
	assert.Equal(t, http.StatusInternalServerError, herr.code)
}

func TestHandleFitbitAuthCallbackWithErrorExchangingCodeForToken(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "come back again later", http.StatusServiceUnavailable)
	})

	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: "ignored", CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		returnVal := args.Get(2).(*CsrfToken)
		returnVal.Csrf = []byte("csrf")
	})
	storer.On("Delete", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil)
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionFitbitURLs(server.URL, server.URL), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	err = sc.HandleFitbitAuth(w, r)

	require.Error(t, err)
	require.IsType(t, new(httpError), err)

	herr := err.(*httpError)
	assert.EqualError(t, herr.err, "error getting access token [503 Service Unavailable]: come back again later\n")
	assert.Equal(t, "Error getting fitbit api access for user [UCODE]", herr.message)
	assert.Equal(t, http.StatusInternalServerError, herr.code)
}

func TestHandleFitbitAuthCallbackWithInvalidTokenResponse(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: "ignored", CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		returnVal := args.Get(2).(*CsrfToken)
		returnVal.Csrf = []byte("csrf")
	})
	storer.On("Delete", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil)
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionFitbitURLs(server.URL, server.URL), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	err = sc.HandleFitbitAuth(w, r)

	require.Error(t, err)
	require.IsType(t, new(httpError), err)

	herr := err.(*httpError)
	assert.EqualError(t, herr.err, "error decoding api access response: unexpected end of JSON input")
	assert.Equal(t, "Error getting fitbit api access for user [UCODE]", herr.message)
	assert.Equal(t, http.StatusInternalServerError, herr.code)
}

func TestHandleFitbitAuthCallbackWithErrorPersistingClientAccess(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		apiAccess := FitbitApiAccess{Token: "token", FitbitUser: "1020", RefreshToken: "refresh"}
		body, _ := json.Marshal(apiAccess)
		w.Write(body)
	})

	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: "ignored", CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		returnVal := args.Get(2).(*CsrfToken)
		returnVal.Csrf = []byte("csrf")
	})
	storer.On("Delete", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil)
	storer.On("Put", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "" && k.Name == "1020" && k.Parent == nil && k.Kind == "FitbitApiAccess"
	}), &FitbitApiAccess{Token: "token", FitbitUser: "1020", RefreshToken: "refresh"}).Return(nil, fmt.Errorf("backend error"))
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionFitbitURLs(server.URL, server.URL), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	err = sc.HandleFitbitAuth(w, r)

	require.Error(t, err)
	require.IsType(t, new(httpError), err)

	herr := err.(*httpError)
	assert.EqualError(t, herr.err, "backend error")
	assert.Equal(t, "Error persisting fitbit api access for fitbit user [1020]", herr.message)
	assert.Equal(t, http.StatusInternalServerError, herr.code)
}

func TestHandleFitbitAuthCallbackWithErrorSendingResultMessage(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		apiAccess := FitbitApiAccess{Token: "token", FitbitUser: "1020", RefreshToken: "refresh"}
		body, _ := json.Marshal(apiAccess)
		w.Write(body)
	})

	mux.HandleFunc("/slackURL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("timeout"))
	})

	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: fmt.Sprintf("%s/slackURL", server.URL), CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		returnVal := args.Get(2).(*CsrfToken)
		returnVal.Csrf = []byte("csrf")
	})
	storer.On("Delete", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil)
	storer.On("Put", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "" && k.Name == "1020" && k.Parent == nil && k.Kind == "FitbitApiAccess"
	}), &FitbitApiAccess{Token: "token", FitbitUser: "1020", RefreshToken: "refresh"}).Return(nil, nil)
	storer.On("Put", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "ClientAccess"
	}), &ClientAccess{SlackUser: "UCODE", FitbitUser: "1020", SlackTeam: "TSOMETHING"}).Return(nil, nil)
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionFitbitURLs(server.URL, server.URL), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	err = sc.HandleFitbitAuth(w, r)

	require.Error(t, err)
	require.IsType(t, new(httpError), err)

	herr := err.(*httpError)
	assert.EqualError(t, herr.err, "Error writing oauth completion message: timeout")
	assert.Equal(t, "", herr.message)
	assert.Equal(t, http.StatusInternalServerError, herr.code)
}

func TestHandleFitbitAuthCallback(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		apiAccess := FitbitApiAccess{Token: "token", FitbitUser: "1020", RefreshToken: "refresh"}
		body, _ := json.Marshal(apiAccess)
		w.Write(body)
	})

	mux.HandleFunc("/slackURL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	stateValue := authIDStateToQueryParam(AuthIdentificationState{SlackUser: "UCODE", SlackChannel: "CGEN", SlackTeam: "TSOMETHING", ResponseURL: fmt.Sprintf("%s/slackURL", server.URL), CsrfToken: CsrfToken{Csrf: []byte("csrf")}})
	callbackURL := fmt.Sprintf("/%s?code=46f595a20e4cd85ce6abf6487eacdaaaf0ecf1c4&state=%s", oauthCallbackPath, stateValue)
	r := httptest.NewRequest(http.MethodGet, callbackURL, strings.NewReader(""))

	w := httptest.NewRecorder()

	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	storer.On("Get", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		returnVal := args.Get(2).(*CsrfToken)
		returnVal.Csrf = []byte("csrf")
	})
	storer.On("Delete", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "CsrfToken"
	}), mock.Anything).Return(nil)
	storer.On("Put", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "" && k.Name == "1020" && k.Parent == nil && k.Kind == "FitbitApiAccess"
	}), &FitbitApiAccess{Token: "token", FitbitUser: "1020", RefreshToken: "refresh"}).Return(nil, nil)
	storer.On("Put", mock.Anything, mock.MatchedBy(func(k *datastore.Key) bool {
		return k.Namespace == "TSOMETHING" && k.Name == "UCODE" && k.Parent == nil && k.Kind == "ClientAccess"
	}), &ClientAccess{SlackUser: "UCODE", SlackTeam: "TSOMETHING", FitbitUser: "1020"}).Return(nil, nil)

	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://localhost", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionFitbitURLs(server.URL, server.URL), OptionVerifier(verifier), OptionStorer(storer), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)
	sc.HandleFitbitAuth(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "<html><head><meta http-equiv=\"refresh\" content=\"0;URL=slack://channel?team=TSOMETHING&id=CGEN\"></head></html>", string(rbody))
}

func authIDStateToQueryParam(authIDState AuthIdentificationState) (queryParam string) {
	json, _ := json.Marshal(authIDState)
	queryParam = base64.URLEncoding.EncodeToString(json)

	return queryParam
}
