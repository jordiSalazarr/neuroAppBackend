package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
	EnvTest  = "test"
)

type AppConfig struct {
	OpenAIKey     string
	SMTP_HOST     string
	SMTP_PORT     int
	SMTP_USERNAME string
	SMTP_PASSWORD string
	SMTP_FROM     string
}

func GetConfig() *AppConfig {
	err := godotenv.Load(".env.local")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		fmt.Println("Error parsing SMTP_PORT:", err)
		port = 587 // Default SMTP port
	}
	return &AppConfig{
		SMTP_FROM:     os.Getenv("SMTP_FROM"),
		SMTP_HOST:     os.Getenv("SMTP_HOST"),
		SMTP_USERNAME: os.Getenv("SMTP_USERNAME"),
		SMTP_PASSWORD: os.Getenv("SMTP_PASSWORD"),
		SMTP_PORT:     port,
		OpenAIKey:     os.Getenv("OPENAI_API_KEY"),
	}
}
func GetS3BucketName() (string, error) {
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		return "", errors.New("S3_BUCKET_NAME environment variable is not set")
	}
	return bucketName, nil
}

func GetCurrentEnvironment() string {
	switch os.Getenv("environment") {
	case EnvLocal:
		return EnvLocal
	case EnvDev:
		return EnvDev
	case EnvProd:
		return EnvProd
	case EnvTest:
		return EnvTest
	default:
		os.Setenv("environment", EnvLocal)
	}
	return EnvLocal
}

// func GetProdConfig() AppConfig {
// 	return AppConfig{
// 		//Logger: , TODO: implementLogger
// 		OpenAIKey: os.Getenv("OPENAI_API_KEY"),
// 	}
// }

// func GetDevConfig() AppConfig {
// 	return AppConfig{
// 		//Logger: , TODO: implementLogger
// 		OpenAIKey: os.Getenv("OPENAI_API_KEY"),
// 	}
// }

// func GetTestConfig() AppConfig {
// 	return AppConfig{
// 		//Logger: , TODO: implementLogger
// 		OpenAIKey: os.Getenv("OPENAI_API_KEY"),
// 	}
// }
