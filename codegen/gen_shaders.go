package main

import (
	"io/ioutil"
	"os"
	"text/template"
	"time"
)

func main() {
	args := os.Args[1:]
	path := args[0]
	name := args[1]

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create("gen/" + name + ".go")
	if err != nil {
		panic(err)
	}
	t := template.Must(template.New("shader").Parse(shaderTemplate))
	t.Execute(outFile, struct {
		Timestamp time.Time
		Name      string
		Path      string
		Content   string
	}{
		Timestamp: time.Now(),
		Name:      name,
		Path:      path,
		Content:   string(contents),
	})
}

const shaderTemplate = `// This shader file was generated using go:generate go run codegen/gen_shaders.go {{.Path}} {{.Name}}
// {{.Timestamp}}"
package shaders
` + "const {{.Name}} = `{{.Content}}`+\"\\x00\""
