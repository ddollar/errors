package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ddollar/errors"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	}
}

func run() error {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		if strings.HasPrefix(path, "vendor/") {
			return nil
		}

		fmt.Printf("processing: %s\n", path) //nolint:forbidigo

		if err := rewriteErrors(path); err != nil {
			return err
		}

		if err := wrapFile(path); err != nil {
			return err
		}

		return nil
	})

	return errors.Wrap(err)
}

func rewriteErrors(path string) error {
	rewrites := []string{
		`"errors" -> "github.com/ddollar/errors"`,
		`"github.com/pkg/errors" -> "github.com/ddollar/errors"`,
		"errors.WithStack -> errors.Wrap",
		"fmt.Errorf -> errors.Errorf",
	}

	for _, rewrite := range rewrites {

		if data, err := exec.Command("gofmt", "-w", "-r", rewrite, path).CombinedOutput(); err != nil {
			return errors.New(string(data))
		}
	}

	return nil
}

var doubleWrap = regexp.MustCompile(`errors\.Wrap\(errors\.(.*?)\)\)`)

func wrapFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return errors.Wrap(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err)
	}

	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		wl, err := wrapLine(line)
		if err != nil {
			return err
		}

		wl = doubleWrap.ReplaceAllString(wl, "errors.$1)")

		lines[i] = wl
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), info.Mode()); err != nil {
		return errors.Wrap(err)
	}

	if data, err := exec.Command("goimports", "-w", path).CombinedOutput(); err != nil {
		return errors.New(string(data))
	}

	return nil
}

var noWrapMatcher = regexp.MustCompile(`//(.*?)nowrap`)

func wrapLine(line string) (string, error) {
	trim := strings.TrimSpace(line)

	if !strings.HasPrefix(trim, "return ") {
		return line, nil
	}

	if noWrapMatcher.MatchString(trim) {
		return line, nil
	}

	args := tokenizeArgs(strings.TrimPrefix(trim, "return "))

	for i, arg := range args {
		if wrappable(arg) {
			args[i] = fmt.Sprintf("errors.Wrap(%s)", arg)
		}
	}

	parts := strings.Split(line, "return ")

	return fmt.Sprintf("%sreturn %s", parts[0], strings.Join(args, ", ")), nil
}

func tokenizeArgs(args string) []string {
	tokens := []string{""}
	i := 0
	depth := 0

	for _, r := range args {
		if r == ',' && depth == 0 {
			tokens = append(tokens, "")
			i += 1
			continue
		}

		if r == '(' {
			depth += 1
		}

		if r == ')' {
			depth -= 1
		}

		tokens[i] += string(r)
	}

	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}

	return tokens
}

func wrappable(arg string) bool {
	if arg == "err" {
		return true
	}

	if strings.HasPrefix(arg, "errors.New") {
		return true
	}

	if strings.HasPrefix(arg, "fmt.Errorf") {
		return true
	}

	if strings.HasPrefix(arg, "log.Error") {
		return true
	}

	return false
}
