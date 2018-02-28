package avurnav

import (
	"net/url"
)

// PremarInterface describes which data a Préfet
// Maritime service should be able to give
type PremarInterface interface {
	// Client gets the HTTP client
	Client() *Client
	// BaseURL returns the base URL of the website
	BaseURL() *url.URL
	// Region returns the region under the authority of
	// the Préfet Maritime
	Region() string
}

// PremarService gives information
// about a specific Préfet Maritime
type PremarService struct {
	client  *Client
	baseURL string
	region  string
}

// Client gets the HTTP client
func (s PremarService) Client() *Client {
	return s.client
}

// BaseURL returns the base URL of the website
func (s PremarService) BaseURL() *url.URL {
	url, err := url.Parse(s.baseURL)
	if err != nil {
		panic(err)
	}
	return url
}

// Region returns the region under the authority of
// the Préfet Maritime
func (s PremarService) Region() string {
	return s.region
}
