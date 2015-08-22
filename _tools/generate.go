package main

import (
	"log"
	"os"
	"strings"
	"text/template"

	"go/token"
	"go/types"
)

var tmpl = template.Must(template.New("bin").Funcs(
	template.FuncMap{
		"ucfirst": func(s string) string {
			return strings.ToUpper(s[0:1]) + s[1:]
		},
	}).
	Parse(`// auto-generated by _tools/generate.go

package genericop

import (
	"fmt"
)
{{range .binOpDefs}}{{$op := .Token.String}}
// {{.Name}} is a generic operator function for "{{$op}}".
func {{.Name}}(x, y interface{}) (r {{if .ReturnBool}}bool{{else}}interface{}{{end}}, err error) {
	{{range .TypeNames}}if x_, ok := x.({{.}}); ok {
		if y_, ok := y.({{.}}); ok {
			r = x_ {{$op}} y_
		} else {
			err = fmt.Errorf("incompatible types: %T %s %T", x, "{{$op}}", y)
		}
		return
	}
	{{end}}
	err = fmt.Errorf("no operation defined: %T %s %T", x, "{{$op}}", y)
	return
}
{{end}}
{{range .types}}
// Must{{.|ucfirst}} receives results from generic operators (e.g. Add)
// and panics if there was any error or the value was not of type {{.|ucfirst}}.
func Must{{.|ucfirst}}(v interface{}, err error) {{.}} {
	if err != nil {
		panic(err)
	}
	return v.({{.}})
}
{{end}}
`))

func main() {
	names := types.Universe.Names()

	allTypeNames := []string{}
	addableTypeNames := []string{}
	numericTypeNames := []string{}
	integerTypeNames := []string{}
	orderedTypeNames := []string{}

	for _, name := range names {
		obj, ok := types.Universe.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}

		ty, ok := obj.Type().(*types.Basic)
		if !ok {
			continue
		}

		allTypeNames = append(allTypeNames, ty.Name())

		if ty.Info()&types.IsInteger != 0 {
			integerTypeNames = append(integerTypeNames, ty.Name())
		}

		if ty.Info()&types.IsNumeric != 0 {
			numericTypeNames = append(numericTypeNames, ty.Name())
		}

		if ty.Info()&types.IsOrdered != 0 {
			orderedTypeNames = append(orderedTypeNames, ty.Name())
		}
	}

	addableTypeNames = append(numericTypeNames, "string")

	binOpDefs := []struct {
		Name       string
		Token      token.Token
		TypeNames  []string
		ReturnBool bool
	}{
		{"Add", token.ADD, addableTypeNames, false},
		{"Sub", token.SUB, numericTypeNames, false},
		{"Mul", token.MUL, numericTypeNames, false},
		{"Quo", token.QUO, numericTypeNames, false},
		{"Rem", token.REM, integerTypeNames, false},
		{"And", token.AND, integerTypeNames, false},
		{"Or", token.OR, integerTypeNames, false},
		{"Xor", token.XOR, integerTypeNames, false},
		{"Lt", token.LSS, orderedTypeNames, true},
		{"Gt", token.GTR, orderedTypeNames, true},
		{"Le", token.LEQ, orderedTypeNames, true},
		{"Ge", token.GEQ, orderedTypeNames, true},
	}

	f, err := os.Create("ops.go")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(f, map[string]interface{}{
		"binOpDefs": binOpDefs,
		"types":     allTypeNames,
	})
	if err != nil {
		log.Fatal(err)
	}
}
