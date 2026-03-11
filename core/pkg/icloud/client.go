package icloud

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	baseURL    = "https://api.apple-cloudkit.com"
	apiVersion = 1
)

// service is the base service used by all CloudKit Web Service APIs.
//
//nolint:structcheck // https://github.com/golangci/golangci-lint/issues/1517
type service struct {
	client   *Client
	basePath string
}

// DefaultHTTPClient returns the default HTTP client used for making requests.
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	}
}

// Client provides the CloudKit Web Services API operations.
type Client struct {
	baseURL        *url.URL
	userAgent      string
	keyID          string
	privateKey     *ecdsa.PrivateKey
	strictDecoding bool

	httpClient *http.Client

	Records *RecordsService
}

// NewClient returns a new CloudKit Web Services API client.
func NewClient(container, keyID string, privateKey *ecdsa.PrivateKey, environment Environment, reqBasePath string, options ...Option) (*Client, error) {
	urlStr := fmt.Sprintf("%s/database/%d/%s/%s", baseURL, apiVersion, container, environment)
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, err
	}

	client := &Client{
		baseURL:    u,
		userAgent:  "icloud-go",
		keyID:      keyID,
		privateKey: privateKey,

		httpClient: DefaultHTTPClient(),
	}

	client.Records = &RecordsService{client, reqBasePath}

	// Apply supplied options.
	if err := client.Options(options...); err != nil {
		return nil, err
	}

	return client, nil
}

// Options applies Options to the Client.
func (c *Client) Options(options ...Option) error {
	for _, option := range options {
		if err := option(c); err != nil {
			return err
		}
	}
	return nil
}

// call creates a new API request and executes it. The response body is JSON
// decoded or directly written to v, depending on v being an io.Writer or not.
func (c *Client) call(ctx context.Context, method, endpoint string, body, v interface{}) error {
	req, err := c.newRequest(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	return c.do(req, v)
}

// newRequest creates an API request. The given body will be included as a JSON
// encoded request body.
func (c *Client) newRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	endpoint = path.Join(c.baseURL.Path, endpoint)
	rel, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, err
	}
	u := c.baseURL.ResolveReference(rel)

	var (
		buf bytes.Buffer
		h   = sha256.New()
	)
	if body != nil {
		w := io.MultiWriter(&buf, h)
		if err = json.NewEncoder(w).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("user-agent", c.userAgent)

	if err = c.signRequest(req, h); err != nil {
		return nil, err
	}

	return req, nil
}

// do sends an API request and returns the API response. The response body is
// JSON decoded or directly written to v, depending on v being an io.Writer or
// not.
func (c *Client) do(req *http.Request, v interface{}) error {
	// bodyBytes, _ := ioutil.ReadAll(req.Body)
	// fmt.Printf("\n req.Body: %s \n", bodyBytes)
	// fmt.Printf("\n req.Header: %+v \n", req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if statusCode := resp.StatusCode; statusCode >= 400 {
		// Handle a generic HTTP error if the response is not JSON formatted.
		if val := resp.Header.Get("content-type"); !strings.HasPrefix(val, "application/json") {
			return Error{
				Reason: http.StatusText(statusCode),
			}
		}

		// For error handling, we want to have access to the raw request body to
		// inspect it further.
		var (
			buf bytes.Buffer
			dec = json.NewDecoder(io.TeeReader(resp.Body, &buf))
		)

		// Handle a properly JSON formatted CloudKit Web Services API error
		// response.
		var errResp Error
		if err = dec.Decode(&errResp); err != nil {
			return fmt.Errorf("error decoding %d error response: %w", statusCode, err)
		}

		// In case something went wrong, include the raw response and hope for
		// the best.
		if errResp.Reason == "" && errResp.Code == Unknown {
			s := strings.ReplaceAll(buf.String(), "\n", " ")
			errResp.Reason = s
		}

		return errResp
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			return err
		}

		dec := json.NewDecoder(resp.Body)
		if c.strictDecoding {
			dec.DisallowUnknownFields()
		}
		return dec.Decode(v)
	}

	return nil
}

// signRequest signs the request with a signature of format date:body:path where
// date is the ISO8601 representation of the current date, body the base64
// string encoded SHA-256 hash of the request body and path the API path without
// base url and query parameters. The SHA-256 hash.Hash must be precomputed
// before calling this function.
func (c *Client) signRequest(req *http.Request, bodyHash hash.Hash) error {
	var (
		buf     bytes.Buffer
		dateStr = time.Now().UTC().Format(time.RFC3339)
	)

	_, _ = buf.WriteString(dateStr)
	_ = buf.WriteByte(':')

	encodedBody := base64.StdEncoding.EncodeToString(bodyHash.Sum(nil))
	_, _ = buf.WriteString(encodedBody)
	_ = buf.WriteByte(':')

	_, _ = buf.WriteString(req.URL.Path)

	// Hash the signature message.
	h := sha256.New()
	if _, err := io.Copy(h, &buf); err != nil {
		return err
	}

	// Generate the signature from the hashed message.
	signature, err := c.privateKey.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		return err
	}

	req.Header.Set("x-apple-cloudkit-request-keyid", c.keyID)
	req.Header.Set("x-apple-cloudkit-request-iso8601date", dateStr)
	req.Header.Set("x-apple-cloudkit-request-signaturev1", base64.StdEncoding.EncodeToString(signature))

	return nil
}

//
// func (c *Client) ReqHeader(_url string) (map[string]interface{}, error) {
// 	var (
// 		buf     bytes.Buffer
// 		dateStr = time.Now().UTC().Format(time.RFC3339)
// 	)
//
// 	_, _ = buf.WriteString(dateStr)
// 	_ = buf.WriteByte(':')
//
// 	encodedBody := base64.StdEncoding.EncodeToString(sha256.New().Sum(nil))
// 	_, _ = buf.WriteString(encodedBody)
// 	_ = buf.WriteByte(':')
//
// 	_, _ = buf.WriteString(_url)
//
// 	// Hash the signature message.
// 	h := sha256.New()
// 	if _, err := io.Copy(h, &buf); err != nil {
// 		return nil, err
// 	}
//
// 	// Generate the signature from the hashed message.
// 	signature, err := c.privateKey.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return map[string]interface{}{
// 		"x-apple-cloudkit-request-keyid":       c.keyID,
// 		"x-apple-cloudkit-request-iso8601date": dateStr,
// 		"x-apple-cloudkit-request-signaturev1": base64.StdEncoding.EncodeToString(signature),
// 		"content-type":                         "application/json",
// 		"accept":                               "application/json",
// 		"user-agent":                           "icloud-go",
// 	}, nil
// }
