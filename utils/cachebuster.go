package main

import (
	"fmt"
	"strconv"
	"time"

	"dagger/utils/internal/dagger"
)

// Define the cache buster strategy
func (m *Utils) WithCacheBurster(
	ctr *dagger.Container,

	// Define if the cache burster level is done per day ('daily'), per hour ('hour'), per minute ('minute'), per second ('default') or no cache buster ('none')
	// +optional
	cacheBursterLevel string,
) *dagger.Container {
	if cacheBursterLevel == "none" {
		return ctr
	}

	utcNow := time.Now().UTC()
	cacheBursterKey := fmt.Sprintf("%d%d%d", utcNow.Year(), utcNow.Month(), utcNow.Day())

	switch cacheBursterLevel {
	case "daily":
	case "hour":
		cacheBursterKey += strconv.Itoa(utcNow.Hour())
	case "minute":
		cacheBursterKey += fmt.Sprintf("%d%d", utcNow.Hour(), utcNow.Minute())
	default:
		cacheBursterKey += fmt.Sprintf("%d%d%d", utcNow.Hour(), utcNow.Minute(), utcNow.Second())
	}

	return ctr.WithEnvVariable("CACHE_BURSTER", cacheBursterKey)
}
