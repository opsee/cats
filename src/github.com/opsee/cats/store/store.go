package store

import (
	"github.com/opsee/basic/com"
)

type Store interface {
	GetAssertions(*com.User, string) ([]*Assertion, error)
	PutAssertions(*com.User, string, []*Assertion) error
	DeleteAssertions(*com.User, string) error
}
