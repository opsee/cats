package results

import (
	"github.com/opsee/basic/schema"
)

type Store interface {
	GetResultByCheckId(bastionId, customerId string) (*schema.CheckResult, error)
	PutResult(result *schema.CheckResult) error
}
