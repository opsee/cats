package service

import (
	"testing"

	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/store"
	"github.com/opsee/cats/testutil"
	"github.com/stretchr/testify/assert"
	// "github.com/opsee/cats/subscriptions"
	"golang.org/x/net/context"
)

type testTeamStore struct {
	curTeam *schema.Team
}

func (q *testTeamStore) WithTX(txfun func(store.TeamStore) error) error {
	return txfun(q)
}

func (q *testTeamStore) Get(id string) (*schema.Team, error) {
	return testutil.Teams["active"], nil
}

func (q *testTeamStore) GetUsers(id string) ([]*schema.User, error) {
	var users []*schema.User
	for _, u := range testutil.Users {
		if u.Active {
			users = append(users, u)
		}
	}
	return users, nil
}

func (q *testTeamStore) GetInvites(id string) ([]*schema.User, error) {
	var users []*schema.User
	return users, nil
}

func (q *testTeamStore) Create(team *schema.Team) error {
	q.curTeam = team
	return nil
}

func (q *testTeamStore) Update(team *schema.Team) error {
	q.curTeam = team
	return nil
}

func (q *testTeamStore) Delete(team *schema.Team) error {
	q.curTeam = team
	return nil
}

func TestUpdateTeam(t *testing.T) {
	assert := assert.New(t)
	ts := &testTeamStore{}
	s := &service{
		teamStore: ts,
	}

	team := new(schema.Team)
	*team = *testutil.Teams["active"]
	assert.Equal("beta", team.SubscriptionPlan)

	// change subscription to team plan, increase quantity
	team.SubscriptionPlan = "team_monthly"
	team.SubscriptionQuantity = 5

	res, err := s.UpdateTeam(context.Background(), &opsee.UpdateTeamRequest{
		Requestor: &schema.User{},
		Team:      team,
	})
	assert.NoError(err)
	assert.NotNil(res.Team)
	assert.Equal("team_monthly", res.Team.SubscriptionPlan)
	assert.EqualValues(5, res.Team.SubscriptionQuantity)
}
