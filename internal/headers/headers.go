package headers

import (
	"errors"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[strings.ToLower(key)]
	return val, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	str := string(data)
	if strings.HasPrefix(str, "\r\n") {
		return 0, true, nil
	}

	if !strings.Contains(str, "\r\n") {
		return 0, false, nil
	}

	header := strings.Split(str, "\r\n")[0]
	key, val, found := strings.Cut(strings.TrimSpace(header), ":")
	if !found {
		return 0, false, errors.New("invalid header")
	}

	match, err := regexp.MatchString("^[A-Za-z\\d!#$%&'*+\\-.^_`|~]+$", key)
	if err != nil {
		return 0, false, err
	}

	if !match {
		return 0, false, errors.New("invalid header key")
	}

	if _, ok := h[strings.ToLower(key)]; ok {
		h[strings.ToLower(key)] = h[strings.ToLower(key)] + ", " + strings.TrimSpace(val)
	} else {
		h[strings.ToLower(key)] = strings.TrimSpace(val)
	}

	return len(header) + 2, false, nil
}
