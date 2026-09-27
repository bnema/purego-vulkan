package integration_test

import (
	"github.com/bnema/purego-vulkan/vulkan"
	"testing"
)

func commandRecordingFixture(tb testing.TB) (*vulkan.DeviceDispatch, vulkan.CommandBuffer, func()) {
	tb.Helper()
	if err := vulkan.Init(); err != nil {
		tb.Skipf("Vulkan loader unavailable: %v", err)
	}
	gd := vulkan.Global()
	app := &vulkan.ApplicationInfo{SType: vulkan.StructureTypeApplicationInfo, ApiVersion: vulkan.MakeVersion(1, 3, 0)}
	create := &vulkan.InstanceCreateInfo{SType: vulkan.StructureTypeInstanceCreateInfo, ApplicationInfo: app}
	var instance vulkan.Instance
	if result := gd.CreateInstance(create, nil, &instance); result != vulkan.Success {
		tb.Skipf("Vulkan instance unavailable: %v", result)
	}
	id, err := vulkan.LoadInstanceDispatch(instance)
	if err != nil {
		vulkan.VkDestroyInstance(instance, nil)
		tb.Fatalf("instance dispatch: %v", err)
	}
	cleanupInstance := func() { id.DestroyInstance(instance, nil) }
	var count uint32
	if result := id.EnumeratePhysicalDevices(instance, &count, nil); result != vulkan.Success || count == 0 {
		cleanupInstance()
		tb.Skipf("no physical devices: %v", result)
	}
	physical := make([]vulkan.PhysicalDevice, count)
	if result := id.EnumeratePhysicalDevices(instance, &count, &physical[0]); result != vulkan.Success {
		cleanupInstance()
		tb.Fatalf("physical devices: %v", result)
	}
	var family uint32
	id.GetPhysicalDeviceQueueFamilyProperties(physical[0], &family, nil)
	families := make([]vulkan.QueueFamilyProperties, family)
	if family == 0 {
		cleanupInstance()
		tb.Skip("no queue families")
	}
	id.GetPhysicalDeviceQueueFamilyProperties(physical[0], &family, &families[0])
	queueIndex := uint32(0)
	for i, f := range families {
		if f.QueueFlags&vulkan.QueueGraphicsBit != 0 {
			queueIndex = uint32(i)
			break
		}
	}
	priority := float32(1)
	queue := &vulkan.DeviceQueueCreateInfo{SType: vulkan.StructureTypeDeviceQueueCreateInfo, QueueFamilyIndex: queueIndex, QueueCount: 1, QueuePriorities: &priority}
	info := &vulkan.DeviceCreateInfo{SType: vulkan.StructureTypeDeviceCreateInfo, QueueCreateInfoCount: 1, QueueCreateInfos: queue}
	var device vulkan.Device
	if result := id.CreateDevice(physical[0], info, nil, &device); result != vulkan.Success {
		cleanupInstance()
		tb.Skipf("device unavailable: %v", result)
	}
	dd, err := vulkan.LoadDeviceDispatch(id, device)
	if err != nil {
		cleanupInstance()
		tb.Skipf("device dispatch unavailable: %v", err)
	}
	poolInfo := &vulkan.CommandPoolCreateInfo{SType: vulkan.StructureTypeCommandPoolCreateInfo, QueueFamilyIndex: queueIndex}
	var pool vulkan.CommandPool
	if result := dd.CreateCommandPool(device, poolInfo, nil, &pool); result != vulkan.Success {
		dd.DestroyDevice(device, nil)
		cleanupInstance()
		tb.Fatalf("command pool: %v", result)
	}
	alloc := &vulkan.CommandBufferAllocateInfo{SType: vulkan.StructureTypeCommandBufferAllocateInfo, CommandPool: pool, Level: vulkan.CommandBufferLevelPrimary, CommandBufferCount: 1}
	var cb vulkan.CommandBuffer
	if result := dd.AllocateCommandBuffers(device, alloc, &cb); result != vulkan.Success {
		dd.DestroyCommandPool(device, pool, nil)
		dd.DestroyDevice(device, nil)
		cleanupInstance()
		tb.Fatalf("command buffer: %v", result)
	}
	return dd, cb, func() { dd.DestroyCommandPool(device, pool, nil); dd.DestroyDevice(device, nil); cleanupInstance() }
}

func recordCommands(dd *vulkan.DeviceDispatch, cb vulkan.CommandBuffer, begin *vulkan.CommandBufferBeginInfo) {
	if dd.ResetCommandBuffer(cb, 0) != vulkan.Success {
		panic("reset command buffer")
	}
	if dd.BeginCommandBuffer(cb, begin) != vulkan.Success {
		panic("begin command buffer")
	}
	dd.CmdPipelineBarrier(cb, 0, 0, 0, 0, nil, 0, nil, 0, nil)
	if dd.EndCommandBuffer(cb) != vulkan.Success {
		panic("end command buffer")
	}
}

func TestCommandRecordingNoAllocs(t *testing.T) {
	dd, cb, cleanup := commandRecordingFixture(t)
	defer cleanup()
	begin := &vulkan.CommandBufferBeginInfo{SType: vulkan.StructureTypeCommandBufferBeginInfo}
	if got := testing.AllocsPerRun(100, func() { recordCommands(dd, cb, begin) }); got != 0 {
		t.Errorf("recording allocs = %v, want 0", got)
	}
}

func BenchmarkCommandRecording(b *testing.B) {
	dd, cb, cleanup := commandRecordingFixture(b)
	defer cleanup()
	begin := &vulkan.CommandBufferBeginInfo{SType: vulkan.StructureTypeCommandBufferBeginInfo}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recordCommands(dd, cb, begin)
	}
}
