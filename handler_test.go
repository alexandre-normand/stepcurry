package stepcurry

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServing(t *testing.T) {
	tests := map[string]struct {
		handler        Handler
		expectedResult int
		expectedBody   string
	}{
		"NoError": {
			handler: func(w http.ResponseWriter, r *http.Request) error {
				w.Write([]byte("success"))
				return nil
			},
			expectedResult: http.StatusOK,
			expectedBody:   "success",
		},
		"StringError": {
			handler: func(http.ResponseWriter, *http.Request) error {
				return errors.New("something wrong")
			},
			expectedResult: http.StatusInternalServerError,
			expectedBody:   "something wrong\n",
		},
		"HttpErrorWithoutMessage": {
			handler: func(http.ResponseWriter, *http.Request) error {
				return newHttpError(errors.New("some error"), "", http.StatusBadGateway)
			},
			expectedResult: http.StatusBadGateway,
			expectedBody:   "some error\n",
		},
		"HttpErrorWithMessage": {
			handler: func(http.ResponseWriter, *http.Request) error {
				return newHttpError(errors.New("some invalid state"), "Bad internal state for user [x]", http.StatusInternalServerError)
			},
			expectedResult: http.StatusInternalServerError,
			expectedBody:   "some invalid state\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
			w := httptest.NewRecorder()
			Handler(tc.handler).ServeHTTP(w, r)

			resp := w.Result()
			rbody, _ := ioutil.ReadAll(resp.Body)

			assert.Equal(t, tc.expectedResult, resp.StatusCode)
			assert.Equal(t, tc.expectedBody, string(rbody))
		})
	}
}
