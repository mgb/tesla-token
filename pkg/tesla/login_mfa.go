package tesla

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (o *ownerAPI) runMFALogin(authcodeFn func() string, formData map[string]string) error {
	transactionID := formData["transaction_id"]
	factorID, err := o.getMFAFactorID(transactionID)
	if err != nil {
		return err
	}

	authcode := authcodeFn()
	if authcode == "" {
		return errors.New("authcode is required when MFA is required")
	}

	err = o.verifyMFACode(transactionID, factorID, authcode)
	if err != nil {
		return err
	}

	return nil
}

func (o *ownerAPI) getMFAFactorID(transactionID string) (string, error) {
	u, err := url.Parse(fmt.Sprintf("%s/authorize/mfa/factors", o.baseURL))
	if err != nil {
		return "", fmt.Errorf("failed to generate mfa factors url: %w", err)
	}

	queryParams := map[string]string{
		"transaction_id": transactionID,
	}

	body, _, err := o.httpGet(u, nil, queryParams)
	if err != nil {
		return "", fmt.Errorf("failed to get mfa factors: %w", err)
	}
	o.saveToHTML(htmlMFAFactorID, body)

	var mfaFactorJson struct {
		Data []struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			FactorType     string `json:"factor_type"`
			FactorProvider string `json:"factor_provider"`
		}
	}

	err = json.Unmarshal([]byte(body), &mfaFactorJson)
	if err != nil {
		return "", fmt.Errorf("failed to json decode mfa factors: %w", err)
	}

	if len(mfaFactorJson.Data) == 0 {
		return "", errors.New("no mfa factors found")
	}

	unsupportedTypes := map[string]struct{}{}
	for _, d := range mfaFactorJson.Data {
		switch d.FactorType {
		case factorTypeSoftwareToken:
			return d.ID, nil

			// TODO(minegoboom): Looks like Tesla is prepared for more than one kind of MFA.
			// When they add more support, we can add more support here. Until then, use the
			// first one we support.

		default:
			unsupportedTypes[d.FactorType] = struct{}{}
		}
	}

	factorTypes := make([]string, 0, len(unsupportedTypes))
	for t := range unsupportedTypes {
		factorTypes = append(factorTypes, t)
	}

	return "", fmt.Errorf(
		"could not find mfa we support, found: %s",
		strings.Join(factorTypes, ", "),
	)
}

func (o *ownerAPI) verifyMFACode(transactionID, factorID, authcode string) error {
	u, err := url.Parse(fmt.Sprintf("%s/authorize/mfa/verify", o.baseURL))
	if err != nil {
		return fmt.Errorf("failed to generate mfa verify url: %w", err)
	}

	mfaPostJson := struct {
		TransactionId string `json:"transaction_id"`
		FactorID      string `json:"factor_id"`
		Passcode      string `json:"passcode"`
	}{
		TransactionId: transactionID,
		FactorID:      factorID,
		Passcode:      authcode,
	}
	jsonBytes, err := json.Marshal(mfaPostJson)
	if err != nil {
		return fmt.Errorf("failed to marshal json for mfa verify: %w", err)
	}

	body, _, err := o.httpPostJson(u, nil, nil, jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to verify mfa: %w", err)
	}
	o.saveToHTML(htmlMFAVerify, body)

	var mfaCodeJson struct {
		Data struct {
			Id          string
			ChallengeId string
			FactorId    string
			PassCode    string
			Approved    bool
			Flagged     bool
			Valid       bool
		}
	}

	err = json.Unmarshal([]byte(body), &mfaCodeJson)
	if err != nil {
		return fmt.Errorf("failed to json decode mfa factors: %w", err)
	}

	if !mfaCodeJson.Data.Approved {
		return errors.New("mfa verification failed")
	}

	return nil
}
