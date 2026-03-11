package icloud

import "net/http"

// An Option modifies the behaviour of the API client. If not otherwise
// specified by a specific option, they are safe to use even after API methods
// have been called. However, they are not safe to use while the client is
// performing an operation.
type Option func(c *Client) error

// SetHTTPClient specifies a custom http client that should be used to make
// requests.
func SetHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return nil
		}
		c.httpClient = client
		return nil
	}
}

// SetUserAgent sets the user agent used by the http client.
func SetUserAgent(userAgent string) Option {
	return func(c *Client) error {
		c.userAgent = userAgent
		return nil
	}
}
