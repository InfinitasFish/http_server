package auth

import (
	"runtime"
	"time"
	"log"
	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	jwToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer: "chirpy-access",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject: userID.String(),
		})

	signed, err := jwToken.SignedString(tokenSecret)
	if err != nil {
		log.Println("Error while signing JWT token", err)
		return "", err
	}

	return signed, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	jwToken, err := jwt.ParseWithClaims(
		tokenString, 
		&jwt.RegisteredClaims{},
		// what key should I use to verify this token's signature?"
		func(token *jwt.Token) (interface{}, error) {
			return tokenSecret, nil
		})

	if err != nil {
		log.Println("Error while parsing JWT token", err)
		return uuid.Nil, err
	}

	claims, _ := jwToken.Claims.(jwt.RegisteredClaims)
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Println("Error while parsing user UUID", err)
		return uuid.Nil, err
	}

	return userID, nil
}
