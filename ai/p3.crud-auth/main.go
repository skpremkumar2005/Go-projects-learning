package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var jwtSecret = []byte("secret-key")

type User struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

type Book struct {
	ID     string `json:"id" bson:"_id,omitempty"`
	Title  string `json:"title" bson:"title"`
	Author string `json:"author" bson:"author"`
}

var userCollection *mongo.Collection
var bookCollection *mongo.Collection

func main() {
	// Connect MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.TODO()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("echo_auth")
	userCollection = db.Collection("users")
	bookCollection = db.Collection("books")

	// Setup Echo
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/register", register)
	e.POST("/login", login)
	e.GET("/logout", logout)

	// JWT Protected
	r := e.Group("/books")
	r.Use(cookieJWTMiddleware)
	r.POST("", createBook)
	r.GET("", getBooks)
	r.PUT("/:id", updateBook)
	r.DELETE("/:id", deleteBook)

	// Start server
	e.Logger.Fatal(e.Start(":8083"))
}

func register(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}

	_, err := userCollection.InsertOne(context.TODO(), u)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "registration failed"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "registered"})
}

func login(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}

	var found User
	err := userCollection.FindOne(context.TODO(), bson.M{"username": u.Username, "password": u.Password}).Decode(&found)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid credentials"})
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": u.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return err
	}

	// Set cookie
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = t
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, echo.Map{"message": "logged in"})
}

func logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, echo.Map{"message": "logged out"})
}

func cookieJWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("token")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "missing token"})
		}
		tokenStr := cookie.Value

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid token"})
		}
		return next(c)
	}
}

func createBook(c echo.Context) error {
	b := new(Book)
	if err := c.Bind(b); err != nil {
		return err
	}
	_, err := bookCollection.InsertOne(context.TODO(), b)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "error creating book"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "book created"})
}

func getBooks(c echo.Context) error {
	cursor, err := bookCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return err
	}
	var books []Book
	if err := cursor.All(context.TODO(), &books); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, books)
}
func updateBook(c echo.Context) error {
	id := c.Param("id")
	b := new(Book)
	if err := c.Bind(b); err != nil {
		return err
	}
	_, err := bookCollection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$set": b})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "book updated"})
}

func deleteBook(c echo.Context) error {
	id := c.Param("id")
	_, err := bookCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "book deleted"})
}
