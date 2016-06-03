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
		p := yaml.NewParser([]byte(file))
		// Root node of a file's tree
		node := p.Parse()
		// List of nodes representing tokens.  Remove the first because YAML
		// always starts with an empty token, and begins any sequence with
		// an empty token (both considered starting at byte 0). If we don't remove
		// it, we will get a duplicate ref key make failure for every file.
		tokenList := yaml.Explore(file, node)[1:]

		for _, tok := range tokenList {
			extension := filepath.Ext(currentFile)
			defUnit := currentFile[0 : len(currentFile)-len(extension)]
			out.Refs = append(out.Refs, &graph.Ref{
				DefUnitType: "URL",
				DefUnit:     defUnit,
				DefPath:     filepath.Dir(currentFile) + "/" + tok.Value,
				Unit:        u.Name,
				File:        filepath.ToSlash(currentFile),
				Start:       uint32(tok.StartByte),
				End:         uint32(tok.EndByte),
				Def:         true,
			})
		}
	}
	return &out, nil
}
