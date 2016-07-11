package testutil

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gopkg.in/check.v1"
)

type DB interface {
	Get(interface{}, string, ...interface{}) error
	Exec(string, ...interface{}) (sql.Result, error)
	Beginx() (*sqlx.Tx, error)
}

func SetupFixtures(db DB, c *check.C) {
	tx, err := db.Beginx()
	if err != nil {
		c.Fatal(err)
	}

	// teardown first since it's nice to have lingering data to play with after a test
	_, err = tx.Exec("delete from signups")
	if err != nil {
		c.Fatal(err)
	}
	_, err = tx.Exec("delete from bastions")
	if err != nil {
		c.Fatal(err)
	}
	_, err = tx.Exec("delete from users")
	if err != nil {
		c.Fatal(err)
	}
	_, err = tx.Exec("delete from customers")
	if err != nil {
		c.Fatal(err)
	}

	// password for both users is "eatshit"
	// create an admin user (fk constraint on customer_id)
	var id string
	err = tx.Get(&id, "insert into customers (name, active) values ('markorg', true) returning id")
	if err != nil {
		c.Fatal(err)
	}
	_, err = tx.Exec(
		"insert into users (id, email, password_hash, admin, active, verified, "+
			"customer_id, name, status, perms) values (1, 'mark@opsee.co', "+
			"'$2a$10$QcgjlXDKnRys50Oc30duFuNcZW6Rmqd7pcIJX9GWheIXJExUooZ7W', true, true, true, "+
			"$1, 'mark', 'active', 1)", id)
	if err != nil {
		c.Fatal(err)
	}

	// create a regular user (fk constraint on customer_id)
	err = tx.Get(&id, "insert into customers (name, active) values ('danorg', true) returning id")
	if err != nil {
		c.Fatal(err)
	}
	_, err = tx.Exec(
		"insert into users (id, email, password_hash, admin, active, verified, "+
			"customer_id, name, status, perms) values (3, 'dan@opsee.co', "+
			"'$2a$10$QcgjlXDKnRys50Oc30duFuNcZW6Rmqd7pcIJX9GWheIXJExUooZ7W', false, true, true, "+
			"$1, 'dan', 'active', 3)", id)
	if err != nil {
		c.Fatal(err)
	}

	tx.Commit()
}
