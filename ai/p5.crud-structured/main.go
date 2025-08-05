package main

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"p5.crud-structured/config"
	"p5.crud-structured/handlers"
	"p5.crud-structured/middleware"
)

func main() {
	config.ConnectMongoDB()

	e := echo.New()

	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	e.POST("/register", handlers.Register)
	e.POST("/login", handlers.Login)
	e.GET("/logout", handlers.Logout)

	books := e.Group("/books")
	books.Use(middleware.CookieJWTMiddleware)
	books.POST("", handlers.CreateBook)
	books.GET("", handlers.GetBooks)
	books.PUT("/:id", handlers.UpdateBook)
	books.DELETE("/:id", handlers.DeleteBook)

	e.Logger.Fatal(e.Start(":8083"))
}
