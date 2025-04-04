package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
    ID               int64       `json:"id"`
    Username         string      `json:"username"`
    Password         string      `json:"password"`
    Email            string      `json:"email"`
    ProfilePicture    string      `json:"profile_picture"`
    Verified          bool        `json:"verified"`
    OauthUser        bool        `json:"oauth_user"`
    CreatedAt        time.Time   `json:"created_at"`
    Followers        []User      `json:"followers"`
    Followings       []User      `json:"followings"`
    FollowersCount   int64       `json:"followers_count"`
    FollowingsCount  int64       `json:"followings_count"`
    Active           bool        `json:"active"`
    LastActive       time.Time   `json:"last_active"`
    BecomingInactive bool        `json:"becoming_inactive"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword checks if the provided password matches the hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}