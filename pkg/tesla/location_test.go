package tesla

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetURLCode(t *testing.T) {
	tests := []struct {
		u     string
		code  string
		found bool
	}{
		{
			u:     "https://auth.tesla.com/void/callback?code=b517e3553eee3478abb1453f867534a5964871a86bc2814b545f7d022ad8&state=eEJsdVQ3VmlXeGo0SlhoSw&issuer=https%3A%2F%2Fauth.tesla.com%2Foauth2%2Fv3",
			code:  "b517e3553eee3478abb1453f867534a5964871a86bc2814b545f7d022ad8",
			found: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.u, func(t *testing.T) {
			u, err := url.Parse(tt.u)
			assert.NoError(t, err)

			code, found := getURLCode(u)
			assert.Equal(t, tt.code, code)
			assert.Equal(t, tt.found, found)
		})
	}
}
