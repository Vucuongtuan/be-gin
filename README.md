# be-gin

Backend for a blog application written in Go using the Gin framework.

## Features

- User authentication
- Blog post management
- Comments
- GraphQL API for blogs

## Running the application

### Prerequisites

- Go
- Docker
- Docker Compose

### Instructions

1.  Clone the repository.
2.  Install dependencies: `go mod tidy`
3.  Run the application: `go run main.go`

The application will be available at `http://localhost:8080`.

## Dependencies

This project uses the following main dependencies:

- **[Gin](https://github.com/gin-gonic/gin):** A popular web framework for Go.
- **[Mongo Driver](https://github.com/mongodb/mongo-go-driver):** The official Go driver for MongoDB.
- **[GORM](https://gorm.io/):** A developer-friendly ORM library for Go.
- **[GoDotEnv](https://github.com/joho/godotenv):** A library to load environment variables from a `.env` file.
- **[JWT-Go](https://github.com/golang-jwt/jwt):** A Go implementation of JSON Web Tokens (JWT).
- **[Gorilla WebSocket](https://github.com/gorilla/websocket):** A Go implementation of the WebSocket protocol.
- **[Firebase Admin SDK](https://firebase.google.com/docs/admin/setup):** The Firebase Admin SDK for Go.
- **[Validator](https://github.com/go-playground/validator):** A Go library for struct and field validation.

For a full list of dependencies, please refer to the `go.mod` file.
