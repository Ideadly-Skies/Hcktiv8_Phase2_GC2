package main

import (
	config "w2/gc2/config/database"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	// "github.com/labstack/echo/v4/middleware"

	student_handler "w2/gc2/internal/studentHandler"
	cust_middleware "w2/gc2/internal/middleware"
	"github.com/swaggo/echo-swagger"
	_ "w2/gc2/cmd/docs"                              
)

func main(){
	// migrate data
	// config.MigrateData()

	// connect to db 
	config.InitDB()
	defer config.CloseDB()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// public routes
	e.POST("students/register", student_handler.Register)	
	e.POST("students/login", student_handler.Login)
	
	// protected route
	e.GET("students/me", student_handler.GetStudentDetails, cust_middleware.JWTMiddleware)
	e.GET("courses", student_handler.GetCourses, cust_middleware.JWTMiddleware)
	e.POST("enrollments", student_handler.EnrollInCourse, cust_middleware.JWTMiddleware)
	e.DELETE("enrollments/:id", student_handler.DeleteEnrollment, cust_middleware.JWTMiddleware)

	// Swagger endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Logger.Fatal(e.Start(":8080"))
}