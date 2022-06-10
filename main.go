package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// in-memory
type datastore struct {
	users         map[string]string
	*sync.RWMutex //concurrent access
}

type userHandler struct {
	store *datastore
}

func (u *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "get called"}`))
	case "POST":
		u.signin(w, r)
	case "PUT":
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"message": "put called"}`))
	case "DELETE":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "delete called"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Kyutae's Awesome"}`))
	}
}

func (u *userHandler) isAuthorised(username, password string) bool {
	pass := u.store.users[username]

	return password == pass
}

func (u *userHandler) signin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "No basic auth present"}`))
		return
	}

	if !u.isAuthorised(username, password) {
		w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid username or password"}`))
		fmt.Printf("%s %s", username, password)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "welcome to golang world!"}`))
}

func main() {
	mux := http.NewServeMux()
	userH := &userHandler{
		store: &datastore{
			users: map[string]string{
				"test": "secret",
				"bob":  "thebuilder123",
			},
			RWMutex: &sync.RWMutex{},
		},
	}
	mux.Handle("/", userH)
	log.Fatal(http.ListenAndServe(":8080", mux))

}
