package service

import (
	"errors"
	"golang.org/x/net/context"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
)

const (
	serviceKey = iota
	userKey
	requestKey
	paramsKey
)

func (s *service) StartHTTP(addr string) error {
	rtr := s.router()
	return http.ListenAndServe(addr, rtr)
}

func (s *service) router() *tp.Router {
	rtr := tp.NewHTTPRouter(context.Background())

	rtr.CORS(
		[]string{"GET", "POST", "DELETE", "HEAD", "PUT"},
		[]string{`https?://localhost:8080`, `https://(\w+\.)?opsee\.com`},
	)

	rtr.Handle("GET", "/api/swagger.json", []tp.DecodeFunc{}, s.swagger())

	// /assertions
	// Retrieve all of a customer assertions -- returns { "items": []*CheckAssertions }
	rtr.Handle("GET", "/assertions", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getAssertions())
	// Update assertions for a check. Accepts *CheckAssertions, returns *CheckAssertions
	rtr.Handle("POST", "/assertions", decoders(com.User{}, CheckAssertions{}), s.putAssertions())

	// /assertions/:check_id
	// Retrieve CheckAssertions for a single check, returns *CheckAssertions
	rtr.Handle("GET", "/assertions/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.getAssertions())
	// Delete assertions for a single check. Accept nil body, returns nil
	rtr.Handle("DELETE", "/assertions/:check_id", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{}), tp.ParamsDecoder(paramsKey)}, s.deleteAssertions())
	// Update assertions for a single check, Accept *CheckAssertions, return *CheckAssertions
	rtr.Handle("PUT", "/assertions/:check_id", decoders(com.User{}, CheckAssertions{}), s.putAssertions())

	rtr.Timeout(5 * time.Minute)

	return rtr
}

func (s *service) swagger() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		return swaggerMap, http.StatusOK, nil
	}
}

func decoders(userType interface{}, requestType interface{}) []tp.DecodeFunc {
	return []tp.DecodeFunc{
		tp.AuthorizationDecodeFunc(userKey, userType),
		tp.RequestDecodeFunc(requestKey, requestType),
		tp.ParamsDecoder(paramsKey),
	}
}

func (s *service) getAssertions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkID string

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkID = params.ByName("check_id")
		}

		assertions, err := s.db.GetAssertions(user, checkID)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
		}

		if len(assertions) == 0 {
			return nil, http.StatusNotFound, nil
		}

		assmap := make(map[string]*CheckAssertions)
		for _, assertion := range assertions {
			if assmap[assertion.CheckID] == nil {
				assmap[assertion.CheckID] = &CheckAssertions{CheckID: assertion.CheckID}
			}

			ca := assmap[assertion.CheckID]
			ca.Assertions = append(ca.Assertions, assertion)
		}

		resp := &GetChecksResponse{}
		for _, v := range assmap {
			resp.Items = append(resp.Items, v)
		}

		return resp, http.StatusOK, nil
	}
}

func (s *service) putAssertions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkID string

		req, ok := ctx.Value(requestKey).(*CheckAssertions)
		if !ok {
			return ctx, http.StatusInternalServerError, errors.New("Unable to process request.")
		} else {
			checkID = req.CheckID
		}

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkID = params.ByName("check_id")
		}

		for _, ass := range req.Assertions {
			ass.CheckID = checkID
			ass.CustomerID = user.CustomerID
		}

		err := s.db.PutAssertions(user, checkID, req.Assertions)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
		}

		resp := &CheckAssertions{
			CheckID:    checkID,
			Assertions: req.Assertions,
		}

		return resp, http.StatusOK, nil
	}
}

func (s *service) deleteAssertions() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		var checkID string

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errors.New("Unable to get user from request context.")
		}

		params, ok := ctx.Value(paramsKey).(httprouter.Params)
		if ok && params.ByName("check_id") != "" {
			checkID = params.ByName("check_id")
		}

		if checkID == "" {
			return ctx, http.StatusBadRequest, errors.New("Must specify check-id in request.")
		}

		err := s.db.DeleteAssertions(user, checkID)
		if err != nil {
			return ctx, http.StatusInternalServerError, err
		}

		return nil, http.StatusOK, nil
	}
}
