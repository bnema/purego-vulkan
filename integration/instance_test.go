package integration_test

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/bnema/purego-vulkan/vulkan"
)

var compositorCriticalExtensions = []string{
	"VK_KHR_get_physical_device_properties2",
	"VK_KHR_external_memory",
	"VK_KHR_external_memory_fd",
	"VK_EXT_external_memory_dma_buf",
	"VK_EXT_image_drm_format_modifier",
	"VK_EXT_physical_device_drm",
	"VK_KHR_external_semaphore",
	"VK_KHR_external_semaphore_fd",
	"VK_KHR_synchronization2",
	"VK_EXT_queue_family_foreign",
}

func TestCreateMinimalInstance(t *testing.T) {
	gd := requireGlobalDispatch(t)
	_, _, cleanup := createTestInstance(t, gd)
	cleanup()
}

func TestEnumeratePhysicalDevices(t *testing.T) {
	gd := requireGlobalDispatch(t)
	instance, id, cleanup := createTestInstance(t, gd)
	defer cleanup()

	devices := enumeratePhysicalDevices(t, id, instance)
	if len(devices) == 0 {
		t.Skip("Vulkan loader has no physical devices")
	}
}

func TestEnumerateInstanceVersion(t *testing.T) {
	gd := requireGlobalDispatch(t)
	if gd.EnumerateInstanceVersion == nil {
		t.Skip("vkEnumerateInstanceVersion unavailable")
	}
	var version uint32
	if err := vulkan.Check(gd.EnumerateInstanceVersion(&version)); err != nil {
		t.Fatalf("EnumerateInstanceVersion: %v", err)
	}
	if version == 0 {
		t.Fatal("EnumerateInstanceVersion returned 0")
	}
}

func TestReportCriticalExtensions(t *testing.T) {
	gd := requireGlobalDispatch(t)
	instance, id, cleanup := createTestInstance(t, gd)
	defer cleanup()

	var physicalDevice vulkan.PhysicalDevice
	if devices := enumeratePhysicalDevices(t, id, instance); len(devices) > 0 {
		physicalDevice = devices[0]
	}
	report := criticalExtensionReport(t, gd, id, physicalDevice)
	for _, name := range compositorCriticalExtensions {
		if _, ok := report[name]; !ok {
			t.Fatalf("critical extension %s missing from report", name)
		}
	}
}

func requireGlobalDispatch(t *testing.T) *vulkan.GlobalDispatch {
	t.Helper()
	if err := vulkan.Init(); err != nil {
		t.Skipf("Vulkan loader unavailable: %v", err)
	}
	gd := vulkan.Global()
	if gd == nil || gd.CreateInstance == nil {
		t.Fatal("global dispatch was not initialized")
	}
	return gd
}

func createTestInstance(t *testing.T, gd *vulkan.GlobalDispatch) (vulkan.Instance, *vulkan.InstanceDispatch, func()) {
	t.Helper()
	appName := cStringBytes("purego-vulkan integration test")
	engineName := cStringBytes("purego-vulkan")
	app := vulkan.ApplicationInfo{
		SType:           vulkan.StructureTypeApplicationInfo,
		ApplicationName: &appName[0],
		EngineName:      &engineName[0],
	}
	info := vulkan.InstanceCreateInfo{
		SType:           vulkan.StructureTypeInstanceCreateInfo,
		ApplicationInfo: &app,
	}
	var instance vulkan.Instance
	result := gd.CreateInstance(&info, nil, &instance)
	runtime.KeepAlive(appName)
	runtime.KeepAlive(engineName)
	if result != vulkan.Success {
		t.Skipf("Vulkan instance unavailable: %s", vulkan.ResultString(result))
	}
	id, err := vulkan.LoadInstanceDispatch(instance)
	if err != nil {
		if vulkan.VkDestroyInstance != nil {
			vulkan.VkDestroyInstance(instance, nil)
		}
		t.Fatalf("LoadInstanceDispatch: %v", err)
	}
	cleanup := func() {
		id.DestroyInstance(instance, nil)
	}
	return instance, id, cleanup
}

func enumeratePhysicalDevices(t *testing.T, id *vulkan.InstanceDispatch, instance vulkan.Instance) []vulkan.PhysicalDevice {
	t.Helper()
	var count uint32
	if err := vulkan.Check(id.EnumeratePhysicalDevices(instance, &count, nil)); err != nil {
		t.Fatalf("EnumeratePhysicalDevices count: %v", err)
	}
	if count == 0 {
		return nil
	}
	devices := make([]vulkan.PhysicalDevice, count)
	if err := vulkan.Check(id.EnumeratePhysicalDevices(instance, &count, &devices[0])); err != nil {
		t.Fatalf("EnumeratePhysicalDevices values: %v", err)
	}
	return devices[:count]
}

func criticalExtensionReport(t *testing.T, gd *vulkan.GlobalDispatch, id *vulkan.InstanceDispatch, physicalDevice vulkan.PhysicalDevice) map[string]bool {
	t.Helper()
	report := make(map[string]bool, len(compositorCriticalExtensions))
	for _, name := range compositorCriticalExtensions {
		report[name] = false
	}
	for _, name := range enumerateInstanceExtensions(t, gd) {
		if _, ok := report[name]; ok {
			report[name] = true
		}
	}
	if physicalDevice != 0 {
		for _, name := range enumerateDeviceExtensions(t, id, physicalDevice) {
			if _, ok := report[name]; ok {
				report[name] = true
			}
		}
	}
	return report
}

func enumerateInstanceExtensions(t *testing.T, gd *vulkan.GlobalDispatch) []string {
	t.Helper()
	var count uint32
	if err := vulkan.Check(gd.EnumerateInstanceExtensionProperties(nil, &count, nil)); err != nil {
		t.Fatalf("EnumerateInstanceExtensionProperties count: %v", err)
	}
	if count == 0 {
		return nil
	}
	props := make([]vulkan.ExtensionProperties, count)
	if err := vulkan.Check(gd.EnumerateInstanceExtensionProperties(nil, &count, &props[0])); err != nil {
		t.Fatalf("EnumerateInstanceExtensionProperties values: %v", err)
	}
	return extensionNames(props[:count])
}

func enumerateDeviceExtensions(t *testing.T, id *vulkan.InstanceDispatch, physicalDevice vulkan.PhysicalDevice) []string {
	t.Helper()
	var count uint32
	if err := vulkan.Check(id.EnumerateDeviceExtensionProperties(physicalDevice, nil, &count, nil)); err != nil {
		t.Fatalf("EnumerateDeviceExtensionProperties count: %v", err)
	}
	if count == 0 {
		return nil
	}
	props := make([]vulkan.ExtensionProperties, count)
	if err := vulkan.Check(id.EnumerateDeviceExtensionProperties(physicalDevice, nil, &count, &props[0])); err != nil {
		t.Fatalf("EnumerateDeviceExtensionProperties values: %v", err)
	}
	return extensionNames(props[:count])
}

func extensionNames(props []vulkan.ExtensionProperties) []string {
	names := make([]string, 0, len(props))
	for _, prop := range props {
		names = append(names, fixedCString(prop.ExtensionName[:]))
	}
	return names
}

func fixedCString(value []byte) string {
	if i := bytes.IndexByte(value, 0); i >= 0 {
		value = value[:i]
	}
	return string(value)
}

func cStringBytes(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}
