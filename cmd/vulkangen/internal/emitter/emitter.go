package emitter

import (
	"bytes"
	"fmt"
	"go/format"
	"strconv"
	"strings"

	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/model"
)

func EmitTypes(sel *model.SelectedRegistry) (string, error) {
	var b bytes.Buffer
	writeHeader(&b, "vulkan")
	b.WriteString("import \"unsafe\"\n\n")
	types := selectedTypeMap(sel)

	for _, t := range sel.Types {
		if t.Source.Alias != "" {
			fmt.Fprintf(&b, "type %s = %s\n\n", t.GoName, goTypeName(t.Source.Alias))
			continue
		}
		switch t.Category {
		case "handle", "basetype", "bitmask", "enum", "funcpointer":
			if t.GoType == "" {
				continue
			}
			fmt.Fprintf(&b, "type %s %s\n\n", t.GoName, t.GoType)
		case "struct":
			fmt.Fprintf(&b, "type %s struct {\n", t.GoName)
			for _, m := range t.Members {
				fmt.Fprintf(&b, "\t%s %s\n", exportName(m.Name), goFieldType(m))
			}
			b.WriteString("}\n\n")
		case "union":
			goType, err := unionGoType(t, types)
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&b, "type %s %s\n\n", t.GoName, goType)
		}
	}
	return formatSource(b.Bytes())
}

func selectedTypeMap(sel *model.SelectedRegistry) map[string]model.SelectedType {
	types := make(map[string]model.SelectedType, len(sel.Types))
	for _, t := range sel.Types {
		types[t.Name] = t
	}
	return types
}

func EmitConstants(sel *model.SelectedRegistry) (string, error) {
	var b bytes.Buffer
	writeHeader(&b, "vulkan")
	seen := make(map[string]string, len(sel.Constants))
	for _, c := range sel.Constants {
		goName := constName(c.Name)
		value := c.Value
		if value == "" && c.Source.Bitpos != "" {
			value = "1 << " + c.Source.Bitpos
		}
		value = goConstValue(value)
		if value == "" {
			continue
		}
		if previous, ok := seen[goName]; ok {
			if previous != value {
				return "", fmt.Errorf("constant Go name collision %s: %s=%s conflicts with %s", goName, previous, c.Name, value)
			}
			continue
		}
		seen[goName] = value
		if strings.HasPrefix(value, "\"") {
			fmt.Fprintf(&b, "const %s = %s\n", goName, value)
			continue
		}
		if c.Extends == "VkResult" || strings.Contains(c.Name, "_ERROR_") || strings.HasPrefix(c.Name, "VK_SUCCESS") {
			fmt.Fprintf(&b, "const %s Result = %s\n", goName, value)
		} else {
			fmt.Fprintf(&b, "const %s = %s\n", goName, value)
		}
	}
	return formatSource(b.Bytes())
}

func goConstValue(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "(~") && strings.HasSuffix(value, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(value, "(~"), ")")
		width := "uint32"
		if trimmed, ok := strings.CutSuffix(inner, "ULL"); ok {
			width = "uint64"
			inner = trimmed
		} else if trimmed, ok := strings.CutSuffix(inner, "UL"); ok {
			inner = trimmed
		} else if trimmed, ok := strings.CutSuffix(inner, "U"); ok {
			inner = trimmed
		}
		return "^" + width + "(" + inner + ")"
	}
	if trimmed, ok := strings.CutSuffix(value, "ULL"); ok {
		return trimmed
	}
	if trimmed, ok := strings.CutSuffix(value, "UL"); ok {
		return trimmed
	}
	if trimmed, ok := strings.CutSuffix(value, "U"); ok {
		return trimmed
	}
	if (strings.HasSuffix(value, "F") || strings.HasSuffix(value, "f")) && strings.ContainsAny(value, ".eE") {
		return value[:len(value)-1]
	}
	return value
}

func EmitCommands(sel *model.SelectedRegistry) (string, error) {
	var b bytes.Buffer
	writeHeader(&b, "vulkan")
	if commandsNeedUnsafe(sel.Commands) {
		b.WriteString("import \"unsafe\"\n\n")
	}
	for _, cmd := range sel.Commands {
		fmt.Fprintf(&b, "var %s func(%s)%s\n", rawFuncName(cmd.Name), paramsSignature(cmd.Params), resultSignature(cmd.Return))
	}
	b.WriteString("\n")
	writeCommandPointerMap(&b, "globalCommandPointers", commandsByDispatch(sel, model.DispatchGlobal))
	writeCommandPointerMap(&b, "instanceCommandPointers", commandsByDispatch(sel, model.DispatchInstance))
	writeCommandPointerMap(&b, "deviceCommandPointers", commandsByDispatch(sel, model.DispatchDevice))
	return formatSource(b.Bytes())
}

func EmitDispatch(sel *model.SelectedRegistry) (string, error) {
	var b bytes.Buffer
	writeHeader(&b, "vulkan")
	if commandsNeedUnsafe(sel.Commands) {
		b.WriteString("import \"unsafe\"\n\n")
	}
	writeDispatchStruct(&b, "GlobalDispatch", "", commandsByDispatch(sel, model.DispatchGlobal))
	writeDispatchStruct(&b, "InstanceDispatch", "Instance Instance", commandsByDispatch(sel, model.DispatchInstance))
	writeDispatchStruct(&b, "DeviceDispatch", "Device Device", commandsByDispatch(sel, model.DispatchDevice))
	return formatSource(b.Bytes())
}

func EmitStrings(sel *model.SelectedRegistry) (string, error) {
	var b bytes.Buffer
	writeHeader(&b, "vulkan")
	b.WriteString("func ResultString(r Result) string {\n\tswitch r {\n")
	for _, c := range sel.Constants {
		if c.Extends == "VkResult" || strings.Contains(c.Name, "_ERROR_") || strings.HasPrefix(c.Name, "VK_SUCCESS") {
			fmt.Fprintf(&b, "\tcase %s:\n\t\treturn %q\n", constName(c.Name), c.Name)
		}
	}
	b.WriteString("\tdefault:\n\t\treturn \"VK_UNKNOWN_RESULT\"\n\t}\n}\n")
	return formatSource(b.Bytes())
}

func EmitRegister(sel *model.SelectedRegistry) (string, error) {
	var b bytes.Buffer
	writeHeader(&b, "capi")
	b.WriteString("import (\n\t\"fmt\"\n\t\"strings\"\n)\n\n")
	b.WriteString("type LookupFunc func(handle uintptr, name string) (uintptr, error)\n\n")
	writeRegisterGroup(&b, "Global", commandsByDispatch(sel, model.DispatchGlobal))
	writeRegisterGroup(&b, "Instance", commandsByDispatch(sel, model.DispatchInstance))
	writeRegisterGroup(&b, "Device", commandsByDispatch(sel, model.DispatchDevice))
	b.WriteString(`func registerRequired(names []string, handle uintptr, lookup LookupFunc, fptrs map[string]any) error {
	if err := requireFunctionPointers(names, fptrs); err != nil {
		return err
	}
	if registerFirstAvailable(names, handle, lookup, fptrs) {
		return nil
	}
	return fmt.Errorf("resolve %s: no symbol found", strings.Join(names, " or "))
}

func registerOptional(names []string, handle uintptr, lookup LookupFunc, fptrs map[string]any) {
	registerFirstAvailable(names, handle, lookup, fptrs)
}

func requireFunctionPointers(names []string, fptrs map[string]any) error {
	for _, name := range names {
		fptr, ok := fptrs[name]
		if !ok || fptr == nil {
			return fmt.Errorf("missing function pointer for %s", name)
		}
	}
	return nil
}

func registerFirstAvailable(names []string, handle uintptr, lookup LookupFunc, fptrs map[string]any) bool {
	for _, name := range names {
		addr, err := lookup(handle, name)
		if err != nil || addr == 0 {
			continue
		}
		registerAddress(names, addr, fptrs)
		return true
	}
	return false
}

func registerAddress(names []string, addr uintptr, fptrs map[string]any) {
	for _, name := range names {
		fptr, ok := fptrs[name]
		if !ok || fptr == nil {
			continue
		}
		RegisterFunc(fptr, addr)
	}
}
`)
	return formatSource(b.Bytes())
}

func writeHeader(b *bytes.Buffer, pkg string) {
	fmt.Fprintf(b, "// Code generated by vulkangen. DO NOT EDIT.\n\npackage %s\n\n", pkg)
}

func writeDispatchStruct(b *bytes.Buffer, name, first string, commands []model.SelectedCommand) {
	fmt.Fprintf(b, "type %s struct {\n", name)
	if first != "" {
		fmt.Fprintf(b, "\t%s\n", first)
	}
	for _, cmd := range commands {
		fmt.Fprintf(b, "\t%s func(%s)%s\n", cmd.GoName, paramsSignature(cmd.Params), resultSignature(cmd.Return))
	}
	b.WriteString("}\n\n")
}

func writeCommandPointerMap(b *bytes.Buffer, name string, commands []model.SelectedCommand) {
	fmt.Fprintf(b, "func %s() map[string]any {\n", name)
	b.WriteString("\treturn map[string]any{\n")
	for _, cmd := range commands {
		fmt.Fprintf(b, "\t\t%q: &%s,\n", cmd.Name, rawFuncName(cmd.Name))
	}
	b.WriteString("\t}\n}\n\n")
}

type registerGroup struct {
	names    []string
	required bool
}

func writeRegisterGroup(b *bytes.Buffer, name string, commands []model.SelectedCommand) {
	fmt.Fprintf(b, "func Register%s(handle uintptr, lookup LookupFunc, fptrs map[string]any) error {\n", name)
	for _, group := range registerGroups(commands) {
		if !group.required {
			fmt.Fprintf(b, "\tregisterOptional([]string{%s}, handle, lookup, fptrs)\n", quotedNames(group.names))
			continue
		}
		fmt.Fprintf(b, "\tif err := registerRequired([]string{%s}, handle, lookup, fptrs); err != nil {\n\t\treturn err\n\t}\n", quotedNames(group.names))
	}
	b.WriteString("\treturn nil\n}\n\n")
}

func registerGroups(commands []model.SelectedCommand) []registerGroup {
	groupByKey := make(map[string]int, len(commands))
	var groups []registerGroup
	for _, cmd := range commands {
		key := cmd.Name
		if cmd.Source.Alias != "" {
			key = cmd.Source.Alias
		}
		idx, ok := groupByKey[key]
		if !ok {
			idx = len(groups)
			groupByKey[key] = idx
			groups = append(groups, registerGroup{})
		}
		groups[idx].names = append(groups[idx].names, cmd.Name)
		if !cmd.Optional {
			groups[idx].required = true
		}
	}
	return groups
}

func quotedNames(names []string) string {
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%q", name))
	}
	return strings.Join(parts, ", ")
}

type typeLayout struct {
	size  int
	align int
	ok    bool
}

func unionGoType(t model.SelectedType, types map[string]model.SelectedType) (string, error) {
	if len(t.Members) == 0 {
		return "[0]byte", nil
	}
	union, bestMember, best := unionLayout(t, types, make(map[string]bool))
	if !union.ok || !best.ok {
		return "", fmt.Errorf("cannot resolve union layout for %s", t.Name)
	}
	if best.size == union.size && best.align == union.align {
		return goFieldType(bestMember), nil
	}
	storage := unionStorageGoType(union.size, union.align)
	if storage == "" {
		return "", fmt.Errorf("cannot represent union %s layout size=%d align=%d", t.Name, union.size, union.align)
	}
	return storage, nil
}

func unionLayout(t model.SelectedType, types map[string]model.SelectedType, seen map[string]bool) (typeLayout, model.MemberDecl, typeLayout) {
	var union typeLayout
	var bestMember model.MemberDecl
	var best typeLayout
	for i, member := range t.Members {
		layout := memberLayout(member, types, seen)
		if !layout.ok {
			return typeLayout{}, t.Members[0], typeLayout{}
		}
		if i == 0 || layout.size > best.size || (layout.size == best.size && layout.align > best.align) {
			bestMember = member
			best = layout
		}
		if layout.size > union.size {
			union.size = layout.size
		}
		if layout.align > union.align {
			union.align = layout.align
		}
	}
	if union.align == 0 {
		union.align = 1
	}
	union.size = alignUp(union.size, union.align)
	union.ok = true
	return union, bestMember, best
}

func unionStorageGoType(size, align int) string {
	size = alignUp(size, align)
	switch {
	case align >= 8 && size%8 == 0:
		return fmt.Sprintf("[%d]uint64", size/8)
	case align >= 4 && size%4 == 0:
		return fmt.Sprintf("[%d]uint32", size/4)
	case align >= 2 && size%2 == 0:
		return fmt.Sprintf("[%d]uint16", size/2)
	default:
		return fmt.Sprintf("[%d]byte", size)
	}
}

func pointerLayout() typeLayout {
	size := strconv.IntSize / 8
	return typeLayout{size: size, align: size, ok: true}
}

func memberLayout(member model.MemberDecl, types map[string]model.SelectedType, seen map[string]bool) typeLayout {
	var layout typeLayout
	if member.PointerDepth > 0 {
		layout = pointerLayout()
	} else {
		layout = vkTypeLayout(member.Type, types, seen)
	}
	if !layout.ok {
		return layout
	}
	for _, lenValue := range member.ArrayLens {
		n, err := strconv.Atoi(strings.TrimSpace(lenValue))
		if err != nil {
			return typeLayout{}
		}
		layout.size *= n
	}
	return layout
}

func vkTypeLayout(name string, types map[string]model.SelectedType, seen map[string]bool) typeLayout {
	if layout := scalarLayout(name); layout.ok {
		return layout
	}
	t, ok := types[name]
	if !ok {
		return typeLayout{}
	}
	if t.Source.Alias != "" {
		return vkTypeLayout(t.Source.Alias, types, seen)
	}
	if seen[name] {
		return typeLayout{}
	}
	seen[name] = true
	defer delete(seen, name)
	switch t.Category {
	case "handle", "basetype", "bitmask", "enum", "funcpointer":
		return scalarLayout(t.GoType)
	case "struct":
		return structLayout(t, types, seen)
	case "union":
		layout, _, _ := unionLayout(t, types, seen)
		return layout
	default:
		return typeLayout{}
	}
}

func structLayout(t model.SelectedType, types map[string]model.SelectedType, seen map[string]bool) typeLayout {
	layout := typeLayout{align: 1, ok: true}
	for _, member := range t.Members {
		fieldLayout := memberLayout(member, types, seen)
		if !fieldLayout.ok {
			return typeLayout{}
		}
		layout.size = alignUp(layout.size, fieldLayout.align)
		layout.size += fieldLayout.size
		if fieldLayout.align > layout.align {
			layout.align = fieldLayout.align
		}
	}
	layout.size = alignUp(layout.size, layout.align)
	return layout
}

func scalarLayout(name string) typeLayout {
	switch name {
	case "char", "byte", "uint8_t", "uint8", "int8_t", "int8":
		return typeLayout{size: 1, align: 1, ok: true}
	case "uint16_t", "uint16", "int16_t", "int16":
		return typeLayout{size: 2, align: 2, ok: true}
	case "int", "float", "uint32_t", "uint32", "int32_t", "int32", "float32", "VkFlags", "Bool32", "Result":
		return typeLayout{size: 4, align: 4, ok: true}
	case "double", "uint64_t", "uint64", "int64_t", "int64", "float64", "VkFlags64", "VkDeviceSize", "DeviceSize":
		return typeLayout{size: 8, align: 8, ok: true}
	case "size_t", "uintptr", "unsafe.Pointer":
		return pointerLayout()
	default:
		return typeLayout{}
	}
}

func alignUp(value, align int) int {
	if align <= 1 {
		return value
	}
	rem := value % align
	if rem == 0 {
		return value
	}
	return value + align - rem
}

func commandsNeedUnsafe(commands []model.SelectedCommand) bool {
	for _, cmd := range commands {
		for _, param := range cmd.Params {
			if param.Type == "void" {
				return true
			}
		}
	}
	return false
}

func commandsByDispatch(sel *model.SelectedRegistry, level model.DispatchLevel) []model.SelectedCommand {
	var cmds []model.SelectedCommand
	for _, cmd := range sel.Commands {
		if cmd.Dispatch == level {
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

func paramsSignature(params []model.ParamDecl) string {
	parts := make([]string, 0, len(params))
	for _, p := range params {
		parts = append(parts, goParamType(p))
	}
	return strings.Join(parts, ", ")
}

func resultSignature(ret string) string {
	if ret == "" || ret == "void" {
		return ""
	}
	return " " + goTypeName(ret)
}

func goParamType(p model.ParamDecl) string {
	base := goTypeName(p.Type)
	if p.Type == "void" && p.PointerDepth > 0 {
		return voidPointerType(p.PointerDepth)
	}
	if len(p.ArrayLens) > 0 {
		for i := len(p.ArrayLens) - 1; i >= 0; i-- {
			base = "[" + arrayLenName(p.ArrayLens[i]) + "]" + base
		}
	}
	for i := 0; i < p.PointerDepth; i++ {
		base = "*" + base
	}
	return base
}

func voidPointerType(pointerDepth int) string {
	base := "unsafe.Pointer"
	for i := 1; i < pointerDepth; i++ {
		base = "*" + base
	}
	return base
}

func goFieldType(m model.MemberDecl) string {
	base := goTypeName(m.Type)
	if m.Type == "void" && m.PointerDepth > 0 {
		return voidPointerType(m.PointerDepth)
	}
	if len(m.ArrayLens) > 0 {
		for i := len(m.ArrayLens) - 1; i >= 0; i-- {
			base = "[" + arrayLenName(m.ArrayLens[i]) + "]" + base
		}
		return base
	}
	for i := 0; i < m.PointerDepth; i++ {
		base = "*" + base
	}
	return base
}

func arrayLenName(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "VK_") {
		return constName(value)
	}
	return value
}

func goTypeName(vkType string) string {
	switch vkType {
	case "void":
		return "unsafe.Pointer"
	case "char":
		return "byte"
	case "int":
		return "int32"
	case "float":
		return "float32"
	case "double":
		return "float64"
	case "uint8_t":
		return "uint8"
	case "uint16_t":
		return "uint16"
	case "uint32_t":
		return "uint32"
	case "uint64_t":
		return "uint64"
	case "int8_t":
		return "int8"
	case "int16_t":
		return "int16"
	case "int32_t":
		return "int32"
	case "int64_t":
		return "int64"
	case "size_t":
		return "uintptr"
	default:
		return strings.TrimPrefix(vkType, "Vk")
	}
}

func rawFuncName(name string) string {
	if strings.HasPrefix(name, "vk") {
		return "Vk" + name[2:]
	}
	return exportName(name)
}

func constName(name string) string {
	name = strings.TrimPrefix(name, "VK_")
	parts := strings.Split(strings.ToLower(name), "_")
	for i, p := range parts {
		parts[i] = initialismTitle(p)
	}
	return strings.Join(parts, "")
}

func exportName(name string) string {
	if len(name) > 1 && name[0] == 'p' && name[1] >= 'A' && name[1] <= 'Z' {
		name = name[1:]
	}
	if name == "" {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func initialismTitle(s string) string {
	upper := strings.ToUpper(s)
	switch upper {
	case "KHR", "EXT", "FD", "DRM", "DMA", "BUF", "ID", "API", "UUID", "LUID", "D3D", "IOS", "MACOS", "QCOM", "NV", "AMD", "ARM":
		return upper
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatSource(src []byte) (string, error) {
	formatted, err := format.Source(src)
	if err != nil {
		return "", fmt.Errorf("format generated source: %w\n%s", err, src)
	}
	return string(formatted), nil
}
