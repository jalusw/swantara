package helper

import (
	"net/http"
	"strings"
)

func IsImage(content []byte) bool {
	return strings.HasPrefix(http.DetectContentType(content), "image/")
}
