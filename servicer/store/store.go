package store

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Tx struct {
	*sqlx.Tx
}

var DB *sqlx.DB

var queries = map[string]string{
	// users
	"list-users":               "select * from users limit $1 offset $2",
	"total-users":              "select count(id) from users",
	"user-by-email":            "select *, length(password_hash) <> 0 as has_password from users where lower(email) = lower($1)",
	"user-by-email-and-active": "select *, length(password_hash) <> 0 as has_password from users where lower(email) = lower($1) and active = $2",
	"user-by-id":               "select *, length(password_hash) <> 0 as has_password from users where id = $1",
	"user-by-cust-id":          "select *, length(password_hash) <> 0 as has_password from users where customer_id = $1",
	"delete-user-by-id":        "delete from users where id = $1",
	"update-user":              "update users set name = :name, email = :email, password_hash = :password_hash, status = :status, verified = :verified, perms = :perms where id = :id",
	"update-user-perms":        "update users set perms = :perms where id = :id",
	"insert-user":              "insert into users (customer_id, email, name, verified, active, password_hash, status, perms) values (:customer_id, :email, :name, :verified, :active, :password_hash, :status, :perms) returning *, length(password_hash) <> 0 as has_password",

	// userdata
	"userdata-by-id": "select data from userdata where user_id = $1",
	"merge-userdata": "update userdata set data = json_merge(data, $2::jsonb) where user_id = $1 returning data",

	// signups
	"signup-by-id":           "select * from signups where id = $1",
	"signups-by-customer-id": "select * from signups where customer_id = $1",
	"delete-signup-by-id":    "delete from signups where id = $1",
	"signup-by-email":        "select * from signups where lower(email) = lower($1)",
	"insert-signup":          "insert into signups (email, name, referrer, activated, customer_id, perms) values (:email, :name, :referrer, :activated, :customer_id, :perms) returning *",
	"list-signups":           "select * from signups limit $1 offset $2",
	"activate-signup":        "update signups set activated = true where id = $1",
	"claim-signup":           "update signups set claimed = true where id = $1",

	// customers
	"customer-by-id-and-active": "select * from customers where id = $1 and active = $2",
	"customer-by-id":            "select * from customers where id = $1",
	"insert-new-customer":       "insert into customers (name, active) values ('default', true) returning id",

	// teams (a subset of customer fields and actions accessible to team admins)
	"update-team":       "update customers set name = :name, subscription = :subscription where id = :id and active = true",
	"team-by-id":        "select id, name, subscription from customers where id = $1 and active = true",
	"team-by-name":      "select id, name, subscription from customers where name = $1",
	"team-users-by-id":  "select id, name, email, status, perms from users where customer_id = $1",
	"delete-team-by-id": "update customers set active = false, where id = $1",

	// bastions
	"insert-bastion":                         "insert into bastions (password_hash, customer_id, active) values (:password_hash, :customer_id, :active) returning *",
	"bastion-join-customer-by-id-and-active": "select b.* from bastions b inner join customers c on b.customer_id = c.id where b.id = $1 and b.active = $2 and c.active = $3",
}

func Init(pgConnection string) error {
	var err error

	db, err := sqlx.Open("postgres", pgConnection)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(8)

	DB = db
	return nil
}

func Get(obj interface{}, queryKey string, args ...interface{}) error {
	return DB.Get(obj, queries[queryKey], args...)
}

func Select(obj interface{}, queryKey string, args ...interface{}) error {
	return DB.Select(obj, queries[queryKey], args...)
}

func Exec(queryKey string, args ...interface{}) (sql.Result, error) {
	return DB.Exec(queries[queryKey], args...)
}

func NamedExec(queryKey string, arg interface{}) (sql.Result, error) {
	return DB.NamedExec(queries[queryKey], arg)
}

func NamedQuery(queryKey string, arg interface{}) (*sqlx.Rows, error) {
	return DB.NamedQuery(queries[queryKey], arg)
}

func NamedInsert(queryKey string, arg interface{}) error {
	rows, err := NamedQuery(queryKey, arg)
	if err != nil {
		return err
	}
	for rows.Next() {
		if err = rows.StructScan(arg); err != nil {
			return err
		}
	}
	return nil
}

func Beginx() (*Tx, error) {
	tx, err := DB.Beginx()
	return &Tx{tx}, err
}

func (tx *Tx) Get(obj interface{}, queryKey string, args ...interface{}) error {
	return tx.Tx.Get(obj, queries[queryKey], args...)
}

func (tx *Tx) Select(obj interface{}, queryKey string, args ...interface{}) error {
	return tx.Tx.Select(obj, queries[queryKey], args...)
}

func (tx *Tx) Exec(queryKey string, args ...interface{}) (sql.Result, error) {
	return tx.Tx.Exec(queries[queryKey], args...)
}

func (tx *Tx) NamedExec(queryKey string, arg interface{}) (sql.Result, error) {
	return tx.Tx.NamedExec(queries[queryKey], arg)
}

func (tx *Tx) NamedQuery(queryKey string, arg interface{}) (*sqlx.Rows, error) {
	return tx.Tx.NamedQuery(queries[queryKey], arg)
}

func (tx *Tx) NamedInsert(queryKey string, arg interface{}) error {
	rows, err := tx.NamedQuery(queryKey, arg)
	if err != nil {
		return err
	}
	for rows.Next() {
		if err = rows.StructScan(arg); err != nil {
			return err
		}
	}
	return nil
}
