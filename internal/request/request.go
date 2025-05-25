package request

import (
	"errors"
	"io"
	"strings"
)

const (
	initialized = iota
	done        = iota
)

type Request struct {
	RequestLine  RequestLine
	RequestState int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const BufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {

	requestParser := Request{}
	buffer := make([]byte, BufferSize)
	bytes_read := 0
	bytes_parsed := 0

	for requestParser.RequestState != done {

		if bytes_read == len(buffer) {
			// if we've filled the buffer, double the size and copy
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		// read into buffer starting at last index where read
		n, err := reader.Read(buffer[bytes_read:])
		if err != nil {
			if err == io.EOF {
				requestParser.RequestState = done
				break
			}
			return nil, err
		}

		bytes_read += n

		bytes_parsed, err = requestParser.parse(buffer)
		if err != nil {
			return nil, err
		}

		if bytes_parsed > 0 {
			// if we parsed some bytes, copy the remaining bytes to the front of the buffer
			copy(buffer, buffer[bytes_parsed:bytes_read])
			bytes_read -= bytes_parsed
		}

	}

	return &requestParser, nil
}

const alpha = "abcdefghijklmnopqrstuvwxyz"

func alphaOnly(s string) bool {
	for _, char := range s {
		if !strings.Contains(strings.ToUpper(alpha), string(char)) {
			return false
		}
	}
	return true
}

func parseRequestLine(requestString string) (*RequestLine, int, error) {
	if !strings.Contains(requestString, "\r\n") {
		return nil, 0, nil
	}

	parts := strings.Split(requestString, "\r\n")
	line := parts[0]

	parts = strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, 0, errors.New("expected 3 parts in request line, got: " + line)
	}

	method, requestTarget, httpVersion := parts[0], parts[1], parts[2]

	if !alphaOnly(method) {
		return nil, 0, errors.New("expected method to be alphanumeric, got: " + method)
	}

	if httpVersion != "HTTP/1.1" {
		return nil, 0, errors.New("expected HTTP version to be HTTP/1.1, got: " + httpVersion)
	}

	res := RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   strings.Split(httpVersion, "/")[1],
	}

	return &res, len(([]byte)(line)), nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.RequestState == done {
		return 0, errors.New("request already parsed")
	} else if r.RequestState == initialized {
		// parse request line
		requestLine, bytes, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}

		if bytes == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.RequestState = done
		return bytes, nil
	} else {
		return 0, errors.New("unknown state")
	}
}
