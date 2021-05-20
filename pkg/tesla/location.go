package tesla

import (
	"net/http"
	"net/url"
	"strings"
)

func (o *ownerAPI) checkRedirect(req *http.Request, via []*http.Request) error {
	// TODO(minegbooom): Check base url (may have redirected to auth.tesla.cn)
	if req == nil || req.URL == nil {
		return nil
	}
	if _, ok := getURLCode(req.URL); ok {
		// Don't redirect if we are going to our code endpoint
		return http.ErrUseLastResponse
	}
	return nil
}

func getURLCode(u *url.URL) (string, bool) {
	if u == nil || !strings.HasPrefix(u.Host, urlVoidCallbackBase) {
		return "", false
	}
	if u.Path != urlVoidCallbackPath {
		return "", false
	}

	code := u.Query().Get("code")
	if code == "" {
		return "", false
	}

	return code, true
}
