package servicer

import (
	"database/sql"
	"github.com/opsee/vape/model"
	"github.com/opsee/vape/store"
)

func CreateBastion(customerId string) (*model.Bastion, string, error) {
	customer := new(model.Customer)
	err := store.Get(customer, "customer-by-id-and-active", customerId, true)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", CustomerNotFound
		}

		return nil, "", err
	}

	bastion, plaintext, err := model.NewBastion(customerId)
	if err != nil {
		return nil, "", err
	}

	// we'll want to activate the bastion as well
	bastion.Active = true

	err = store.NamedInsert("insert-bastion", bastion)
	if err != nil {
		return nil, "", err
	}

	return bastion, plaintext, nil
}

func AuthenticateBastion(id, password string) error {
	bastion := new(model.Bastion)
	if err := store.Get(bastion, "bastion-join-customer-by-id-and-active", id, true, true); err != nil {
		return err
	}
	return bastion.Authenticate(password)
}
