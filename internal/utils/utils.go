package utils

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// PadNumber adds a leading zero to an integer if it is less than 10.
// This is used for consistent formatting of date and time components.
func PadNumber(num int) string {
	if num < 10 {
		return fmt.Sprintf("0%d", num)
	}
	return fmt.Sprintf("%d", num)
}

// GetFormattedTimestamp generates a UTC timestamp in the "YYYY:MM:DD-HH:MM:SS"
// format required by the Circular Protocol.
func GetFormattedTimestamp() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d:%s:%s-%s:%s:%s",
		now.Year(),
		PadNumber(int(now.Month())),
		PadNumber(now.Day()),
		PadNumber(now.Hour()),
		PadNumber(now.Minute()),
		PadNumber(now.Second()),
	)
}

// HexFix removes the "0x" prefix from a hexadecimal string, if present.
// It ensures that hex strings are in a consistent format for processing.
func HexFix(word string) string {
	if strings.HasPrefix(strings.ToLower(word), "0x") {
		return word[2:]
	}
	return word
}

// StringToHex converts a standard UTF-8 string into its hexadecimal representation.
func StringToHex(str string) string {
	return hex.EncodeToString([]byte(str))
}

// HexToString converts a hexadecimal string back into its original UTF-8 string form.
// It returns an empty string if the input is not valid hex. It also strips null
// bytes to match the behavior of the reference implementations.
func HexToString(hexStr string) (string, error) {
	cleanedHex := HexFix(hexStr)
	bytes, err := hex.DecodeString(cleanedHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex string: %w", err)
	}
	// Strip null bytes to match the reference implementation's behavior.
	return strings.ReplaceAll(string(bytes), "\x00", ""), nil
}
