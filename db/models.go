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

func Register(username, password, email string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	row := db.QueryRow("INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3) RETURNING id", username, string(hash), email)
	var user User
	err = row.Scan(&user.ID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}