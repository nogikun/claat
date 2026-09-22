package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validManual = `summary: Sample manual
id: sample-manual
categories: dev, docs
environments: Web
status: Draft
feedback link: https://example.com/issues
analytics account: UA-123

# Sample manual

## First step
Duration: 0:05:00

Explain the expected result.

~~~text
# this is not a page title
## this is not a step
~~~
`

func TestLintValidManual(t *testing.T) {
	path := writeManual(t, validManual)
	diagnostics, err := lintFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
}

func TestLintReportsContractErrors(t *testing.T) {
	path := writeManual(t, `summary: Sample
id: bad/id
categories: dev,
environments: Desktop
status: Live
feedback link:
analytics account: UA-123

# Sample

## Step
This should have a duration first.
`)
	diagnostics, err := lintFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"META003", "STEP002"} {
		if !hasCode(diagnostics, code) {
			t.Errorf("missing diagnostic %s: %v", code, diagnostics)
		}
	}
}

func TestLintReportsUnclosedFenceAndMissingBody(t *testing.T) {
	path := writeManual(t, `summary: Sample
id: sample
categories: dev
environments: Web
status: Draft
feedback link: https://example.com
analytics account: UA-123

# Sample

## Empty step
Duration: 0:05:00

~~~text
still open
`)
	diagnostics, err := lintFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"STEP003", "MD001"} {
		if !hasCode(diagnostics, code) {
			t.Errorf("missing diagnostic %s: %v", code, diagnostics)
		}
	}
}

func TestLintChecksLocalImages(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "ok.png"), []byte("image"), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "manual.md")
	contents := strings.Replace(validManual, "Explain the expected result.", "![ok](ok.png)\n![missing](missing.png)", 1)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	diagnostics, err := lintFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCode(diagnostics, "ASSET001") {
		t.Fatalf("missing image diagnostic: %v", diagnostics)
	}
}

func TestBuildDoesNotRunClaatWhenLintFails(t *testing.T) {
	path := writeManual(t, "# invalid\n")
	originalFind := findClaat
	originalRun := runClaat
	t.Cleanup(func() {
		findClaat = originalFind
		runClaat = originalRun
	})
	called := false
	findClaat = func(string) (string, error) {
		called = true
		return "claat", nil
	}
	runClaat = func(string, string, string, io.Writer, io.Writer) error {
		called = true
		return nil
	}
	code := run([]string{"build", path}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != exitLint || called {
		t.Fatalf("code=%d called=%v", code, called)
	}
}

func TestBuildRunsClaatAfterLint(t *testing.T) {
	path := writeManual(t, validManual)
	originalFind := findClaat
	originalRun := runClaat
	t.Cleanup(func() {
		findClaat = originalFind
		runClaat = originalRun
	})
	findClaat = func(string) (string, error) {
		return "fake-claat", nil
	}
	var gotPath, gotInput, gotOutput string
	runClaat = func(path, input, output string, _, _ io.Writer) error {
		gotPath, gotInput, gotOutput = path, input, output
		return nil
	}
	code := run([]string{"build", "-output", "generated", path}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != exitOK {
		t.Fatalf("code=%d", code)
	}
	if gotPath != "fake-claat" || gotInput != path || gotOutput != "generated" {
		t.Fatalf("claat args: path=%q input=%q output=%q", gotPath, gotInput, gotOutput)
	}
}

func writeManual(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manual.md")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func hasCode(diagnostics []diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.code == code {
			return true
		}
	}
	return false
}
