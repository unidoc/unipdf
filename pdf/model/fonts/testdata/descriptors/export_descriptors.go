// +build unidev

// Parse character metrics from a php into a static go code declaration.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func main() {
	parts := []string{}
	for _, nf := range nameFile {
		path := filepath.Join(".", nf.path)
		text, err := createFontDescriptor(nf.name, path)
		if err != nil {
			panic(err)
		}
		parts = append(parts, text)
	}
	fmt.Println("\n// =========================================================\n")
	fmt.Println(" Standard14Descriptors = map[string]DescriptorLiteral {")
	for _, text := range parts {
		fmt.Printf("%s,\n", text)
	}
	fmt.Println("}")
}

var nameFile = []struct{ name, path string }{
	{"Courier", "courier.php"},
	{"Courier-Bold", "courierb.php"},
	{"Courier-BoldOblique", "courierbi.php"},
	{"Courier-Oblique", "courieri.php"},
	{"Helvetica", "helvetica.php"},
	{"Helvetica-Bold", "helveticab.php"},
	{"Helvetica-BoldOblique", "helveticabi.php"},
	{"Helvetica-Oblique", "helveticai.php"},
	{"Times-Roman", "times.php"},
	{"Times-Bold", "timesb.php"},
	{"Times-BoldItalic", "timesbi.php"},
	{"Times-talic", "timesi.php"},
	{"Symbol", "symbol.php"},
	{"ZapfDingbats", "zapfdingbats.php"},
}

// Generate a glyph to charmetrics (width and height) map.
func createFontDescriptor(name, path string) (string, error) {
	fmt.Println("----------------------------------------------------")
	fmt.Printf("createFontDescriptor: name=%#q path=%q\n", name, path)
	descriptor, err := getDescriptor(path)
	if err != nil {
		return "", err
	}

	return writeDescriptorLiteral(name, descriptor), nil
}

func getDescriptor(filename string) (fonts.DescriptorLiteral, error) {
	descriptor := fonts.DescriptorLiteral{}

	f, err := os.Open(filename)
	if err != nil {
		return descriptor, err
	}
	defer f.Close()

	nameVal := map[string]string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "$") {
			continue
		}
		updateDescMap(nameVal, line)
	}

	fmt.Printf("nameVal=%d\n", len(nameVal))

	if err := scanner.Err(); err != nil {
		return descriptor, err
	}

	return parseNameVal(nameVal)
}

func writeDescriptorLiteral(name string, descriptor fonts.DescriptorLiteral) string {

	var b bytes.Buffer

	fmt.Fprintf(&b, "%q: DescriptorLiteral {\n", name)
	fmt.Fprintf(&b, "\tFontName: %q,\n", descriptor.FontName)
	fmt.Fprintf(&b, "\tFontFamily: %q,\n", descriptor.FontFamily)
	fmt.Fprintf(&b, "\tFlags: 0x%04x,\n", descriptor.Flags)
	fmt.Fprintf(&b, "\tFontBBox: %#v,\n", descriptor.FontBBox)
	fmt.Fprintf(&b, "\tItalicAngle: %g,\n", descriptor.ItalicAngle)
	fmt.Fprintf(&b, "\tAscent: %g,\n", descriptor.Ascent)
	fmt.Fprintf(&b, "\tDescent: %g,\n", descriptor.Descent)
	fmt.Fprintf(&b, "\tLeading: %d,\n", descriptor.Leading)
	fmt.Fprintf(&b, "\tCapHeight: %g,\n", descriptor.CapHeight)
	fmt.Fprintf(&b, "\tXHeight: %g,\n", descriptor.XHeight)
	fmt.Fprintf(&b, "\tStemV: %g,\n", descriptor.StemV)
	fmt.Fprintf(&b, "\tStemH: %g,\n", descriptor.StemH)
	fmt.Fprintf(&b, "\tAvgWidth: %g,\n", descriptor.AvgWidth)
	fmt.Fprintf(&b, "\tMaxWidth: %g,\n", descriptor.MaxWidth)
	fmt.Fprintf(&b, "\tMissingWidth: %g,\n", descriptor.MissingWidth)
	fmt.Fprintf(&b, "}")
	return b.String()
}

func parseNameVal(nameVal map[string]string) (fonts.DescriptorLiteral, error) {
	fmt.Printf("nameVal=%q\n", nameVal)
	descriptor := fonts.DescriptorLiteral{}
	descriptor.FontName = nameVal["name"]
	descriptor.FontFamily = strings.Split(descriptor.FontName, "-")[0]
	keyVal := parseDesc(nameVal["desc"])
	descriptor.FontBBox = toBBox(keyVal["FontBBox"])
	descriptor.Flags = toInt(keyVal, "Flags")
	descriptor.ItalicAngle = toFloat(keyVal, "ItalicAngle")
	descriptor.Ascent = toFloat(keyVal, "Ascent")
	descriptor.Descent = toFloat(keyVal, "Descent")
	descriptor.Leading = toInt(keyVal, "Leading")
	descriptor.CapHeight = toFloat(keyVal, "CapHeight")
	descriptor.XHeight = toFloat(keyVal, "XHeight")
	descriptor.StemV = toFloat(keyVal, "StemV")
	descriptor.StemH = toFloat(keyVal, "StemH")
	descriptor.AvgWidth = toFloat(keyVal, "AvgWidth")
	descriptor.MissingWidth = toFloat(keyVal, "MissingWidth")

	return descriptor, nil
}

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

func toFloat(keyVal map[string]string, key string) float64 {
	text, ok := keyVal[key]
	if !ok {
		// Symbol and ZapfDingbats don't have XHeight entries
		if key == "XHeight" {
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

func updateDescMap(nameVal map[string]string, line string) {
	if len(line) == 0 {
		panic("1")
	}
	for name, re := range nameRegex {
		if _, ok := nameVal[name]; ok {
			continue
		}
		if val, ok := getRHS(re, line); ok {
			nameVal[name] = val
			break
		}
	}
}

var wantedKeys = map[string]bool{
	"Flags":        true,
	"FontBBox":     true,
	"ItalicAngle":  true,
	"Ascent":       true,
	"Descent":      true,
	"Leading":      true,
	"CapHeight":    true,
	"XHeight":      true,
	"StemV":        true,
	"StemH":        true,
	"AvgWidth":     true,
	"MaxWidth":     true,
	"MissingWidth": true,
}

func getRHS(re *regexp.Regexp, line string) (string, bool) {
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return "", false
	}
	if len(matches[1]) == 0 {
		fmt.Printf("line=%q\n", line)
		fmt.Printf("matches=%+v\n", matches[1:])
		fmt.Printf("re=%s\n", re.String())
		panic("2")
	}
	return matches[1], true
}

func parseDesc(text string) map[string]string {
	keyVal := map[string]string{}
	parts := strings.Split(text, ",")
	for _, p := range parts {
		if k, v, ok := getKeyVal(p); ok {
			if _, ok := wantedKeys[k]; !ok {
				continue
			}
			if k == "FontBBox" {
				v = getBBox(v)
			}
			keyVal[k] = v
		}
	}
	fmt.Printf("parseDesc: %q\n->%q\n", text, keyVal)
	if bbox, ok := keyVal["FontBBox"]; !ok || len(bbox) == 0 {
		panic("No FontBBox")
	}
	return keyVal
}

func getKeyVal(line string) (string, string, bool) {
	matches := reKeyVal.FindStringSubmatch(line)
	if matches == nil {
		return "", "", false
	}
	return matches[1], matches[2], true
}

func getBBox(line string) string {
	matches := reBBox.FindStringSubmatch(line)
	if matches == nil {
		panic("Not BBox")
	}
	text := matches[1]
	if len(text) == 0 {
		panic("Empty BBox")
	}
	return text
}

var (
	nameRegex = map[string]*regexp.Regexp{
		"name": regexp.MustCompile(`\$name\s*=\s*'(.*?)'`),
		"up":   regexp.MustCompile(`\$up\s*=\s*(.*?)\s*;\s*$`),
		"ut":   regexp.MustCompile(`\$ut\s*=\s*(.*?)\s*;\s*$`),
		"dw":   regexp.MustCompile(`\$dw\s*=\s*(.*?)\s*;\s*$`),
		"desc": regexp.MustCompile(`\$desc\s*=\s*array\(\s*(.*?)\s*\)\s*;\s*$`),
		"cw":   regexp.MustCompile(`\$cw\s*=\s*(.*?)\s*;\s*$`),
	}
	reKeyVal = regexp.MustCompile(`^\s*'(.*?)'\s*=>\s*(.*?)\s*$`)
	reBBox   = regexp.MustCompile(`'\[\s*(.*?)\s*\]'`)
)
