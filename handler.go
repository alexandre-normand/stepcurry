package stepcurry

import (
	"log"
	"net/http"
)

// httpError holds an application error and attributes to return a proper
// http response to the caller
type httpError struct {
	err     error
	message string
	code    int
}

func newHttpError(err error, message string, code int) (herr *httpError) {
	herr = new(httpError)
	herr.err = err
	herr.message = message
	herr.code = code

	return herr
}

func (e *httpError) Error() string {
	return e.err.Error()
}

// Handler is a http.HandlerFunc that returns an error
type Handler func(http.ResponseWriter, *http.Request) error

// ServerHTTP runs the requestHandler and returns an error with the appropriate code if an error is returned
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		if herr, ok := err.(*httpError); ok {
			if len(herr.message) > 0 {
				log.Printf("%s: %s", herr.message, herr.Error())
			} else {
				log.Printf("%s", herr.Error())
			}

			http.Error(w, herr.Error(), herr.code)
		} else {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
