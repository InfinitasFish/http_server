-- name: ResetAllUsers :exec
DELETE FROM users;

-- name: ResetAllChirps :exec
DELETE FROM chirps;