// +build unidev

package main

// Utility to generate static maps of glyph <-> rune conversions for a glyphlist.

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	glyphlistFile := flag.String("glyphfile", "", "Glyph file to parse")
	method := flag.String("method", "glyph-to-rune", "glyph-to-rune/rune-to-glyph")

	flag.Parse()

	if len(*glyphlistFile) == 0 {
		fmt.Printf("Need to specify glyph list file via glyphfile\n")
		flag.Usage()
		os.Exit(1)
	}

	glyphToUnicodeMap, err := parseGlyphList(*glyphlistFile)
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
		os.Exit(1)
	}

	switch *method {
	case "glyph-to-rune":
		printGlyphToRuneList(glyphToUnicodeMap, true)
	case "rune-to-glyph":
		printRuneToGlyphList(glyphToUnicodeMap, true)
	case "glyph-to-string":
		printGlyphToRuneList(glyphToUnicodeMap, false)
	case "string-to-glyph":
		printRuneToGlyphList(glyphToUnicodeMap, false)
	default:
		fmt.Printf("Unsupported method: %s, see -h for options\n", *method)
	}

	/*
		glyphs, err := loadGlyphlist("symbol.txt")
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			os.Exit(1)
		}
		_ = glyphs
	*/

	//printGlyphList(glyphToUnicodeMap)
	//printEncodingGlyphToRuneMap(glyphs, glyphToUnicodeMap)
	//printEncodingRuneToGlyphMap(glyphs, glyphToUnicodeMap)

}

func printGlyphToRuneList(glyphToUnicodeMap map[string]string, asRune bool) {
	keys := []string{}
	for key := range glyphToUnicodeMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Printf("var glyphlistGlyphToRuneMap = map[string]rune{ // %d entries \n", len(keys))
	for _, glyph := range keys {
		s := glyphToUnicodeMap[glyph]
		if asRune {
			r := []rune(s)[0]
			fmt.Printf("\t\t%q:\t%+q, %s\n", glyph, r, showRune(r))
		} else {
			fmt.Printf("\t\t%q:\t%+q, %s\n", glyph, s, showString(s))
		}
	}
	fmt.Printf("}\n")
}

func printRuneToGlyphList(glyphToUnicodeMap map[string]string, asRune bool) {
	keys := []string{}
	for key := range glyphToUnicodeMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	uniqueList := map[string]bool{}

	fmt.Printf("var glyphlistRuneToGlyphMap = map[rune]string{ // %d entries \n", len(keys))
	for _, glyph := range keys {
		s := glyphToUnicodeMap[glyph]
		line := ""
		runes := []rune(s)
		if len(runes) > 1 {
			line = fmt.Sprintf("%+q:\t%q, %s", s, glyph, showString(s))
		} else {
			if asRune {
				r := runes[0]
				line = fmt.Sprintf("%+q:\t%q, %s", r, glyph, showRune(r))
			} else {
				line = fmt.Sprintf("%+q:\t%q, %s", s, glyph, showString(s))
			}
		}
		_, duplicate := uniqueList[s]
		if !duplicate {
			fmt.Printf("\t\t%s\n", line)
			uniqueList[s] = true
		} else if len(runes) > 1 {
			fmt.Printf("\t\t// %s -- ambiguous - multiple characters\n", line)
		} else {
			fmt.Printf("\t\t// %s -- duplicate\n", line)
		}
	}
	fmt.Printf("}\n")
}

// showString returns a string with the Go code for string `u` and a comment showing how it prints
// if it is printable.
func showString(u string) string {
	s := ""
	printable := false
	for _, r := range u {
		if unicode.IsPrint(r) {
			printable = true
		}
	}
	if printable {
		s = fmt.Sprintf("%#q", u)
		s = fmt.Sprintf("// %s", s[1:len(s)-1])
	}
	return s
}

// showRune returns a string with the Go code for rune `r` and a comment showing how it prints if
// it is printable.
func showRune(r rune) string {
	s := ""
	if unicode.IsPrint(r) {
		s = fmt.Sprintf("%#q", r)
		s = fmt.Sprintf("// %s", s[1:len(s)-1])
	}
	return s
}

func printEncodingGlyphToRuneMap(glyphs []string, glyphToUnicodeMap map[string]string) {
	fmt.Printf("var nameEncodingGlyphToRuneMap map[string]rune = map[string]rune{\n")
	for _, glyph := range glyphs {
		ucode, has := glyphToUnicodeMap[glyph]
		if has {
			fmt.Printf("\t\"%s\":\t'\\u%s',\n", glyph, strings.ToLower(ucode))
		} else {
			fmt.Printf("'%s' - NOT FOUND\n", glyph)
		}

	}
	fmt.Printf("}\n")
}

func printEncodingRuneToGlyphMap(glyphs []string, glyphToUnicodeMap map[string]string) {
	fmt.Printf("var nameEncodingRuneToGlyphMap map[rune]string = map[rune]string{\n")
	for _, glyph := range glyphs {
		ucode, has := glyphToUnicodeMap[glyph]
		if has {
			fmt.Printf("\t'\\u%s':\t\"%s\",\n", strings.ToLower(ucode), glyph)
		} else {
			fmt.Printf("'%s' - NOT FOUND\n", glyph)
		}
	}
	fmt.Printf("}\n")
}

func parseGlyphList(filename string) (map[string]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	glyphToUnicodeMap := map[string]string{}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		line = strings.Trim(line, " \r\n")
		if line[0] == '#' {
			continue
		}
		glyph, s, err := parseGlyphString(line)
		if err != nil {
			return nil, err
		}
		glyphToUnicodeMap[glyph] = s

	}

	return glyphToUnicodeMap, nil
}

func loadGlyphlist(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	glyphs := []string{}
	reader := bufio.NewReader(f)

	index := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		line = strings.Trim(line, " \r\n")

		parts := strings.Split(line, " ")
		for _, part := range parts {
			index++
			if part == "notdef" {
				continue
			}
			glyphs = append(glyphs, part)
		}
	}

	return glyphs, nil
}

// reGlyphCodes extracts codes from string like "z;007A" which would give "z", "007A"
var reGlyphCodes = regexp.MustCompile(`^\s*(\w+)\s*;\s*(.+?)\s*$`)

func parseGlyphString(line string) (string, string, error) {
	groups := reGlyphCodes.FindStringSubmatch(line)
	if groups == nil {
		return "", "", errors.New("no match")
	}
	glyph, codesStr := groups[1], groups[2]
	runes, err := parseRunes(codesStr)
	if err != nil {
		return "", "", errors.New("no match")
	}
	return glyph, string(runes), nil
}

func parseGlyphRune(line string) (string, rune, error) {
	groups := reGlyphCodes.FindStringSubmatch(line)
	if groups == nil {
		return "", rune(0), errors.New("no match")
	}
	glyph, codesStr := groups[1], groups[2]
	runes, err := parseRunes(codesStr)
	if err != nil {
		return "", rune(0), err
	}
	return glyph, runes[0], nil
}

// parseRunes parses the string `s` for rune codes.
// An example of `s` is "FFIsmall;F766 F766 F769,0066 0066 0069"
func parseRunes(s string) ([]rune, error) {
	codeStrings := strings.Split(s, ",")
	// We only want the first string
	s = codeStrings[0]
	parts := strings.Split(s, " ")
	runes := []rune{}
	for _, p := range parts {
		h, err := strconv.ParseUint(p, 16, 32)
		if err != nil {
			return []rune{}, err
		}
		runes = append(runes, rune(h))
	}
	return runes, nil
}
