package headers

import (
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

const validChars = "abcdefghijklmnopqrstuvwxyz0123456789!#$%&'*+-.^_`|~"

func containsOnlyValidCharacters(s string) bool {
	set := make(map[rune]bool)
	for _, r := range validChars {
		set[r] = true
	}

	for _, char := range s {
		if !set[char] {
			return false
		}
	}
	return true
}

func (h Headers) Get(key string) string {
	// Convert key to lowercase to ensure case-insensitive access
	return h[strings.ToLower(key)]
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	dataString := string(data)

	// fmt.Println(dataString)
	// fmt.Println("Data length:", len(dataString))

	// Check if the data buffer contains CRLF - if not, re-parse later with more data
	if !strings.Contains(dataString, "\r\n") {
		return 0, false, nil
	}

	lines := strings.Split(dataString, "\r\n")

	// Check if the first line is empty, if it is we assume the headers are done
	if len(lines[0]) == 0 {
		// fmt.Println("lines empty")
		return 0, true, nil
	}

	n = 0

	// only parse the first line
	line := lines[0]

	split := strings.Split(line, ": ")

	if len(split) != 2 {
		return 0, false, errors.New("Invalid header, expected <key>: <value> format: " + line)
	}

	if strings.HasSuffix(split[0], " ") {
		return 0, false, errors.New("Invalid header, space after key: " + line)
	}

	key := strings.ToLower(strings.TrimSpace(split[0]))

	if !containsOnlyValidCharacters(key) {
		return 0, false, errors.New("Invalid header key, only alphanumeric and certain special characters are allowed: " + key)
	}

	value := strings.TrimSpace(split[1])

	if len(h[key]) > 0 {
		// If the header already exists, append the new value
		h[key] += ", " + value
	} else {
		// Otherwise, set the header
		h[key] = value
	}
	n = len(line) + 2
	return n, false, nil
}
