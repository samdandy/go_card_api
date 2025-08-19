package authentication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	db_tools "github.com/samdandy/go_card_api/internal/tools"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type User struct {
	ID       int
	Username string
	Password string
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
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
	var creds User
	_ = json.NewDecoder(r.Body).Decode(&creds)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pw, err := db.GetUserPassword(creds.Username)
	if err != nil {
		http.Error(w, "error getting user password", http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(pw), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	// create JWT
	expiration := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Could not login", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		var tokenStr string
		fmt.Sscanf(authHeader, "Bearer %s", &tokenStr)

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
