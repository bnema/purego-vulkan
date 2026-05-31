package model

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type DispatchLevel string

const (
	DispatchGlobal   DispatchLevel = "global"
	DispatchInstance DispatchLevel = "instance"
	DispatchDevice   DispatchLevel = "device"
)

type SelectionConfig struct {
	RootTypes        []string
	Commands         []string
	Extensions       []string
	CoreVersions     []string
	CommandOverrides map[string]CommandOverride
}

type CommandOverride struct {
	Dispatch DispatchLevel
	Optional bool
}

type SelectedRegistry struct {
	Types      []SelectedType
	Constants  []SelectedConstant
	Commands   []SelectedCommand
	Extensions []SelectedExtension
}

func (r *SelectedRegistry) TypeByName(name string) *SelectedType {
	for i := range r.Types {
		if r.Types[i].Name == name {
			return &r.Types[i]
		}
	}
	return nil
}

func (r *SelectedRegistry) CommandByName(name string) *SelectedCommand {
	for i := range r.Commands {
		if r.Commands[i].Name == name {
			return &r.Commands[i]
		}
	}
	return nil
}

func (r *SelectedRegistry) ExtensionByName(name string) *SelectedExtension {
	for i := range r.Extensions {
		if r.Extensions[i].Name == name {
			return &r.Extensions[i]
		}
	}
	return nil
}

type SelectedType struct {
	Name         string
	Category     string
	GoName       string
	GoType       string
	Dispatchable bool
	Members      []MemberDecl
	Source       TypeDecl
}

type SelectedConstant struct {
	Name    string
	Value   string
	Extends string
	Source  EnumDecl
}

type SelectedCommand struct {
	Name     string
	GoName   string
	Return   string
	Dispatch DispatchLevel
	Optional bool
	Params   []ParamDecl
	Source   CommandDecl
}

type SelectedExtension struct {
	Name     string
	Type     string
	Commands []string
	Types    []string
	Enums    []EnumDecl
	Source   ExtensionDecl
}

// Select filters the raw Vulkan registry to the configured command/extension
// subset and adds the dependency closure of referenced types.
func Select(reg *Registry, cfg SelectionConfig) (*SelectedRegistry, error) {
	if reg == nil {
		return nil, fmt.Errorf("select registry: nil registry")
	}
	idx := newIndex(reg)
	state := selectionState{
		reg:              reg,
		idx:              idx,
		cfg:              cfg,
		selectedTypes:    make(map[string]bool),
		selectedCommands: make(map[string]bool),
		commandOptional:  make(map[string]bool),
		selectedExts:     make(map[string]bool),
		constantNames:    make(map[string]bool),
		selectedCore:     make(map[string]bool),
		configuredExts:   make(map[string]bool),
		missingNames:     make(map[string]bool),
	}
	coreVersions := cfg.CoreVersions
	if len(coreVersions) == 0 {
		coreVersions = defaultCoreVersions()
	}
	for _, version := range coreVersions {
		state.selectedCore[version] = true
	}
	for _, ext := range cfg.Extensions {
		state.configuredExts[ext] = true
	}

	for _, name := range cfg.RootTypes {
		state.addType(name, "configured root type")
	}
	for _, name := range cfg.Commands {
		if _, ok := idx.commands[name]; !ok {
			state.addMissing("command", name, "configured command")
			continue
		}
		state.addCommand(name, cfg.CommandOverrides[name].Optional, "configured command")
	}
	for _, name := range cfg.Extensions {
		state.addExtension(name, true, "configured extension")
	}

	for len(state.typeQueue) > 0 {
		name := state.typeQueue[0]
		state.typeQueue = state.typeQueue[1:]
		decl := idx.types[name]
		state.addEnumGroupConstants(decl.Name)
		if decl.Category != "handle" {
			state.addType(decl.Type, "type "+decl.Name)
		}
		state.addType(decl.Requires, "type "+decl.Name)
		state.addType(decl.Bitvalues, "type "+decl.Name)
		state.addEnumGroupConstants(decl.Requires)
		state.addEnumGroupConstants(decl.Bitvalues)
		state.addType(decl.Alias, "type "+decl.Name)
		for _, member := range decl.Members {
			state.addType(member.Type, "member "+decl.Name+"."+member.Name)
			state.addFeatureConstant(member.Values)
		}
	}

	if err := state.missingError(); err != nil {
		return nil, err
	}
	return state.buildSelected()
}

type registryIndex struct {
	types        map[string]TypeDecl
	commands     map[string]CommandDecl
	commandOrder []string
	extensions   map[string]ExtensionDecl
	enumGroups   map[string]EnumGroup
	featureEnums map[string]EnumDecl
}

func preferCommand(candidate, existing CommandDecl) bool {
	candidateScore := commandVariantScore(candidate)
	existingScore := commandVariantScore(existing)
	return candidateScore > existingScore
}

func commandVariantScore(cmd CommandDecl) int {
	api := strings.ToLower(cmd.API)
	export := strings.ToLower(cmd.Export)
	score := 0
	if api == "" || strings.Contains(api, "vulkan") {
		score += 2
	}
	if strings.Contains(api, "vulkansc") && !strings.Contains(api, "vulkan,") {
		score -= 2
	}
	if export == "" || strings.Contains(export, "vulkan") {
		score++
	}
	if strings.Contains(export, "vulkansc") && !strings.Contains(export, "vulkan,") {
		score -= 2
	}
	if cmd.Return != "" || len(cmd.Params) > 0 {
		score++
	}
	return score
}

func newIndex(reg *Registry) registryIndex {
	idx := registryIndex{
		types:        make(map[string]TypeDecl, len(reg.Types)),
		commands:     make(map[string]CommandDecl, len(reg.Commands)),
		extensions:   make(map[string]ExtensionDecl, len(reg.Extensions)),
		enumGroups:   make(map[string]EnumGroup, len(reg.EnumGroups)),
		featureEnums: make(map[string]EnumDecl),
	}
	for _, t := range reg.Types {
		idx.types[t.Name] = t
	}
	for _, c := range reg.Commands {
		if existing, ok := idx.commands[c.Name]; ok {
			if preferCommand(c, existing) {
				idx.commands[c.Name] = c
			}
			continue
		}
		idx.commands[c.Name] = c
		idx.commandOrder = append(idx.commandOrder, c.Name)
	}
	for _, e := range reg.Extensions {
		idx.extensions[e.Name] = e
	}
	for _, g := range reg.EnumGroups {
		idx.enumGroups[g.Name] = g
	}
	for _, feature := range reg.Features {
		if feature.API != "" && !strings.Contains(feature.API, "vulkan") {
			continue
		}
		for _, req := range feature.Requires {
			for _, enum := range req.Enums {
				if enum.Name != "" {
					idx.featureEnums[enum.Name] = enum
				}
			}
		}
	}
	for _, ext := range reg.Extensions {
		if ext.Supported == "disabled" {
			continue
		}
		for _, req := range ext.Requires {
			for _, enum := range req.Enums {
				if enum.Name != "" {
					idx.featureEnums[enum.Name] = enum
				}
			}
		}
	}
	for name, enum := range idx.featureEnums {
		idx.featureEnums[name] = idx.resolveEnumAlias(enum, nil)
	}
	for groupName, group := range idx.enumGroups {
		for i, enum := range group.Enums {
			group.Enums[i] = idx.resolveEnumAlias(enum, nil)
		}
		idx.enumGroups[groupName] = group
	}
	return idx
}

func (idx registryIndex) normalizeEnum(enum EnumDecl) EnumDecl {
	if enum.Value == "" && enum.Bitpos == "" && enum.Offset != "" && enum.ExtNumber != "" {
		extNumber, extErr := strconv.ParseInt(enum.ExtNumber, 10, 64)
		offset, offsetErr := strconv.ParseInt(enum.Offset, 10, 64)
		if extErr == nil && offsetErr == nil {
			value := int64(1000000000) + (extNumber-1)*1000 + offset
			if enum.Dir == "-" {
				value = -value
			}
			enum.Value = strconv.FormatInt(value, 10)
		}
	}
	return enum
}

func (idx registryIndex) resolveEnumAlias(enum EnumDecl, seen map[string]bool) EnumDecl {
	enum = idx.normalizeEnum(enum)
	if enum.Alias == "" {
		return enum
	}
	if seen == nil {
		seen = make(map[string]bool)
	}
	if seen[enum.Name] {
		return enum
	}
	seen[enum.Name] = true
	if target, ok := idx.featureEnums[enum.Alias]; ok {
		resolved := idx.resolveEnumAlias(target, seen)
		resolved.Name = enum.Name
		resolved.Alias = enum.Alias
		if enum.Extends != "" {
			resolved.Extends = enum.Extends
		}
		return resolved
	}
	for _, group := range idx.enumGroups {
		for _, candidate := range group.Enums {
			if candidate.Name == enum.Alias {
				resolved := idx.resolveEnumAlias(candidate, seen)
				resolved.Name = enum.Name
				resolved.Alias = enum.Alias
				if enum.Extends != "" {
					resolved.Extends = enum.Extends
				}
				return resolved
			}
		}
	}
	return enum
}

type selectionState struct {
	reg *Registry
	idx registryIndex
	cfg SelectionConfig

	selectedTypes    map[string]bool
	selectedCommands map[string]bool
	commandOptional  map[string]bool
	selectedExts     map[string]bool
	constantNames    map[string]bool
	constants        []EnumDecl
	typeQueue        []string
	selectedCore     map[string]bool
	configuredExts   map[string]bool
	missingNames     map[string]bool
	missing          []error
}

func (s *selectionState) addType(name, context string) {
	if name == "" || isBuiltinType(name) || s.selectedTypes[name] {
		return
	}
	if _, ok := s.idx.types[name]; !ok {
		s.addMissing("type", name, context)
		return
	}
	s.selectedTypes[name] = true
	s.typeQueue = append(s.typeQueue, name)
}

func defaultCoreVersions() []string {
	return []string{"VK_VERSION_1_0", "VK_VERSION_1_1", "VK_VERSION_1_2"}
}

func (s *selectionState) extensionDepsToAdd(expr string) ([]string, bool) {
	expr = trimOuterParens(strings.TrimSpace(expr))
	if expr == "" {
		return nil, true
	}
	if parts := splitTopLevel(expr, ','); len(parts) > 1 {
		bestOK := false
		var best []string
		for _, part := range parts {
			deps, ok := s.extensionDepsToAdd(part)
			if !ok {
				continue
			}
			if !bestOK || len(deps) < len(best) {
				bestOK = true
				best = deps
			}
		}
		return uniqueStrings(best), bestOK
	}
	if parts := splitTopLevel(expr, '+'); len(parts) > 1 {
		var deps []string
		for _, part := range parts {
			partDeps, ok := s.extensionDepsToAdd(part)
			if !ok {
				return nil, false
			}
			deps = append(deps, partDeps...)
		}
		return uniqueStrings(deps), true
	}
	if strings.HasPrefix(expr, "VK_VERSION_") {
		return nil, s.selectedCore[expr]
	}
	if strings.HasPrefix(expr, "VK_") {
		if s.selectedExts[expr] || s.configuredExts[expr] {
			return nil, true
		}
		ext, ok := s.idx.extensions[expr]
		if !ok || ext.Platform != "" || ext.Supported == "disabled" {
			return nil, false
		}
		return []string{expr}, true
	}
	return nil, true
}

func (s *selectionState) dependencyExprSatisfied(expr string) bool {
	deps, ok := s.extensionDepsToAdd(expr)
	return ok && len(deps) == 0
}

func trimOuterParens(expr string) string {
	for strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		depth := 0
		balanced := true
		for i, r := range expr {
			switch r {
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 && i != len(expr)-1 {
					balanced = false
				}
			}
			if depth < 0 {
				balanced = false
				break
			}
		}
		if !balanced || depth != 0 {
			break
		}
		expr = strings.TrimSpace(expr[1 : len(expr)-1])
	}
	return expr
}

func splitTopLevel(expr string, sep rune) []string {
	var parts []string
	depth := 0
	start := 0
	for i, r := range expr {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		default:
			if r == sep && depth == 0 {
				parts = append(parts, strings.TrimSpace(expr[start:i]))
				start = i + len(string(r))
			}
		}
	}
	if len(parts) == 0 {
		return []string{strings.TrimSpace(expr)}
	}
	parts = append(parts, strings.TrimSpace(expr[start:]))
	return parts
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, value := range in {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func (s *selectionState) addExtension(name string, optionalCommands bool, context string) {
	ext, ok := s.idx.extensions[name]
	if !ok {
		s.addMissing("extension", name, context)
		return
	}
	if ext.Platform != "" || ext.Supported == "disabled" {
		return
	}
	if s.selectedExts[name] {
		return
	}
	s.selectedExts[name] = true

	deps, ok := s.extensionDepsToAdd(ext.Depends)
	if !ok {
		s.addMissing("dependency", ext.Depends, "extension "+name)
	}
	for _, depName := range deps {
		s.addExtension(depName, optionalCommands, "dependency of extension "+name)
	}
	for _, req := range ext.Requires {
		if !s.dependencyExprSatisfied(req.Depends) {
			continue
		}
		for _, typeName := range req.Types {
			s.addType(typeName, "extension "+name)
		}
		for _, cmdName := range req.Commands {
			if _, ok := s.idx.commands[cmdName]; !ok {
				s.addMissing("command", cmdName, "extension "+name)
				continue
			}
			s.addCommand(cmdName, optionalCommands, "extension "+name)
		}
		for _, enum := range req.Enums {
			s.addResolvedConstant(enum)
		}
	}
}

func (s *selectionState) addCommand(name string, optional bool, context string) {
	if s.selectedCommands[name] {
		if optional {
			s.commandOptional[name] = true
		}
		return
	}
	cmd, ok := s.resolveCommand(name)
	if !ok {
		s.addMissing("command", name, context)
		return
	}
	s.selectedCommands[name] = true
	if optional || s.cfg.CommandOverrides[name].Optional {
		s.commandOptional[name] = true
	}
	s.addType(cmd.Return, "return type for "+name)
	for _, param := range cmd.Params {
		s.addType(param.Type, "parameter "+name+"."+param.Name)
	}
}

func (s *selectionState) addResolvedConstant(enum EnumDecl) {
	if resolved, ok := s.idx.featureEnums[enum.Name]; ok {
		s.addConstant(resolved, resolved.Extends)
		return
	}
	s.addConstant(s.idx.resolveEnumAlias(enum, nil), enum.Extends)
}

func (s *selectionState) addConstant(enum EnumDecl, defaultExtends string) {
	if enum.Name == "" || s.constantNames[enum.Name] {
		return
	}
	if enum.Extends == "" {
		enum.Extends = defaultExtends
	}
	s.constantNames[enum.Name] = true
	s.constants = append(s.constants, enum)
}

func (s *selectionState) addFeatureConstant(name string) {
	if name == "" {
		return
	}
	for _, enumName := range strings.Split(name, ",") {
		enumName = strings.TrimSpace(enumName)
		s.addFeatureConstantRecursive(enumName, make(map[string]bool))
	}
}

func (s *selectionState) addFeatureConstantRecursive(enumName string, seen map[string]bool) {
	if enumName == "" || seen[enumName] {
		return
	}
	seen[enumName] = true
	if enum, ok := s.idx.featureEnums[enumName]; ok {
		s.addConstant(enum, enum.Extends)
	}
	for _, enum := range s.aliasConstantsFor(enumName) {
		s.addConstant(enum, enum.Extends)
		s.addFeatureConstantRecursive(enum.Name, seen)
	}
}

func (s *selectionState) aliasConstantsFor(targetName string) []EnumDecl {
	var aliases []EnumDecl
	for _, enum := range s.idx.featureEnums {
		if enum.Alias == targetName {
			aliases = append(aliases, enum)
		}
	}
	return aliases
}

func (s *selectionState) addEnumGroupConstants(groupName string) {
	if groupName == "" {
		return
	}
	group, ok := s.idx.enumGroups[groupName]
	if !ok {
		return
	}
	for _, enum := range group.Enums {
		s.addConstant(enum, group.Name)
	}
}

func (s *selectionState) addMissing(kind, name, context string) {
	if name == "" || isBuiltinType(name) {
		return
	}
	key := kind + ":" + name + ":" + context
	if s.missingNames[key] {
		return
	}
	s.missingNames[key] = true
	s.missing = append(s.missing, fmt.Errorf("missing %s %s referenced by %s", kind, name, context))
}

func (s *selectionState) missingError() error {
	if len(s.missing) == 0 {
		return nil
	}
	return fmt.Errorf("select Vulkan subset: %w", errors.Join(s.missing...))
}

func (s *selectionState) buildSelected() (*SelectedRegistry, error) {
	out := &SelectedRegistry{}
	for _, decl := range s.reg.Types {
		if !s.selectedTypes[decl.Name] {
			continue
		}
		out.Types = append(out.Types, s.selectType(decl))
	}
	for _, enum := range s.constants {
		out.Constants = append(out.Constants, SelectedConstant{Name: enum.Name, Value: enum.Value, Extends: enum.Extends, Source: enum})
	}
	for _, name := range s.idx.commandOrder {
		if !s.selectedCommands[name] {
			continue
		}
		decl, ok := s.resolveCommand(name)
		if !ok {
			continue
		}
		dispatch, err := s.dispatchFor(decl)
		if err != nil {
			return nil, err
		}
		out.Commands = append(out.Commands, SelectedCommand{
			Name:     decl.Name,
			GoName:   goName(decl.Name),
			Return:   decl.Return,
			Dispatch: dispatch,
			Optional: s.commandOptional[decl.Name],
			Params:   slices.Clone(decl.Params),
			Source:   decl,
		})
	}
	for _, decl := range s.reg.Extensions {
		if !s.selectedExts[decl.Name] {
			continue
		}
		sel := SelectedExtension{Name: decl.Name, Type: decl.Type, Source: decl}
		for _, req := range decl.Requires {
			sel.Types = append(sel.Types, req.Types...)
			sel.Commands = append(sel.Commands, req.Commands...)
			sel.Enums = append(sel.Enums, req.Enums...)
		}
		out.Extensions = append(out.Extensions, sel)
	}
	return out, nil
}

func (s *selectionState) resolveCommand(name string) (CommandDecl, bool) {
	cmd, ok := s.idx.commands[name]
	if !ok {
		return CommandDecl{}, false
	}
	if cmd.Alias == "" || (cmd.Return != "" || len(cmd.Params) > 0) {
		return cmd, true
	}
	target, ok := s.resolveCommand(cmd.Alias)
	if !ok {
		return CommandDecl{}, false
	}
	target.Name = cmd.Name
	target.Alias = cmd.Alias
	target.API = cmd.API
	target.Export = cmd.Export
	return target, true
}

func (s *selectionState) dispatchFor(cmd CommandDecl) (DispatchLevel, error) {
	if ov := s.cfg.CommandOverrides[cmd.Name]; ov.Dispatch != "" {
		return ov.Dispatch, nil
	}
	if isGlobalCommand(cmd.Name) {
		return DispatchGlobal, nil
	}
	for _, param := range cmd.Params {
		t, ok := s.idx.types[param.Type]
		if !ok || t.Category != "handle" || t.Type != "VK_DEFINE_HANDLE" {
			continue
		}
		switch param.Type {
		case "VkInstance", "VkPhysicalDevice":
			return DispatchInstance, nil
		case "VkDevice", "VkQueue", "VkCommandBuffer":
			return DispatchDevice, nil
		default:
			if t.Parent == "VkDevice" {
				return DispatchDevice, nil
			}
		}
	}
	return DispatchGlobal, nil
}

func (s *selectionState) selectType(decl TypeDecl) SelectedType {
	resolved := s.resolveTypeDecl(decl, nil)
	goType, dispatchable := s.goTypeFor(resolved, nil)
	return SelectedType{
		Name:         decl.Name,
		Category:     resolved.Category,
		GoName:       goName(decl.Name),
		GoType:       goType,
		Dispatchable: dispatchable,
		Members:      slices.Clone(resolved.Members),
		Source:       resolved,
	}
}

func (s *selectionState) resolveTypeDecl(decl TypeDecl, seen map[string]bool) TypeDecl {
	if decl.Alias == "" {
		return decl
	}
	if seen == nil {
		seen = make(map[string]bool)
	}
	if seen[decl.Name] {
		return decl
	}
	seen[decl.Name] = true
	target, ok := s.idx.types[decl.Alias]
	if !ok {
		return decl
	}
	resolved := s.resolveTypeDecl(target, seen)
	resolved.Name = decl.Name
	resolved.Alias = decl.Alias
	resolved.API = decl.API
	return resolved
}

func (s *selectionState) goTypeFor(decl TypeDecl, seen map[string]bool) (string, bool) {
	if decl.Alias != "" && (decl.Type == "" || goScalarType(decl.Type) == decl.Type) {
		if seen == nil {
			seen = make(map[string]bool)
		}
		if !seen[decl.Name] {
			seen[decl.Name] = true
			if target, ok := s.idx.types[decl.Alias]; ok {
				return s.goTypeFor(target, seen)
			}
		}
	}
	if decl.Category == "handle" {
		if decl.Type == "VK_DEFINE_HANDLE" {
			return "uintptr", true
		}
		return "uint64", false
	}
	if decl.Category == "enum" {
		return "int32", false
	}
	if decl.Category == "basetype" || decl.Category == "bitmask" {
		return goScalarType(decl.Type), false
	}
	return "", false
}

func goScalarType(cType string) string {
	switch cType {
	case "uint8_t":
		return "uint8"
	case "uint16_t":
		return "uint16"
	case "uint32_t", "VkFlags":
		return "uint32"
	case "uint64_t", "VkFlags64", "VkDeviceSize":
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
		return goName(cType)
	}
}

func isGlobalCommand(name string) bool {
	switch name {
	case "vkGetInstanceProcAddr", "vkCreateInstance", "vkEnumerateInstanceVersion", "vkEnumerateInstanceExtensionProperties", "vkEnumerateInstanceLayerProperties":
		return true
	default:
		return false
	}
}

func isBuiltinType(name string) bool {
	switch name {
	case "", "void", "char", "int", "float", "double", "size_t",
		"uint8_t", "uint16_t", "uint32_t", "uint64_t",
		"int8_t", "int16_t", "int32_t", "int64_t":
		return true
	default:
		return false
	}
}

func goName(name string) string {
	if len(name) > 2 && name[:2] == "Vk" {
		return name[2:]
	}
	if len(name) > 2 && name[:2] == "vk" {
		return name[2:]
	}
	return name
}
