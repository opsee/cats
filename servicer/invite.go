package servicer

import (
	"fmt"

	"github.com/opsee/basic/schema"
	log "github.com/opsee/logrus"
)

func CreateActiveInvite(teamName, senderEmail, customerId, email string, perms *schema.UserFlags) (*schema.Invite, error) {
	referrer := ""
	signup, err := createSignup(customerId, email, "", referrer, true, perms)
	if err != nil {
		// only return err if signupexists and was previously claimed
		if err == SignupExists && signup.Claimed == true {
			return nil, err
		} else if err != SignupExists {
			return nil, err
		}
	}

	log.Debugf("invite signup %v permissions: %v", signup, signup.Perms)
	invite := &schema.Invite{
		Id:         int32(signup.Id),
		Email:      signup.Email,
		CustomerId: signup.CustomerId,
		Perms:      signup.Perms,
	}

	// send an email, create a lead and notify slack here!
	go func() {
		mergeVars := map[string]interface{}{
			"signup_id":    fmt.Sprint(signup.Id),
			"signup_token": VerificationToken(fmt.Sprintf("%d", signup.Id)),
			"team_name":    teamName,
			"sender_email": senderEmail,
		}
		mailTemplatedMessage(signup.Email, signup.Name, "team-invitation", mergeVars)
	}()

	return invite, nil
}
