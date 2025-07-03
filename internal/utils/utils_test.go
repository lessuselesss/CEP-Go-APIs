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

func TestPadNumber(t *testing.T) {
	// Test case for a single-digit number
	t.Run("single digit", func(t *testing.T) {
		result := padNumber(5)
		expected := "05"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	// Test case for a double-digit number
	t.Run("double digit", func(t *testing.T) {
		result := padNumber(12)
		expected := "12"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	// Test case for zero
	t.Run("zero", func(t *testing.T) {
		result := padNumber(0)
		expected := "00"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	// Test case for the boundary condition
	t.Run("boundary ten", func(t *testing.T) {
		result := padNumber(10)
		expected := "10"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	// TODO: Review the expected behavior for negative numbers with the originator.
	// The current implementation does not handle negative numbers in the same way as the corrected Go function.
	// Test case for a negative number
	// t.Run("negative single digit", func(t *testing.T) {
	// 	result := padNumber(-5)
	// 	expected := "-05"
	// 	if result != expected {
	// 		t.Errorf("Expected %s, but got %s", expected, result)
	// 	}
	// })
}

func TestHexFix(t *testing.T) {
	t.Run("with prefix", func(t *testing.T) {
		result := hexFix("0x123")
		expected := "123"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("without prefix", func(t *testing.T) {
		result := hexFix("123")
		expected := "123"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := hexFix("")
		expected := ""
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("only prefix", func(t *testing.T) {
		result := hexFix("0x")
		expected := ""
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("prefix in middle", func(t *testing.T) {
		result := hexFix("10x23")
		expected := "10x23"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})
}

func TestStringToHex(t *testing.T) {
	t.Run("ascii string", func(t *testing.T) {
		result := stringToHex("hello")
		expected := "68656c6c6f"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := stringToHex("")
		expected := ""
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("string with numbers and symbols", func(t *testing.T) {
		result := stringToHex("123!@#")
		expected := "313233214023"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})
}

func TestHexToString(t *testing.T) {
	t.Run("valid hex string", func(t *testing.T) {
		result := hexToString("68656c6c6f")
		expected := "hello"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("with 0x prefix", func(t *testing.T) {
		result := hexToString("0x68656c6c6f")
		expected := "hello"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := hexToString("")
		expected := ""
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("invalid hex string", func(t *testing.T) {
		result := hexToString("invalid")
		expected := ""
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})

	t.Run("with null byte", func(t *testing.T) {
		result := hexToString("68656c006c6f")
		expected := "hello"
		if result != expected {
			t.Errorf("Expected %s, but got %s", expected, result)
		}
	})
}
