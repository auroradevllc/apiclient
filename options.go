package apiclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/auroradevllc/apiclient/multipart"
)

// WithClient sets the http.Client to use for the request
func WithClient(c *http.Client) Option {
	return func(r *Request) error {
		r.client = c
		return nil
	}
}

// WithMethod sets the request method
func WithMethod(method string) Option {
	return func(r *Request) error {
		r.method = method
		return nil
	}
}

// WithBody sets a request's body
// For common types like JSON/XML, use WithJSON and WithXML
//
// This supports the following types:
// - io.Reader
// - multipart.Streamer
// - []byte
// - string
// - url.Values
func WithBody(v any) Option {
	return func(r *Request) error {
		// Set method to POST automatically for GET requests
		if r.method == http.MethodGet {
			r.method = http.MethodPost
		}

		switch t := v.(type) {
		case *multipart.Streamer: // support for multipart streaming
			setHeader(r.headers, HeaderContentType, t.ContentType())
			setHeader(r.headers, HeaderContentLength, t.Len())

			r.body = t
			return nil
		case io.Reader: // support for basic readers
			r.body = t
			return nil
		case []byte: // support for basic byte slices
			setHeader(r.headers, HeaderContentLength, len(t))
			r.body = bytes.NewReader(t)
			return nil
		case string:
			setHeader(r.headers, HeaderContentLength, len(t))
			r.body = strings.NewReader(t)
			return nil
		case url.Values:
			encoded := t.Encode()

			setHeader(r.headers, HeaderContentType, MIMEApplicationForm)
			setHeader(r.headers, HeaderContentLength, len(encoded))

			r.body = strings.NewReader(encoded)
			return nil
		}

		return ErrUnsupportedBody
	}
}

// WithJSON is a wrapper to set a JSON request body
// This automatically sets Content-Length, if available
func WithJSON(v any) Option {
	return func(r *Request) error {
		// Set method to POST automatically for GET requests
		if r.method == http.MethodGet {
			r.method = http.MethodPost
		}

		// Set Content-Type header
		r.headers.Set(HeaderContentType, MIMEApplicationJSON)

		switch t := v.(type) {
		case []byte:
			setHeader(r.headers, HeaderContentLength, len(t))

			r.body = bytes.NewReader(t)
		case io.Reader:
			r.body = t
		default:
			b, err := json.Marshal(v)

			if err != nil {
				return err
			}

			// This wrapper allows us to set a header value without converting to string ourselves
			setHeader(r.headers, HeaderContentLength, len(b))

			r.body = bytes.NewReader(b)
		}

		return nil
	}
}

// WithXML is a wrapper to set an XML request body
// This automatically sets Content-Length, if available
func WithXML(v any) Option {
	return func(r *Request) error {
		// Set method to POST automatically for GET requests
		if r.method == http.MethodGet {
			r.method = http.MethodPost
		}

		// Set Content-Type header
		r.headers.Set(HeaderContentType, MIMETextXML)

		switch t := v.(type) {
		case []byte:
			setHeader(r.headers, HeaderContentLength, len(t))

			r.body = bytes.NewReader(t)
		case io.Reader:
			r.body = t
		default:
			b, err := xml.Marshal(v)

			if err != nil {
				return err
			}

			// This wrapper allows us to set a header value without converting to string ourselves
			setHeader(r.headers, HeaderContentLength, len(b))

			r.body = bytes.NewReader(b)
		}

		return nil
	}
}

// WithHeader sets a request header
func WithHeader(key string, value any) Option {
	return func(r *Request) error {
		setHeader(r.headers, key, value)
		return nil
	}
}

// WithBasicAuth sets basic authentication on the request
func WithBasicAuth(username, password string) Option {
	return func(r *Request) error {
		authStr := fmt.Sprintf("%s:%s", username, password)
		r.headers.Set(HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte(authStr)))
		return nil
	}
}

// setHeader is an internal helper which can format header values
func setHeader(headers http.Header, key string, value any) {
	var val string

	switch t := value.(type) {
	case int8:
		val = strconv.FormatInt(int64(t), 10)
	case uint8:
		val = strconv.FormatInt(int64(t), 10)
	case int16:
		val = strconv.FormatInt(int64(t), 10)
	case uint16:
		val = strconv.FormatInt(int64(t), 10)
	case int32:
		val = strconv.FormatInt(int64(t), 10)
	case uint32:
		val = strconv.FormatInt(int64(t), 10)
	case int:
		val = strconv.FormatInt(int64(t), 10)
	case uint:
		val = strconv.FormatInt(int64(t), 10)
	case int64:
		val = strconv.FormatInt(t, 10)
	case uint64:
		val = strconv.FormatInt(int64(t), 10)
	case bool:
		val = strconv.FormatBool(t)
	case float32:
		val = strconv.FormatFloat(float64(t), 'f', 0, 64)
	case float64:
		val = strconv.FormatFloat(t, 'f', 0, 64)
	case []byte:
		val = string(t)
	case string:
		val = t
	default:
		val = fmt.Sprintf("%v", t)
	}

	headers.Set(key, val)
}
