# purego-vulkan

Linux-first Vulkan bindings for Go without cgo.

`purego-vulkan` loads the system Vulkan loader (`libvulkan.so.1`) through `github.com/bnema/purego` and generates its raw Vulkan API surface from the pinned Khronos registry at `registry/vk.xml`.

The package is intentionally low-level. It exposes Vulkan ABI types, constants, command signatures, result helpers, and explicit global/instance/device dispatch tables for compositor and renderer code that owns its rendering architecture.

## Status

Early v0.x implementation. The committed generated package uses the `wsi` coverage profile: CPU-visible buffer uploads, memory-property queries, buffer/image copies, pipeline barriers, graphics pipeline creation/destruction, draw commands, dynamic rendering, Linux surface/swapchain bindings, and image readback commands are covered. Other platforms, high-level rendering helpers, swapchain management policy, and scene abstractions are out of scope for this milestone.

## Usage

```go
if err := vulkan.Init(); err != nil {
    return err
}

gd := vulkan.Global()
var version uint32
if gd.HasEnumerateInstanceVersion() {
    if err := vulkan.Check(gd.EnumerateInstanceVersion(&version)); err != nil {
        return err
    }
    fmt.Println(vulkan.FormatVersion(version))
}
```

Use `LoadInstanceDispatch(instance)` and `LoadDeviceDispatch(instanceDispatch, device)` after creating Vulkan handles. Dispatch commands are pointer-receiver methods; use `HasX()` (for example, `dd.HasGetMemoryFdKHR()`) before invoking optional commands. Unavailable commands return false from `HasX()`. Package-level `VkXxx` function variables remain available for compatibility.

Run the smoke example on a Vulkan-capable Linux machine:

```sh
go run ./examples/enumerate
```

The example prints the loader version, compositor-critical extension availability, and physical device names.

## Development

Stable commands:

```sh
make generate
make test
make check
```

`make check` runs generation, tests, and `git diff --exit-code` to ensure generated files are fresh.

The development target is raw Vulkan binding coverage, not framework code. Keep upload orchestration, swapchain ownership, pipeline policy, render-target ownership, and command-buffer recording in the consumer; this module supplies the low-level Vulkan commands and structs needed for CPU-visible staging/upload buffers, buffer/image copies, graphics pipelines, draw calls, dynamic rendering, and Linux WSI.

Coverage profiles:

```sh
go run ./cmd/vulkangen --profile renderer  # renderer-only subset
go run ./cmd/vulkangen --profile wsi       # default committed profile
go run ./cmd/vulkangen --profile complete  # broad pinned-registry audit profile
```

See `docs/binding-coverage.md` for current registry coverage counts and intentional exclusions.

## Generated files

Generated files are committed. To update them, modify `cmd/vulkangen`, `registry/vk.xml`, or overrides, then run:

```sh
go generate ./...
go test ./...
git diff --exit-code
```

## Consumer integration

```sh
go get github.com/bnema/purego-vulkan@latest
```

To work on the bindings and a consumer together, use a Go workspace (`go work init . ../purego-vulkan`) instead of a `replace` directive. Rendering policy, image import and presentation belong in the consumer; do not copy generated files into it.

See `docs/design-notes.md` for dispatch design, selected extensions, exclusions, and the consumer handoff checklist.
