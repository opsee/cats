package store

import (
	"github.com/opsee/basic/schema"
	"github.com/opsee/cats/checks"
)

type CheckStore interface {
	GetAndLockState(customerId, checkId string) (*checks.State, error)
	UpdateState(state *checks.State) error
	PutState(state *checks.State) error
	PutMemo(memo *checks.ResultMemo) error
	GetMemo(checkId, bastionId string) (*checks.ResultMemo, error)
	CreateStateTransitionLogEntry(checkId, customerId string, fromState, toState checks.StateId) (*checks.StateTransitionLogEntry, error)
	GetCheckCount(user *schema.User, prorated bool) (float32, error)
	GetLiveBastions(customerId, checkId string) ([]string, error)
}
