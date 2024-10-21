# Vansify

This is a Vansify API built with Go and Gin. It provides endpoints for user registration, authentication, and managing user profiles, including following and unfollowing functionality.

### Table of Contents
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
  - [Version 1 Endpoints](#version-1-endpoints)
    - [Authorization Routes](#authorization-routes)
    - [Follow/Unfollow System Routes](#followunfollow-system-routes)
    - [User Profile Retrieval](#user-profile-retrieval)
- [Technologies Used](#technologies-used)
- [Contributing](#contributing)

### Getting Started

To run the API locally, ensure you have Go installed on your machine. Clone the repository and navigate into the project directory. Use the following commands to install dependencies and run the server:

```bash
go mod tidy
go run main.go
The API will be accessible at http://localhost:8080.
```

### Version 1 Endpoints
#### Authorization Routes
```bash
• POST /v1/register: Register a new user.

• POST /v1/login: Log in an existing user. Accepts an optional remember_me field to generate a long-lived token.

• GET /v1/verify: Verify user email which will be sent to your email.

• DELETE /v1/delete-account: Delete a user account.

• POST /v1/forgot-password: Send a password reset email.

• POST /v1/reset-password: Reset the user’s password using a token.

```
#### Follow/Unfollow System Routes
```bash
• POST /v1/follow/:username: Follow a user.

• DELETE /v1/unfollow/:username: Unfollow a user.
```

#### Chat Routes
```bash
• POST /v1/create-chat: Create a new chat.

• GET /v1/chat/:chatID: Connect to a chat WebSocket.

• GET /v1/chat/:chatID/history: Get chat history.
```

#### User Profile Retrieval
```bash
• GET /v1/me: Get current user profile.

• GET /v1/me/chats: Get chats for the current user.

• GET /v1/user/:username: Get user profile by username.
```

#### Technologies Used
• Go: The programming language used for the API.

• Gin: The web framework used for building the API.


### Summary of Changes:
- **Authorization Routes**: Added details about the `remember_me` functionality.