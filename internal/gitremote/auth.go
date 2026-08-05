package gitremote

import (
	"encoding/base64"
	"strings"
)

// authHeaderGit — значение для git -c http.extraHeader=... (HTTPS + GitHub PAT).
// Формат: Authorization: Basic base64("oauth2:" + pat)
func authHeaderGit(pat string) string {
	pat = strings.TrimSpace(pat)
	b := base64.StdEncoding.EncodeToString([]byte("oauth2:" + pat))
	return "Authorization: Basic " + b
}
