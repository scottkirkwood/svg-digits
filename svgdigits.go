// Package main creates multiple svg images from one template
// and  combines them into one image.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	templateFlag = flag.String("template", "template.svg", "Filename to use as a template.")
	heightFlag   = flag.Int("height", 64, "Height of output png in pixels")
	charsFlag    = flag.String("chars", "A,B,C", "Comma separated list of characters to generate.")
)

func convert(infile, outfile string, height int) error {
	args := []string{
		infile,
		"--export-png=" + outfile,
		fmt.Sprintf("--export-height=%d", height),
	}
	cmd := exec.Command("inkscape", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed args: %q, %v, %s", args, err, out)
	}
	return nil
}

func convertAll(template string, chars []string, height int) error {
	outfile := strings.Replace(*templateFlag, ".svg", "-out.png", -1)
	if err := convert(*inFileFlag, outfile, *heightFlag); err != nil {
		return err
	}
	return nil
}
func main() {
	fmt.Printf("Converting template %q\n", *templateFlag)
	chars := strings.Split(*charsFlag, ",")
	if err := convertAll(*templateFlag, chars, *heightFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create png: %v\n", err)
	}
}
