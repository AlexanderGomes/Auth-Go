package controllers

import (
	"auth-go/database"
	"auth-go/schemas"
	"context"
	"log"
	"os"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func LoadKey() string {
	envError := godotenv.Load()

	if envError != nil {
		log.Fatal(".env file couldn't be loaded")
	}

	secretKey := os.Getenv("JWT_SECRET")

	return secretKey
}

var userCollection *mongo.Collection = database.GetCollection(database.DB, "users")

func Register(c *fiber.Ctx) error {
	var data map[string]string

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := schemas.User{
		Name:     data["name"],
		Email:    data["email"],
		Password: string(password),
	}

	result, _ := userCollection.InsertOne(ctx, user)

	return c.JSON(result)
}

func Login(c *fiber.Ctx) error {
	secretKey := LoadKey()
	var data map[string]string

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	var user schemas.User

	filter := bson.M{"email": data["email"]}

	if userNotFound := userCollection.FindOne(ctx, filter).Decode(&user); userNotFound != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	mismatch := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data["password"]))

	if mismatch != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "incorrect password",
		})
	}

	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Hour * 10))
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    user.Id.String(),
		ExpiresAt: expiresAt,
	})

	token, err := claims.SignedString([]byte(secretKey))

	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"message": "could not login",
		})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func GetUser(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	secretKey := LoadKey()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := jwt.ParseWithClaims(cookie, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return err
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	var user schemas.User

	filter := bson.M{"_id": claims.Issuer}

	if userNotFound := userCollection.FindOne(ctx, filter).Decode(&user); userNotFound != nil {
	  return userNotFound
	}

	return c.JSON(user)
}
