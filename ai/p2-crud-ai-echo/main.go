package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Book struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title  string             `json:"title" bson:"title"`
	Author string             `json:"author" bson:"author"`
}

var bookCollection *mongo.Collection
var ctx = context.Background()

func connectDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	bookCollection = client.Database("library").Collection("books")
	fmt.Println("✅ MongoDB connected")
}

// POST /books
func createBook(c echo.Context) error {
	var book Book
	if err := c.Bind(&book); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	res, err := bookCollection.InsertOne(ctx, book)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, res)
}

// GET /books
func getBooks(c echo.Context) error {
	cursor, err := bookCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer cursor.Close(ctx)

	var books []Book
	for cursor.Next(ctx) {
		var book Book
		cursor.Decode(&book)
		books = append(books, book)
	}
	return c.JSON(http.StatusOK, books)
}

// PUT /books/:id
func updateBook(c echo.Context) error {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID")
	}

	var book Book
	if err := c.Bind(&book); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	update := bson.M{"$set": book}
	_, err = bookCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Book updated")
}

// DELETE /books/:id
func deleteBook(c echo.Context) error {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID")
	}
	_, err = bookCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Book deleted")
}

func main() {
	connectDB()
	e := echo.New()

	e.POST("/books", createBook)
	e.GET("/books", getBooks)
	e.PUT("/books/:id", updateBook)
	e.DELETE("/books/:id", deleteBook)

	fmt.Println("🚀 Server started at http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
