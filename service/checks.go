package service

import (
	"fmt"
	"sync"

	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	log "github.com/opsee/logrus"
	"golang.org/x/net/context"
)

func (s *service) GetCheckCount(ctx context.Context, req *opsee.GetCheckCountRequest) (*opsee.GetCheckCountResponse, error) {
	if req.User == nil {
		log.Error("no user in request")
		return nil, fmt.Errorf("user is required")
	}

	if err := req.User.Validate(); err != nil {
		log.WithError(err).Error("user is invalid")
		return nil, err
	}

	count, err := s.checkStore.GetCheckCount(req.User, req.Prorated)
	if err != nil {
		log.WithError(err).Error("db request failed")
		return nil, err
	}

	return &opsee.GetCheckCountResponse{
		Prorated: req.Prorated,
		Count:    count,
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
			var result *schema.CheckResult
			result, err = s.resultStore.GetResultByCheckId(bastionId, req.CheckId)
			if err != nil {
				log.WithFields(log.Fields{
					"fn":         "GetCheckResults",
					"check_id":   req.CheckId,
					"bastion_id": bastionId,
				}).WithError(err).Error("Error getting result from result store.")
			}
			results[idx] = result
			wg.Done()
		}(b, i)
	}
	wg.Wait()
	// We can't return partial results, so we return an error if there was any error.
	if err != nil {
		return nil, err
	}
	return &opsee.GetCheckResultsResponse{results}, nil
}
