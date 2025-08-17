package authentication

import (
	"encoding/json"
	"fmt"
	"net/http"

	db_tools "github.com/samdandy/go_card_api/internal/tools"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
	Password string
}

type UserSignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func UserSignup(w http.ResponseWriter, r *http.Request, db *db_tools.PGDatabase) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UserSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	signUpUser := &User{
		Username: req.Username,
		Password: string(hashed),
	}
	exists, err := db.CheckUserExists(signUpUser.Username)
	if err != nil {
		http.Error(w, "error checking user existence", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "username already taken", http.StatusConflict)
		return
	}

	_, err = db.CreateUser(signUpUser.Username, signUpUser.Password)
	fmt.Println(err)
	if err != nil {
		http.Error(w, "error creating user", http.StatusInternalServerError)
		return
	}
	fmt.Println("User created:", signUpUser.Username)
}

func UserLogin(w http.ResponseWriter, r *http.Request, db *db_tools.PGDatabase) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	exists, err := db.CheckUserExists(req.Username)
	if err != nil {
		http.Error(w, "error checking user existence", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Here you would typically compare the hashed password with the stored hash
	// For simplicity, we're just going to return a success message
	fmt.Println("User logged in:", req.Username)
	w.WriteHeader(http.StatusOK)
}
