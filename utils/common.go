package utils

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"strings"
	"unicode"
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func GenerateMD5(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

func Slugify(input string) string {
	var cleaned strings.Builder

	for _, char := range input {
		if unicode.IsLetter(char) {
			cleaned.WriteRune(unicode.ToLower(char))
		}
	}

	return cleaned.String()
}

func LogVerbose() bool {
	return os.Getenv("VERBOSE_LOG") == "yes"
}
