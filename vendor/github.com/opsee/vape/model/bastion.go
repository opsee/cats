package model

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Bastion struct {
	Id           string    `json:"id"`
	PasswordHash string    `json:"_" db:"password_hash"`
	CustomerId   string    `json:"customer_id" db:"customer_id"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// also returns a plaintext password generated here
func NewBastion(customerId string) (*Bastion, string, error) {
	pwbytes := make([]byte, 18)
	if _, err := rand.Read(pwbytes); err != nil {
		return nil, "", err
	}

	pw := base64.StdEncoding.EncodeToString(pwbytes)
	pwhash, err := bcrypt.GenerateFromPassword([]byte(pw), 10)
	if err != nil {
		return nil, "", err
	}

	bastion := &Bastion{PasswordHash: string(pwhash), CustomerId: customerId}
	return bastion, pw, nil
}

func (bastion *Bastion) Authenticate(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(bastion.PasswordHash), []byte(password))
}
