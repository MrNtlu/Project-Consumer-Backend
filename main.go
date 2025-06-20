package main

import (
	"app/db"
	"app/docs"
	"app/helpers"
	"app/routes"
	"app/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	limit "github.com/yangxikun/gin-limit-by-key"
	"golang.org/x/time/rate"
)

// @title Watchlistfy API
// @version 1.0
// @description REST Api of Watchlistfy.
// @termsOfService https://watchlistfy.com/terms-conditions.html

// @contact.name Burak Fidan
// @contact.email mrntlu@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host http://localhost:8080
// @BasePath /api/v1
// @schemes https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if os.Getenv("ENV") != "Production" {
		if err := godotenv.Load(".env"); err != nil {
			log.Default().Println(os.Getenv("ENV"))
			log.Fatal("Error loading .env file")
		}
	}

	mongoDB, ctx, cancel := db.Connect(os.Getenv("MONGO_ATLAS_URI"))
	defer db.Close(ctx, mongoDB.Client, cancel)

	pinecone, pineconeClient, err := db.InitPinecone()
	if err != nil {
		log.Fatal("Error initializing Pinecone: ", err)
	}
	defer pinecone.Close()

	redisClient, err := db.InitRedis()
	if err != nil {
		log.Fatal("Error initializing Redis: ", err)
	}
	defer redisClient.Close()

	utils.InitCipher()

	jwtHandler := helpers.SetupJWTHandler(mongoDB)

	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC822,
		PrettyPrint:     true,
	})

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"

	const (
		burstTime       = 100 * time.Millisecond
		requestCount    = 20
		restrictionTime = 5 * time.Second
	)
	// Burst of 0.1 sec 20 requests. 5 second restriction.
	router.Use(limit.NewRateLimiter(func(ctx *gin.Context) string {
		return ctx.ClientIP()
	}, func(ctx *gin.Context) (*rate.Limiter, time.Duration) {
		return rate.NewLimiter(rate.Every(burstTime), requestCount), restrictionTime
	}, func(ctx *gin.Context) {
		const tooManyRequestError = "Too many requests. Rescricted for 5 seconds."
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": tooManyRequestError, "message": tooManyRequestError})
		ctx.Abort()
	}))

	// config := cors.DefaultConfig()
	// config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	// config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	// config.AllowHeaders = []string{"*"}
	// config.AllowCredentials = true
	// router.Use(cors.New(config))

	routes.SetupRoutes(router, jwtHandler, mongoDB, &pineconeClient, &pinecone, redisClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	if err := router.Run(":" + port); err != nil {
		panic(err)
	}
}
