// Package main creates multiple svg images from one template
// and  combines them into one image.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

var (
	templateFlag = flag.String("template", "template.svg", "Filename to use as a template.")
	heightFlag   = flag.Int("height", 64, "Height of output png in pixels")
	charsFlag    = flag.String("chars", "A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z", "Comma separated list of characters to generate.")
)

var findTx = regexp.MustCompile(`>A</tspan>`)

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

func replaceCopyTemplate(tempdir, template, char string) (tmpFilename string, err error) {
	bytes, err := ioutil.ReadFile(template)
	bytes = findTx.ReplaceAll(bytes, []byte(">"+char+"</tspan>"))

	basename := path.Base(template)
	tmpFilename = path.Join(tempdir, basename)

	err = ioutil.WriteFile(tmpFilename, bytes, 0644)
	return tmpFilename, err
}

func convertAll(tmpDir, template string, chars []string, height int) error {
	outfilenames := make([]string, 0, len(chars))
	for _, char := range chars {
		tmpFilename, err := replaceCopyTemplate(tmpDir, template, char)
		if err != nil {
			return err
		}
		tail := fmt.Sprintf("-%s-out.png", char)
		outfilename := path.Join(tmpDir, path.Base(strings.Replace(template, ".svg", tail, -1)))
		if err := convert(tmpFilename, outfilename, height); err != nil {
			return err
		}
		outfilenames = append(outfilenames, outfilename)
	}
	images, err := readAllImages(outfilenames)
	if err != nil {
		return err
	}
	outputFilename := path.Join(tmpDir, "output.png")

	if err = combineImagesHorizontally(images, outputFilename); err != nil {
		return err
	}
	fmt.Printf("Output in %s\n", outputFilename)
	return nil
}

func readAllImages(filenames []string) (images []image.Image, err error) {
	for _, filename := range filenames {
		reader, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		image, err := png.Decode(reader)
		reader.Close()
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}

func combineImagesHorizontally(images []image.Image, outputFilename string) error {
	width, height := 0, 0
	for _, image := range images {
		b := image.Bounds()
		width += b.Dx()
		if b.Dy() > height {
			height = b.Dy()
		}
	}
	m := image.NewRGBA(image.Rect(0, 0, width, height))

	x := 0
	for _, curImage := range images {
		b := curImage.Bounds()
		r := image.Rect(x, 0, x+b.Dx(), b.Dy())
		draw.Draw(m, r, curImage, image.ZP, draw.Src)
		x += b.Dx()
	}
	writer, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer writer.Close()
	return png.Encode(writer, m)
}

func main() {
	fmt.Printf("Converting template %q\n", *templateFlag)
	chars := strings.Split(*charsFlag, ",")
	tmpDir, err := ioutil.TempDir("", "svgdigits")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create temporary dir: %v\n", err)
		return
	}
	fmt.Printf("Output to folder %q\n", tmpDir)
	if err := convertAll(tmpDir, *templateFlag, chars, *heightFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create png: %v\n", err)
	}
}
