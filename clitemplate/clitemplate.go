package clitemplate

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type Command struct {
	Name               string
	Alias              string
	SubCommands        []Command
	RequestMessageType string
	Dataflag           string
	HasRequest         bool
	Runnable           bool
}

var (
	tmplFuncs = template.FuncMap{
		"joinSubCommands": func(cmds []Command) string {
			buf := bytes.Buffer{}
			for i, cmd := range cmds {
				if i == 0 {
					fmt.Fprintf(&buf, "%sCommand()", cmd.Name)
				} else {
					fmt.Fprintf(&buf, ", %sCommand()", cmd.Name)
				}
			}
			return buf.String()
		},
		"ToLower": func(s string) string { return strings.ToLower(s) },
	}

	CmdTemplate = template.Must(template.New("cmdTemplate").Funcs(tmplFuncs).Parse(`
func {{.Name}}Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "{{ToLower .Name}}",
		{{- if .Runnable }}
		RunE: run{{.Name}}Command,
		{{- end }}
	}
	{{- if .SubCommands}}
	cmd.AddCommand({{joinSubCommands .SubCommands}})
	{{- end}}
	{{- if .HasRequest}}{{- if .Dataflag}}
	cmd.Flags().StringP("request", "{{.Dataflag}}", "", "request object in json format")
	{{- else}}
	cmd.Flags().String("request", "", "request object in json format")
	{{- end}}{{- end}}
	{{- if .RequestMessageType}}
	{{.RequestMessageType}}Flags(cmd.Flags())
	{{- end}}
	return cmd
}
{{ if .Runnable }}
func run{{.Name}}Command(cmd *cobra.Command, args []string) error {
	// TODO: implement
	panic("unimplemented")
}
{{- end }}
`))
)
