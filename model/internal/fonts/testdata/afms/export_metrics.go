// +build unidev

// Parse character metrics from an AFM file to convert into a static go code declaration.

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"flag"

	pdfcommon "github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/internal/fonts"
	"github.com/unidoc/unipdf/v3/model"
)

func main() {
	filepath := flag.String("file", "", "AFM input file")
	method := flag.String("method", "charmetrics", "charmetrics/charcodes/glyph-to-charcode")

	flag.Parse()

	if len(*filepath) == 0 && *method != "descriptor" {
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
	case "descriptor":
		err = runFontDescriptorOnFiles()
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
		fmt.Printf("\t\"%s\":\t{Wx:%f, Wy:%f},\n", key, metric.Wx, metric.Wy)
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

func GetCharmetricsFromAfmFile(filename string) (map[string]model.CharMetrics, error) {
	glyphMetricsMap := map[string]model.CharMetrics{}

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
		var metrics model.CharMetrics
		glyphName := ""
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
					return nil, errors.New("invalid C line")
				}
				glyphName = strings.TrimSpace(args[1])
			case "WX":
				if len(args) != 2 {
					pdfcommon.Log.Debug("WX: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("invalid range")
				}
				wx, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wx = wx
			case "WY":
				if len(args) != 2 {
					pdfcommon.Log.Debug("WY: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("invalid range")
				}
				wy, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wy = wy
			case "W":
				if len(args) != 2 {
					pdfcommon.Log.Debug("W: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("invalid range")
				}
				w, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wy = w
				metrics.Wx = w
			}
		}

		if len(glyphName) > 0 {
			glyphMetricsMap[glyphName] = metrics
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
					return nil, errors.New("invalid C line")
				}
				charcode, err = strconv.ParseInt(strings.TrimSpace(args[1]), 10, 64)
				if err != nil {
					return nil, err
				}
			case "N":
				if len(args) != 2 {
					pdfcommon.Log.Debug("Failed C line: %s", line)
					return nil, errors.New("invalid C line")
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

func runFontDescriptorOnFiles() error {
	parts := []string{}
	for _, name := range standard14Names {
		path := filepath.Join(".", fmt.Sprintf("%s.afm", name))
		text, err := createFontDescriptor(name, path)
		if err != nil {
			return err
		}
		parts = append(parts, text)
	}
	fmt.Println("\n// =========================================================\n")
	fmt.Println(" Standard14Descriptors = map[string]DescriptorLiteral {")
	for _, text := range parts {
		fmt.Printf("%s,\n", text)
	}
	fmt.Println("}")
	return nil
}

var standard14Names = []string{
	"Courier",
	"Courier-Bold",
	"Courier-BoldOblique",
	"Courier-Oblique",
	"Helvetica",
	"Helvetica-Bold",
	"Helvetica-BoldOblique",
	"Helvetica-Oblique",
	"Times-Roman",
	"Times-Bold",
	"Times-BoldItalic",
	"Times-Italic",
	"Symbol",
	"ZapfDingbats",
}

// createFontDescriptor creates a literal DescriptorLiteral that can be copied into
// pdf/internal/fonts/const.go.
func createFontDescriptor(name, filename string) (string, error) {
	fmt.Println("----------------------------------------------------")
	fmt.Printf("createFontDescriptor: name=%#q filename=%q\n", name, filename)

	nameVal, err := parseAfmFile(filename)
	if err != nil {
		return "", err
	}
	descriptor, err := parseNameVal(nameVal)
	if err != nil {
		fmt.Printf("ERROR: createFontDescriptor failed. name=%q err=%v\n", name, err)
		return "", err
	}
	return writeDescriptorLiteral(name, descriptor), nil
}

func parseAfmFile(filename string) (map[string]string, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	nameVal := map[string]string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		k, v, ok := asKeyValue(line)
		if !ok {
			continue
		}
		if k == "StartCharMetrics" {
			break
		}
		nameVal[k] = v
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nameVal, nil
}

// firstWord returns the first word in `line`.
func asKeyValue(line string) (string, string, bool) {
	line = strings.Trim(line, "\r\n\t ")
	i := strings.IndexByte(line, ' ')
	if i < 0 {
		return "", "", false
	}
	return line[:i], line[i+1:], true
}

func parseNameVal(nameVal map[string]string) (fonts.DescriptorLiteral, error) {
	fmt.Printf("nameVal=%q\n", nameVal)
	descriptor := fonts.DescriptorLiteral{}
	descriptor.FontName = nameVal["FontName"]
	descriptor.FontFamily = nameVal["FamilyName"]
	weight, ok := textWeight[nameVal["Weight"]]
	if !ok {
		return descriptor, fmt.Errorf("Invalid Weight=%#q. nameVal=%q",
			nameVal["Weight"], nameVal)
	}
	descriptor.FontWeight = weight
	symbolic := descriptor.FontFamily == "Symbol" || descriptor.FontFamily == "ZapfDingbats"

	descriptor.FontBBox = toBBox(nameVal["FontBBox"])
	descriptor.ItalicAngle = toFloat(nameVal, false, "ItalicAngle")
	descriptor.CapHeight = toFloat(nameVal, symbolic, "CapHeight")
	descriptor.XHeight = toFloat(nameVal, symbolic, "XHeight")
	descriptor.Ascent = toFloat(nameVal, symbolic, "Ascender")
	descriptor.Descent = toFloat(nameVal, symbolic, "Descender")
	descriptor.StemH = toFloat(nameVal, symbolic, "StdHW")
	descriptor.StemV = toFloat(nameVal, symbolic, "StdVW")

	fixedPitch := toBool(nameVal, "IsFixedPitch")
	italic := descriptor.ItalicAngle != 0
	descriptor.Flags = getFlags(descriptor.FontName, fixedPitch, symbolic, italic)

	return descriptor, nil
}

func getFlags(name string, fixedPitch, symbolic, italic bool) uint {
	flags := uint(0)
	if fixedPitch {
		flags |= fontFlagFixedPitch
	}
	if symbolic {
		flags |= fontFlagSymbolic
	} else {
		flags |= fontFlagNonsymbolic
	}
	if italic {
		flags |= fontFlagItalic
	}
	return flags
}

// These consts are copied from model/font.go
// 9.8.2 Font Descriptor Flags (page 283)
const (
	fontFlagFixedPitch  = 0x00001
	fontFlagSerif       = 0x00002
	fontFlagSymbolic    = 0x00004
	fontFlagScript      = 0x00008
	fontFlagNonsymbolic = 0x00020
	fontFlagItalic      = 0x00040
	fontFlagAllCap      = 0x10000
	fontFlagSmallCap    = 0x20000
	fontFlagForceBold   = 0x40000
)

func toInt(keyVal map[string]string, key string) uint {
	text, ok := keyVal[key]
	if !ok {
		panic(fmt.Sprintf("No %#q key", key))
	}
	x, err := strconv.Atoi(text)
	if err != nil {
		panic(err)
	}
	return uint(x)
}

func toBool(keyVal map[string]string, key string) bool {
	text, ok := keyVal[key]
	if !ok {
		panic(fmt.Sprintf("No %#q key", key))
	}
	switch text {
	case "true":
		return true
	case "false":
		return false
	default:
		panic(fmt.Sprintf("Bad bool val %q. key=%#q", text, key))
	}
}

func toFloat(keyVal map[string]string, optional bool, key string) float64 {
	text, ok := keyVal[key]
	if !ok {
		// Symbol and ZapfDingbats don't have XHeight and key == "CapHeight" entries
		if optional {
			return 0.0
		}
		panic(fmt.Sprintf("No %#q key", key))
	}
	x, err := strconv.ParseFloat(text, 64)
	if err != nil {
		panic(err)
	}
	return x
}

func toBBox(text string) [4]float64 {
	fmt.Printf("toBBox: text=%q\n", text)
	parts := strings.Split(text, " ")
	vals := [4]float64{}
	for i, p := range parts {
		x, err := strconv.ParseFloat(p, 64)
		if err != nil {
			panic(err)
		}
		vals[i] = x
	}
	return vals
}

var textWeight = map[string]fonts.FontWeight{
	"Medium": fonts.FontWeightMedium,
	"Bold":   fonts.FontWeightBold,
	"Roman":  fonts.FontWeightRoman,
}

var weightVarname = map[fonts.FontWeight]string{
	fonts.FontWeightMedium: "FontWeightMedium",
	fonts.FontWeightBold:   "FontWeightBold",
	fonts.FontWeightRoman:  "FontWeightRoman",
}

func writeDescriptorLiteral(name string, descriptor fonts.DescriptorLiteral) string {

	var b bytes.Buffer

	fmt.Fprintf(&b, "%q: DescriptorLiteral {\n", name)
	fmt.Fprintf(&b, "\tFontName: %q,\n", descriptor.FontName)
	fmt.Fprintf(&b, "\tFontFamily: %q,\n", descriptor.FontFamily)
	fmt.Fprintf(&b, "\tFontWeight: %s,\n", weightVarname[descriptor.FontWeight])
	fmt.Fprintf(&b, "\tFlags: 0x%04x,\n", descriptor.Flags)
	fmt.Fprintf(&b, "\tFontBBox: %#v,\n", descriptor.FontBBox)
	fmt.Fprintf(&b, "\tItalicAngle: %g,\n", descriptor.ItalicAngle)
	fmt.Fprintf(&b, "\tAscent: %g,\n", descriptor.Ascent)
	fmt.Fprintf(&b, "\tDescent: %g,\n", descriptor.Descent)
	fmt.Fprintf(&b, "\tCapHeight: %g,\n", descriptor.CapHeight)
	fmt.Fprintf(&b, "\tXHeight: %g,\n", descriptor.XHeight)
	fmt.Fprintf(&b, "\tStemV: %g,\n", descriptor.StemV)
	fmt.Fprintf(&b, "\tStemH: %g,\n", descriptor.StemH)
	fmt.Fprintf(&b, "}")
	return b.String()
}
