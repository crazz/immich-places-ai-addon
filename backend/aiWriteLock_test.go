package main

import (
	"path/filepath"
	"testing"
)

func TestAIWriteRuntimeLockExcludesOtherProcessHandles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "writer.lock")
	first, err := acquireAIWriteLock(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := acquireAIWriteLock(path)
	if err == nil {
		second.Close()
		first.Close()
		t.Fatal("second runtime acquired writer lock")
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	third, err := acquireAIWriteLock(path)
	if err != nil {
		t.Fatal(err)
	}
	third.Close()
}
