package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HostAddress       string
	HostPort          string
	ControllerAddress string
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

	return &Config{
		HostAddress:       getRequiredParameter("HOST_ADDRESS"),
		HostPort:          getRequiredParameter("HOST_PORT"),
		ControllerAddress: getRequiredParameter("CONTROLLER_ADDRESS"),
	}
}

func getRequiredParameter(envParameter string) string {
	value, ok := os.LookupEnv(envParameter)
	if !ok {
		log.Fatalf("required %v environment is not set. exiting", envParameter)
	}
	return value
}
