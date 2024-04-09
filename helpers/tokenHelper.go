package helper

import (
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UID       string
	jwt.StandardClaims
}

var SecretKey string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UID:       uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	log.Println("Fetched secret key is: ", SecretKey)

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SecretKey))
	if err != nil {
		log.Panic(err)
		return
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SecretKey))
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}
