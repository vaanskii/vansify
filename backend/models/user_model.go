package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID 		   		 int64  	 `json:"id"`
	Username   		 string 	 `json:"username"`
	Password   		 string 	 `json:"password"`
	Email 	   		 string 	 `json:"email"`
	ProfilePicture    string      `json:"profile_picture"`
	Gender 			 string   	 `json:"gender"`
	Verified 		  bool 	  	  `json:"verified"`
	CreatedAt  		 time.Time   `json:"created_at"`
	Followers        []User      `json:"followers"`
	Followings       []User      `json:"followings"`
	FollowersCount   int64       `json:"followers_count"`
	FollowingsCount  int64       `json:"followings_count"`
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