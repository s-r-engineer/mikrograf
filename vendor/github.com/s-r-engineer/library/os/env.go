package libraryOS

import "os"

func GetEnvOrDefault(envVariableName, defaultValue string) string {
	var found bool
	var result string
	if result, found = os.LookupEnv(envVariableName); !found {
		return defaultValue
	}
	return result
}
