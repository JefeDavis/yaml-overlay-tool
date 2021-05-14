// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Nodes      []*yaml.Node
	Origin     string
	Path       string
	outputPath string
}

func (src *Source) processOverlays(o []Overlay, nodeIndex int) error {
	for i := range o {
		if err := o[i].applyOverlay(src, nodeIndex); err != nil {
			return err
		}
	}

	return nil
}

func (src *Source) Save(o *Options, output string) error {
	fileName := path.Join(o.OutputDir, "yamlFiles", src.outputPath)
	dirName := path.Dir(fileName)

	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		if err := os.MkdirAll(dirName, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %s, %w", dirName, err)
		}
	}

	//nolint:gosec //output files with read and write permissions so that end-users can continue to leverage these files
	if err := ioutil.WriteFile(fileName, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to save file %s: %w", fileName, err)
	}

	return nil
}

func (src *Source) doPostProcessing(o *Options) error {
	output := new(bytes.Buffer)
	ye := yaml.NewEncoder(output)
	ye.SetIndent(o.Indent)

	for _, node := range src.Nodes {
		err := ye.Encode(node)
		if err != nil {
			return fmt.Errorf("unable to marshal final document %s, error: %w", src.Path, err)
		}
	}

	final := fmt.Sprintf("---\n%s\n", output)

	log.Noticef("Final: >>>\n%s\n", output)
	// added so we can quickly see the results of the run
	if o.StdOut {
		fmt.Print(final) //nolint:forbidigo

		return nil
	}

	if err := src.Save(o, final); err != nil {
		return fmt.Errorf("failed to save, %w", err)
	}

	return nil
}
