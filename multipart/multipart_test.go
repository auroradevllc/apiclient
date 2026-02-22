package multipart

import (
	"bytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
)

var _ = Describe("Multipart Streaming", func() {
	var s *Streamer
	BeforeEach(func() {
		s = New()
	})
	Context("Length matching", func() {
		Context("Basic fields", func() {
			It("Should create a request with a basic field, matching Len before Finalize", func() {
				Expect(s.WriteField("test", "test123")).To(BeNil())

				// This is called before reading, so Finalize is never called
				lenBefore := s.Len()

				var b bytes.Buffer
				_, _ = io.Copy(&b, s)

				Expect(lenBefore).To(Equal(s.Len()))
				Expect(b.Len()).To(BeEquivalentTo(s.Len()))
			})
			It("Should create a request with a basic field, matching Len after Finalize", func() {
				Expect(s.WriteField("test", "test123")).To(BeNil())

				var b bytes.Buffer
				_, _ = io.Copy(&b, s)

				Expect(b.Len()).To(BeEquivalentTo(s.Len()))
			})
		})
		Context("Files", func() {
			It("Should create a request with a file, matching Len before Finalize", func() {
				Expect(s.CreateFormFile("file", "test.txt", strings.NewReader("something"), 9))

				// This is called before reading, so Finalize is never called
				lenBefore := s.Len()

				var b bytes.Buffer
				_, _ = io.Copy(&b, s)

				Expect(lenBefore).To(Equal(s.Len()))
				Expect(b.Len()).To(BeEquivalentTo(s.Len()))
			})
			It("Should create a request with a file, matching Len after Finalize", func() {
				Expect(s.CreateFormFile("file", "test.txt", strings.NewReader("something"), 9))

				var b bytes.Buffer
				_, _ = io.Copy(&b, s)

				Expect(b.Len()).To(BeEquivalentTo(s.Len()))
			})
		})
		Context("Parts with headers", func() {
			It("Should create a request with a part, matching Len before Finalize", func() {
				h := make(textproto.MIMEHeader)
				h.Set("X-Test", "test")
				Expect(s.CreatePart(h, strings.NewReader("something"), 9)).To(BeNil())

				// This is called before reading, so Finalize is never called
				lenBefore := s.Len()

				var b bytes.Buffer
				_, _ = io.Copy(&b, s)

				Expect(lenBefore).To(Equal(s.Len()))
				Expect(b.Len()).To(BeEquivalentTo(s.Len()))
			})
			It("Should create a request with a part, matching Len after Finalize", func() {
				h := make(textproto.MIMEHeader)
				h.Set("X-Test", "test")
				Expect(s.CreatePart(h, strings.NewReader("something"), 9)).To(BeNil())

				var b bytes.Buffer
				_, _ = io.Copy(&b, s)

				Expect(b.Len()).To(BeEquivalentTo(s.Len()))
			})
		})
		Context("Combined fields + files", func() {
			It("Should create a request with a field and file, matching Len before Finalize", func() {
				Expect(s.WriteField("test", "test123")).To(BeNil())
				Expect(s.CreateFormFile("file", "test.txt", strings.NewReader("something"), 9))

				// This is called before reading, so Finalize is never called
				lenBefore := s.Len()

				var b bytes.Buffer
				_, _ = io.Copy(&b, s)

				Expect(lenBefore).To(Equal(s.Len()))
				Expect(b.Len()).To(BeEquivalentTo(s.Len()))
			})
			It("Should create a request with a field and file, matching Len after Finalize", func() {
				Expect(s.WriteField("test", "test123")).To(BeNil())
				Expect(s.CreateFormFile("file", "test.txt", strings.NewReader("something"), 9))

				var b bytes.Buffer
				_, _ = io.Copy(&b, s)

				Expect(b.Len()).To(BeEquivalentTo(s.Len()))
			})
		})
	})
	Context("Requests", func() {
		It("Should perform a successful request to a test server", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.ParseMultipartForm(10 * 1024 * 1024)).To(BeNil())

				Expect(r.FormValue("test")).To(Equal("testing123"))

				f, h, err := r.FormFile("file")

				Expect(err).To(BeNil())
				Expect(h.Filename).To(Equal("file.txt"))

				b, _ := io.ReadAll(f)

				Expect(string(b)).To(Equal("something"))

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))

			defer server.Close()

			Expect(s.WriteField("test", "testing123")).To(BeNil())
			Expect(s.CreateFormFile("file", "file.txt", strings.NewReader("something"), 9)).To(BeNil())

			req, err := http.NewRequest(http.MethodPost, server.URL, s)

			Expect(err).To(BeNil())

			req.Header.Set("Content-Type", s.ContentType())
			req.Header.Set("Content-Length", s.LenString())

			res, err := http.DefaultClient.Do(req)

			Expect(err).To(BeNil())

			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusOK))

			b, err := io.ReadAll(res.Body)

			Expect(err).To(BeNil())

			Expect(string(b)).To(Equal("OK"))
		})
	})
})
