// +build unidev

package main

// Utility to generate static maps of glyph <-> character codes for text encoding.

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {

	encodingfile := flag.String("encodingfile", "", "Encoding glyph list file")
	method := flag.String("method", "charcode-to-glyph", "charcode-to-glyph/glyph-to-charcode")

	flag.Parse()

	if len(*encodingfile) == 0 {
		fmt.Printf("Please specify an encoding file, see -h for options\n")
		os.Exit(1)
	}

	var err error
	switch *method {
	case "charcode-to-glyph":
		err = charcodeToGlyphListPath(*encodingfile)
	case "glyph-to-charcode":
		err = glyphToCharcodeListPath(*encodingfile)
	default:
		fmt.Printf("Unsupported method, see -h for options\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func charcodeToGlyphListPath(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	index := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.Trim(line, " \r\n")

		parts := strings.Split(line, " ")
		for _, part := range parts {
			index++
			if part == "notdef" {
				continue
			}
			fmt.Printf("\t%d: \"%s\",\n", index, part)
		}
	}

	return nil
}

func glyphToCharcodeListPath(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	index := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.Trim(line, " \r\n")

		parts := strings.Split(line, " ")
		for _, part := range parts {
			index++
			if part == "notdef" {
				continue
			}
			fmt.Printf("\t\"%s\": %d,\n", part, index)
		}
	}

	return nil
}
