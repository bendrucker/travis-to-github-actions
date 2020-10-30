package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	travisPath   = ".travis.yml"
	workflowPath = ".github/workflows/test.yml"
	readmePath   = "readme.md"
)

func main() {
	// b, err := ioutil.ReadFile(travisPath)
	// if err != nil {
	// 	panic(err)
	// }

	flags := flag.NewFlagSet("travis-to-github-actions", flag.PanicOnError)

	var versions []string
	flags.StringSliceVar(&versions, "versions", []string{"14"}, "the versions of node to test with")

	flags.Parse(os.Args[1:])

	workflow := map[string]interface{}{
		"name": "tests",
		"on":   []string{"push"},
		"jobs": map[string]interface{}{
			"build": map[string]interface{}{
				"runs-on": "ubuntu-latest",
				"strategy": map[string]interface{}{
					"matrix": map[string][]string{
						"node-version": versions,
					},
				},
				"steps": []map[string]interface{}{
					{"uses": "actions/checkout@v2"},
					{
						"name": "Use Node.js ${{ matrix.node-version }}",
						"uses": "actions/setup-node@v1",
						"with": map[string]string{"node-version": "${{ matrix.node-version }}"},
					},
					{"run": "npm install"},
					{"run": "npm test"},
				},
			},
		},
	}

	err := os.MkdirAll(filepath.Dir(workflowPath), 0700)
	if err != nil {
		panic(err)
	}

	workflowJSON, err := yaml.Marshal(workflow)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(workflowPath, workflowJSON, 0644); err != nil {
		panic(err)
	}

	readme, err := ioutil.ReadFile(readmePath)
	if err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	badge := template.New("badge")
	badge.Parse(`[![tests](https://github.com/{{.Author}}/{{.Repo}}/workflows/tests/badge.svg)](https://github.com/{{.Author}}/{{.Repo}}/actions?query=workflow%3Atests)`)

	var b bytes.Buffer
	err = badge.Execute(&b, map[string]interface{}{
		"Author": "bendrucker",
		"Repo":   filepath.Base(wd),
	})
	if err != nil {
		panic(err)
	}

	regex := regexp.MustCompile(`\[\!\[Build Status]\(https:\/\/travis-ci.org\/.*(?:\)\])\(.*(?:\))`)
	updated := regex.ReplaceAll(readme, b.Bytes())

	if err := ioutil.WriteFile(readmePath, updated, 0644); err != nil {
		panic(err)
	}

	os.Remove(travisPath)
}
