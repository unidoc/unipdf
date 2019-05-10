// +build unidev

package main

// Utility to generate static maps of glyph <-> rune conversions for a glyphlist.
// This variant of the utility relies on text files in the unidoc source tree
//  ../../glyphlist.txt
//  ../../texglyphlist.txt
//  ../../additional.txt
//  ../../unimathsymbols.txt
//  ../../Unicode.txt
//
// It builds 3 maps
//   var glyphAliases = map[string]string { // 2461 entries
//   var glyphlistGlyphToRuneMap = map[string]rune{ // 6340 entries
//   var glyphlistRuneToGlyphMap = map[rune]string{ // 6340 entries
//
// glyphlistGlyphToRuneMap and glyphlistRuneToGlyphMap map between glyphs and runes.
// More than one glyph can map to the same rune. The additional mapping are handled in glyphAliases
// NOTE: This forces us to choose a primary glyph for each rune which has multiple glyphs map to it.
//       This choice is implemented in the code below by the order of the gs.update() calls.

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	err := buildAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed. err=%v\n")
	}
}

// buildAll builds rune->glyph for glyph->rune maps for the sources as well as a glyph alias map
// for the cases where multiple glyphs map to same rune.
// NOTE: In cases where multiple glyphs map to the same rune, the order of gs.update() calls
//       determines which glyphs go in the rune<->glyph maps. The first glyph that is found goes
//       in the rune<->glyph maps and subsequent go in the aliases map.
func buildAll() error {
	gs := newGlyphState()
	gs.update(".notdef", map[string]rune{".notdef": 0xfffd})

	// Start with the base encodings
	for _, name := range baseNames {
		gr := getBaseGlyphRune(name)
		if err := gs.update(name, gr); err != nil {
			return err
		}
	}

	// Next do these mapping files
	filenames := []string{
		"glyphlist.txt",
		"texglyphlist.txt",
		"additional.txt",
	}
	for _, filename := range filenames {
		path := filepath.Join("..", "..", filename)
		gr, err := parseGlyphList(path)
		if err != nil {
			fmt.Printf("Failed to parse %q: %v\n", path, err)
			return err
		}
		if err := gs.update(filename, gr); err != nil {
			return err
		}
	}

	// Finally do these other types of mapping files
	path := filepath.Join("..", "..", "unimathsymbols.txt")
	gr, err := parseLatex(path)
	if err != nil {
		fmt.Printf("Failed to parse %q: %v\n", path, err)
		return err
	}
	if err := gs.update(path, gr); err != nil {
		return err
	}

	path = filepath.Join("..", "..", "Unicode.txt")
	gr, err = parseUnknown(path)
	if err != nil {
		fmt.Printf("Failed to parse %q: %v\n", path, err)
		return err
	}
	if err := gs.update(path, gr); err != nil {
		return err
	}

	gs.update("elipsis", map[string]rune{"elipsis": 0x2026}) // …

	printAliases(gs.aliases)
	printGlyphToRuneList(gs.glyphRune)
	printRuneToGlyphList(gs.glyphRune)
	return nil
}

type glyphState struct {
	glyphRune map[string]rune
	runeGlyph map[rune]string
	aliases   map[string]string
}

func newGlyphState() glyphState {
	return glyphState{
		glyphRune: map[string]rune{},
		runeGlyph: map[rune]string{},
		aliases:   map[string]string{},
	}
}

// update updates the glyph map state with a new glyph->rune map `gr`.
func (gs *glyphState) update(name string, gr map[string]rune) error {
	for g, r := range gr {
		if _, ok := gs.glyphRune[g]; ok {
			// Duplicate glyph. 1st definition has precedence
			continue
		}
		if g0, ok := gs.runeGlyph[r]; ok {
			// Two glyphs map to same rune. Make this one an alias.
			// g -> g0
			if base0, ok := gs.aliases[g0]; ok {
				if base0 == g {
					// Use the existing alias direction
					delete(gs.glyphRune, g0)
					gs.glyphRune[g] = r
					continue
				}
				fmt.Printf("// Existing: %q->%q\n", g0, base0)
				fmt.Printf("// New:      %q->%q\n", g, base0)
				gs.aliases[g] = base0 // transitive
			}
			gs.aliases[g] = g0
			continue
		}
		gs.glyphRune[g] = r
		gs.runeGlyph[r] = g
	}

	for g, r := range gs.glyphRune {
		if _, ok := gs.runeGlyph[r]; !ok {
			fmt.Fprintf(os.Stderr, "duplicate glyphRune[%q]=0x%04x\n", g, r)
			return errors.New("duplicate glyph")
		}
	}
	for r, g := range gs.runeGlyph {
		if _, ok := gs.glyphRune[g]; !ok {
			fmt.Fprintf(os.Stderr, "duplicate runeGlyph[0x%04x]=%q\n", r, g)
			return errors.New("duplicate rune")
		}
	}

	runeGlyph := map[rune]string{}
	for g, r := range gs.glyphRune {
		runeGlyph[r] = g
	}
	if len(gs.glyphRune) != len(runeGlyph) {
		return errors.New("inconsistent glyphRune runeGlyph")
	}
	fmt.Printf("// glyphRune=%d + %d (%d) %s\n", len(gs.glyphRune), len(gs.aliases), len(gr), name)
	return nil
}

// printGlyphToRuneList writes `glyphRune` as Go code to stdout so that it can be copied and pasted
// into glyphs_glyphlist.go
func printGlyphToRuneList(glyphRune map[string]rune) {
	keys := sorted(glyphRune)

	fmt.Printf("var glyphlistGlyphToRuneMap = map[string]rune{ // %d entries \n", len(keys))
	for _, glyph := range keys {
		r := glyphRune[glyph]
		fmt.Printf("\t\t%q:\t0x%04x, %s\n", glyph, r, showRune(r))
	}
	fmt.Printf("}\n")
}

// printGlyphToRuneList writes the reverse map of `glyphRune` as Go code to stdout so that it can be
// copied and pasted into glyphs_glyphlist.go
func printRuneToGlyphList(glyphRune map[string]rune) {
	keys := sorted(glyphRune)

	fmt.Printf("var glyphlistRuneToGlyphMap = map[rune]string{ // %d entries \n", len(keys))
	for _, glyph := range keys {
		r := glyphRune[glyph]
		fmt.Printf("\t\t0x%04x:\t%q, %s\n", r, glyph, showRune(r))
	}
	fmt.Printf("}\n")
}

// printGlyphToRuneList writes `aliases` as Go code to stdout so that it can be copied and pasted
// into glyphs_glyphlist.go
func printAliases(aliases map[string]string) {
	keys := sorted2(aliases)

	fmt.Printf("var glyphAliases = map[string]string{ // %d entries \n", len(keys))
	for _, derived := range keys {
		base := aliases[derived]
		fmt.Printf("\t\t%q:\t%q, \n", derived, base)
	}
	fmt.Printf("}\n")
}

// showRune returns a string with the Go code for rune `r` and a comment showing how it prints if
// it is printable.
func showRune(r rune) string {
	s := ""
	if unicode.IsPrint(r) {
		s = fmt.Sprintf("%#q", r)
		s = fmt.Sprintf("%s", s[1:len(s)-1])
	}
	return fmt.Sprintf("// %s %+q", s, r)
}

// sorted returns the keys of glyphRune sorted alphanumerically
func sorted(glyphRune map[string]rune) []string {
	keys := []string{}
	for key := range glyphRune {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		si, sj := keys[i], keys[j]
		ti, ni, oki := alphaNum(si)
		tj, nj, okj := alphaNum(sj)
		if oki && okj {
			if ti != tj {
				return ti < tj
			}
			return ni < nj
		}
		return si < sj
	})
	return keys
}

// sorted2 returns the keys of glyphAliases sorted alphanumerically by value then by key.
func sorted2(glyphAliases map[string]string) []string {
	keys := []string{}
	for key := range glyphAliases {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		_i := strings.Contains(keys[i], "_")
		_j := strings.Contains(keys[j], "_")
		if _i != _j {
			return _i
		}
		si, sj := glyphAliases[keys[i]], glyphAliases[keys[j]]
		ti, ni, oki := alphaNum(si)
		tj, nj, okj := alphaNum(sj)
		if oki && okj {
			if ti != tj {
				return ti < tj
			}
			return ni < nj
		}
		if si != sj {
			return si < sj
		}

		si, sj = keys[i], keys[j]
		ti, ni, oki = alphaNum(si)
		tj, nj, okj = alphaNum(sj)
		if oki && okj {
			if ti != tj {
				return ti < tj
			}
			return ni < nj
		}
		return si < sj
	})
	return keys
}

// alphaNum returns the character and numerical parts of a string like "a111". The boolean return
// is true if there is a match.
func alphaNum(s string) (string, int, bool) {
	groups := reNum.FindStringSubmatch(s)
	if len(groups) == 0 {
		return "", 0, false
	}
	n, err := strconv.Atoi(groups[2])
	if err != nil {
		return "", 0, false
	}
	return groups[1], n, true
}

// reNum extracts the character and numerical parts of a string like "a21"
var reNum = regexp.MustCompile(`([A-Za-z]+)(\d+)`)

// parseGlyphList parses a file in the format of glyphlist.txt and returns a glyph->rune map.
func parseGlyphList(filename string) (map[string]rune, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	glyphRune := map[string]rune{}
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
		glyph, r, err := parseGlyphRune(line)
		if err != nil {
			return nil, err
		}
		glyphRune[glyph] = r
	}

	return glyphRune, nil
}

// reGlyphCodes extracts codes from string like "z;007A" which would give "z", "007A"
var reGlyphCodes = regexp.MustCompile(`^\s*(\w+)\s*;\s*(.+?)\s*$`)

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

// unimathsymbols.txt is a table where the rows are lines in the text fie and the cells are
// separated by ^ symbols.
func parseLatex(filename string) (map[string]rune, error) {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parseLatex: Coudn't read %q. err=%v\n", filename, err)
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	fmt.Printf("// data=%d\n", len(data))
	fmt.Printf("// lines=%d\n", len(lines))

	cells := [][]string{}
	for i, ln := range lines {
		ln = strings.TrimSpace(ln)
		if len(ln) == 0 || ln[0] == '#' {
			continue
		}
		row := strings.Split(ln, "^")
		if len(row) != 8 {
			fmt.Fprintf(os.Stderr, "parseLatex: %d cells (expected 8) line %d: %q\n",
				len(row), i, ln)
			return nil, errors.New("bad Latex line")
		}
		cells = append(cells, row)
	}

	glyph2rune := map[string]rune{}
	glyph2char := map[string]string{}
	glyph2cols := map[string][]string{}
	glyphs := []string{}
	for i, row := range cells {
		n, err := strconv.ParseInt(row[0], 16, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parseLatex bad 0 column: row %d: %q\n", i, row)
			return nil, errors.New("bad Latex row")
		}
		char := row[1]
		glyph := row[3]
		comment := row[7]
		if glyph == "" || strings.Contains(comment, "deprecated") {
			continue
		}
		if glyph[0] != '\\' {
			fmt.Fprintf(os.Stderr, "parseLatex bad glyph: row %d: %q\n", i, row)
			return nil, errors.New("bad Latex glyph")
		}
		glyph = glyph[1:]
		r := rune(n)
		if _, ok := glyph2rune[glyph]; ok {
			fmt.Printf("// %4d: %q %q dup glyph\n", i, glyph, row)
			fmt.Printf("//      %q %q\n", glyph2cols[glyph][3], glyph2cols[glyph])
			continue
		}
		glyph2rune[glyph] = r
		glyph2char[glyph] = char
		glyph2cols[glyph] = row
		glyphs = append(glyphs, glyph)
	}
	return glyph2rune, nil
}

// parseUnknown parses a file in the format of ../..Unicode.txt and returns a glyph->rune map.
func parseUnknown(filename string) (map[string]rune, error) {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parseUnknown: Coudn't read %q. err=%v\n", filename, err)
		return nil, err

	}
	fmt.Printf("// data=%d\n", len(data))
	lines := strings.Split(string(data), "\n")
	fmt.Printf("// lines=%d\n", len(lines))

	groups := [][]string{}
	for i, ln := range lines {
		if len(ln) == 0 || ln[0] == '%' {
			continue
		}
		g := reLine.FindStringSubmatch(ln)
		if g == nil {
			fmt.Fprintf(os.Stderr, "parseUnknown: No match line %d: %q\n", i, ln)
			return nil, errors.New("bad line")
		}
		groups = append(groups, g[1:])
	}
	fmt.Printf("// groups=%d\n", len(groups))

	glyph2rune := map[string]rune{}
	for _, row := range groups {
		n, err := strconv.ParseInt(row[0], 16, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parseUnknown: Not int row=%q\n", row)
			return nil, err
		}
		r := rune(n)
		parts := reSpace.Split(row[1], -1)
		for _, g := range parts {
			if g == "" || strings.Contains(g, "000") || strings.Contains(g, "0.0") {
				continue
			}
			glyph2rune[g] = r
		}
	}
	fmt.Printf("// entries=%d\n", len(glyph2rune))
	return glyph2rune, nil
}

// The lines look like.
// 16#002D  hyphen      SP100000 hyphen-minus hyphenminus
// 16#0082        SC040000
// var reLine = regexp.MustCompile(`16#([\dA-F]{4})\s+(\S+)(?:\s+(.+?))?\s*$`)
var reLine = regexp.MustCompile(`16#([\dA-F]{4})\s+(.+?)\s*$`)
var reSpace = regexp.MustCompile(`\s+`)

// getBaseGlyphRune returns the glyph->rune map from the basicEncodings.
func getBaseGlyphRune(name string) map[string]rune {
	baseGlyphs := map[string]rune{}
	for _, glyphRune := range basicEncodings[name] {
		baseGlyphs[glyphRune.glyph] = glyphRune.r
	}
	return baseGlyphs
}

type glyphRune struct {
	glyph string
	r     rune
}

var baseNames = []string{"SymbolEncoding", "WinAnsiEncoding", "ZapfDingbatsEncoding"}

var basicEncodings = map[string]map[uint16]glyphRune{
	"SymbolEncoding": map[uint16]glyphRune{ // 189 entries
		0x20: {"space", '\u0020'},         //
		0x21: {"exclam", '\u0021'},        // !
		0x22: {"universal", '\u2200'},     // ∀
		0x23: {"numbersign", '\u0023'},    // #
		0x24: {"existential", '\u2203'},   // ∃
		0x25: {"percent", '\u0025'},       // %
		0x26: {"ampersand", '\u0026'},     // &
		0x27: {"suchthat", '\u220b'},      // ∋
		0x28: {"parenleft", '\u0028'},     // (
		0x29: {"parenright", '\u0029'},    // )
		0x2a: {"asteriskmath", '\u2217'},  // ∗
		0x2b: {"plus", '\u002b'},          // +
		0x2c: {"comma", '\u002c'},         // ,
		0x2d: {"minus", '\u2212'},         // −
		0x2e: {"period", '\u002e'},        // .
		0x2f: {"slash", '\u002f'},         // /
		0x30: {"zero", '\u0030'},          // 0
		0x31: {"one", '\u0031'},           // 1
		0x32: {"two", '\u0032'},           // 2
		0x33: {"three", '\u0033'},         // 3
		0x34: {"four", '\u0034'},          // 4
		0x35: {"five", '\u0035'},          // 5
		0x36: {"six", '\u0036'},           // 6
		0x37: {"seven", '\u0037'},         // 7
		0x38: {"eight", '\u0038'},         // 8
		0x39: {"nine", '\u0039'},          // 9
		0x3a: {"colon", '\u003a'},         // :
		0x3b: {"semicolon", '\u003b'},     // ;
		0x3c: {"less", '\u003c'},          // <
		0x3d: {"equal", '\u003d'},         // =
		0x3e: {"greater", '\u003e'},       // >
		0x3f: {"question", '\u003f'},      // ?
		0x40: {"congruent", '\u2245'},     // ≅
		0x41: {"Alpha", '\u0391'},         // Α
		0x42: {"Beta", '\u0392'},          // Β
		0x43: {"Chi", '\u03a7'},           // Χ
		0x44: {"Delta", '\u2206'},         // ∆
		0x45: {"Epsilon", '\u0395'},       // Ε
		0x46: {"Phi", '\u03a6'},           // Φ
		0x47: {"Gamma", '\u0393'},         // Γ
		0x48: {"Eta", '\u0397'},           // Η
		0x49: {"Iota", '\u0399'},          // Ι
		0x4a: {"theta1", '\u03d1'},        // ϑ
		0x4b: {"Kappa", '\u039a'},         // Κ
		0x4c: {"Lambda", '\u039b'},        // Λ
		0x4d: {"Mu", '\u039c'},            // Μ
		0x4e: {"Nu", '\u039d'},            // Ν
		0x4f: {"Omicron", '\u039f'},       // Ο
		0x50: {"Pi", '\u03a0'},            // Π
		0x51: {"Theta", '\u0398'},         // Θ
		0x52: {"Rho", '\u03a1'},           // Ρ
		0x53: {"Sigma", '\u03a3'},         // Σ
		0x54: {"Tau", '\u03a4'},           // Τ
		0x55: {"Upsilon", '\u03a5'},       // Υ
		0x56: {"sigma1", '\u03c2'},        // ς
		0x57: {"Omega", '\u2126'},         // Ω
		0x58: {"Xi", '\u039e'},            // Ξ
		0x59: {"Psi", '\u03a8'},           // Ψ
		0x5a: {"Zeta", '\u0396'},          // Ζ
		0x5b: {"bracketleft", '\u005b'},   // [
		0x5c: {"therefore", '\u2234'},     // ∴
		0x5d: {"bracketright", '\u005d'},  // ]
		0x5e: {"perpendicular", '\u22a5'}, // ⊥
		0x5f: {"underscore", '\u005f'},    // _
		0x60: {"radicalex", '\uf8e5'},
		0x61: {"alpha", '\u03b1'},        // α
		0x62: {"beta", '\u03b2'},         // β
		0x63: {"chi", '\u03c7'},          // χ
		0x64: {"delta", '\u03b4'},        // δ
		0x65: {"epsilon", '\u03b5'},      // ε
		0x66: {"phi", '\u03c6'},          // φ
		0x67: {"gamma", '\u03b3'},        // γ
		0x68: {"eta", '\u03b7'},          // η
		0x69: {"iota", '\u03b9'},         // ι
		0x6a: {"phi1", '\u03d5'},         // ϕ
		0x6b: {"kappa", '\u03ba'},        // κ
		0x6c: {"lambda", '\u03bb'},       // λ
		0x6d: {"mu", '\u00b5'},           // µ
		0x6e: {"nu", '\u03bd'},           // ν
		0x6f: {"omicron", '\u03bf'},      // ο
		0x70: {"pi", '\u03c0'},           // π
		0x71: {"theta", '\u03b8'},        // θ
		0x72: {"rho", '\u03c1'},          // ρ
		0x73: {"sigma", '\u03c3'},        // σ
		0x74: {"tau", '\u03c4'},          // τ
		0x75: {"upsilon", '\u03c5'},      // υ
		0x76: {"omega1", '\u03d6'},       // ϖ
		0x77: {"omega", '\u03c9'},        // ω
		0x78: {"xi", '\u03be'},           // ξ
		0x79: {"psi", '\u03c8'},          // ψ
		0x7a: {"zeta", '\u03b6'},         // ζ
		0x7b: {"braceleft", '\u007b'},    // {
		0x7c: {"bar", '\u007c'},          // |
		0x7d: {"braceright", '\u007d'},   // }
		0x7e: {"similar", '\u223c'},      // ∼
		0xa0: {"Euro", '\u20ac'},         // €
		0xa1: {"Upsilon1", '\u03d2'},     // ϒ
		0xa2: {"minute", '\u2032'},       // ′
		0xa3: {"lessequal", '\u2264'},    // ≤
		0xa4: {"fraction", '\u2044'},     // ⁄
		0xa5: {"infinity", '\u221e'},     // ∞
		0xa6: {"florin", '\u0192'},       // ƒ
		0xa7: {"club", '\u2663'},         // ♣
		0xa8: {"diamond", '\u2666'},      // ♦
		0xa9: {"heart", '\u2665'},        // ♥
		0xaa: {"spade", '\u2660'},        // ♠
		0xab: {"arrowboth", '\u2194'},    // ↔
		0xac: {"arrowleft", '\u2190'},    // ←
		0xad: {"arrowup", '\u2191'},      // ↑
		0xae: {"arrowright", '\u2192'},   // →
		0xaf: {"arrowdown", '\u2193'},    // ↓
		0xb0: {"degree", '\u00b0'},       // °
		0xb1: {"plusminus", '\u00b1'},    // ±
		0xb2: {"second", '\u2033'},       // ″
		0xb3: {"greaterequal", '\u2265'}, // ≥
		0xb4: {"multiply", '\u00d7'},     // ×
		0xb5: {"proportional", '\u221d'}, // ∝
		0xb6: {"partialdiff", '\u2202'},  // ∂
		0xb7: {"bullet", '\u2022'},       // •
		0xb8: {"divide", '\u00f7'},       // ÷
		0xb9: {"notequal", '\u2260'},     // ≠
		0xba: {"equivalence", '\u2261'},  // ≡
		0xbb: {"approxequal", '\u2248'},  // ≈
		0xbc: {"ellipsis", '\u2026'},     // …
		0xbd: {"arrowvertex", '\uf8e6'},
		0xbe: {"arrowhorizex", '\uf8e7'},
		0xbf: {"carriagereturn", '\u21b5'}, // ↵
		0xc0: {"aleph", '\u2135'},          // ℵ
		0xc1: {"Ifraktur", '\u2111'},       // ℑ
		0xc2: {"Rfraktur", '\u211c'},       // ℜ
		0xc3: {"weierstrass", '\u2118'},    // ℘
		0xc4: {"circlemultiply", '\u2297'}, // ⊗
		0xc5: {"circleplus", '\u2295'},     // ⊕
		0xc6: {"emptyset", '\u2205'},       // ∅
		0xc7: {"intersection", '\u2229'},   // ∩
		0xc8: {"union", '\u222a'},          // ∪
		0xc9: {"propersuperset", '\u2283'}, // ⊃
		0xca: {"reflexsuperset", '\u2287'}, // ⊇
		0xcb: {"notsubset", '\u2284'},      // ⊄
		0xcc: {"propersubset", '\u2282'},   // ⊂
		0xcd: {"reflexsubset", '\u2286'},   // ⊆
		0xce: {"element", '\u2208'},        // ∈
		0xcf: {"notelement", '\u2209'},     // ∉
		0xd0: {"angle", '\u2220'},          // ∠
		0xd1: {"gradient", '\u2207'},       // ∇
		0xd2: {"registerserif", '\uf6da'},
		0xd3: {"copyrightserif", '\uf6d9'},
		0xd4: {"trademarkserif", '\uf6db'},
		0xd5: {"product", '\u220f'},       // ∏
		0xd6: {"radical", '\u221a'},       // √
		0xd7: {"dotmath", '\u22c5'},       // ⋅
		0xd8: {"logicalnot", '\u00ac'},    // ¬
		0xd9: {"logicaland", '\u2227'},    // ∧
		0xda: {"logicalor", '\u2228'},     // ∨
		0xdb: {"arrowdblboth", '\u21d4'},  // ⇔
		0xdc: {"arrowdblleft", '\u21d0'},  // ⇐
		0xdd: {"arrowdblup", '\u21d1'},    // ⇑
		0xde: {"arrowdblright", '\u21d2'}, // ⇒
		0xdf: {"arrowdbldown", '\u21d3'},  // ⇓
		0xe0: {"lozenge", '\u25ca'},       // ◊
		0xe1: {"angleleft", '\u2329'},     // 〈
		0xe2: {"registersans", '\uf8e8'},
		0xe3: {"copyrightsans", '\uf8e9'},
		0xe4: {"trademarksans", '\uf8ea'},
		0xe5: {"summation", '\u2211'}, // ∑
		0xe6: {"parenlefttp", '\uf8eb'},
		0xe7: {"parenleftex", '\uf8ec'},
		0xe8: {"parenleftbt", '\uf8ed'},
		0xe9: {"bracketlefttp", '\uf8ee'},
		0xea: {"bracketleftex", '\uf8ef'},
		0xeb: {"bracketleftbt", '\uf8f0'},
		0xec: {"bracelefttp", '\uf8f1'},
		0xed: {"braceleftmid", '\uf8f2'},
		0xee: {"braceleftbt", '\uf8f3'},
		0xef: {"braceex", '\uf8f4'},
		0xf1: {"angleright", '\u232a'}, // 〉
		0xf2: {"integral", '\u222b'},   // ∫
		0xf3: {"integraltp", '\u2320'}, // ⌠
		0xf4: {"integralex", '\uf8f5'},
		0xf5: {"integralbt", '\u2321'}, // ⌡
		0xf6: {"parenrighttp", '\uf8f6'},
		0xf7: {"parenrightex", '\uf8f7'},
		0xf8: {"parenrightbt", '\uf8f8'},
		0xf9: {"bracketrighttp", '\uf8f9'},
		0xfa: {"bracketrightex", '\uf8fa'},
		0xfb: {"bracketrightbt", '\uf8fb'},
		0xfc: {"bracerighttp", '\uf8fc'},
		0xfd: {"bracerightmid", '\uf8fd'},
		0xfe: {"bracerightbt", '\uf8fe'},
	},
	"WinAnsiEncoding": map[uint16]glyphRune{ // 224 entries
		0x20: {"space", '\u0020'},          //
		0x21: {"exclam", '\u0021'},         // !
		0x22: {"quotedbl", '\u0022'},       // "
		0x23: {"numbersign", '\u0023'},     // #
		0x24: {"dollar", '\u0024'},         // $
		0x25: {"percent", '\u0025'},        // %
		0x26: {"ampersand", '\u0026'},      // &
		0x27: {"quotesingle", '\u0027'},    // \'
		0x28: {"parenleft", '\u0028'},      // (
		0x29: {"parenright", '\u0029'},     // )
		0x2a: {"asterisk", '\u002a'},       // *
		0x2b: {"plus", '\u002b'},           // +
		0x2c: {"comma", '\u002c'},          // ,
		0x2d: {"hyphen", '\u002d'},         // -
		0x2e: {"period", '\u002e'},         // .
		0x2f: {"slash", '\u002f'},          // /
		0x30: {"zero", '\u0030'},           // 0
		0x31: {"one", '\u0031'},            // 1
		0x32: {"two", '\u0032'},            // 2
		0x33: {"three", '\u0033'},          // 3
		0x34: {"four", '\u0034'},           // 4
		0x35: {"five", '\u0035'},           // 5
		0x36: {"six", '\u0036'},            // 6
		0x37: {"seven", '\u0037'},          // 7
		0x38: {"eight", '\u0038'},          // 8
		0x39: {"nine", '\u0039'},           // 9
		0x3a: {"colon", '\u003a'},          // :
		0x3b: {"semicolon", '\u003b'},      // ;
		0x3c: {"less", '\u003c'},           // <
		0x3d: {"equal", '\u003d'},          // =
		0x3e: {"greater", '\u003e'},        // >
		0x3f: {"question", '\u003f'},       // ?
		0x40: {"at", '\u0040'},             // @
		0x41: {"A", '\u0041'},              // A
		0x42: {"B", '\u0042'},              // B
		0x43: {"C", '\u0043'},              // C
		0x44: {"D", '\u0044'},              // D
		0x45: {"E", '\u0045'},              // E
		0x46: {"F", '\u0046'},              // F
		0x47: {"G", '\u0047'},              // G
		0x48: {"H", '\u0048'},              // H
		0x49: {"I", '\u0049'},              // I
		0x4a: {"J", '\u004a'},              // J
		0x4b: {"K", '\u004b'},              // K
		0x4c: {"L", '\u004c'},              // L
		0x4d: {"M", '\u004d'},              // M
		0x4e: {"N", '\u004e'},              // N
		0x4f: {"O", '\u004f'},              // O
		0x50: {"P", '\u0050'},              // P
		0x51: {"Q", '\u0051'},              // Q
		0x52: {"R", '\u0052'},              // R
		0x53: {"S", '\u0053'},              // S
		0x54: {"T", '\u0054'},              // T
		0x55: {"U", '\u0055'},              // U
		0x56: {"V", '\u0056'},              // V
		0x57: {"W", '\u0057'},              // W
		0x58: {"X", '\u0058'},              // X
		0x59: {"Y", '\u0059'},              // Y
		0x5a: {"Z", '\u005a'},              // Z
		0x5b: {"bracketleft", '\u005b'},    // [
		0x5c: {"backslash", '\u005c'},      // \\
		0x5d: {"bracketright", '\u005d'},   // ]
		0x5e: {"asciicircum", '\u005e'},    // ^
		0x5f: {"underscore", '\u005f'},     // _
		0x60: {"grave", '\u0060'},          // `
		0x61: {"a", '\u0061'},              // a
		0x62: {"b", '\u0062'},              // b
		0x63: {"c", '\u0063'},              // c
		0x64: {"d", '\u0064'},              // d
		0x65: {"e", '\u0065'},              // e
		0x66: {"f", '\u0066'},              // f
		0x67: {"g", '\u0067'},              // g
		0x68: {"h", '\u0068'},              // h
		0x69: {"i", '\u0069'},              // i
		0x6a: {"j", '\u006a'},              // j
		0x6b: {"k", '\u006b'},              // k
		0x6c: {"l", '\u006c'},              // l
		0x6d: {"m", '\u006d'},              // m
		0x6e: {"n", '\u006e'},              // n
		0x6f: {"o", '\u006f'},              // o
		0x70: {"p", '\u0070'},              // p
		0x71: {"q", '\u0071'},              // q
		0x72: {"r", '\u0072'},              // r
		0x73: {"s", '\u0073'},              // s
		0x74: {"t", '\u0074'},              // t
		0x75: {"u", '\u0075'},              // u
		0x76: {"v", '\u0076'},              // v
		0x77: {"w", '\u0077'},              // w
		0x78: {"x", '\u0078'},              // x
		0x79: {"y", '\u0079'},              // y
		0x7a: {"z", '\u007a'},              // z
		0x7b: {"braceleft", '\u007b'},      // {
		0x7c: {"bar", '\u007c'},            // |
		0x7d: {"braceright", '\u007d'},     // }
		0x7e: {"asciitilde", '\u007e'},     // ~
		0x7f: {"bullet", '\u2022'},         // •
		0x80: {"Euro", '\u20ac'},           // €
		0x81: {"bullet", '\u2022'},         // •
		0x82: {"quotesinglbase", '\u201a'}, // ‚
		0x83: {"florin", '\u0192'},         // ƒ
		0x84: {"quotedblbase", '\u201e'},   // „
		0x85: {"ellipsis", '\u2026'},       // …
		0x86: {"dagger", '\u2020'},         // †
		0x87: {"daggerdbl", '\u2021'},      // ‡
		0x88: {"circumflex", '\u02c6'},     // ˆ
		0x89: {"perthousand", '\u2030'},    // ‰
		0x8a: {"Scaron", '\u0160'},         // Š
		0x8b: {"guilsinglleft", '\u2039'},  // ‹
		0x8c: {"OE", '\u0152'},             // Œ
		0x8d: {"bullet", '\u2022'},         // •
		0x8e: {"Zcaron", '\u017d'},         // Ž
		0x8f: {"bullet", '\u2022'},         // •
		0x90: {"bullet", '\u2022'},         // •
		0x91: {"quoteleft", '\u2018'},      // ‘
		0x92: {"quoteright", '\u2019'},     // ’
		0x93: {"quotedblleft", '\u201c'},   // “
		0x94: {"quotedblright", '\u201d'},  // ”
		0x95: {"bullet", '\u2022'},         // •
		0x96: {"endash", '\u2013'},         // –
		0x97: {"emdash", '\u2014'},         // —
		0x98: {"tilde", '\u02dc'},          // ˜
		0x99: {"trademark", '\u2122'},      // ™
		0x9a: {"scaron", '\u0161'},         // š
		0x9b: {"guilsinglright", '\u203a'}, // ›
		0x9c: {"oe", '\u0153'},             // œ
		0x9d: {"bullet", '\u2022'},         // •
		0x9e: {"zcaron", '\u017e'},         // ž
		0x9f: {"Ydieresis", '\u0178'},      // Ÿ
		0xa0: {"space", '\u0020'},          //
		0xa1: {"exclamdown", '\u00a1'},     // ¡
		0xa2: {"cent", '\u00a2'},           // ¢
		0xa3: {"sterling", '\u00a3'},       // £
		0xa4: {"currency", '\u00a4'},       // ¤
		0xa5: {"yen", '\u00a5'},            // ¥
		0xa6: {"brokenbar", '\u00a6'},      // ¦
		0xa7: {"section", '\u00a7'},        // §
		0xa8: {"dieresis", '\u00a8'},       // ¨
		0xa9: {"copyright", '\u00a9'},      // ©
		0xaa: {"ordfeminine", '\u00aa'},    // ª
		0xab: {"guillemotleft", '\u00ab'},  // «
		0xac: {"logicalnot", '\u00ac'},     // ¬
		0xad: {"hyphen", '\u002d'},         // -
		0xae: {"registered", '\u00ae'},     // ®
		0xaf: {"macron", '\u00af'},         // ¯
		0xb0: {"degree", '\u00b0'},         // °
		0xb1: {"plusminus", '\u00b1'},      // ±
		0xb2: {"twosuperior", '\u00b2'},    // ²
		0xb3: {"threesuperior", '\u00b3'},  // ³
		0xb4: {"acute", '\u00b4'},          // ´
		0xb5: {"mu", '\u00b5'},             // µ
		0xb6: {"paragraph", '\u00b6'},      // ¶
		0xb7: {"periodcentered", '\u00b7'}, // ·
		0xb8: {"cedilla", '\u00b8'},        // ¸
		0xb9: {"onesuperior", '\u00b9'},    // ¹
		0xba: {"ordmasculine", '\u00ba'},   // º
		0xbb: {"guillemotright", '\u00bb'}, // »
		0xbc: {"onequarter", '\u00bc'},     // ¼
		0xbd: {"onehalf", '\u00bd'},        // ½
		0xbe: {"threequarters", '\u00be'},  // ¾
		0xbf: {"questiondown", '\u00bf'},   // ¿
		0xc0: {"Agrave", '\u00c0'},         // À
		0xc1: {"Aacute", '\u00c1'},         // Á
		0xc2: {"Acircumflex", '\u00c2'},    // Â
		0xc3: {"Atilde", '\u00c3'},         // Ã
		0xc4: {"Adieresis", '\u00c4'},      // Ä
		0xc5: {"Aring", '\u00c5'},          // Å
		0xc6: {"AE", '\u00c6'},             // Æ
		0xc7: {"Ccedilla", '\u00c7'},       // Ç
		0xc8: {"Egrave", '\u00c8'},         // È
		0xc9: {"Eacute", '\u00c9'},         // É
		0xca: {"Ecircumflex", '\u00ca'},    // Ê
		0xcb: {"Edieresis", '\u00cb'},      // Ë
		0xcc: {"Igrave", '\u00cc'},         // Ì
		0xcd: {"Iacute", '\u00cd'},         // Í
		0xce: {"Icircumflex", '\u00ce'},    // Î
		0xcf: {"Idieresis", '\u00cf'},      // Ï
		0xd0: {"Eth", '\u00d0'},            // Ð
		0xd1: {"Ntilde", '\u00d1'},         // Ñ
		0xd2: {"Ograve", '\u00d2'},         // Ò
		0xd3: {"Oacute", '\u00d3'},         // Ó
		0xd4: {"Ocircumflex", '\u00d4'},    // Ô
		0xd5: {"Otilde", '\u00d5'},         // Õ
		0xd6: {"Odieresis", '\u00d6'},      // Ö
		0xd7: {"multiply", '\u00d7'},       // ×
		0xd8: {"Oslash", '\u00d8'},         // Ø
		0xd9: {"Ugrave", '\u00d9'},         // Ù
		0xda: {"Uacute", '\u00da'},         // Ú
		0xdb: {"Ucircumflex", '\u00db'},    // Û
		0xdc: {"Udieresis", '\u00dc'},      // Ü
		0xdd: {"Yacute", '\u00dd'},         // Ý
		0xde: {"Thorn", '\u00de'},          // Þ
		0xdf: {"germandbls", '\u00df'},     // ß
		0xe0: {"agrave", '\u00e0'},         // à
		0xe1: {"aacute", '\u00e1'},         // á
		0xe2: {"acircumflex", '\u00e2'},    // â
		0xe3: {"atilde", '\u00e3'},         // ã
		0xe4: {"adieresis", '\u00e4'},      // ä
		0xe5: {"aring", '\u00e5'},          // å
		0xe6: {"ae", '\u00e6'},             // æ
		0xe7: {"ccedilla", '\u00e7'},       // ç
		0xe8: {"egrave", '\u00e8'},         // è
		0xe9: {"eacute", '\u00e9'},         // é
		0xea: {"ecircumflex", '\u00ea'},    // ê
		0xeb: {"edieresis", '\u00eb'},      // ë
		0xec: {"igrave", '\u00ec'},         // ì
		0xed: {"iacute", '\u00ed'},         // í
		0xee: {"icircumflex", '\u00ee'},    // î
		0xef: {"idieresis", '\u00ef'},      // ï
		0xf0: {"eth", '\u00f0'},            // ð
		0xf1: {"ntilde", '\u00f1'},         // ñ
		0xf2: {"ograve", '\u00f2'},         // ò
		0xf3: {"oacute", '\u00f3'},         // ó
		0xf4: {"ocircumflex", '\u00f4'},    // ô
		0xf5: {"otilde", '\u00f5'},         // õ
		0xf6: {"odieresis", '\u00f6'},      // ö
		0xf7: {"divide", '\u00f7'},         // ÷
		0xf8: {"oslash", '\u00f8'},         // ø
		0xf9: {"ugrave", '\u00f9'},         // ù
		0xfa: {"uacute", '\u00fa'},         // ú
		0xfb: {"ucircumflex", '\u00fb'},    // û
		0xfc: {"udieresis", '\u00fc'},      // ü
		0xfd: {"yacute", '\u00fd'},         // ý
		0xfe: {"thorn", '\u00fe'},          // þ
		0xff: {"ydieresis", '\u00ff'},      // ÿ
	},
	"ZapfDingbatsEncoding": map[uint16]glyphRune{ // 202 entries
		0x20: {"space", '\u0020'}, //
		0x21: {"a1", '\u2701'},    // ✁
		0x22: {"a2", '\u2702'},    // ✂
		0x23: {"a202", '\u2703'},  // ✃
		0x24: {"a3", '\u2704'},    // ✄
		0x25: {"a4", '\u260e'},    // ☎
		0x26: {"a5", '\u2706'},    // ✆
		0x27: {"a119", '\u2707'},  // ✇
		0x28: {"a118", '\u2708'},  // ✈
		0x29: {"a117", '\u2709'},  // ✉
		0x2a: {"a11", '\u261b'},   // ☛
		0x2b: {"a12", '\u261e'},   // ☞
		0x2c: {"a13", '\u270c'},   // ✌
		0x2d: {"a14", '\u270d'},   // ✍
		0x2e: {"a15", '\u270e'},   // ✎
		0x2f: {"a16", '\u270f'},   // ✏
		0x30: {"a105", '\u2710'},  // ✐
		0x31: {"a17", '\u2711'},   // ✑
		0x32: {"a18", '\u2712'},   // ✒
		0x33: {"a19", '\u2713'},   // ✓
		0x34: {"a20", '\u2714'},   // ✔
		0x35: {"a21", '\u2715'},   // ✕
		0x36: {"a22", '\u2716'},   // ✖
		0x37: {"a23", '\u2717'},   // ✗
		0x38: {"a24", '\u2718'},   // ✘
		0x39: {"a25", '\u2719'},   // ✙
		0x3a: {"a26", '\u271a'},   // ✚
		0x3b: {"a27", '\u271b'},   // ✛
		0x3c: {"a28", '\u271c'},   // ✜
		0x3d: {"a6", '\u271d'},    // ✝
		0x3e: {"a7", '\u271e'},    // ✞
		0x3f: {"a8", '\u271f'},    // ✟
		0x40: {"a9", '\u2720'},    // ✠
		0x41: {"a10", '\u2721'},   // ✡
		0x42: {"a29", '\u2722'},   // ✢
		0x43: {"a30", '\u2723'},   // ✣
		0x44: {"a31", '\u2724'},   // ✤
		0x45: {"a32", '\u2725'},   // ✥
		0x46: {"a33", '\u2726'},   // ✦
		0x47: {"a34", '\u2727'},   // ✧
		0x48: {"a35", '\u2605'},   // ★
		0x49: {"a36", '\u2729'},   // ✩
		0x4a: {"a37", '\u272a'},   // ✪
		0x4b: {"a38", '\u272b'},   // ✫
		0x4c: {"a39", '\u272c'},   // ✬
		0x4d: {"a40", '\u272d'},   // ✭
		0x4e: {"a41", '\u272e'},   // ✮
		0x4f: {"a42", '\u272f'},   // ✯
		0x50: {"a43", '\u2730'},   // ✰
		0x51: {"a44", '\u2731'},   // ✱
		0x52: {"a45", '\u2732'},   // ✲
		0x53: {"a46", '\u2733'},   // ✳
		0x54: {"a47", '\u2734'},   // ✴
		0x55: {"a48", '\u2735'},   // ✵
		0x56: {"a49", '\u2736'},   // ✶
		0x57: {"a50", '\u2737'},   // ✷
		0x58: {"a51", '\u2738'},   // ✸
		0x59: {"a52", '\u2739'},   // ✹
		0x5a: {"a53", '\u273a'},   // ✺
		0x5b: {"a54", '\u273b'},   // ✻
		0x5c: {"a55", '\u273c'},   // ✼
		0x5d: {"a56", '\u273d'},   // ✽
		0x5e: {"a57", '\u273e'},   // ✾
		0x5f: {"a58", '\u273f'},   // ✿
		0x60: {"a59", '\u2740'},   // ❀
		0x61: {"a60", '\u2741'},   // ❁
		0x62: {"a61", '\u2742'},   // ❂
		0x63: {"a62", '\u2743'},   // ❃
		0x64: {"a63", '\u2744'},   // ❄
		0x65: {"a64", '\u2745'},   // ❅
		0x66: {"a65", '\u2746'},   // ❆
		0x67: {"a66", '\u2747'},   // ❇
		0x68: {"a67", '\u2748'},   // ❈
		0x69: {"a68", '\u2749'},   // ❉
		0x6a: {"a69", '\u274a'},   // ❊
		0x6b: {"a70", '\u274b'},   // ❋
		0x6c: {"a71", '\u25cf'},   // ●
		0x6d: {"a72", '\u274d'},   // ❍
		0x6e: {"a73", '\u25a0'},   // ■
		0x6f: {"a74", '\u274f'},   // ❏
		0x70: {"a203", '\u2750'},  // ❐
		0x71: {"a75", '\u2751'},   // ❑
		0x72: {"a204", '\u2752'},  // ❒
		0x73: {"a76", '\u25b2'},   // ▲
		0x74: {"a77", '\u25bc'},   // ▼
		0x75: {"a78", '\u25c6'},   // ◆
		0x76: {"a79", '\u2756'},   // ❖
		0x77: {"a81", '\u25d7'},   // ◗
		0x78: {"a82", '\u2758'},   // ❘
		0x79: {"a83", '\u2759'},   // ❙
		0x7a: {"a84", '\u275a'},   // ❚
		0x7b: {"a97", '\u275b'},   // ❛
		0x7c: {"a98", '\u275c'},   // ❜
		0x7d: {"a99", '\u275d'},   // ❝
		0x7e: {"a100", '\u275e'},  // ❞
		0x80: {"a89", '\uf8d7'},
		0x81: {"a90", '\uf8d8'},
		0x82: {"a93", '\uf8d9'},
		0x83: {"a94", '\uf8da'},
		0x84: {"a91", '\uf8db'},
		0x85: {"a92", '\uf8dc'},
		0x86: {"a205", '\uf8dd'},
		0x87: {"a85", '\uf8de'},
		0x88: {"a206", '\uf8df'},
		0x89: {"a86", '\uf8e0'},
		0x8a: {"a87", '\uf8e1'},
		0x8b: {"a88", '\uf8e2'},
		0x8c: {"a95", '\uf8e3'},
		0x8d: {"a96", '\uf8e4'},
		0xa1: {"a101", '\u2761'}, // ❡
		0xa2: {"a102", '\u2762'}, // ❢
		0xa3: {"a103", '\u2763'}, // ❣
		0xa4: {"a104", '\u2764'}, // ❤
		0xa5: {"a106", '\u2765'}, // ❥
		0xa6: {"a107", '\u2766'}, // ❦
		0xa7: {"a108", '\u2767'}, // ❧
		0xa8: {"a112", '\u2663'}, // ♣
		0xa9: {"a111", '\u2666'}, // ♦
		0xaa: {"a110", '\u2665'}, // ♥
		0xab: {"a109", '\u2660'}, // ♠
		0xac: {"a120", '\u2460'}, // ①
		0xad: {"a121", '\u2461'}, // ②
		0xae: {"a122", '\u2462'}, // ③
		0xaf: {"a123", '\u2463'}, // ④
		0xb0: {"a124", '\u2464'}, // ⑤
		0xb1: {"a125", '\u2465'}, // ⑥
		0xb2: {"a126", '\u2466'}, // ⑦
		0xb3: {"a127", '\u2467'}, // ⑧
		0xb4: {"a128", '\u2468'}, // ⑨
		0xb5: {"a129", '\u2469'}, // ⑩
		0xb6: {"a130", '\u2776'}, // ❶
		0xb7: {"a131", '\u2777'}, // ❷
		0xb8: {"a132", '\u2778'}, // ❸
		0xb9: {"a133", '\u2779'}, // ❹
		0xba: {"a134", '\u277a'}, // ❺
		0xbb: {"a135", '\u277b'}, // ❻
		0xbc: {"a136", '\u277c'}, // ❼
		0xbd: {"a137", '\u277d'}, // ❽
		0xbe: {"a138", '\u277e'}, // ❾
		0xbf: {"a139", '\u277f'}, // ❿
		0xc0: {"a140", '\u2780'}, // ➀
		0xc1: {"a141", '\u2781'}, // ➁
		0xc2: {"a142", '\u2782'}, // ➂
		0xc3: {"a143", '\u2783'}, // ➃
		0xc4: {"a144", '\u2784'}, // ➄
		0xc5: {"a145", '\u2785'}, // ➅
		0xc6: {"a146", '\u2786'}, // ➆
		0xc7: {"a147", '\u2787'}, // ➇
		0xc8: {"a148", '\u2788'}, // ➈
		0xc9: {"a149", '\u2789'}, // ➉
		0xca: {"a150", '\u278a'}, // ➊
		0xcb: {"a151", '\u278b'}, // ➋
		0xcc: {"a152", '\u278c'}, // ➌
		0xcd: {"a153", '\u278d'}, // ➍
		0xce: {"a154", '\u278e'}, // ➎
		0xcf: {"a155", '\u278f'}, // ➏
		0xd0: {"a156", '\u2790'}, // ➐
		0xd1: {"a157", '\u2791'}, // ➑
		0xd2: {"a158", '\u2792'}, // ➒
		0xd3: {"a159", '\u2793'}, // ➓
		0xd4: {"a160", '\u2794'}, // ➔
		0xd5: {"a161", '\u2192'}, // →
		0xd6: {"a163", '\u2194'}, // ↔
		0xd7: {"a164", '\u2195'}, // ↕
		0xd8: {"a196", '\u2798'}, // ➘
		0xd9: {"a165", '\u2799'}, // ➙
		0xda: {"a192", '\u279a'}, // ➚
		0xdb: {"a166", '\u279b'}, // ➛
		0xdc: {"a167", '\u279c'}, // ➜
		0xdd: {"a168", '\u279d'}, // ➝
		0xde: {"a169", '\u279e'}, // ➞
		0xdf: {"a170", '\u279f'}, // ➟
		0xe0: {"a171", '\u27a0'}, // ➠
		0xe1: {"a172", '\u27a1'}, // ➡
		0xe2: {"a173", '\u27a2'}, // ➢
		0xe3: {"a162", '\u27a3'}, // ➣
		0xe4: {"a174", '\u27a4'}, // ➤
		0xe5: {"a175", '\u27a5'}, // ➥
		0xe6: {"a176", '\u27a6'}, // ➦
		0xe7: {"a177", '\u27a7'}, // ➧
		0xe8: {"a178", '\u27a8'}, // ➨
		0xe9: {"a179", '\u27a9'}, // ➩
		0xea: {"a193", '\u27aa'}, // ➪
		0xeb: {"a180", '\u27ab'}, // ➫
		0xec: {"a199", '\u27ac'}, // ➬
		0xed: {"a181", '\u27ad'}, // ➭
		0xee: {"a200", '\u27ae'}, // ➮
		0xef: {"a182", '\u27af'}, // ➯
		0xf1: {"a201", '\u27b1'}, // ➱
		0xf2: {"a183", '\u27b2'}, // ➲
		0xf3: {"a184", '\u27b3'}, // ➳
		0xf4: {"a197", '\u27b4'}, // ➴
		0xf5: {"a185", '\u27b5'}, // ➵
		0xf6: {"a194", '\u27b6'}, // ➶
		0xf7: {"a198", '\u27b7'}, // ➷
		0xf8: {"a186", '\u27b8'}, // ➸
		0xf9: {"a195", '\u27b9'}, // ➹
		0xfa: {"a187", '\u27ba'}, // ➺
		0xfb: {"a188", '\u27bb'}, // ➻
		0xfc: {"a189", '\u27bc'}, // ➼
		0xfd: {"a190", '\u27bd'}, // ➽
		0xfe: {"a191", '\u27be'}, // ➾
	},
}
