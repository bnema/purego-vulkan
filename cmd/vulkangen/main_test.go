package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/overrides"
)

func TestGenerationProfilesEmitCompilablePackages(t *testing.T) {
	tests := []struct {
		name    string
		profile string
	}{
		{name: "default", profile: ""},
		{name: string(overrides.ProfileRenderer), profile: string(overrides.ProfileRenderer)},
		{name: string(overrides.ProfileWSI), profile: string(overrides.ProfileWSI)},
		{name: string(overrides.ProfileComplete), profile: string(overrides.ProfileComplete)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outDir := t.TempDir()
			if err := run(config{
				registryPath: filepath.Join("..", "..", "registry", "vk.xml"),
				outDir:       outDir,
				profile:      tt.profile,
			}); err != nil {
				t.Fatalf("run(%s) error = %v", tt.name, err)
			}
			writeGeneratedCompileHarness(t, outDir)
			tidy := exec.Command("go", "mod", "tidy")
			tidy.Dir = outDir
			if out, err := tidy.CombinedOutput(); err != nil {
				t.Fatalf("tidy generated module: %v: %s", err, out)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, "go", "test", "./...")
			cmd.Dir = outDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("generated %s profile did not compile: %v\n%s", tt.name, err, out)
			}
		})
	}
}

func writeGeneratedCompileHarness(t *testing.T, outDir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(outDir, "go.mod"), []byte("module generatedprofile\n\ngo 1.25\n\nrequire github.com/bnema/purego v0.11.0-bnema.3\n\nreplace github.com/bnema/purego => /home/brice/dev/projects/purego/.worktrees/typed-callbacks\n"), 0o644); err != nil {
		t.Fatalf("write generated go.mod: %v", err)
	}
	registerPath := filepath.Join(outDir, "internal", "capi", "register.go")
	if err := os.MkdirAll(filepath.Dir(registerPath), 0o750); err != nil {
		t.Fatalf("create capi dir: %v", err)
	}
	if err := os.WriteFile(registerPath, []byte("package capi\n\nfunc RegisterFunc(any, uintptr) {}\n"), 0o644); err != nil {
		t.Fatalf("write capi register stub: %v", err)
	}
}
