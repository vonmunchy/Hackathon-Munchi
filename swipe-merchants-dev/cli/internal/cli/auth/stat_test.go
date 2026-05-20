package auth_test

import "os"

// statFunc is split out so the main test file doesn't need os/syscall
// imports. Defined here for keyword-search clarity.
func statFunc(path string) (any, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return info, nil
}
