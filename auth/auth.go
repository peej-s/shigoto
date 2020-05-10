package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	r "shigoto/repositories"
	u "shigoto/resources"
	"strings"
	"time"

	g "github.com/google/uuid"
	"github.com/gorilla/mux"
	b "golang.org/x/crypto/bcrypt"
)

func authenticator(rw http.ResponseWriter, req *http.Request, handler http.HandlerFunc) {
	// Get UserID from request
	vars := mux.Vars(req)
	userID := vars["userID"]
	var token *u.AccessToken = &u.AccessToken{}

	// Get Token from Header
	headerToken := req.Header.Get("Authorization")
	splitToken := strings.Split(headerToken, "Bearer")
	if len(splitToken) != 2 {
		http.Error(rw, "Error: Bearer token not in proper format", http.StatusBadRequest)
		return
	}
	headerToken = strings.TrimSpace(splitToken[1])
	token.Token = headerToken

	// Validate Token and UserID
	err := validateToken(token, userID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	handler(rw, req)

}

func generateToken(userID string) *u.AccessToken {
	token := make([]byte, 16)
	rand.Read(token)
	accessToken := &u.AccessToken{Token: fmt.Sprintf("%x", token), UserID: userID, Expiry: time.Now().AddDate(0, 0, 1)}
	tokenRepository := &r.TokenRepository{}
	tokenRepository.Upsert(accessToken)
	return accessToken
}

func validateToken(token *u.AccessToken, currentUser string) error {
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

func AuthenticationFilter(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		authenticator(rw, req, handler)
	})
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

	return generateToken(savedUser.UserID), nil
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

	userRepository.Create(registerRequest)
	return generateToken(registerRequest.UserID), nil
}
