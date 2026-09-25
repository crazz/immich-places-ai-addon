//go:build !linux && !darwin && !freebsd

package main

import (
	"errors"
	"os"
)

func acquireAIWriteLock(string) (*os.File, error) {
	return nil, errors.New("writer runtime lock unsupported")
}
