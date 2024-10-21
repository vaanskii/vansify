# Vansify

This is a Vansify API built with Go and Gin. It provides endpoints for user registration, authentication, and managing user profiles, including following and unfollowing functionality.

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
The API will be accessible at http://localhost:8080.
```

### Database Setup
This project uses MySQL as its database. Ensure you have MySQL installed and running. Create a new database and update the database connection settings in your environment variables.

### Running Migrations
We use the `migrate` tool to manage database migrations. The Makefile simplifies running these migrations.

#### Commands:
- To create a new migration:

``` 
migrate create -ext sql -dir 'your-directory-folder-name' 'table-name'
```

- To run migrations:
```
make migrate
```
- To apply up migrations:
``` 
make up
```
- To roll back migrations:
```
make down
```


### Authentication
For endpoints that require authentication, include the token in the Authorization header using the Bearer scheme:
#### Header:
`Key`: Authorization -
`Value`: Bearer yourJWTtoken

## Version 1 Endpoints
### Authorization Routes

- **POST** `/v1/register`: Register a new user.

- **POST** `/v1/login`: Log in an existing user. Accepts an optional remember_me field to generate a long-lived token.

- **GET** `/v1/verify`: Verify user email which will be sent to your email.

- **DELETE** `/v1/delete-account`: Delete a user account.

- **POST** `/v1/forgot-password`: Send a password reset email.

- **POST** `/v1/reset-password`: Reset the user’s password using a token.


### Follow/Unfollow System Routes

- **POST** `/v1/follow/:username`: Follow a user.

- **DELETE** `/v1/unfollow/:username`: Unfollow a user.


### Chat Routes

- **POST** `/v1/create-chat`: Create a new chat.

- **GET** `/v1/chat/:chatID`: Connect to a chat WebSocket.

- **GET** `/v1/chat/:chatID/history`: Get chat history.


### User Profile Retrieval
- **GET** `/v1/me`: Get current user profile.

- **GET** `/v1/me/chats`: Get chats for the current user.

- **GET** `/v1/user/:username`: Get user profile by username.

### Technologies Used
- **Go**: The programming language used for the API.

- **Gin**: The web framework used for building the API.


### Contributing
- Contributions are welcome! Please fork the repository and create a pull request.

### Testing the API
Here’s how to test the API endpoints using Postman or any API testing tool:

### Authorization Routes
- **POST** `/v1/register`
```
{
    "username": "yourUsername",
    "password": "yourPassword",
    "email": "yourEmail@example.com"
}
```
- **POST** `/v1/login`
```
{
    "username": "yourUsername",
    "password": "yourPassword",
    "remember_me": true or false
}
```
- **POST•• `/v1/forgot-password`
```
{
    "email": "yourEmail@example.com"
}
```

- **POST** `/v1/reset-password`
```
{
    "token": "yourResetToken",
    "new_password": "yourNewPassword"
}

```
