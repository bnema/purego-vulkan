package puregovulkan_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

const generatedPathspec = ":(glob)**/*_gen.go"

func TestGeneratedFilesAreFresh(t *testing.T) {
	if _, err := os.Stat(".git"); err != nil {
		t.Skip("git metadata unavailable")
	}

	before := generatedState(t)
	cmd := exec.Command("go", "generate", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go generate ./...: %v", err)
	}
	after := generatedState(t)
	if !bytes.Equal(before, after) {
		t.Fatalf("generated files are stale; go generate changed generated output\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

func generatedState(t *testing.T) []byte {
	t.Helper()
	var state bytes.Buffer
	for _, args := range [][]string{
		{"diff", "--", generatedPathspec},
		{"ls-files", "--others", "--exclude-standard", "--", generatedPathspec},
	} {
		cmd := exec.Command("git", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		fmt.Fprintf(&state, "$ git %v\n", args)
		state.Write(out)
	}
	return state.Bytes()
}
