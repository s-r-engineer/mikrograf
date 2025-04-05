package main

import (
	"strings"

	libraryOS "github.com/s-r-engineer/library/os"
)

const mikrotikHostENVName = "MIKROGRAF_TARGET_HOSTS"
const defaultDelimeter = ";"

func parseTheEnv() []string {
	return strings.Split(libraryOS.GetEnvOrDefault(mikrotikHostENVName, ""), defaultDelimeter)
}
