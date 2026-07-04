module http_server

go 1.25.5

require internal/database v1.0.0

replace internal/database => ./internal/database

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.12.3 // indirect
)
