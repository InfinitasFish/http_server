module http_server

go 1.25.5

require internal/database v1.0.0
require internal/auth v1.0.0

replace internal/database => ./internal/database
replace internal/auth => ./internal/auth

require (
	github.com/alexedwards/argon2id v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.12.3 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
