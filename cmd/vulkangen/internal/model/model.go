package model

// Registry is the raw, normalized-enough representation decoded from vk.xml.
type Registry struct {
	Types      []TypeDecl
	EnumGroups []EnumGroup
	Commands   []CommandDecl
	Features   []FeatureDecl
	Extensions []ExtensionDecl
}

func (r *Registry) TypeByName(name string) *TypeDecl {
	for i := range r.Types {
		if r.Types[i].Name == name {
			return &r.Types[i]
		}
	}
	return nil
}

func (r *Registry) CommandByName(name string) *CommandDecl {
	for i := range r.Commands {
		if r.Commands[i].Name == name {
			return &r.Commands[i]
		}
	}
	return nil
}

func (r *Registry) EnumGroupByName(name string) *EnumGroup {
	for i := range r.EnumGroups {
		if r.EnumGroups[i].Name == name {
			return &r.EnumGroups[i]
		}
	}
	return nil
}

type TypeDecl struct {
	Name          string
	Category      string
	Type          string
	Alias         string
	API           string
	Requires      string
	Bitvalues     string
	Parent        string
	ReturnedOnly  bool
	StructExtends string
	RawText       string
	Members       []MemberDecl
}

type MemberDecl struct {
	Name         string
	Type         string
	RawText      string
	Const        bool
	PointerDepth int
	ArrayLen     string
	ArrayLens    []string
	Len          string
	Optional     string
	Values       string
}

type EnumGroup struct {
	Name     string
	Type     string
	BitWidth string
	Enums    []EnumDecl
}

type EnumDecl struct {
	Name    string
	Value   string
	Bitpos  string
	Alias   string
	Extends string
	Offset  string
	Dir     string
}

type CommandDecl struct {
	Name         string
	Return       string
	Alias        string
	API          string
	Export       string
	SuccessCodes string
	ErrorCodes   string
	Params       []ParamDecl
}

type ParamDecl struct {
	Name         string
	Type         string
	RawText      string
	Const        bool
	PointerDepth int
	ArrayLen     string
	ArrayLens    []string
	Len          string
	Optional     string
}

type FeatureDecl struct {
	Name     string
	Number   string
	API      string
	Requires []RequireDecl
}

type ExtensionDecl struct {
	Name      string
	Number    string
	Type      string
	Depends   string
	Supported string
	Platform  string
	Requires  []RequireDecl
}

type RequireDecl struct {
	Depends  string
	Types    []string
	Commands []string
	Enums    []EnumDecl
}
