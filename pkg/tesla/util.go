package tesla

import (
	"crypto/sha256"
	"encoding/base64"
	"math/rand"
	"strings"
)

var randomCharacters = []rune(
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789",
)

// Sourced from: https://raw.githubusercontent.com/teslascope/tokens/master/auth.tokens.py
func generateCodeAndState() (codeVerifier, codeChallenge, state string) {
	var b strings.Builder
	for i := 0; i < 86; i++ {
		b.WriteRune(randomCharacters[rand.Intn(len(randomCharacters))])
	}
	codeVerifier = b.String()

	msg := sha256.Sum256([]byte(codeVerifier))
	codeChallenge = base64URLSafe(msg[:])

	var s strings.Builder
	for i := 0; i < 16; i++ {
		s.WriteRune(randomCharacters[rand.Intn(len(randomCharacters))])
	}
	state = base64URLSafe([]byte(s.String()))

	return
}

func base64URLSafe(s []byte) string {
	b := base64.StdEncoding.EncodeToString(s)
	return strings.TrimRightFunc(b, func(r rune) bool {
		return r == '='
	})
}
