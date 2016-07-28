package service

import (
	"os"
	"testing"
	"time"

	"github.com/opsee/basic/schema"
	"github.com/opsee/cats/checks"
	"github.com/opsee/cats/store"
	"github.com/opsee/cats/testutil"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
)

type testSluiceClient struct{}

func (s *testSluiceClient) Send(name string, data interface{}) error {
	return nil
}

type testTeamStore struct {
	curTeam *schema.Team
}

func (q *testTeamStore) WithTX(txfun func(store.TeamStore) error) error {
	return txfun(q)
}

func (q *testTeamStore) Get(id string) (*schema.Team, error) {
	return q.curTeam, nil
}

func (q *testTeamStore) GetByStripeId(id string) (*schema.Team, error) {
	return q.curTeam, nil
}

func (q *testTeamStore) GetUsers(id string) ([]*schema.User, error) {
	var users []*schema.User
	for _, u := range testutil.Users {
		if u.Active {
			users = append(users, &u)
		}
	}
	return users, nil
}

func (q *testTeamStore) GetInvites(id string) ([]*schema.User, error) {
	var users []*schema.User
	return users, nil
}

func (q *testTeamStore) Create(team *schema.Team) error {
	team.Id = "666"
	q.curTeam = team
	return nil
}

func (q *testTeamStore) Update(team *schema.Team) error {
	q.curTeam = team
	return nil
}

func (q *testTeamStore) UpdateSubscription(team *schema.Team) error {
	q.curTeam = team
	return nil
}

func (q *testTeamStore) Delete(team *schema.Team) error {
	q.curTeam = team
	return nil
}

func (q *testTeamStore) List(page, perPage int) ([]*schema.Team, store.ListMeta, error) {
	return []*schema.Team{q.curTeam}, store.ListMeta{Page: page, PerPage: perPage, Total: uint64(1)}, nil
}

type testCheckStore struct{}

func (q *testCheckStore) GetAndLockState(customerId, checkId string) (*checks.State, error) {
	return nil, nil
}
func (q *testCheckStore) UpdateState(state *checks.State) error { return nil }
func (q *testCheckStore) PutState(state *checks.State) error    { return nil }
func (q *testCheckStore) PutMemo(memo *checks.ResultMemo) error { return nil }
func (q *testCheckStore) GetMemo(checkId, bastionId string) (*checks.ResultMemo, error) {
	return nil, nil
}
func (q *testCheckStore) CreateStateTransitionLogEntry(checkId, customerId string, fromState, toState checks.StateId) (*checks.StateTransitionLogEntry, error) {
	return nil, nil
}
func (q *testCheckStore) GetLiveBastions(customerId, checkId string) ([]string, error) {
	return []string{}, nil
}
func (q *testCheckStore) GetCheckStateTransitionLogEntries(checkId, customerId string, from, to time.Time) ([]*checks.StateTransitionLogEntry, error) {
	return nil, nil
}
func (q *testCheckStore) GetCheck(user *schema.User, checkId string) (*schema.Check, error) {
	return nil, nil
}
func (q *testCheckStore) GetChecks(user *schema.User) ([]*schema.Check, error) { return nil, nil }
func (q *testCheckStore) GetCheckCount(customerId string) (int32, error)       { return int32(2), nil }

func TestMain(m *testing.M) {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	// initialize stripe with our account key
	stripe.Key = viper.GetString("stripe_key")

	os.Exit(m.Run())
}
