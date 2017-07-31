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
	"sort"
	"strings"
)

func main() {
	glyphlistFile := flag.String("glyphfile", "", "Glyph file to parse")
	method := flag.String("method", "glyph-to-rune", "glyph-to-rune/rune-to-glyph")

	flag.Parse()

	if len(*glyphlistFile) == 0 {
		fmt.Printf("Need to specify glyph list file via glyphfile, see -h for options\n")
		os.Exit(1)
	}

	glyphToUnicodeMap, err := parseGlyphList(*glyphlistFile)
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
		os.Exit(1)
	}

	switch *method {
	case "glyph-to-rune":
		printGlyphToRuneList(glyphToUnicodeMap)
	case "rune-to-glyph":
		printRuneToGlyphList(glyphToUnicodeMap)
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

func printGlyphToRuneList(glyphToUnicodeMap map[string]string) {
	keys := []string{}
	for key := range glyphToUnicodeMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Printf("var glyphlistGlyphToRuneMap = map[string]rune{\n")
	for _, glyph := range keys {
		ucode := glyphToUnicodeMap[glyph]
		fmt.Printf("\t\"%s\":\t'\\u%s',\n", glyph, strings.ToLower(ucode))
	}
	fmt.Printf("}\n")
}

func printRuneToGlyphList(glyphToUnicodeMap map[string]string) {
	keys := []string{}
	for key := range glyphToUnicodeMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	uniqueList := map[string]bool{}

	fmt.Printf("var glyphlistRuneToGlyphMap = map[rune]string{\n")
	for _, glyph := range keys {
		ucode := glyphToUnicodeMap[glyph]
		ucode = strings.ToLower(ucode)

		_, duplicate := uniqueList[ucode]
		if !duplicate {
			fmt.Printf("\t'\\u%s':\t\"%s\",\n", ucode, glyph)
			uniqueList[ucode] = true
		} else {
			fmt.Printf("//\t'\\u%s':\t\"%s\", // duplicate\n", ucode, glyph)
		}
	}
	fmt.Printf("}\n")
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

	gmap := map[string]bool{}
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

		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			return nil, errors.New("Invalid part")
		}

		if len(parts[1]) > 4 {
			subparts := strings.Split(parts[1], " ")
			for _, subpart := range subparts {
				//fmt.Printf("\"%s\": '\\u%s', //%s (non unique)\n", parts[0], parts[1][0:4], parts[1][4:])
				if _, has := gmap[subpart]; !has {
					//fmt.Printf("'\\u%s': \"%s\",\n", subpart, parts[0])
					gmap[subpart] = true
					glyphToUnicodeMap[parts[0]] = subpart
				} else {
					//fmt.Printf("// '\\u%s': \"%s\", (duplicate)\n", subpart, parts[0])
					glyphToUnicodeMap[parts[0]] = subpart
				}
			}
		} else {
			//fmt.Printf("\"%s\": '\\u%s',\n", parts[0], parts[1])

			if _, has := gmap[parts[1]]; !has {
				//fmt.Printf("'\\u%s': \"%s\",\n", parts[1], parts[0])
				gmap[parts[1]] = true
				glyphToUnicodeMap[parts[0]] = parts[1]
			} else {
				//fmt.Printf("// '\\u%s': \"%s\", (duplicate)\n", parts[1], parts[0])
				glyphToUnicodeMap[parts[0]] = parts[1]
			}
		}
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

		//fmt.Printf("%s\n", line)

		parts := strings.Split(line, " ")
		for _, part := range parts {
			index++
			if part == "notdef" {
				continue
			}
			//fmt.Printf("%d: \"%s\",\n", index, part)
			glyphs = append(glyphs, part)
		}
	}

	return glyphs, nil
}
