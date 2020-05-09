package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	r "shigoto/repositories"
	u "shigoto/resources"
	"strings"
	"time"

	g "github.com/google/uuid"
	b "golang.org/x/crypto/bcrypt"
)

func GenerateToken(userID string) *u.AccessToken {
	token := make([]byte, 16)
	rand.Read(token)
	accessToken := &u.AccessToken{Token: fmt.Sprintf("%x", token), UserID: userID, Expiry: time.Now().AddDate(0, 0, 1)}
	tokenRepository := &r.TokenRepository{}
	go tokenRepository.Upsert(accessToken)
	return accessToken
}

func ValidatePassword(loginRequest *u.User) (*u.AccessToken, error) {
	if loginRequest.Username == "" {
		return nil, errors.New("Missing Username in Login Request")
	}
	if loginRequest.Password == "" {
		return nil, errors.New("Missing Password in Login Request")
	}
	userRepository := &r.UserRepository{}
	savedUser := userRepository.ReadByUsername(loginRequest.Username)
	if savedUser == nil {
		return nil, errors.New("Username does not exist")
	}
	err := b.CompareHashAndPassword([]byte(savedUser.Password), []byte(loginRequest.Password))
	if err == b.ErrMismatchedHashAndPassword {
		return nil, errors.New("Incorrect Password")
	} else if err != nil {
		return nil, err
	}

	return GenerateToken(savedUser.UserID), nil
}

func RegisterUser(registerRequest *u.User) (*u.AccessToken, error) {
	if len(registerRequest.Username) < 4 {
		return nil, errors.New("Username must have at least 4 characters in Login Request")
	}
	if len(registerRequest.Password) < 8 {
		return nil, errors.New("Password must have at least 8 characters in Login Request")

	}
	userRepository := &r.UserRepository{}
	existingUser := userRepository.ReadByUsername(registerRequest.Username)
	if existingUser != nil {
		return nil, errors.New("Username already exists")
	}
	registerRequest.UserID = strings.ReplaceAll(g.New().String(), "-", "")

	go userRepository.Create(registerRequest)
	return GenerateToken(registerRequest.UserID), nil
}

func ValidateToken(token *u.AccessToken, currentUser string) error {
	tokenRepository := &r.TokenRepository{}
	savedToken := tokenRepository.ReadByUserID(currentUser)
	if savedToken == nil {
		return errors.New("Username does not exist")
	}
	if token.Token != savedToken.Token {
		return errors.New(fmt.Sprintf("Token does not match saved token for user %s", currentUser))
	}
	if time.Now().After(savedToken.Expiry) {
		return errors.New(fmt.Sprintf("Token expired for user %s", currentUser))
	}
	return nil
}
