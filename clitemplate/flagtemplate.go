package clitemplate

import (
	"text/template"

	pgs "github.com/lyft/protoc-gen-star"
)

type PFlagType string

const (
	Bool    PFlagType = "Bool"
	Int32   PFlagType = "Int32"
	Int64   PFlagType = "Int64"
	Uint32  PFlagType = "Uint32"
	Uint64  PFlagType = "Uint64"
	Float32 PFlagType = "Float32"
	Float64 PFlagType = "Float64"
	Bytes   PFlagType = "Bytes"
	String  PFlagType = "String"
)

func (p PFlagType) ZeroValue() string {
	switch p {
	case Bool:
		return "false"
	case Int32:
		return "0"
	case Int64:
		return "0"
	case Uint32:
		return "0"
	case Uint64:
		return "0"
	case Float32:
		return "0"
	case Float64:
		return "0"
	case Bytes:
		return "[]"
	case String:
		return `""`
	default:
		panic("unrecognised PFlagType")
	}
}

var (
	protoFlagTypeMap = map[pgs.ProtoType]PFlagType{

		pgs.FloatT:  Float32,
		pgs.DoubleT: Float64,

		pgs.Int32T:   Int32,
		pgs.SInt32:   Int32,
		pgs.SFixed32: Int32,

		pgs.UInt32T:  Uint32,
		pgs.Fixed32T: Uint32,

		pgs.Int64T:   Int64,
		pgs.SInt64:   Int64,
		pgs.SFixed64: Int64,

		pgs.UInt64T:  Uint64,
		pgs.Fixed64T: Uint64,

		pgs.BytesT:  Bytes,
		pgs.StringT: String,
		pgs.BoolT:   Bool,

		// Enum values can be identified by their string identifier
		pgs.EnumT: String,

		pgs.GroupT:   "",
		pgs.MessageT: "",
	}
)

type PFlag struct {
	Type      PFlagType
	Slice     bool
	StringMap bool
}

func NewPFlag(t pgs.ProtoType, slice bool) PFlag {
	return PFlag{
		Type:  protoFlagTypeMap[t],
		Slice: slice,
	}
}

func NewEnumFlag(slice bool) PFlag {
	return PFlag{
		Type:  String,
		Slice: slice,
	}
}

func NewStringMapFlag() PFlag {
	return PFlag{
		StringMap: true,
	}
}

func (p PFlag) FlagFunc() string {
	if p.StringMap {
		return "StringToString"
	}
	if p.Type == Bytes {
		return "BytesBase64"
	}
	if p.Slice {
		return string(p.Type) + "Slice"
	}
	return string(p.Type)
}

func (p PFlag) DefaultValue() string {
	if p.StringMap {
		return "map[string]string{}"
	}
	if p.Slice {
		return "[]"
	}
	return p.Type.ZeroValue()
}

type Flagset struct {
	Name  string
	Flags map[string]PFlag
}

var (
	flagTemplateFuncs = template.FuncMap{
		"FlagFunc": func(p PFlag) string {
			return p.FlagFunc()
		},
		"DefaultValue": func(p PFlag) string {
			return p.DefaultValue()
		},
	}

	FlagsetTemplate = template.Must(template.New("flagset").Funcs(flagTemplateFuncs).Parse(`
func {{.Name}}Flags(flags *pflag.FlagSet) {
	{{- range $field, $flag := .Flags}}
	flags.{{FlagFunc $flag}}("{{$field}}", {{DefaultValue $flag}}, "")
	{{- end}}
}
`))

	// Current strategy with this one is to convert to json then marshal to proto type
	ParseToProtoTemplate = template.Must(template.New("flagset").Funcs(flagTemplateFuncs).Parse(`
func {{.Name}}FromFlags(flags *pflag.FlagSet, dataFlagKey string) (*{{.Name}}, err) {
	var target {{.Name}}
	dataFlag := flags.Lookup(dataFlagKey)
	if dataFlag.Changed() {
		if err := json.Unmarshal([]byte(dataFlag.Value.String()), &target); err != nil {
			return nil, err
		}
		return target, nil
	}

	// read from specific flags
	panic("reading request values from flags not yet implemented")
}
`))
)
