-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAllChrips :many
SELECT * FROM chirps;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1
LIMIT 1;
