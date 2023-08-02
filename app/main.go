package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/linqcod/postsRestAPI/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
)

var (
	server      *gin.Engine
	ctx         context.Context
	mongoClient *mongo.Client
	redisClient *redis.Client
)

func init() {
	//Loading config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("could not load environment variables: ", err)
	}

	ctx = context.TODO()

	//Connecting to mongo db
	mongoConn := options.Client().ApplyURI(cfg.DBUri)
	mongoClient, err = mongo.Connect(ctx, mongoConn)
	if err != nil {
		panic(err)
	}

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println("MongoDB successfully connected...")

	//Connecting to redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: cfg.RedisUri,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	err = redisClient.Set(ctx, "test", "Welcome to Golang Redis and MongoDB", 0).Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("Redis client connected successfully...")

	//Initializing gin server
	server = gin.Default()
}

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("could not load config: ", err)
	}

	defer mongoClient.Disconnect(ctx)

	value, err := redisClient.Get(ctx, "test").Result()
	if err == redis.Nil {
		fmt.Println("key: test does not exist")
	} else if err != nil {
		panic(err)
	}

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": value,
		})
	})

	log.Fatal(server.Run(":" + config.Port))
}
