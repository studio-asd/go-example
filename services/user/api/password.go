package api

import (
	"errors"
	"math/rand"

	"github.com/studio-asd/go-example/services/user"
	"golang.org/x/crypto/bcrypt"
)

var char = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// password representation of safe string to be displayed.
type password string

func (p password) String() string {
	return "*****"
}

// encryptUserPassword encrypts the user password from the existing rawPassword.
func encryptUserPassword(rawPassword, salt string) (password, error) {
	if salt == "" {
		return "", user.ErrPasswordSaltEmpty
	}
	if len(rawPassword) < 8 {
		return "", user.ErrPasswordTooShort
	}
	if len(rawPassword) > 36 {
		return "", user.ErrPasswordTooLong
	}

	prefixSalt := salt[0 : len(salt)/2]
	suffixSalt := salt[len(salt)/2:]
	// The raw passwrod is generated through hashing the password with a salt and constructed in a specific way.
	// raw_password := prefixSalt + value.SecretValue + suffixSalt
	finalPassword := prefixSalt + rawPassword + suffixSalt

	// Please NOTE that we are using bcrypt to hash the password and the algorithm has a limitation of 72 characters.
	//
	// Okta previosuly has security incident because they are allowing more than 72 characters for the password while
	// they are using bcrypt algorithm to hash the password. https://trust.okta.com/security-advisories/okta-ad-ldap-delegated-authentication-username/.
	//
	// Yo can look at one of the interesting read here: https://n0rdy.foo/posts/20250121/okta-bcrypt-lessons-for-better-apis/.
	//
	// Go's bcrypt package has already done this, but we will still do this for awareness.
	if len(finalPassword) > 72 {
		return "", errors.New("password length cannot be more than 72 characters")
	}
	out, err := bcrypt.GenerateFromPassword([]byte(finalPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return password(out), nil
}

// randSalt returns a random salt characters with length 16.
func randSalt() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = char[rand.Intn(len(char))] // #nosec
	}
	return string(b)
}
