package apiclient

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Response is a simple wrapper to add functionality into http.Response
// Responses MUST be closed with Response.Close or Response.Body.Close
type Response struct {
	*http.Response
}

// Close will close the underlying response body
func (r *Response) Close() error {
	return r.Body.Close()
}

// Bytes returns the response body as a byte slice
// This is equivalent to just reading the body using io.ReadAll
func (r *Response) Bytes() ([]byte, error) {
	defer r.Close()

	return io.ReadAll(r.Body)
}

// String returns the response body as a string
// This is equivalent to just reading the body using io.ReadAll, and casting to string
func (r *Response) String() (string, error) {
	defer r.Close()

	b, err := r.Bytes()

	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Unmarshal decodes the response into the specified value
func (r *Response) Unmarshal(v any) error {
	contentType := r.Header.Get(HeaderContentType)

	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[0:idx]
	}

	switch contentType {
	case MIMEApplicationJSON:
		defer r.Close()

		return json.NewDecoder(r.Body).Decode(&v)
	case MIMETextXML, MIMEApplicationXML:
		defer r.Close()

		return xml.NewDecoder(r.Body).Decode(&v)
	}

	return fmt.Errorf("unsupported content type %s", r.Header.Get("Content-Type"))
}
