package db

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Count        int   `json:"count"`
}

func Authenticate(username, password string) (*User, error) {
	row := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username=$1", username)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Println("Error getting user by username:", err)
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func Register(username, password, email, name string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	row := db.QueryRow("INSERT INTO users (username, password_hash, email, name, count) VALUES ($1, $2, $3, $4, $5) RETURNING id", username, string(hash), email, name, 0)
	var user User
	err = row.Scan(&user.ID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// function to fetch matching name on username field
func FetchDetails(username string) (*User, error) {
	var user User
	row := db.QueryRow("SELECT name, email, count FROM users WHERE username=$1", username)
	err := row.Scan(&user.Name, &user.Email, &user.Count)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no user found with username: %s", username)
		}
		return nil, err
	}

	return &user, nil
}

func UpdateCount(username string) error {
	_, err := db.Exec("UPDATE users SET count = count + 1 WHERE username=$1", username)
	if err != nil {
		return err
	}
	return nil
}