package service

import (
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/servicer"
	"golang.org/x/net/context"
)

// Fetches team, including users
func (s *service) GetTeam(ctx context.Context, req *opsee.GetTeamRequest) (*opsee.GetTeamResponse, error) {
	t, err := servicer.GetTeam(req.Team.Id)
	return &opsee.GetTeamResponse{
		Team: t,
	}, err
}

// Updates team name or subscription
func (s *service) UpdateTeam(ctx context.Context, req *opsee.UpdateTeamRequest) (*opsee.UpdateTeamResponse, error) {
	t, err := servicer.UpdateTeam(req.Team, req.Team.Name, req.Team.Subscription)
	return &opsee.UpdateTeamResponse{
		Team: t,
	}, err
}

// Sets team to inactive
func (s *service) DeleteTeam(ctx context.Context, req *opsee.DeleteTeamRequest) (*opsee.DeleteTeamResponse, error) {
	err := servicer.DeleteTeam(req.Team.Id)
	return &opsee.DeleteTeamResponse{
		Team: req.Team,
	}, err
}
