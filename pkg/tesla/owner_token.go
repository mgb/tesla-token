package tesla

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

func (o *ownerAPI) GetOwnerToken(accessToken string) (string, error) {
	u, err := url.Parse(urlOwnerAPI)
	if err != nil {
		return "", fmt.Errorf("failed to generate owner token url: %w", err)
	}

	headers := map[string]string{
		"authorization": fmt.Sprintf("bearer %s", accessToken),
	}

	tokenJson := struct {
		GrantType string `json:"grant_type"`
		ClientID  string `json:"client_id"`
	}{
		GrantType: grantType,
		ClientID:  teslaClientID,
	}
	jsonBytes, err := json.Marshal(tokenJson)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json for owner token: %w", err)
	}

	body, _, err := o.httpPostJson(u, headers, nil, jsonBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	o.saveToHTML(htmlOwnerToken, body)

	var ownerTokenJson struct {
		AccessToken string `json:"access_token"`
	}

	err = json.Unmarshal([]byte(body), &ownerTokenJson)
	if err != nil {
		return "", fmt.Errorf("failed to json decode generate token: %w", err)
	}

	if ownerTokenJson.AccessToken == "" {
		return "", errors.New("access token value not set from owner token response")
	}

	return ownerTokenJson.AccessToken, nil
}
