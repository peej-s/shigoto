package main

import (
	"encoding/json"
	"net/http"
	"strings"

	auth "shigoto/auth"
	r "shigoto/repositories"
	u "shigoto/resources"

	g "github.com/google/uuid"
	"github.com/gorilla/mux"
)

func getTaskListByUser(userID string) ([]byte, error) {
	taskRepository := &r.TaskRepository{}
	responseTaskList := taskRepository.ReadByUserID(userID)

	js, err := json.Marshal(responseTaskList)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func writeTask(t *u.TaskItem, userID string) ([]byte, error) {
	taskRepository := &r.TaskRepository{}

	t.UserID = userID
	t.TaskID = strings.ReplaceAll(g.New().String(), "-", "")

	go taskRepository.Create(t)
	response := u.TaskCreatedResponse{Success: t.TaskID}

	js, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return js, nil
}

func shigotoHandler(rw http.ResponseWriter, req *http.Request) {
	// Todo: Make sure this is as simple as possible
	vars := mux.Vars(req)
	userID := vars["userID"]

	switch req.Method {
	case "GET":
		js, err := getTaskListByUser(userID)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Write(js)
	case "POST":
		var t u.TaskItem
		err := json.NewDecoder(req.Body).Decode(&t)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		js, err := writeTask(&t, userID)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Write(js)
	}
}

func shigotoAuthHandler(rw http.ResponseWriter, req *http.Request) {
	var loginRequest u.User
	err := json.NewDecoder(req.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	accessToken, err := auth.ValidatePassword(&loginRequest)
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
	var registerRequest u.User
	err := json.NewDecoder(req.Body).Decode(&registerRequest)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := auth.RegisterUser(&registerRequest)
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

func Authenticator(rw http.ResponseWriter, req *http.Request, handler http.HandlerFunc) {
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
	err := auth.ValidateToken(token, userID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	handler(rw, req)

}

func AuthenticationFilter(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		Authenticator(rw, req, handler)
	})
}

func main() {
	u.InitializeResources()

	rtr := mux.NewRouter().PathPrefix("/api/v1").Subrouter()
	rtr.HandleFunc("/login", shigotoAuthHandler).Methods("POST")
	rtr.HandleFunc("/register", shigotoUserHandler).Methods("POST")

	rtr.Handle("/{userID:[a-zA-Z0-9]+}/tasks", AuthenticationFilter(shigotoHandler))

	http.Handle("/", rtr)
	http.ListenAndServe(":8080", nil)
}
