package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"

	"github.com/bnema/purego-vulkan/vulkan"
)

var criticalExtensions = []string{
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

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if err := vulkan.Init(); err != nil {
		return err
	}
	gd := vulkan.Global()
	if gd.EnumerateInstanceVersion != nil {
		var version uint32
		if err := vulkan.Check(gd.EnumerateInstanceVersion(&version)); err != nil {
			return err
		}
		fmt.Printf("Vulkan loader version: %s\n", vulkan.FormatVersion(version))
	} else {
		fmt.Println("Vulkan loader version: <vkEnumerateInstanceVersion unavailable>")
	}

	instance, id, cleanup, err := createInstance(gd)
	if err != nil {
		return err
	}
	defer cleanup()

	devices, err := enumerateDevices(id, instance)
	if err != nil {
		return err
	}

	report, err := criticalExtensionReport(gd, id, firstDevice(devices))
	if err != nil {
		return err
	}
	fmt.Println("Critical extensions:")
	for _, name := range criticalExtensions {
		fmt.Printf("  %-45s %s\n", name, yesNo(report[name]))
	}

	fmt.Println("Physical devices:")
	if len(devices) == 0 {
		fmt.Println("  <none>")
		return nil
	}
	for _, device := range devices {
		fmt.Printf("  %s\n", deviceName(id, device))
	}
	return nil
}

func createInstance(gd *vulkan.GlobalDispatch) (vulkan.Instance, *vulkan.InstanceDispatch, func(), error) {
	appName := cStringBytes("purego-vulkan enumerate")
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
	if err := vulkan.Check(result); err != nil {
		return 0, nil, nil, err
	}
	id, err := vulkan.LoadInstanceDispatch(instance)
	if err != nil {
		if vulkan.VkDestroyInstance != nil {
			vulkan.VkDestroyInstance(instance, nil)
		}
		return 0, nil, nil, err
	}
	cleanup := func() { id.DestroyInstance(instance, nil) }
	return instance, id, cleanup, nil
}

func enumerateDevices(id *vulkan.InstanceDispatch, instance vulkan.Instance) ([]vulkan.PhysicalDevice, error) {
	var count uint32
	if err := vulkan.Check(id.EnumeratePhysicalDevices(instance, &count, nil)); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	devices := make([]vulkan.PhysicalDevice, count)
	if err := vulkan.Check(id.EnumeratePhysicalDevices(instance, &count, &devices[0])); err != nil {
		return nil, err
	}
	return devices[:count], nil
}

func criticalExtensionReport(gd *vulkan.GlobalDispatch, id *vulkan.InstanceDispatch, physicalDevice vulkan.PhysicalDevice) (map[string]bool, error) {
	report := make(map[string]bool, len(criticalExtensions))
	for _, name := range criticalExtensions {
		report[name] = false
	}
	instanceExtensions, err := enumerateInstanceExtensions(gd)
	if err != nil {
		return nil, err
	}
	for _, name := range instanceExtensions {
		if _, ok := report[name]; ok {
			report[name] = true
		}
	}
	if physicalDevice == 0 {
		return report, nil
	}
	deviceExtensions, err := enumerateDeviceExtensions(id, physicalDevice)
	if err != nil {
		return nil, err
	}
	for _, name := range deviceExtensions {
		if _, ok := report[name]; ok {
			report[name] = true
		}
	}
	return report, nil
}

func enumerateInstanceExtensions(gd *vulkan.GlobalDispatch) ([]string, error) {
	var count uint32
	if err := vulkan.Check(gd.EnumerateInstanceExtensionProperties(nil, &count, nil)); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	props := make([]vulkan.ExtensionProperties, count)
	if err := vulkan.Check(gd.EnumerateInstanceExtensionProperties(nil, &count, &props[0])); err != nil {
		return nil, err
	}
	return extensionNames(props[:count]), nil
}

func enumerateDeviceExtensions(id *vulkan.InstanceDispatch, physicalDevice vulkan.PhysicalDevice) ([]string, error) {
	var count uint32
	if err := vulkan.Check(id.EnumerateDeviceExtensionProperties(physicalDevice, nil, &count, nil)); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	props := make([]vulkan.ExtensionProperties, count)
	if err := vulkan.Check(id.EnumerateDeviceExtensionProperties(physicalDevice, nil, &count, &props[0])); err != nil {
		return nil, err
	}
	return extensionNames(props[:count]), nil
}

func extensionNames(props []vulkan.ExtensionProperties) []string {
	names := make([]string, 0, len(props))
	for _, prop := range props {
		names = append(names, fixedCString(prop.ExtensionName[:]))
	}
	return names
}

func firstDevice(devices []vulkan.PhysicalDevice) vulkan.PhysicalDevice {
	if len(devices) == 0 {
		return 0
	}
	return devices[0]
}

func deviceName(id *vulkan.InstanceDispatch, device vulkan.PhysicalDevice) string {
	var props vulkan.PhysicalDeviceProperties
	id.GetPhysicalDeviceProperties(device, &props)
	return fixedCString(props.DeviceName[:])
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

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
