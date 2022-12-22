package helper

import (
	"context"
	"errors"
	database "gojwt/database"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Signedetails struct {
	Email      string
	First_name string
	Last_name  string
	User_type  string
	User_id    string
	jwt.StandardClaims
}

var usercollection *mongo.Collection = database.OpenCollection(database.Client, "user")

//var err error= godotenv.Load(".env")
//if err != nil {
//log.Fatal("error occured in .env file")
//}
//var SECRET_KEY string = os.Getenv("SECRET_KEY")
//var Usercollection *mongo Collection = database.OpenCollection(database.Client,"users")
func Generatealltoken(email string, firstname string, lastname string, usertype string, userid string) (string, string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error occured in .env file")
	}
	var SECRET_KEY string = os.Getenv("SECRET_KEY")
	claims := &Signedetails{
		Email:      email,
		First_name: firstname,
		Last_name:  lastname,
		User_type:  usertype,
		User_id:    userid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	refreshClaims := &Signedetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return "", "", err
	}
	refreshtoken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return "", "", err
	}
	return token, refreshtoken, err

}
func ValidateToken(SignedToken string) (*Signedetails, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error occured in .env file")
	}
	var SECRET_KEY string = os.Getenv("SECRET_KEY")
	token, err := jwt.ParseWithClaims(
		SignedToken,
		&Signedetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Signedetails)
	if !ok {

		err = errors.New("token invalid")
		return nil, err
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return nil, err
	}
	return claims, err

}
func UpdateAllTokens(Signedtoken string, Signedrefreshtoken string, Userid string) {
	//var updateobj primitive.D
	//updateobj = append(updateobj, bson.D{primiteve.E{Key:"token",Value:Signedtoken}})
	//updateobj = append(updateobj, bson.E{"refresh_token", Signedrefreshtoken})
	updatedat, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	//updateobj = append(updateobj, bson.E{"updated_at", updatedat})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	filter := bson.M{"user_id": Userid}
	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := usercollection.UpdateOne(
		ctx,
		filter,
		bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "token", Value: Signedtoken},
			{Key: "refresh_token", Value: Signedrefreshtoken},
			{Key: "updated_at", Value: updatedat}}}},
		&opt,
	)
	defer cancel()
	if err != nil {
		log.Panic(err)
		return
	}
}
