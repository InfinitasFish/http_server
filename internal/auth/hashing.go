package auth

import (
	"runtime"
	"log"
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashParams := &argon2id.Params{
		Memory:      128 * 1024,
		Iterations:  4,
		Parallelism: uint8(runtime.NumCPU()),
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(password, hashParams)
	if err != nil {
		log.Println("Error while creating hash for password", err)
	}

	return hash, err
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		log.Println("Error while comparing password with hash", err)
	}

	return match, err
}

