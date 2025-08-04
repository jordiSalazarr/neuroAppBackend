package config

import (
	"errors"
	"os"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
	EnvTest  = "test"
)

type AppConfig struct {
	OpenAIKey    string
	S3BucketName string
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
