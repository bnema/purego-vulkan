package emitter

import (
	"strings"
	"testing"

	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/model"
)

func TestEmitTypes(t *testing.T) {
	out, err := EmitTypes(testSelectedRegistry())
	if err != nil {
		t.Fatalf("EmitTypes() error = %v", err)
	}
	for _, want := range []string{
		"type Instance uintptr",
		"type PhysicalDevice uintptr",
		"type Image uint64",
		"type Bool32 uint32",
		"type Result int32",
		"type PFN_vkVoidFunction uintptr",
		"type AccessFlags2KHR = AccessFlags2",
		"type PhysicalDeviceDrmPropertiesEXT struct",
		"HasPrimary   Bool32",
		"type PhysicalDeviceProperties struct",
		"DeviceName [MaxPhysicalDeviceNameSize]byte",
		"type ImageDrmFormatModifierExplicitCreateInfoEXT struct",
		"PlaneLayouts                *SubresourceLayout",
		"type ClearColorValue [4]uint32",
		"type ClearValue [4]uint32",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitTypes() missing %q\n%s", want, out)
		}
	}
	for _, bad := range []string{
		"type ClearColorValue struct",
		"type ClearValue struct",
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("EmitTypes() unexpectedly emitted union as struct %q\n%s", bad, out)
		}
	}
}

func TestEmitConstants(t *testing.T) {
	out, err := EmitConstants(testSelectedRegistry())
	if err != nil {
		t.Fatalf("EmitConstants() error = %v", err)
	}
	for _, want := range []string{
		"const Success Result = 0",
		"const ErrorOutOfHostMemory Result = -1",
		"const KHRExternalMemoryFDExtensionName = \"VK_KHR_external_memory_fd\"",
		"const QueueFamilyExternalKHR = ^uint32(1)",
		"const WholeSize = ^uint64(0)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitConstants() missing %q\n%s", want, out)
		}
	}
}

func TestEmitRegisterHelpers(t *testing.T) {
	out, err := EmitRegister(testSelectedRegistry())
	if err != nil {
		t.Fatalf("EmitRegister() error = %v", err)
	}
	for _, want := range []string{
		"func RegisterGlobal(handle uintptr, lookup LookupFunc, fptrs map[string]any) error",
		"registerRequired([]string{\"vkCreateInstance\"}, handle, lookup, fptrs)",
		"registerRequired([]string{\"vkGetPhysicalDeviceProperties2\", \"vkGetPhysicalDeviceProperties2KHR\"}, handle, lookup, fptrs)",
		"func RegisterDevice(handle uintptr, lookup LookupFunc, fptrs map[string]any) error",
		"registerOptional([]string{\"vkGetMemoryFdKHR\"}, handle, lookup, fptrs)",
		"addr == 0",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitRegister() missing %q\n%s", want, out)
		}
	}
}

func TestEmitTypesDefinesOpaqueNativeBasetypes(t *testing.T) {
	sel := &model.SelectedRegistry{Types: []model.SelectedType{
		{Name: "ANativeWindow", GoName: "ANativeWindow", Category: "basetype", GoType: "uintptr", Source: model.TypeDecl{Category: "basetype", RawText: "struct ANativeWindow ;"}},
		{Name: "VkRemoteAddressNV", GoName: "RemoteAddressNV", Category: "basetype", GoType: "unsafe.Pointer", Source: model.TypeDecl{Category: "basetype", Type: "void"}},
	}}
	out, err := EmitTypes(sel)
	if err != nil {
		t.Fatalf("EmitTypes() error = %v", err)
	}
	for _, want := range []string{
		"type ANativeWindow uintptr",
		"type RemoteAddressNV unsafe.Pointer",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitTypes() missing %q\n%s", want, out)
		}
	}
}

func TestEmitTypesUsesSafeRepresentationsForNativePlatformHandles(t *testing.T) {
	sel := &model.SelectedRegistry{Types: []model.SelectedType{
		{Name: "VkWaylandSurfaceCreateInfoKHR", GoName: "WaylandSurfaceCreateInfoKHR", Category: "struct", Members: []model.MemberDecl{
			{Name: "display", Type: "wl_display", PointerDepth: 1},
			{Name: "surface", Type: "wl_surface", PointerDepth: 1},
		}},
		{Name: "VkXcbSurfaceCreateInfoKHR", GoName: "XcbSurfaceCreateInfoKHR", Category: "struct", Members: []model.MemberDecl{
			{Name: "connection", Type: "xcb_connection_t", PointerDepth: 1},
			{Name: "window", Type: "xcb_window_t"},
		}},
	}}
	out, err := EmitTypes(sel)
	if err != nil {
		t.Fatalf("EmitTypes() error = %v", err)
	}
	for _, want := range []string{
		"import \"unsafe\"",
		"Display unsafe.Pointer",
		"Surface unsafe.Pointer",
		"Connection unsafe.Pointer",
		"Window     uint32",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitTypes() missing %q\n%s", want, out)
		}
	}
	for _, bad := range []string{"*wl_display", "*wl_surface", "*xcb_connection_t", "xcb_window_t"} {
		if strings.Contains(out, bad) {
			t.Fatalf("EmitTypes() emitted native C type %q\n%s", bad, out)
		}
	}
}

func TestEmitCommands(t *testing.T) {
	out, err := EmitCommands(testSelectedRegistry())
	if err != nil {
		t.Fatalf("EmitCommands() error = %v", err)
	}
	for _, want := range []string{
		"var VkCreateInstance func(*InstanceCreateInfo, *AllocationCallbacks, *Instance) Result",
		"var VkDestroyInstance func(Instance, *AllocationCallbacks)",
		"var VkGetMemoryFdKHR func(Device, *MemoryGetFdInfoKHR, *int32) Result",
		"var VkMapMemory func(Device, DeviceMemory, DeviceSize, DeviceSize, MemoryMapFlags, *unsafe.Pointer) Result",
		"func globalCommandPointers() map[string]any",
		`"vkCreateInstance": &VkCreateInstance`,
		"func deviceCommandPointers() map[string]any",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitCommands() missing %q\n%s", want, out)
		}
	}
}

func TestEmitDispatch(t *testing.T) {
	out, err := EmitDispatch(testSelectedRegistry())
	if err != nil {
		t.Fatalf("EmitDispatch() error = %v", err)
	}
	for _, want := range []string{
		"type GlobalDispatch struct",
		"CreateInstance func(*InstanceCreateInfo, *AllocationCallbacks, *Instance) Result",
		"type InstanceDispatch struct",
		"Instance                        Instance",
		"DestroyInstance",
		"type DeviceDispatch struct",
		"Device         Device",
		"GetMemoryFdKHR func(Device, *MemoryGetFdInfoKHR, *int32) Result",
		"MapMemory      func(Device, DeviceMemory, DeviceSize, DeviceSize, MemoryMapFlags, *unsafe.Pointer) Result",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitDispatch() missing %q\n%s", want, out)
		}
	}
}

func TestEmitCommandsOmitsUnsafeImportWhenUnused(t *testing.T) {
	out, err := EmitCommands(testSelectedRegistryNoUnsafe())
	if err != nil {
		t.Fatalf("EmitCommands() error = %v", err)
	}
	for _, want := range []string{
		"var VkCreateInstance func(*Instance) Result",
		`"vkCreateInstance": &VkCreateInstance`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitCommands() missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "import \"unsafe\"") {
		t.Fatalf("EmitCommands() unexpectedly imported unsafe\n%s", out)
	}
}

func TestEmitDispatchOmitsUnsafeImportWhenUnused(t *testing.T) {
	out, err := EmitDispatch(testSelectedRegistryNoUnsafe())
	if err != nil {
		t.Fatalf("EmitDispatch() error = %v", err)
	}
	for _, want := range []string{
		"type GlobalDispatch struct",
		"CreateInstance func(*Instance) Result",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EmitDispatch() missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "import \"unsafe\"") {
		t.Fatalf("EmitDispatch() unexpectedly imported unsafe\n%s", out)
	}
}

func testSelectedRegistry() *model.SelectedRegistry {
	return &model.SelectedRegistry{
		Types: []model.SelectedType{
			{Name: "VkInstance", GoName: "Instance", GoType: "uintptr", Category: "handle", Dispatchable: true},
			{Name: "VkPhysicalDevice", GoName: "PhysicalDevice", GoType: "uintptr", Category: "handle", Dispatchable: true},
			{Name: "VkDevice", GoName: "Device", GoType: "uintptr", Category: "handle", Dispatchable: true},
			{Name: "VkImage", GoName: "Image", GoType: "uint64", Category: "handle"},
			{Name: "VkDeviceMemory", GoName: "DeviceMemory", GoType: "uint64", Category: "handle"},
			{Name: "VkBool32", GoName: "Bool32", GoType: "uint32", Category: "basetype"},
			{Name: "VkResult", GoName: "Result", GoType: "int32", Category: "basetype"},
			{Name: "VkStructureType", GoName: "StructureType", GoType: "int32", Category: "enum"},
			{Name: "PFN_vkVoidFunction", GoName: "PFN_vkVoidFunction", GoType: "uintptr", Category: "funcpointer"},
			{Name: "VkDeviceSize", GoName: "DeviceSize", GoType: "uint64", Category: "basetype"},
			{Name: "VkMemoryMapFlags", GoName: "MemoryMapFlags", GoType: "uint32", Category: "bitmask"},
			{Name: "VkAccessFlags2", GoName: "AccessFlags2", GoType: "uint64", Category: "bitmask"},
			{Name: "VkAccessFlags2KHR", GoName: "AccessFlags2KHR", GoType: "uint64", Category: "bitmask", Source: model.TypeDecl{Alias: "VkAccessFlags2"}},
			{Name: "VkAllocationCallbacks", GoName: "AllocationCallbacks", Category: "struct"},
			{Name: "VkPhysicalDeviceProperties", GoName: "PhysicalDeviceProperties", Category: "struct", Members: []model.MemberDecl{
				{Name: "deviceName", Type: "char", ArrayLens: []string{"VK_MAX_PHYSICAL_DEVICE_NAME_SIZE"}},
			}},
			{Name: "VkInstanceCreateInfo", GoName: "InstanceCreateInfo", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType"},
				{Name: "pNext", Type: "void", PointerDepth: 1},
			}},
			{Name: "VkMemoryGetFdInfoKHR", GoName: "MemoryGetFdInfoKHR", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType"},
				{Name: "pNext", Type: "void", Const: true, PointerDepth: 1},
				{Name: "memory", Type: "VkDeviceMemory"},
			}},
			{Name: "VkPhysicalDeviceDrmPropertiesEXT", GoName: "PhysicalDeviceDrmPropertiesEXT", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType"},
				{Name: "pNext", Type: "void", PointerDepth: 1},
				{Name: "hasPrimary", Type: "VkBool32"},
				{Name: "hasRender", Type: "VkBool32"},
				{Name: "primaryMajor", Type: "int64_t"},
			}},
			{Name: "VkImageDrmFormatModifierExplicitCreateInfoEXT", GoName: "ImageDrmFormatModifierExplicitCreateInfoEXT", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType"},
				{Name: "pNext", Type: "void", Const: true, PointerDepth: 1},
				{Name: "drmFormatModifier", Type: "uint64_t"},
				{Name: "drmFormatModifierPlaneCount", Type: "uint32_t"},
				{Name: "pPlaneLayouts", Type: "VkSubresourceLayout", Const: true, PointerDepth: 1},
			}},
			{Name: "VkClearColorValue", GoName: "ClearColorValue", Category: "union", Members: []model.MemberDecl{
				{Name: "float32", Type: "float", ArrayLens: []string{"4"}},
				{Name: "int32", Type: "int32_t", ArrayLens: []string{"4"}},
				{Name: "uint32", Type: "uint32_t", ArrayLens: []string{"4"}},
			}},
			{Name: "VkClearDepthStencilValue", GoName: "ClearDepthStencilValue", Category: "struct", Members: []model.MemberDecl{
				{Name: "depth", Type: "float"},
				{Name: "stencil", Type: "uint32_t"},
			}},
			{Name: "VkClearValue", GoName: "ClearValue", Category: "union", Members: []model.MemberDecl{
				{Name: "color", Type: "VkClearColorValue"},
				{Name: "depthStencil", Type: "VkClearDepthStencilValue"},
			}},
			{Name: "VkSubresourceLayout", GoName: "SubresourceLayout", Category: "struct"},
		},
		Constants: []model.SelectedConstant{
			{Name: "VK_SUCCESS", Value: "0", Extends: "VkResult"},
			{Name: "VK_ERROR_OUT_OF_HOST_MEMORY", Value: "-1", Extends: "VkResult"},
			{Name: "VK_KHR_EXTERNAL_MEMORY_FD_EXTENSION_NAME", Value: "\"VK_KHR_external_memory_fd\""},
			{Name: "VK_QUEUE_FAMILY_EXTERNAL_KHR", Value: "(~1U)"},
			{Name: "VK_WHOLE_SIZE", Value: "(~0ULL)"},
			{Name: "VK_KHR_MAINTENANCE_1_SPEC_VERSION", Value: "2"},
			{Name: "VK_KHR_MAINTENANCE1_SPEC_VERSION", Value: "2"},
		},
		Commands: []model.SelectedCommand{
			{Name: "vkCreateInstance", GoName: "CreateInstance", Return: "VkResult", Dispatch: model.DispatchGlobal, Params: []model.ParamDecl{
				{Name: "pCreateInfo", Type: "VkInstanceCreateInfo", Const: true, PointerDepth: 1},
				{Name: "pAllocator", Type: "VkAllocationCallbacks", Const: true, PointerDepth: 1},
				{Name: "pInstance", Type: "VkInstance", PointerDepth: 1},
			}},
			{Name: "vkDestroyInstance", GoName: "DestroyInstance", Return: "void", Dispatch: model.DispatchInstance, Params: []model.ParamDecl{
				{Name: "instance", Type: "VkInstance"},
				{Name: "pAllocator", Type: "VkAllocationCallbacks", Const: true, PointerDepth: 1},
			}},
			{Name: "vkGetPhysicalDeviceProperties2", GoName: "GetPhysicalDeviceProperties2", Return: "void", Dispatch: model.DispatchInstance, Params: []model.ParamDecl{
				{Name: "physicalDevice", Type: "VkPhysicalDevice"},
				{Name: "pProperties", Type: "VkPhysicalDeviceProperties2", PointerDepth: 1},
			}},
			{Name: "vkGetPhysicalDeviceProperties2KHR", GoName: "GetPhysicalDeviceProperties2KHR", Return: "void", Dispatch: model.DispatchInstance, Optional: true, Source: model.CommandDecl{Alias: "vkGetPhysicalDeviceProperties2"}, Params: []model.ParamDecl{
				{Name: "physicalDevice", Type: "VkPhysicalDevice"},
				{Name: "pProperties", Type: "VkPhysicalDeviceProperties2", PointerDepth: 1},
			}},
			{Name: "vkGetMemoryFdKHR", GoName: "GetMemoryFdKHR", Return: "VkResult", Dispatch: model.DispatchDevice, Optional: true, Params: []model.ParamDecl{
				{Name: "device", Type: "VkDevice"},
				{Name: "pGetFdInfo", Type: "VkMemoryGetFdInfoKHR", Const: true, PointerDepth: 1},
				{Name: "pFd", Type: "int", PointerDepth: 1},
			}},
			{Name: "vkMapMemory", GoName: "MapMemory", Return: "VkResult", Dispatch: model.DispatchDevice, Params: []model.ParamDecl{
				{Name: "device", Type: "VkDevice"},
				{Name: "memory", Type: "VkDeviceMemory"},
				{Name: "offset", Type: "VkDeviceSize"},
				{Name: "size", Type: "VkDeviceSize"},
				{Name: "flags", Type: "VkMemoryMapFlags"},
				{Name: "ppData", Type: "void", PointerDepth: 2},
			}},
		},
	}
}

func testSelectedRegistryNoUnsafe() *model.SelectedRegistry {
	return &model.SelectedRegistry{
		Types: []model.SelectedType{
			{Name: "VkInstance", GoName: "Instance", GoType: "uintptr", Category: "handle", Dispatchable: true},
			{Name: "VkResult", GoName: "Result", GoType: "int32", Category: "basetype"},
		},
		Commands: []model.SelectedCommand{
			{Name: "vkCreateInstance", GoName: "CreateInstance", Return: "VkResult", Dispatch: model.DispatchGlobal, Params: []model.ParamDecl{
				{Name: "pInstance", Type: "VkInstance", PointerDepth: 1},
			}},
		},
	}
}
