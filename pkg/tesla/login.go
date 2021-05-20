package tesla

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func (o *ownerAPI) Login(usernameFn, passwordFn, authcodeFn func() string) (accessToken, refreshToken string, err error) {
	username := usernameFn()
	if username == "" {
		return "", "", errors.New("username is required")
	}

	formData, err := o.getLoginForm()
	if err != nil {
		return "", "", err
	}

	password := passwordFn()
	if password == "" {
		return "", "", errors.New("password is required")
	}

	code, mfaRequired, err := o.postAuthorize(username, password, formData)
	if err != nil {
		return "", "", err
	}

	if mfaRequired {
		err := o.runMFALogin(authcodeFn, formData)
		if err != nil {
			return "", "", err
		}

		code, mfaRequired, err = o.postAuthorize(username, password, formData)
		if err != nil {
			return "", "", err
		}
		if mfaRequired {
			return "", "", errors.New("mfa was required after mfa verified; mfa expired?")
		}
	}

	if code == "" {
		return "", "", errors.New("failed to get authorization code")
	}

	accessToken, refreshToken, err = o.postToken(code)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (o *ownerAPI) getLoginForm() (map[string]string, error) {
	u, err := url.Parse(fmt.Sprintf("%s/authorize", o.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to generate login form url: %w", err)
	}

	body, _, err := o.httpGet(u, nil, o.getQueryParam())
	if err != nil {
		return nil, fmt.Errorf("failed to get Login Form: %w", err)
	}
	o.saveToHTML(htmlLoginForm, body)

	// Use the DOM to get all our hidden input values
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed parsing dom from login form: %w", err)
	}

	hiddenValues := map[string]string{}
	dom.Find("input").Each(func(i int, selection *goquery.Selection) {
		t, _ := selection.Attr("type")
		if t != "hidden" {
			return
		}

		name, _ := selection.Attr("name")
		value, _ := selection.Attr("value")
		hiddenValues[name] = value
	})

	for _, f := range requiredHiddenFields {
		if hiddenValues[f] == "" {
			return nil, fmt.Errorf("missing required hidden input field: %s", f)
		}
	}

	return hiddenValues, nil
}

func (o *ownerAPI) postAuthorize(username, password string, formData map[string]string) (code string, mfaRequired bool, err error) {
	u, err := url.Parse(fmt.Sprintf("%s/authorize", o.baseURL))
	if err != nil {
		return "", false, fmt.Errorf("failed to generate authorize form url: %w", err)
	}

	formParams := map[string]string{
		"identity":   username,
		"credential": password,
	}
	for k, v := range formData {
		formParams[k] = v
	}

	body, location, err := o.httpPostForm(u, nil, o.getQueryParam(), formParams)
	if err != nil {
		return "", false, fmt.Errorf("failed to get Login Form: %w", err)
	}

	code, ok := getURLCode(location)
	if ok {
		return code, false, nil
	}

	o.saveToHTML(htmlAuthorize, body)

	// Check if MFA is required
	if strings.Contains(body, "mfa/verify") {
		return "", true, nil
	}

	return "", false, errors.New("authorize login failed")
}

func (o *ownerAPI) postToken(code string) (string, string, error) {
	u, err := url.Parse(fmt.Sprintf("%s/token", o.baseURL))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token url: %w", err)
	}

	tokenJson := struct {
		ClientID     string `json:"client_id"`
		Code         string `json:"code"`
		CodeVerifier string `json:"code_verifier"`
		GrantType    string `json:"grant_type"`
		RedirectURI  string `json:"redirect_uri"`
	}{
		ClientID:     "ownerapi",
		Code:         code,
		CodeVerifier: o.codeVerifier,
		GrantType:    "authorization_code",
		// TODO(mgb): Support auth.tesla.cn
		RedirectURI: urlVoidCallback,
	}
	jsonBytes, err := json.Marshal(tokenJson)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal json for token: %w", err)
	}

	body, _, err := o.httpPostJson(u, nil, nil, jsonBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}
	o.saveToHTML(htmlMFAToken, body)

	var tokenAuthJson struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
	}

	err = json.Unmarshal([]byte(body), &tokenAuthJson)
	if err != nil {
		return "", "", fmt.Errorf("failed to json decode generate token: %w", err)
	}

	return tokenAuthJson.AccessToken, tokenAuthJson.RefreshToken, nil
}

func (o *ownerAPI) getQueryParam() map[string]string {
	return map[string]string{
		"audience":              "",
		"client_id":             "ownerapi",
		"code_challenge_method": "S256",
		"code_challenge":        o.codeChallenge,
		"locale":                "en",
		"prompt":                "login",
		"redirect_uri":          urlVoidCallback,
		"response_type":         "code",
		"scope":                 "openid email offline_access",
		"state":                 o.state,
	}
}

func (o *ownerAPI) saveToHTML(filename, body string) error {
	if !o.saveHTML {
		return nil
	}

	filename = filepath.Join(o.saveDir,
		fmt.Sprintf("%s-%s-%x.html",
			filename,
			time.Now().Format("2006-01-02-15-04-05"),
			sha256.Sum256([]byte(body)),
		),
	)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed writing file %q: %w", filename, err)
	}
	defer f.Close()

	_, err = io.Copy(f, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed copying data to file %q: %w", filename, err)
	}

	return nil
}
