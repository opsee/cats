package service

import (
	"fmt"

	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/subscriptions"
	"golang.org/x/net/context"
)

// Fetches team, including users
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

// Updates team name or subscription
func (s *service) UpdateTeam(ctx context.Context, req *opsee.UpdateTeamRequest) (*opsee.UpdateTeamResponse, error) {
	if req.Team == nil {
		return nil, fmt.Errorf("invalid request, missing team")
	}

	if err := req.Team.Validate(); err != nil {
		return nil, err
	}

	currentTeam, err := s.teamStore.Get(req.Team.Id)
	if err != nil {
		return nil, err
	}

	if currentTeam == nil {
		return nil, fmt.Errorf("no team found")
	}

	// update subscription if necessary
	if currentTeam.Subscription != req.Team.Subscription || currentTeam.SubscriptionQuantity != req.Team.SubscriptionQuantity || req.StripeToken != "" {
		currentTeam.Subscription = req.Team.Subscription
		currentTeam.SubscriptionQuantity = req.Team.SubscriptionQuantity

		err = subscriptions.Update(currentTeam, req.StripeToken)
		if err != nil {
			return nil, err
		}
	}

	// update other pertinent fields in the team (name i guess)
	currentTeam.Name = req.Team.Name

	err = s.teamStore.Upsert(currentTeam)
	if err != nil {
		return nil, err
	}

	return &opsee.UpdateTeamResponse{
		Team: currentTeam,
	}, nil
}

// Sets team to inactive
func (s *service) DeleteTeam(ctx context.Context, req *opsee.DeleteTeamRequest) (*opsee.DeleteTeamResponse, error) {
	if req.Team == nil {
		return nil, fmt.Errorf("invalid request, missing team")
	}

	if err := req.Team.Validate(); err != nil {
		return nil, err
	}

	err := s.teamStore.Delete(req.Team)
	if err != nil {
		return nil, err
	}

	return &opsee.DeleteTeamResponse{
		Team: req.Team,
	}, nil
}
