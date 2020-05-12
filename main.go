package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"shigoto/auth"
	r "shigoto/repositories"
	u "shigoto/resources"

	g "github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func getTaskListByUser(userID string) ([]byte, error) {
	taskRepository := &r.TaskRepository{}
	response := taskRepository.ReadByUserID(userID)

	js, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func writeTask(t *u.TaskItem, userID string) ([]byte, error) {
	taskRepository := &r.TaskRepository{}

	t.UserID = userID
	t.TaskID = strings.ReplaceAll(g.New().String(), "-", "")

	response := taskRepository.Create(t)

	js, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return js, nil
}

func updateTask(userID string, taskID string, updates *u.TaskUpdate) ([]byte, error) {
	taskRepository := &r.TaskRepository{}
	response := taskRepository.Update(userID, taskID, updates)

	js, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func deleteTask(userID string, taskID string) ([]byte, error) {
	taskRepository := &r.TaskRepository{}
	response := taskRepository.Delete(userID, taskID)

	js, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func shigotoHandler(rw http.ResponseWriter, req *http.Request) {
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

func shigotoTaskHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userID := vars["userID"]
	taskID := vars["taskID"]

	switch req.Method {
	case "PATCH":
		var updates u.TaskUpdate
		err := json.NewDecoder(req.Body).Decode(&updates)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		if (updates.Priority == nil) && (updates.Task == nil) {
			http.Error(rw, "No update fields provided in request", http.StatusBadRequest)
			return
		}

		js, err := updateTask(userID, taskID, &updates)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Write(js)

	case "DELETE":
		js, err := deleteTask(userID, taskID)
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
	loginRequest.Username = strings.ToLower(loginRequest.Username)

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
	registerRequest.Username = strings.ToLower(registerRequest.Username)

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

func main() {
	u.InitializeResources()

	rtr := mux.NewRouter().PathPrefix("/api/v1").Subrouter()
	rtr.HandleFunc("/login", shigotoAuthHandler).Methods("POST")
	rtr.HandleFunc("/register", shigotoUserHandler).Methods("POST")
	rtr.Handle("/{userID:[a-zA-Z0-9]+}/tasks", auth.AuthenticationFilter(shigotoHandler))
	rtr.Handle("/{userID:[a-zA-Z0-9]+}/tasks/{taskID:[a-zA-Z0-9-]+}", auth.AuthenticationFilter(shigotoTaskHandler))
	http.Handle("/", rtr)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PATCH", "OPTIONS", "DELETE"})
	origins := handlers.AllowedOrigins([]string{"https://shigoto-app.netlify.app"})
	corsFunc := handlers.CORS(headers, methods, origins)
	if err := http.ListenAndServe(":"+port, corsFunc(rtr)); err != nil {
		log.Fatal(err)
	}
}
