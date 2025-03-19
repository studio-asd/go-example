package api

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// password representation of safe string to be displayed.
type password string

func (p password) String() string {
	return "*****"
}

// createUserPassword creates the user password from the existing rawPassword.
func createUserPassword(rawPassword string) (password, error) {
	// Please NOTE that we are using bcrypt to hash the password and the algorithm has a limitation of 72 characters.
	//
	// Okta previosuly has security incident because they are allowing more than 72 characters for the password while
	// they are using bcrypt algorithm to hash the password. https://trust.okta.com/security-advisories/okta-ad-ldap-delegated-authentication-username/.
	//
	// Yo can look at one of the interesting read here: https://n0rdy.foo/posts/20250121/okta-bcrypt-lessons-for-better-apis/.
	if len(rawPassword) > 72 {
		return "", errors.New("password length cannot be more than 72 characters")
	}
	out, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return password(out), nil
}
