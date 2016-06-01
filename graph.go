package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/attfarhan/yaml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

type GraphCmd struct{}

var graphCmd GraphCmd

func init() {
	_, err := parser.AddCommand("graph",
		"graph for YAML files",
		"Graph the directory tree rooted at the current directory for YAML files.",
		&graphCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *GraphCmd) Execute(args []string) error {
	inputBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var units unit.SourceUnits
	if err := json.NewDecoder(bytes.NewReader(inputBytes)).Decode(&units); err != nil {
		var u *unit.SourceUnit
		if err := json.NewDecoder(bytes.NewReader(inputBytes)).Decode(&u); err != nil {
			return err
		}
		units = unit.SourceUnits{u}
	}
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	if len(units) == 0 {
		log.Fatal("input contains no source unit data.")
	}
	out, err := Graph(units)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		return err
	}
	return nil
}

type T struct {
	value  []string
	line   []int
	column []int
}

// graph.Output is a struct with fields:
// 				Defs []*Def
//				Refs []*Ref
//				Docs []*Doc
//				Anns []*ann.Ann
func Graph(units unit.SourceUnits) (*graph.Output, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(fmt.Errorf("failed to read file: %s", r))
		}
	}()
	if len(units) > 1 {
		return nil, errors.New("unexpected multiple units")
	}
	u := units[0]
	// out is a graph.Output struct with a Ref field of pointers to graph.Ref
	out := graph.Output{Refs: []*graph.Ref{}}

	// Decode source unit
	// Get files
	// Iterate over files, parse YAML
	// For each token, get the byte ranges, token string, and add to Refs

	for _, currentFile := range u.Files {
		f, err := ioutil.ReadFile(currentFile)
		if err != nil {
			log.Printf("failed to read a source unit file: %s", err)
			continue
		}
		file := string(f)
		t := &T{}
		var x []*yaml.Node
		p := yaml.NewParser([]byte(file))
		node := p.Parse()
		tokenList := yaml.Explore(node, x)
		getLineAndColumn(tokenList, file, t)
		for i, _ := range t.value {
			start, end, value := findOffsets(file, t.line[i], t.column[i], t.value[i])
			extension := filepath.Ext(currentFile)
			defUnit := currentFile[0 : len(currentFile)-len(extension)]
			out.Refs = append(out.Refs, &graph.Ref{
				DefUnitType: "URL",
				DefUnit:     defUnit,
				DefPath:     filepath.Dir(currentFile) + "/" + value,
				Unit:        u.Name,
				File:        filepath.ToSlash(currentFile),
				Start:       uint32(start),
				End:         uint32(end),
				Def:         true,
			})
		}
	}
	return &out, nil
}

func getLineAndColumn(tokenList []*yaml.Node, fileString string, out *T) {
	for _, token := range tokenList {
		out.value = append(out.value, token.Value)
		out.line = append(out.line, token.Line)
		out.column = append(out.column, token.Column)
		// a, b := findOffsets(data, token.Line, token.Column, token.Value)
		// fmt.Println("start: ", a, "End: ", b)
		getLineAndColumn(token.Children, fileString, out)
	}
}
func findOffsets(fileText string, line, column int, token string) (start, end int, value string) {

	// we count our current line and column position.
	currentCol := 0
	currentLine := 0
	for offset, ch := range fileText {
		// see if we found where we wanted to go to.
		if currentLine == line && currentCol == column {
			end = offset + len([]byte(token))
			return offset, end, token
		}

		// line break - increment the line counter and reset the column.
		if ch == '\n' {
			currentLine++
			currentCol = 0
		} else {
			currentCol++
		}
	}
	return -1, -1, token // not found.
}

// package main

// import (
// 	"fmt"
// 	// "log"
// 	"github.com/attfarhan/yaml"
// 	// "github.com/shurcooL/go-goon"
// )

// var data = `language: go

// go:
//     - 1.4
//     - 1.5
//     - 1.6
//     - tip

// go_import_path: gopkg.in/yaml.v2`

// func getLineAndColumn(tokenList []*yaml.Node, fileString string) {
// 	for _, token := range tokenList {
// 		fmt.Println("line: ", token.Line)
// 		fmt.Println("column: ", token.Column)
// 		fmt.Println("value:", token.Value)
// 		fmt.Println(findOffsets(data, token.Line, token.Column, token.Value))
// 		getLineAndColumn(token.Children, data)
// 	}
// }

// // refs = append(refs, &graph.Ref{
// // 	DefUnitType: "URL",
// // 	DefUnit:     "MDN",
// // 	DefPath:     mdnDefPath(d.Property),
// // 	Unit:        u.Name,
// // 	File:        filepath.ToSlash(filePath),
// // 	Start:       uint32(s),
// // 	End:         uint32(e),
