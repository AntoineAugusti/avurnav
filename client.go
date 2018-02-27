package avurnav

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.1"
	userAgent      = "Go AVURNAV v" + libraryVersion
	mediaType      = "application/json"
)

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string
}

// Error gives information about the error
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

// Client manages communication the API
type Client struct {
	// HTTP client used to communicate with the API
	client *http.Client
	// User agent for client
	UserAgent string
	// Services used for communications with the API
	Manche       AVURNAVFetcher
	Atlantique   AVURNAVFetcher
	Mediterranee AVURNAVFetcher

	Fetchers []AVURNAVFetcher
}

// NewClient returns a new API client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{
		client:    httpClient,
		UserAgent: userAgent,
	}

	c.Manche = AVURNAVFetcher{
		service: PremarService{
			client:  c,
			baseURL: "https://www.premar-manche.gouv.fr",
			region:  "Manche",
		}}
	c.Atlantique = AVURNAVFetcher{
		service: PremarService{
			client:  c,
			baseURL: "https://www.premar-atlantique.gouv.fr",
			region:  "Atlantique",
		}}
	c.Mediterranee = AVURNAVFetcher{
		service: PremarService{
			client:  c,
			baseURL: "https://www.premar-mediterranee.gouv.fr",
			region:  "Méditerranée",
		}}

	c.Fetchers = []AVURNAVFetcher{c.Manche, c.Atlantique, c.Mediterranee}

	return c
}

// NewRequest creates an API request to a given URL.
// If specified, the value pointed to by body is JSON encoded and included in as the request body.
func (c *Client) NewRequest(method string, theURL *url.URL, body interface{}) (*http.Request, error) {
	buf := new(bytes.Buffer)
	if body != nil {
		if w, ok := body.(url.Values); ok {
			buf = bytes.NewBufferString(w.Encode())
		} else {
			if err := json.NewEncoder(buf).Encode(body); err != nil {
				return nil, err
			}
		}
	}

	req, err := http.NewRequest(method, theURL.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", userAgent)
	return req, nil
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	response, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if rerr := response.Body.Close(); err == nil {
			err = rerr
		}
	}()

	err = CheckResponse(response)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err := io.Copy(w, response.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err := json.NewDecoder(response.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, err
}

// CheckResponse checks the API response for errors, and returns them if present. A response is considered an
// error if it has a status code outside the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			return err
		}
	}

	return errorResponse
}
