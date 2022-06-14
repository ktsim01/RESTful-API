package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type user struct {
	UserID   string `json:"id"`
	Password string `json:"pw"`
}

// in-memory
type datastore struct {
	users         map[string]user
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
		if r.URL.Path == "/signup" {
			u.signup(w, r)
		} else if r.URL.Path == "/signin" {
			u.signin(w, r)
		} else {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"message": "post called"}`))
		}
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

func (u *userHandler) signup(w http.ResponseWriter, r *http.Request) {
	var u1 user
	if err := json.NewDecoder(r.Body).Decode(&u1); err != nil {
		fmt.Println(err)
		internalServerError(w, r)
		return
	}
	u.store.Lock()
	u.store.users[u1.UserID] = u1
	u.store.Unlock()
	jsonBytes, err := json.Marshal(u1)

	if err != nil {
		fmt.Println(err)
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Write([]byte("New account created!\n"))
	r.SetBasicAuth(u1.UserID, u1.Password)
	r.Header.Add("Content-Type", "application/json")
	r.Close = true

}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal server error"))
}

func (u *userHandler) isAuthorised(username, password string) bool {
	pass := u.store.users[username].Password

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
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "welcome to golang world!"}`))
}

func main() {
	mux := http.NewServeMux()
	userH := &userHandler{
		store: &datastore{
			users: map[string]user{
				"test": {UserID: "test", Password: "secret"},
			},
			RWMutex: &sync.RWMutex{},
		},
	}
	mux.Handle("/", userH)
	mux.Handle("/signin", userH)
	mux.Handle("/signup", userH)
	log.Fatal(http.ListenAndServe(":8080", mux))

}
