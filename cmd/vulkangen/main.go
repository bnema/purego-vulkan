package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/emitter"
	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/model"
	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/overrides"
	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/parser"
)

type config struct {
	registryPath string
	outDir       string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.registryPath, "registry", "./registry/vk.xml", "path to Vulkan vk.xml registry")
	flag.StringVar(&cfg.outDir, "out", ".", "project root output directory")
	flag.Parse()

	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	if cfg.registryPath == "" {
		return fmt.Errorf("registry path is required")
	}
	if cfg.outDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if info, err := os.Stat(cfg.outDir); err != nil {
		return fmt.Errorf("read output directory %s: %w", cfg.outDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("output path %s is not a directory", cfg.outDir)
	}

	reg, err := parser.ParseFile(cfg.registryPath)
	if err != nil {
		return err
	}
	sel, err := model.Select(reg, overrides.DefaultSelection())
	if err != nil {
		return err
	}
	return emitGeneratedFiles(cfg.outDir, sel)
}

func emitGeneratedFiles(outDir string, sel *model.SelectedRegistry) error {
	outputs := []struct {
		path string
		emit func(*model.SelectedRegistry) (string, error)
	}{
		{filepath.Join("vulkan", "types_gen.go"), emitter.EmitTypes},
		{filepath.Join("vulkan", "constants_gen.go"), emitter.EmitConstants},
		{filepath.Join("vulkan", "commands_gen.go"), emitter.EmitCommands},
		{filepath.Join("vulkan", "dispatch_gen.go"), emitter.EmitDispatch},
		{filepath.Join("vulkan", "strings_gen.go"), emitter.EmitStrings},
		{filepath.Join("internal", "capi", "register_gen.go"), emitter.EmitRegister},
	}
	for _, output := range outputs {
		code, err := output.emit(sel)
		if err != nil {
			return fmt.Errorf("emit %s: %w", output.path, err)
		}
		path := filepath.Join(outDir, output.path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create output directory for %s: %w", output.path, err)
		}
		if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", output.path, err)
		}
	}
	return nil
}
