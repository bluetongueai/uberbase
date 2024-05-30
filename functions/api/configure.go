package main

import "os"

// use go to declare a
type Configuration struct {
	AllowedOrigins string
}

func Configure() Configuration {
	origins, _ := os.LookupEnv("UBERBASE_ALLOWED_ORIGINS")
	config := Configuration{
		AllowedOrigins: origins,
	}

	return config
}
