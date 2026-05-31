package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRegistryFixture(t *testing.T) {
	reg, err := ParseFile(filepath.Join("testdata", "minimal_registry.xml"))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if len(reg.Types) != 9 {
		t.Fatalf("types = %d, want 9", len(reg.Types))
	}
	instance := reg.TypeByName("VkInstance")
	if instance == nil {
		t.Fatal("VkInstance type missing")
	}
	if instance.Category != "handle" || instance.Type != "VK_DEFINE_HANDLE" {
		t.Fatalf("VkInstance = %+v", *instance)
	}

	image := reg.TypeByName("VkImage")
	if image == nil {
		t.Fatal("VkImage type missing")
	}
	if image.Parent != "VkDevice" || image.Type != "VK_DEFINE_NON_DISPATCHABLE_HANDLE" {
		t.Fatalf("VkImage = %+v", *image)
	}

	info := reg.TypeByName("VkInstanceCreateInfo")
	if info == nil {
		t.Fatal("VkInstanceCreateInfo type missing")
	}
	if len(info.Members) != 4 {
		t.Fatalf("members = %d, want 4", len(info.Members))
	}
	if info.Members[1].Name != "pNext" || info.Members[1].Type != "void" || !info.Members[1].Const || info.Members[1].PointerDepth != 1 {
		t.Fatalf("pNext member parsed as %+v", info.Members[1])
	}
	if info.Members[3].Name != "ppEnabledExtensionNames" || info.Members[3].PointerDepth != 2 || info.Members[3].Len != "enabledExtensionCount" {
		t.Fatalf("ppEnabledExtensionNames member parsed as %+v", info.Members[3])
	}

	if len(reg.EnumGroups) != 1 || len(reg.EnumGroups[0].Enums) != 2 {
		t.Fatalf("enum groups = %+v", reg.EnumGroups)
	}

	cmd := reg.CommandByName("vkCreateInstance")
	if cmd == nil {
		t.Fatal("vkCreateInstance missing")
	}
	if cmd.Return != "VkResult" || len(cmd.Params) != 3 {
		t.Fatalf("vkCreateInstance = %+v", *cmd)
	}
	if cmd.Params[0].Name != "pCreateInfo" || cmd.Params[0].Type != "VkInstanceCreateInfo" || !cmd.Params[0].Const || cmd.Params[0].PointerDepth != 1 {
		t.Fatalf("pCreateInfo parsed as %+v", cmd.Params[0])
	}

	if len(reg.Features) != 1 || reg.Features[0].Name != "VK_VERSION_1_0" {
		t.Fatalf("features = %+v", reg.Features)
	}
	aliasType := reg.TypeByName("VkAccessFlags2KHR")
	if aliasType == nil || aliasType.Alias != "VkAccessFlags2" {
		t.Fatalf("alias type parsed as %+v", aliasType)
	}
	matrix := reg.TypeByName("VkTransformMatrixKHR")
	if matrix == nil || len(matrix.Members) != 1 || len(matrix.Members[0].ArrayLens) != 2 || matrix.Members[0].ArrayLens[0] != "VK_MAX_MATRIX_ROWS" || matrix.Members[0].ArrayLens[1] != "4" {
		t.Fatalf("matrix member parsed as %+v", matrix)
	}
	aliasCommand := reg.CommandByName("vkGetPhysicalDeviceProperties2KHR")
	if aliasCommand == nil || aliasCommand.Alias != "vkGetPhysicalDeviceProperties2" {
		t.Fatalf("alias command parsed as %+v", aliasCommand)
	}

	if len(reg.Extensions) != 2 || reg.Extensions[1].Name != "VK_KHR_external_memory_fd" {
		t.Fatalf("extensions = %+v", reg.Extensions)
	}
	ext := reg.Extensions[1]
	if ext.Type != "device" || len(ext.Requires) != 1 || ext.Requires[0].Commands[0] != "vkGetMemoryFdKHR" {
		t.Fatalf("extension parsed as %+v", ext)
	}
	if ext.Requires[0].Enums[0].Value != "\"VK_KHR_external_memory_fd\"" {
		t.Fatalf("extension enum parsed as %+v", ext.Requires[0].Enums[0])
	}
}

func TestParseRejectsEmptyRegistry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.xml")
	if err := os.WriteFile(path, []byte("<registry></registry>"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("ParseFile() error = nil")
	}
}
