package handlers

import (
	"context"
	"p5.crud-structured/config"
	"p5.crud-structured/middleware"
	"p5.crud-structured/models"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Register(c echo.Context) error {
	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	_, err := config.UserCollection.InsertOne(context.TODO(), u)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "registration failed"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "registered"})
}

func Login(c echo.Context) error {
	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	var found models.User
	err := config.UserCollection.FindOne(context.TODO(), map[string]interface{}{
		"username": u.Username,
		"password": u.Password,
	}).Decode(&found)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid credentials"})
	}

	// JWT generation
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": u.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	t, err := token.SignedString(middleware.JwtSecret)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:    "token",
		Value:   t,
		Expires: time.Now().Add(24 * time.Hour),
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, echo.Map{"message": "logged in"})
}

func Logout(c echo.Context) error {
	cookie := &http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
	}
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, echo.Map{"message": "logged out"})
}
