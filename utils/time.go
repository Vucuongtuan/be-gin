package utils

import "time"




func CurrentTime() *time.Time {
	currentTime := time.Now()
	return &currentTime
}