package handlers

import (
	"context"
	"p5.crud-structured/config"
	"p5.crud-structured/models"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateBook(c echo.Context) error {
	b := new(models.Book)
	if err := c.Bind(b); err != nil {
		return err
	}
	_, err := config.BookCollection.InsertOne(context.TODO(), b)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "error creating book"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "book created"})
}

func GetBooks(c echo.Context) error {
	cursor, err := config.BookCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return err
	}
	var books []models.Book
	if err := cursor.All(context.TODO(), &books); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, books)
}

func UpdateBook(c echo.Context) error {
	id := c.Param("id")
	b := new(models.Book)
	if err := c.Bind(b); err != nil {
		return err
	}
	_, err := config.BookCollection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$set": b})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "book updated"})
}

func DeleteBook(c echo.Context) error {
	id := c.Param("id")
	_, err := config.BookCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "book deleted"})
}
