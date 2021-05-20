package tesla

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (o *ownerAPI) httpGet(u *url.URL, headers map[string]string, queryParams map[string]string) (body string, location *url.URL, err error) {
	r := http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	}

	o.fillHTTPRequest(&r, headers, queryParams, nil)

	return o.httpDo(&r)
}

func (o *ownerAPI) httpPostForm(u *url.URL, headers map[string]string, queryParams map[string]string, formParams map[string]string) (body string, location *url.URL, err error) {
	r := http.Request{
		Method: "POST",
		URL:    u,
		Header: make(http.Header),
	}

	r.Header.Add("content-type", "application/x-www-form-urlencoded")

	var postBody io.Reader
	if formParams != nil {
		values := url.Values{}
		for k, v := range formParams {
			values.Set(k, v)
		}
		postBody = strings.NewReader(values.Encode())
	}

	o.fillHTTPRequest(&r, headers, queryParams, postBody)

	return o.httpDo(&r)
}

func (o *ownerAPI) httpPostJson(u *url.URL, headers map[string]string, queryParams map[string]string, jsonBytes []byte) (body string, location *url.URL, err error) {
	r := http.Request{
		Method: "POST",
		URL:    u,
		Header: make(http.Header),
	}

	r.Header.Add("content-type", "application/json")

	o.fillHTTPRequest(&r, headers, queryParams, bytes.NewReader(jsonBytes))

	return o.httpDo(&r)
}

func (o *ownerAPI) fillHTTPRequest(r *http.Request, headers map[string]string, queryParams map[string]string, body io.Reader) {
	r.Header.Add("user-agent", o.userAgent)
	for k, v := range headers {
		r.Header.Add(k, v)
	}

	if queryParams != nil {
		q := r.URL.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		r.URL.RawQuery = q.Encode()
	}

	if body != nil {
		r.Body = io.NopCloser(body)
	}
}

func (o *ownerAPI) httpDo(r *http.Request) (body string, location *url.URL, err error) {
	resp, err := o.client.Do(r)
	if err != nil {
		return "", nil, fmt.Errorf("failed to %s %s: %w", r.Method, r.URL, err)
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read body from %s %s: %w", r.Method, r.URL, err)
	}

	u, _ := resp.Location()

	return string(b), u, nil
}
