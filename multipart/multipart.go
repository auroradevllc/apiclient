package multipart

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strconv"
	"strings"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// part is a section of a multipart request
// header will never contain any actual file/field data, it is all in reader.
type part struct {
	header []byte
	reader io.Reader
	length int64
}

// Streamer is an implementation of a multipart.Writer which allows you to read directly from it
type Streamer struct {
	b     *bytes.Buffer
	mw    *multipart.Writer
	parts []*part

	reader io.Reader
}

// New creates a new Multipart streamer
func New() *Streamer {
	var b bytes.Buffer

	return &Streamer{
		b:  &b,
		mw: multipart.NewWriter(&b),
	}
}

// CreatePart functions the same as multipart's CreatePart, but takes a reader instead of returning a writer.
func (s *Streamer) CreatePart(h textproto.MIMEHeader, r io.Reader, length int64) error {
	_, err := s.mw.CreatePart(h)

	if err != nil {
		return err
	}

	s.parts = append(s.parts, &part{
		header: s.readHeader(),
		reader: r,
		length: length,
	})

	return nil
}

// readHeader reads our buffer into a new byte slice, then resets the buffer position for re-use.
func (s *Streamer) readHeader() []byte {
	newBuf := make([]byte, s.b.Len())
	copy(newBuf, s.b.Bytes())

	// Reset our buffer, so that we only have the necessary data for the next request
	s.b.Reset()

	return newBuf
}

// CreateFormFile is a convenience wrapper around CreatePart. It creates
// a new form-data header with the provided field name and file name.
func (s *Streamer) CreateFormFile(fieldname, filename string, r io.Reader, length int64) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", "application/octet-stream")
	return s.CreatePart(h, r, length)
}

// CreateFormField calls CreatePart with a header using the
// given field name.
func (s *Streamer) CreateFormField(fieldname string, r io.Reader, length int64) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(fieldname)))
	return s.CreatePart(h, r, length)
}

// WriteField calls CreateFormField and then writes the given value.
func (s *Streamer) WriteField(fieldname, value string) error {
	_, err := s.mw.CreateFormField(fieldname)

	if err != nil {
		return err
	}

	s.parts = append(s.parts, &part{
		header: s.readHeader(),
		reader: bytes.NewReader([]byte(value)),
		length: int64(len(value)),
	})

	return nil
}

// Finalize is called when the reader is attempted to be read from, and it does not have an underlying
// reader to use. It will close the multipart writer, flushing the final part, and append all readers together
// making us a reader that is able to simply append all the parts as one stream.
func (s *Streamer) Finalize() error {
	if s.reader != nil {
		return nil
	}

	err := s.mw.Close()

	if err != nil {
		return err
	}

	// Final data
	final := s.readHeader()

	s.parts = append(s.parts, &part{
		header: final,
	})

	readers := make([]io.Reader, 0)

	for _, part := range s.parts {
		readers = append(readers, bytes.NewReader(part.header))

		if part.reader != nil {
			readers = append(readers, part.reader)
		}
	}

	s.reader = io.MultiReader(readers...)

	return nil
}

// Len should only be called after Finalize, which calculates the final length
func (s *Streamer) Len() int64 {
	var final int64

	for _, part := range s.parts {
		final += part.length + int64(len(part.header))
	}

	if s.reader == nil {
		// Append closing because we don't have it as a part yet.
		// if Finalize is called before this, this won't be called because it's also a part
		final += int64(len(fmt.Sprintf("\r\n--%s--\r\n", s.mw.Boundary())))
	}

	return final
}

// LenString is Len, but returned as a string for ease of use in headers
func (s *Streamer) LenString() string {
	return strconv.FormatInt(s.Len(), 10)
}

// ContentType returns the request's content type
func (s *Streamer) ContentType() string {
	return s.mw.FormDataContentType()
}

// Read implements an io.Reader that can be used to read from this multipart streamer
// This finalizes the request (with mw.Close) and prepares a reader for reading if it is not ready.
func (s *Streamer) Read(buf []byte) (int, error) {
	if s.reader == nil {
		// Construct io.MultiReader from parts
		if err := s.Finalize(); err != nil {
			return -1, err
		}
	}

	return s.reader.Read(buf)
}

// Close check every part that may be an io.ReadCloser
// this isn't necessary if the ReadClosers are manually closed after the request
// This function itself makes the Streamer an io.Closer
func (s *Streamer) Close() error {
	for _, p := range s.parts {
		if rc, ok := p.reader.(io.ReadCloser); ok {
			if err := rc.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
