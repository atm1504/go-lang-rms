package helper

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	db "atm1504.in/rms/database"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Phone     string
	jwt.StandardClaims
}

var SecretKey string = os.Getenv("SECRET_KEY")

// Function to handle database connection errors
func ISEInjection(c *gin.Context, err error, errorMessage string) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMessage, "errMsg": err.Error()})
		return true
	}
	return false
}

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GenerateAllTokens(email string, firstName string, lastName string, phone string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
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

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId int64) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	dbConn, dbErr := db.DBInstanceSql()
	if dbErr != nil {
		log.Panic(dbErr.Error())
		return
	}
	defer dbConn.Close()
	query := "UPDATE user SET updated_at=?, token=?, refresh_token=? WHERE id=?"
	values := []interface{}{time.Now(), signedToken, signedRefreshToken, userId}
	_, err := dbConn.ExecContext(ctx, query, values...)
	if err != nil {
		log.Panic(dbErr.Error())
		return
	}
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {

	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		},
	)

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "the token is invalid"
		msg = err.Error()
		return
	}
	//the token is expired
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "token is expired"
		msg = err.Error()
		return
	}
	return claims, msg
}
