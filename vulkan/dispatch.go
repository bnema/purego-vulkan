package vulkan

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/bnema/purego-vulkan/internal/capi"
)

type Error struct {
	Result Result
}

func (e Error) Error() string {
	return fmt.Sprintf("vulkan: %s", ResultString(e.Result))
}

func Check(r Result) error {
	if r == Success {
		return nil
	}
	return Error{Result: r}
}

var (
	dispatchMu     sync.Mutex
	globalDispatch GlobalDispatch
)

func Global() *GlobalDispatch {
	return &globalDispatch
}

func loadGlobalDispatch() error {
	if vkGetInstanceProcAddr == nil && VkGetInstanceProcAddr == nil {
		return fmt.Errorf("vulkan: vkGetInstanceProcAddr is not loaded")
	}
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	globalDispatch = GlobalDispatch{}
	if err := capi.RegisterGlobal(0, lookupInstanceProcAddr, globalDispatch.commandPointers()); err != nil {
		return fmt.Errorf("vulkan: load global dispatch: %w", err)
	}
	fptrs := globalCommandPointers()
	clearCommandPointers(fptrs)
	if err := capi.RegisterGlobal(0, lookupInstanceProcAddr, fptrs); err != nil {
		return err
	}
	return nil
}

func LoadInstanceDispatch(instance Instance) (*InstanceDispatch, error) {
	if vkGetInstanceProcAddr == nil && VkGetInstanceProcAddr == nil {
		return nil, fmt.Errorf("vulkan: vkGetInstanceProcAddr is not loaded")
	}
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	dispatch := &InstanceDispatch{Instance: instance}
	if err := capi.RegisterInstance(uintptr(instance), lookupInstanceProcAddr, dispatch.commandPointers()); err != nil {
		return nil, fmt.Errorf("vulkan: load instance dispatch: %w", err)
	}
	fptrs := instanceCommandPointers()
	clearCommandPointers(fptrs)
	if err := capi.RegisterInstance(uintptr(instance), lookupInstanceProcAddr, fptrs); err != nil {
		return nil, err
	}
	return dispatch, nil
}

func LoadDeviceDispatch(instance *InstanceDispatch, device Device) (*DeviceDispatch, error) {
	if instance == nil {
		return nil, fmt.Errorf("vulkan: instance dispatch is nil")
	}
	if !instance.HasGetDeviceProcAddr() {
		return nil, fmt.Errorf("vulkan: vkGetDeviceProcAddr is not loaded")
	}
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	lookup := func(handle uintptr, name string) (uintptr, error) {
		return lookupDeviceProcAddr(instance.GetDeviceProcAddr, Device(handle), name), nil
	}
	dispatch := &DeviceDispatch{Device: device}
	if err := capi.RegisterDevice(uintptr(device), lookup, dispatch.commandPointers()); err != nil {
		return nil, fmt.Errorf("vulkan: load device dispatch: %w", err)
	}
	fptrs := deviceCommandPointers()
	clearCommandPointers(fptrs)
	if err := capi.RegisterDevice(uintptr(device), lookup, fptrs); err != nil {
		return nil, err
	}
	return dispatch, nil
}

func lookupInstanceProcAddr(handle uintptr, name string) (uintptr, error) {
	nameBytes := cStringBytes(name)
	if vkGetInstanceProcAddr != nil {
		addr := vkGetInstanceProcAddr(handle, &nameBytes[0])
		runtime.KeepAlive(nameBytes)
		return addr, nil
	}
	if VkGetInstanceProcAddr != nil {
		addr := VkGetInstanceProcAddr(Instance(handle), &nameBytes[0])
		runtime.KeepAlive(nameBytes)
		return uintptr(addr), nil
	}
	runtime.KeepAlive(nameBytes)
	return 0, fmt.Errorf("vkGetInstanceProcAddr is not loaded")
}

func lookupDeviceProcAddr(getProcAddr func(Device, *byte) PFN_vkVoidFunction, device Device, name string) uintptr {
	nameBytes := cStringBytes(name)
	addr := getProcAddr(device, &nameBytes[0])
	runtime.KeepAlive(nameBytes)
	return uintptr(addr)
}

func cStringBytes(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

func clearCommandPointers(fptrs map[string]any) {
	for _, fptr := range fptrs {
		v := reflect.ValueOf(fptr).Elem()
		v.Set(reflect.Zero(v.Type()))
	}
}
