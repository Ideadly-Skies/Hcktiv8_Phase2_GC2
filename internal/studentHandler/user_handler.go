package handler

import (
	"fmt"
	"net/http"
	config "w2/gc2/config/database"
	
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// student struct
type Student struct {
	ID int `json:"id"`
	First string `json:"first"`
	Last string `json:"last"`
	Address string `json:"address"`
	Email string `json:"email"`
	Password string `json:"password"`
	Dob       time.Time `json:"date_of_birth"`
	Jwt_token string `json:"jwt_token"`
}

// register student struct
type RegisterRequest struct {
    First    string `json:"first" validate:"required"`
    Last     string `json:"last" validate:"required"`
    Address  string `json:"address" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
    Dob      string `json:"dob" validate:"required"`
}

// login request struct
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

var jwtSecret = []byte("12345")

// @Summary Register a new student
// @Description Register a new student with personal information and credentials
// @Tags Students
// @Accept json
// @Produce json
// @Param student body RegisterRequest true "Student Registration Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /students/register [post]
func Register(c echo.Context) error {
    var req RegisterRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Request"})
    }

    // Hash the password
    hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
    }

    // Insert into the database
    query := "INSERT INTO students (first_name, last_name, email, address, date_of_birth, password_hash) VALUES ($1, $2, $3, $4, $5, $6) returning id"
    var userID int
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = config.Pool.QueryRow(ctx, query, req.First, req.Last, req.Email, req.Address, req.Dob, string(hashPassword)).Scan(&userID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // Unique violation (email already registered)
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Email already registered"})
			}
		}
		fmt.Printf("Error executing query: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	}

    return c.JSON(http.StatusOK, map[string]interface{}{
        "message": "User registered successfully",
        "user_id": userID,
        "email": req.Email,
    })
}

// @Summary Login as a student
// @Description Authenticate a student and return a JWT token
// @Tags Students
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Student Login Request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /students/login [post]
func Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message":"Invalid Request"})
	}
	
	var student Student
	query := "SELECT id, email, password_hash FROM students WHERE email = $1"
	err := config.Pool.QueryRow(context.Background(), query, req.Email).Scan(&student.ID, &student.Email, &student.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Email or Password"})
	}

	// compare password to see if it matches the student password provided
	if err := bcrypt.CompareHashAndPassword([]byte(student.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Email or Password"})
	}

	// create new jwt claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": student.ID,
		"exp":     jwt.NewNumericDate(time.Now().Add(72 * time.Hour)), // Use `jwt.NewNumericDate` for expiry
	})
	
	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid Generate Token"})
	}

	fmt.Println(student.ID)

	// update jwt token column in supabase
	updateQuery := "UPDATE students SET jwt_token = $1 WHERE id = $2"
	_, err = config.Pool.Exec(context.Background(), updateQuery, tokenString, student.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to Update Token"})
	}

	// return ok status and login response
	return c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
}

// @Summary Get student details
// @Description Get the details of the currently authenticated student
// @Tags Students
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /students/me [get]
// @Security Bearer
func GetStudentDetails(c echo.Context) error {
	// Extract the user ID from the JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims) 
	
	// safely convert user id into string
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid user_id in token"})
	}
	userID := int32(userIDFloat)

	fmt.Printf("userId from get student details %d\n",userID)

	// Query to get student details
	var student Student
	query := "SELECT first_name, last_name, address, date_of_birth FROM students WHERE id = $1"
	err := config.Pool.QueryRow(context.Background(), query, userID).Scan(&student.First, &student.Last, &student.Address, &student.Dob)
	if err != nil {
		fmt.Printf("Error retrieving student details: %v\n", err)
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Student not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve student details"})
	}
	
	// Query to get enrolled courses
	type Course struct {
		Name          string `json:"name"`
		EnrollmentDate time.Time `json:"enrollment_date"`
	}

	var courses []Course
	courseQuery := `
		SELECT c.name, e.enrollment_date
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1
	`
	rows, err := config.Pool.Query(context.Background(), courseQuery, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve courses"})
	}
	defer rows.Close()

	for rows.Next() {
		var course Course
		err := rows.Scan(&course.Name, &course.EnrollmentDate)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to parse courses"})
		}
		courses = append(courses, course)
	}

	// Combine the data into a single response
	response := map[string]interface{}{
		"full_name":     fmt.Sprintf("%s %s", student.First, student.Last),
		"address":       student.Address,
		"date_of_birth": student.Dob,
		"courses":       courses,
	}

	return c.JSON(http.StatusOK, response)
}

// @Summary Get all courses
// @Description Retrieve all available courses
// @Tags Courses
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /courses [get]
func GetCourses(c echo.Context) error {
	var courses []struct {
		ID    int    `json:"id"`
		Name string `json:"name"`
	}
	query := "SELECT id, name FROM courses"

	rows, err := config.Pool.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("Error retrieving courses: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve courses"})
	}
	defer rows.Close()

	for rows.Next() {
		var course struct {
			ID    int    `json:"id"`
			Name string `json:"name"`
		}
		err := rows.Scan(&course.ID, &course.Name)
		if err != nil {
			fmt.Printf("Error scanning course row: %v\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to parse courses"})
		}
		courses = append(courses, course)
	}

	return c.JSON(http.StatusOK, courses)
}

// @Summary Enroll in a course
// @Description Enroll a student in a specific course
// @Tags Enrollments
// @Accept json
// @Produce json
// @Param request body struct{CourseID int} true "Enrollment Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enrollments [post]
// @Security Bearer
func EnrollInCourse(c echo.Context) error {
	// Extract student ID from the JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := int32(claims["user_id"].(float64))

	// Parse request body
	var req struct {
		CourseID int `json:"course_id" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	// Check if the student is already enrolled in the course
	var exists bool
	checkQuery := "SELECT EXISTS (SELECT 1 FROM enrollments WHERE student_id = $1 AND course_id = $2)"
	err := config.Pool.QueryRow(context.Background(), checkQuery, userID, req.CourseID).Scan(&exists)
	if err != nil {
		fmt.Printf("Error checking enrollment: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to check enrollment"})
	}
	if exists {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Student already enrolled in this course"})
	}

	// Enroll the student in the course
	var enrollmentID int
	
	insertQuery := `
		INSERT INTO enrollments (student_id, course_id, enrollment_date)
		VALUES ($1, $2, CURRENT_DATE)
		RETURNING id
	`
	err = config.Pool.QueryRow(context.Background(), insertQuery, userID, req.CourseID).Scan(&enrollmentID)
	if err != nil {
		fmt.Printf("Error enrolling in course: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to enroll in course"})
	}

	// Get course details for response
	var course struct {
		Name          string `json:"name"`
		EnrollmentDate time.Time `json:"enrollment_date"`
	}

	courseQuery := `
		SELECT c.name, e.enrollment_date
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.id = $1
	`
	err = config.Pool.QueryRow(context.Background(), courseQuery, enrollmentID).Scan(&course.Name, &course.EnrollmentDate)
	if err != nil {
		fmt.Printf("Error retrieving enrollment details: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve enrollment details"})
	}

	return c.JSON(http.StatusOK, course)
}

// @Summary Delete enrollment
// @Description Remove a student's enrollment in a specific course
// @Tags Enrollments
// @Produce json
// @Param id path int true "Enrollment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enrollments/{id} [delete]
// @Security Bearer
func DeleteEnrollment(c echo.Context) error {
	// Extract enrollment ID from the URL
	enrollmentID := c.Param("id")

	// Check if the enrollment exists
	var exists bool
	checkQuery := "SELECT EXISTS (SELECT 1 FROM enrollments WHERE id = $1)"
	err := config.Pool.QueryRow(context.Background(), checkQuery, enrollmentID).Scan(&exists)
	if err != nil {
		fmt.Printf("Error checking enrollment: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to check enrollment"})
	}
	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Enrollment not found"})
	}

	// Delete the enrollment
	deleteQuery := `
		DELETE FROM enrollments
		WHERE id = $1
		RETURNING course_id, enrollment_date
	`
	var course struct {
		CourseID       int    `json:"course_id"`
		EnrollmentDate time.Time `json:"enrollment_date"`
	}
	err = config.Pool.QueryRow(context.Background(), deleteQuery, enrollmentID).Scan(&course.CourseID, &course.EnrollmentDate)
	if err != nil {
		fmt.Printf("Error deleting enrollment: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to delete enrollment"})
	}

	return c.JSON(http.StatusOK, course)
}