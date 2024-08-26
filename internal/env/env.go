package env

import "os"

// GetEnvOrDefault returns the environment variable value with the key or the default value if the key is not exists.
func GetEnvOrDefault(key, defaultValue string) string {
	got, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return got
}
