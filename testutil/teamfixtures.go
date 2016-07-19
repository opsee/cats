package testutil

import (
	"github.com/opsee/basic/schema"
)

var Teams = map[string]*schema.Team{
	"active": {
		Id:                   "11111111-1111-1111-1111-111111111111",
		Name:                 "barbell brigade death squad crew",
		SubscriptionPlan:     "beta",
		SubscriptionStatus:   "active",
		StripeCustomerId:     "cus_8oux3kULDWgU8F",
		StripeSubscriptionId: "sub_8owgXA5pkRDs31",
		SubscriptionQuantity: int32(3),
	},
	"inactive": {
		Id:                   "00000000-0000-0000-0000-000000000000",
		Name:                 "INACTIVE",
		SubscriptionPlan:     "beta",
		SubscriptionStatus:   "canceled",
		StripeCustomerId:     "",
		StripeSubscriptionId: "",
		SubscriptionQuantity: int32(3),
	},
}

var Users = map[string]*schema.User{
	"active_admin": {
		Email:      "opsee+active+admin@opsee.com",
		Active:     true,
		Verified:   true,
		CustomerId: "11111111-1111-1111-1111-111111111111",
		Status:     "active",
		Perms:      &schema.UserFlags{Admin: true, Edit: true, Billing: true},
	},
	"active_editor": {
		Email:      "opsee+active+edit@opsee.com",
		Active:     true,
		Verified:   true,
		CustomerId: "11111111-1111-1111-1111-111111111111",
		Status:     "active",
		Perms:      &schema.UserFlags{Admin: false, Edit: true, Billing: false},
	},
	"inactive_admin": {
		Email:      "opsee+inactive@opsee.com",
		Active:     false,
		Verified:   true,
		CustomerId: "11111111-1111-1111-1111-111111111111",
		Status:     "active",
		Perms:      &schema.UserFlags{Admin: true, Edit: true, Billing: true},
	},
}

var Invites = map[string]*schema.Invite{
	"invited_admin": {
		Email:      "opsee+invited+admin+pending@opsee.com",
		Claimed:    false,
		Perms:      &schema.UserFlags{Admin: true, Edit: true, Billing: true},
		CustomerId: "11111111-1111-1111-1111-111111111111",
	},
	"invited_viewer": {
		Email:      "opsee+invited+noperms+pending@opsee.com",
		Claimed:    false,
		Perms:      &schema.UserFlags{Admin: false, Edit: false, Billing: false},
		CustomerId: "11111111-1111-1111-1111-111111111111",
	},
	"claimed_admin": {
		Email:      "opsee+invited+admin+claimed@opsee.com",
		Claimed:    true,
		Perms:      &schema.UserFlags{Admin: true, Edit: true, Billing: true},
		CustomerId: "11111111-1111-1111-1111-111111111111",
	},
}
