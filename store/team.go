package store

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/opsee/basic/schema"
)

type teamStore struct {
	sqlx.Ext
}

func NewTeamStore(q sqlx.Ext) TeamStore {
	return &teamStore{q}
}

func (q *teamStore) Get(id string) (*schema.Team, error) {
	team := new(schema.Team)
	err := sqlx.Get(q, team, "select id, name, subscription, stripe_customer_id, stripe_subscription_id from customers where id = $1 and active = true", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	users, err := q.GetUsers(team.Id)
	if err != nil {
		return nil, err
	}

	invites, err := q.GetInvites(team.Id)
	if err != nil {
		return nil, err
	}

	team.Users = append(users, invites...)

	return team, nil
}

func (q *teamStore) GetUsers(id string) ([]*schema.User, error) {
	var users []*schema.User

	err := sqlx.Select(q, &users, "select * from users where customer_id = $1 and active = true", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return users, nil
		}

		return nil, err
	}

	return users, nil
}

func (q *teamStore) GetInvites(id string) ([]*schema.User, error) {
	var (
		invites []*schema.Invite
		users   []*schema.User
	)

	err := sqlx.Select(q, &invites, "select * from signups where customer_id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return users, nil
		}

		return nil, err
	}

	for _, invite := range invites {
		if invite.Claimed == false {
			u := &schema.User{
				Id:         0, // meh
				CustomerId: id,
				Email:      invite.Email,
				Perms:      invite.Perms,
				Status:     "invited",
			}

			if u.Perms == nil {
				u.Perms = &schema.UserFlags{Admin: false, Edit: false, Billing: false}
			}

			users = append(users, u)
		}
	}

	return users, nil
}

// Inserts a new team into the database, and mutates the team pointer to fill in the returned id.
func (q *teamStore) Create(team *schema.Team) error {
	rows, err := sqlx.NamedQuery(
		q,
		`insert into customers (name, active, subscription, stripe_customer_id, stripe_subscription_id, subscription_quantity)
		values (:name, true, :subscription, :stripe_customer_id, :stripe_subscription_id, :subscription_quantity)
		returning id, name, subscription, stripe_customer_id, stripe_subscription_id`,
		team,
	)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(team)
		if err != nil {
			return err
		}
	}

	return nil
}

// Updates an existing team in the database,
// and mutates the team pointer that is passed to it. Will return error if team is not active (soft-deleted).
func (q *teamStore) Update(team *schema.Team) error {
	rows, err := sqlx.NamedQuery(
		q,
		`update customers set
		name = :name, subscription = :subscription, stripe_customer_id = :stripe_customer_id, 
		stripe_subscription_id = :stripe_subscription_id, subscription_quantity = :subscription_quantity where id = :id and active = true
		returning id, name, subscription, stripe_customer_id, stripe_subscription_id`,
		team,
	)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(team)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete deletes the team
func (q *teamStore) Delete(team *schema.Team) error {
	_, err := q.Exec("update customers set active = false where id = $1", team.Id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}
