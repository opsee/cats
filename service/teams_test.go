package service

import (
	"testing"
	"time"

	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/testutil"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestCreateUpdateDeleteTeam(t *testing.T) {
	assert := assert.New(t)

	curTeam := testutil.Teams["active"]
	ts := &testTeamStore{&curTeam}
	cs := &testCheckStore{}
	s := &service{
		teamStore:  ts,
		checkStore: cs,
	}

	team := &schema.Team{
		Name:             "http://www.customink.com/team/bowling-team-names",
		SubscriptionPlan: "beta",
	}

	resp, err := s.CreateTeam(context.Background(), &opsee.CreateTeamRequest{
		Requestor: &schema.User{
			Email: "testin@opsee.com",
		},
		Team:     team,
		TrialEnd: time.Now().Add(3 * 24 * time.Hour).Unix(),
	})
	assert.NoError(err)
	assert.Equal("beta", resp.Team.SubscriptionPlan)
	assert.EqualValues(2, resp.Team.SubscriptionQuantity)
	assert.Equal("trialing", resp.Team.SubscriptionStatus)
	assert.NotEmpty(resp.Team.StripeSubscriptionId)
	assert.NotEmpty(resp.Team.StripeCustomerId)
	assert.NotEmpty(resp.Team.Id)

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
	assert.EqualValues(2, res.Team.SubscriptionQuantity)

	_, err = s.DeleteTeam(context.Background(), &opsee.DeleteTeamRequest{
		Team: team,
	})
	assert.NoError(err)
}
