//go:build e2e

package e2e_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIProcessContract(t *testing.T) {
	root := repositoryRoot(t)
	binary := filepath.Join(t.TempDir(), "calathea")

	build := exec.Command("go", "build", "-o", binary, "./cmd/calathea")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	t.Run("default command", func(t *testing.T) {
		code, stdout, stderr := runCLI(t, binary)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
		if strings.TrimSpace(stdout) == "" {
			t.Fatal("stdout is empty, want human-facing orientation output")
		}
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}
	})

	t.Run("version", func(t *testing.T) {
		code, stdout, stderr := runCLI(t, binary, "version")
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}

		fields := strings.Fields(stdout)
		if len(fields) != 2 || fields[0] != "calathea" || fields[1] == "" {
			t.Fatalf("stdout = %q, want version identity in 'calathea <version>' form", stdout)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		code, stdout, stderr := runCLI(t, binary, "unknown")
		if code != 2 {
			t.Fatalf("exit code = %d, want 2", code)
		}
		if stdout != "" {
			t.Fatalf("stdout = %q, want empty", stdout)
		}
		if strings.TrimSpace(stderr) == "" {
			t.Fatal("stderr is empty, want human-facing invocation diagnostic")
		}
	})
}

func runCLI(t *testing.T, binary string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run CLI: %v", err)
		}
		code = exitErr.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get test working directory: %v", err)
	}

	for {
		moduleFile := filepath.Join(dir, "go.mod")
		info, statErr := os.Stat(moduleFile)
		if statErr == nil && !info.IsDir() {
			return dir
		}
		if statErr != nil && !os.IsNotExist(statErr) {
			t.Fatalf("inspect %s: %v", moduleFile, statErr)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatal("could not locate repository go.mod from test working directory")
	return ""
}
