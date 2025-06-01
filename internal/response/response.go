package response

import (
	"io"
	"strconv"

	"github.com/jsleep/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	successCode             StatusCode = 200
	badRequestCode          StatusCode = 400
	internalServerErrorCode StatusCode = 500
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := "HTTP/1.1 " + strconv.Itoa(int(statusCode)) + " " + statusText(statusCode) + "\r\n"
	_, err := w.Write([]byte(statusLine))
	return err
}

func statusText(code StatusCode) string {
	switch code {
	case successCode:
		return "OK"
	case badRequestCode:
		return "Bad Request"
	case internalServerErrorCode:
		return "Internal Server Error"
	default:
		return "Unknown Status"
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": strconv.Itoa(contentLen),
		"Content-Type":   "text/plain",
		"Connection":     "close",
	}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, value := range headers {
		headerLine := key + ": " + value + "\r\n"
		if _, err := w.Write([]byte(headerLine)); err != nil {
			return err
		}
	}
	// Write the final CRLF to indicate the end of headers
	w.Write([]byte("\r\n"))

	return nil
}

func (w *Writer) WriteBody(Body []byte) (int, error) {
	n, err := w.Write(Body)
	if err != nil {
		return 0, err
	}
	return n, err
}

type Writer struct {
	io.Writer
}
