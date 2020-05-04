package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/anzboi/protoc-gen-cli/clioption"
	"github.com/anzboi/protoc-gen-cli/clitemplate"
	"github.com/jinzhu/inflection"
	pgs "github.com/lyft/protoc-gen-star"
	"github.com/pkg/errors"
)

type PrinterModule struct {
	*pgs.ModuleBase
	pgs.Visitor
}

func CliPrinter() *PrinterModule { return &PrinterModule{ModuleBase: &pgs.ModuleBase{}} }

func (p *PrinterModule) Name() string { return "cli-gen" }

func (p *PrinterModule) Execute(targets map[string]pgs.File, packages map[string]pgs.Package) []pgs.Artifact {
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "package cli\n")
	cliVisitor := NewCLIVisitor(&buf, p)
	for _, pkg := range packages {
		pgs.Walk(cliVisitor, pkg)
	}
	p.AddGeneratorFile("name.cli.go", buf.String())
	return p.Artifacts()
}

type cliVisitor struct {
	Out io.Writer
	pgs.Visitor
	pgs.DebuggerCommon
}

func NewCLIVisitor(output io.Writer, d pgs.DebuggerCommon) pgs.Visitor {
	return &cliVisitor{
		Out:            output,
		Visitor:        pgs.NilVisitor(),
		DebuggerCommon: d,
	}
}

func (cv *cliVisitor) VisitPackage(pgs.Package) (pgs.Visitor, error) { return cv, nil }
func (cv *cliVisitor) VisitFile(pgs.File) (pgs.Visitor, error)       { return cv, nil }

func (cv *cliVisitor) VisitService(svc pgs.Service) (pgs.Visitor, error) {
	if svc.BuildTarget() {
		cv.Debugf("Found targeted service %s", svc.Name())

		rpcs := map[string]bool{}
		resources := map[string]map[string]string{}
		messageFlagSets := map[string]clitemplate.Flagset{}

		for _, method := range svc.Methods() {
			// Create a command for each method
			rpcs[method.FullyQualifiedName()] = true

			// analyze the method name for potential resource-verb combos
			res, verb, ok := MethodResourceVerb(cv, method)
			if ok {
				resources = AddResourceVerb(resources, res, verb, method.Input().Name().String())
			}

			// Produce a flagset for the request message type
			if _, ok := messageFlagSets[method.Input().FullyQualifiedName()]; !ok {
				m := &FlagsetVisitor{
					Visitor:        pgs.NilVisitor(),
					DebuggerCommon: cv,
					flags:          map[string]clitemplate.PFlag{},
				}
				pgs.Walk(m, method.Input())
				messageFlagSets[method.Input().FullyQualifiedName()] = clitemplate.Flagset{
					Name:  method.Input().Name().String(),
					Flags: m.flags,
				}
			}
		}

		cv.Debug(messageFlagSets)

		// for each resource, create another command
		for resource, verbs := range resources {
			verbCommands := []clitemplate.Command{}
			for verb, msg := range verbs {
				cmdName := PascalCaseName(svc.Name().String(), resource, verb)
				cmd := clitemplate.Command{
					Name:               cmdName,
					RequestMessageType: msg,
					Dataflag:           "d",
					HasRequest:         true,
					Runnable:           true,
				}
				verbCommands = append(verbCommands, cmd)
				clitemplate.CmdTemplate.Execute(cv.Out, cmd)
				fmt.Fprintf(cv.Out, "\n")
			}
			cmdName := PascalCaseName(svc.Name().String(), resource)
			resCommand := clitemplate.Command{
				Name:        cmdName,
				SubCommands: verbCommands,
			}
			clitemplate.CmdTemplate.Execute(cv.Out, resCommand)
		}

		// for each request type generate a flagset function
		for _, flagset := range messageFlagSets {
			clitemplate.FlagsetTemplate.Execute(cv.Out, flagset)
		}

	}
	return nil, nil
}

func MethodResourceVerb(logger pgs.DebuggerCommon, method pgs.Method) (resource string, verb string, ok bool) {
	// interpret the name of the RPC
	res, ver := SplitRPCName(logger, method.Name().LowerSnakeCase().String())

	// override with cli option if present
	resOK, err := method.Extension(clioption.E_Resource, &resource)
	if err != nil {
		panic(errors.Wrap(err, "error reading CLI resource option, should be a string"))
	}
	if !resOK {
		resource = res
	}

	// Read the verb extension field
	verbOK, err := method.Extension(clioption.E_Verb, &verb)
	if err != nil {
		panic(errors.Wrap(err, "error reading cli verb option, should be a string"))
	}
	if !verbOK {
		verb = ver
	}

	// return true only if both resource and verb are non-empty
	ok = resource != "" && verb != ""
	return
}

func SplitRPCName(logger pgs.DebuggerCommon, name string) (string, string) {
	split := strings.Split(strings.ToLower(name), "_")
	if len(split) != 2 && len(split) != 3 {
		return "", ""
	}

	// Assume name is of the form VERB_RESOURCE(_LIST)
	// If VERB = GET and LIST is present, assume the actual verb is LIST
	verb := split[0]
	resource := inflection.Singular(split[1])
	if len(split) == 3 && verb == "get" && split[2] == "list" {
		verb = "list"
	}
	return verb, resource
}

func AddResourceVerb(resources map[string]map[string]string, resource, verb, rpc string) map[string]map[string]string {
	var verbs map[string]string
	if vs, ok := resources[resource]; ok {
		verbs = vs
	} else {
		verbs = map[string]string{}
	}
	if existing, ok := verbs[verb]; ok {
		panic(fmt.Errorf("verb clash. resource %s, verbs %s %s", resource, verb, existing))
	}
	verbs[verb] = rpc
	resources[resource] = verbs
	return resources
}

func PascalCaseName(parts ...string) string {
	builder := strings.Builder{}
	for _, part := range parts {
		builder.WriteString(strings.Title(part))
	}
	return builder.String()
}
