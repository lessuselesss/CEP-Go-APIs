package utils

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// PadNumber adds a leading zero to numbers less than 10.
func PadNumber(num int) string {
	if num < 10 {
		return fmt.Sprintf("0%d", num)
	}
	return fmt.Sprintf("%d", num)
}

// GetFormattedTimestamp generates a UTC timestamp in YYYY:MM:DD-HH:MM:SS format.
func GetFormattedTimestamp() string {
	now := time.Now().UTC()
	year := now.Year()
	month := PadNumber(int(now.Month()))
	day := PadNumber(now.Day())
	hours := PadNumber(now.Hour())
	minutes := PadNumber(now.Minute())
	seconds := PadNumber(now.Second())
	return fmt.Sprintf("%d:%s:%s-%s:%s:%s", year, month, day, hours, minutes, seconds)
}

// HexFix removes '0x' prefix from hexadecimal strings if present.
func HexFix(word string) string {
	if strings.HasPrefix(word, "0x") {
		return word[2:]
	}
	return word
}

// StringToHex converts a string to its hexadecimal representation.
func StringToHex(str string) string {
	return hex.EncodeToString([]byte(str))
}

// HexToString converts a hexadecimal string back to its original string form.
func HexToString(hexStr string) string {
	cleanedHex := HexFix(hexStr)
	bytes, err := hex.DecodeString(cleanedHex)
	if err != nil {
		return ""
	}
	// Strip null bytes to match the reference implementation
	return strings.ReplaceAll(string(bytes), "\x00", "")
}
