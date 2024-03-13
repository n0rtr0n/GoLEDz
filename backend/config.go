package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const MAX_FRAMES_PER_SECOND = 120

type Config struct {
	HostAddress           string
	HostPort              string
	ControllerAddress     string
	TargetFramesPerSecond int
}

func loadConfig() *Config {

	// since GOLEDZ_ENV will determine which .env file gets loaded, this needs to
	// be set directly in the environment, not the .env file
	envPrefix := ""
	env := os.Getenv("GOLEDZ_ENV")
	switch env {
	case "":
		envPrefix = "development"
	}

	err := godotenv.Load(envPrefix + `.env`)
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	targetFramesPerSecondString := getRequiredParameter("TARGET_FRAMES_PER_SECOND")
	targetFramesPerSecond, err := strconv.Atoi(targetFramesPerSecondString)
	if err != nil || targetFramesPerSecond < 0 {
		log.Fatalf("invalid value type for targetFramesPerSecond from env configuration")
	}
	if targetFramesPerSecond > MAX_FRAMES_PER_SECOND {
		log.Fatalf("maximum targetFramesPerSecond value %v exceeded, got %v", MAX_FRAMES_PER_SECOND, targetFramesPerSecond)
	}

	return &Config{
		HostAddress:           getRequiredParameter("HOST_ADDRESS"),
		HostPort:              getRequiredParameter("HOST_PORT"),
		ControllerAddress:     getRequiredParameter("CONTROLLER_ADDRESS"),
		TargetFramesPerSecond: targetFramesPerSecond,
	}
}

func getRequiredParameter(envParameter string) string {
	value, ok := os.LookupEnv(envParameter)
	if !ok {
		log.Fatalf("required %v environment is not set. exiting", envParameter)
	}
	return value
}
