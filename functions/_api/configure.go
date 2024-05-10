package main

// use go to declare a 
type Configuration {
	AllowedOrigins string
}

func Configure() Configuration {
	config := Configuration{}

	config.AllowedOrigins, = := os.LookupEnv("UBERBASE_ALLOWED_ORIGINS")

	return config
}
