package service

import (
	"crypto/tls"
	"fmt"
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
	db store.Store
}

func New(pgConn string) (*service, error) {
	svc := new(service)
	db, err := store.NewPostgres(pgConn)
	if err != nil {
		return nil, err
	}

	svc.db = db
	return svc, nil
}

// http / grpc multiplexer for http health checks
func (s *service) StartMux(addr, certfile, certkeyfile string) error {
	router := tp.NewHTTPRouter(context.Background())
	server := grpc.NewServer()

	opsee.RegisterCatsServer(server, s)

	httpServer := &http.Server{
		Addr:      addr,
		Handler:   grpcutil.GRPCHandlerFunc(server, router),
		TLSConfig: &tls.Config{},
	}

	return httpServer.ListenAndServeTLS(certfile, certkeyfile)
}

func (s *service) GetCheckCount(ctx context.Context, req *opsee.GetCheckCountRequest) (*opsee.GetCheckCountResponse, error) {
	if req.User == nil {
		log.Error("no user in request")
		return nil, fmt.Errorf("user is required")
	}

	if err := req.User.Validate(); err != nil {
		log.WithError(err).Error("user is invalid")
		return nil, err
	}

	count, err := s.db.GetCheckCount(req.User, req.Prorated)
	if err != nil {
		log.WithError(err).Error("db request failed")
		return nil, err
	}

	return &opsee.GetCheckCountResponse{
		Prorated: req.Prorated,
		Count:    count,
	}, nil
}
