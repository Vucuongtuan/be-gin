# be-gin

Backend for a blog application written in Go using the Gin framework.

## Features

- User authentication
- Blog post management
- Comments

## Running the application

### Run locally with Go

1. Clone the repository.
2. Install dependencies: `go mod tidy`
3. Run the application: `go run main.go`

The application will be available at `http://localhost:8080`.

### Run with Docker Compose (recommended)

1. Create a `.env` file at the project root (example):

   ```
   PORT=8080
   NAME_DB=be
   MONGO_INITDB_ROOT_USERNAME=root
   MONGO_INITDB_ROOT_PASSWORD=example
   DATABASE_URL="mongodb://root:example@mongo:27017/be?authSource=admin"
   ```

2. Start the services (from the project directory):

   - Run in foreground (show logs):  
     `docker compose up --build`
   - Run in background (detached):  
     `docker compose up -d --build`

3. Stop and remove containers and the MongoDB volume (if needed):  
   `docker compose down -v`

4. Access the application: `http://localhost:8080`

Notes:

- docker-compose configures the `be-gin` service (built from the Dockerfile in the project root) and a `mongo` service.
- If there is no Dockerfile yet, add a suitable Dockerfile to build the Go application.
- You may add a `.env.example` file to the repo to show required environment variables.

## Dependencies

This project uses the following main dependencies:

- Gin: https://github.com/gin-gonic/gin
- MongoDB Go Driver: https://github.com/mongodb/mongo-go-driver
- GORM: https://gorm.io/
- GoDotEnv: https://github.com/joho/godotenv
- JWT-Go: https://github.com/golang-jwt/jwt
- Gorilla WebSocket: https://github.com/gorilla/websocket
- Firebase Admin SDK for Go: https://firebase.google.com/docs/admin/setup
- Validator: https://github.com/go-playground/validator

For the full list of modules, see `go.mod`.
