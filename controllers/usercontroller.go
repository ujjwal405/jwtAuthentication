package controllers

import (
	"context"
	database "gojwt/database"
	helper "gojwt/helpers"
	models "gojwt/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var Usercollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(userpassword string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(userpassword), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	check := true
	msg := ""
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	if err != nil {
		msg = "password incorect"
		check = false
	}
	return check, msg
}
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		var founduser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		err := Usercollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email doesn't exist"})
			return
		}
		passwordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
		var val bool = true
		if passwordIsValid != val {
			c.JSON(http.StatusInternalServerError, gin.H{"err": msg})
			return
		}
		defer cancel()
		token, refreshtoken, err := helper.Generatealltoken(*founduser.Email, *founduser.First_name, *founduser.Last_name, *founduser.User_type, founduser.User_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}
		helper.UpdateAllTokens(token, refreshtoken, founduser.User_id)
		err = Usercollection.FindOne(ctx, bson.M{"user_id": founduser.User_id}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}
		c.JSON(http.StatusOK, founduser)

	}

}
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validateerr := validate.Struct(user)
		if validateerr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validateerr.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		count, err := Usercollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		password := HashPassword(*user.Password)
		user.Password = &password

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking"})

		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "This email already exist"})
			return

		}
		count, err = Usercollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for phone number"})

		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "This phone number  already exist"})
			return
		}
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.Generatealltoken(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken
		insertnumber, inserterr := Usercollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user was not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, insertnumber)
	}
}
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.Checkusertype(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordperpage, err := strconv.Atoi(c.Query("recordperpage"))
		if err != nil || recordperpage < 1 {
			recordperpage = 10
		}
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		var startindex int = (page - 1) * recordperpage
		startindex, _ = strconv.Atoi(c.Query("startindex"))
		matchstage := bson.D{primitive.E{Key: "$match", Value: bson.D{primitive.E{}}}}

		groupstage := bson.D{primitive.E{Key: "$group", Value: bson.D{
			primitive.E{Key: "_id", Value: bson.D{primitive.E{Key: "_id", Value: "null"}}},
			{Key: "totalcount", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{primitive.E{Key: "$push", Value: "$$ROOT"}}},
		}}}

		projectstage := bson.D{primitive.E{Key: "$project", Value: bson.D{primitive.E{Key: "_id", Value: 0},
			{Key: "totalcount", Value: 1},
			{Key: "userdata", Value: bson.D{primitive.E{Key: "$slice", Value: []interface{}{"$data", startindex, recordperpage}}}},
		}}}

		result, err := Usercollection.Aggregate(ctx, mongo.Pipeline{
			matchstage, groupstage, projectstage,
		})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": "error occured while making list of user"})

		}
		var allusers []bson.M
		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allusers[0])
	}
}
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Param("user_id")
		if err := helper.Matchusertypeid(c, userid); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := Usercollection.FindOne(ctx, bson.M{"user_id": userid}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
