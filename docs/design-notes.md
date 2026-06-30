# Design notes

`purego-vulkan` is a Linux-first, no-cgo Vulkan binding for Go. It exposes a low-level raw surface generated from `registry/vk.xml` plus a small handwritten runtime layer for loader initialization, result helpers, and explicit dispatch loading.

## Loader and dispatch design

`vulkan.Init` opens `libvulkan.so.1` through `github.com/bnema/purego`, resolves `vkGetInstanceProcAddr` with `dlsym`, registers it, and loads global commands. Failed initialization is retryable, so callers can supply `vulkan.WithLibraryPath` after a default lookup failure.

Vulkan commands are loaded through Vulkan dispatch rather than as a flat `dlsym` table:

- global commands are available through `vulkan.Global()` after `vulkan.Init()`;
- instance commands are loaded with `vulkan.LoadInstanceDispatch(instance)`;
- device commands are loaded with `vulkan.LoadDeviceDispatch(instanceDispatch, device)`.

Required commands return descriptive errors when missing. Optional extension commands remain nil when the loader or driver does not expose them. The dispatch loaders serialize and clear their package-level purego registration staging before each load so optional symbols cannot leak between devices or instances.

## Generated profiles

The generator reads the pinned Khronos registry at `registry/vk.xml` and emits one of three profiles:

- `renderer`: the original low-level compositor renderer subset;
- `wsi`: the default committed profile, adding Linux WSI/swapchain commands and image readback;
- `complete`: a broad pinned-registry profile for audit and consumers that need commands outside the default package.

See `docs/binding-coverage.md` for current command/type/constant/extension counts.

The default `wsi` profile includes:

- core instance and device enumeration;
- CPU-visible buffer upload primitives (`vkMapMemory`, `vkUnmapMemory`, memory binding, buffers, and barriers);
- buffer/image transfer commands (`vkCmdCopyBufferToImage`, `vkCmdCopyImageToBuffer`, and image layout/access structures);
- required memory-property queries for allocation decisions (`vkGetPhysicalDeviceMemoryProperties` and `vkGetPhysicalDeviceMemoryProperties2`);
- required classic `vkCmdPipelineBarrier` support as the Vulkan 1.0 synchronization fallback, with synchronization2 extension commands exposed when present;
- command buffer, fence, semaphore, image, memory, descriptor, sampler, shader module, pipeline-layout, graphics-pipeline creation, and pipeline destruction setup;
- draw command recording (`vkCmdBindPipeline`, descriptor/vertex/index binding, `vkCmdDraw`, and `vkCmdDrawIndexed`);
- dynamic rendering commands and structures for render-target setup without generated render-pass builders;
- external memory fd / DMA-BUF support;
- external semaphore fd support;
- DRM format modifier and physical-device DRM property structures;
- synchronization2 command and structure support where present;
- Linux WSI and presentation bindings for `VK_KHR_surface`, `VK_KHR_swapchain`, `VK_KHR_wayland_surface`, `VK_KHR_xcb_surface`, and `VK_KHR_xlib_surface`.

The selected extension set includes:

- `VK_KHR_surface`
- `VK_KHR_swapchain`
- `VK_KHR_wayland_surface`
- `VK_KHR_xcb_surface`
- `VK_KHR_xlib_surface`
- `VK_KHR_dynamic_rendering`
- `VK_KHR_get_physical_device_properties2`
- `VK_KHR_bind_memory2`
- `VK_KHR_get_memory_requirements2`
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

## Renderer surface and render-target choice

The selected renderer surface is deliberately close to Vulkan: consumers receive generated handle types, create-info structs, constants, result values, and dispatch-table function fields. That is enough for a compositor renderer to allocate host-visible staging buffers, copy pixels into images, create and destroy graphics pipelines, bind descriptors/buffers/pipelines, issue draw calls, and begin/end dynamic rendering.

Render targets should use dynamic rendering (`vkCmdBeginRendering` / `vkCmdEndRendering`, or the `KHR` aliases when that is the exposed path) instead of requiring classic render-pass/framebuffer objects in this binding milestone. Dynamic rendering fits imported images and compositor-owned render-target decisions better: the consumer can choose color/depth attachments at command-record time, while this package only exposes the raw structures and commands. `LoadDeviceDispatch` requires either the core dynamic-rendering symbols or their `KHR` aliases and returns an error when neither path is exposed, so consumers should filter devices by Vulkan version/extension support before creating renderer devices.

## v0.x exclusions

This package intentionally avoids high-level Vulkan ownership abstractions. It provides raw surface/swapchain bindings but does not provide swapchain management, render-pass builders, framebuffer builders, descriptor allocators, pipeline-cache policy, scene graphs, callback-heavy debug messenger surfaces, upload managers, render-graph abstractions, native window creation, or compositor-specific import policy. Those belong in consumer packages that can make product-specific trade-offs.

Classic render-pass/framebuffer commands and pipeline-cache commands are not part of the renderer-ready target unless already needed as parameter types (for example, `PipelineCache` as a nullable handle passed to `vkCreateGraphicsPipelines`). Prefer `VK_NULL_HANDLE` / zero pipeline cache and dynamic rendering for the current compositor path.

## Consumer integration checklist

- Use a local `replace github.com/bnema/purego-vulkan => ../purego-vulkan` during bring-up.
- Keep Vulkan binding code inside `purego-vulkan`.
- Keep compositor-specific rendering, import policy, and DRM presentation inside `go-wm-poc/adapters/vulkan`.
- Do not copy generated files into the compositor repository.
- Upload handoff: consumer chooses memory types, staging-buffer lifetime, flush/invalidate policy, and barriers; this package supplies buffer, memory-map, bind, copy, and synchronization command bindings.
- Pipeline handoff: consumer owns shader modules, pipeline layouts, descriptor layouts, graphics-pipeline create-info assembly, and destruction ordering; this package supplies the raw create/destroy commands and structs.
- Draw handoff: consumer records command buffers, binds pipelines/descriptors/vertex/index buffers, and issues draw calls through the device dispatch table.
- Render-target handoff: consumer imports or creates images, selects layouts/load-store ops/attachment formats, enables core/KHR dynamic rendering, and treats `LoadDeviceDispatch` failure as a device-capability rejection when dynamic-rendering symbols are unavailable.
- Run `make check` in this repository before updating the compositor integration.

The `examples/enumerate` program is the smoke test for consumers: it initializes Vulkan, prints loader version, reports compositor-critical extension availability, creates an instance, and lists physical devices.
