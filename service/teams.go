package service

import (
	"fmt"

	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/store"
	"github.com/opsee/cats/subscriptions"
	"golang.org/x/net/context"
)

// Fetches team, including users and invites.
func (s *service) GetTeam(ctx context.Context, req *opsee.GetTeamRequest) (*opsee.GetTeamResponse, error) {
	if req.Team == nil {
		return nil, fmt.Errorf("invalid request, missing team")
	}

	t, err := s.teamStore.Get(req.Team.Id)
	if err != nil {
		return nil, err
	}

	return &opsee.GetTeamResponse{
		Team: t,
	}, nil
}

// Creates a team. This call will also make a stripe API request to create a stripe customer / subscription,
// and update the stripe subscription and customer id in the database.
func (s *service) CreateTeam(ctx context.Context, req *opsee.CreateTeamRequest) (*opsee.CreateTeamResponse, error) {
	team := req.Team

	if team == nil {
		return nil, fmt.Errorf("invalid request, missing team")
	}

	if err := team.Validate(); err != nil {
		return nil, err
	}

	if req.Requestor == nil {
		return nil, fmt.Errorf("invalid request, missing requestor")
	}

	if req.Requestor.Email == "" {
		return nil, fmt.Errorf("invalid request, missing requestor email")
	}

	if team.SubscriptionPlan == "" {
		team.SubscriptionPlan = "beta"
	}

	if err := s.teamStore.Create(team); err != nil {
		return nil, err
	}

	checkCount, err := s.checkStore.GetCheckCount(team.Id)
	if err != nil {
		return nil, err
	}

	team.SubscriptionQuantity = checkCount

	if err := subscriptions.Create(team, req.Requestor.Email, req.StripeToken, req.TrialEnd); err != nil {
		return nil, err
	}

	// update with stripe info
	if err := s.teamStore.UpdateSubscription(team); err != nil {
		return nil, err
	}

	return &opsee.CreateTeamResponse{
		Team: team,
	}, nil
}

// Updates team name or subscription. This call will make a stripe API request inline if
// the team subscription or subscription quantity is changed, or if a stripe credit card
// token is present in the request.
func (s *service) UpdateTeam(ctx context.Context, req *opsee.UpdateTeamRequest) (*opsee.UpdateTeamResponse, error) {
	if req.Team == nil {
		return nil, fmt.Errorf("invalid request, missing team")
	}

	if err := req.Team.Validate(); err != nil {
		return nil, err
	}

	var (
		currentTeam *schema.Team
		err         error
	)

	if err := s.teamStore.WithTX(func(ts store.TeamStore) error {
		currentTeam, err = ts.Get(req.Team.Id)
		if err != nil {
			return err
		}

		if currentTeam == nil {
			return fmt.Errorf("no team found")
		}

		checkCount, err := s.checkStore.GetCheckCount(req.Team.Id)
		if err != nil {
			return err
		}

		req.Team.SubscriptionQuantity = checkCount

		// update subscription if necessary
		if currentTeam.StripeSubscriptionId != "" && (currentTeam.SubscriptionPlan != req.Team.SubscriptionPlan || currentTeam.SubscriptionQuantity != req.Team.SubscriptionQuantity || req.StripeToken != "") {
			currentTeam.SubscriptionPlan = req.Team.SubscriptionPlan
			currentTeam.SubscriptionQuantity = req.Team.SubscriptionQuantity

			err = subscriptions.Update(currentTeam, req.StripeToken)
			if err != nil {
				return err
			}

			err = ts.UpdateSubscription(currentTeam)
			if err != nil {
				return err
			}
		}

		// update other pertinent fields in the team (name i guess)
		currentTeam.Name = req.Team.Name

		return ts.Update(currentTeam)

	}); err != nil {
		return nil, err
	}

	return &opsee.UpdateTeamResponse{
		Team: currentTeam,
	}, nil
}

// Soft deletes a team and cancels their stripe subscription if present.
func (s *service) DeleteTeam(ctx context.Context, req *opsee.DeleteTeamRequest) (*opsee.DeleteTeamResponse, error) {
	if req.Team == nil {
		return nil, fmt.Errorf("invalid request, missing team")
	}

	if req.Team.Id == "" {
		return nil, fmt.Errorf("invalid request, missing team id")
	}

	currentTeam, err := s.teamStore.Get(req.Team.Id)
	if err != nil {
		return nil, err
	}

	err = subscriptions.Cancel(currentTeam)
	if err != nil {
		return nil, err
	}

	err = s.teamStore.Delete(currentTeam)
	if err != nil {
		return nil, err
	}

	return &opsee.DeleteTeamResponse{
		Team: currentTeam,
	}, nil
}
