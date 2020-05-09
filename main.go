package main

import (
	"encoding/json"
	"net/http"
	"strings"

	auth "shigoto/auth"
	u "shigoto/resources"

	"github.com/gorilla/mux"
)

func shigotoAuthHandler(rw http.ResponseWriter, req *http.Request) {
	var loginRequest *u.User
	err := json.NewDecoder(req.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	accessToken, err := auth.ValidatePassword(loginRequest)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	js, err := json.Marshal(accessToken)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(js)
}

func shigotoUserHandler(rw http.ResponseWriter, req *http.Request) {
	var registerRequest *u.User
	err := json.NewDecoder(req.Body).Decode(&registerRequest)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := auth.RegisterUser(registerRequest)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(js)
}

func shigotoTokenHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userID := vars["userID"]
	var token *u.AccessToken = &u.AccessToken{}

	headerToken := req.Header.Get("Authorization")
	splitToken := strings.Split(headerToken, "Bearer")
	if len(splitToken) != 2 {
		http.Error(rw, "Error: Bearer token not in proper format", http.StatusBadRequest)
		return
	}
	headerToken = strings.TrimSpace(splitToken[1])
	token.Token = headerToken

	err := auth.ValidateToken(token, userID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Write([]byte("Valid Token"))
}

func main() {
	u.InitializeResources()

	rtr := mux.NewRouter()
	rtr.HandleFunc("/login", shigotoAuthHandler).Methods("POST")
	rtr.HandleFunc("/register", shigotoUserHandler).Methods("POST")
	// For debugging only, used to validate a user token
	rtr.HandleFunc("/{userID:[a-zA-Z0-9-]+}/token", shigotoTokenHandler)

	http.Handle("/", rtr)
	http.ListenAndServe(":8080", nil)
}
