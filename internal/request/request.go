package request

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/jsleep/httpfromtcp/internal/headers"
)

const (
	initialized    = iota
	parsingHeaders = iota
	parsingBody    = iota
	done           = iota
)

type Request struct {
	RequestLine RequestLine
	State       int
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const BufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {

	requestParser := Request{Headers: headers.NewHeaders(), State: initialized, Body: make([]byte, 0)}
	buffer := make([]byte, BufferSize)
	bytes_read := 0
	bytes_parsed := 0
	hitEOF := false
	last_parsed := 0

	for requestParser.State != done {

		if hitEOF && last_parsed == 0 {
			// if we hit EOF and nothing was parsed, we can't continue
			return nil, errors.New("EOF and can't parse further")
		}

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
				hitEOF = true
			} else {
				return nil, err
			}
		}
		bytes_read += n

		// parse the request starting from last parsed index
		n, err = requestParser.parse(buffer[bytes_parsed:bytes_read])
		//fmt.Println(requestParser.RequestState)
		if err != nil {
			return nil, err
		}

		last_parsed = n
		bytes_parsed += n
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

	return &res, len(line) + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.State == done {
		return 0, errors.New("request already parsed")
	} else if r.State == initialized {
		// parse request line
		requestLine, bytes, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}

		if bytes == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.State = parsingHeaders
		return bytes, nil
	} else if r.State == parsingHeaders {
		// parse headers
		bytes, finished, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if finished {
			bytes += 2 // account for the \r\n after headers
			if r.Headers.Get("content-length") != "" {
				r.State = parsingBody
			} else {
				r.State = done
			}
		}

		return bytes, nil
	} else if r.State == parsingBody {
		// parse body
		length, err := r.parseBody(data)
		if err != nil {
			return 0, err
		}

		return length, nil
	} else {
		return 0, errors.New("unknown state")
	}
}

func (r *Request) parseBody(data []byte) (int, error) {
	length, err := strconv.Atoi(r.Headers.Get("content-length"))
	if err != nil {
		return 0, errors.New("invalid content-length header: " + r.Headers.Get("content-length"))
	}

	bytes := min(length-len(r.Body), len(data))

	r.Body = append(r.Body, data[:bytes]...)

	if len(r.Body) > length {
		return 0, errors.New("body length exceeds content-length header, expected: " + strconv.Itoa(length) + ", got: " + strconv.Itoa(len(r.Body)) + " full body: " + string(r.Body) + " full data: " + string(data))
	}

	if len(r.Body) == length {
		r.State = done
	}

	return bytes, nil
}
