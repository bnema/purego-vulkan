package main

import (
	"flag"
	"fmt"
	"os"

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
	if _, err := model.Select(reg, overrides.DefaultSelection()); err != nil {
		return err
	}
	return nil
}
