// +build unidev

// Parse character metrics from an AFM file to convert into a static go code declaration.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"flag"

	pdfcommon "github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func main() {
	filepath := flag.String("file", "", "AFM input file")
	method := flag.String("method", "charmetrics", "charmetrics/charcodes/glyph-to-charcode")

	flag.Parse()

	if len(*filepath) == 0 {
		fmt.Println("Please specify an input file.  Run with -h to get options.")
		return
	}

	var err error
	switch *method {
	case "charmetrics":
		err = runCharmetricsOnFile(*filepath)
	case "charcodes":
		err = runCharcodeToGlyphRetrievalOnFile(*filepath)
	case "glyph-to-charcode":
		err = runGlyphToCharcodeRetrievalOnFile(*filepath)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// --charmetrics to get char metric data.
	// --charcodes to get charcode to glyph data

}

// Generate a glyph to charmetrics (width and height) map.
func runCharmetricsOnFile(path string) error {
	metrics, err := GetCharmetricsFromAfmFile(path)
	if err != nil {
		return err
	}

	keys := []string{}
	for key := range metrics {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	fmt.Printf("var xxfontCharMetrics map[string]CharMetrics = map[string]CharMetrics{\n")
	for _, key := range keys {
		metric := metrics[key]
		fmt.Printf("\t\"%s\":\t{GlyphName:\"%s\", Wx:%f, Wy:%f},\n", key, metric.GlyphName, metric.Wx, metric.Wy)
	}
	fmt.Printf("}\n")
	return nil
}

func runCharcodeToGlyphRetrievalOnFile(afmpath string) error {
	charcodeToGlyphMap, err := GetCharcodeToGlyphEncodingFromAfmFile(afmpath)
	if err != nil {
		return err
	}

	keys := []int{}
	for key := range charcodeToGlyphMap {
		keys = append(keys, int(key))
	}
	sort.Ints(keys)

	fmt.Printf("var xxfontCharcodeToGlyphMap map[byte]string =  map[byte]string{\n")
	for _, key := range keys {
		fmt.Printf("\t%d: \"%s\",\n", key, charcodeToGlyphMap[byte(key)])
	}
	fmt.Printf("}\n")

	return nil
}

func runGlyphToCharcodeRetrievalOnFile(afmpath string) error {
	charcodeToGlyphMap, err := GetCharcodeToGlyphEncodingFromAfmFile(afmpath)
	if err != nil {
		return err
	}

	keys := []int{}
	for key := range charcodeToGlyphMap {
		keys = append(keys, int(key))
	}
	sort.Ints(keys)

	fmt.Printf("var xxfontGlyphToCharcodeMap map[string]byte =  map[string]byte ={\n")
	for _, key := range keys {
		fmt.Printf("\t\"%s\":\t%d,\n", charcodeToGlyphMap[byte(key)], key)
	}
	fmt.Printf("}\n")

	return nil
}

func GetCharmetricsFromAfmFile(filename string) (map[string]fonts.CharMetrics, error) {
	glyphMetricsMap := map[string]fonts.CharMetrics{}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	readingCharMetrics := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, " ")
		if len(parts) < 1 {
			continue
		}
		if !readingCharMetrics && parts[0] == "StartCharMetrics" {
			readingCharMetrics = true
			continue
		}
		if readingCharMetrics && parts[0] == "EndCharMetrics" {
			break
		}
		if !readingCharMetrics {
			continue
		}
		if parts[0] != "C" {
			continue
		}

		parts = strings.Split(line, ";")
		metrics := fonts.CharMetrics{}
		metrics.GlyphName = ""
		for _, part := range parts {
			cmd := strings.TrimSpace(part)
			if len(cmd) == 0 {
				continue
			}
			args := strings.Split(cmd, " ")
			if len(args) < 1 {
				continue
			}

			switch args[0] {
			case "N":
				if len(args) != 2 {
					pdfcommon.Log.Debug("Failed C line: ", line)
					return nil, errors.New("Invalid C line")
				}
				metrics.GlyphName = strings.TrimSpace(args[1])
			case "WX":
				if len(args) != 2 {
					pdfcommon.Log.Debug("WX: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("Invalid range")
				}
				wx, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wx = wx
			case "WY":
				if len(args) != 2 {
					pdfcommon.Log.Debug("WY: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("Invalid range")
				}
				wy, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wy = wy
			case "W":
				if len(args) != 2 {
					pdfcommon.Log.Debug("W: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("Invalid range")
				}
				w, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wy = w
				metrics.Wx = w
			}
		}

		if len(metrics.GlyphName) > 0 {
			glyphMetricsMap[metrics.GlyphName] = metrics
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return glyphMetricsMap, nil
}

func GetCharcodeToGlyphEncodingFromAfmFile(filename string) (map[byte]string, error) {
	charcodeToGlypMap := map[byte]string{}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	readingCharMetrics := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, " ")
		if len(parts) < 1 {
			continue
		}
		if !readingCharMetrics && parts[0] == "StartCharMetrics" {
			readingCharMetrics = true
			continue
		}
		if readingCharMetrics && parts[0] == "EndCharMetrics" {
			break
		}
		if !readingCharMetrics {
			continue
		}
		if parts[0] != "C" {
			continue
		}

		parts = strings.Split(line, ";")
		var charcode int64
		var glyph string

		for _, part := range parts {
			cmd := strings.TrimSpace(part)
			if len(cmd) == 0 {
				continue
			}
			args := strings.Split(cmd, " ")
			if len(args) < 1 {
				continue
			}

			switch args[0] {
			case "C":
				if len(args) != 2 {
					pdfcommon.Log.Debug("Failed C line: %s", line)
					return nil, errors.New("Invalid C line")
				}
				charcode, err = strconv.ParseInt(strings.TrimSpace(args[1]), 10, 64)
				if err != nil {
					return nil, err
				}
			case "N":
				if len(args) != 2 {
					pdfcommon.Log.Debug("Failed C line: %s", line)
					return nil, errors.New("Invalid C line")
				}

				glyph = strings.TrimSpace(args[1])
				if charcode >= 0 && charcode <= 255 {
					charcodeToGlypMap[byte(charcode)] = glyph
				} else {
					fmt.Printf("NOT included: %d -> %s\n", charcode, glyph)
				}
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return charcodeToGlypMap, nil
}
