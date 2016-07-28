package service

import (
	"crypto/tls"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/opsee/basic/grpcutil"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/basic/tp"
	"github.com/opsee/cats/checks/results"
	"github.com/opsee/cats/store"
	sluice "github.com/opsee/gmunch/client"
	log "github.com/opsee/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type service struct {
	httpServer   *http.Server
	checkStore   store.CheckStore
	teamStore    store.TeamStore
	resultStore  results.Store
	sluiceClient sluice.Client
}

func New(pgConn string, resultStore results.Store) (*service, error) {
	db, err := store.NewPostgres(pgConn)
	if err != nil {
		return nil, err
	}

	sluiceClient, err := sluice.New("sluice.in.opsee.com:8443", sluice.Config{
		TLSConfig: tls.Config{},
	})
	if err != nil {
		return nil, err
	}

	svc := &service{
		checkStore:   store.NewCheckStore(db),
		teamStore:    store.NewTeamStore(db),
		resultStore:  resultStore,
		sluiceClient: sluiceClient,
	}

	return svc, nil
}

// http / grpc multiplexer for http health checks
func (s *service) StartMux(addr, certfile, certkeyfile string) error {
	// The grpc service
	server := grpc.NewServer()
	opsee.RegisterCatsServer(server, s)
	log.Infof("starting cats service at %s", addr)

	s.httpServer = &http.Server{
		Addr:      addr,
		Handler:   grpcutil.GRPCHandlerFunc(server, s.NewHandler()),
		TLSConfig: &tls.Config{},
	}

	return s.httpServer.ListenAndServeTLS(certfile, certkeyfile)
}

// Returns a new service http.Handler for testing. Sets up an opsee/tp router, which comes with a basic health endpoint for free.
// It's going to have at least another endpoint for stripe webhooks
func (s *service) NewHandler() http.Handler {
	router := tp.NewHTTPRouter(context.Background())
	router.Handle("POST", "/hooks/stripe", []tp.DecodeFunc{s.httpLogger(), s.stripeHookDecoder()}, s.stripeHookHandler())
	return router
}

// HTTP decoder that logs requests
func (s *service) httpLogger() tp.DecodeFunc {
	return func(ctx context.Context, rw http.ResponseWriter, r *http.Request, p httprouter.Params) (context.Context, int, error) {
		log.Info("http request: ", r.URL.String())
		return ctx, 0, nil
	}
}
