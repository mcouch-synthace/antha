package composer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

type Config struct {
	ElementSources ElementSources `json:"ElementSources"`
	TmpDir         string
}

func ConfigFromReader(r io.Reader) (*Config, error) {
	c := &Config{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(c); err != nil {
		return nil, err
	} else {
		c.ElementSources.Sort()
		if len(c.TmpDir) == 0 {
			if c.TmpDir, err = ioutil.TempDir("", "antha-composer"); err != nil {
				return nil, err
			}
		}
		if err := os.MkdirAll(c.TmpDir, 0700); err != nil {
			return nil, err
		}
		return c, nil
	}
}

type Composer struct {
	Config   *Config
	Workflow *Workflow

	classes map[string]*ElementClass

	varCount uint64
	varMemo  map[string]string
}

type ElementClass struct {
	importPath    string
	directoryPath string
	packageName   string
	filesContent  map[string]string
}

func NewComposer(cfg *Config, workflow *Workflow) *Composer {
	return &Composer{
		Config:   cfg,
		Workflow: workflow,

		classes: make(map[string]*ElementClass),

		varMemo: make(map[string]string),
	}
}

func (c *Composer) Render(w io.Writer) error {
	for _, class := range c.Workflow.ElementClasses() {
		if elemSource, revision, pathTail, err := c.Config.ElementSources.Match(class); err != nil {
			return err
		} else if elemSource == nil {
			return fmt.Errorf("Unable to resolve component name: %s", class)
		} else if files, err := elemSource.FetchFiles(revision, pathTail); err != nil {
			return err
		} else {
			c.classes[class] = &ElementClass{
				importPath:   path.Join(elemSource.Prefix, pathTail),
				filesContent: files,
				packageName:  path.Base(pathTail),
			}
		}
	}

	funcs := template.FuncMap{
		"varName":      c.varName,
		"importPaths":  c.importPaths,
		"packageName":  c.packageName,
		"elementPaths": c.elementPaths,
	}
	if t, err := template.New("main").Funcs(funcs).Parse(tpl); err != nil {
		return err
	} else {
		return t.Execute(w, c.Workflow)
	}
}

func (c *Composer) varName(name string) string {
	if res, found := c.varMemo[name]; found {
		return res
	}

	res := make([]rune, 0, len(name))
	ensureUpper := false
	for _, r := range []rune(name) {
		switch {
		case 'a' <= r && r <= 'z' && ensureUpper:
			ensureUpper = false
			res = append(res, unicode.ToUpper(r))
		case 'a' <= r && r <= 'z':
			res = append(res, r)
		case 'A' <= r && r <= 'Z' && len(res) == 0:
			res = append(res, unicode.ToLower(r))
		case 'A' <= r && r <= 'Z':
			res = append(res, r)
			ensureUpper = false
		case strings.ContainsRune(" -_", r):
			ensureUpper = true
		}
	}
	resStr := fmt.Sprintf("%s%d", string(res), c.varCount)
	c.varCount++
	c.varMemo[name] = resStr
	return resStr
}

func (c *Composer) importPaths() []string {
	res := make([]string, 0, len(c.classes))
	for _, ec := range c.classes {
		res = append(res, ec.importPath)
	}
	sort.Strings(res)
	return res
}

func (c *Composer) packageName(name string) string {
	return path.Base(name)
}

func (c *Composer) elementPaths() map[string]string {
	res := make(map[string]string, len(c.classes))
	for _, ec := range c.classes {
		res[ec.importPath] = ec.packageName
	}
	return res
}

var tpl = `// Code generated by antha composer. DO NOT EDIT.
package main

import (
	"log"

	"github.com/antha-lang/antha/laboratory"

{{range importPaths}}	{{printf "%q" .}}
{{end}})

func main() {
	lab := laboratory.NewLaboratory()
	// Register line maps for the elements we're using
{{range $path,$packageName := elementPaths}}	lab.RegisterLineMap({{printf "%q" $path}}, {{$packageName}}.LineMap)
{{end}}
	// Create the elements
{{range $name, $proc := .Processes}}	{{varName $name}} := {{packageName $proc.Component}}.New(lab)
{{end}}
	// Add wiring
{{range .Connections}}	lab.AddLink({{varName .Source.Process}}, {{varName .Target.Process}}, func () { {{varName .Target.Process}}.Inputs.{{.Target.Port}} = {{varName .Source.Process}}.Outputs.{{.Source.Port}} })
{{end}}
	// Set parameters
{{range $name, $params := .Parameters}}{{range $param, $value := $params}}	if err := {{varName $name}}.Inputs.{{$param}}.SetFromJSON([]byte({{printf "%q" $value}})); err != nil {
		log.Fatal(err)
	}
{{end}}{{end}}
	// Run!
	if err := lab.Run(); err != nil {
		log.Fatal(err)
	}
}
`
