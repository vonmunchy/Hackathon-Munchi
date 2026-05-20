// syncspec performs the small filesystem operations the Makefile needs,
// replacing the unix-only `cp`, `diff -q`, and `mkdir -p` invocations.
// It exists so `make build` works identically on macOS, Linux, and
// Windows without depending on an `sh`-style shell being in PATH.
//
// Usage:
//
//	go run ./tools/syncspec <src> <dst>           # copy src to dst
//	go run ./tools/syncspec -check <src> <dst>    # exit 1 if src != dst
//	go run ./tools/syncspec -mkdir <dir>          # mkdir -p <dir>
//	go run ./tools/syncspec -cat <path>           # print file contents
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	check := flag.Bool("check", false, "verify dst matches src; do not write")
	mkdir := flag.Bool("mkdir", false, "create the given directory tree (mkdir -p)")
	cat := flag.Bool("cat", false, "print file contents to stdout (trailing newline trimmed)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: syncspec [-check] <src> <dst>  |  syncspec -mkdir <dir>  |  syncspec -cat <path>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *mkdir {
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(2)
		}
		if err := os.MkdirAll(flag.Arg(0), 0o750); err != nil {
			fmt.Fprintf(os.Stderr, "syncspec: mkdir %q: %v\n", flag.Arg(0), err)
			os.Exit(1)
		}
		return
	}

	if *cat {
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(2)
		}
		data, err := os.ReadFile(flag.Arg(0)) //nolint:gosec // tool-local; path comes from Makefile.
		if err != nil {
			fmt.Fprintf(os.Stderr, "syncspec: read %q: %v\n", flag.Arg(0), err)
			os.Exit(1)
		}
		// Trim trailing newline so $(shell ...) callers don't have to
		// strip — matches how `cat` is used by the Makefile for short,
		// single-line metadata files.
		fmt.Print(strings.TrimRight(string(data), "\r\n"))
		return
	}

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	src, dst := flag.Arg(0), flag.Arg(1)

	if *check {
		if err := verify(src, dst); err != nil {
			fmt.Fprintf(os.Stderr, "syncspec: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := copyFile(src, dst); err != nil {
		fmt.Fprintf(os.Stderr, "syncspec: %v\n", err)
		os.Exit(1)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // tool-local; src comes from Makefile.
	if err != nil {
		return fmt.Errorf("open src %q: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(dst), err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), filepath.Base(dst)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp in %q: %w", filepath.Dir(dst), err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("copy to temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		cleanup()
		return fmt.Errorf("rename %q -> %q: %w", tmpName, dst, err)
	}
	return nil
}

func verify(src, dst string) error {
	a, err := os.ReadFile(src) //nolint:gosec // tool-local; src comes from Makefile.
	if err != nil {
		return fmt.Errorf("read src %q: %w", src, err)
	}
	b, err := os.ReadFile(dst) //nolint:gosec // tool-local; dst comes from Makefile.
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s is missing — run 'make sync-spec'", dst)
		}
		return fmt.Errorf("read dst %q: %w", dst, err)
	}
	if !bytes.Equal(a, b) {
		return fmt.Errorf("%s is out of sync with %s — run 'make sync-spec'", dst, src)
	}
	return nil
}
