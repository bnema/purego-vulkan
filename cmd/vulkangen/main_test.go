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
	for _, profile := range []overrides.Profile{overrides.ProfileRenderer, overrides.ProfileWSI, overrides.ProfileComplete} {
		t.Run(string(profile), func(t *testing.T) {
			outDir := t.TempDir()
			if err := run(config{
				registryPath: filepath.Join("..", "..", "registry", "vk.xml"),
				outDir:       outDir,
				profile:      string(profile),
			}); err != nil {
				t.Fatalf("run(%s) error = %v", profile, err)
			}
			writeGeneratedCompileHarness(t, outDir)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, "go", "test", "./...")
			cmd.Dir = outDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("generated %s profile did not compile: %v\n%s", profile, err, out)
			}
		})
	}
}

func writeGeneratedCompileHarness(t *testing.T, outDir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(outDir, "go.mod"), []byte("module generatedprofile\n\ngo 1.24\n"), 0o644); err != nil {
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
