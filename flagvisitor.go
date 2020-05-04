package main

import (
	"strings"

	"github.com/anzboi/protoc-gen-cli/clitemplate"
	pgs "github.com/lyft/protoc-gen-star"
)

type FlagsetVisitor struct {
	pgs.Visitor
	pgs.DebuggerCommon
	flags map[string]clitemplate.PFlag
}

func NewFlagsetVisitor(parent *FlagsetVisitor) *FlagsetVisitor {
	return &FlagsetVisitor{
		Visitor:        parent.Visitor,
		DebuggerCommon: parent.DebuggerCommon,
		flags:          map[string]clitemplate.PFlag{},
	}
}

func (m *FlagsetVisitor) VisitMessage(msg pgs.Message) (v pgs.Visitor, err error) {
	return m, nil
}

func (m *FlagsetVisitor) VisitField(field pgs.Field) (pgs.Visitor, error) {
	t := field.Type()
	fieldName := strings.ToLower(field.Name().String())

	// repeated
	if t.IsRepeated() {
		valType := t.Element()
		if !valType.IsEmbed() {
			if valType.IsEnum() {
				m.flags[fieldName] = clitemplate.NewEnumFlag(true)
			} else {
				m.flags[fieldName] = clitemplate.NewPFlag(valType.ProtoType(), true)
			}
		}
		return nil, nil
	}

	// enums
	if t.IsEnum() {
		m.flags[fieldName] = clitemplate.NewEnumFlag(false)
		return nil, nil
	}

	// maps
	if t.IsMap() {
		if t.Key().ProtoType().String() == "string" && !t.Element().IsEmbed() {
			m.flags[fieldName] = clitemplate.NewStringMapFlag()
		}
		return nil, nil
	}

	// embedded messages
	if t.IsEmbed() {
		child := NewFlagsetVisitor(m)
		if err := pgs.Walk(child, t.Embed()); err != nil {
			return nil, err
		}

		for childFlag, flagType := range child.flags {
			m.flags[strings.ToLower(fieldName+"."+childFlag)] = flagType
		}

		return nil, nil
	}

	// primitives
	m.flags[strings.ToLower(field.Name().String())] = clitemplate.NewPFlag(t.ProtoType(), false)
	return nil, nil
}
