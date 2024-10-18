
# This is a User Management API built with Go and Gin. It provides endpoints for user registration, authentication, and managing user profiles, including following and unfollowing functionality.

## Table of Contents
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
  - [Version 1 Endpoints](#version-1-endpoints)
    - [Authorization Routes](#authorization-routes)
    - [Follow/Unfollow System Routes](#followunfollow-system-routes)
    - [User Profile Retrieval](#user-profile-retrieval)
- [Technologies Used](#technologies-used)
- [Contributing](#contributing)

## Getting Started

To run the API locally, ensure you have Go installed on your machine. Clone the repository and navigate into the project directory. Use the following commands to install dependencies and run the server:

```bash
go mod tidy
go run main.go
```

The API will be accessible at http://localhost:8080.

# Version 1 Endpoints
```bash API Endpoints 
Authorization Routes
POST /v1/register: Register a new user.
POST /v1/login: Log in an existing user.
GET /v1/verify: Verify user email which will be sended your email.
DELETE /v1/delete-account: Delete a user account.
```
# Follow/Unfollow System Routes
```bash
POST /v1/follow/:username: Follow a user.
DELETE /v1/unfollow/:username: Unfollow a user.
User Profile Retrieval
GET /v1/user/:username: Get user profile by username.
Technologies Used
Go: The programming language used for the API.
Gin: The web framework used for building the API.
```