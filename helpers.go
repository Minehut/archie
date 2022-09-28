package main

import (
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
)

func LookupEnvOrUint(key string, defaultVal uint) uint {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			log.Fatal().Msgf("LookupEnvOrUint[%s]: %v", key, err)
		}
		return uint(v)
	}
	return defaultVal
}

func LookupEnvOrUint64(key string, defaultVal uint64) uint64 {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			log.Fatal().Msgf("LookupEnvOrUint64[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal().Msgf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func LookupEnvOrInt64(key string, defaultVal int64) int64 {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Fatal().Msgf("LookupEnvOrInt64[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatal().Msgf("LookupEnvOrBool[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}
