package main

import (
	"fmt"
	"regexp"
)

// safeParam covers: IQNs, IP:port, WWIDs, device names, map names
// Examples: iqn.2001-04.com.ibm:storage.9.71.253.36, 10.0.0.1:3260, 600507681abc
var safeParam = regexp.MustCompile(`^[a-zA-Z0-9.:\-_/]+$`)

func sanitize(key, value string) error {
	if value == "" {
		return fmt.Errorf("param %q is empty", key)
	}
	if !safeParam.MatchString(value) {
		return fmt.Errorf("param %q contains invalid characters: %q", key, value)
	}
	return nil
}
