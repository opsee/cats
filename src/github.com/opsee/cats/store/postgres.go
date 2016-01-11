package store

import (
	"errors"

	_ "database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/com"
	"github.com/opsee/cats/checker"
	log "github.com/sirupsen/logrus"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (Store, error) {
	db, err := sqlx.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(8)

	return &Postgres{
		db: db,
	}, nil
}

// Assertion and checker.Assertion could become confusing eventually, but it seems
// necessary atm.
type Assertion struct {
	CheckID      string `json:"-" db:"check_id"`
	CustomerID   string `json:"-" db:"customer_id"`
	Key          string `json:"key"`
	Relationship string `json:"relationship"`
	Value        string `json:"value,omitempty"`
	Operand      string `json:"operand"`
}

func NewAssertion(customerID string, checkID string, ass *checker.Assertion) *Assertion {
	return &Assertion{
		CheckID:      checkID,
		CustomerID:   customerID,
		Key:          ass.Key,
		Relationship: ass.Relationship,
		Value:        ass.Value,
		Operand:      ass.Operand,
	}
}

func (pg *Postgres) GetAssertions(user *com.User, checks []string) ([]*Assertion, error) {
	var (
		assertions []*Assertion
		query      string
		args       []interface{}
		err        error
	)

	if len(checks) > 0 {
		query, args, err = sqlx.In("SELECT * FROM assertions WHERE customer_id = ? AND check_id IN (?)", user.CustomerID, checks)
		if err != nil {
			return nil, err
		}
	} else {
		query = "SELECT * FROM assertions WHERE customer_id = ?"
		args = []interface{}{user.CustomerID}
	}

	query = pg.db.Rebind(query)
	rows, err := pg.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var a Assertion
		err := rows.StructScan(&a)
		if err != nil {
			panic(err)
		}
		assertions = append(assertions, &a)
	}

	return assertions, err
}

func (pg *Postgres) PutAssertions(user *com.User, checkID string, assertions []*Assertion) error {
	if user == nil {
		return errors.New("PutAssertion received nil user.")
	}

	if checkID == "" {
		return errors.New("PutAssertion received empty checkID")
	}

	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM assertions WHERE customer_id = $1 AND check_id = $2", user.CustomerID, checkID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.WithError(err).Error("Error calling rollback on transaction.")
		}
		return err
	}

	for _, assertion := range assertions {
		_, err = tx.NamedExec(
			`INSERT INTO assertions 
			  VALUES (:check_id, :customer_id, :key, 
			  :relationship, :value, :operand);`,
			assertion,
		)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				log.WithError(err).Error("Error calling rollback on transaction.")
			}
			return err
		}
	}

	return tx.Commit()
}

func (pg *Postgres) DeleteAssertions(user *com.User, checkID string) error {
	_, err := pg.db.Exec("DELETE FROM assertions WHERE customer_id = $1 AND check_id = $2", user.CustomerID, checkID)
	return err
}
