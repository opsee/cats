package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/servicer"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"golang.org/x/net/context"
)

func (s *service) GetUser(ctx context.Context, req *opsee.GetUserRequest) (*opsee.GetUserResponse, error) {
	var (
		user *schema.User
		err  error
	)

	if req.CustomerId != "" {
		user, err = servicer.GetUserCustID(req.CustomerId)
		if err != nil {
			return nil, err
		}
	} else if req.Email != "" {
		user, err = servicer.GetUserEmail(req.Email)
		if err != nil {
			return nil, err
		}

	} else {
		user, err = servicer.GetUser(int(req.Id))
		if err != nil {
			return nil, err
		}
	}

	toke, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	return &opsee.GetUserResponse{
		User:       user,
		BasicToken: base64.StdEncoding.EncodeToString(toke),
	}, nil
}

func (s *service) ListUsers(ctx context.Context, req *opsee.ListUsersRequest) (*opsee.ListUsersResponse, error) {
	users, err := servicer.ListUsers(int(req.PerPage), int(req.Page))
	if err != nil {
		return nil, err
	}

	return &opsee.ListUsersResponse{
		PerPage: int32(users.PerPage),
		Page:    int32(users.Page),
		Total:   int32(users.Total),
		Users:   users.Users,
	}, nil
}

// Delete a user
// TODO(dan) This also needs to delete the users subscription.
func (s *service) DeleteUser(ctx context.Context, req *opsee.DeleteUserRequest) (*opsee.DeleteUserResponse, error) {
	err := servicer.DeleteUser(int(req.User.Id))
	if err != nil {
		return nil, err
	}

	return &opsee.DeleteUserResponse{
		User: req.User,
	}, nil
}

// Update a user's email, name, or password
func (s *service) UpdateUser(ctx context.Context, req *opsee.UpdateUserRequest) (*opsee.UserTokenResponse, error) {
	user, err := servicer.GetUser(int(req.User.Id))
	if err != nil {
		return nil, err
	}

	if !req.Requestor.IsOpseeAdmin() && req.Requestor.CustomerId != user.CustomerId {
		return nil, opsee_types.NewPermissionsError("must be on same team")
	}

	if req.Status != "" {
		user.Status = req.Status
	}

	if req.Perms != nil {
		user.Perms = req.Perms
	}

	// TODO(dan) add status and perms to updateuser
	token, err := servicer.UpdateUser(user, req.Email, req.Name, req.Password, time.Hour*24)
	if err != nil {
		return nil, err
	}

	updatedUser, err := servicer.GetUser(int(req.User.Id))
	if err != nil {
		return nil, err
	}

	return &opsee.UserTokenResponse{
		User: &schema.User{
			Id:         updatedUser.Id,
			CustomerId: updatedUser.CustomerId,
			Name:       updatedUser.Name,
			Email:      updatedUser.Email,
			Status:     updatedUser.Status,
			Perms:      updatedUser.Perms,
		},
		Token: token,
	}, nil
}

func (s *service) InviteUser(ctx context.Context, req *opsee.InviteUserRequest) (*opsee.InviteUserResponse, error) {
	// TODO(dan) this could be used as a side-channel to find valid email addresses maybe
	user, err := servicer.GetUserEmail(req.Email)
	if err != nil && err != servicer.UserNotFound {
		return nil, err
	}
	if user != nil {
		return nil, fmt.Errorf("user exists with this email address")
	}

	teamName := ""
	team, _ := servicer.GetTeam(req.Requestor.CustomerId)
	if team != nil {
		teamName = team.Name
	}

	senderEmail := req.Requestor.Email

	invite, err := servicer.CreateActiveInvite(teamName, senderEmail, req.Requestor.CustomerId, req.Email, req.Perms)
	if err != nil {
		return nil, err
	}

	return &opsee.InviteUserResponse{
		Invite: invite,
	}, nil
}
