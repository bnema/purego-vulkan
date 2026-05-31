package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/model"
)

// ParseFile decodes a Vulkan registry XML file into the generator's raw model.
func ParseFile(path string) (*model.Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read registry: %w", err)
	}
	var raw xmlRegistry
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse registry XML: %w", err)
	}
	reg := raw.toModel()
	if len(reg.Types) == 0 && len(reg.Commands) == 0 && len(reg.EnumGroups) == 0 && len(reg.Extensions) == 0 && len(reg.Features) == 0 {
		return nil, fmt.Errorf("registry %s is empty", path)
	}
	return reg, nil
}

type xmlRegistry struct {
	XMLName    xml.Name       `xml:"registry"`
	Types      xmlTypes       `xml:"types"`
	EnumGroups []xmlEnumGroup `xml:"enums"`
	Commands   xmlCommands    `xml:"commands"`
	Features   []xmlFeature   `xml:"feature"`
	Extensions xmlExtensions  `xml:"extensions"`
}

type xmlTypes struct {
	Types []xmlType `xml:"type"`
}

type xmlType struct {
	Category      string      `xml:"category,attr"`
	NameAttr      string      `xml:"name,attr"`
	Alias         string      `xml:"alias,attr"`
	API           string      `xml:"api,attr"`
	Requires      string      `xml:"requires,attr"`
	Bitvalues     string      `xml:"bitvalues,attr"`
	Parent        string      `xml:"parent,attr"`
	ReturnedOnly  string      `xml:"returnedonly,attr"`
	StructExtends string      `xml:"structextends,attr"`
	InnerXML      string      `xml:",innerxml"`
	Members       []xmlMember `xml:"member"`
}

type xmlMember struct {
	InnerXML string `xml:",innerxml"`
	Len      string `xml:"len,attr"`
	Optional string `xml:"optional,attr"`
	Values   string `xml:"values,attr"`
}

type xmlEnumGroup struct {
	Name     string    `xml:"name,attr"`
	Type     string    `xml:"type,attr"`
	BitWidth string    `xml:"bitwidth,attr"`
	Enums    []xmlEnum `xml:"enum"`
}

type xmlEnum struct {
	Name    string `xml:"name,attr"`
	Value   string `xml:"value,attr"`
	Bitpos  string `xml:"bitpos,attr"`
	Alias   string `xml:"alias,attr"`
	Extends string `xml:"extends,attr"`
	Offset  string `xml:"offset,attr"`
	Dir     string `xml:"dir,attr"`
}

type xmlCommands struct {
	Commands []xmlCommand `xml:"command"`
}

type xmlCommand struct {
	NameAttr     string     `xml:"name,attr"`
	Alias        string     `xml:"alias,attr"`
	API          string     `xml:"api,attr"`
	Export       string     `xml:"export,attr"`
	SuccessCodes string     `xml:"successcodes,attr"`
	ErrorCodes   string     `xml:"errorcodes,attr"`
	Proto        xmlProto   `xml:"proto"`
	Params       []xmlParam `xml:"param"`
}

type xmlProto struct {
	InnerXML string `xml:",innerxml"`
}

type xmlParam struct {
	InnerXML string `xml:",innerxml"`
	Len      string `xml:"len,attr"`
	Optional string `xml:"optional,attr"`
}

type xmlFeature struct {
	Name     string       `xml:"name,attr"`
	Number   string       `xml:"number,attr"`
	API      string       `xml:"api,attr"`
	Requires []xmlRequire `xml:"require"`
}

type xmlExtensions struct {
	Extensions []xmlExtension `xml:"extension"`
}

type xmlExtension struct {
	Name      string       `xml:"name,attr"`
	Number    string       `xml:"number,attr"`
	Type      string       `xml:"type,attr"`
	Depends   string       `xml:"depends,attr"`
	Supported string       `xml:"supported,attr"`
	Platform  string       `xml:"platform,attr"`
	Requires  []xmlRequire `xml:"require"`
}

type xmlRequire struct {
	Depends  string    `xml:"depends,attr"`
	Types    []xmlRef  `xml:"type"`
	Commands []xmlRef  `xml:"command"`
	Enums    []xmlEnum `xml:"enum"`
}

type xmlRef struct {
	Name string `xml:"name,attr"`
}

func (r xmlRegistry) toModel() *model.Registry {
	reg := &model.Registry{}
	for _, t := range r.Types.Types {
		decl := t.toModel()
		if decl.Name != "" {
			reg.Types = append(reg.Types, decl)
		}
	}
	for _, g := range r.EnumGroups {
		reg.EnumGroups = append(reg.EnumGroups, g.toModel())
	}
	for _, c := range r.Commands.Commands {
		decl := c.toModel()
		if decl.Name != "" {
			reg.Commands = append(reg.Commands, decl)
		}
	}
	for _, f := range r.Features {
		reg.Features = append(reg.Features, f.toModel())
	}
	for _, e := range r.Extensions.Extensions {
		reg.Extensions = append(reg.Extensions, e.toModel())
	}
	return reg
}

func (t xmlType) toModel() model.TypeDecl {
	typed := parseTypedElement(t.InnerXML, "")
	decl := model.TypeDecl{
		Name:          typed.Name,
		Category:      t.Category,
		Type:          typed.Type,
		Alias:         t.Alias,
		API:           t.API,
		Requires:      t.Requires,
		Bitvalues:     t.Bitvalues,
		Parent:        t.Parent,
		ReturnedOnly:  t.ReturnedOnly == "true",
		StructExtends: t.StructExtends,
		RawText:       typed.RawText,
	}
	if t.NameAttr != "" {
		decl.Name = t.NameAttr
	}
	if t.Category == "struct" || t.Category == "union" || t.Category == "enum" {
		decl.Type = ""
	}
	for _, m := range t.Members {
		decl.Members = append(decl.Members, m.toModel())
	}
	return decl
}

func (m xmlMember) toModel() model.MemberDecl {
	t := parseTypedElement(m.InnerXML, "")
	return model.MemberDecl{
		Name:         t.Name,
		Type:         t.Type,
		RawText:      t.RawText,
		Const:        t.Const,
		PointerDepth: t.PointerDepth,
		ArrayLen:     t.ArrayLen,
		ArrayLens:    t.ArrayLens,
		Len:          m.Len,
		Optional:     m.Optional,
		Values:       m.Values,
	}
}

func (g xmlEnumGroup) toModel() model.EnumGroup {
	decl := model.EnumGroup{Name: g.Name, Type: g.Type, BitWidth: g.BitWidth}
	for _, e := range g.Enums {
		decl.Enums = append(decl.Enums, e.toModel())
	}
	return decl
}

func (e xmlEnum) toModel() model.EnumDecl {
	return model.EnumDecl{
		Name:    e.Name,
		Value:   e.Value,
		Bitpos:  e.Bitpos,
		Alias:   e.Alias,
		Extends: e.Extends,
		Offset:  e.Offset,
		Dir:     e.Dir,
	}
}

func (c xmlCommand) toModel() model.CommandDecl {
	proto := parseTypedElement(c.Proto.InnerXML, c.NameAttr)
	decl := model.CommandDecl{
		Name:         proto.Name,
		Return:       proto.Type,
		Alias:        c.Alias,
		API:          c.API,
		Export:       c.Export,
		SuccessCodes: c.SuccessCodes,
		ErrorCodes:   c.ErrorCodes,
	}
	if decl.Name == "" {
		decl.Name = c.NameAttr
	}
	for _, p := range c.Params {
		decl.Params = append(decl.Params, p.toModel())
	}
	return decl
}

func (p xmlParam) toModel() model.ParamDecl {
	t := parseTypedElement(p.InnerXML, "")
	return model.ParamDecl{
		Name:         t.Name,
		Type:         t.Type,
		RawText:      t.RawText,
		Const:        t.Const,
		PointerDepth: t.PointerDepth,
		ArrayLen:     t.ArrayLen,
		ArrayLens:    t.ArrayLens,
		Len:          p.Len,
		Optional:     p.Optional,
	}
}

func (f xmlFeature) toModel() model.FeatureDecl {
	decl := model.FeatureDecl{Name: f.Name, Number: f.Number, API: f.API}
	for _, req := range f.Requires {
		decl.Requires = append(decl.Requires, req.toModel())
	}
	return decl
}

func (e xmlExtension) toModel() model.ExtensionDecl {
	decl := model.ExtensionDecl{Name: e.Name, Number: e.Number, Type: e.Type, Depends: e.Depends, Supported: e.Supported, Platform: e.Platform}
	for _, req := range e.Requires {
		decl.Requires = append(decl.Requires, req.toModel())
	}
	return decl
}

func (r xmlRequire) toModel() model.RequireDecl {
	decl := model.RequireDecl{Depends: r.Depends}
	for _, ref := range r.Types {
		if ref.Name != "" {
			decl.Types = append(decl.Types, ref.Name)
		}
	}
	for _, ref := range r.Commands {
		if ref.Name != "" {
			decl.Commands = append(decl.Commands, ref.Name)
		}
	}
	for _, e := range r.Enums {
		decl.Enums = append(decl.Enums, e.toModel())
	}
	return decl
}

type typedElement struct {
	Name         string
	Type         string
	RawText      string
	Const        bool
	PointerDepth int
	ArrayLen     string
	ArrayLens    []string
}

var (
	typeRE       = regexp.MustCompile(`<type>([^<]+)</type>`)
	nameRE       = regexp.MustCompile(`<name>([^<]+)</name>`)
	tagRE        = regexp.MustCompile(`<[^>]+>`)
	arrayRE      = regexp.MustCompile(`<name>[^<]+</name>((?:\s*\[[^\]]+\])+)`)
	arrayDimRE   = regexp.MustCompile(`\[([^\]]+)\]`)
	spaceRE      = regexp.MustCompile(`\s+`)
	pointerStarR = regexp.MustCompile(`\*`)
)

func parseTypedElement(innerXML, nameAttr string) typedElement {
	out := typedElement{}
	if m := typeRE.FindStringSubmatch(innerXML); len(m) == 2 {
		out.Type = strings.TrimSpace(m[1])
	}
	if m := nameRE.FindStringSubmatch(innerXML); len(m) == 2 {
		out.Name = strings.TrimSpace(m[1])
	}
	if out.Name == "" {
		out.Name = nameAttr
	}
	if m := arrayRE.FindStringSubmatch(innerXML); len(m) == 2 {
		for _, dim := range arrayDimRE.FindAllStringSubmatch(m[1], -1) {
			if len(dim) == 2 {
				out.ArrayLens = append(out.ArrayLens, strings.TrimSpace(dim[1]))
			}
		}
		if len(out.ArrayLens) > 0 {
			out.ArrayLen = out.ArrayLens[0]
		}
	}

	raw := tagRE.ReplaceAllString(innerXML, " ")
	raw = strings.TrimSpace(spaceRE.ReplaceAllString(raw, " "))
	out.RawText = raw
	out.Const = strings.Contains(" "+raw+" ", " const ")
	out.PointerDepth = len(pointerStarR.FindAllString(raw, -1))
	return out
}
