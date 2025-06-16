package utils

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewID(size int, prefix string) string {
	l := size - len(prefix)
	if l <= 0 {
		return ""
	}
	id, _ := gonanoid.Generate(alphabet, l)
	return prefix + id
}
