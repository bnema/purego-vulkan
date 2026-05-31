package model

import (
	"fmt"
	"slices"
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
	}

	for _, name := range cfg.RootTypes {
		state.addType(name)
	}
	for _, name := range cfg.Commands {
		if _, ok := idx.commands[name]; ok {
			state.addCommand(name, cfg.CommandOverrides[name].Optional)
		}
	}
	for _, name := range cfg.Extensions {
		ext, ok := idx.extensions[name]
		if !ok || ext.Platform != "" || ext.Supported == "disabled" {
			continue
		}
		state.selectedExts[name] = true
		for _, req := range ext.Requires {
			for _, typeName := range req.Types {
				state.addType(typeName)
			}
			for _, cmdName := range req.Commands {
				if _, ok := idx.commands[cmdName]; ok {
					state.addCommand(cmdName, true)
				}
			}
			for _, enum := range req.Enums {
				state.addConstant(enum)
			}
		}
	}

	for len(state.typeQueue) > 0 {
		name := state.typeQueue[0]
		state.typeQueue = state.typeQueue[1:]
		decl := idx.types[name]
		state.addType(decl.Type)
		state.addType(decl.Requires)
		state.addType(decl.Bitvalues)
		state.addType(decl.Alias)
		for _, member := range decl.Members {
			state.addType(member.Type)
		}
	}

	return state.buildSelected()
}

type registryIndex struct {
	types      map[string]TypeDecl
	commands   map[string]CommandDecl
	extensions map[string]ExtensionDecl
}

func newIndex(reg *Registry) registryIndex {
	idx := registryIndex{
		types:      make(map[string]TypeDecl, len(reg.Types)),
		commands:   make(map[string]CommandDecl, len(reg.Commands)),
		extensions: make(map[string]ExtensionDecl, len(reg.Extensions)),
	}
	for _, t := range reg.Types {
		idx.types[t.Name] = t
	}
	for _, c := range reg.Commands {
		idx.commands[c.Name] = c
	}
	for _, e := range reg.Extensions {
		idx.extensions[e.Name] = e
	}
	return idx
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
}

func (s *selectionState) addType(name string) {
	if name == "" || isBuiltinType(name) || s.selectedTypes[name] {
		return
	}
	if _, ok := s.idx.types[name]; !ok {
		return
	}
	s.selectedTypes[name] = true
	s.typeQueue = append(s.typeQueue, name)
}

func (s *selectionState) addCommand(name string, optional bool) {
	if s.selectedCommands[name] {
		if optional {
			s.commandOptional[name] = true
		}
		return
	}
	cmd, ok := s.idx.commands[name]
	if !ok {
		return
	}
	s.selectedCommands[name] = true
	if optional || s.cfg.CommandOverrides[name].Optional {
		s.commandOptional[name] = true
	}
	s.addType(cmd.Return)
	for _, param := range cmd.Params {
		s.addType(param.Type)
	}
}

func (s *selectionState) addConstant(enum EnumDecl) {
	if enum.Name == "" || s.constantNames[enum.Name] {
		return
	}
	s.constantNames[enum.Name] = true
	s.constants = append(s.constants, enum)
}

func (s *selectionState) buildSelected() (*SelectedRegistry, error) {
	out := &SelectedRegistry{}
	for _, decl := range s.reg.Types {
		if !s.selectedTypes[decl.Name] {
			continue
		}
		out.Types = append(out.Types, selectType(decl))
	}
	for _, enum := range s.constants {
		out.Constants = append(out.Constants, SelectedConstant{Name: enum.Name, Value: enum.Value, Extends: enum.Extends, Source: enum})
	}
	for _, decl := range s.reg.Commands {
		if !s.selectedCommands[decl.Name] {
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

func selectType(decl TypeDecl) SelectedType {
	goType, dispatchable := goTypeFor(decl)
	return SelectedType{
		Name:         decl.Name,
		Category:     decl.Category,
		GoName:       goName(decl.Name),
		GoType:       goType,
		Dispatchable: dispatchable,
		Members:      slices.Clone(decl.Members),
		Source:       decl,
	}
}

func goTypeFor(decl TypeDecl) (string, bool) {
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
