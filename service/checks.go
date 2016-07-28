package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	log "github.com/opsee/logrus"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"golang.org/x/net/context"
)

func (s *service) GetChecks(ctx context.Context, req *opsee.GetChecksRequest) (*opsee.GetChecksResponse, error) {
	if req.Requestor == nil {
		log.Error("no user in request")
		return nil, fmt.Errorf("user is required")
	}

	if err := req.Requestor.Validate(); err != nil {
		log.WithError(err).Error("user is invalid")
		return nil, err
	}

	if req.CheckId != "" {
		checkId := req.CheckId
		check, err := s.checkStore.GetCheck(req.Requestor, checkId)
		if err != nil {
			log.WithError(err).Errorf("failed to get check from db: %s", checkId)
			return nil, err
		}

		return &opsee.GetChecksResponse{[]*schema.Check{check}}, nil
	}

	checks, err := s.checkStore.GetChecks(req.Requestor)
	if err != nil {
		log.WithError(err).Errorf("failed to get checks from db")
		return nil, err
	}

	return &opsee.GetChecksResponse{checks}, nil
}

func (s *service) GetCheckCount(ctx context.Context, req *opsee.GetCheckCountRequest) (*opsee.GetCheckCountResponse, error) {
	if req.User == nil {
		log.Error("no user in request")
		return nil, fmt.Errorf("user is required")
	}

	if err := req.User.Validate(); err != nil {
		log.WithError(err).Error("user is invalid")
		return nil, err
	}

	count, err := s.checkStore.GetCheckCount(req.User.CustomerId)
	if err != nil {
		log.WithError(err).Error("db request failed")
		return nil, err
	}

	return &opsee.GetCheckCountResponse{
		Count: int32(count),
	}, nil
}

func (s *service) GetCheckResults(ctx context.Context, req *opsee.GetCheckResultsRequest) (response *opsee.GetCheckResultsResponse, err error) {
	if req.CustomerId == "" {
		return nil, fmt.Errorf("Request missing CustomerID")
	}

	if req.CheckId == "" {
		return nil, fmt.Errorf("Request missing CheckID")
	}

	bastions, err := s.checkStore.GetLiveBastions(req.CustomerId, req.CheckId)
	if err != nil {
		return nil, err
	}

	results := make([]*schema.CheckResult, len(bastions))
	wg := &sync.WaitGroup{}
	for i, b := range bastions {
		wg.Add(1)
		go func(bastionId string, idx int) {
			defer wg.Done()
			var result *schema.CheckResult
			result, err = s.resultStore.GetResultByCheckId(bastionId, req.CheckId)
			if err != nil {
				log.WithFields(log.Fields{
					"fn":         "GetCheckResults",
					"check_id":   req.CheckId,
					"bastion_id": bastionId,
				}).WithError(err).Error("Error getting result from result store.")
				return
			}
			results[idx] = result
		}(b, i)
	}
	wg.Wait()
	// We can't return partial results, so we return an error if there was any error.
	if err != nil {
		return nil, err
	}
	return &opsee.GetCheckResultsResponse{results}, nil
}

func (s *service) GetCheckStateTransitions(ctx context.Context, req *opsee.GetCheckStateTransitionsRequest) (response *opsee.GetCheckStateTransitionsResponse, err error) {
	if req.CustomerId == "" {
		return nil, fmt.Errorf("Request missing CustomerID")
	}

	if req.CheckId == "" {
		return nil, fmt.Errorf("Request missing CheckID")
	}

	if req.AbsoluteStartTime == nil {
		return nil, fmt.Errorf("Request missing AbsoluteStartTime")
	}

	if req.AbsoluteEndTime == nil {
		return nil, fmt.Errorf("Request missing AbsoluteEndTime")
	}

	st, err := req.AbsoluteStartTime.Value()
	if err != nil {
		return nil, fmt.Errorf("Invalid AbsoluteStartTime")
	}
	et, err := req.AbsoluteEndTime.Value()
	if err != nil {
		return nil, fmt.Errorf("Invalid AbsoluteEndTime")
	}
	ast, aok := st.(time.Time)
	if !aok {
		return nil, fmt.Errorf("invalid AbsoluteStartTime")
	}
	aet, eok := et.(time.Time)
	if !eok {
		return nil, fmt.Errorf("invalid AbsoluteEndTime")
	}

	var logEntries []*schema.CheckStateTransition
	entries, err := s.checkStore.GetCheckStateTransitionLogEntries(req.CheckId, req.CustomerId, ast, aet)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		timestamp := &opsee_types.Timestamp{}
		if err := timestamp.Scan(e.CreatedAt); err != nil {
			continue
		}

		logEntries = append(logEntries, &schema.CheckStateTransition{
			CheckId:    req.CheckId,
			From:       e.From.String(),
			To:         e.To.String(),
			OccurredAt: timestamp,
		})
	}

	return &opsee.GetCheckStateTransitionsResponse{logEntries}, nil
}

func (s *service) GetCheckSnapshot(ctx context.Context, req *opsee.GetCheckSnapshotRequest) (*opsee.GetCheckSnapshotResponse, error) {
	user := req.Requestor
	if user == nil {
		return nil, fmt.Errorf("Request requires a user.")
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	ss, err := s.resultStore.GetCheckSnapshot(req.TransitionId, req.CheckId)
	if err != nil {
		return nil, err
	}

	resp := &opsee.GetCheckSnapshotResponse{
		Check: ss,
	}

	return resp, nil
}
