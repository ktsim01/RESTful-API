package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

/*
{ "id": "f5 demo", "pw": "new account"}

{ "id": "f5 demo", "subject": "I just made a new account", "content":"Let’s go"}

{ "id": "test", "pw": "secret"}

{ "id": "test", "subject": "testing", "content":"Does this work?"}

*/

/*
CREATE TABLE users(
    id SERIAL,
    userID varchar(50) NOT NULL,
    password varchar(50) NOT NULL,
    PRIMARY KEY (id)
)


INSERT INTO users(
	userID,
	password
)
VALUES
    ('id1', 'pw1'),
    ('id2', 'pw2'),
    ('id3', 'pw3');
*/
const (
	DB_HOST     = "172.17.0.2"
	DB_PORT     = 5432
	DB_USER     = "postgres"
	DB_PASSWORD = "f5demo"
	DB_NAME     = "serverdb"
)

func setupDB() *sql.DB {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s \npassword=%s dbname=%s sslmode=disable", DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)

	if err != nil {
		panic(err)
	}
	fmt.Println("Success")

	return db
}

type User struct {
	UserID   string `json:"id"`
	Password string `json:"pw"`
}

type message struct {
	UserID  string `json:"id"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

type userHandler struct {
	DB *sql.DB
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
func (u *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		if r.URL.Path == "/welcome" {
			u.welcome(w, r)
		} else if r.URL.Path == "/users" {
			u.getUsers(w, r)
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

func (u *userHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := u.DB.Query("SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}
	var users []User

	for rows.Next() {
		var user User
		rows.Scan(&user.UserID, &user.Password)
		users = append(users, user)
		fmt.Sprintf("%+v", user)
	}
	userBytes, _ := json.MarshalIndent(users, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(userBytes)

	defer rows.Close()

}
func (u *userHandler) getMsg(w http.ResponseWriter, r *http.Request) {

	rows, err := u.DB.Query("SELECT * FROM messages")
	if err != nil {
		log.Fatal(err)
	}
	var messages []message

	for rows.Next() {
		var msg message
		rows.Scan(&msg.UserID, &msg.Subject, &msg.Content)
		messages = append(messages, msg)
	}
	userBytes, _ := json.MarshalIndent(messages, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(userBytes)
	defer rows.Close()
}

func (u *userHandler) signup(w http.ResponseWriter, r *http.Request) {
	var cred User
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	statement := `INSERT INTO users(id,pw) VALUES ($1, $2)`
	_, err := u.DB.Exec(statement, cred.UserID, cred.Password)
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
	var creds User
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var id_found string
	var pw_found string
	var ctx context.Context
	found := u.DB.QueryRowContext(ctx, "SELECT id, pw FROM users WHERE id=?", creds.UserID).Scan(&id_found, &pw_found)

	if found == sql.ErrNoRows || found != nil {
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
	// u.store.Lock()
	// u.store.messages[post.UserID] = post
	// u.store.Unlock()
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

func main() {
	db := setupDB()
	mux := http.NewServeMux()
	userH := &userHandler{DB: db}
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
