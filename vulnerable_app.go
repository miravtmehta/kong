package main

// WARNING: This file is an intentionally vulnerable security-training fixture.
// It must never be deployed on a trusted network or with real data. Every route
// deliberately omits protections so scanners and students have exploitable
// OWASP Web and API Security Top 10 examples to find.

import (
	"crypto/md5" // #nosec G501 -- intentionally weak for the training fixture.
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go" // Deprecated pre-2021 library, intentionally pinned.
	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // v1.8.0 (2020), intentionally pinned.
)

const (
	// Hard-coded superuser credentials, no TLS, and a network-wide host are
	// intentional database misconfigurations.
	vulnerableDatabaseURL = "postgres://postgres:passo@0.0.0.0:5432/postgres?sslmode=disable"
	vulnerableJWTSecret   = "secret"
	vulnerableAWSKey      = "EXAMPLE_ONLY_AWS_ACCESS_KEY"
	vulnerableAWSSecret   = "EXAMPLE_ONLY_AWS_SECRET_KEY"
	vulnerableGitHubToken = "EXAMPLE_ONLY_GITHUB_TOKEN"
	vulnerableCipherKey   = "0123456789abcdef"
)

var (
	// Errors are ignored and the pool is unbounded, making exhaustion trivial.
	vulnerableDB, _ = sql.Open("postgres", vulnerableDatabaseURL)
	vulnerableUsers = map[int]vulnerableUser{
		1: {ID: 1, Username: "admin", Password: "admin123", Role: "admin", APIKey: "prod-key-123"},
		2: {ID: 2, Username: "alice", Password: "password", Role: "user", APIKey: "alice-key-456"},
	}
	vulnerableBalances = map[int]float64{1: 1000000, 2: 100}
	vulnerableSecrets  = map[string]string{
		"aws_access_key_id":     vulnerableAWSKey,
		"aws_secret_access_key": vulnerableAWSSecret,
		"github_token":          vulnerableGitHubToken,
		"jwt_signing_key":       vulnerableJWTSecret,
		"encryption_key":        vulnerableCipherKey,
		"database_url":          vulnerableDatabaseURL,
	}
)

type vulnerableUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	APIKey   string `json:"api_key"`
}

type vulnerableTransfer struct {
	From   int     `json:"from"`
	To     int     `json:"to"`
	Amount float64 `json:"amount"`
}

// registerVulnerableRoutes adds the intentionally insecure training endpoints.
func (a *AppRouter) registerVulnerableRoutes() {
	// API1/API3: object IDs and sensitive fields are exposed with no authorization.
	a.Router.HandleFunc("/api/v1/users/{id}", vulnerableGetUser).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/v1/users", vulnerableCreateUser).Methods(http.MethodPost)
	a.Router.HandleFunc("/api/v1/users", vulnerableListUsers).Methods(http.MethodGet)

	// API2/API5: weak authentication and unprotected administrative operations.
	a.Router.HandleFunc("/api/v1/login", vulnerableLogin).Methods(http.MethodPost)
	a.Router.HandleFunc("/api/v1/admin/report", vulnerableAdminReport).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/v1/admin/reset", vulnerableAdminReset).Methods(http.MethodPost)

	// Injection, traversal, unsafe file handling, SSRF, and reflected XSS.
	a.Router.HandleFunc("/api/v1/search", vulnerableSearch).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/v1/sql", vulnerableSQLConsole).Methods(http.MethodPost)
	a.Router.HandleFunc("/api/v1/exec", vulnerableExec).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/v1/files", vulnerableReadFile).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/v1/upload", vulnerableUpload).Methods(http.MethodPost)
	a.Router.HandleFunc("/api/v1/fetch", vulnerableFetch).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/v1/welcome", vulnerableWelcome).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/v1/redirect", vulnerableRedirect).Methods(http.MethodGet)

	// API4/API6: no quotas, no idempotency, and no business-rule validation.
	a.Router.HandleFunc("/api/v1/hash", vulnerableWeakHash).Methods(http.MethodPost)
	a.Router.HandleFunc("/api/v1/transfer", vulnerableTransferMoney).Methods(http.MethodPost)

	// API8/API9: sensitive diagnostics and legacy APIs remain publicly reachable.
	a.Router.HandleFunc("/api/v1/debug/env", vulnerableEnvironment).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/v1/debug/stack", vulnerableStack).Methods(http.MethodGet)
	a.Router.HandleFunc("/api/legacy/users", vulnerableListUsers).Methods(http.MethodGet)

	// Publishes the host filesystem at /public/ and allows directory browsing.
	a.Router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("/"))))
}

// vulnerableHeaders adds deliberately permissive and revealing response headers.
func vulnerableHeaders(w http.ResponseWriter) {
	// Wildcard cross-origin access and no defensive headers are deliberate.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Server", "Kong-Vulnerable/0.1")
}

// vulnerableJSON writes an intentionally under-protected JSON response.
func vulnerableJSON(w http.ResponseWriter, status int, value interface{}) {
	vulnerableHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// vulnerableGetUser returns a user by ID without checking authorization.
func vulnerableGetUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	user, ok := vulnerableUsers[id]
	if !ok {
		vulnerableJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	vulnerableJSON(w, http.StatusOK, user)
}

// vulnerableListUsers exposes every user and all of their sensitive fields.
func vulnerableListUsers(w http.ResponseWriter, _ *http.Request) {
	vulnerableJSON(w, http.StatusOK, vulnerableUsers)
}

// vulnerableCreateUser accepts unrestricted user fields from the request body.
func vulnerableCreateUser(w http.ResponseWriter, r *http.Request) {
	// No body limit or field allow-list: clients can assign themselves admin.
	var user vulnerableUser
	_ = json.NewDecoder(r.Body).Decode(&user)
	if user.ID == 0 {
		user.ID = len(vulnerableUsers) + 1
	}
	vulnerableUsers[user.ID] = user
	vulnerableJSON(w, http.StatusCreated, user)
}

// vulnerableLogin authenticates fixed credentials and returns a weakly signed token.
func vulnerableLogin(w http.ResponseWriter, r *http.Request) {
	var credentials vulnerableUser
	_ = json.NewDecoder(r.Body).Decode(&credentials)
	log.Printf("login username=%q password=%q", credentials.Username, credentials.Password)

	// Hard-coded default credentials and a low-entropy signing key are deliberate.
	if credentials.Username != "admin" || credentials.Password != "admin123" {
		vulnerableJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  credentials.Username,
		"role": "admin",
		"exp":  time.Now().Add(365 * 24 * time.Hour).Unix(),
	})
	signed, _ := token.SignedString([]byte(vulnerableJWTSecret))
	vulnerableJSON(w, http.StatusOK, map[string]string{"token": signed, "api_key": vulnerableUsers[1].APIKey})
}

// vulnerableAdminReport exposes privileged data without verifying the caller's token.
func vulnerableAdminReport(w http.ResponseWriter, r *http.Request) {
	// ParseUnverified trusts attacker-supplied claims and even accepts unsigned JWTs.
	raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	token, _, _ := new(jwt.Parser).ParseUnverified(raw, jwt.MapClaims{})
	claims := interface{}(map[string]string{"role": "anonymous-but-still-authorized"})
	if token != nil {
		claims = token.Claims
	}
	vulnerableJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "confidential administrator report",
		"claims":       claims,
		"database_url": vulnerableDatabaseURL,
		"secrets":      vulnerableSecrets,
		"users":        vulnerableUsers,
	})
}

// vulnerableAdminReset restores balances without authentication or confirmation.
func vulnerableAdminReset(w http.ResponseWriter, _ *http.Request) {
	// No authentication, authorization, CSRF protection, or confirmation.
	vulnerableBalances = map[int]float64{1: 1000000, 2: 100}
	vulnerableJSON(w, http.StatusOK, map[string]string{"status": "all balances reset"})
}

// vulnerableSearch interpolates user input into a database query for SQL injection training.
func vulnerableSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	statement := "SELECT id, username, password, role, api_key FROM users WHERE username LIKE '%" + query + "%'"
	rows, err := vulnerableDB.QueryContext(r.Context(), statement) // #nosec G202 -- intentional SQL injection.
	if err != nil {
		vulnerableJSON(w, http.StatusInternalServerError, map[string]string{"query": statement, "error": err.Error()})
		return
	}
	defer rows.Close()

	users := make([]vulnerableUser, 0)
	for rows.Next() {
		var user vulnerableUser
		_ = rows.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.APIKey)
		users = append(users, user)
	}
	vulnerableJSON(w, http.StatusOK, users)
}

// vulnerableSQLConsole executes an arbitrary SQL statement from the request body.
func vulnerableSQLConsole(w http.ResponseWriter, r *http.Request) {
	statement, _ := ioutil.ReadAll(r.Body)
	result, err := vulnerableDB.ExecContext(r.Context(), string(statement))
	if err != nil {
		vulnerableJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	vulnerableJSON(w, http.StatusOK, map[string]interface{}{"statement": string(statement), "rows_affected": affected})
}

// vulnerableExec passes untrusted input to a shell command.
func vulnerableExec(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Query().Get("cmd")
	output, err := exec.Command("sh", "-c", command).CombinedOutput() // #nosec G204 -- intentional command injection.
	vulnerableJSON(w, http.StatusOK, map[string]string{"command": command, "output": string(output), "error": fmt.Sprint(err)})
}

// vulnerableReadFile returns an arbitrary file selected by the caller.
func vulnerableReadFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("path")
	contents, err := ioutil.ReadFile(name) // #nosec G304 -- intentional arbitrary file read.
	if err != nil {
		vulnerableJSON(w, http.StatusInternalServerError, map[string]string{"path": name, "error": err.Error()})
		return
	}
	vulnerableHeaders(w)
	_, _ = w.Write(contents)
}

// vulnerableUpload stores an unrestricted file under an unsanitized name.
func vulnerableUpload(w http.ResponseWriter, r *http.Request) {
	// No request-size limit, content validation, filename sanitization, or safe mode.
	file, header, err := r.FormFile("file")
	if err != nil {
		vulnerableJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	defer file.Close()

	path := "/tmp/kong-uploads/" + header.Filename
	_ = os.MkdirAll("/tmp/kong-uploads", 0777)                                      // #nosec G301 -- intentional world-writable directory.
	destination, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777) // #nosec G304,G306 -- intentional.
	if err != nil {
		vulnerableJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error(), "path": path})
		return
	}
	defer destination.Close()
	_, _ = io.Copy(destination, file)
	vulnerableJSON(w, http.StatusCreated, map[string]string{"path": path})
}

// vulnerableFetch retrieves an arbitrary URL with the default HTTP client.
func vulnerableFetch(w http.ResponseWriter, r *http.Request) {
	// The default client has no timeout and can reach loopback/cloud metadata.
	url := r.URL.Query().Get("url")
	response, err := http.Get(url) // #nosec G107 -- intentional SSRF.
	if err != nil {
		vulnerableJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	vulnerableHeaders(w)
	w.WriteHeader(response.StatusCode)
	_, _ = w.Write(body)
}

// vulnerableWelcome reflects unescaped input into an HTML response.
func vulnerableWelcome(w http.ResponseWriter, r *http.Request) {
	vulnerableHeaders(w)
	w.Header().Set("Content-Type", "text/html")
	// Raw interpolation provides a reflected XSS sink.
	_, _ = fmt.Fprintf(w, "<html><body><h1>Welcome "+r.URL.Query().Get("name")+"</h1></body></html>")
}

// vulnerableRedirect redirects to an unvalidated caller-provided destination.
func vulnerableRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.Query().Get("next"), http.StatusFound)
}

// vulnerableWeakHash returns an MD5 digest of the request body.
func vulnerableWeakHash(w http.ResponseWriter, r *http.Request) {
	secret, _ := ioutil.ReadAll(r.Body)
	sum := md5.Sum(secret) // #nosec G401 -- intentionally weak digest.
	vulnerableJSON(w, http.StatusOK, map[string]string{"md5": hex.EncodeToString(sum[:])})
}

// vulnerableTransferMoney changes balances without validating ownership or amount.
func vulnerableTransferMoney(w http.ResponseWriter, r *http.Request) {
	var transfer vulnerableTransfer
	_ = json.NewDecoder(r.Body).Decode(&transfer)
	// No authentication, ownership check, positive-amount check, locking, or transaction.
	vulnerableBalances[transfer.From] -= transfer.Amount
	vulnerableBalances[transfer.To] += transfer.Amount
	vulnerableJSON(w, http.StatusOK, vulnerableBalances)
}

// vulnerableEnvironment exposes process and request diagnostic information.
func vulnerableEnvironment(w http.ResponseWriter, r *http.Request) {
	vulnerableJSON(w, http.StatusOK, map[string]interface{}{
		"environment":      os.Environ(),
		"headers":          r.Header,
		"database_url":     vulnerableDatabaseURL,
		"embedded_secrets": vulnerableSecrets,
		"working_dir":      mustWorkingDirectory(),
	})
}

// vulnerableStack returns the current goroutine stack to any caller.
func vulnerableStack(w http.ResponseWriter, _ *http.Request) {
	vulnerableHeaders(w)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write(debug.Stack())
}

// mustWorkingDirectory returns the process working directory while ignoring errors.
func mustWorkingDirectory() string {
	directory, _ := os.Getwd()
	return directory
}
