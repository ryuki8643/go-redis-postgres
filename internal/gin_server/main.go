package gin_server

import (
	"github.com/geolocket/batch_redis/internal/momentType"
	"github.com/geolocket/batch_redis/internal/postgres"
	"github.com/geolocket/batch_redis/internal/redis_server"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"time"

	_ "github.com/geolocket/batch_redis/internal/gin_server/docs"
)

type Message struct {
	Message string `json:"message"`
}

// Hello ...
// @Summary helloを返す
// @Tags helloWorld
// @Produce  json
// @Success 200 {object} Message
// @Router / [get]
func hello(c *gin.Context) {
	c.JSONP(http.StatusOK, gin.H{
		"message": "hello",
	})
}

// readRedis ...
// @Summary redis読み取り
// @Tags redis
// @Produce  json
// @Param user_id path int true "User ID"
// @Success 200 {object} Message
// @Failure 400 {object} Message
// @Router /{user_id} [get]
func readRedis(c *gin.Context) {

	m, err := redis_server.ReadRedis(c.Param("user_id"))
	if err != nil {
		log.Printf("%+v", err)
		c.JSONP(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSONP(http.StatusOK, gin.H{
		"message": m,
	})
}

// RegisterMessage ...
// @Summary データ送信
// @Tags redis
// @Produce  json
// @Param article_json body UserRequest true "Article Json"
// @Success 200 {object} Message
// @Failure 400 {object} Message
// @Router / [put]
func registerMessages(c *gin.Context) {
	var userRequest momentType.UserRequest
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		log.Printf("%+v", err)
		c.JSONP(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	if len(userRequest.LatLng) != 2 {
		log.Println("lat lang should be two para")
		c.JSONP(http.StatusBadRequest, gin.H{
			"message": "lat lang should be two para",
		})
		return
	}
	UserMomentWriteInQueue(userRequest)
	c.JSONP(http.StatusOK, gin.H{
		"message": "momentRegistered",
	})
}

// Swagger ...
// @Summary /swagger/index.html#/にアクセスするとswaggerを返す
// @Tags helloWorld
// @Produce  json
// @Failure 400 {object} Message
// @Router /swagger [get]
func ginSwaggerDoc() func(c *gin.Context) {
	return ginSwagger.WrapHandler(swaggerFiles.Handler)
}

func dbDelete(c *gin.Context) {
	postgres.DeleteMessages()
	c.JSONP(http.StatusOK, gin.H{
		"message": "deleted",
	})
}

// @title batch-redis
// @version 2.0
// @license.name ryuki
func NewHTTPServer() {
	go func() {
		ticker := time.NewTicker(momentType.RegisterInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if len(momentType.MomentQueue) > 0 {
					redis_server.BatchRedis(time.Now())
				}
			}
		}

	}()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowMethods: []string{
			"PUT",
			"GET",
		},
	}))

	r.GET("/", hello)
	r.GET("/delete", dbDelete)
	r.GET("/:user_id", readRedis)
	r.PUT("/", registerMessages)
	r.GET("/swagger/*any", ginSwaggerDoc())

	r.Run()
}
