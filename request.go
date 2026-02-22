package apiclient

import (
	"io"
	"net/http"
	"time"
)

var client = &http.Client{
	Timeout: 5 * time.Minute,
}

type Request struct {
	client  *http.Client
	method  string
	url     string
	headers http.Header
	opts    []Option
	body    io.Reader
}

type Option func(r *Request) error

func NewRequest(url string, opts ...Option) (*Request, error) {
	return DefaultClient.NewRequest(url, opts...)
}

func (c *Client) NewRequest(url string, opts ...Option) (*Request, error) {
	r := &Request{
		client:  client,
		method:  http.MethodGet,
		url:     url,
		headers: make(http.Header),
	}

	if c.defaultOpts != nil {
		opts = append(c.defaultOpts, opts...)
	}

	var err error

	for _, opt := range opts {
		err = opt(r)

		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Request) Send() (*Response, error) {
	req, err := http.NewRequest(r.method, r.url, r.body)

	if err != nil {
		return nil, err
	}

	for key, header := range r.headers {
		req.Header[key] = header
	}

	res, err := r.client.Do(req)

	if err != nil {
		return nil, err
	}

	return &Response{
		Response: res,
	}, nil
}
