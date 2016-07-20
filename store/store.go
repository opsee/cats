package store

import (
	"time"

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
	GetCheckStateTransitionLogEntries(checkId, customerId string, from, to time.Time) ([]*checks.StateTransitionLogEntry, error)
	GetCheck(user *schema.User, checkId string) (*schema.Check, error)
	GetChecks(user *schema.User) ([]*schema.Check, error)
}

type TeamStore interface {
	WithTX(txfun func(TeamStore) error) error
	Get(id string) (*schema.Team, error)
	GetUsers(id string) ([]*schema.User, error)
	GetInvites(id string) ([]*schema.User, error)
	Create(team *schema.Team) error
	Update(team *schema.Team) error
	Delete(team *schema.Team) error
}
