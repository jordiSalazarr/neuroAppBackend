package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"neuro.app.jordi/internal/api"
	"neuro.app.jordi/internal/shared/mysql"
)

type Config struct {
	Port string `envconfig:"PORT" default:"8080"`
	Env  string `envconfig:"ENV" default:"development"`
}

func main() {
	db, err := mysql.NewMySQL()
	if err != nil {
		log.Fatal(err)
	}

	router := api.NewApp(db).SetupRouter()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}

	srv := &http.Server{
		Addr:    ":8401",
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on port 8401")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %s", err)
	}
	log.Println("Server exited cleanly")
}
