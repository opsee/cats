package service

import (
	"errors"
	"golang.org/x/net/context"
	"net/http"
	"time"

	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/cats/store"
)

const (
	serviceKey = iota
	userKey
	requestKey
	paramsKey
)

func (s *service) StartHTTP(addr string) {
	rtr := s.router()
	http.ListenAndServe(addr, rtr)
}

func (s *service) router() *tp.Router {
	rtr := tp.NewHTTPRouter(context.Background())

	rtr.CORS(
		[]string{"GET", "POST", "DELETE", "HEAD"},
		[]string{`https?://localhost:8080`, `https://(\w+\.)?opsee\.com`},
	)

	//rtr.Handle("GET", "/api/swagger.json", []tp.DecodeFunc{}, s.swagger())

	// The request types may seem nonintuitive, but eventually cats will handle checks and assertions will
	// be included in checks--leading to the eventual deprecation of the /assertions endpoint. So instead of
	// sending only assertions here, we will allow the request body to include the entire check for now and
	// we will get the assertions from it.
	rtr.Handle("GET", "/assertions", decoders(com.User{}, GetChecksRequest{}), s.getAssertions())
	rtr.Handle("PUT", "/assertions", decoders(com.User{}, PutCheckRequest{}), s.putAssertions())
	rtr.Handle("DELETE", "/assertions", decoders(com.User{}, PutCheckRequest{}), s.deleteAssertions())

	rtr.Timeout(5 * time.Minute)

	return rtr
}

func decoders(userType interface{}, requestType interface{}) []tp.DecodeFunc {
	return []tp.DecodeFunc{
		tp.AuthorizationDecodeFunc(userKey, userType),
		tp.RequestDecodeFunc(requestKey, requestType),
	}
}

func (s *service) getAssertions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		request, ok := ctx.Value(requestKey).(*GetChecksRequest)
		if !ok {
			return ctx, http.StatusInternalServerError, errors.New("Unable to process request.")
		}

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		assertions, err := s.db.GetAssertions(user, request.Checks)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
		}

		if len(assertions) == 0 {
			return nil, http.StatusNotFound, nil
		}
		return assertions, http.StatusOK, nil
	}
}

func (s *service) putAssertions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		req, ok := ctx.Value(requestKey).(*PutCheckRequest)
		if !ok {
			return ctx, http.StatusInternalServerError, errors.New("Unable to process request.")
		}

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		var asses []*store.Assertion
		for _, ass := range req.Check.Assertions {
			storeAss := store.NewAssertion(user.CustomerID, req.Check.Id, ass)
			asses = append(asses, storeAss)
		}

		err := s.db.PutAssertions(user, req.Check.Id, asses)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
		}
		return req.Check, http.StatusOK, nil
	}
}

func (s *service) deleteAssertions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		req, ok := ctx.Value(requestKey).(*PutCheckRequest)
		if !ok {
			return ctx, http.StatusInternalServerError, errors.New("Unable to process request.")
		}

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		err := s.db.DeleteAssertions(user, req.Check.Id)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
		}

		return req.Check, http.StatusOK, nil
	}
}
