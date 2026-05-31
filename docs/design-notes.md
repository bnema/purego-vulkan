# Design notes

`purego-vulkan` is a Linux-first, no-cgo Vulkan binding for Go. It exposes a low-level raw surface generated from `registry/vk.xml` plus a small handwritten runtime layer for loader initialization, result helpers, and explicit dispatch loading.

## Loader and dispatch design

`vulkan.Init` opens `libvulkan.so.1` through `github.com/bnema/purego`, resolves `vkGetInstanceProcAddr` with `dlsym`, registers it, and loads global commands. Failed initialization is retryable, so callers can supply `vulkan.WithLibraryPath` after a default lookup failure.

Vulkan commands are loaded through Vulkan dispatch rather than as a flat `dlsym` table:

- global commands are available through `vulkan.Global()` after `vulkan.Init()`;
- instance commands are loaded with `vulkan.LoadInstanceDispatch(instance)`;
- device commands are loaded with `vulkan.LoadDeviceDispatch(instanceDispatch, device)`.

Required commands return descriptive errors when missing. Optional extension commands remain nil when the loader or driver does not expose them. The dispatch loaders serialize and clear their package-level purego registration staging before each load so optional symbols cannot leak between devices or instances.

## Generated subset

The generator reads the pinned Khronos registry at `registry/vk.xml` and emits the selected Linux compositor subset:

- core instance and device enumeration;
- command buffer, fence, semaphore, image, memory, descriptor, sampler, shader module, and pipeline-layout setup;
- external memory fd / DMA-BUF support;
- external semaphore fd support;
- DRM format modifier and physical-device DRM property structures;
- synchronization2 command and structure support where present.

The selected extension set includes:

- `VK_KHR_get_physical_device_properties2`
- `VK_KHR_external_memory`
- `VK_KHR_external_memory_fd`
- `VK_EXT_external_memory_dma_buf`
- `VK_EXT_image_drm_format_modifier`
- `VK_EXT_physical_device_drm`
- `VK_KHR_external_semaphore`
- `VK_KHR_external_semaphore_fd`
- `VK_KHR_synchronization2`
- `VK_EXT_queue_family_foreign`

The generator resolves aliases, feature constants, extension offset constants, guarded extension requirements, and core/KHR promoted command fallbacks for the selected subset. Generated files are committed and checked for freshness by `go generate ./...` plus `git diff --exit-code`.

## v0.x exclusions

This package intentionally avoids high-level Vulkan ownership abstractions. It does not provide swapchain management, render-pass builders, descriptor allocators, pipeline caches, scene graphs, callback-heavy debug messenger surfaces, or compositor-specific import policy. Those belong in consumer packages that can make product-specific trade-offs.

## Consumer integration checklist

- Use a local `replace github.com/bnema/purego-vulkan => ../purego-vulkan` during bring-up.
- Keep Vulkan binding code inside `purego-vulkan`.
- Keep compositor-specific rendering, import policy, and DRM presentation inside `go-wm-poc/adapters/vulkan`.
- Do not copy generated files into the compositor repository.
- Run `make check` in this repository before updating the compositor integration.

The `examples/enumerate` program is the smoke test for consumers: it initializes Vulkan, prints loader version, reports compositor-critical extension availability, creates an instance, and lists physical devices.
