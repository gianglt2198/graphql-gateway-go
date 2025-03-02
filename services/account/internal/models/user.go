package models

import (
	"fmt"
	"time"

	"github.com/gianglt2198/graphql-gateway-go/services/account/graph/model"
)

// For simplicity, we'll use an in-memory store
var Users = []*model.User{
	{
		ID:        "1",
		UserName:  "admin",
		Email:     "admin@example.com",
		FullName:  "Admin User",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	},
	{
		ID:        "2",
		UserName:  "johndoe",
		Email:     "john@example.com",
		FullName:  "John Doe",
		CreatedAt: time.Now().Add(-48 * time.Hour),
	},
}

// Helper functions for our mock database
func GetUserByID(id string) *model.User {
	for _, user := range Users {
		if user.ID == id {
			return user
		}
	}
	return nil
}

func GetUsers() []*model.User {
	return Users
}

func CreateUser(username, email, fullName string) *model.User {
	user := &model.User{
		ID:        fmt.Sprintf("%d", len(Users)+1),
		UserName:  username,
		Email:     email,
		FullName:  fullName,
		CreatedAt: time.Now(),
	}
	Users = append(Users, user)
	return user
}

func UpdateUser(id, username, email, fullName string) *model.User {
	user := GetUserByID(id)
	if user == nil {
		return nil
	}

	if username != "" {
		user.UserName = username
	}
	if email != "" {
		user.Email = email
	}
	if fullName != "" {
		user.FullName = fullName
	}

	return user
}

func DeleteUser(id string) bool {
	for i, user := range Users {
		if user.ID == id {
			// Remove the user
			Users = append(Users[:i], Users[i+1:]...)
			return true
		}
	}
	return false
}
