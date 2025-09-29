package tmplinspect

import (
	"html/template"
	"maps"
	"slices"
	"strings"
	"text/template/parse"
)

func Inspect(tmpl string) (fields []string, funcs []string, _ error) {
	t := parse.New("")
	t.Mode = parse.SkipFuncCheck

	t, err := t.Parse(tmpl, "", "", make(map[string]*parse.Tree))
	if err != nil {
		return nil, nil, err
	}

	return InspectTree(t)
}

func InspectTemplate(t *template.Template) (fields []string, funcs []string, err error) {
	return InspectTree(t.Tree)
}

func InspectTree(t *parse.Tree) (fields []string, funcs []string, err error) {
	// 1. Prepare maps for storing results.
	fieldMap := make(map[string]struct{})
	funcMap := make(map[string]struct{})

	// 2. Walk the AST from the template's root node.
	walk(t.Root, fieldMap, funcMap, nil)

	// 3. Convert maps to sorted slices for consistent output.
	fields = slices.Sorted(maps.Keys(fieldMap))
	funcs = slices.Sorted(maps.Keys(funcMap))

	return fields, funcs, nil
}

// walk recursively traverses the template's AST, populating the field and function maps.
// It carries a prefix to handle nested contexts like in `range` and `with` blocks.
func walk(node parse.Node, fieldMap, funcMap map[string]struct{}, prefix []string) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parse.ActionNode:
		walk(n.Pipe, fieldMap, funcMap, prefix)

	case *parse.CommandNode:
		if len(n.Args) > 0 {
			if identifier, ok := n.Args[0].(*parse.IdentifierNode); ok {
				funcMap[identifier.Ident] = struct{}{}
			}
		}
		for _, arg := range n.Args {
			walk(arg, fieldMap, funcMap, prefix)
		}

	case *parse.FieldNode:
		fullPath := make([]string, len(prefix), len(prefix)+len(n.Ident))
		copy(fullPath, prefix)
		fullPath = append(fullPath, n.Ident...)

		if len(fullPath) > 0 {
			field := strings.Join(fullPath, ".")
			fieldMap[field] = struct{}{}
		}

	case *parse.PipeNode:
		for _, cmd := range n.Cmds {
			walk(cmd, fieldMap, funcMap, prefix)
		}

	case *parse.ListNode:
		if n != nil {
			for _, subNode := range n.Nodes {
				walk(subNode, fieldMap, funcMap, prefix)
			}
		}

	case *parse.IfNode:
		walk(n.Pipe, fieldMap, funcMap, prefix)
		walk(n.List, fieldMap, funcMap, prefix)
		walk(n.ElseList, fieldMap, funcMap, prefix)

	case *parse.RangeNode:
		walk(n.Pipe, fieldMap, funcMap, prefix)
		newPrefixPart := extractFieldPath(n.Pipe)
		newPrefix := append(prefix, newPrefixPart...)
		walk(n.List, fieldMap, funcMap, newPrefix)
		walk(n.ElseList, fieldMap, funcMap, prefix)

	case *parse.WithNode:
		walk(n.Pipe, fieldMap, funcMap, prefix)
		newPrefixPart := extractFieldPath(n.Pipe)
		newPrefix := append(prefix, newPrefixPart...)
		walk(n.List, fieldMap, funcMap, newPrefix)
		walk(n.ElseList, fieldMap, funcMap, prefix)
	}
}

// extractFieldPath tries to find the field being used in a range/with pipe.
func extractFieldPath(pipe parse.Node) []string {
	if pipe == nil {
		return nil
	}

	// A range/with pipe is usually a PipeNode with one command.
	if pipeNode, ok := pipe.(*parse.PipeNode); ok {
		if len(pipeNode.Cmds) > 0 {
			cmd := pipeNode.Cmds[0]
			// The command's first argument is what we're ranging over.
			if len(cmd.Args) > 0 {
				if fieldNode, ok := cmd.Args[0].(*parse.FieldNode); ok {
					return fieldNode.Ident
				}
			}
		}
	}

	return nil
}
