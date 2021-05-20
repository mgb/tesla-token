package tesla

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path/filepath"

	"golang.org/x/net/publicsuffix"
)

// OwnerAPI is a listing of all the compatible OwnerAPI we have access to
type OwnerAPI interface {
	// Login will, given a username, password, and optional MFA authcode, return the accessToken and refreshToken
	Login(usernameFn, passwordFn, authcodeFn func() string) (accessToken, refreshToken string, err error)
	GetOwnerToken(accessToken string) (string, error)
}

// Params allows you to configure how the OwnerAPI actions in the background
type Params struct {
	// SaveHTML will, if set, output the raw outputs of the HTTP calls made. Useful for debugging and upgrading unit tests.
	SaveHTML bool
	SaveDir  string
}

// New creates a new OwnerAPI session
func New(p Params) (OwnerAPI, error) {
	codeVerifier, codeChallenge, state := generateCodeAndState()
	o := ownerAPI{
		userAgent: "tesla-token",

		codeVerifier:  codeVerifier,
		codeChallenge: codeChallenge,
		state:         state,

		saveHTML: p.SaveHTML,
	}

	u, err := url.Parse("https://auth.tesla.com/oauth2/v3")
	if err != nil {
		return nil, fmt.Errorf("failed to convert OAuth2 URL to url.URL: %w", err)
	}
	o.baseURL = u

	saveDir, err := filepath.Abs(p.SaveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to convert SaveDir %q to absolute path: %w", saveDir, err)
	}
	o.saveDir = saveDir

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		Jar:           jar,
		CheckRedirect: o.checkRedirect,
	}
	o.client = &client

	return &o, nil
}

type ownerAPI struct {
	baseURL   *url.URL
	userAgent string

	codeVerifier  string
	codeChallenge string
	state         string

	saveHTML bool
	saveDir  string

	client *http.Client
}
