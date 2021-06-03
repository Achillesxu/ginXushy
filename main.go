// Package ginXushy
// Time    : 2021/5/26 7:59 下午
// Author  : xushiyin
// contact : yuqingxushiyin@gmail.com
package main

import (
	"fmt"
	_ "github.com/Achillesxu/ginXushy/docs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/thinkerou/favicon"
	"log"
	"net/http"
	"time"
)

var db = make(map[string]string)

// Login Binding from JSON
type Login struct {
	User     string `form:"user" json:"user" xml:"user"  binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

// Booking contains binded and validated data.
type Booking struct {
	CheckIn  time.Time `form:"check_in" binding:"required,bookableDate" time_format:"2006-01-02"`
	CheckOut time.Time `form:"check_out" binding:"required,gtfield=CheckIn" time_format:"2006-01-02"`
}

var bookableDate = func(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if ok {
		today := time.Now()
		if today.After(date) {
			return false
		}
	}
	return true
}

func addXuRoutes(rg *gin.RouterGroup) {
	xu := rg.Group("xu")
	xu.GET("/comments", func(c *gin.Context) {
		c.JSON(http.StatusOK, "users comments")
	})
	xu.GET("/pictures", func(c *gin.Context) {
		c.JSON(http.StatusOK, "users pictures")
	})
}

func getBookable(c *gin.Context) {
	var b Booking
	if err := c.ShouldBindWith(&b, binding.Query); err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Booking dates are valid!"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /v2
func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()
	r.Use(favicon.New("./favicon.ico"))

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("bookableDate", bookableDate)
		if err != nil {
			return nil
		}
	}

	addXuRoutes(r.Group("v1"))

	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json") // The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ping": "pong"})
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "name": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth middleware)
	authorized := r.Group("/a", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		var json struct {
			Value string `json:"value" binding:"required"`
		}
		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	authorized.GET("comments", func(c *gin.Context) {
		c.JSON(http.StatusOK, "users comments")
	})
	authorized.GET("pictures", func(c *gin.Context) {
		c.JSON(http.StatusOK, "users pictures")
	})

	r.POST("/loginJson", func(c *gin.Context) {
		var j Login
		if err := c.ShouldBindJSON(&j); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if j.User != "manu" || j.Password != "123" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
	})

	// Example for binding a HTML form (user=manu&password=123)
	r.POST("/loginForm", func(c *gin.Context) {
		var form Login
		// This will infer what binder to use depending on the content-type header.
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if form.User != "manu" || form.Password != "123" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
	})

	// upload single file
	r.POST("/upload_single", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		log.Println(file.Filename)
		// Upload the file to specific dst.
		err := c.SaveUploadedFile(file, "single.txt")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": fmt.Sprintf("'%s' uploaded!", file.Filename)})

	})
	// Custom Validators
	r.GET("/bookable", getBookable)

	r.GET("/someJSON", func(c *gin.Context) {
		names := []string{"lena", "austin", "foo"}

		// Will output  :   while(1);["lena","austin","foo"]
		c.SecureJSON(http.StatusOK, names)
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	log.Fatal(r.Run(":8080"))
}
