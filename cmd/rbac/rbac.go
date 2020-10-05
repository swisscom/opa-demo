package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Employee struct {
	Name string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
	Salary float64 `json:"salary,omitempty" binding:"required"`
}

type Role string

const RoleAdmin Role = "ROLE_ADMIN"
const RoleHR Role = "ROLE_HR"
const RoleUser Role = "ROLE_EMPLOYEE"

type User struct {
	Name string
	Role Role
}

var db map[string]Employee
var userDb map[string]User

func setupRouter() *gin.Engine {
	r := gin.Default()

	db = map[string]Employee{}
	userDb = map[string]User{
		"admin": {
			Name: "Admin",
			Role: RoleAdmin,
		},
		"hr": {
			Name: "HR Employee",
			Role: RoleHR,
		},
		"user": {
			Name: "User 1",
			Role: RoleUser,
		},

	}

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"admin":  "admin",
		"hr": "hr",
		"user": "user",
	}))

	// Get employee
	authorized.GET("/employee/:name", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		userFromDb, ok := userDb[user]


		employee := c.Params.ByName("name")
		value, ok := db[employee]

		if ok {
			if userFromDb.Role != RoleAdmin && userFromDb.Role != RoleHR {
				value.Salary = -1
			}
			c.JSON(http.StatusOK, value)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"status": "not found"})
		}
	})

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		userFromDb := userDb[user]
		json := Employee{}

		if userFromDb.Role != RoleAdmin {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		}

		if c.Bind(&json) == nil {
			db[json.Name] = json
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}