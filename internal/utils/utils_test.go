package utils

import (
	"strings"
	"testing"
	"time"
)

func TestGetFormattedTimestamp(t *testing.T) {
	// Call the function to get the formatted timestamp
	timestamp := GetFormattedTimestamp()

	// Split the timestamp into date and time parts
	parts := strings.Split(timestamp, "-")
	if len(parts) != 2 {
		t.Errorf("Timestamp does not have the expected format (YYYY:MM:DD-HH:MM:SS). Received: %s", timestamp)
		return
	}

	// Validate the date part
	dateParts := strings.Split(parts[0], ":")
	if len(dateParts) != 3 {
		t.Errorf("Date part does not have the expected format (YYYY:MM:DD). Received: %s", parts[0])
		return
	}

	// Validate the time part
	timeParts := strings.Split(parts[1], ":")
	if len(timeParts) != 3 {
		t.Errorf("Time part does not have the expected format (HH:MM:SS). Received: %s", parts[1])
		return
	}

	// Ensure the timestamp represents a valid time
	parsedTime, err := time.Parse("2006:01:02-15:04:05", timestamp)
	if err != nil {
		t.Errorf("Timestamp is not valid. Error: %v", err)
		return
	}

	// Ensure the parsed time is close to the current time (within a few seconds)
	now := time.Now().UTC()
	diff := now.Sub(parsedTime)
	if diff.Seconds() > 5 || diff.Seconds() < -5 {
		t.Errorf("Timestamp is not close to the current UTC time. Difference: %v seconds", diff.Seconds())
	}
}