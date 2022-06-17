package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

/*
{ "id": "f5 demo", "pw": "new account"}

{ "id": "f5 demo", "subject": "I just made a new account", "content":"Let’s go"}

{ "id": "test", "pw": "secret"}

{ "id": "test", "subject": "testing", "content":"Does this work?"}

*/

type user struct {
	UserID   string `json:"id"`
	Password string `json:"pw"`
}

type message struct {
	UserID  string `json:"id"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

// in-memory
type datastore struct {
	users         map[string]user
	messages      map[string]message
	*sync.RWMutex //concurrent access
}

type userHandler struct {
	store *datastore
}

// users sessions
var sessions = map[string]session{}

// session contains the username and the time of expiration
type session struct {
	username string
	expiry   time.Time
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}
func (u *userHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%+v", u.store.users)))
}

func (u *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		if r.URL.Path == "/welcome" {
			u.welcome(w, r)
		} else if r.URL.Path == "/users" {
			u.listUsers(w, r)
		} else if r.URL.Path == "/data" {
			u.getMsg(w, r)
		} else {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"message": "get called"}`))
		}
	case "POST":
		if r.URL.Path == "/signup" {
			u.signup(w, r)
		} else if r.URL.Path == "/signin" {
			u.signin(w, r)
		} else if r.URL.Path == "/refresh" {
			u.refresh(w, r)
		} else if r.URL.Path == "/logout" {
			u.logout(w, r)
		} else if r.URL.Path == "/data" {
			u.sendMsg(w, r)
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
	var cred user
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	u.store.Lock()
	u.store.users[cred.UserID] = cred
	u.store.Unlock()
	jsonBytes, err := json.Marshal(cred)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	w.Write([]byte("\nNew account created!\n"))
	r.SetBasicAuth(cred.UserID, cred.Password)
	r.Header.Add("Content-Type", "application/json")
	r.Close = true

}

func (u *userHandler) signin(w http.ResponseWriter, r *http.Request) {
	var creds user
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	expectedPassword := u.store.users[creds.UserID].Password
	if expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create a new random session token
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(30 * time.Second)

	// Set the token in the session map
	sessions[sessionToken] = session{
		username: creds.UserID,
		expiry:   expiresAt,
	}

	// set the client cookie as the session token we just created and set the expiration time to 2 minutes
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})
}
func (u *userHandler) welcome(w http.ResponseWriter, r *http.Request) {
	// obtaining the session token from the request cookie
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// return an unauthorized status if the cookie is not set
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Your session has expired."))
			return
		}
		// Return a bad request status for any other errors
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
		return
	}
	sessionToken := c.Value

	// Retreive the session from our session map
	userSession, exists := sessions[sessionToken]
	if !exists {
		// Unauthorized error if the seesion token is not present
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// For expired sessions, delete token and return an unauthorized status
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// For valid sessions, return a welcome message
	w.Write([]byte(fmt.Sprintf("Welcome %s!", userSession.username)))
}

func (u *userHandler) refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	userSession, exists := sessions[sessionToken]
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// If the previous session is valid, create a new session token for the current user
	newSessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	// Set the token in the session map, along with the user whom it represents
	sessions[newSessionToken] = session{
		username: userSession.username,
		expiry:   expiresAt,
	}

	// Delete the old session token
	delete(sessions, sessionToken)

	// Set the new token as the users `session_token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newSessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
	w.Write([]byte("Your session has been refreshed!"))
}

func (u *userHandler) logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// remove the users session from the session map
	delete(sessions, sessionToken)

	// Cookie is expired
	// Set the session token to an empty value and set its expiry as the current time
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})
	w.Write([]byte("Logged out succesfully. Goodbye!"))
}

func (u *userHandler) sendMsg(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Error: unauthorized."))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error: bad request."))
		return
	}
	sessionToken := c.Value

	userSession, exists := sessions[sessionToken]
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	//If the session is valid, users can send send a message
	var post message
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	u.store.Lock()
	u.store.messages[post.UserID] = post
	u.store.Unlock()
	jsonBytes, err := json.Marshal(post)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
	r.Header.Add("Content-Type", "application/json")
	r.Close = true
	w.Write([]byte("\nMessage received!"))
}
func (u *userHandler) getMsg(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%+v", u.store.messages)))
}

func main() {
	mux := http.NewServeMux()
	userH := &userHandler{
		store: &datastore{
			users: map[string]user{
				"test": {UserID: "test", Password: "secret"},
			},
			messages: map[string]message{},
			RWMutex:  &sync.RWMutex{},
		},
	}
	mux.Handle("/", userH)
	mux.Handle("/signin", userH)
	mux.Handle("/signup", userH)
	mux.Handle("/refresh", userH)
	mux.Handle("/welcome", userH)
	mux.Handle("/logout", userH)
	mux.Handle("/users", userH)
	mux.Handle("/data", userH)
	log.Fatal(http.ListenAndServe(":8080", mux))

}
