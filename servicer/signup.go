package servicer

import (
	"database/sql"
	"fmt"

	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	log "github.com/opsee/logrus"
	"github.com/opsee/vape/model"
	"github.com/opsee/vape/store"
	"github.com/snorecone/closeio-go"
	"golang.org/x/net/context"
)

func GetSignupsByCustomerId(id string) ([]*model.Signup, error) {
	var signups []*model.Signup
	err := store.Select(&signups, "signups-by-customer-id", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, SignupNotFound
		}

		return nil, err
	}

	return signups, nil
}

func GetSignup(id int) (*model.Signup, error) {
	signup := new(model.Signup)
	err := store.Get(signup, "signup-by-id", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, SignupNotFound
		}

		return nil, err
	}

	return signup, nil
}

func DeleteSignup(id int) error {
	_, err := store.Exec("delete-signup-by-id", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}

		return err
	}

	return nil
}

func CreateActiveSignup(email, name, referrer string) (*model.Signup, error) {
	signup, err := createSignup("", email, name, referrer, true, &schema.UserFlags{Admin: true, Billing: true, Edit: true})
	if err != nil {
		return nil, err
	}

	// send an email, create a lead and notify slack here!
	go func() {
		mergeVars := map[string]interface{}{
			"signup_id":    fmt.Sprint(signup.Id),
			"signup_token": VerificationToken(fmt.Sprint(signup.Id)),
			"name":         signup.Name,
		}
		mailTemplatedMessage(signup.Email, signup.Name, "instant-approval", mergeVars)

		lead := &closeio.Lead{
			Name: signup.Email,
			Contacts: []*closeio.Contact{
				{
					Name: signup.Email,
					Emails: []*closeio.Email{
						{
							Type:  "work",
							Email: signup.Email,
						},
					},
				},
			},
		}

		if referrer != "" {
			lead.Custom = map[string]string{
				"referrer": referrer,
			}
		}

		createLead(lead)

		slackMap := map[string]interface{}{
			"user_name":  signup.Name,
			"user_email": signup.Email,
		}

		// work around template shortcomings
		if referrer != "" {
			slackMap["referrer"] = signup.Referrer
		}

		notifySlack("new-signup", slackMap)
	}()

	return signup, err
}

func createSignup(customerId, email, name, referrer string, activated bool, perms *schema.UserFlags) (*model.Signup, error) {
	existingSignup := new(model.Signup)
	err := store.Get(existingSignup, "signup-by-email", email)
	if err == nil {
		return existingSignup, SignupExists
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	if name == "" {
		name = "default"
	}

	signup := &model.Signup{
		Email:      email,
		Name:       name,
		Referrer:   referrer,
		Activated:  activated,
		CustomerId: customerId,
		Perms:      perms,
	}

	if len(signup.Name) > 254 {
		signup.Name = signup.Name[:254]
	}
	if len(signup.Referrer) > 254 {
		signup.Referrer = signup.Referrer[:254]
	}

	err = store.NamedInsert("insert-signup", signup)
	if err != nil {
		return nil, err
	}

	return signup, err
}

func ActivateSignup(id int) (*model.Signup, error) {
	signup, err := GetSignup(id)
	if err != nil {
		return nil, err
	}

	_, err = store.Exec("activate-signup", signup.Id)
	if err != nil {
		return nil, err
	}

	// send an email here!
	go func() {
		mergeVars := map[string]interface{}{
			"signup_id":    fmt.Sprint(signup.Id),
			"signup_token": VerificationToken(fmt.Sprint(signup.Id)),
			"name":         signup.Name,
		}
		mailTemplatedMessage(signup.Email, signup.Name, "beta-approval", mergeVars)
	}()

	return signup, nil
}

func ListSignups(perPage int, page int) ([]*model.Signup, error) {
	if perPage < 1 {
		perPage = 20
	}

	if page < 1 {
		page = 1
	}

	limit := perPage
	offset := (perPage * page) - perPage

	signups := []*model.Signup{}
	err := store.Select(&signups, "list-signups", limit, offset)
	if err != nil {
		return nil, err
	}

	return signups, nil
}

func ClaimSignup(id int, token, name, password string, verified bool) (*schema.User, error) {
	signup, err := GetSignup(id)
	if err != nil {
		return nil, err
	}

	if VerifyToken(fmt.Sprint(signup.Id), token) == false {
		return nil, SignupInvalidToken
	}

	// make sure user hasn't been claimed
	if signup.Claimed {
		return nil, SignupAlreadyClaimed
	}

	// ok, pop that stuff in the user, and make sure they're verified
	user, err := NewUser(name, signup.Email, password)
	if err != nil {
		return nil, err
	}
	user.Verified = verified
	user.Active = true
	user.Perms = signup.Perms
	user.Status = "active"

	tx, err := store.Beginx()
	if err != nil {
		return nil, err
	}

	customerId := signup.CustomerId
	if signup.CustomerId == "" {
		// signup is a new signup -- not user invite. must generate customer
		if err = tx.Get(&customerId, "insert-new-customer"); err != nil {
			tx.Rollback()
			return nil, err
		}
		// ensure that user has admin privs (0111b)
		user.Perms = &schema.UserFlags{Admin: true, Edit: true, Billing: true}
	}
	user.CustomerId = customerId
	if _, err := tx.Exec("claim-signup", signup.Id); err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.NamedInsert("insert-user", user)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	if spanxClient != nil {
		go func() {
			spanxResp, err := spanxClient.EnhancedCombatMode(context.Background(), &opsee.EnhancedCombatModeRequest{
				User: user,
			})

			logger := log.WithFields(log.Fields{
				"email":       user.Email,
				"name":        user.Name,
				"customer_id": user.CustomerId,
			})

			if err != nil {
				logger.WithError(err).Error("error saving new role stack template")
			}

			logger.Infof("saved new role stack template: %s", spanxResp.StackUrl)
		}()
	}

	return user, nil
}
