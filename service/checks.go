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
	agent := s.newrelicAgent.StartTransaction("GetChecks", nil, nil)
	defer agent.End()

	if req.Requestor == nil {
		log.Error("no user in request")
		return nil, fmt.Errorf("user is required")
	}

	if err := req.Requestor.Validate(); err != nil {
		log.WithError(err).Error("user is invalid")
		return nil, err
	}

	agent.AddAttribute("user_email", req.Requestor.Email)

	if req.CheckId != "" {
		defer agent.EndSegment(agent.StartSegment(), "checkStore.GetCheck")

		checkId := req.CheckId
		check, err := s.checkStore.GetCheck(req.Requestor, checkId)
		if err != nil {
			log.WithError(err).Errorf("failed to get check from db: %s", checkId)
			return nil, err
		}

		return &opsee.GetChecksResponse{[]*schema.Check{check}}, nil
	}

	defer agent.EndSegment(agent.StartSegment(), "checkStore.GetChecks")

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
		log.WithError(err).Error("Error getting check count from check store.")
		return nil, err
	}

	return &opsee.GetCheckCountResponse{
		Count: int32(count),
	}, nil
}

func (s *service) GetCheckResults(ctx context.Context, req *opsee.GetCheckResultsRequest) (response *opsee.GetCheckResultsResponse, err error) {
	agent := s.newrelicAgent.StartTransaction("GetChecks", nil, nil)
	defer agent.End()

	if req.CustomerId == "" {
		return nil, fmt.Errorf("Request missing CustomerID")
	}

	if req.CheckId == "" {
		return nil, fmt.Errorf("Request missing CheckID")
	}

	logger := log.WithFields(log.Fields{
		"customer_id": req.CustomerId,
		"check_id":    req.CheckId,
	})

	defer agent.EndSegment(agent.StartSegment(), "checkStore.GetLiveBastions")
	bastions, err := s.checkStore.GetLiveBastions(req.CustomerId, req.CheckId)
	if err != nil {
		logger.WithError(err).Error("Error getting live bastions from check store.")
		return nil, err
	}

	defer agent.EndSegment(agent.StartSegment(), "resultStore.GetResultByCheckId")
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
		logger.WithError(err).Error("Error getting check result for bastions.")
		return nil, err
	}
	return &opsee.GetCheckResultsResponse{Results: results}, nil
}

func (s *service) GetCheckStateTransitions(ctx context.Context, req *opsee.GetCheckStateTransitionsRequest) (response *opsee.GetCheckStateTransitionsResponse, err error) {
	if req.CustomerId == "" {
		log.Error("Request missing CheckID")
		return nil, fmt.Errorf("Request missing CustomerID")
	}

	if req.CheckId == "" {
		log.Error("Request missing CheckID")
		return nil, fmt.Errorf("Request missing CheckID")
	}

	logger := log.WithFields(log.Fields{
		"customer_id": req.CustomerId,
		"check_id":    req.CheckId,
	})

	if req.StateTransitionId != 0 {
		// We are getting a specific state transition.
		entry, err := s.checkStore.GetCheckStateTransitionLogEntry(req.CheckId, req.CustomerId, req.StateTransitionId)
		if err != nil {
			logger.WithError(err).Error("Error getting check state transition from DB.")
			return nil, err
		}

		t := &opsee_types.Timestamp{}
		if err := t.Scan(entry.CreatedAt); err != nil {
			logger.WithError(err).Error("Error scanning log entry created at field.")
			return nil, err
		}

		return &opsee.GetCheckStateTransitionsResponse{
			Transitions: []*schema.CheckStateTransition{&schema.CheckStateTransition{
				CheckId:    entry.CheckId,
				From:       entry.From.String(),
				To:         entry.To.String(),
				OccurredAt: t,
				CustomerId: entry.CustomerId,
				Id:         entry.Id,
			}},
		}, nil
	}

	if req.AbsoluteStartTime == nil {
		err := fmt.Errorf("Request missing AbsoluteStartTime")
		logger.WithError(err).Error("Invalid request.")
		return nil, err
	}

	if req.AbsoluteEndTime == nil {
		err := fmt.Errorf("Request missing AbsoluteEndTime")
		logger.WithError(err).Error("Invalid request.")
		return nil, err
	}

	st, err := req.AbsoluteStartTime.Value()
	if err != nil {
		err := fmt.Errorf("Invalid AbsoluteStartTime")
		log.WithError(err).Error("Invalid request.")
		return nil, err
	}
	et, err := req.AbsoluteEndTime.Value()
	if err != nil {
		err := fmt.Errorf("Invalid AbsoluteEndTime")
		logger.WithError(err).Error("Invalid request.")
		return nil, err
	}
	ast, aok := st.(time.Time)
	if !aok {
		err := fmt.Errorf("invalid AbsoluteStartTime")
		logger.WithError(err).Error("Invalid request.")
		return nil, err
	}
	aet, eok := et.(time.Time)
	if !eok {
		err := fmt.Errorf("invalid AbsoluteEndTime")
		logger.WithError(err).Error("Invalid request.")
		return nil, err
	}

	var logEntries []*schema.CheckStateTransition
	entries, err := s.checkStore.GetCheckStateTransitionLogEntries(req.CheckId, req.CustomerId, ast, aet)
	if err != nil {
		logger.WithError(err).Error("Error getting check state transition log entry from check store.")
		return nil, err
	}

	for _, e := range entries {
		timestamp := &opsee_types.Timestamp{}
		if err := timestamp.Scan(e.CreatedAt); err != nil {
			log.WithError(err).Error("Error scaning log entry timestamp.")
			continue
		}

		logEntries = append(logEntries, &schema.CheckStateTransition{
			CheckId:    req.CheckId,
			From:       e.From.String(),
			To:         e.To.String(),
			OccurredAt: timestamp,
			Id:         e.Id,
		})
	}

	return &opsee.GetCheckStateTransitionsResponse{
		Transitions: logEntries,
	}, nil
}

func (s *service) GetCheckSnapshot(ctx context.Context, req *opsee.GetCheckSnapshotRequest) (*opsee.GetCheckSnapshotResponse, error) {
	user := req.Requestor
	if user == nil {
		log.Error("Request requires a user.")
		return nil, fmt.Errorf("Request requires a user.")
	}

	if err := user.Validate(); err != nil {
		log.WithError(err).Error("Error validating user object in request.")
		return nil, err
	}

	ss, err := s.resultStore.GetCheckSnapshot(req.TransitionId, req.CheckId)
	if err != nil {
		log.WithError(err).Error("Error getting check snapshot.")
		return nil, err
	}

	resp := &opsee.GetCheckSnapshotResponse{
		Check: ss,
	}

	return resp, nil
}
