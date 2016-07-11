package service

import (
	"crypto/tls"
	"net/http"

	"github.com/opsee/basic/grpcutil"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/basic/tp"
	"github.com/opsee/cats/store"
	log "github.com/opsee/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type service struct {
	checkStore store.CheckStore
}

func New(pgConn string) (*service, error) {
	db, err := store.NewPostgres(pgConn)
	if err != nil {
		return nil, err
	}

	svc := &service{
		checkStore: store.NewCheckStore(db),
	}

	return svc, nil
}

// http / grpc multiplexer for http health checks
func (s *service) StartMux(addr, certfile, certkeyfile string) error {
	router := tp.NewHTTPRouter(context.Background())
	server := grpc.NewServer()

	opsee.RegisterCatsServer(server, s)
	log.Infof("starting cats service at %s", addr)

	httpServer := &http.Server{
		Addr:      addr,
		Handler:   grpcutil.GRPCHandlerFunc(server, router),
		TLSConfig: &tls.Config{},
	}

	return httpServer.ListenAndServeTLS(certfile, certkeyfile)
}
