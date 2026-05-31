# purego-vulkan

Linux-first Vulkan bindings for Go without cgo.

`purego-vulkan` loads the system Vulkan loader (`libvulkan.so.1`) through `github.com/bnema/purego` and generates its raw Vulkan API surface from the pinned Khronos registry at `registry/vk.xml`.

The package is intentionally low-level. It exposes Vulkan ABI types, constants, command signatures, result helpers, and explicit global/instance/device dispatch tables for compositor and renderer code that owns its rendering architecture.

## Status

Early v0.x implementation. The initial target is a Linux/DRM compositor renderer path. Other platforms, high-level rendering helpers, swapchain management, and scene abstractions are out of scope for this milestone.

## Usage

```go
if err := vulkan.Init(); err != nil {
    return err
}

gd := vulkan.Global()
var version uint32
if gd.EnumerateInstanceVersion != nil {
    if err := vulkan.Check(gd.EnumerateInstanceVersion(&version)); err != nil {
        return err
    }
    fmt.Println(vulkan.FormatVersion(version))
}
```

Use `LoadInstanceDispatch(instance)` and `LoadDeviceDispatch(instanceDispatch, device)` after creating Vulkan handles. Optional extension commands are nil when unavailable.

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

## Generated files

Generated files are committed. To update them, modify `cmd/vulkangen`, `registry/vk.xml`, or overrides, then run:

```sh
go generate ./...
go test ./...
git diff --exit-code
```

## Consumer integration

During compositor bring-up, use a local module replacement such as:

```go
replace github.com/bnema/purego-vulkan => ../purego-vulkan
```

Keep generated Vulkan binding code in this repository. Compositor-specific rendering, import policy, and DRM presentation belong in `go-wm-poc/adapters/vulkan`; do not copy generated files into the compositor repository.

See `docs/design-notes.md` for dispatch design, selected extensions, exclusions, and the consumer handoff checklist.
