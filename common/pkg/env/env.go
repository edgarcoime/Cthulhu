package env

import (
	"os"

	"github.com/joho/godotenv"
)

var Env map[string]string

func GetEnv(key, def string) string {
	// First check the Env map (loaded from .env file)
	if val, ok := Env[key]; ok {
		return val
	}
	// Fall back to system environment variables
	if val := os.Getenv(key); val != "" {
		return val
	}
	// Finally, return the default value
	return def
}

// SetupEnvFile loads environment variables from a .env file
// It populates the Env map and also loads them into os.Getenv()
func SetupEnvFile(path string) {
	// Load into os.Getenv() (standard behavior)
	err := godotenv.Load(path)
	if err != nil {
		// If file doesn't exist, that's okay - we'll use defaults
		return
	}

	// Also populate the Env map for direct access
	envMap, err := godotenv.Read(path)
	if err != nil {
		// If we can't read it, at least os.Getenv() will have it
		return
	}
	Env = envMap
}
