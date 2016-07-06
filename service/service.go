package service

import (
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	grpcauth "google.golang.org/grpc/credentials"
	"net"
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

func (s *service) Start(listenAddr, cert, certkey string) error {
	auth, err := grpcauth.NewServerTLSFromFile(cert, certkey)
	if err != nil {
		return err
	}

	server := grpc.NewServer(grpc.Creds(auth))
	opsee.RegisterCatsServer(server, s)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	return server.Serve(lis)
}
