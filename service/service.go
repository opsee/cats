package service

import (
	"errors"

	"github.com/opsee/cats/store"
)

type service struct {
	db store.Store
}

func NewService(connect string) (*service, error) {
	svc := new(service)
	db, err := store.NewPostgres(connect)
	if err != nil {
		return nil, err
	}

	svc.db = db

	return svc, nil
}

type CheckAssertions struct {
	CheckID    string             `json:"check-id"`
	Assertions []*store.Assertion `json:"assertions"`
}

func (c *CheckAssertions) Validate() error {
	if c.CheckID == "" {
		return errors.New("CheckAssertions must have valid check-id")
	}

	if len(c.Assertions) < 1 {
		return errors.New("Use DELETE method to delete assertions.")
	}

	return nil
}

type GetChecksResponse struct {
	Items []*CheckAssertions `json:"items"`
}
