/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
/*
 * The embedded glyph to unicode mappings specified in this file are distributed under the terms
 * listed in ./glyphlist/glyphlist.txt, ./glyphlist/texglyphlist.txt and ./glyphlist/addtional.txt
 */

package textencoding

import (
	"regexp"
	"strconv"
)

// MissingCodeRune is the rune returned when there is no matching glyph.
const MissingCodeRune = '\ufffd' // �

// GlyphToRune returns true if `glyph` is in our GlyphToRune mapping.
func KnownGlyph(glyph string) bool {
	_, ok := GlyphToRune(glyph)
	return ok
}

// GlyphToRune returns the rune corresponding to glyph `glyph` if there is one.
// XXX: TODO: Can we return a string here? e.g. When we are extracting text, we want to get "ffi"
// rather than 'ﬃ'. We only need a glyph ➞ rune map when we need to convert back to glyphs.
func GlyphToRune(glyph string) (rune, bool) {
	if alias, ok := glyphAliases[glyph]; ok {
		glyph = alias
	}
	if r, ok := glyphlistGlyphToRuneMap[glyph]; ok {
		return r, true
	}
	if s, ok := texGlyphlistGlyphToStringMap[glyph]; ok {
		// XXX: Hack. Use the first rune in the teX mapping
		return []rune(s)[0], true
	}
	if r, ok := additionalGlyphlistGlyphToRuneMap[glyph]; ok {
		return r, true
	}

	if glyph == ".notdef" {
		return MissingCodeRune, true
	}

	if groups := reUniEncoding.FindStringSubmatch(glyph); groups != nil {
		n, err := strconv.ParseInt(groups[1], 16, 32)
		if err == nil {
			return rune(n), true
		}
	}

	if groups := reEncoding.FindStringSubmatch(glyph); groups != nil {
		n, err := strconv.Atoi(groups[1])
		if err == nil {
			return rune(n), true
		}
	}
	return rune(0), false
}

var (
	reEncoding    = regexp.MustCompile(`^[A-Z](\d{1,4})$`)  // C211
	reUniEncoding = regexp.MustCompile(`^uni([\dA-F]{4})$`) //uniFB03
	glyphAliases  = map[string]string{
		"f_f":     "ff",
		"f_f_i":   "ffi",
		"f_f_l":   "ffl",
		"f_i":     "fi",
		"f_l":     "fl",
		"s_t":     "st",
		"I_J":     "IJ",
		"i_j":     "ij",
		"elipsis": "ellipsis",
	}
)

var glyphlistGlyphToRuneMap = map[string]rune{ // 4281 entries
	"A":                             'A',      // A
	"AE":                            '\u00c6', // Æ
	"AEacute":                       '\u01fc', // Ǽ
	"AEmacron":                      '\u01e2', // Ǣ
	"AEsmall":                       '\uf7e6',
	"Aacute":                        '\u00c1', // Á
	"Aacutesmall":                   '\uf7e1',
	"Abreve":                        '\u0102', // Ă
	"Abreveacute":                   '\u1eae', // Ắ
	"Abrevecyrillic":                '\u04d0', // Ӑ
	"Abrevedotbelow":                '\u1eb6', // Ặ
	"Abrevegrave":                   '\u1eb0', // Ằ
	"Abrevehookabove":               '\u1eb2', // Ẳ
	"Abrevetilde":                   '\u1eb4', // Ẵ
	"Acaron":                        '\u01cd', // Ǎ
	"Acircle":                       '\u24b6', // Ⓐ
	"Acircumflex":                   '\u00c2', // Â
	"Acircumflexacute":              '\u1ea4', // Ấ
	"Acircumflexdotbelow":           '\u1eac', // Ậ
	"Acircumflexgrave":              '\u1ea6', // Ầ
	"Acircumflexhookabove":          '\u1ea8', // Ẩ
	"Acircumflexsmall":              '\uf7e2',
	"Acircumflextilde":              '\u1eaa', // Ẫ
	"Acute":                         '\uf6c9',
	"Acutesmall":                    '\uf7b4',
	"Acyrillic":                     '\u0410', // А
	"Adblgrave":                     '\u0200', // Ȁ
	"Adieresis":                     '\u00c4', // Ä
	"Adieresiscyrillic":             '\u04d2', // Ӓ
	"Adieresismacron":               '\u01de', // Ǟ
	"Adieresissmall":                '\uf7e4',
	"Adotbelow":                     '\u1ea0', // Ạ
	"Adotmacron":                    '\u01e0', // Ǡ
	"Agrave":                        '\u00c0', // À
	"Agravesmall":                   '\uf7e0',
	"Ahookabove":                    '\u1ea2', // Ả
	"Aiecyrillic":                   '\u04d4', // Ӕ
	"Ainvertedbreve":                '\u0202', // Ȃ
	"Alpha":                         '\u0391', // Α
	"Alphatonos":                    '\u0386', // Ά
	"Amacron":                       '\u0100', // Ā
	"Amonospace":                    '\uff21', // Ａ
	"Aogonek":                       '\u0104', // Ą
	"Aring":                         '\u00c5', // Å
	"Aringacute":                    '\u01fa', // Ǻ
	"Aringbelow":                    '\u1e00', // Ḁ
	"Aringsmall":                    '\uf7e5',
	"Asmall":                        '\uf761',
	"Atilde":                        '\u00c3', // Ã
	"Atildesmall":                   '\uf7e3',
	"Aybarmenian":                   '\u0531', // Ա
	"B":                             'B',      // B
	"Bcircle":                       '\u24b7', // Ⓑ
	"Bdotaccent":                    '\u1e02', // Ḃ
	"Bdotbelow":                     '\u1e04', // Ḅ
	"Becyrillic":                    '\u0411', // Б
	"Benarmenian":                   '\u0532', // Բ
	"Beta":                          '\u0392', // Β
	"Bhook":                         '\u0181', // Ɓ
	"Blinebelow":                    '\u1e06', // Ḇ
	"Bmonospace":                    '\uff22', // Ｂ
	"Brevesmall":                    '\uf6f4',
	"Bsmall":                        '\uf762',
	"Btopbar":                       '\u0182', // Ƃ
	"C":                             'C',      // C
	"Caarmenian":                    '\u053e', // Ծ
	"Cacute":                        '\u0106', // Ć
	"Caron":                         '\uf6ca',
	"Caronsmall":                    '\uf6f5',
	"Ccaron":                        '\u010c', // Č
	"Ccedilla":                      '\u00c7', // Ç
	"Ccedillaacute":                 '\u1e08', // Ḉ
	"Ccedillasmall":                 '\uf7e7',
	"Ccircle":                       '\u24b8', // Ⓒ
	"Ccircumflex":                   '\u0108', // Ĉ
	"Cdot":                          '\u010a', // Ċ
	"Cdotaccent":                    '\u010a', // Ċ
	"Cedillasmall":                  '\uf7b8',
	"Chaarmenian":                   '\u0549', // Չ
	"Cheabkhasiancyrillic":          '\u04bc', // Ҽ
	"Checyrillic":                   '\u0427', // Ч
	"Chedescenderabkhasiancyrillic": '\u04be', // Ҿ
	"Chedescendercyrillic":          '\u04b6', // Ҷ
	"Chedieresiscyrillic":           '\u04f4', // Ӵ
	"Cheharmenian":                  '\u0543', // Ճ
	"Chekhakassiancyrillic":         '\u04cb', // Ӌ
	"Cheverticalstrokecyrillic":     '\u04b8', // Ҹ
	"Chi":                  '\u03a7', // Χ
	"Chook":                '\u0187', // Ƈ
	"Circumflexsmall":      '\uf6f6',
	"Cmonospace":           '\uff23', // Ｃ
	"Coarmenian":           '\u0551', // Ց
	"Csmall":               '\uf763',
	"D":                    'D',      // D
	"DZ":                   '\u01f1', // Ǳ
	"DZcaron":              '\u01c4', // Ǆ
	"Daarmenian":           '\u0534', // Դ
	"Dafrican":             '\u0189', // Ɖ
	"Dcaron":               '\u010e', // Ď
	"Dcedilla":             '\u1e10', // Ḑ
	"Dcircle":              '\u24b9', // Ⓓ
	"Dcircumflexbelow":     '\u1e12', // Ḓ
	"Dcroat":               '\u0110', // Đ
	"Ddotaccent":           '\u1e0a', // Ḋ
	"Ddotbelow":            '\u1e0c', // Ḍ
	"Decyrillic":           '\u0414', // Д
	"Deicoptic":            '\u03ee', // Ϯ
	"Delta":                '\u2206', // ∆
	"Deltagreek":           '\u0394', // Δ
	"Dhook":                '\u018a', // Ɗ
	"Dieresis":             '\uf6cb',
	"DieresisAcute":        '\uf6cc',
	"DieresisGrave":        '\uf6cd',
	"Dieresissmall":        '\uf7a8',
	"Digammagreek":         '\u03dc', // Ϝ
	"Djecyrillic":          '\u0402', // Ђ
	"Dlinebelow":           '\u1e0e', // Ḏ
	"Dmonospace":           '\uff24', // Ｄ
	"Dotaccentsmall":       '\uf6f7',
	"Dslash":               '\u0110', // Đ
	"Dsmall":               '\uf764',
	"Dtopbar":              '\u018b', // Ƌ
	"Dz":                   '\u01f2', // ǲ
	"Dzcaron":              '\u01c5', // ǅ
	"Dzeabkhasiancyrillic": '\u04e0', // Ӡ
	"Dzecyrillic":          '\u0405', // Ѕ
	"Dzhecyrillic":         '\u040f', // Џ
	"E":                    'E',      // E
	"Eacute":               '\u00c9', // É
	"Eacutesmall":          '\uf7e9',
	"Ebreve":               '\u0114', // Ĕ
	"Ecaron":               '\u011a', // Ě
	"Ecedillabreve":        '\u1e1c', // Ḝ
	"Echarmenian":          '\u0535', // Ե
	"Ecircle":              '\u24ba', // Ⓔ
	"Ecircumflex":          '\u00ca', // Ê
	"Ecircumflexacute":     '\u1ebe', // Ế
	"Ecircumflexbelow":     '\u1e18', // Ḙ
	"Ecircumflexdotbelow":  '\u1ec6', // Ệ
	"Ecircumflexgrave":     '\u1ec0', // Ề
	"Ecircumflexhookabove": '\u1ec2', // Ể
	"Ecircumflexsmall":     '\uf7ea',
	"Ecircumflextilde":     '\u1ec4', // Ễ
	"Ecyrillic":            '\u0404', // Є
	"Edblgrave":            '\u0204', // Ȅ
	"Edieresis":            '\u00cb', // Ë
	"Edieresissmall":       '\uf7eb',
	"Edot":                 '\u0116', // Ė
	"Edotaccent":           '\u0116', // Ė
	"Edotbelow":            '\u1eb8', // Ẹ
	"Efcyrillic":           '\u0424', // Ф
	"Egrave":               '\u00c8', // È
	"Egravesmall":          '\uf7e8',
	"Eharmenian":           '\u0537', // Է
	"Ehookabove":           '\u1eba', // Ẻ
	"Eightroman":           '\u2167', // Ⅷ
	"Einvertedbreve":       '\u0206', // Ȇ
	"Eiotifiedcyrillic":    '\u0464', // Ѥ
	"Elcyrillic":           '\u041b', // Л
	"Elevenroman":          '\u216a', // Ⅺ
	"Emacron":              '\u0112', // Ē
	"Emacronacute":         '\u1e16', // Ḗ
	"Emacrongrave":         '\u1e14', // Ḕ
	"Emcyrillic":           '\u041c', // М
	"Emonospace":           '\uff25', // Ｅ
	"Encyrillic":           '\u041d', // Н
	"Endescendercyrillic":  '\u04a2', // Ң
	"Eng":                 '\u014a', // Ŋ
	"Enghecyrillic":       '\u04a4', // Ҥ
	"Enhookcyrillic":      '\u04c7', // Ӈ
	"Eogonek":             '\u0118', // Ę
	"Eopen":               '\u0190', // Ɛ
	"Epsilon":             '\u0395', // Ε
	"Epsilontonos":        '\u0388', // Έ
	"Ercyrillic":          '\u0420', // Р
	"Ereversed":           '\u018e', // Ǝ
	"Ereversedcyrillic":   '\u042d', // Э
	"Escyrillic":          '\u0421', // С
	"Esdescendercyrillic": '\u04aa', // Ҫ
	"Esh":                     '\u01a9', // Ʃ
	"Esmall":                  '\uf765',
	"Eta":                     '\u0397', // Η
	"Etarmenian":              '\u0538', // Ը
	"Etatonos":                '\u0389', // Ή
	"Eth":                     '\u00d0', // Ð
	"Ethsmall":                '\uf7f0',
	"Etilde":                  '\u1ebc', // Ẽ
	"Etildebelow":             '\u1e1a', // Ḛ
	"Euro":                    '\u20ac', // €
	"Ezh":                     '\u01b7', // Ʒ
	"Ezhcaron":                '\u01ee', // Ǯ
	"Ezhreversed":             '\u01b8', // Ƹ
	"F":                       'F',      // F
	"Fcircle":                 '\u24bb', // Ⓕ
	"Fdotaccent":              '\u1e1e', // Ḟ
	"Feharmenian":             '\u0556', // Ֆ
	"Feicoptic":               '\u03e4', // Ϥ
	"Fhook":                   '\u0191', // Ƒ
	"Fitacyrillic":            '\u0472', // Ѳ
	"Fiveroman":               '\u2164', // Ⅴ
	"Fmonospace":              '\uff26', // Ｆ
	"Fourroman":               '\u2163', // Ⅳ
	"Fsmall":                  '\uf766',
	"G":                       'G',      // G
	"GBsquare":                '\u3387', // ㎇
	"Gacute":                  '\u01f4', // Ǵ
	"Gamma":                   '\u0393', // Γ
	"Gammaafrican":            '\u0194', // Ɣ
	"Gangiacoptic":            '\u03ea', // Ϫ
	"Gbreve":                  '\u011e', // Ğ
	"Gcaron":                  '\u01e6', // Ǧ
	"Gcedilla":                '\u0122', // Ģ
	"Gcircle":                 '\u24bc', // Ⓖ
	"Gcircumflex":             '\u011c', // Ĝ
	"Gcommaaccent":            '\u0122', // Ģ
	"Gdot":                    '\u0120', // Ġ
	"Gdotaccent":              '\u0120', // Ġ
	"Gecyrillic":              '\u0413', // Г
	"Ghadarmenian":            '\u0542', // Ղ
	"Ghemiddlehookcyrillic":   '\u0494', // Ҕ
	"Ghestrokecyrillic":       '\u0492', // Ғ
	"Gheupturncyrillic":       '\u0490', // Ґ
	"Ghook":                   '\u0193', // Ɠ
	"Gimarmenian":             '\u0533', // Գ
	"Gjecyrillic":             '\u0403', // Ѓ
	"Gmacron":                 '\u1e20', // Ḡ
	"Gmonospace":              '\uff27', // Ｇ
	"Grave":                   '\uf6ce',
	"Gravesmall":              '\uf760',
	"Gsmall":                  '\uf767',
	"Gsmallhook":              '\u029b', // ʛ
	"Gstroke":                 '\u01e4', // Ǥ
	"H":                       'H',      // H
	"H18533":                  '\u25cf', // ●
	"H18543":                  '\u25aa', // ▪
	"H18551":                  '\u25ab', // ▫
	"H22073":                  '\u25a1', // □
	"HPsquare":                '\u33cb', // ㏋
	"Haabkhasiancyrillic":     '\u04a8', // Ҩ
	"Hadescendercyrillic":     '\u04b2', // Ҳ
	"Hardsigncyrillic":        '\u042a', // Ъ
	"Hbar":                    '\u0126', // Ħ
	"Hbrevebelow":             '\u1e2a', // Ḫ
	"Hcedilla":                '\u1e28', // Ḩ
	"Hcircle":                 '\u24bd', // Ⓗ
	"Hcircumflex":             '\u0124', // Ĥ
	"Hdieresis":               '\u1e26', // Ḧ
	"Hdotaccent":              '\u1e22', // Ḣ
	"Hdotbelow":               '\u1e24', // Ḥ
	"Hmonospace":              '\uff28', // Ｈ
	"Hoarmenian":              '\u0540', // Հ
	"Horicoptic":              '\u03e8', // Ϩ
	"Hsmall":                  '\uf768',
	"Hungarumlaut":            '\uf6cf',
	"Hungarumlautsmall":       '\uf6f8',
	"Hzsquare":                '\u3390', // ㎐
	"I":                       'I',      // I
	"IAcyrillic":              '\u042f', // Я
	"IJ":                      '\u0132', // Ĳ
	"IUcyrillic":              '\u042e', // Ю
	"Iacute":                  '\u00cd', // Í
	"Iacutesmall":             '\uf7ed',
	"Ibreve":                  '\u012c', // Ĭ
	"Icaron":                  '\u01cf', // Ǐ
	"Icircle":                 '\u24be', // Ⓘ
	"Icircumflex":             '\u00ce', // Î
	"Icircumflexsmall":        '\uf7ee',
	"Icyrillic":               '\u0406', // І
	"Idblgrave":               '\u0208', // Ȉ
	"Idieresis":               '\u00cf', // Ï
	"Idieresisacute":          '\u1e2e', // Ḯ
	"Idieresiscyrillic":       '\u04e4', // Ӥ
	"Idieresissmall":          '\uf7ef',
	"Idot":                    '\u0130', // İ
	"Idotaccent":              '\u0130', // İ
	"Idotbelow":               '\u1eca', // Ị
	"Iebrevecyrillic":         '\u04d6', // Ӗ
	"Iecyrillic":              '\u0415', // Е
	"Ifraktur":                '\u2111', // ℑ
	"Igrave":                  '\u00cc', // Ì
	"Igravesmall":             '\uf7ec',
	"Ihookabove":              '\u1ec8', // Ỉ
	"Iicyrillic":              '\u0418', // И
	"Iinvertedbreve":          '\u020a', // Ȋ
	"Iishortcyrillic":         '\u0419', // Й
	"Imacron":                 '\u012a', // Ī
	"Imacroncyrillic":         '\u04e2', // Ӣ
	"Imonospace":              '\uff29', // Ｉ
	"Iniarmenian":             '\u053b', // Ի
	"Iocyrillic":              '\u0401', // Ё
	"Iogonek":                 '\u012e', // Į
	"Iota":                    '\u0399', // Ι
	"Iotaafrican":             '\u0196', // Ɩ
	"Iotadieresis":            '\u03aa', // Ϊ
	"Iotatonos":               '\u038a', // Ί
	"Ismall":                  '\uf769',
	"Istroke":                 '\u0197', // Ɨ
	"Itilde":                  '\u0128', // Ĩ
	"Itildebelow":             '\u1e2c', // Ḭ
	"Izhitsacyrillic":         '\u0474', // Ѵ
	"Izhitsadblgravecyrillic": '\u0476', // Ѷ
	"J":                        'J',      // J
	"Jaarmenian":               '\u0541', // Ձ
	"Jcircle":                  '\u24bf', // Ⓙ
	"Jcircumflex":              '\u0134', // Ĵ
	"Jecyrillic":               '\u0408', // Ј
	"Jheharmenian":             '\u054b', // Ջ
	"Jmonospace":               '\uff2a', // Ｊ
	"Jsmall":                   '\uf76a',
	"K":                        'K',      // K
	"KBsquare":                 '\u3385', // ㎅
	"KKsquare":                 '\u33cd', // ㏍
	"Kabashkircyrillic":        '\u04a0', // Ҡ
	"Kacute":                   '\u1e30', // Ḱ
	"Kacyrillic":               '\u041a', // К
	"Kadescendercyrillic":      '\u049a', // Қ
	"Kahookcyrillic":           '\u04c3', // Ӄ
	"Kappa":                    '\u039a', // Κ
	"Kastrokecyrillic":         '\u049e', // Ҟ
	"Kaverticalstrokecyrillic": '\u049c', // Ҝ
	"Kcaron":                   '\u01e8', // Ǩ
	"Kcedilla":                 '\u0136', // Ķ
	"Kcircle":                  '\u24c0', // Ⓚ
	"Kcommaaccent":             '\u0136', // Ķ
	"Kdotbelow":                '\u1e32', // Ḳ
	"Keharmenian":              '\u0554', // Ք
	"Kenarmenian":              '\u053f', // Կ
	"Khacyrillic":              '\u0425', // Х
	"Kheicoptic":               '\u03e6', // Ϧ
	"Khook":                    '\u0198', // Ƙ
	"Kjecyrillic":              '\u040c', // Ќ
	"Klinebelow":               '\u1e34', // Ḵ
	"Kmonospace":               '\uff2b', // Ｋ
	"Koppacyrillic":            '\u0480', // Ҁ
	"Koppagreek":               '\u03de', // Ϟ
	"Ksicyrillic":              '\u046e', // Ѯ
	"Ksmall":                   '\uf76b',
	"L":                        'L',      // L
	"LJ":                       '\u01c7', // Ǉ
	"LL":                       '\uf6bf',
	"Lacute":                   '\u0139', // Ĺ
	"Lambda":                   '\u039b', // Λ
	"Lcaron":                   '\u013d', // Ľ
	"Lcedilla":                 '\u013b', // Ļ
	"Lcircle":                  '\u24c1', // Ⓛ
	"Lcircumflexbelow":         '\u1e3c', // Ḽ
	"Lcommaaccent":             '\u013b', // Ļ
	"Ldot":                     '\u013f', // Ŀ
	"Ldotaccent":               '\u013f', // Ŀ
	"Ldotbelow":                '\u1e36', // Ḷ
	"Ldotbelowmacron":          '\u1e38', // Ḹ
	"Liwnarmenian":             '\u053c', // Լ
	"Lj":                       '\u01c8', // ǈ
	"Ljecyrillic":              '\u0409', // Љ
	"Llinebelow":               '\u1e3a', // Ḻ
	"Lmonospace":               '\uff2c', // Ｌ
	"Lslash":                   '\u0141', // Ł
	"Lslashsmall":              '\uf6f9',
	"Lsmall":                   '\uf76c',
	"M":                        'M',      // M
	"MBsquare":                 '\u3386', // ㎆
	"Macron":                   '\uf6d0',
	"Macronsmall":              '\uf7af',
	"Macute":                   '\u1e3e', // Ḿ
	"Mcircle":                  '\u24c2', // Ⓜ
	"Mdotaccent":               '\u1e40', // Ṁ
	"Mdotbelow":                '\u1e42', // Ṃ
	"Menarmenian":              '\u0544', // Մ
	"Mmonospace":               '\uff2d', // Ｍ
	"Msmall":                   '\uf76d',
	"Mturned":                  '\u019c', // Ɯ
	"Mu":                       '\u039c', // Μ
	"N":                        'N',      // N
	"NJ":                       '\u01ca', // Ǌ
	"Nacute":                   '\u0143', // Ń
	"Ncaron":                   '\u0147', // Ň
	"Ncedilla":                 '\u0145', // Ņ
	"Ncircle":                  '\u24c3', // Ⓝ
	"Ncircumflexbelow":         '\u1e4a', // Ṋ
	"Ncommaaccent":             '\u0145', // Ņ
	"Ndotaccent":               '\u1e44', // Ṅ
	"Ndotbelow":                '\u1e46', // Ṇ
	"Nhookleft":                '\u019d', // Ɲ
	"Nineroman":                '\u2168', // Ⅸ
	"Nj":                       '\u01cb', // ǋ
	"Njecyrillic":              '\u040a', // Њ
	"Nlinebelow":               '\u1e48', // Ṉ
	"Nmonospace":               '\uff2e', // Ｎ
	"Nowarmenian":              '\u0546', // Ն
	"Nsmall":                   '\uf76e',
	"Ntilde":                   '\u00d1', // Ñ
	"Ntildesmall":              '\uf7f1',
	"Nu":                       '\u039d', // Ν
	"O":                        'O',      // O
	"OE":                       '\u0152', // Œ
	"OEsmall":                  '\uf6fa',
	"Oacute":                   '\u00d3', // Ó
	"Oacutesmall":              '\uf7f3',
	"Obarredcyrillic":          '\u04e8', // Ө
	"Obarreddieresiscyrillic":  '\u04ea', // Ӫ
	"Obreve":                   '\u014e', // Ŏ
	"Ocaron":                   '\u01d1', // Ǒ
	"Ocenteredtilde":           '\u019f', // Ɵ
	"Ocircle":                  '\u24c4', // Ⓞ
	"Ocircumflex":              '\u00d4', // Ô
	"Ocircumflexacute":         '\u1ed0', // Ố
	"Ocircumflexdotbelow":      '\u1ed8', // Ộ
	"Ocircumflexgrave":         '\u1ed2', // Ồ
	"Ocircumflexhookabove":     '\u1ed4', // Ổ
	"Ocircumflexsmall":         '\uf7f4',
	"Ocircumflextilde":         '\u1ed6', // Ỗ
	"Ocyrillic":                '\u041e', // О
	"Odblacute":                '\u0150', // Ő
	"Odblgrave":                '\u020c', // Ȍ
	"Odieresis":                '\u00d6', // Ö
	"Odieresiscyrillic":        '\u04e6', // Ӧ
	"Odieresissmall":           '\uf7f6',
	"Odotbelow":                '\u1ecc', // Ọ
	"Ogoneksmall":              '\uf6fb',
	"Ograve":                   '\u00d2', // Ò
	"Ogravesmall":              '\uf7f2',
	"Oharmenian":               '\u0555', // Օ
	"Ohm":                      '\u2126', // Ω
	"Ohookabove":               '\u1ece', // Ỏ
	"Ohorn":                    '\u01a0', // Ơ
	"Ohornacute":               '\u1eda', // Ớ
	"Ohorndotbelow":            '\u1ee2', // Ợ
	"Ohorngrave":               '\u1edc', // Ờ
	"Ohornhookabove":           '\u1ede', // Ở
	"Ohorntilde":               '\u1ee0', // Ỡ
	"Ohungarumlaut":            '\u0150', // Ő
	"Oi":                       '\u01a2', // Ƣ
	"Oinvertedbreve":           '\u020e', // Ȏ
	"Omacron":                  '\u014c', // Ō
	"Omacronacute":             '\u1e52', // Ṓ
	"Omacrongrave":             '\u1e50', // Ṑ
	"Omega":                    '\u2126', // Ω
	"Omegacyrillic":            '\u0460', // Ѡ
	"Omegagreek":               '\u03a9', // Ω
	"Omegaroundcyrillic":       '\u047a', // Ѻ
	"Omegatitlocyrillic":       '\u047c', // Ѽ
	"Omegatonos":               '\u038f', // Ώ
	"Omicron":                  '\u039f', // Ο
	"Omicrontonos":             '\u038c', // Ό
	"Omonospace":               '\uff2f', // Ｏ
	"Oneroman":                 '\u2160', // Ⅰ
	"Oogonek":                  '\u01ea', // Ǫ
	"Oogonekmacron":            '\u01ec', // Ǭ
	"Oopen":                    '\u0186', // Ɔ
	"Oslash":                   '\u00d8', // Ø
	"Oslashacute":              '\u01fe', // Ǿ
	"Oslashsmall":              '\uf7f8',
	"Osmall":                   '\uf76f',
	"Ostrokeacute":             '\u01fe', // Ǿ
	"Otcyrillic":               '\u047e', // Ѿ
	"Otilde":                   '\u00d5', // Õ
	"Otildeacute":              '\u1e4c', // Ṍ
	"Otildedieresis":           '\u1e4e', // Ṏ
	"Otildesmall":              '\uf7f5',
	"P":                        'P',      // P
	"Pacute":                   '\u1e54', // Ṕ
	"Pcircle":                  '\u24c5', // Ⓟ
	"Pdotaccent":               '\u1e56', // Ṗ
	"Pecyrillic":               '\u041f', // П
	"Peharmenian":              '\u054a', // Պ
	"Pemiddlehookcyrillic":     '\u04a6', // Ҧ
	"Phi":                    '\u03a6', // Φ
	"Phook":                  '\u01a4', // Ƥ
	"Pi":                     '\u03a0', // Π
	"Piwrarmenian":           '\u0553', // Փ
	"Pmonospace":             '\uff30', // Ｐ
	"Psi":                    '\u03a8', // Ψ
	"Psicyrillic":            '\u0470', // Ѱ
	"Psmall":                 '\uf770',
	"Q":                      'Q',      // Q
	"Qcircle":                '\u24c6', // Ⓠ
	"Qmonospace":             '\uff31', // Ｑ
	"Qsmall":                 '\uf771',
	"R":                      'R',      // R
	"Raarmenian":             '\u054c', // Ռ
	"Racute":                 '\u0154', // Ŕ
	"Rcaron":                 '\u0158', // Ř
	"Rcedilla":               '\u0156', // Ŗ
	"Rcircle":                '\u24c7', // Ⓡ
	"Rcommaaccent":           '\u0156', // Ŗ
	"Rdblgrave":              '\u0210', // Ȑ
	"Rdotaccent":             '\u1e58', // Ṙ
	"Rdotbelow":              '\u1e5a', // Ṛ
	"Rdotbelowmacron":        '\u1e5c', // Ṝ
	"Reharmenian":            '\u0550', // Ր
	"Rfraktur":               '\u211c', // ℜ
	"Rho":                    '\u03a1', // Ρ
	"Ringsmall":              '\uf6fc',
	"Rinvertedbreve":         '\u0212', // Ȓ
	"Rlinebelow":             '\u1e5e', // Ṟ
	"Rmonospace":             '\uff32', // Ｒ
	"Rsmall":                 '\uf772',
	"Rsmallinverted":         '\u0281', // ʁ
	"Rsmallinvertedsuperior": '\u02b6', // ʶ
	"S":                              'S',      // S
	"SF010000":                       '\u250c', // ┌
	"SF020000":                       '\u2514', // └
	"SF030000":                       '\u2510', // ┐
	"SF040000":                       '\u2518', // ┘
	"SF050000":                       '\u253c', // ┼
	"SF060000":                       '\u252c', // ┬
	"SF070000":                       '\u2534', // ┴
	"SF080000":                       '\u251c', // ├
	"SF090000":                       '\u2524', // ┤
	"SF100000":                       '\u2500', // ─
	"SF110000":                       '\u2502', // │
	"SF190000":                       '\u2561', // ╡
	"SF200000":                       '\u2562', // ╢
	"SF210000":                       '\u2556', // ╖
	"SF220000":                       '\u2555', // ╕
	"SF230000":                       '\u2563', // ╣
	"SF240000":                       '\u2551', // ║
	"SF250000":                       '\u2557', // ╗
	"SF260000":                       '\u255d', // ╝
	"SF270000":                       '\u255c', // ╜
	"SF280000":                       '\u255b', // ╛
	"SF360000":                       '\u255e', // ╞
	"SF370000":                       '\u255f', // ╟
	"SF380000":                       '\u255a', // ╚
	"SF390000":                       '\u2554', // ╔
	"SF400000":                       '\u2569', // ╩
	"SF410000":                       '\u2566', // ╦
	"SF420000":                       '\u2560', // ╠
	"SF430000":                       '\u2550', // ═
	"SF440000":                       '\u256c', // ╬
	"SF450000":                       '\u2567', // ╧
	"SF460000":                       '\u2568', // ╨
	"SF470000":                       '\u2564', // ╤
	"SF480000":                       '\u2565', // ╥
	"SF490000":                       '\u2559', // ╙
	"SF500000":                       '\u2558', // ╘
	"SF510000":                       '\u2552', // ╒
	"SF520000":                       '\u2553', // ╓
	"SF530000":                       '\u256b', // ╫
	"SF540000":                       '\u256a', // ╪
	"Sacute":                         '\u015a', // Ś
	"Sacutedotaccent":                '\u1e64', // Ṥ
	"Sampigreek":                     '\u03e0', // Ϡ
	"Scaron":                         '\u0160', // Š
	"Scarondotaccent":                '\u1e66', // Ṧ
	"Scaronsmall":                    '\uf6fd',
	"Scedilla":                       '\u015e', // Ş
	"Schwa":                          '\u018f', // Ə
	"Schwacyrillic":                  '\u04d8', // Ә
	"Schwadieresiscyrillic":          '\u04da', // Ӛ
	"Scircle":                        '\u24c8', // Ⓢ
	"Scircumflex":                    '\u015c', // Ŝ
	"Scommaaccent":                   '\u0218', // Ș
	"Sdotaccent":                     '\u1e60', // Ṡ
	"Sdotbelow":                      '\u1e62', // Ṣ
	"Sdotbelowdotaccent":             '\u1e68', // Ṩ
	"Seharmenian":                    '\u054d', // Ս
	"Sevenroman":                     '\u2166', // Ⅶ
	"Shaarmenian":                    '\u0547', // Շ
	"Shacyrillic":                    '\u0428', // Ш
	"Shchacyrillic":                  '\u0429', // Щ
	"Sheicoptic":                     '\u03e2', // Ϣ
	"Shhacyrillic":                   '\u04ba', // Һ
	"Shimacoptic":                    '\u03ec', // Ϭ
	"Sigma":                          '\u03a3', // Σ
	"Sixroman":                       '\u2165', // Ⅵ
	"Smonospace":                     '\uff33', // Ｓ
	"Softsigncyrillic":               '\u042c', // Ь
	"Ssmall":                         '\uf773',
	"Stigmagreek":                    '\u03da', // Ϛ
	"T":                              'T',      // T
	"Tau":                            '\u03a4', // Τ
	"Tbar":                           '\u0166', // Ŧ
	"Tcaron":                         '\u0164', // Ť
	"Tcedilla":                       '\u0162', // Ţ
	"Tcircle":                        '\u24c9', // Ⓣ
	"Tcircumflexbelow":               '\u1e70', // Ṱ
	"Tcommaaccent":                   '\u0162', // Ţ
	"Tdotaccent":                     '\u1e6a', // Ṫ
	"Tdotbelow":                      '\u1e6c', // Ṭ
	"Tecyrillic":                     '\u0422', // Т
	"Tedescendercyrillic":            '\u04ac', // Ҭ
	"Tenroman":                       '\u2169', // Ⅹ
	"Tetsecyrillic":                  '\u04b4', // Ҵ
	"Theta":                          '\u0398', // Θ
	"Thook":                          '\u01ac', // Ƭ
	"Thorn":                          '\u00de', // Þ
	"Thornsmall":                     '\uf7fe',
	"Threeroman":                     '\u2162', // Ⅲ
	"Tildesmall":                     '\uf6fe',
	"Tiwnarmenian":                   '\u054f', // Տ
	"Tlinebelow":                     '\u1e6e', // Ṯ
	"Tmonospace":                     '\uff34', // Ｔ
	"Toarmenian":                     '\u0539', // Թ
	"Tonefive":                       '\u01bc', // Ƽ
	"Tonesix":                        '\u0184', // Ƅ
	"Tonetwo":                        '\u01a7', // Ƨ
	"Tretroflexhook":                 '\u01ae', // Ʈ
	"Tsecyrillic":                    '\u0426', // Ц
	"Tshecyrillic":                   '\u040b', // Ћ
	"Tsmall":                         '\uf774',
	"Twelveroman":                    '\u216b', // Ⅻ
	"Tworoman":                       '\u2161', // Ⅱ
	"U":                              'U',      // U
	"Uacute":                         '\u00da', // Ú
	"Uacutesmall":                    '\uf7fa',
	"Ubreve":                         '\u016c', // Ŭ
	"Ucaron":                         '\u01d3', // Ǔ
	"Ucircle":                        '\u24ca', // Ⓤ
	"Ucircumflex":                    '\u00db', // Û
	"Ucircumflexbelow":               '\u1e76', // Ṷ
	"Ucircumflexsmall":               '\uf7fb',
	"Ucyrillic":                      '\u0423', // У
	"Udblacute":                      '\u0170', // Ű
	"Udblgrave":                      '\u0214', // Ȕ
	"Udieresis":                      '\u00dc', // Ü
	"Udieresisacute":                 '\u01d7', // Ǘ
	"Udieresisbelow":                 '\u1e72', // Ṳ
	"Udieresiscaron":                 '\u01d9', // Ǚ
	"Udieresiscyrillic":              '\u04f0', // Ӱ
	"Udieresisgrave":                 '\u01db', // Ǜ
	"Udieresismacron":                '\u01d5', // Ǖ
	"Udieresissmall":                 '\uf7fc',
	"Udotbelow":                      '\u1ee4', // Ụ
	"Ugrave":                         '\u00d9', // Ù
	"Ugravesmall":                    '\uf7f9',
	"Uhookabove":                     '\u1ee6', // Ủ
	"Uhorn":                          '\u01af', // Ư
	"Uhornacute":                     '\u1ee8', // Ứ
	"Uhorndotbelow":                  '\u1ef0', // Ự
	"Uhorngrave":                     '\u1eea', // Ừ
	"Uhornhookabove":                 '\u1eec', // Ử
	"Uhorntilde":                     '\u1eee', // Ữ
	"Uhungarumlaut":                  '\u0170', // Ű
	"Uhungarumlautcyrillic":          '\u04f2', // Ӳ
	"Uinvertedbreve":                 '\u0216', // Ȗ
	"Ukcyrillic":                     '\u0478', // Ѹ
	"Umacron":                        '\u016a', // Ū
	"Umacroncyrillic":                '\u04ee', // Ӯ
	"Umacrondieresis":                '\u1e7a', // Ṻ
	"Umonospace":                     '\uff35', // Ｕ
	"Uogonek":                        '\u0172', // Ų
	"Upsilon":                        '\u03a5', // Υ
	"Upsilon1":                       '\u03d2', // ϒ
	"Upsilonacutehooksymbolgreek":    '\u03d3', // ϓ
	"Upsilonafrican":                 '\u01b1', // Ʊ
	"Upsilondieresis":                '\u03ab', // Ϋ
	"Upsilondieresishooksymbolgreek": '\u03d4', // ϔ
	"Upsilonhooksymbol":              '\u03d2', // ϒ
	"Upsilontonos":                   '\u038e', // Ύ
	"Uring":                          '\u016e', // Ů
	"Ushortcyrillic":                 '\u040e', // Ў
	"Usmall":                         '\uf775',
	"Ustraightcyrillic":              '\u04ae', // Ү
	"Ustraightstrokecyrillic":        '\u04b0', // Ұ
	"Utilde":                         '\u0168', // Ũ
	"Utildeacute":                    '\u1e78', // Ṹ
	"Utildebelow":                    '\u1e74', // Ṵ
	"V":                              'V',      // V
	"Vcircle":                        '\u24cb', // Ⓥ
	"Vdotbelow":                      '\u1e7e', // Ṿ
	"Vecyrillic":                     '\u0412', // В
	"Vewarmenian":                    '\u054e', // Վ
	"Vhook":                          '\u01b2', // Ʋ
	"Vmonospace":                     '\uff36', // Ｖ
	"Voarmenian":                     '\u0548', // Ո
	"Vsmall":                         '\uf776',
	"Vtilde":                         '\u1e7c', // Ṽ
	"W":                              'W',      // W
	"Wacute":                         '\u1e82', // Ẃ
	"Wcircle":                        '\u24cc', // Ⓦ
	"Wcircumflex":                    '\u0174', // Ŵ
	"Wdieresis":                      '\u1e84', // Ẅ
	"Wdotaccent":                     '\u1e86', // Ẇ
	"Wdotbelow":                      '\u1e88', // Ẉ
	"Wgrave":                         '\u1e80', // Ẁ
	"Wmonospace":                     '\uff37', // Ｗ
	"Wsmall":                         '\uf777',
	"X":                              'X',      // X
	"Xcircle":                        '\u24cd', // Ⓧ
	"Xdieresis":                      '\u1e8c', // Ẍ
	"Xdotaccent":                     '\u1e8a', // Ẋ
	"Xeharmenian":                    '\u053d', // Խ
	"Xi":                             '\u039e', // Ξ
	"Xmonospace":                     '\uff38', // Ｘ
	"Xsmall":                         '\uf778',
	"Y":                              'Y',      // Y
	"Yacute":                         '\u00dd', // Ý
	"Yacutesmall":                    '\uf7fd',
	"Yatcyrillic":                    '\u0462', // Ѣ
	"Ycircle":                        '\u24ce', // Ⓨ
	"Ycircumflex":                    '\u0176', // Ŷ
	"Ydieresis":                      '\u0178', // Ÿ
	"Ydieresissmall":                 '\uf7ff',
	"Ydotaccent":                     '\u1e8e', // Ẏ
	"Ydotbelow":                      '\u1ef4', // Ỵ
	"Yericyrillic":                   '\u042b', // Ы
	"Yerudieresiscyrillic":           '\u04f8', // Ӹ
	"Ygrave":                         '\u1ef2', // Ỳ
	"Yhook":                          '\u01b3', // Ƴ
	"Yhookabove":                     '\u1ef6', // Ỷ
	"Yiarmenian":                     '\u0545', // Յ
	"Yicyrillic":                     '\u0407', // Ї
	"Yiwnarmenian":                   '\u0552', // Ւ
	"Ymonospace":                     '\uff39', // Ｙ
	"Ysmall":                         '\uf779',
	"Ytilde":                         '\u1ef8', // Ỹ
	"Yusbigcyrillic":                 '\u046a', // Ѫ
	"Yusbigiotifiedcyrillic":         '\u046c', // Ѭ
	"Yuslittlecyrillic":              '\u0466', // Ѧ
	"Yuslittleiotifiedcyrillic":      '\u0468', // Ѩ
	"Z":                         'Z',      // Z
	"Zaarmenian":                '\u0536', // Զ
	"Zacute":                    '\u0179', // Ź
	"Zcaron":                    '\u017d', // Ž
	"Zcaronsmall":               '\uf6ff',
	"Zcircle":                   '\u24cf', // Ⓩ
	"Zcircumflex":               '\u1e90', // Ẑ
	"Zdot":                      '\u017b', // Ż
	"Zdotaccent":                '\u017b', // Ż
	"Zdotbelow":                 '\u1e92', // Ẓ
	"Zecyrillic":                '\u0417', // З
	"Zedescendercyrillic":       '\u0498', // Ҙ
	"Zedieresiscyrillic":        '\u04de', // Ӟ
	"Zeta":                      '\u0396', // Ζ
	"Zhearmenian":               '\u053a', // Ժ
	"Zhebrevecyrillic":          '\u04c1', // Ӂ
	"Zhecyrillic":               '\u0416', // Ж
	"Zhedescendercyrillic":      '\u0496', // Җ
	"Zhedieresiscyrillic":       '\u04dc', // Ӝ
	"Zlinebelow":                '\u1e94', // Ẕ
	"Zmonospace":                '\uff3a', // Ｚ
	"Zsmall":                    '\uf77a',
	"Zstroke":                   '\u01b5', // Ƶ
	"a":                         'a',      // a
	"aabengali":                 '\u0986', // আ
	"aacute":                    '\u00e1', // á
	"aadeva":                    '\u0906', // आ
	"aagujarati":                '\u0a86', // આ
	"aagurmukhi":                '\u0a06', // ਆ
	"aamatragurmukhi":           '\u0a3e', // ਾ
	"aarusquare":                '\u3303', // ㌃
	"aavowelsignbengali":        '\u09be', // া
	"aavowelsigndeva":           '\u093e', // ा
	"aavowelsigngujarati":       '\u0abe', // ા
	"abbreviationmarkarmenian":  '\u055f', // ՟
	"abbreviationsigndeva":      '\u0970', // ॰
	"abengali":                  '\u0985', // অ
	"abopomofo":                 '\u311a', // ㄚ
	"abreve":                    '\u0103', // ă
	"abreveacute":               '\u1eaf', // ắ
	"abrevecyrillic":            '\u04d1', // ӑ
	"abrevedotbelow":            '\u1eb7', // ặ
	"abrevegrave":               '\u1eb1', // ằ
	"abrevehookabove":           '\u1eb3', // ẳ
	"abrevetilde":               '\u1eb5', // ẵ
	"acaron":                    '\u01ce', // ǎ
	"acircle":                   '\u24d0', // ⓐ
	"acircumflex":               '\u00e2', // â
	"acircumflexacute":          '\u1ea5', // ấ
	"acircumflexdotbelow":       '\u1ead', // ậ
	"acircumflexgrave":          '\u1ea7', // ầ
	"acircumflexhookabove":      '\u1ea9', // ẩ
	"acircumflextilde":          '\u1eab', // ẫ
	"acute":                     '\u00b4', // ´
	"acutebelowcmb":             '\u0317', // ̗
	"acutecmb":                  '\u0301', // ́
	"acutecomb":                 '\u0301', // ́
	"acutedeva":                 '\u0954', // ॔
	"acutelowmod":               '\u02cf', // ˏ
	"acutetonecmb":              '\u0341', // ́
	"acyrillic":                 '\u0430', // а
	"adblgrave":                 '\u0201', // ȁ
	"addakgurmukhi":             '\u0a71', // ੱ
	"adeva":                     '\u0905', // अ
	"adieresis":                 '\u00e4', // ä
	"adieresiscyrillic":         '\u04d3', // ӓ
	"adieresismacron":           '\u01df', // ǟ
	"adotbelow":                 '\u1ea1', // ạ
	"adotmacron":                '\u01e1', // ǡ
	"ae":                        '\u00e6', // æ
	"aeacute":                   '\u01fd', // ǽ
	"aekorean":                  '\u3150', // ㅐ
	"aemacron":                  '\u01e3', // ǣ
	"afii00208":                 '\u2015', // ―
	"afii08941":                 '\u20a4', // ₤
	"afii10017":                 '\u0410', // А
	"afii10018":                 '\u0411', // Б
	"afii10019":                 '\u0412', // В
	"afii10020":                 '\u0413', // Г
	"afii10021":                 '\u0414', // Д
	"afii10022":                 '\u0415', // Е
	"afii10023":                 '\u0401', // Ё
	"afii10024":                 '\u0416', // Ж
	"afii10025":                 '\u0417', // З
	"afii10026":                 '\u0418', // И
	"afii10027":                 '\u0419', // Й
	"afii10028":                 '\u041a', // К
	"afii10029":                 '\u041b', // Л
	"afii10030":                 '\u041c', // М
	"afii10031":                 '\u041d', // Н
	"afii10032":                 '\u041e', // О
	"afii10033":                 '\u041f', // П
	"afii10034":                 '\u0420', // Р
	"afii10035":                 '\u0421', // С
	"afii10036":                 '\u0422', // Т
	"afii10037":                 '\u0423', // У
	"afii10038":                 '\u0424', // Ф
	"afii10039":                 '\u0425', // Х
	"afii10040":                 '\u0426', // Ц
	"afii10041":                 '\u0427', // Ч
	"afii10042":                 '\u0428', // Ш
	"afii10043":                 '\u0429', // Щ
	"afii10044":                 '\u042a', // Ъ
	"afii10045":                 '\u042b', // Ы
	"afii10046":                 '\u042c', // Ь
	"afii10047":                 '\u042d', // Э
	"afii10048":                 '\u042e', // Ю
	"afii10049":                 '\u042f', // Я
	"afii10050":                 '\u0490', // Ґ
	"afii10051":                 '\u0402', // Ђ
	"afii10052":                 '\u0403', // Ѓ
	"afii10053":                 '\u0404', // Є
	"afii10054":                 '\u0405', // Ѕ
	"afii10055":                 '\u0406', // І
	"afii10056":                 '\u0407', // Ї
	"afii10057":                 '\u0408', // Ј
	"afii10058":                 '\u0409', // Љ
	"afii10059":                 '\u040a', // Њ
	"afii10060":                 '\u040b', // Ћ
	"afii10061":                 '\u040c', // Ќ
	"afii10062":                 '\u040e', // Ў
	"afii10063":                 '\uf6c4',
	"afii10064":                 '\uf6c5',
	"afii10065":                 '\u0430', // а
	"afii10066":                 '\u0431', // б
	"afii10067":                 '\u0432', // в
	"afii10068":                 '\u0433', // г
	"afii10069":                 '\u0434', // д
	"afii10070":                 '\u0435', // е
	"afii10071":                 '\u0451', // ё
	"afii10072":                 '\u0436', // ж
	"afii10073":                 '\u0437', // з
	"afii10074":                 '\u0438', // и
	"afii10075":                 '\u0439', // й
	"afii10076":                 '\u043a', // к
	"afii10077":                 '\u043b', // л
	"afii10078":                 '\u043c', // м
	"afii10079":                 '\u043d', // н
	"afii10080":                 '\u043e', // о
	"afii10081":                 '\u043f', // п
	"afii10082":                 '\u0440', // р
	"afii10083":                 '\u0441', // с
	"afii10084":                 '\u0442', // т
	"afii10085":                 '\u0443', // у
	"afii10086":                 '\u0444', // ф
	"afii10087":                 '\u0445', // х
	"afii10088":                 '\u0446', // ц
	"afii10089":                 '\u0447', // ч
	"afii10090":                 '\u0448', // ш
	"afii10091":                 '\u0449', // щ
	"afii10092":                 '\u044a', // ъ
	"afii10093":                 '\u044b', // ы
	"afii10094":                 '\u044c', // ь
	"afii10095":                 '\u044d', // э
	"afii10096":                 '\u044e', // ю
	"afii10097":                 '\u044f', // я
	"afii10098":                 '\u0491', // ґ
	"afii10099":                 '\u0452', // ђ
	"afii10100":                 '\u0453', // ѓ
	"afii10101":                 '\u0454', // є
	"afii10102":                 '\u0455', // ѕ
	"afii10103":                 '\u0456', // і
	"afii10104":                 '\u0457', // ї
	"afii10105":                 '\u0458', // ј
	"afii10106":                 '\u0459', // љ
	"afii10107":                 '\u045a', // њ
	"afii10108":                 '\u045b', // ћ
	"afii10109":                 '\u045c', // ќ
	"afii10110":                 '\u045e', // ў
	"afii10145":                 '\u040f', // Џ
	"afii10146":                 '\u0462', // Ѣ
	"afii10147":                 '\u0472', // Ѳ
	"afii10148":                 '\u0474', // Ѵ
	"afii10192":                 '\uf6c6',
	"afii10193":                 '\u045f', // џ
	"afii10194":                 '\u0463', // ѣ
	"afii10195":                 '\u0473', // ѳ
	"afii10196":                 '\u0475', // ѵ
	"afii10831":                 '\uf6c7',
	"afii10832":                 '\uf6c8',
	"afii10846":                 '\u04d9', // ә
	"afii299":                   '\u200e',
	"afii300":                   '\u200f',
	"afii301":                   '\u200d',
	"afii57381":                 '\u066a', // ٪
	"afii57388":                 '\u060c', // ،
	"afii57392":                 '\u0660', // ٠
	"afii57393":                 '\u0661', // ١
	"afii57394":                 '\u0662', // ٢
	"afii57395":                 '\u0663', // ٣
	"afii57396":                 '\u0664', // ٤
	"afii57397":                 '\u0665', // ٥
	"afii57398":                 '\u0666', // ٦
	"afii57399":                 '\u0667', // ٧
	"afii57400":                 '\u0668', // ٨
	"afii57401":                 '\u0669', // ٩
	"afii57403":                 '\u061b', // ؛
	"afii57407":                 '\u061f', // ؟
	"afii57409":                 '\u0621', // ء
	"afii57410":                 '\u0622', // آ
	"afii57411":                 '\u0623', // أ
	"afii57412":                 '\u0624', // ؤ
	"afii57413":                 '\u0625', // إ
	"afii57414":                 '\u0626', // ئ
	"afii57415":                 '\u0627', // ا
	"afii57416":                 '\u0628', // ب
	"afii57417":                 '\u0629', // ة
	"afii57418":                 '\u062a', // ت
	"afii57419":                 '\u062b', // ث
	"afii57420":                 '\u062c', // ج
	"afii57421":                 '\u062d', // ح
	"afii57422":                 '\u062e', // خ
	"afii57423":                 '\u062f', // د
	"afii57424":                 '\u0630', // ذ
	"afii57425":                 '\u0631', // ر
	"afii57426":                 '\u0632', // ز
	"afii57427":                 '\u0633', // س
	"afii57428":                 '\u0634', // ش
	"afii57429":                 '\u0635', // ص
	"afii57430":                 '\u0636', // ض
	"afii57431":                 '\u0637', // ط
	"afii57432":                 '\u0638', // ظ
	"afii57433":                 '\u0639', // ع
	"afii57434":                 '\u063a', // غ
	"afii57440":                 '\u0640', // ـ
	"afii57441":                 '\u0641', // ف
	"afii57442":                 '\u0642', // ق
	"afii57443":                 '\u0643', // ك
	"afii57444":                 '\u0644', // ل
	"afii57445":                 '\u0645', // م
	"afii57446":                 '\u0646', // ن
	"afii57448":                 '\u0648', // و
	"afii57449":                 '\u0649', // ى
	"afii57450":                 '\u064a', // ي
	"afii57451":                 '\u064b', // ً
	"afii57452":                 '\u064c', // ٌ
	"afii57453":                 '\u064d', // ٍ
	"afii57454":                 '\u064e', // َ
	"afii57455":                 '\u064f', // ُ
	"afii57456":                 '\u0650', // ِ
	"afii57457":                 '\u0651', // ّ
	"afii57458":                 '\u0652', // ْ
	"afii57470":                 '\u0647', // ه
	"afii57505":                 '\u06a4', // ڤ
	"afii57506":                 '\u067e', // پ
	"afii57507":                 '\u0686', // چ
	"afii57508":                 '\u0698', // ژ
	"afii57509":                 '\u06af', // گ
	"afii57511":                 '\u0679', // ٹ
	"afii57512":                 '\u0688', // ڈ
	"afii57513":                 '\u0691', // ڑ
	"afii57514":                 '\u06ba', // ں
	"afii57519":                 '\u06d2', // ے
	"afii57534":                 '\u06d5', // ە
	"afii57636":                 '\u20aa', // ₪
	"afii57645":                 '\u05be', // ־
	"afii57658":                 '\u05c3', // ׃
	"afii57664":                 '\u05d0', // א
	"afii57665":                 '\u05d1', // ב
	"afii57666":                 '\u05d2', // ג
	"afii57667":                 '\u05d3', // ד
	"afii57668":                 '\u05d4', // ה
	"afii57669":                 '\u05d5', // ו
	"afii57670":                 '\u05d6', // ז
	"afii57671":                 '\u05d7', // ח
	"afii57672":                 '\u05d8', // ט
	"afii57673":                 '\u05d9', // י
	"afii57674":                 '\u05da', // ך
	"afii57675":                 '\u05db', // כ
	"afii57676":                 '\u05dc', // ל
	"afii57677":                 '\u05dd', // ם
	"afii57678":                 '\u05de', // מ
	"afii57679":                 '\u05df', // ן
	"afii57680":                 '\u05e0', // נ
	"afii57681":                 '\u05e1', // ס
	"afii57682":                 '\u05e2', // ע
	"afii57683":                 '\u05e3', // ף
	"afii57684":                 '\u05e4', // פ
	"afii57685":                 '\u05e5', // ץ
	"afii57686":                 '\u05e6', // צ
	"afii57687":                 '\u05e7', // ק
	"afii57688":                 '\u05e8', // ר
	"afii57689":                 '\u05e9', // ש
	"afii57690":                 '\u05ea', // ת
	"afii57694":                 '\ufb2a', // שׁ
	"afii57695":                 '\ufb2b', // שׂ
	"afii57700":                 '\ufb4b', // וֹ
	"afii57705":                 '\ufb1f', // ײַ
	"afii57716":                 '\u05f0', // װ
	"afii57717":                 '\u05f1', // ױ
	"afii57718":                 '\u05f2', // ײ
	"afii57723":                 '\ufb35', // וּ
	"afii57793":                 '\u05b4', // ִ
	"afii57794":                 '\u05b5', // ֵ
	"afii57795":                 '\u05b6', // ֶ
	"afii57796":                 '\u05bb', // ֻ
	"afii57797":                 '\u05b8', // ָ
	"afii57798":                 '\u05b7', // ַ
	"afii57799":                 '\u05b0', // ְ
	"afii57800":                 '\u05b2', // ֲ
	"afii57801":                 '\u05b1', // ֱ
	"afii57802":                 '\u05b3', // ֳ
	"afii57803":                 '\u05c2', // ׂ
	"afii57804":                 '\u05c1', // ׁ
	"afii57806":                 '\u05b9', // ֹ
	"afii57807":                 '\u05bc', // ּ
	"afii57839":                 '\u05bd', // ֽ
	"afii57841":                 '\u05bf', // ֿ
	"afii57842":                 '\u05c0', // ׀
	"afii57929":                 '\u02bc', // ʼ
	"afii61248":                 '\u2105', // ℅
	"afii61289":                 '\u2113', // ℓ
	"afii61352":                 '\u2116', // №
	"afii61573":                 '\u202c',
	"afii61574":                 '\u202d',
	"afii61575":                 '\u202e',
	"afii61664":                 '\u200c',
	"afii63167":                 '\u066d', // ٭
	"afii64937":                 '\u02bd', // ʽ
	"agrave":                    '\u00e0', // à
	"agujarati":                 '\u0a85', // અ
	"agurmukhi":                 '\u0a05', // ਅ
	"ahiragana":                 '\u3042', // あ
	"ahookabove":                '\u1ea3', // ả
	"aibengali":                 '\u0990', // ঐ
	"aibopomofo":                '\u311e', // ㄞ
	"aideva":                    '\u0910', // ऐ
	"aiecyrillic":               '\u04d5', // ӕ
	"aigujarati":                '\u0a90', // ઐ
	"aigurmukhi":                '\u0a10', // ਐ
	"aimatragurmukhi":           '\u0a48', // ੈ
	"ainarabic":                 '\u0639', // ع
	"ainfinalarabic":            '\ufeca', // ﻊ
	"aininitialarabic":          '\ufecb', // ﻋ
	"ainmedialarabic":           '\ufecc', // ﻌ
	"ainvertedbreve":            '\u0203', // ȃ
	"aivowelsignbengali":        '\u09c8', // ৈ
	"aivowelsigndeva":           '\u0948', // ै
	"aivowelsigngujarati":       '\u0ac8', // ૈ
	"akatakana":                 '\u30a2', // ア
	"akatakanahalfwidth":        '\uff71', // ｱ
	"akorean":                   '\u314f', // ㅏ
	"alef":                      '\u05d0', // א
	"alefarabic":                '\u0627', // ا
	"alefdageshhebrew":          '\ufb30', // אּ
	"aleffinalarabic":           '\ufe8e', // ﺎ
	"alefhamzaabovearabic":      '\u0623', // أ
	"alefhamzaabovefinalarabic": '\ufe84', // ﺄ
	"alefhamzabelowarabic":      '\u0625', // إ
	"alefhamzabelowfinalarabic": '\ufe88', // ﺈ
	"alefhebrew":                '\u05d0', // א
	"aleflamedhebrew":           '\ufb4f', // ﭏ
	"alefmaddaabovearabic":      '\u0622', // آ
	"alefmaddaabovefinalarabic": '\ufe82', // ﺂ
	"alefmaksuraarabic":         '\u0649', // ى
	"alefmaksurafinalarabic":    '\ufef0', // ﻰ
	"alefmaksurainitialarabic":  '\ufef3', // ﻳ
	"alefmaksuramedialarabic":   '\ufef4', // ﻴ
	"alefpatahhebrew":           '\ufb2e', // אַ
	"alefqamatshebrew":          '\ufb2f', // אָ
	"aleph":                     '\u2135', // ℵ
	"allequal":                  '\u224c', // ≌
	"alpha":                     '\u03b1', // α
	"alphatonos":                '\u03ac', // ά
	"amacron":                   '\u0101', // ā
	"amonospace":                '\uff41', // ａ
	"ampersand":                 '&',      // &
	"ampersandmonospace":        '\uff06', // ＆
	"ampersandsmall":            '\uf726',
	"amsquare":                  '\u33c2', // ㏂
	"anbopomofo":                '\u3122', // ㄢ
	"angbopomofo":               '\u3124', // ㄤ
	"angkhankhuthai":            '\u0e5a', // ๚
	"angle":                     '\u2220', // ∠
	"anglebracketleft":          '\u3008', // 〈
	"anglebracketleftvertical":  '\ufe3f', // ︿
	"anglebracketright":         '\u3009', // 〉
	"anglebracketrightvertical": '\ufe40', // ﹀
	"angleleft":                 '\u2329', // 〈
	"angleright":                '\u232a', // 〉
	"angstrom":                  '\u212b', // Å
	"anoteleia":                 '\u0387', // ·
	"anudattadeva":              '\u0952', // ॒
	"anusvarabengali":           '\u0982', // ং
	"anusvaradeva":              '\u0902', // ं
	"anusvaragujarati":          '\u0a82', // ં
	"aogonek":                   '\u0105', // ą
	"apaatosquare":              '\u3300', // ㌀
	"aparen":                    '\u249c', // ⒜
	"apostrophearmenian":        '\u055a', // ՚
	"apostrophemod":             '\u02bc', // ʼ
	"apple":                     '\uf8ff',
	"approaches":                '\u2250', // ≐
	"approxequal":               '\u2248', // ≈
	"approxequalorimage":        '\u2252', // ≒
	"approximatelyequal":        '\u2245', // ≅
	"araeaekorean":              '\u318e', // ㆎ
	"araeakorean":               '\u318d', // ㆍ
	"arc":                       '\u2312', // ⌒
	"arighthalfring":            '\u1e9a', // ẚ
	"aring":                     '\u00e5', // å
	"aringacute":                '\u01fb', // ǻ
	"aringbelow":                '\u1e01', // ḁ
	"arrowboth":                 '\u2194', // ↔
	"arrowdashdown":             '\u21e3', // ⇣
	"arrowdashleft":             '\u21e0', // ⇠
	"arrowdashright":            '\u21e2', // ⇢
	"arrowdashup":               '\u21e1', // ⇡
	"arrowdblboth":              '\u21d4', // ⇔
	"arrowdbldown":              '\u21d3', // ⇓
	"arrowdblleft":              '\u21d0', // ⇐
	"arrowdblright":             '\u21d2', // ⇒
	"arrowdblup":                '\u21d1', // ⇑
	"arrowdown":                 '\u2193', // ↓
	"arrowdownleft":             '\u2199', // ↙
	"arrowdownright":            '\u2198', // ↘
	"arrowdownwhite":            '\u21e9', // ⇩
	"arrowheaddownmod":          '\u02c5', // ˅
	"arrowheadleftmod":          '\u02c2', // ˂
	"arrowheadrightmod":         '\u02c3', // ˃
	"arrowheadupmod":            '\u02c4', // ˄
	"arrowhorizex":              '\uf8e7',
	"arrowleft":                 '\u2190', // ←
	"arrowleftdbl":              '\u21d0', // ⇐
	"arrowleftdblstroke":        '\u21cd', // ⇍
	"arrowleftoverright":        '\u21c6', // ⇆
	"arrowleftwhite":            '\u21e6', // ⇦
	"arrowright":                '\u2192', // →
	"arrowrightdblstroke":       '\u21cf', // ⇏
	"arrowrightheavy":           '\u279e', // ➞
	"arrowrightoverleft":        '\u21c4', // ⇄
	"arrowrightwhite":           '\u21e8', // ⇨
	"arrowtableft":              '\u21e4', // ⇤
	"arrowtabright":             '\u21e5', // ⇥
	"arrowup":                   '\u2191', // ↑
	"arrowupdn":                 '\u2195', // ↕
	"arrowupdnbse":              '\u21a8', // ↨
	"arrowupdownbase":           '\u21a8', // ↨
	"arrowupleft":               '\u2196', // ↖
	"arrowupleftofdown":         '\u21c5', // ⇅
	"arrowupright":              '\u2197', // ↗
	"arrowupwhite":              '\u21e7', // ⇧
	"arrowvertex":               '\uf8e6',
	"asciicircum":               '^',      // ^
	"asciicircummonospace":      '\uff3e', // ＾
	"asciitilde":                '~',      // ~
	"asciitildemonospace":       '\uff5e', // ～
	"ascript":                   '\u0251', // ɑ
	"ascriptturned":             '\u0252', // ɒ
	"asmallhiragana":            '\u3041', // ぁ
	"asmallkatakana":            '\u30a1', // ァ
	"asmallkatakanahalfwidth":   '\uff67', // ｧ
	"asterisk":                  '*',      // *
	"asteriskaltonearabic":      '\u066d', // ٭
	"asteriskarabic":            '\u066d', // ٭
	"asteriskmath":              '\u2217', // ∗
	"asteriskmonospace":         '\uff0a', // ＊
	"asterisksmall":             '\ufe61', // ﹡
	"asterism":                  '\u2042', // ⁂
	"asuperior":                 '\uf6e9',
	"asymptoticallyequal":       '\u2243', // ≃
	"at":                                  '@',      // @
	"atilde":                              '\u00e3', // ã
	"atmonospace":                         '\uff20', // ＠
	"atsmall":                             '\ufe6b', // ﹫
	"aturned":                             '\u0250', // ɐ
	"aubengali":                           '\u0994', // ঔ
	"aubopomofo":                          '\u3120', // ㄠ
	"audeva":                              '\u0914', // औ
	"augujarati":                          '\u0a94', // ઔ
	"augurmukhi":                          '\u0a14', // ਔ
	"aulengthmarkbengali":                 '\u09d7', // ৗ
	"aumatragurmukhi":                     '\u0a4c', // ੌ
	"auvowelsignbengali":                  '\u09cc', // ৌ
	"auvowelsigndeva":                     '\u094c', // ौ
	"auvowelsigngujarati":                 '\u0acc', // ૌ
	"avagrahadeva":                        '\u093d', // ऽ
	"aybarmenian":                         '\u0561', // ա
	"ayin":                                '\u05e2', // ע
	"ayinaltonehebrew":                    '\ufb20', // ﬠ
	"ayinhebrew":                          '\u05e2', // ע
	"b":                                   'b',      // b
	"babengali":                           '\u09ac', // ব
	"backslash":                           '\\',     // \\
	"backslashmonospace":                  '\uff3c', // ＼
	"badeva":                              '\u092c', // ब
	"bagujarati":                          '\u0aac', // બ
	"bagurmukhi":                          '\u0a2c', // ਬ
	"bahiragana":                          '\u3070', // ば
	"bahtthai":                            '\u0e3f', // ฿
	"bakatakana":                          '\u30d0', // バ
	"bar":                                 '|',      // |
	"barmonospace":                        '\uff5c', // ｜
	"bbopomofo":                           '\u3105', // ㄅ
	"bcircle":                             '\u24d1', // ⓑ
	"bdotaccent":                          '\u1e03', // ḃ
	"bdotbelow":                           '\u1e05', // ḅ
	"beamedsixteenthnotes":                '\u266c', // ♬
	"because":                             '\u2235', // ∵
	"becyrillic":                          '\u0431', // б
	"beharabic":                           '\u0628', // ب
	"behfinalarabic":                      '\ufe90', // ﺐ
	"behinitialarabic":                    '\ufe91', // ﺑ
	"behiragana":                          '\u3079', // べ
	"behmedialarabic":                     '\ufe92', // ﺒ
	"behmeeminitialarabic":                '\ufc9f', // ﲟ
	"behmeemisolatedarabic":               '\ufc08', // ﰈ
	"behnoonfinalarabic":                  '\ufc6d', // ﱭ
	"bekatakana":                          '\u30d9', // ベ
	"benarmenian":                         '\u0562', // բ
	"bet":                                 '\u05d1', // ב
	"beta":                                '\u03b2', // β
	"betasymbolgreek":                     '\u03d0', // ϐ
	"betdagesh":                           '\ufb31', // בּ
	"betdageshhebrew":                     '\ufb31', // בּ
	"bethebrew":                           '\u05d1', // ב
	"betrafehebrew":                       '\ufb4c', // בֿ
	"bhabengali":                          '\u09ad', // ভ
	"bhadeva":                             '\u092d', // भ
	"bhagujarati":                         '\u0aad', // ભ
	"bhagurmukhi":                         '\u0a2d', // ਭ
	"bhook":                               '\u0253', // ɓ
	"bihiragana":                          '\u3073', // び
	"bikatakana":                          '\u30d3', // ビ
	"bilabialclick":                       '\u0298', // ʘ
	"bindigurmukhi":                       '\u0a02', // ਂ
	"birusquare":                          '\u3331', // ㌱
	"blackcircle":                         '\u25cf', // ●
	"blackdiamond":                        '\u25c6', // ◆
	"blackdownpointingtriangle":           '\u25bc', // ▼
	"blackleftpointingpointer":            '\u25c4', // ◄
	"blackleftpointingtriangle":           '\u25c0', // ◀
	"blacklenticularbracketleft":          '\u3010', // 【
	"blacklenticularbracketleftvertical":  '\ufe3b', // ︻
	"blacklenticularbracketright":         '\u3011', // 】
	"blacklenticularbracketrightvertical": '\ufe3c', // ︼
	"blacklowerlefttriangle":              '\u25e3', // ◣
	"blacklowerrighttriangle":             '\u25e2', // ◢
	"blackrectangle":                      '\u25ac', // ▬
	"blackrightpointingpointer":           '\u25ba', // ►
	"blackrightpointingtriangle":          '\u25b6', // ▶
	"blacksmallsquare":                    '\u25aa', // ▪
	"blacksmilingface":                    '\u263b', // ☻
	"blacksquare":                         '\u25a0', // ■
	"blackstar":                           '\u2605', // ★
	"blackupperlefttriangle":              '\u25e4', // ◤
	"blackupperrighttriangle":             '\u25e5', // ◥
	"blackuppointingsmalltriangle":        '\u25b4', // ▴
	"blackuppointingtriangle":             '\u25b2', // ▲
	"blank":                               '\u2423', // ␣
	"blinebelow":                          '\u1e07', // ḇ
	"block":                               '\u2588', // █
	"bmonospace":                          '\uff42', // ｂ
	"bobaimaithai":                        '\u0e1a', // บ
	"bohiragana":                          '\u307c', // ぼ
	"bokatakana":                          '\u30dc', // ボ
	"bparen":                              '\u249d', // ⒝
	"bqsquare":                            '\u33c3', // ㏃
	"braceex":                             '\uf8f4',
	"braceleft":                           '{', // {
	"braceleftbt":                         '\uf8f3',
	"braceleftmid":                        '\uf8f2',
	"braceleftmonospace":                  '\uff5b', // ｛
	"braceleftsmall":                      '\ufe5b', // ﹛
	"bracelefttp":                         '\uf8f1',
	"braceleftvertical":                   '\ufe37', // ︷
	"braceright":                          '}',      // }
	"bracerightbt":                        '\uf8fe',
	"bracerightmid":                       '\uf8fd',
	"bracerightmonospace":                 '\uff5d', // ｝
	"bracerightsmall":                     '\ufe5c', // ﹜
	"bracerighttp":                        '\uf8fc',
	"bracerightvertical":                  '\ufe38', // ︸
	"bracketleft":                         '[',      // [
	"bracketleftbt":                       '\uf8f0',
	"bracketleftex":                       '\uf8ef',
	"bracketleftmonospace":                '\uff3b', // ［
	"bracketlefttp":                       '\uf8ee',
	"bracketright":                        ']', // ]
	"bracketrightbt":                      '\uf8fb',
	"bracketrightex":                      '\uf8fa',
	"bracketrightmonospace":               '\uff3d', // ］
	"bracketrighttp":                      '\uf8f9',
	"breve":                               '\u02d8', // ˘
	"brevebelowcmb":                       '\u032e', // ̮
	"brevecmb":                            '\u0306', // ̆
	"breveinvertedbelowcmb":               '\u032f', // ̯
	"breveinvertedcmb":                    '\u0311', // ̑
	"breveinverteddoublecmb":              '\u0361', // ͡
	"bridgebelowcmb":                      '\u032a', // ̪
	"bridgeinvertedbelowcmb":              '\u033a', // ̺
	"brokenbar":                           '\u00a6', // ¦
	"bstroke":                             '\u0180', // ƀ
	"bsuperior":                           '\uf6ea',
	"btopbar":                             '\u0183', // ƃ
	"buhiragana":                          '\u3076', // ぶ
	"bukatakana":                          '\u30d6', // ブ
	"bullet":                              '\u2022', // •
	"bulletinverse":                       '\u25d8', // ◘
	"bulletoperator":                      '\u2219', // ∙
	"bullseye":                            '\u25ce', // ◎
	"c":                                   'c',      // c
	"caarmenian":                          '\u056e', // ծ
	"cabengali":                           '\u099a', // চ
	"cacute":                              '\u0107', // ć
	"cadeva":                              '\u091a', // च
	"cagujarati":                          '\u0a9a', // ચ
	"cagurmukhi":                          '\u0a1a', // ਚ
	"calsquare":                           '\u3388', // ㎈
	"candrabindubengali":                  '\u0981', // ঁ
	"candrabinducmb":                      '\u0310', // ̐
	"candrabindudeva":                     '\u0901', // ँ
	"candrabindugujarati":                 '\u0a81', // ઁ
	"capslock":                            '\u21ea', // ⇪
	"careof":                              '\u2105', // ℅
	"caron":                               '\u02c7', // ˇ
	"caronbelowcmb":                       '\u032c', // ̬
	"caroncmb":                            '\u030c', // ̌
	"carriagereturn":                      '\u21b5', // ↵
	"cbopomofo":                           '\u3118', // ㄘ
	"ccaron":                              '\u010d', // č
	"ccedilla":                            '\u00e7', // ç
	"ccedillaacute":                       '\u1e09', // ḉ
	"ccircle":                             '\u24d2', // ⓒ
	"ccircumflex":                         '\u0109', // ĉ
	"ccurl":                               '\u0255', // ɕ
	"cdot":                                '\u010b', // ċ
	"cdotaccent":                          '\u010b', // ċ
	"cdsquare":                            '\u33c5', // ㏅
	"cedilla":                             '\u00b8', // ¸
	"cedillacmb":                          '\u0327', // ̧
	"cent":                                '\u00a2', // ¢
	"centigrade":                          '\u2103', // ℃
	"centinferior":                        '\uf6df',
	"centmonospace":                       '\uffe0', // ￠
	"centoldstyle":                        '\uf7a2',
	"centsuperior":                        '\uf6e0',
	"chaarmenian":                         '\u0579', // չ
	"chabengali":                          '\u099b', // ছ
	"chadeva":                             '\u091b', // छ
	"chagujarati":                         '\u0a9b', // છ
	"chagurmukhi":                         '\u0a1b', // ਛ
	"chbopomofo":                          '\u3114', // ㄔ
	"cheabkhasiancyrillic":                '\u04bd', // ҽ
	"checkmark":                           '\u2713', // ✓
	"checyrillic":                         '\u0447', // ч
	"chedescenderabkhasiancyrillic":       '\u04bf', // ҿ
	"chedescendercyrillic":                '\u04b7', // ҷ
	"chedieresiscyrillic":                 '\u04f5', // ӵ
	"cheharmenian":                        '\u0573', // ճ
	"chekhakassiancyrillic":               '\u04cc', // ӌ
	"cheverticalstrokecyrillic":           '\u04b9', // ҹ
	"chi": '\u03c7', // χ
	"chieuchacirclekorean":                '\u3277', // ㉷
	"chieuchaparenkorean":                 '\u3217', // ㈗
	"chieuchcirclekorean":                 '\u3269', // ㉩
	"chieuchkorean":                       '\u314a', // ㅊ
	"chieuchparenkorean":                  '\u3209', // ㈉
	"chochangthai":                        '\u0e0a', // ช
	"chochanthai":                         '\u0e08', // จ
	"chochingthai":                        '\u0e09', // ฉ
	"chochoethai":                         '\u0e0c', // ฌ
	"chook":                               '\u0188', // ƈ
	"cieucacirclekorean":                  '\u3276', // ㉶
	"cieucaparenkorean":                   '\u3216', // ㈖
	"cieuccirclekorean":                   '\u3268', // ㉨
	"cieuckorean":                         '\u3148', // ㅈ
	"cieucparenkorean":                    '\u3208', // ㈈
	"cieucuparenkorean":                   '\u321c', // ㈜
	"circle":                              '\u25cb', // ○
	"circlemultiply":                      '\u2297', // ⊗
	"circleot":                            '\u2299', // ⊙
	"circleplus":                          '\u2295', // ⊕
	"circlepostalmark":                    '\u3036', // 〶
	"circlewithlefthalfblack":             '\u25d0', // ◐
	"circlewithrighthalfblack":            '\u25d1', // ◑
	"circumflex":                          '\u02c6', // ˆ
	"circumflexbelowcmb":                  '\u032d', // ̭
	"circumflexcmb":                       '\u0302', // ̂
	"clear":                               '\u2327', // ⌧
	"clickalveolar":                       '\u01c2', // ǂ
	"clickdental":                         '\u01c0', // ǀ
	"clicklateral":                        '\u01c1', // ǁ
	"clickretroflex":                      '\u01c3', // ǃ
	"club":                                '\u2663', // ♣
	"clubsuitblack":                       '\u2663', // ♣
	"clubsuitwhite":                       '\u2667', // ♧
	"cmcubedsquare":                       '\u33a4', // ㎤
	"cmonospace":                          '\uff43', // ｃ
	"cmsquaredsquare":                     '\u33a0', // ㎠
	"coarmenian":                          '\u0581', // ց
	"colon":                               ':',      // :
	"colonmonetary":                       '\u20a1', // ₡
	"colonmonospace":                      '\uff1a', // ：
	"colonsign":                           '\u20a1', // ₡
	"colonsmall":                          '\ufe55', // ﹕
	"colontriangularhalfmod":              '\u02d1', // ˑ
	"colontriangularmod":                  '\u02d0', // ː
	"comma":                               ',',      // ,
	"commaabovecmb":                       '\u0313', // ̓
	"commaaboverightcmb":                  '\u0315', // ̕
	"commaaccent":                         '\uf6c3',
	"commaarabic":                         '\u060c', // ،
	"commaarmenian":                       '\u055d', // ՝
	"commainferior":                       '\uf6e1',
	"commamonospace":                      '\uff0c', // ，
	"commareversedabovecmb":               '\u0314', // ̔
	"commareversedmod":                    '\u02bd', // ʽ
	"commasmall":                          '\ufe50', // ﹐
	"commasuperior":                       '\uf6e2',
	"commaturnedabovecmb":                 '\u0312', // ̒
	"commaturnedmod":                      '\u02bb', // ʻ
	"compass":                             '\u263c', // ☼
	"congruent":                           '\u2245', // ≅
	"contourintegral":                     '\u222e', // ∮
	"control":                             '\u2303', // ⌃
	"controlACK":                          '\x06',
	"controlBEL":                          '\a',
	"controlBS":                           '\b',
	"controlCAN":                          '\x18',
	"controlCR":                           '\r',
	"controlDC1":                          '\x11',
	"controlDC2":                          '\x12',
	"controlDC3":                          '\x13',
	"controlDC4":                          '\x14',
	"controlDEL":                          '\u007f',
	"controlDLE":                          '\x10',
	"controlEM":                           '\x19',
	"controlENQ":                          '\x05',
	"controlEOT":                          '\x04',
	"controlESC":                          '\x1b',
	"controlETB":                          '\x17',
	"controlETX":                          '\x03',
	"controlFF":                           '\f',
	"controlFS":                           '\x1c',
	"controlGS":                           '\x1d',
	"controlHT":                           '\t',
	"controlLF":                           '\n',
	"controlNAK":                          '\x15',
	"controlRS":                           '\x1e',
	"controlSI":                           '\x0f',
	"controlSO":                           '\x0e',
	"controlSOT":                          '\x02',
	"controlSTX":                          '\x01',
	"controlSUB":                          '\x1a',
	"controlSYN":                          '\x16',
	"controlUS":                           '\x1f',
	"controlVT":                           '\v',
	"copyright":                           '\u00a9', // ©
	"copyrightsans":                       '\uf8e9',
	"copyrightserif":                      '\uf6d9',
	"cornerbracketleft":                   '\u300c', // 「
	"cornerbracketlefthalfwidth":          '\uff62', // ｢
	"cornerbracketleftvertical":           '\ufe41', // ﹁
	"cornerbracketright":                  '\u300d', // 」
	"cornerbracketrighthalfwidth":         '\uff63', // ｣
	"cornerbracketrightvertical":          '\ufe42', // ﹂
	"corporationsquare":                   '\u337f', // ㍿
	"cosquare":                            '\u33c7', // ㏇
	"coverkgsquare":                       '\u33c6', // ㏆
	"cparen":                              '\u249e', // ⒞
	"cruzeiro":                            '\u20a2', // ₢
	"cstretched":                          '\u0297', // ʗ
	"curlyand":                            '\u22cf', // ⋏
	"curlyor":                             '\u22ce', // ⋎
	"currency":                            '\u00a4', // ¤
	"cyrBreve":                            '\uf6d1',
	"cyrFlex":                             '\uf6d2',
	"cyrbreve":                            '\uf6d4',
	"cyrflex":                             '\uf6d5',
	"d":                                   'd',      // d
	"daarmenian":                          '\u0564', // դ
	"dabengali":                           '\u09a6', // দ
	"dadarabic":                           '\u0636', // ض
	"dadeva":                              '\u0926', // द
	"dadfinalarabic":                      '\ufebe', // ﺾ
	"dadinitialarabic":                    '\ufebf', // ﺿ
	"dadmedialarabic":                     '\ufec0', // ﻀ
	"dagesh":                              '\u05bc', // ּ
	"dageshhebrew":                        '\u05bc', // ּ
	"dagger":                              '\u2020', // †
	"daggerdbl":                           '\u2021', // ‡
	"dagujarati":                          '\u0aa6', // દ
	"dagurmukhi":                          '\u0a26', // ਦ
	"dahiragana":                          '\u3060', // だ
	"dakatakana":                          '\u30c0', // ダ
	"dalarabic":                           '\u062f', // د
	"dalet":                               '\u05d3', // ד
	"daletdagesh":                         '\ufb33', // דּ
	"daletdageshhebrew":                   '\ufb33', // דּ
	"dalethatafpatah":                     '\u05d3', // ד
	"dalethatafpatahhebrew":               '\u05d3', // ד
	"dalethatafsegol":                     '\u05d3', // ד
	"dalethatafsegolhebrew":               '\u05d3', // ד
	"dalethebrew":                         '\u05d3', // ד
	"dalethiriq":                          '\u05d3', // ד
	"dalethiriqhebrew":                    '\u05d3', // ד
	"daletholam":                          '\u05d3', // ד
	"daletholamhebrew":                    '\u05d3', // ד
	"daletpatah":                          '\u05d3', // ד
	"daletpatahhebrew":                    '\u05d3', // ד
	"daletqamats":                         '\u05d3', // ד
	"daletqamatshebrew":                   '\u05d3', // ד
	"daletqubuts":                         '\u05d3', // ד
	"daletqubutshebrew":                   '\u05d3', // ד
	"daletsegol":                          '\u05d3', // ד
	"daletsegolhebrew":                    '\u05d3', // ד
	"daletsheva":                          '\u05d3', // ד
	"daletshevahebrew":                    '\u05d3', // ד
	"dalettsere":                          '\u05d3', // ד
	"dalettserehebrew":                    '\u05d3', // ד
	"dalfinalarabic":                      '\ufeaa', // ﺪ
	"dammaarabic":                         '\u064f', // ُ
	"dammalowarabic":                      '\u064f', // ُ
	"dammatanaltonearabic":                '\u064c', // ٌ
	"dammatanarabic":                      '\u064c', // ٌ
	"danda":                               '\u0964', // ।
	"dargahebrew":                         '\u05a7', // ֧
	"dargalefthebrew":                     '\u05a7', // ֧
	"dasiapneumatacyrilliccmb":            '\u0485', // ҅
	"dblGrave":                            '\uf6d3',
	"dblanglebracketleft":                 '\u300a', // 《
	"dblanglebracketleftvertical":         '\ufe3d', // ︽
	"dblanglebracketright":                '\u300b', // 》
	"dblanglebracketrightvertical":        '\ufe3e', // ︾
	"dblarchinvertedbelowcmb":             '\u032b', // ̫
	"dblarrowleft":                        '\u21d4', // ⇔
	"dblarrowright":                       '\u21d2', // ⇒
	"dbldanda":                            '\u0965', // ॥
	"dblgrave":                            '\uf6d6',
	"dblgravecmb":                         '\u030f', // ̏
	"dblintegral":                         '\u222c', // ∬
	"dbllowline":                          '\u2017', // ‗
	"dbllowlinecmb":                       '\u0333', // ̳
	"dbloverlinecmb":                      '\u033f', // ̿
	"dblprimemod":                         '\u02ba', // ʺ
	"dblverticalbar":                      '\u2016', // ‖
	"dblverticallineabovecmb":             '\u030e', // ̎
	"dbopomofo":                           '\u3109', // ㄉ
	"dbsquare":                            '\u33c8', // ㏈
	"dcaron":                              '\u010f', // ď
	"dcedilla":                            '\u1e11', // ḑ
	"dcircle":                             '\u24d3', // ⓓ
	"dcircumflexbelow":                    '\u1e13', // ḓ
	"dcroat":                              '\u0111', // đ
	"ddabengali":                          '\u09a1', // ড
	"ddadeva":                             '\u0921', // ड
	"ddagujarati":                         '\u0aa1', // ડ
	"ddagurmukhi":                         '\u0a21', // ਡ
	"ddalarabic":                          '\u0688', // ڈ
	"ddalfinalarabic":                     '\ufb89', // ﮉ
	"dddhadeva":                           '\u095c', // ड़
	"ddhabengali":                         '\u09a2', // ঢ
	"ddhadeva":                            '\u0922', // ढ
	"ddhagujarati":                        '\u0aa2', // ઢ
	"ddhagurmukhi":                        '\u0a22', // ਢ
	"ddotaccent":                          '\u1e0b', // ḋ
	"ddotbelow":                           '\u1e0d', // ḍ
	"decimalseparatorarabic":              '\u066b', // ٫
	"decimalseparatorpersian":             '\u066b', // ٫
	"decyrillic":                          '\u0434', // д
	"degree":                              '\u00b0', // °
	"dehihebrew":                          '\u05ad', // ֭
	"dehiragana":                          '\u3067', // で
	"deicoptic":                           '\u03ef', // ϯ
	"dekatakana":                          '\u30c7', // デ
	"deleteleft":                          '\u232b', // ⌫
	"deleteright":                         '\u2326', // ⌦
	"delta":                               '\u03b4', // δ
	"deltaturned":                         '\u018d', // ƍ
	"denominatorminusonenumeratorbengali": '\u09f8', // ৸
	"dezh":                        '\u02a4', // ʤ
	"dhabengali":                  '\u09a7', // ধ
	"dhadeva":                     '\u0927', // ध
	"dhagujarati":                 '\u0aa7', // ધ
	"dhagurmukhi":                 '\u0a27', // ਧ
	"dhook":                       '\u0257', // ɗ
	"dialytikatonos":              '\u0385', // ΅
	"dialytikatonoscmb":           '\u0344', // ̈́
	"diamond":                     '\u2666', // ♦
	"diamondsuitwhite":            '\u2662', // ♢
	"dieresis":                    '\u00a8', // ¨
	"dieresisacute":               '\uf6d7',
	"dieresisbelowcmb":            '\u0324', // ̤
	"dieresiscmb":                 '\u0308', // ̈
	"dieresisgrave":               '\uf6d8',
	"dieresistonos":               '\u0385', // ΅
	"dihiragana":                  '\u3062', // ぢ
	"dikatakana":                  '\u30c2', // ヂ
	"dittomark":                   '\u3003', // 〃
	"divide":                      '\u00f7', // ÷
	"divides":                     '\u2223', // ∣
	"divisionslash":               '\u2215', // ∕
	"djecyrillic":                 '\u0452', // ђ
	"dkshade":                     '\u2593', // ▓
	"dlinebelow":                  '\u1e0f', // ḏ
	"dlsquare":                    '\u3397', // ㎗
	"dmacron":                     '\u0111', // đ
	"dmonospace":                  '\uff44', // ｄ
	"dnblock":                     '\u2584', // ▄
	"dochadathai":                 '\u0e0e', // ฎ
	"dodekthai":                   '\u0e14', // ด
	"dohiragana":                  '\u3069', // ど
	"dokatakana":                  '\u30c9', // ド
	"dollar":                      '$',      // $
	"dollarinferior":              '\uf6e3',
	"dollarmonospace":             '\uff04', // ＄
	"dollaroldstyle":              '\uf724',
	"dollarsmall":                 '\ufe69', // ﹩
	"dollarsuperior":              '\uf6e4',
	"dong":                        '\u20ab', // ₫
	"dorusquare":                  '\u3326', // ㌦
	"dotaccent":                   '\u02d9', // ˙
	"dotaccentcmb":                '\u0307', // ̇
	"dotbelowcmb":                 '\u0323', // ̣
	"dotbelowcomb":                '\u0323', // ̣
	"dotkatakana":                 '\u30fb', // ・
	"dotlessi":                    '\u0131', // ı
	"dotlessj":                    '\uf6be',
	"dotlessjstrokehook":          '\u0284', // ʄ
	"dotmath":                     '\u22c5', // ⋅
	"dottedcircle":                '\u25cc', // ◌
	"doubleyodpatah":              '\ufb1f', // ײַ
	"doubleyodpatahhebrew":        '\ufb1f', // ײַ
	"downtackbelowcmb":            '\u031e', // ̞
	"downtackmod":                 '\u02d5', // ˕
	"dparen":                      '\u249f', // ⒟
	"dsuperior":                   '\uf6eb',
	"dtail":                       '\u0256', // ɖ
	"dtopbar":                     '\u018c', // ƌ
	"duhiragana":                  '\u3065', // づ
	"dukatakana":                  '\u30c5', // ヅ
	"dz":                          '\u01f3', // ǳ
	"dzaltone":                    '\u02a3', // ʣ
	"dzcaron":                     '\u01c6', // ǆ
	"dzcurl":                      '\u02a5', // ʥ
	"dzeabkhasiancyrillic":        '\u04e1', // ӡ
	"dzecyrillic":                 '\u0455', // ѕ
	"dzhecyrillic":                '\u045f', // џ
	"e":                           'e',      // e
	"eacute":                      '\u00e9', // é
	"earth":                       '\u2641', // ♁
	"ebengali":                    '\u098f', // এ
	"ebopomofo":                   '\u311c', // ㄜ
	"ebreve":                      '\u0115', // ĕ
	"ecandradeva":                 '\u090d', // ऍ
	"ecandragujarati":             '\u0a8d', // ઍ
	"ecandravowelsigndeva":        '\u0945', // ॅ
	"ecandravowelsigngujarati":    '\u0ac5', // ૅ
	"ecaron":                      '\u011b', // ě
	"ecedillabreve":               '\u1e1d', // ḝ
	"echarmenian":                 '\u0565', // ե
	"echyiwnarmenian":             '\u0587', // և
	"ecircle":                     '\u24d4', // ⓔ
	"ecircumflex":                 '\u00ea', // ê
	"ecircumflexacute":            '\u1ebf', // ế
	"ecircumflexbelow":            '\u1e19', // ḙ
	"ecircumflexdotbelow":         '\u1ec7', // ệ
	"ecircumflexgrave":            '\u1ec1', // ề
	"ecircumflexhookabove":        '\u1ec3', // ể
	"ecircumflextilde":            '\u1ec5', // ễ
	"ecyrillic":                   '\u0454', // є
	"edblgrave":                   '\u0205', // ȅ
	"edeva":                       '\u090f', // ए
	"edieresis":                   '\u00eb', // ë
	"edot":                        '\u0117', // ė
	"edotaccent":                  '\u0117', // ė
	"edotbelow":                   '\u1eb9', // ẹ
	"eegurmukhi":                  '\u0a0f', // ਏ
	"eematragurmukhi":             '\u0a47', // ੇ
	"efcyrillic":                  '\u0444', // ф
	"egrave":                      '\u00e8', // è
	"egujarati":                   '\u0a8f', // એ
	"eharmenian":                  '\u0567', // է
	"ehbopomofo":                  '\u311d', // ㄝ
	"ehiragana":                   '\u3048', // え
	"ehookabove":                  '\u1ebb', // ẻ
	"eibopomofo":                  '\u311f', // ㄟ
	"eight":                       '8',      // 8
	"eightarabic":                 '\u0668', // ٨
	"eightbengali":                '\u09ee', // ৮
	"eightcircle":                 '\u2467', // ⑧
	"eightcircleinversesansserif": '\u2791', // ➑
	"eightdeva":                   '\u096e', // ८
	"eighteencircle":              '\u2471', // ⑱
	"eighteenparen":               '\u2485', // ⒅
	"eighteenperiod":              '\u2499', // ⒙
	"eightgujarati":               '\u0aee', // ૮
	"eightgurmukhi":               '\u0a6e', // ੮
	"eighthackarabic":             '\u0668', // ٨
	"eighthangzhou":               '\u3028', // 〨
	"eighthnotebeamed":            '\u266b', // ♫
	"eightideographicparen":       '\u3227', // ㈧
	"eightinferior":               '\u2088', // ₈
	"eightmonospace":              '\uff18', // ８
	"eightoldstyle":               '\uf738',
	"eightparen":                  '\u247b', // ⑻
	"eightperiod":                 '\u248f', // ⒏
	"eightpersian":                '\u06f8', // ۸
	"eightroman":                  '\u2177', // ⅷ
	"eightsuperior":               '\u2078', // ⁸
	"eightthai":                   '\u0e58', // ๘
	"einvertedbreve":              '\u0207', // ȇ
	"eiotifiedcyrillic":           '\u0465', // ѥ
	"ekatakana":                   '\u30a8', // エ
	"ekatakanahalfwidth":          '\uff74', // ｴ
	"ekonkargurmukhi":             '\u0a74', // ੴ
	"ekorean":                     '\u3154', // ㅔ
	"elcyrillic":                  '\u043b', // л
	"element":                     '\u2208', // ∈
	"elevencircle":                '\u246a', // ⑪
	"elevenparen":                 '\u247e', // ⑾
	"elevenperiod":                '\u2492', // ⒒
	"elevenroman":                 '\u217a', // ⅺ
	"ellipsis":                    '\u2026', // …
	"ellipsisvertical":            '\u22ee', // ⋮
	"emacron":                     '\u0113', // ē
	"emacronacute":                '\u1e17', // ḗ
	"emacrongrave":                '\u1e15', // ḕ
	"emcyrillic":                  '\u043c', // м
	"emdash":                      '\u2014', // —
	"emdashvertical":              '\ufe31', // ︱
	"emonospace":                  '\uff45', // ｅ
	"emphasismarkarmenian":        '\u055b', // ՛
	"emptyset":                    '\u2205', // ∅
	"enbopomofo":                  '\u3123', // ㄣ
	"encyrillic":                  '\u043d', // н
	"endash":                      '\u2013', // –
	"endashvertical":              '\ufe32', // ︲
	"endescendercyrillic":         '\u04a3', // ң
	"eng":                 '\u014b', // ŋ
	"engbopomofo":         '\u3125', // ㄥ
	"enghecyrillic":       '\u04a5', // ҥ
	"enhookcyrillic":      '\u04c8', // ӈ
	"enspace":             '\u2002',
	"eogonek":             '\u0119', // ę
	"eokorean":            '\u3153', // ㅓ
	"eopen":               '\u025b', // ɛ
	"eopenclosed":         '\u029a', // ʚ
	"eopenreversed":       '\u025c', // ɜ
	"eopenreversedclosed": '\u025e', // ɞ
	"eopenreversedhook":   '\u025d', // ɝ
	"eparen":              '\u24a0', // ⒠
	"epsilon":             '\u03b5', // ε
	"epsilontonos":        '\u03ad', // έ
	"equal":               '=',      // =
	"equalmonospace":      '\uff1d', // ＝
	"equalsmall":          '\ufe66', // ﹦
	"equalsuperior":       '\u207c', // ⁼
	"equivalence":         '\u2261', // ≡
	"erbopomofo":          '\u3126', // ㄦ
	"ercyrillic":          '\u0440', // р
	"ereversed":           '\u0258', // ɘ
	"ereversedcyrillic":   '\u044d', // э
	"escyrillic":          '\u0441', // с
	"esdescendercyrillic": '\u04ab', // ҫ
	"esh":                         '\u0283', // ʃ
	"eshcurl":                     '\u0286', // ʆ
	"eshortdeva":                  '\u090e', // ऎ
	"eshortvowelsigndeva":         '\u0946', // ॆ
	"eshreversedloop":             '\u01aa', // ƪ
	"eshsquatreversed":            '\u0285', // ʅ
	"esmallhiragana":              '\u3047', // ぇ
	"esmallkatakana":              '\u30a7', // ェ
	"esmallkatakanahalfwidth":     '\uff6a', // ｪ
	"estimated":                   '\u212e', // ℮
	"esuperior":                   '\uf6ec',
	"eta":                         '\u03b7', // η
	"etarmenian":                  '\u0568', // ը
	"etatonos":                    '\u03ae', // ή
	"eth":                         '\u00f0', // ð
	"etilde":                      '\u1ebd', // ẽ
	"etildebelow":                 '\u1e1b', // ḛ
	"etnahtafoukhhebrew":          '\u0591', // ֑
	"etnahtafoukhlefthebrew":      '\u0591', // ֑
	"etnahtahebrew":               '\u0591', // ֑
	"etnahtalefthebrew":           '\u0591', // ֑
	"eturned":                     '\u01dd', // ǝ
	"eukorean":                    '\u3161', // ㅡ
	"euro":                        '\u20ac', // €
	"evowelsignbengali":           '\u09c7', // ে
	"evowelsigndeva":              '\u0947', // े
	"evowelsigngujarati":          '\u0ac7', // ે
	"exclam":                      '!',      // !
	"exclamarmenian":              '\u055c', // ՜
	"exclamdbl":                   '\u203c', // ‼
	"exclamdown":                  '\u00a1', // ¡
	"exclamdownsmall":             '\uf7a1',
	"exclammonospace":             '\uff01', // ！
	"exclamsmall":                 '\uf721',
	"existential":                 '\u2203', // ∃
	"ezh":                         '\u0292', // ʒ
	"ezhcaron":                    '\u01ef', // ǯ
	"ezhcurl":                     '\u0293', // ʓ
	"ezhreversed":                 '\u01b9', // ƹ
	"ezhtail":                     '\u01ba', // ƺ
	"f":                           'f',      // f
	"fadeva":                      '\u095e', // फ़
	"fagurmukhi":                  '\u0a5e', // ਫ਼
	"fahrenheit":                  '\u2109', // ℉
	"fathaarabic":                 '\u064e', // َ
	"fathalowarabic":              '\u064e', // َ
	"fathatanarabic":              '\u064b', // ً
	"fbopomofo":                   '\u3108', // ㄈ
	"fcircle":                     '\u24d5', // ⓕ
	"fdotaccent":                  '\u1e1f', // ḟ
	"feharabic":                   '\u0641', // ف
	"feharmenian":                 '\u0586', // ֆ
	"fehfinalarabic":              '\ufed2', // ﻒ
	"fehinitialarabic":            '\ufed3', // ﻓ
	"fehmedialarabic":             '\ufed4', // ﻔ
	"feicoptic":                   '\u03e5', // ϥ
	"female":                      '\u2640', // ♀
	"ff":                          '\ufb00', // ﬀ
	"ffi":                         '\ufb03', // ﬃ
	"ffl":                         '\ufb04', // ﬄ
	"fi":                          '\ufb01', // ﬁ
	"fifteencircle":               '\u246e', // ⑮
	"fifteenparen":                '\u2482', // ⒂
	"fifteenperiod":               '\u2496', // ⒖
	"figuredash":                  '\u2012', // ‒
	"filledbox":                   '\u25a0', // ■
	"filledrect":                  '\u25ac', // ▬
	"finalkaf":                    '\u05da', // ך
	"finalkafdagesh":              '\ufb3a', // ךּ
	"finalkafdageshhebrew":        '\ufb3a', // ךּ
	"finalkafhebrew":              '\u05da', // ך
	"finalkafqamats":              '\u05da', // ך
	"finalkafqamatshebrew":        '\u05da', // ך
	"finalkafsheva":               '\u05da', // ך
	"finalkafshevahebrew":         '\u05da', // ך
	"finalmem":                    '\u05dd', // ם
	"finalmemhebrew":              '\u05dd', // ם
	"finalnun":                    '\u05df', // ן
	"finalnunhebrew":              '\u05df', // ן
	"finalpe":                     '\u05e3', // ף
	"finalpehebrew":               '\u05e3', // ף
	"finaltsadi":                  '\u05e5', // ץ
	"finaltsadihebrew":            '\u05e5', // ץ
	"firsttonechinese":            '\u02c9', // ˉ
	"fisheye":                     '\u25c9', // ◉
	"fitacyrillic":                '\u0473', // ѳ
	"five":                        '5',      // 5
	"fivearabic":                  '\u0665', // ٥
	"fivebengali":                 '\u09eb', // ৫
	"fivecircle":                  '\u2464', // ⑤
	"fivecircleinversesansserif":  '\u278e', // ➎
	"fivedeva":                    '\u096b', // ५
	"fiveeighths":                 '\u215d', // ⅝
	"fivegujarati":                '\u0aeb', // ૫
	"fivegurmukhi":                '\u0a6b', // ੫
	"fivehackarabic":              '\u0665', // ٥
	"fivehangzhou":                '\u3025', // 〥
	"fiveideographicparen":        '\u3224', // ㈤
	"fiveinferior":                '\u2085', // ₅
	"fivemonospace":               '\uff15', // ５
	"fiveoldstyle":                '\uf735',
	"fiveparen":                   '\u2478', // ⑸
	"fiveperiod":                  '\u248c', // ⒌
	"fivepersian":                 '\u06f5', // ۵
	"fiveroman":                   '\u2174', // ⅴ
	"fivesuperior":                '\u2075', // ⁵
	"fivethai":                    '\u0e55', // ๕
	"fl":                          '\ufb02', // ﬂ
	"florin":                      '\u0192', // ƒ
	"fmonospace":                  '\uff46', // ｆ
	"fmsquare":                    '\u3399', // ㎙
	"fofanthai":                   '\u0e1f', // ฟ
	"fofathai":                    '\u0e1d', // ฝ
	"fongmanthai":                 '\u0e4f', // ๏
	"forall":                      '\u2200', // ∀
	"four":                        '4',      // 4
	"fourarabic":                  '\u0664', // ٤
	"fourbengali":                 '\u09ea', // ৪
	"fourcircle":                  '\u2463', // ④
	"fourcircleinversesansserif":  '\u278d', // ➍
	"fourdeva":                    '\u096a', // ४
	"fourgujarati":                '\u0aea', // ૪
	"fourgurmukhi":                '\u0a6a', // ੪
	"fourhackarabic":              '\u0664', // ٤
	"fourhangzhou":                '\u3024', // 〤
	"fourideographicparen":        '\u3223', // ㈣
	"fourinferior":                '\u2084', // ₄
	"fourmonospace":               '\uff14', // ４
	"fournumeratorbengali":        '\u09f7', // ৷
	"fouroldstyle":                '\uf734',
	"fourparen":                   '\u2477', // ⑷
	"fourperiod":                  '\u248b', // ⒋
	"fourpersian":                 '\u06f4', // ۴
	"fourroman":                   '\u2173', // ⅳ
	"foursuperior":                '\u2074', // ⁴
	"fourteencircle":              '\u246d', // ⑭
	"fourteenparen":               '\u2481', // ⒁
	"fourteenperiod":              '\u2495', // ⒕
	"fourthai":                    '\u0e54', // ๔
	"fourthtonechinese":           '\u02cb', // ˋ
	"fparen":                      '\u24a1', // ⒡
	"fraction":                    '\u2044', // ⁄
	"franc":                       '\u20a3', // ₣
	"g":                           'g',      // g
	"gabengali":                   '\u0997', // গ
	"gacute":                      '\u01f5', // ǵ
	"gadeva":                      '\u0917', // ग
	"gafarabic":                   '\u06af', // گ
	"gaffinalarabic":              '\ufb93', // ﮓ
	"gafinitialarabic":            '\ufb94', // ﮔ
	"gafmedialarabic":             '\ufb95', // ﮕ
	"gagujarati":                  '\u0a97', // ગ
	"gagurmukhi":                  '\u0a17', // ਗ
	"gahiragana":                  '\u304c', // が
	"gakatakana":                  '\u30ac', // ガ
	"gamma":                       '\u03b3', // γ
	"gammalatinsmall":             '\u0263', // ɣ
	"gammasuperior":               '\u02e0', // ˠ
	"gangiacoptic":                '\u03eb', // ϫ
	"gbopomofo":                   '\u310d', // ㄍ
	"gbreve":                      '\u011f', // ğ
	"gcaron":                      '\u01e7', // ǧ
	"gcedilla":                    '\u0123', // ģ
	"gcircle":                     '\u24d6', // ⓖ
	"gcircumflex":                 '\u011d', // ĝ
	"gcommaaccent":                '\u0123', // ģ
	"gdot":                        '\u0121', // ġ
	"gdotaccent":                  '\u0121', // ġ
	"gecyrillic":                  '\u0433', // г
	"gehiragana":                  '\u3052', // げ
	"gekatakana":                  '\u30b2', // ゲ
	"geometricallyequal":          '\u2251', // ≑
	"gereshaccenthebrew":          '\u059c', // ֜
	"gereshhebrew":                '\u05f3', // ׳
	"gereshmuqdamhebrew":          '\u059d', // ֝
	"germandbls":                  '\u00df', // ß
	"gershayimaccenthebrew":       '\u059e', // ֞
	"gershayimhebrew":             '\u05f4', // ״
	"getamark":                    '\u3013', // 〓
	"ghabengali":                  '\u0998', // ঘ
	"ghadarmenian":                '\u0572', // ղ
	"ghadeva":                     '\u0918', // घ
	"ghagujarati":                 '\u0a98', // ઘ
	"ghagurmukhi":                 '\u0a18', // ਘ
	"ghainarabic":                 '\u063a', // غ
	"ghainfinalarabic":            '\ufece', // ﻎ
	"ghaininitialarabic":          '\ufecf', // ﻏ
	"ghainmedialarabic":           '\ufed0', // ﻐ
	"ghemiddlehookcyrillic":       '\u0495', // ҕ
	"ghestrokecyrillic":           '\u0493', // ғ
	"gheupturncyrillic":           '\u0491', // ґ
	"ghhadeva":                    '\u095a', // ग़
	"ghhagurmukhi":                '\u0a5a', // ਗ਼
	"ghook":                       '\u0260', // ɠ
	"ghzsquare":                   '\u3393', // ㎓
	"gihiragana":                  '\u304e', // ぎ
	"gikatakana":                  '\u30ae', // ギ
	"gimarmenian":                 '\u0563', // գ
	"gimel":                       '\u05d2', // ג
	"gimeldagesh":                 '\ufb32', // גּ
	"gimeldageshhebrew":           '\ufb32', // גּ
	"gimelhebrew":                 '\u05d2', // ג
	"gjecyrillic":                 '\u0453', // ѓ
	"glottalinvertedstroke":       '\u01be', // ƾ
	"glottalstop":                 '\u0294', // ʔ
	"glottalstopinverted":         '\u0296', // ʖ
	"glottalstopmod":              '\u02c0', // ˀ
	"glottalstopreversed":         '\u0295', // ʕ
	"glottalstopreversedmod":      '\u02c1', // ˁ
	"glottalstopreversedsuperior": '\u02e4', // ˤ
	"glottalstopstroke":           '\u02a1', // ʡ
	"glottalstopstrokereversed":   '\u02a2', // ʢ
	"gmacron":                     '\u1e21', // ḡ
	"gmonospace":                  '\uff47', // ｇ
	"gohiragana":                  '\u3054', // ご
	"gokatakana":                  '\u30b4', // ゴ
	"gparen":                      '\u24a2', // ⒢
	"gpasquare":                   '\u33ac', // ㎬
	"gradient":                    '\u2207', // ∇
	"grave":                       '`',      // `
	"gravebelowcmb":               '\u0316', // ̖
	"gravecmb":                    '\u0300', // ̀
	"gravecomb":                   '\u0300', // ̀
	"gravedeva":                   '\u0953', // ॓
	"gravelowmod":                 '\u02ce', // ˎ
	"gravemonospace":              '\uff40', // ｀
	"gravetonecmb":                '\u0340', // ̀
	"greater":                     '>',      // >
	"greaterequal":                '\u2265', // ≥
	"greaterequalorless":          '\u22db', // ⋛
	"greatermonospace":            '\uff1e', // ＞
	"greaterorequivalent":         '\u2273', // ≳
	"greaterorless":               '\u2277', // ≷
	"greateroverequal":            '\u2267', // ≧
	"greatersmall":                '\ufe65', // ﹥
	"gscript":                     '\u0261', // ɡ
	"gstroke":                     '\u01e5', // ǥ
	"guhiragana":                  '\u3050', // ぐ
	"guillemotleft":               '\u00ab', // «
	"guillemotright":              '\u00bb', // »
	"guilsinglleft":               '\u2039', // ‹
	"guilsinglright":              '\u203a', // ›
	"gukatakana":                  '\u30b0', // グ
	"guramusquare":                '\u3318', // ㌘
	"gysquare":                    '\u33c9', // ㏉
	"h":                           'h',      // h
	"haabkhasiancyrillic":            '\u04a9', // ҩ
	"haaltonearabic":                 '\u06c1', // ہ
	"habengali":                      '\u09b9', // হ
	"hadescendercyrillic":            '\u04b3', // ҳ
	"hadeva":                         '\u0939', // ह
	"hagujarati":                     '\u0ab9', // હ
	"hagurmukhi":                     '\u0a39', // ਹ
	"haharabic":                      '\u062d', // ح
	"hahfinalarabic":                 '\ufea2', // ﺢ
	"hahinitialarabic":               '\ufea3', // ﺣ
	"hahiragana":                     '\u306f', // は
	"hahmedialarabic":                '\ufea4', // ﺤ
	"haitusquare":                    '\u332a', // ㌪
	"hakatakana":                     '\u30cf', // ハ
	"hakatakanahalfwidth":            '\uff8a', // ﾊ
	"halantgurmukhi":                 '\u0a4d', // ੍
	"hamzaarabic":                    '\u0621', // ء
	"hamzadammaarabic":               '\u0621', // ء
	"hamzadammatanarabic":            '\u0621', // ء
	"hamzafathaarabic":               '\u0621', // ء
	"hamzafathatanarabic":            '\u0621', // ء
	"hamzalowarabic":                 '\u0621', // ء
	"hamzalowkasraarabic":            '\u0621', // ء
	"hamzalowkasratanarabic":         '\u0621', // ء
	"hamzasukunarabic":               '\u0621', // ء
	"hangulfiller":                   '\u3164', // ㅤ
	"hardsigncyrillic":               '\u044a', // ъ
	"harpoonleftbarbup":              '\u21bc', // ↼
	"harpoonrightbarbup":             '\u21c0', // ⇀
	"hasquare":                       '\u33ca', // ㏊
	"hatafpatah":                     '\u05b2', // ֲ
	"hatafpatah16":                   '\u05b2', // ֲ
	"hatafpatah23":                   '\u05b2', // ֲ
	"hatafpatah2f":                   '\u05b2', // ֲ
	"hatafpatahhebrew":               '\u05b2', // ֲ
	"hatafpatahnarrowhebrew":         '\u05b2', // ֲ
	"hatafpatahquarterhebrew":        '\u05b2', // ֲ
	"hatafpatahwidehebrew":           '\u05b2', // ֲ
	"hatafqamats":                    '\u05b3', // ֳ
	"hatafqamats1b":                  '\u05b3', // ֳ
	"hatafqamats28":                  '\u05b3', // ֳ
	"hatafqamats34":                  '\u05b3', // ֳ
	"hatafqamatshebrew":              '\u05b3', // ֳ
	"hatafqamatsnarrowhebrew":        '\u05b3', // ֳ
	"hatafqamatsquarterhebrew":       '\u05b3', // ֳ
	"hatafqamatswidehebrew":          '\u05b3', // ֳ
	"hatafsegol":                     '\u05b1', // ֱ
	"hatafsegol17":                   '\u05b1', // ֱ
	"hatafsegol24":                   '\u05b1', // ֱ
	"hatafsegol30":                   '\u05b1', // ֱ
	"hatafsegolhebrew":               '\u05b1', // ֱ
	"hatafsegolnarrowhebrew":         '\u05b1', // ֱ
	"hatafsegolquarterhebrew":        '\u05b1', // ֱ
	"hatafsegolwidehebrew":           '\u05b1', // ֱ
	"hbar":                           '\u0127', // ħ
	"hbopomofo":                      '\u310f', // ㄏ
	"hbrevebelow":                    '\u1e2b', // ḫ
	"hcedilla":                       '\u1e29', // ḩ
	"hcircle":                        '\u24d7', // ⓗ
	"hcircumflex":                    '\u0125', // ĥ
	"hdieresis":                      '\u1e27', // ḧ
	"hdotaccent":                     '\u1e23', // ḣ
	"hdotbelow":                      '\u1e25', // ḥ
	"he":                             '\u05d4', // ה
	"heart":                          '\u2665', // ♥
	"heartsuitblack":                 '\u2665', // ♥
	"heartsuitwhite":                 '\u2661', // ♡
	"hedagesh":                       '\ufb34', // הּ
	"hedageshhebrew":                 '\ufb34', // הּ
	"hehaltonearabic":                '\u06c1', // ہ
	"heharabic":                      '\u0647', // ه
	"hehebrew":                       '\u05d4', // ה
	"hehfinalaltonearabic":           '\ufba7', // ﮧ
	"hehfinalalttwoarabic":           '\ufeea', // ﻪ
	"hehfinalarabic":                 '\ufeea', // ﻪ
	"hehhamzaabovefinalarabic":       '\ufba5', // ﮥ
	"hehhamzaaboveisolatedarabic":    '\ufba4', // ﮤ
	"hehinitialaltonearabic":         '\ufba8', // ﮨ
	"hehinitialarabic":               '\ufeeb', // ﻫ
	"hehiragana":                     '\u3078', // へ
	"hehmedialaltonearabic":          '\ufba9', // ﮩ
	"hehmedialarabic":                '\ufeec', // ﻬ
	"heiseierasquare":                '\u337b', // ㍻
	"hekatakana":                     '\u30d8', // ヘ
	"hekatakanahalfwidth":            '\uff8d', // ﾍ
	"hekutaarusquare":                '\u3336', // ㌶
	"henghook":                       '\u0267', // ɧ
	"herutusquare":                   '\u3339', // ㌹
	"het":                            '\u05d7', // ח
	"hethebrew":                      '\u05d7', // ח
	"hhook":                          '\u0266', // ɦ
	"hhooksuperior":                  '\u02b1', // ʱ
	"hieuhacirclekorean":             '\u327b', // ㉻
	"hieuhaparenkorean":              '\u321b', // ㈛
	"hieuhcirclekorean":              '\u326d', // ㉭
	"hieuhkorean":                    '\u314e', // ㅎ
	"hieuhparenkorean":               '\u320d', // ㈍
	"hihiragana":                     '\u3072', // ひ
	"hikatakana":                     '\u30d2', // ヒ
	"hikatakanahalfwidth":            '\uff8b', // ﾋ
	"hiriq":                          '\u05b4', // ִ
	"hiriq14":                        '\u05b4', // ִ
	"hiriq21":                        '\u05b4', // ִ
	"hiriq2d":                        '\u05b4', // ִ
	"hiriqhebrew":                    '\u05b4', // ִ
	"hiriqnarrowhebrew":              '\u05b4', // ִ
	"hiriqquarterhebrew":             '\u05b4', // ִ
	"hiriqwidehebrew":                '\u05b4', // ִ
	"hlinebelow":                     '\u1e96', // ẖ
	"hmonospace":                     '\uff48', // ｈ
	"hoarmenian":                     '\u0570', // հ
	"hohipthai":                      '\u0e2b', // ห
	"hohiragana":                     '\u307b', // ほ
	"hokatakana":                     '\u30db', // ホ
	"hokatakanahalfwidth":            '\uff8e', // ﾎ
	"holam":                          '\u05b9', // ֹ
	"holam19":                        '\u05b9', // ֹ
	"holam26":                        '\u05b9', // ֹ
	"holam32":                        '\u05b9', // ֹ
	"holamhebrew":                    '\u05b9', // ֹ
	"holamnarrowhebrew":              '\u05b9', // ֹ
	"holamquarterhebrew":             '\u05b9', // ֹ
	"holamwidehebrew":                '\u05b9', // ֹ
	"honokhukthai":                   '\u0e2e', // ฮ
	"hookabovecomb":                  '\u0309', // ̉
	"hookcmb":                        '\u0309', // ̉
	"hookpalatalizedbelowcmb":        '\u0321', // ̡
	"hookretroflexbelowcmb":          '\u0322', // ̢
	"hoonsquare":                     '\u3342', // ㍂
	"horicoptic":                     '\u03e9', // ϩ
	"horizontalbar":                  '\u2015', // ―
	"horncmb":                        '\u031b', // ̛
	"hotsprings":                     '\u2668', // ♨
	"house":                          '\u2302', // ⌂
	"hparen":                         '\u24a3', // ⒣
	"hsuperior":                      '\u02b0', // ʰ
	"hturned":                        '\u0265', // ɥ
	"huhiragana":                     '\u3075', // ふ
	"huiitosquare":                   '\u3333', // ㌳
	"hukatakana":                     '\u30d5', // フ
	"hukatakanahalfwidth":            '\uff8c', // ﾌ
	"hungarumlaut":                   '\u02dd', // ˝
	"hungarumlautcmb":                '\u030b', // ̋
	"hv":                             '\u0195', // ƕ
	"hyphen":                         '-',      // -
	"hypheninferior":                 '\uf6e5',
	"hyphenmonospace":                '\uff0d', // －
	"hyphensmall":                    '\ufe63', // ﹣
	"hyphensuperior":                 '\uf6e6',
	"hyphentwo":                      '\u2010', // ‐
	"i":                              'i',      // i
	"iacute":                         '\u00ed', // í
	"iacyrillic":                     '\u044f', // я
	"ibengali":                       '\u0987', // ই
	"ibopomofo":                      '\u3127', // ㄧ
	"ibreve":                         '\u012d', // ĭ
	"icaron":                         '\u01d0', // ǐ
	"icircle":                        '\u24d8', // ⓘ
	"icircumflex":                    '\u00ee', // î
	"icyrillic":                      '\u0456', // і
	"idblgrave":                      '\u0209', // ȉ
	"ideographearthcircle":           '\u328f', // ㊏
	"ideographfirecircle":            '\u328b', // ㊋
	"ideographicallianceparen":       '\u323f', // ㈿
	"ideographiccallparen":           '\u323a', // ㈺
	"ideographiccentrecircle":        '\u32a5', // ㊥
	"ideographicclose":               '\u3006', // 〆
	"ideographiccomma":               '\u3001', // 、
	"ideographiccommaleft":           '\uff64', // ､
	"ideographiccongratulationparen": '\u3237', // ㈷
	"ideographiccorrectcircle":       '\u32a3', // ㊣
	"ideographicearthparen":          '\u322f', // ㈯
	"ideographicenterpriseparen":     '\u323d', // ㈽
	"ideographicexcellentcircle":     '\u329d', // ㊝
	"ideographicfestivalparen":       '\u3240', // ㉀
	"ideographicfinancialcircle":     '\u3296', // ㊖
	"ideographicfinancialparen":      '\u3236', // ㈶
	"ideographicfireparen":           '\u322b', // ㈫
	"ideographichaveparen":           '\u3232', // ㈲
	"ideographichighcircle":          '\u32a4', // ㊤
	"ideographiciterationmark":       '\u3005', // 々
	"ideographiclaborcircle":         '\u3298', // ㊘
	"ideographiclaborparen":          '\u3238', // ㈸
	"ideographicleftcircle":          '\u32a7', // ㊧
	"ideographiclowcircle":           '\u32a6', // ㊦
	"ideographicmedicinecircle":      '\u32a9', // ㊩
	"ideographicmetalparen":          '\u322e', // ㈮
	"ideographicmoonparen":           '\u322a', // ㈪
	"ideographicnameparen":           '\u3234', // ㈴
	"ideographicperiod":              '\u3002', // 。
	"ideographicprintcircle":         '\u329e', // ㊞
	"ideographicreachparen":          '\u3243', // ㉃
	"ideographicrepresentparen":      '\u3239', // ㈹
	"ideographicresourceparen":       '\u323e', // ㈾
	"ideographicrightcircle":         '\u32a8', // ㊨
	"ideographicsecretcircle":        '\u3299', // ㊙
	"ideographicselfparen":           '\u3242', // ㉂
	"ideographicsocietyparen":        '\u3233', // ㈳
	"ideographicspace":               '\u3000',
	"ideographicspecialparen":        '\u3235', // ㈵
	"ideographicstockparen":          '\u3231', // ㈱
	"ideographicstudyparen":          '\u323b', // ㈻
	"ideographicsunparen":            '\u3230', // ㈰
	"ideographicsuperviseparen":      '\u323c', // ㈼
	"ideographicwaterparen":          '\u322c', // ㈬
	"ideographicwoodparen":           '\u322d', // ㈭
	"ideographiczero":                '\u3007', // 〇
	"ideographmetalcircle":           '\u328e', // ㊎
	"ideographmooncircle":            '\u328a', // ㊊
	"ideographnamecircle":            '\u3294', // ㊔
	"ideographsuncircle":             '\u3290', // ㊐
	"ideographwatercircle":           '\u328c', // ㊌
	"ideographwoodcircle":            '\u328d', // ㊍
	"ideva":                          '\u0907', // इ
	"idieresis":                      '\u00ef', // ï
	"idieresisacute":                 '\u1e2f', // ḯ
	"idieresiscyrillic":              '\u04e5', // ӥ
	"idotbelow":                      '\u1ecb', // ị
	"iebrevecyrillic":                '\u04d7', // ӗ
	"iecyrillic":                     '\u0435', // е
	"ieungacirclekorean":             '\u3275', // ㉵
	"ieungaparenkorean":              '\u3215', // ㈕
	"ieungcirclekorean":              '\u3267', // ㉧
	"ieungkorean":                    '\u3147', // ㅇ
	"ieungparenkorean":               '\u3207', // ㈇
	"igrave":                         '\u00ec', // ì
	"igujarati":                      '\u0a87', // ઇ
	"igurmukhi":                      '\u0a07', // ਇ
	"ihiragana":                      '\u3044', // い
	"ihookabove":                     '\u1ec9', // ỉ
	"iibengali":                      '\u0988', // ঈ
	"iicyrillic":                     '\u0438', // и
	"iideva":                         '\u0908', // ई
	"iigujarati":                     '\u0a88', // ઈ
	"iigurmukhi":                     '\u0a08', // ਈ
	"iimatragurmukhi":                '\u0a40', // ੀ
	"iinvertedbreve":                 '\u020b', // ȋ
	"iishortcyrillic":                '\u0439', // й
	"iivowelsignbengali":             '\u09c0', // ী
	"iivowelsigndeva":                '\u0940', // ी
	"iivowelsigngujarati":            '\u0ac0', // ી
	"ij":                        '\u0133', // ĳ
	"ikatakana":                 '\u30a4', // イ
	"ikatakanahalfwidth":        '\uff72', // ｲ
	"ikorean":                   '\u3163', // ㅣ
	"ilde":                      '\u02dc', // ˜
	"iluyhebrew":                '\u05ac', // ֬
	"imacron":                   '\u012b', // ī
	"imacroncyrillic":           '\u04e3', // ӣ
	"imageorapproximatelyequal": '\u2253', // ≓
	"imatragurmukhi":            '\u0a3f', // ਿ
	"imonospace":                '\uff49', // ｉ
	"increment":                 '\u2206', // ∆
	"infinity":                  '\u221e', // ∞
	"iniarmenian":               '\u056b', // ի
	"integral":                  '\u222b', // ∫
	"integralbottom":            '\u2321', // ⌡
	"integralbt":                '\u2321', // ⌡
	"integralex":                '\uf8f5',
	"integraltop":               '\u2320', // ⌠
	"integraltp":                '\u2320', // ⌠
	"intersection":              '\u2229', // ∩
	"intisquare":                '\u3305', // ㌅
	"invbullet":                 '\u25d8', // ◘
	"invcircle":                 '\u25d9', // ◙
	"invsmileface":              '\u263b', // ☻
	"iocyrillic":                '\u0451', // ё
	"iogonek":                   '\u012f', // į
	"iota":                      '\u03b9', // ι
	"iotadieresis":              '\u03ca', // ϊ
	"iotadieresistonos":         '\u0390', // ΐ
	"iotalatin":                 '\u0269', // ɩ
	"iotatonos":                 '\u03af', // ί
	"iparen":                    '\u24a4', // ⒤
	"irigurmukhi":               '\u0a72', // ੲ
	"ismallhiragana":            '\u3043', // ぃ
	"ismallkatakana":            '\u30a3', // ィ
	"ismallkatakanahalfwidth":   '\uff68', // ｨ
	"issharbengali":             '\u09fa', // ৺
	"istroke":                   '\u0268', // ɨ
	"isuperior":                 '\uf6ed',
	"iterationhiragana":         '\u309d', // ゝ
	"iterationkatakana":         '\u30fd', // ヽ
	"itilde":                    '\u0129', // ĩ
	"itildebelow":               '\u1e2d', // ḭ
	"iubopomofo":                '\u3129', // ㄩ
	"iucyrillic":                '\u044e', // ю
	"ivowelsignbengali":         '\u09bf', // ি
	"ivowelsigndeva":            '\u093f', // ि
	"ivowelsigngujarati":        '\u0abf', // િ
	"izhitsacyrillic":           '\u0475', // ѵ
	"izhitsadblgravecyrillic":   '\u0477', // ѷ
	"j":                               'j',      // j
	"jaarmenian":                      '\u0571', // ձ
	"jabengali":                       '\u099c', // জ
	"jadeva":                          '\u091c', // ज
	"jagujarati":                      '\u0a9c', // જ
	"jagurmukhi":                      '\u0a1c', // ਜ
	"jbopomofo":                       '\u3110', // ㄐ
	"jcaron":                          '\u01f0', // ǰ
	"jcircle":                         '\u24d9', // ⓙ
	"jcircumflex":                     '\u0135', // ĵ
	"jcrossedtail":                    '\u029d', // ʝ
	"jdotlessstroke":                  '\u025f', // ɟ
	"jecyrillic":                      '\u0458', // ј
	"jeemarabic":                      '\u062c', // ج
	"jeemfinalarabic":                 '\ufe9e', // ﺞ
	"jeeminitialarabic":               '\ufe9f', // ﺟ
	"jeemmedialarabic":                '\ufea0', // ﺠ
	"jeharabic":                       '\u0698', // ژ
	"jehfinalarabic":                  '\ufb8b', // ﮋ
	"jhabengali":                      '\u099d', // ঝ
	"jhadeva":                         '\u091d', // झ
	"jhagujarati":                     '\u0a9d', // ઝ
	"jhagurmukhi":                     '\u0a1d', // ਝ
	"jheharmenian":                    '\u057b', // ջ
	"jis":                             '\u3004', // 〄
	"jmonospace":                      '\uff4a', // ｊ
	"jparen":                          '\u24a5', // ⒥
	"jsuperior":                       '\u02b2', // ʲ
	"k":                               'k',      // k
	"kabashkircyrillic":               '\u04a1', // ҡ
	"kabengali":                       '\u0995', // ক
	"kacute":                          '\u1e31', // ḱ
	"kacyrillic":                      '\u043a', // к
	"kadescendercyrillic":             '\u049b', // қ
	"kadeva":                          '\u0915', // क
	"kaf":                             '\u05db', // כ
	"kafarabic":                       '\u0643', // ك
	"kafdagesh":                       '\ufb3b', // כּ
	"kafdageshhebrew":                 '\ufb3b', // כּ
	"kaffinalarabic":                  '\ufeda', // ﻚ
	"kafhebrew":                       '\u05db', // כ
	"kafinitialarabic":                '\ufedb', // ﻛ
	"kafmedialarabic":                 '\ufedc', // ﻜ
	"kafrafehebrew":                   '\ufb4d', // כֿ
	"kagujarati":                      '\u0a95', // ક
	"kagurmukhi":                      '\u0a15', // ਕ
	"kahiragana":                      '\u304b', // か
	"kahookcyrillic":                  '\u04c4', // ӄ
	"kakatakana":                      '\u30ab', // カ
	"kakatakanahalfwidth":             '\uff76', // ｶ
	"kappa":                           '\u03ba', // κ
	"kappasymbolgreek":                '\u03f0', // ϰ
	"kapyeounmieumkorean":             '\u3171', // ㅱ
	"kapyeounphieuphkorean":           '\u3184', // ㆄ
	"kapyeounpieupkorean":             '\u3178', // ㅸ
	"kapyeounssangpieupkorean":        '\u3179', // ㅹ
	"karoriisquare":                   '\u330d', // ㌍
	"kashidaautoarabic":               '\u0640', // ـ
	"kashidaautonosidebearingarabic":  '\u0640', // ـ
	"kasmallkatakana":                 '\u30f5', // ヵ
	"kasquare":                        '\u3384', // ㎄
	"kasraarabic":                     '\u0650', // ِ
	"kasratanarabic":                  '\u064d', // ٍ
	"kastrokecyrillic":                '\u049f', // ҟ
	"katahiraprolongmarkhalfwidth":    '\uff70', // ｰ
	"kaverticalstrokecyrillic":        '\u049d', // ҝ
	"kbopomofo":                       '\u310e', // ㄎ
	"kcalsquare":                      '\u3389', // ㎉
	"kcaron":                          '\u01e9', // ǩ
	"kcedilla":                        '\u0137', // ķ
	"kcircle":                         '\u24da', // ⓚ
	"kcommaaccent":                    '\u0137', // ķ
	"kdotbelow":                       '\u1e33', // ḳ
	"keharmenian":                     '\u0584', // ք
	"kehiragana":                      '\u3051', // け
	"kekatakana":                      '\u30b1', // ケ
	"kekatakanahalfwidth":             '\uff79', // ｹ
	"kenarmenian":                     '\u056f', // կ
	"kesmallkatakana":                 '\u30f6', // ヶ
	"kgreenlandic":                    '\u0138', // ĸ
	"khabengali":                      '\u0996', // খ
	"khacyrillic":                     '\u0445', // х
	"khadeva":                         '\u0916', // ख
	"khagujarati":                     '\u0a96', // ખ
	"khagurmukhi":                     '\u0a16', // ਖ
	"khaharabic":                      '\u062e', // خ
	"khahfinalarabic":                 '\ufea6', // ﺦ
	"khahinitialarabic":               '\ufea7', // ﺧ
	"khahmedialarabic":                '\ufea8', // ﺨ
	"kheicoptic":                      '\u03e7', // ϧ
	"khhadeva":                        '\u0959', // ख़
	"khhagurmukhi":                    '\u0a59', // ਖ਼
	"khieukhacirclekorean":            '\u3278', // ㉸
	"khieukhaparenkorean":             '\u3218', // ㈘
	"khieukhcirclekorean":             '\u326a', // ㉪
	"khieukhkorean":                   '\u314b', // ㅋ
	"khieukhparenkorean":              '\u320a', // ㈊
	"khokhaithai":                     '\u0e02', // ข
	"khokhonthai":                     '\u0e05', // ฅ
	"khokhuatthai":                    '\u0e03', // ฃ
	"khokhwaithai":                    '\u0e04', // ค
	"khomutthai":                      '\u0e5b', // ๛
	"khook":                           '\u0199', // ƙ
	"khorakhangthai":                  '\u0e06', // ฆ
	"khzsquare":                       '\u3391', // ㎑
	"kihiragana":                      '\u304d', // き
	"kikatakana":                      '\u30ad', // キ
	"kikatakanahalfwidth":             '\uff77', // ｷ
	"kiroguramusquare":                '\u3315', // ㌕
	"kiromeetorusquare":               '\u3316', // ㌖
	"kirosquare":                      '\u3314', // ㌔
	"kiyeokacirclekorean":             '\u326e', // ㉮
	"kiyeokaparenkorean":              '\u320e', // ㈎
	"kiyeokcirclekorean":              '\u3260', // ㉠
	"kiyeokkorean":                    '\u3131', // ㄱ
	"kiyeokparenkorean":               '\u3200', // ㈀
	"kiyeoksioskorean":                '\u3133', // ㄳ
	"kjecyrillic":                     '\u045c', // ќ
	"klinebelow":                      '\u1e35', // ḵ
	"klsquare":                        '\u3398', // ㎘
	"kmcubedsquare":                   '\u33a6', // ㎦
	"kmonospace":                      '\uff4b', // ｋ
	"kmsquaredsquare":                 '\u33a2', // ㎢
	"kohiragana":                      '\u3053', // こ
	"kohmsquare":                      '\u33c0', // ㏀
	"kokaithai":                       '\u0e01', // ก
	"kokatakana":                      '\u30b3', // コ
	"kokatakanahalfwidth":             '\uff7a', // ｺ
	"kooposquare":                     '\u331e', // ㌞
	"koppacyrillic":                   '\u0481', // ҁ
	"koreanstandardsymbol":            '\u327f', // ㉿
	"koroniscmb":                      '\u0343', // ̓
	"kparen":                          '\u24a6', // ⒦
	"kpasquare":                       '\u33aa', // ㎪
	"ksicyrillic":                     '\u046f', // ѯ
	"ktsquare":                        '\u33cf', // ㏏
	"kturned":                         '\u029e', // ʞ
	"kuhiragana":                      '\u304f', // く
	"kukatakana":                      '\u30af', // ク
	"kukatakanahalfwidth":             '\uff78', // ｸ
	"kvsquare":                        '\u33b8', // ㎸
	"kwsquare":                        '\u33be', // ㎾
	"l":                               'l',      // l
	"labengali":                       '\u09b2', // ল
	"lacute":                          '\u013a', // ĺ
	"ladeva":                          '\u0932', // ल
	"lagujarati":                      '\u0ab2', // લ
	"lagurmukhi":                      '\u0a32', // ਲ
	"lakkhangyaothai":                 '\u0e45', // ๅ
	"lamaleffinalarabic":              '\ufefc', // ﻼ
	"lamalefhamzaabovefinalarabic":    '\ufef8', // ﻸ
	"lamalefhamzaaboveisolatedarabic": '\ufef7', // ﻷ
	"lamalefhamzabelowfinalarabic":    '\ufefa', // ﻺ
	"lamalefhamzabelowisolatedarabic": '\ufef9', // ﻹ
	"lamalefisolatedarabic":           '\ufefb', // ﻻ
	"lamalefmaddaabovefinalarabic":    '\ufef6', // ﻶ
	"lamalefmaddaaboveisolatedarabic": '\ufef5', // ﻵ
	"lamarabic":                       '\u0644', // ل
	"lambda":                          '\u03bb', // λ
	"lambdastroke":                    '\u019b', // ƛ
	"lamed":                           '\u05dc', // ל
	"lameddagesh":                     '\ufb3c', // לּ
	"lameddageshhebrew":               '\ufb3c', // לּ
	"lamedhebrew":                     '\u05dc', // ל
	"lamedholam":                      '\u05dc', // ל
	"lamedholamdagesh":                '\u05dc', // ל
	"lamedholamdageshhebrew":          '\u05dc', // ל
	"lamedholamhebrew":                '\u05dc', // ל
	"lamfinalarabic":                  '\ufede', // ﻞ
	"lamhahinitialarabic":             '\ufcca', // ﳊ
	"laminitialarabic":                '\ufedf', // ﻟ
	"lamjeeminitialarabic":            '\ufcc9', // ﳉ
	"lamkhahinitialarabic":            '\ufccb', // ﳋ
	"lamlamhehisolatedarabic":         '\ufdf2', // ﷲ
	"lammedialarabic":                 '\ufee0', // ﻠ
	"lammeemhahinitialarabic":         '\ufd88', // ﶈ
	"lammeeminitialarabic":            '\ufccc', // ﳌ
	"lammeemjeeminitialarabic":        '\ufedf', // ﻟ
	"lammeemkhahinitialarabic":        '\ufedf', // ﻟ
	"largecircle":                     '\u25ef', // ◯
	"lbar":                            '\u019a', // ƚ
	"lbelt":                           '\u026c', // ɬ
	"lbopomofo":                       '\u310c', // ㄌ
	"lcaron":                          '\u013e', // ľ
	"lcedilla":                        '\u013c', // ļ
	"lcircle":                         '\u24db', // ⓛ
	"lcircumflexbelow":                '\u1e3d', // ḽ
	"lcommaaccent":                    '\u013c', // ļ
	"ldot":                            '\u0140', // ŀ
	"ldotaccent":                      '\u0140', // ŀ
	"ldotbelow":                       '\u1e37', // ḷ
	"ldotbelowmacron":                 '\u1e39', // ḹ
	"leftangleabovecmb":               '\u031a', // ̚
	"lefttackbelowcmb":                '\u0318', // ̘
	"less":                            '<',      // <
	"lessequal":                       '\u2264', // ≤
	"lessequalorgreater":              '\u22da', // ⋚
	"lessmonospace":                   '\uff1c', // ＜
	"lessorequivalent":                '\u2272', // ≲
	"lessorgreater":                   '\u2276', // ≶
	"lessoverequal":                   '\u2266', // ≦
	"lesssmall":                       '\ufe64', // ﹤
	"lezh":                            '\u026e', // ɮ
	"lfblock":                         '\u258c', // ▌
	"lhookretroflex":                  '\u026d', // ɭ
	"lira":                            '\u20a4', // ₤
	"liwnarmenian":                    '\u056c', // լ
	"lj":                              '\u01c9', // ǉ
	"ljecyrillic":                     '\u0459', // љ
	"ll":                              '\uf6c0',
	"lladeva":                         '\u0933', // ळ
	"llagujarati":                     '\u0ab3', // ળ
	"llinebelow":                      '\u1e3b', // ḻ
	"llladeva":                        '\u0934', // ऴ
	"llvocalicbengali":                '\u09e1', // ৡ
	"llvocalicdeva":                   '\u0961', // ॡ
	"llvocalicvowelsignbengali":       '\u09e3', // ৣ
	"llvocalicvowelsigndeva":          '\u0963', // ॣ
	"lmiddletilde":                    '\u026b', // ɫ
	"lmonospace":                      '\uff4c', // ｌ
	"lmsquare":                        '\u33d0', // ㏐
	"lochulathai":                     '\u0e2c', // ฬ
	"logicaland":                      '\u2227', // ∧
	"logicalnot":                      '\u00ac', // ¬
	"logicalnotreversed":              '\u2310', // ⌐
	"logicalor":                       '\u2228', // ∨
	"lolingthai":                      '\u0e25', // ล
	"longs":                           '\u017f', // ſ
	"lowlinecenterline":               '\ufe4e', // ﹎
	"lowlinecmb":                      '\u0332', // ̲
	"lowlinedashed":                   '\ufe4d', // ﹍
	"lozenge":                         '\u25ca', // ◊
	"lparen":                          '\u24a7', // ⒧
	"lslash":                          '\u0142', // ł
	"lsquare":                         '\u2113', // ℓ
	"lsuperior":                       '\uf6ee',
	"ltshade":                         '\u2591', // ░
	"luthai":                          '\u0e26', // ฦ
	"lvocalicbengali":                 '\u098c', // ঌ
	"lvocalicdeva":                    '\u090c', // ऌ
	"lvocalicvowelsignbengali":        '\u09e2', // ৢ
	"lvocalicvowelsigndeva":           '\u0962', // ॢ
	"lxsquare":                        '\u33d3', // ㏓
	"m":                               'm',      // m
	"mabengali":                       '\u09ae', // ম
	"macron":                          '\u00af', // ¯
	"macronbelowcmb":                  '\u0331', // ̱
	"macroncmb":                       '\u0304', // ̄
	"macronlowmod":                    '\u02cd', // ˍ
	"macronmonospace":                 '\uffe3', // ￣
	"macute":                          '\u1e3f', // ḿ
	"madeva":                          '\u092e', // म
	"magujarati":                      '\u0aae', // મ
	"magurmukhi":                      '\u0a2e', // ਮ
	"mahapakhhebrew":                  '\u05a4', // ֤
	"mahapakhlefthebrew":              '\u05a4', // ֤
	"mahiragana":                      '\u307e', // ま
	"maichattawalowleftthai":          '\uf895',
	"maichattawalowrightthai":         '\uf894',
	"maichattawathai":                 '\u0e4b', // ๋
	"maichattawaupperleftthai":        '\uf893',
	"maieklowleftthai":                '\uf88c',
	"maieklowrightthai":               '\uf88b',
	"maiekthai":                       '\u0e48', // ่
	"maiekupperleftthai":              '\uf88a',
	"maihanakatleftthai":              '\uf884',
	"maihanakatthai":                  '\u0e31', // ั
	"maitaikhuleftthai":               '\uf889',
	"maitaikhuthai":                   '\u0e47', // ็
	"maitholowleftthai":               '\uf88f',
	"maitholowrightthai":              '\uf88e',
	"maithothai":                      '\u0e49', // ้
	"maithoupperleftthai":             '\uf88d',
	"maitrilowleftthai":               '\uf892',
	"maitrilowrightthai":              '\uf891',
	"maitrithai":                      '\u0e4a', // ๊
	"maitriupperleftthai":             '\uf890',
	"maiyamokthai":                    '\u0e46', // ๆ
	"makatakana":                      '\u30de', // マ
	"makatakanahalfwidth":             '\uff8f', // ﾏ
	"male":                            '\u2642', // ♂
	"mansyonsquare":                   '\u3347', // ㍇
	"maqafhebrew":                     '\u05be', // ־
	"mars":                            '\u2642', // ♂
	"masoracirclehebrew":              '\u05af', // ֯
	"masquare":                        '\u3383', // ㎃
	"mbopomofo":                       '\u3107', // ㄇ
	"mbsquare":                        '\u33d4', // ㏔
	"mcircle":                         '\u24dc', // ⓜ
	"mcubedsquare":                    '\u33a5', // ㎥
	"mdotaccent":                      '\u1e41', // ṁ
	"mdotbelow":                       '\u1e43', // ṃ
	"meemarabic":                      '\u0645', // م
	"meemfinalarabic":                 '\ufee2', // ﻢ
	"meeminitialarabic":               '\ufee3', // ﻣ
	"meemmedialarabic":                '\ufee4', // ﻤ
	"meemmeeminitialarabic":           '\ufcd1', // ﳑ
	"meemmeemisolatedarabic":          '\ufc48', // ﱈ
	"meetorusquare":                   '\u334d', // ㍍
	"mehiragana":                      '\u3081', // め
	"meizierasquare":                  '\u337e', // ㍾
	"mekatakana":                      '\u30e1', // メ
	"mekatakanahalfwidth":             '\uff92', // ﾒ
	"mem":                        '\u05de', // מ
	"memdagesh":                  '\ufb3e', // מּ
	"memdageshhebrew":            '\ufb3e', // מּ
	"memhebrew":                  '\u05de', // מ
	"menarmenian":                '\u0574', // մ
	"merkhahebrew":               '\u05a5', // ֥
	"merkhakefulahebrew":         '\u05a6', // ֦
	"merkhakefulalefthebrew":     '\u05a6', // ֦
	"merkhalefthebrew":           '\u05a5', // ֥
	"mhook":                      '\u0271', // ɱ
	"mhzsquare":                  '\u3392', // ㎒
	"middledotkatakanahalfwidth": '\uff65', // ･
	"middot":                     '\u00b7', // ·
	"mieumacirclekorean":         '\u3272', // ㉲
	"mieumaparenkorean":          '\u3212', // ㈒
	"mieumcirclekorean":          '\u3264', // ㉤
	"mieumkorean":                '\u3141', // ㅁ
	"mieumpansioskorean":         '\u3170', // ㅰ
	"mieumparenkorean":           '\u3204', // ㈄
	"mieumpieupkorean":           '\u316e', // ㅮ
	"mieumsioskorean":            '\u316f', // ㅯ
	"mihiragana":                 '\u307f', // み
	"mikatakana":                 '\u30df', // ミ
	"mikatakanahalfwidth":        '\uff90', // ﾐ
	"minus":                      '\u2212', // −
	"minusbelowcmb":              '\u0320', // ̠
	"minuscircle":                '\u2296', // ⊖
	"minusmod":                   '\u02d7', // ˗
	"minusplus":                  '\u2213', // ∓
	"minute":                     '\u2032', // ′
	"miribaarusquare":            '\u334a', // ㍊
	"mirisquare":                 '\u3349', // ㍉
	"mlonglegturned":             '\u0270', // ɰ
	"mlsquare":                   '\u3396', // ㎖
	"mmcubedsquare":              '\u33a3', // ㎣
	"mmonospace":                 '\uff4d', // ｍ
	"mmsquaredsquare":            '\u339f', // ㎟
	"mohiragana":                 '\u3082', // も
	"mohmsquare":                 '\u33c1', // ㏁
	"mokatakana":                 '\u30e2', // モ
	"mokatakanahalfwidth":        '\uff93', // ﾓ
	"molsquare":                  '\u33d6', // ㏖
	"momathai":                   '\u0e21', // ม
	"moverssquare":               '\u33a7', // ㎧
	"moverssquaredsquare":        '\u33a8', // ㎨
	"mparen":                     '\u24a8', // ⒨
	"mpasquare":                  '\u33ab', // ㎫
	"mssquare":                   '\u33b3', // ㎳
	"msuperior":                  '\uf6ef',
	"mturned":                    '\u026f', // ɯ
	"mu":                         '\u00b5', // µ
	"mu1":                        '\u00b5', // µ
	"muasquare":                  '\u3382', // ㎂
	"muchgreater":                '\u226b', // ≫
	"muchless":                   '\u226a', // ≪
	"mufsquare":                  '\u338c', // ㎌
	"mugreek":                    '\u03bc', // μ
	"mugsquare":                  '\u338d', // ㎍
	"muhiragana":                 '\u3080', // む
	"mukatakana":                 '\u30e0', // ム
	"mukatakanahalfwidth":        '\uff91', // ﾑ
	"mulsquare":                  '\u3395', // ㎕
	"multiply":                   '\u00d7', // ×
	"mumsquare":                  '\u339b', // ㎛
	"munahhebrew":                '\u05a3', // ֣
	"munahlefthebrew":            '\u05a3', // ֣
	"musicalnote":                '\u266a', // ♪
	"musicalnotedbl":             '\u266b', // ♫
	"musicflatsign":              '\u266d', // ♭
	"musicsharpsign":             '\u266f', // ♯
	"mussquare":                  '\u33b2', // ㎲
	"muvsquare":                  '\u33b6', // ㎶
	"muwsquare":                  '\u33bc', // ㎼
	"mvmegasquare":               '\u33b9', // ㎹
	"mvsquare":                   '\u33b7', // ㎷
	"mwmegasquare":               '\u33bf', // ㎿
	"mwsquare":                   '\u33bd', // ㎽
	"n":                          'n',      // n
	"nabengali":                  '\u09a8', // ন
	"nabla":                      '\u2207', // ∇
	"nacute":                     '\u0144', // ń
	"nadeva":                     '\u0928', // न
	"nagujarati":                 '\u0aa8', // ન
	"nagurmukhi":                 '\u0a28', // ਨ
	"nahiragana":                 '\u306a', // な
	"nakatakana":                 '\u30ca', // ナ
	"nakatakanahalfwidth":        '\uff85', // ﾅ
	"napostrophe":                '\u0149', // ŉ
	"nasquare":                   '\u3381', // ㎁
	"nbopomofo":                  '\u310b', // ㄋ
	"nbspace":                    '\u00a0',
	"ncaron":                     '\u0148', // ň
	"ncedilla":                   '\u0146', // ņ
	"ncircle":                    '\u24dd', // ⓝ
	"ncircumflexbelow":           '\u1e4b', // ṋ
	"ncommaaccent":               '\u0146', // ņ
	"ndotaccent":                 '\u1e45', // ṅ
	"ndotbelow":                  '\u1e47', // ṇ
	"nehiragana":                 '\u306d', // ね
	"nekatakana":                 '\u30cd', // ネ
	"nekatakanahalfwidth":        '\uff88', // ﾈ
	"newsheqelsign":              '\u20aa', // ₪
	"nfsquare":                   '\u338b', // ㎋
	"ngabengali":                 '\u0999', // ঙ
	"ngadeva":                    '\u0919', // ङ
	"ngagujarati":                '\u0a99', // ઙ
	"ngagurmukhi":                '\u0a19', // ਙ
	"ngonguthai":                 '\u0e07', // ง
	"nhiragana":                  '\u3093', // ん
	"nhookleft":                  '\u0272', // ɲ
	"nhookretroflex":             '\u0273', // ɳ
	"nieunacirclekorean":         '\u326f', // ㉯
	"nieunaparenkorean":          '\u320f', // ㈏
	"nieuncieuckorean":           '\u3135', // ㄵ
	"nieuncirclekorean":          '\u3261', // ㉡
	"nieunhieuhkorean":           '\u3136', // ㄶ
	"nieunkorean":                '\u3134', // ㄴ
	"nieunpansioskorean":         '\u3168', // ㅨ
	"nieunparenkorean":           '\u3201', // ㈁
	"nieunsioskorean":            '\u3167', // ㅧ
	"nieuntikeutkorean":          '\u3166', // ㅦ
	"nihiragana":                 '\u306b', // に
	"nikatakana":                 '\u30cb', // ニ
	"nikatakanahalfwidth":        '\uff86', // ﾆ
	"nikhahitleftthai":           '\uf899',
	"nikhahitthai":               '\u0e4d', // ํ
	"nine":                       '9',      // 9
	"ninearabic":                 '\u0669', // ٩
	"ninebengali":                '\u09ef', // ৯
	"ninecircle":                 '\u2468', // ⑨
	"ninecircleinversesansserif": '\u2792', // ➒
	"ninedeva":                   '\u096f', // ९
	"ninegujarati":               '\u0aef', // ૯
	"ninegurmukhi":               '\u0a6f', // ੯
	"ninehackarabic":             '\u0669', // ٩
	"ninehangzhou":               '\u3029', // 〩
	"nineideographicparen":       '\u3228', // ㈨
	"nineinferior":               '\u2089', // ₉
	"ninemonospace":              '\uff19', // ９
	"nineoldstyle":               '\uf739',
	"nineparen":                  '\u247c', // ⑼
	"nineperiod":                 '\u2490', // ⒐
	"ninepersian":                '\u06f9', // ۹
	"nineroman":                  '\u2178', // ⅸ
	"ninesuperior":               '\u2079', // ⁹
	"nineteencircle":             '\u2472', // ⑲
	"nineteenparen":              '\u2486', // ⒆
	"nineteenperiod":             '\u249a', // ⒚
	"ninethai":                   '\u0e59', // ๙
	"nj":                         '\u01cc', // ǌ
	"njecyrillic":                '\u045a', // њ
	"nkatakana":                  '\u30f3', // ン
	"nkatakanahalfwidth":         '\uff9d', // ﾝ
	"nlegrightlong":              '\u019e', // ƞ
	"nlinebelow":                 '\u1e49', // ṉ
	"nmonospace":                 '\uff4e', // ｎ
	"nmsquare":                   '\u339a', // ㎚
	"nnabengali":                 '\u09a3', // ণ
	"nnadeva":                    '\u0923', // ण
	"nnagujarati":                '\u0aa3', // ણ
	"nnagurmukhi":                '\u0a23', // ਣ
	"nnnadeva":                   '\u0929', // ऩ
	"nohiragana":                 '\u306e', // の
	"nokatakana":                 '\u30ce', // ノ
	"nokatakanahalfwidth":        '\uff89', // ﾉ
	"nonbreakingspace":           '\u00a0',
	"nonenthai":                  '\u0e13', // ณ
	"nonuthai":                   '\u0e19', // น
	"noonarabic":                 '\u0646', // ن
	"noonfinalarabic":            '\ufee6', // ﻦ
	"noonghunnaarabic":           '\u06ba', // ں
	"noonghunnafinalarabic":      '\ufb9f', // ﮟ
	"noonhehinitialarabic":       '\ufee7', // ﻧ
	"nooninitialarabic":          '\ufee7', // ﻧ
	"noonjeeminitialarabic":      '\ufcd2', // ﳒ
	"noonjeemisolatedarabic":     '\ufc4b', // ﱋ
	"noonmedialarabic":           '\ufee8', // ﻨ
	"noonmeeminitialarabic":      '\ufcd5', // ﳕ
	"noonmeemisolatedarabic":     '\ufc4e', // ﱎ
	"noonnoonfinalarabic":        '\ufc8d', // ﲍ
	"notcontains":                '\u220c', // ∌
	"notelement":                 '\u2209', // ∉
	"notelementof":               '\u2209', // ∉
	"notequal":                   '\u2260', // ≠
	"notgreater":                 '\u226f', // ≯
	"notgreaternorequal":         '\u2271', // ≱
	"notgreaternorless":          '\u2279', // ≹
	"notidentical":               '\u2262', // ≢
	"notless":                    '\u226e', // ≮
	"notlessnorequal":            '\u2270', // ≰
	"notparallel":                '\u2226', // ∦
	"notprecedes":                '\u2280', // ⊀
	"notsubset":                  '\u2284', // ⊄
	"notsucceeds":                '\u2281', // ⊁
	"notsuperset":                '\u2285', // ⊅
	"nowarmenian":                '\u0576', // ն
	"nparen":                     '\u24a9', // ⒩
	"nssquare":                   '\u33b1', // ㎱
	"nsuperior":                  '\u207f', // ⁿ
	"ntilde":                     '\u00f1', // ñ
	"nu":                         '\u03bd', // ν
	"nuhiragana":                 '\u306c', // ぬ
	"nukatakana":                 '\u30cc', // ヌ
	"nukatakanahalfwidth":        '\uff87', // ﾇ
	"nuktabengali":               '\u09bc', // ়
	"nuktadeva":                  '\u093c', // ़
	"nuktagujarati":              '\u0abc', // ઼
	"nuktagurmukhi":              '\u0a3c', // ਼
	"numbersign":                 '#',      // #
	"numbersignmonospace":        '\uff03', // ＃
	"numbersignsmall":            '\ufe5f', // ﹟
	"numeralsigngreek":           '\u0374', // ʹ
	"numeralsignlowergreek":      '\u0375', // ͵
	"numero":                     '\u2116', // №
	"nun":                        '\u05e0', // נ
	"nundagesh":                  '\ufb40', // נּ
	"nundageshhebrew":            '\ufb40', // נּ
	"nunhebrew":                  '\u05e0', // נ
	"nvsquare":                   '\u33b5', // ㎵
	"nwsquare":                   '\u33bb', // ㎻
	"nyabengali":                 '\u099e', // ঞ
	"nyadeva":                    '\u091e', // ञ
	"nyagujarati":                '\u0a9e', // ઞ
	"nyagurmukhi":                '\u0a1e', // ਞ
	"o":                          'o',      // o
	"oacute":                     '\u00f3', // ó
	"oangthai":                   '\u0e2d', // อ
	"obarred":                    '\u0275', // ɵ
	"obarredcyrillic":            '\u04e9', // ө
	"obarreddieresiscyrillic":    '\u04eb', // ӫ
	"obengali":                   '\u0993', // ও
	"obopomofo":                  '\u311b', // ㄛ
	"obreve":                     '\u014f', // ŏ
	"ocandradeva":                '\u0911', // ऑ
	"ocandragujarati":            '\u0a91', // ઑ
	"ocandravowelsigndeva":       '\u0949', // ॉ
	"ocandravowelsigngujarati":   '\u0ac9', // ૉ
	"ocaron":                     '\u01d2', // ǒ
	"ocircle":                    '\u24de', // ⓞ
	"ocircumflex":                '\u00f4', // ô
	"ocircumflexacute":           '\u1ed1', // ố
	"ocircumflexdotbelow":        '\u1ed9', // ộ
	"ocircumflexgrave":           '\u1ed3', // ồ
	"ocircumflexhookabove":       '\u1ed5', // ổ
	"ocircumflextilde":           '\u1ed7', // ỗ
	"ocyrillic":                  '\u043e', // о
	"odblacute":                  '\u0151', // ő
	"odblgrave":                  '\u020d', // ȍ
	"odeva":                      '\u0913', // ओ
	"odieresis":                  '\u00f6', // ö
	"odieresiscyrillic":          '\u04e7', // ӧ
	"odotbelow":                  '\u1ecd', // ọ
	"oe":                         '\u0153', // œ
	"oekorean":                   '\u315a', // ㅚ
	"ogonek":                     '\u02db', // ˛
	"ogonekcmb":                  '\u0328', // ̨
	"ograve":                     '\u00f2', // ò
	"ogujarati":                  '\u0a93', // ઓ
	"oharmenian":                 '\u0585', // օ
	"ohiragana":                  '\u304a', // お
	"ohookabove":                 '\u1ecf', // ỏ
	"ohorn":                      '\u01a1', // ơ
	"ohornacute":                 '\u1edb', // ớ
	"ohorndotbelow":              '\u1ee3', // ợ
	"ohorngrave":                 '\u1edd', // ờ
	"ohornhookabove":             '\u1edf', // ở
	"ohorntilde":                 '\u1ee1', // ỡ
	"ohungarumlaut":              '\u0151', // ő
	"oi":                         '\u01a3', // ƣ
	"oinvertedbreve":             '\u020f', // ȏ
	"okatakana":                  '\u30aa', // オ
	"okatakanahalfwidth":         '\uff75', // ｵ
	"okorean":                    '\u3157', // ㅗ
	"olehebrew":                  '\u05ab', // ֫
	"omacron":                    '\u014d', // ō
	"omacronacute":               '\u1e53', // ṓ
	"omacrongrave":               '\u1e51', // ṑ
	"omdeva":                     '\u0950', // ॐ
	"omega":                      '\u03c9', // ω
	"omega1":                     '\u03d6', // ϖ
	"omegacyrillic":              '\u0461', // ѡ
	"omegalatinclosed":           '\u0277', // ɷ
	"omegaroundcyrillic":         '\u047b', // ѻ
	"omegatitlocyrillic":         '\u047d', // ѽ
	"omegatonos":                 '\u03ce', // ώ
	"omgujarati":                 '\u0ad0', // ૐ
	"omicron":                    '\u03bf', // ο
	"omicrontonos":               '\u03cc', // ό
	"omonospace":                 '\uff4f', // ｏ
	"one":                        '1',      // 1
	"onearabic":                  '\u0661', // ١
	"onebengali":                 '\u09e7', // ১
	"onecircle":                  '\u2460', // ①
	"onecircleinversesansserif":  '\u278a', // ➊
	"onedeva":                    '\u0967', // १
	"onedotenleader":             '\u2024', // ․
	"oneeighth":                  '\u215b', // ⅛
	"onefitted":                  '\uf6dc',
	"onegujarati":                '\u0ae7', // ૧
	"onegurmukhi":                '\u0a67', // ੧
	"onehackarabic":              '\u0661', // ١
	"onehalf":                    '\u00bd', // ½
	"onehangzhou":                '\u3021', // 〡
	"oneideographicparen":        '\u3220', // ㈠
	"oneinferior":                '\u2081', // ₁
	"onemonospace":               '\uff11', // １
	"onenumeratorbengali":        '\u09f4', // ৴
	"oneoldstyle":                '\uf731',
	"oneparen":                   '\u2474', // ⑴
	"oneperiod":                  '\u2488', // ⒈
	"onepersian":                 '\u06f1', // ۱
	"onequarter":                 '\u00bc', // ¼
	"oneroman":                   '\u2170', // ⅰ
	"onesuperior":                '\u00b9', // ¹
	"onethai":                    '\u0e51', // ๑
	"onethird":                   '\u2153', // ⅓
	"oogonek":                    '\u01eb', // ǫ
	"oogonekmacron":              '\u01ed', // ǭ
	"oogurmukhi":                 '\u0a13', // ਓ
	"oomatragurmukhi":            '\u0a4b', // ੋ
	"oopen":                      '\u0254', // ɔ
	"oparen":                     '\u24aa', // ⒪
	"openbullet":                 '\u25e6', // ◦
	"option":                     '\u2325', // ⌥
	"ordfeminine":                '\u00aa', // ª
	"ordmasculine":               '\u00ba', // º
	"orthogonal":                 '\u221f', // ∟
	"oshortdeva":                 '\u0912', // ऒ
	"oshortvowelsigndeva":        '\u094a', // ॊ
	"oslash":                     '\u00f8', // ø
	"oslashacute":                '\u01ff', // ǿ
	"osmallhiragana":             '\u3049', // ぉ
	"osmallkatakana":             '\u30a9', // ォ
	"osmallkatakanahalfwidth":    '\uff6b', // ｫ
	"ostrokeacute":               '\u01ff', // ǿ
	"osuperior":                  '\uf6f0',
	"otcyrillic":                 '\u047f', // ѿ
	"otilde":                     '\u00f5', // õ
	"otildeacute":                '\u1e4d', // ṍ
	"otildedieresis":             '\u1e4f', // ṏ
	"oubopomofo":                 '\u3121', // ㄡ
	"overline":                   '\u203e', // ‾
	"overlinecenterline":         '\ufe4a', // ﹊
	"overlinecmb":                '\u0305', // ̅
	"overlinedashed":             '\ufe49', // ﹉
	"overlinedblwavy":            '\ufe4c', // ﹌
	"overlinewavy":               '\ufe4b', // ﹋
	"overscore":                  '\u00af', // ¯
	"ovowelsignbengali":          '\u09cb', // ো
	"ovowelsigndeva":             '\u094b', // ो
	"ovowelsigngujarati":         '\u0acb', // ો
	"p":                          'p',      // p
	"paampssquare":               '\u3380', // ㎀
	"paasentosquare":             '\u332b', // ㌫
	"pabengali":                  '\u09aa', // প
	"pacute":                     '\u1e55', // ṕ
	"padeva":                     '\u092a', // प
	"pagedown":                   '\u21df', // ⇟
	"pageup":                     '\u21de', // ⇞
	"pagujarati":                 '\u0aaa', // પ
	"pagurmukhi":                 '\u0a2a', // ਪ
	"pahiragana":                 '\u3071', // ぱ
	"paiyannoithai":              '\u0e2f', // ฯ
	"pakatakana":                 '\u30d1', // パ
	"palatalizationcyrilliccmb":  '\u0484', // ҄
	"palochkacyrillic":           '\u04c0', // Ӏ
	"pansioskorean":              '\u317f', // ㅿ
	"paragraph":                  '\u00b6', // ¶
	"parallel":                   '\u2225', // ∥
	"parenleft":                  '(',      // (
	"parenleftaltonearabic":      '\ufd3e', // ﴾
	"parenleftbt":                '\uf8ed',
	"parenleftex":                '\uf8ec',
	"parenleftinferior":          '\u208d', // ₍
	"parenleftmonospace":         '\uff08', // （
	"parenleftsmall":             '\ufe59', // ﹙
	"parenleftsuperior":          '\u207d', // ⁽
	"parenlefttp":                '\uf8eb',
	"parenleftvertical":          '\ufe35', // ︵
	"parenright":                 ')',      // )
	"parenrightaltonearabic":     '\ufd3f', // ﴿
	"parenrightbt":               '\uf8f8',
	"parenrightex":               '\uf8f7',
	"parenrightinferior":         '\u208e', // ₎
	"parenrightmonospace":        '\uff09', // ）
	"parenrightsmall":            '\ufe5a', // ﹚
	"parenrightsuperior":         '\u207e', // ⁾
	"parenrighttp":               '\uf8f6',
	"parenrightvertical":         '\ufe36', // ︶
	"partialdiff":                '\u2202', // ∂
	"paseqhebrew":                '\u05c0', // ׀
	"pashtahebrew":               '\u0599', // ֙
	"pasquare":                   '\u33a9', // ㎩
	"patah":                      '\u05b7', // ַ
	"patah11":                    '\u05b7', // ַ
	"patah1d":                    '\u05b7', // ַ
	"patah2a":                    '\u05b7', // ַ
	"patahhebrew":                '\u05b7', // ַ
	"patahnarrowhebrew":          '\u05b7', // ַ
	"patahquarterhebrew":         '\u05b7', // ַ
	"patahwidehebrew":            '\u05b7', // ַ
	"pazerhebrew":                '\u05a1', // ֡
	"pbopomofo":                  '\u3106', // ㄆ
	"pcircle":                    '\u24df', // ⓟ
	"pdotaccent":                 '\u1e57', // ṗ
	"pe":                         '\u05e4', // פ
	"pecyrillic":                 '\u043f', // п
	"pedagesh":                   '\ufb44', // פּ
	"pedageshhebrew":             '\ufb44', // פּ
	"peezisquare":                '\u333b', // ㌻
	"pefinaldageshhebrew":        '\ufb43', // ףּ
	"peharabic":                  '\u067e', // پ
	"peharmenian":                '\u057a', // պ
	"pehebrew":                   '\u05e4', // פ
	"pehfinalarabic":             '\ufb57', // ﭗ
	"pehinitialarabic":           '\ufb58', // ﭘ
	"pehiragana":                 '\u307a', // ぺ
	"pehmedialarabic":            '\ufb59', // ﭙ
	"pekatakana":                 '\u30da', // ペ
	"pemiddlehookcyrillic":       '\u04a7', // ҧ
	"perafehebrew":               '\ufb4e', // פֿ
	"percent":                    '%',      // %
	"percentarabic":              '\u066a', // ٪
	"percentmonospace":           '\uff05', // ％
	"percentsmall":               '\ufe6a', // ﹪
	"period":                     '.',      // .
	"periodarmenian":             '\u0589', // ։
	"periodcentered":             '\u00b7', // ·
	"periodhalfwidth":            '\uff61', // ｡
	"periodinferior":             '\uf6e7',
	"periodmonospace":            '\uff0e', // ．
	"periodsmall":                '\ufe52', // ﹒
	"periodsuperior":             '\uf6e8',
	"perispomenigreekcmb":        '\u0342', // ͂
	"perpendicular":              '\u22a5', // ⊥
	"perthousand":                '\u2030', // ‰
	"peseta":                     '\u20a7', // ₧
	"pfsquare":                   '\u338a', // ㎊
	"phabengali":                 '\u09ab', // ফ
	"phadeva":                    '\u092b', // फ
	"phagujarati":                '\u0aab', // ફ
	"phagurmukhi":                '\u0a2b', // ਫ
	"phi":                        '\u03c6', // φ
	"phi1":                       '\u03d5', // ϕ
	"phieuphacirclekorean":       '\u327a', // ㉺
	"phieuphaparenkorean":        '\u321a', // ㈚
	"phieuphcirclekorean":        '\u326c', // ㉬
	"phieuphkorean":              '\u314d', // ㅍ
	"phieuphparenkorean":         '\u320c', // ㈌
	"philatin":                   '\u0278', // ɸ
	"phinthuthai":                '\u0e3a', // ฺ
	"phisymbolgreek":             '\u03d5', // ϕ
	"phook":                      '\u01a5', // ƥ
	"phophanthai":                '\u0e1e', // พ
	"phophungthai":               '\u0e1c', // ผ
	"phosamphaothai":             '\u0e20', // ภ
	"pi":                         '\u03c0', // π
	"pieupacirclekorean":         '\u3273', // ㉳
	"pieupaparenkorean":          '\u3213', // ㈓
	"pieupcieuckorean":           '\u3176', // ㅶ
	"pieupcirclekorean":          '\u3265', // ㉥
	"pieupkiyeokkorean":          '\u3172', // ㅲ
	"pieupkorean":                '\u3142', // ㅂ
	"pieupparenkorean":           '\u3205', // ㈅
	"pieupsioskiyeokkorean":      '\u3174', // ㅴ
	"pieupsioskorean":            '\u3144', // ㅄ
	"pieupsiostikeutkorean":      '\u3175', // ㅵ
	"pieupthieuthkorean":         '\u3177', // ㅷ
	"pieuptikeutkorean":          '\u3173', // ㅳ
	"pihiragana":                 '\u3074', // ぴ
	"pikatakana":                 '\u30d4', // ピ
	"pisymbolgreek":              '\u03d6', // ϖ
	"piwrarmenian":               '\u0583', // փ
	"plus":                       '+',      // +
	"plusbelowcmb":               '\u031f', // ̟
	"pluscircle":                 '\u2295', // ⊕
	"plusminus":                  '\u00b1', // ±
	"plusmod":                    '\u02d6', // ˖
	"plusmonospace":              '\uff0b', // ＋
	"plussmall":                  '\ufe62', // ﹢
	"plussuperior":               '\u207a', // ⁺
	"pmonospace":                 '\uff50', // ｐ
	"pmsquare":                   '\u33d8', // ㏘
	"pohiragana":                 '\u307d', // ぽ
	"pointingindexdownwhite":     '\u261f', // ☟
	"pointingindexleftwhite":     '\u261c', // ☜
	"pointingindexrightwhite":    '\u261e', // ☞
	"pointingindexupwhite":       '\u261d', // ☝
	"pokatakana":                 '\u30dd', // ポ
	"poplathai":                  '\u0e1b', // ป
	"postalmark":                 '\u3012', // 〒
	"postalmarkface":             '\u3020', // 〠
	"pparen":                     '\u24ab', // ⒫
	"precedes":                   '\u227a', // ≺
	"prescription":               '\u211e', // ℞
	"primemod":                   '\u02b9', // ʹ
	"primereversed":              '\u2035', // ‵
	"product":                    '\u220f', // ∏
	"projective":                 '\u2305', // ⌅
	"prolongedkana":              '\u30fc', // ー
	"propellor":                  '\u2318', // ⌘
	"propersubset":               '\u2282', // ⊂
	"propersuperset":             '\u2283', // ⊃
	"proportion":                 '\u2237', // ∷
	"proportional":               '\u221d', // ∝
	"psi":                        '\u03c8', // ψ
	"psicyrillic":                '\u0471', // ѱ
	"psilipneumatacyrilliccmb":   '\u0486', // ҆
	"pssquare":                   '\u33b0', // ㎰
	"puhiragana":                 '\u3077', // ぷ
	"pukatakana":                 '\u30d7', // プ
	"pvsquare":                   '\u33b4', // ㎴
	"pwsquare":                   '\u33ba', // ㎺
	"q":                          'q',      // q
	"qadeva":                     '\u0958', // क़
	"qadmahebrew":                '\u05a8', // ֨
	"qafarabic":                  '\u0642', // ق
	"qaffinalarabic":             '\ufed6', // ﻖ
	"qafinitialarabic":           '\ufed7', // ﻗ
	"qafmedialarabic":            '\ufed8', // ﻘ
	"qamats":                     '\u05b8', // ָ
	"qamats10":                   '\u05b8', // ָ
	"qamats1a":                   '\u05b8', // ָ
	"qamats1c":                   '\u05b8', // ָ
	"qamats27":                   '\u05b8', // ָ
	"qamats29":                   '\u05b8', // ָ
	"qamats33":                   '\u05b8', // ָ
	"qamatsde":                   '\u05b8', // ָ
	"qamatshebrew":               '\u05b8', // ָ
	"qamatsnarrowhebrew":         '\u05b8', // ָ
	"qamatsqatanhebrew":          '\u05b8', // ָ
	"qamatsqatannarrowhebrew":    '\u05b8', // ָ
	"qamatsqatanquarterhebrew":   '\u05b8', // ָ
	"qamatsqatanwidehebrew":      '\u05b8', // ָ
	"qamatsquarterhebrew":        '\u05b8', // ָ
	"qamatswidehebrew":           '\u05b8', // ָ
	"qarneyparahebrew":           '\u059f', // ֟
	"qbopomofo":                  '\u3111', // ㄑ
	"qcircle":                    '\u24e0', // ⓠ
	"qhook":                      '\u02a0', // ʠ
	"qmonospace":                 '\uff51', // ｑ
	"qof":                        '\u05e7', // ק
	"qofdagesh":                  '\ufb47', // קּ
	"qofdageshhebrew":            '\ufb47', // קּ
	"qofhatafpatah":              '\u05e7', // ק
	"qofhatafpatahhebrew":        '\u05e7', // ק
	"qofhatafsegol":              '\u05e7', // ק
	"qofhatafsegolhebrew":        '\u05e7', // ק
	"qofhebrew":                  '\u05e7', // ק
	"qofhiriq":                   '\u05e7', // ק
	"qofhiriqhebrew":             '\u05e7', // ק
	"qofholam":                   '\u05e7', // ק
	"qofholamhebrew":             '\u05e7', // ק
	"qofpatah":                   '\u05e7', // ק
	"qofpatahhebrew":             '\u05e7', // ק
	"qofqamats":                  '\u05e7', // ק
	"qofqamatshebrew":            '\u05e7', // ק
	"qofqubuts":                  '\u05e7', // ק
	"qofqubutshebrew":            '\u05e7', // ק
	"qofsegol":                   '\u05e7', // ק
	"qofsegolhebrew":             '\u05e7', // ק
	"qofsheva":                   '\u05e7', // ק
	"qofshevahebrew":             '\u05e7', // ק
	"qoftsere":                   '\u05e7', // ק
	"qoftserehebrew":             '\u05e7', // ק
	"qparen":                     '\u24ac', // ⒬
	"quarternote":                '\u2669', // ♩
	"qubuts":                     '\u05bb', // ֻ
	"qubuts18":                   '\u05bb', // ֻ
	"qubuts25":                   '\u05bb', // ֻ
	"qubuts31":                   '\u05bb', // ֻ
	"qubutshebrew":               '\u05bb', // ֻ
	"qubutsnarrowhebrew":         '\u05bb', // ֻ
	"qubutsquarterhebrew":        '\u05bb', // ֻ
	"qubutswidehebrew":           '\u05bb', // ֻ
	"question":                   '?',      // ?
	"questionarabic":             '\u061f', // ؟
	"questionarmenian":           '\u055e', // ՞
	"questiondown":               '\u00bf', // ¿
	"questiondownsmall":          '\uf7bf',
	"questiongreek":              '\u037e', // ;
	"questionmonospace":          '\uff1f', // ？
	"questionsmall":              '\uf73f',
	"quotedbl":                   '"',      // "
	"quotedblbase":               '\u201e', // „
	"quotedblleft":               '\u201c', // “
	"quotedblmonospace":          '\uff02', // ＂
	"quotedblprime":              '\u301e', // 〞
	"quotedblprimereversed":      '\u301d', // 〝
	"quotedblright":              '\u201d', // ”
	"quoteleft":                  '\u2018', // ‘
	"quoteleftreversed":          '\u201b', // ‛
	"quotereversed":              '\u201b', // ‛
	"quoteright":                 '\u2019', // ’
	"quoterightn":                '\u0149', // ŉ
	"quotesinglbase":             '\u201a', // ‚
	"quotesingle":                '\'',     // \'
	"quotesinglemonospace":       '\uff07', // ＇
	"r":                          'r',      // r
	"raarmenian":                 '\u057c', // ռ
	"rabengali":                  '\u09b0', // র
	"racute":                     '\u0155', // ŕ
	"radeva":                     '\u0930', // र
	"radical":                    '\u221a', // √
	"radicalex":                  '\uf8e5',
	"radoverssquare":             '\u33ae', // ㎮
	"radoverssquaredsquare":      '\u33af', // ㎯
	"radsquare":                  '\u33ad', // ㎭
	"rafe":                       '\u05bf', // ֿ
	"rafehebrew":                 '\u05bf', // ֿ
	"ragujarati":                 '\u0ab0', // ર
	"ragurmukhi":                 '\u0a30', // ਰ
	"rahiragana":                 '\u3089', // ら
	"rakatakana":                 '\u30e9', // ラ
	"rakatakanahalfwidth":        '\uff97', // ﾗ
	"ralowerdiagonalbengali":     '\u09f1', // ৱ
	"ramiddlediagonalbengali":    '\u09f0', // ৰ
	"ramshorn":                   '\u0264', // ɤ
	"ratio":                      '\u2236', // ∶
	"rbopomofo":                  '\u3116', // ㄖ
	"rcaron":                     '\u0159', // ř
	"rcedilla":                   '\u0157', // ŗ
	"rcircle":                    '\u24e1', // ⓡ
	"rcommaaccent":               '\u0157', // ŗ
	"rdblgrave":                  '\u0211', // ȑ
	"rdotaccent":                 '\u1e59', // ṙ
	"rdotbelow":                  '\u1e5b', // ṛ
	"rdotbelowmacron":            '\u1e5d', // ṝ
	"referencemark":              '\u203b', // ※
	"reflexsubset":               '\u2286', // ⊆
	"reflexsuperset":             '\u2287', // ⊇
	"registered":                 '\u00ae', // ®
	"registersans":               '\uf8e8',
	"registerserif":              '\uf6da',
	"reharabic":                  '\u0631', // ر
	"reharmenian":                '\u0580', // ր
	"rehfinalarabic":             '\ufeae', // ﺮ
	"rehiragana":                 '\u308c', // れ
	"rehyehaleflamarabic":        '\u0631', // ر
	"rekatakana":                 '\u30ec', // レ
	"rekatakanahalfwidth":        '\uff9a', // ﾚ
	"resh":                       '\u05e8', // ר
	"reshdageshhebrew":           '\ufb48', // רּ
	"reshhatafpatah":             '\u05e8', // ר
	"reshhatafpatahhebrew":       '\u05e8', // ר
	"reshhatafsegol":             '\u05e8', // ר
	"reshhatafsegolhebrew":       '\u05e8', // ר
	"reshhebrew":                 '\u05e8', // ר
	"reshhiriq":                  '\u05e8', // ר
	"reshhiriqhebrew":            '\u05e8', // ר
	"reshholam":                  '\u05e8', // ר
	"reshholamhebrew":            '\u05e8', // ר
	"reshpatah":                  '\u05e8', // ר
	"reshpatahhebrew":            '\u05e8', // ר
	"reshqamats":                 '\u05e8', // ר
	"reshqamatshebrew":           '\u05e8', // ר
	"reshqubuts":                 '\u05e8', // ר
	"reshqubutshebrew":           '\u05e8', // ר
	"reshsegol":                  '\u05e8', // ר
	"reshsegolhebrew":            '\u05e8', // ר
	"reshsheva":                  '\u05e8', // ר
	"reshshevahebrew":            '\u05e8', // ר
	"reshtsere":                  '\u05e8', // ר
	"reshtserehebrew":            '\u05e8', // ר
	"reversedtilde":              '\u223d', // ∽
	"reviahebrew":                '\u0597', // ֗
	"reviamugrashhebrew":         '\u0597', // ֗
	"revlogicalnot":              '\u2310', // ⌐
	"rfishhook":                  '\u027e', // ɾ
	"rfishhookreversed":          '\u027f', // ɿ
	"rhabengali":                 '\u09dd', // ঢ়
	"rhadeva":                    '\u095d', // ढ़
	"rho":                        '\u03c1', // ρ
	"rhook":                      '\u027d', // ɽ
	"rhookturned":                '\u027b', // ɻ
	"rhookturnedsuperior":        '\u02b5', // ʵ
	"rhosymbolgreek":             '\u03f1', // ϱ
	"rhotichookmod":              '\u02de', // ˞
	"rieulacirclekorean":         '\u3271', // ㉱
	"rieulaparenkorean":          '\u3211', // ㈑
	"rieulcirclekorean":          '\u3263', // ㉣
	"rieulhieuhkorean":           '\u3140', // ㅀ
	"rieulkiyeokkorean":          '\u313a', // ㄺ
	"rieulkiyeoksioskorean":      '\u3169', // ㅩ
	"rieulkorean":                '\u3139', // ㄹ
	"rieulmieumkorean":           '\u313b', // ㄻ
	"rieulpansioskorean":         '\u316c', // ㅬ
	"rieulparenkorean":           '\u3203', // ㈃
	"rieulphieuphkorean":         '\u313f', // ㄿ
	"rieulpieupkorean":           '\u313c', // ㄼ
	"rieulpieupsioskorean":       '\u316b', // ㅫ
	"rieulsioskorean":            '\u313d', // ㄽ
	"rieulthieuthkorean":         '\u313e', // ㄾ
	"rieultikeutkorean":          '\u316a', // ㅪ
	"rieulyeorinhieuhkorean":     '\u316d', // ㅭ
	"rightangle":                 '\u221f', // ∟
	"righttackbelowcmb":          '\u0319', // ̙
	"righttriangle":              '\u22bf', // ⊿
	"rihiragana":                 '\u308a', // り
	"rikatakana":                 '\u30ea', // リ
	"rikatakanahalfwidth":        '\uff98', // ﾘ
	"ring":                       '\u02da', // ˚
	"ringbelowcmb":               '\u0325', // ̥
	"ringcmb":                    '\u030a', // ̊
	"ringhalfleft":               '\u02bf', // ʿ
	"ringhalfleftarmenian":       '\u0559', // ՙ
	"ringhalfleftbelowcmb":       '\u031c', // ̜
	"ringhalfleftcentered":       '\u02d3', // ˓
	"ringhalfright":              '\u02be', // ʾ
	"ringhalfrightbelowcmb":      '\u0339', // ̹
	"ringhalfrightcentered":      '\u02d2', // ˒
	"rinvertedbreve":             '\u0213', // ȓ
	"rittorusquare":              '\u3351', // ㍑
	"rlinebelow":                 '\u1e5f', // ṟ
	"rlongleg":                   '\u027c', // ɼ
	"rlonglegturned":             '\u027a', // ɺ
	"rmonospace":                 '\uff52', // ｒ
	"rohiragana":                 '\u308d', // ろ
	"rokatakana":                 '\u30ed', // ロ
	"rokatakanahalfwidth":        '\uff9b', // ﾛ
	"roruathai":                  '\u0e23', // ร
	"rparen":                     '\u24ad', // ⒭
	"rrabengali":                 '\u09dc', // ড়
	"rradeva":                    '\u0931', // ऱ
	"rragurmukhi":                '\u0a5c', // ੜ
	"rreharabic":                 '\u0691', // ڑ
	"rrehfinalarabic":            '\ufb8d', // ﮍ
	"rrvocalicbengali":           '\u09e0', // ৠ
	"rrvocalicdeva":              '\u0960', // ॠ
	"rrvocalicgujarati":          '\u0ae0', // ૠ
	"rrvocalicvowelsignbengali":  '\u09c4', // ৄ
	"rrvocalicvowelsigndeva":     '\u0944', // ॄ
	"rrvocalicvowelsigngujarati": '\u0ac4', // ૄ
	"rsuperior":                  '\uf6f1',
	"rtblock":                    '\u2590', // ▐
	"rturned":                    '\u0279', // ɹ
	"rturnedsuperior":            '\u02b4', // ʴ
	"ruhiragana":                 '\u308b', // る
	"rukatakana":                 '\u30eb', // ル
	"rukatakanahalfwidth":        '\uff99', // ﾙ
	"rupeemarkbengali":           '\u09f2', // ৲
	"rupeesignbengali":           '\u09f3', // ৳
	"rupiah":                     '\uf6dd',
	"ruthai":                     '\u0e24', // ฤ
	"rvocalicbengali":            '\u098b', // ঋ
	"rvocalicdeva":               '\u090b', // ऋ
	"rvocalicgujarati":           '\u0a8b', // ઋ
	"rvocalicvowelsignbengali":   '\u09c3', // ৃ
	"rvocalicvowelsigndeva":      '\u0943', // ृ
	"rvocalicvowelsigngujarati":  '\u0ac3', // ૃ
	"s":                               's',      // s
	"sabengali":                       '\u09b8', // স
	"sacute":                          '\u015b', // ś
	"sacutedotaccent":                 '\u1e65', // ṥ
	"sadarabic":                       '\u0635', // ص
	"sadeva":                          '\u0938', // स
	"sadfinalarabic":                  '\ufeba', // ﺺ
	"sadinitialarabic":                '\ufebb', // ﺻ
	"sadmedialarabic":                 '\ufebc', // ﺼ
	"sagujarati":                      '\u0ab8', // સ
	"sagurmukhi":                      '\u0a38', // ਸ
	"sahiragana":                      '\u3055', // さ
	"sakatakana":                      '\u30b5', // サ
	"sakatakanahalfwidth":             '\uff7b', // ｻ
	"sallallahoualayhewasallamarabic": '\ufdfa', // ﷺ
	"samekh":                                  '\u05e1', // ס
	"samekhdagesh":                            '\ufb41', // סּ
	"samekhdageshhebrew":                      '\ufb41', // סּ
	"samekhhebrew":                            '\u05e1', // ס
	"saraaathai":                              '\u0e32', // า
	"saraaethai":                              '\u0e41', // แ
	"saraaimaimalaithai":                      '\u0e44', // ไ
	"saraaimaimuanthai":                       '\u0e43', // ใ
	"saraamthai":                              '\u0e33', // ำ
	"saraathai":                               '\u0e30', // ะ
	"saraethai":                               '\u0e40', // เ
	"saraiileftthai":                          '\uf886',
	"saraiithai":                              '\u0e35', // ี
	"saraileftthai":                           '\uf885',
	"saraithai":                               '\u0e34', // ิ
	"saraothai":                               '\u0e42', // โ
	"saraueeleftthai":                         '\uf888',
	"saraueethai":                             '\u0e37', // ื
	"saraueleftthai":                          '\uf887',
	"sarauethai":                              '\u0e36', // ึ
	"sarauthai":                               '\u0e38', // ุ
	"sarauuthai":                              '\u0e39', // ู
	"sbopomofo":                               '\u3119', // ㄙ
	"scaron":                                  '\u0161', // š
	"scarondotaccent":                         '\u1e67', // ṧ
	"scedilla":                                '\u015f', // ş
	"schwa":                                   '\u0259', // ə
	"schwacyrillic":                           '\u04d9', // ә
	"schwadieresiscyrillic":                   '\u04db', // ӛ
	"schwahook":                               '\u025a', // ɚ
	"scircle":                                 '\u24e2', // ⓢ
	"scircumflex":                             '\u015d', // ŝ
	"scommaaccent":                            '\u0219', // ș
	"sdotaccent":                              '\u1e61', // ṡ
	"sdotbelow":                               '\u1e63', // ṣ
	"sdotbelowdotaccent":                      '\u1e69', // ṩ
	"seagullbelowcmb":                         '\u033c', // ̼
	"second":                                  '\u2033', // ″
	"secondtonechinese":                       '\u02ca', // ˊ
	"section":                                 '\u00a7', // §
	"seenarabic":                              '\u0633', // س
	"seenfinalarabic":                         '\ufeb2', // ﺲ
	"seeninitialarabic":                       '\ufeb3', // ﺳ
	"seenmedialarabic":                        '\ufeb4', // ﺴ
	"segol":                                   '\u05b6', // ֶ
	"segol13":                                 '\u05b6', // ֶ
	"segol1f":                                 '\u05b6', // ֶ
	"segol2c":                                 '\u05b6', // ֶ
	"segolhebrew":                             '\u05b6', // ֶ
	"segolnarrowhebrew":                       '\u05b6', // ֶ
	"segolquarterhebrew":                      '\u05b6', // ֶ
	"segoltahebrew":                           '\u0592', // ֒
	"segolwidehebrew":                         '\u05b6', // ֶ
	"seharmenian":                             '\u057d', // ս
	"sehiragana":                              '\u305b', // せ
	"sekatakana":                              '\u30bb', // セ
	"sekatakanahalfwidth":                     '\uff7e', // ｾ
	"semicolon":                               ';',      // ;
	"semicolonarabic":                         '\u061b', // ؛
	"semicolonmonospace":                      '\uff1b', // ；
	"semicolonsmall":                          '\ufe54', // ﹔
	"semivoicedmarkkana":                      '\u309c', // ゜
	"semivoicedmarkkanahalfwidth":             '\uff9f', // ﾟ
	"sentisquare":                             '\u3322', // ㌢
	"sentosquare":                             '\u3323', // ㌣
	"seven":                                   '7',      // 7
	"sevenarabic":                             '\u0667', // ٧
	"sevenbengali":                            '\u09ed', // ৭
	"sevencircle":                             '\u2466', // ⑦
	"sevencircleinversesansserif":             '\u2790', // ➐
	"sevendeva":                               '\u096d', // ७
	"seveneighths":                            '\u215e', // ⅞
	"sevengujarati":                           '\u0aed', // ૭
	"sevengurmukhi":                           '\u0a6d', // ੭
	"sevenhackarabic":                         '\u0667', // ٧
	"sevenhangzhou":                           '\u3027', // 〧
	"sevenideographicparen":                   '\u3226', // ㈦
	"seveninferior":                           '\u2087', // ₇
	"sevenmonospace":                          '\uff17', // ７
	"sevenoldstyle":                           '\uf737',
	"sevenparen":                              '\u247a', // ⑺
	"sevenperiod":                             '\u248e', // ⒎
	"sevenpersian":                            '\u06f7', // ۷
	"sevenroman":                              '\u2176', // ⅶ
	"sevensuperior":                           '\u2077', // ⁷
	"seventeencircle":                         '\u2470', // ⑰
	"seventeenparen":                          '\u2484', // ⒄
	"seventeenperiod":                         '\u2498', // ⒘
	"seventhai":                               '\u0e57', // ๗
	"sfthyphen":                               '\u00ad',
	"shaarmenian":                             '\u0577', // շ
	"shabengali":                              '\u09b6', // শ
	"shacyrillic":                             '\u0448', // ш
	"shaddaarabic":                            '\u0651', // ّ
	"shaddadammaarabic":                       '\ufc61', // ﱡ
	"shaddadammatanarabic":                    '\ufc5e', // ﱞ
	"shaddafathaarabic":                       '\ufc60', // ﱠ
	"shaddafathatanarabic":                    '\u0651', // ّ
	"shaddakasraarabic":                       '\ufc62', // ﱢ
	"shaddakasratanarabic":                    '\ufc5f', // ﱟ
	"shade":                                   '\u2592', // ▒
	"shadedark":                               '\u2593', // ▓
	"shadelight":                              '\u2591', // ░
	"shademedium":                             '\u2592', // ▒
	"shadeva":                                 '\u0936', // श
	"shagujarati":                             '\u0ab6', // શ
	"shagurmukhi":                             '\u0a36', // ਸ਼
	"shalshelethebrew":                        '\u0593', // ֓
	"shbopomofo":                              '\u3115', // ㄕ
	"shchacyrillic":                           '\u0449', // щ
	"sheenarabic":                             '\u0634', // ش
	"sheenfinalarabic":                        '\ufeb6', // ﺶ
	"sheeninitialarabic":                      '\ufeb7', // ﺷ
	"sheenmedialarabic":                       '\ufeb8', // ﺸ
	"sheicoptic":                              '\u03e3', // ϣ
	"sheqel":                                  '\u20aa', // ₪
	"sheqelhebrew":                            '\u20aa', // ₪
	"sheva":                                   '\u05b0', // ְ
	"sheva115":                                '\u05b0', // ְ
	"sheva15":                                 '\u05b0', // ְ
	"sheva22":                                 '\u05b0', // ְ
	"sheva2e":                                 '\u05b0', // ְ
	"shevahebrew":                             '\u05b0', // ְ
	"shevanarrowhebrew":                       '\u05b0', // ְ
	"shevaquarterhebrew":                      '\u05b0', // ְ
	"shevawidehebrew":                         '\u05b0', // ְ
	"shhacyrillic":                            '\u04bb', // һ
	"shimacoptic":                             '\u03ed', // ϭ
	"shin":                                    '\u05e9', // ש
	"shindagesh":                              '\ufb49', // שּ
	"shindageshhebrew":                        '\ufb49', // שּ
	"shindageshshindot":                       '\ufb2c', // שּׁ
	"shindageshshindothebrew":                 '\ufb2c', // שּׁ
	"shindageshsindot":                        '\ufb2d', // שּׂ
	"shindageshsindothebrew":                  '\ufb2d', // שּׂ
	"shindothebrew":                           '\u05c1', // ׁ
	"shinhebrew":                              '\u05e9', // ש
	"shinshindot":                             '\ufb2a', // שׁ
	"shinshindothebrew":                       '\ufb2a', // שׁ
	"shinsindot":                              '\ufb2b', // שׂ
	"shinsindothebrew":                        '\ufb2b', // שׂ
	"shook":                                   '\u0282', // ʂ
	"sigma":                                   '\u03c3', // σ
	"sigma1":                                  '\u03c2', // ς
	"sigmafinal":                              '\u03c2', // ς
	"sigmalunatesymbolgreek":                  '\u03f2', // ϲ
	"sihiragana":                              '\u3057', // し
	"sikatakana":                              '\u30b7', // シ
	"sikatakanahalfwidth":                     '\uff7c', // ｼ
	"siluqhebrew":                             '\u05bd', // ֽ
	"siluqlefthebrew":                         '\u05bd', // ֽ
	"similar":                                 '\u223c', // ∼
	"sindothebrew":                            '\u05c2', // ׂ
	"siosacirclekorean":                       '\u3274', // ㉴
	"siosaparenkorean":                        '\u3214', // ㈔
	"sioscieuckorean":                         '\u317e', // ㅾ
	"sioscirclekorean":                        '\u3266', // ㉦
	"sioskiyeokkorean":                        '\u317a', // ㅺ
	"sioskorean":                              '\u3145', // ㅅ
	"siosnieunkorean":                         '\u317b', // ㅻ
	"siosparenkorean":                         '\u3206', // ㈆
	"siospieupkorean":                         '\u317d', // ㅽ
	"siostikeutkorean":                        '\u317c', // ㅼ
	"six":                                     '6',      // 6
	"sixarabic":                               '\u0666', // ٦
	"sixbengali":                              '\u09ec', // ৬
	"sixcircle":                               '\u2465', // ⑥
	"sixcircleinversesansserif":               '\u278f', // ➏
	"sixdeva":                                 '\u096c', // ६
	"sixgujarati":                             '\u0aec', // ૬
	"sixgurmukhi":                             '\u0a6c', // ੬
	"sixhackarabic":                           '\u0666', // ٦
	"sixhangzhou":                             '\u3026', // 〦
	"sixideographicparen":                     '\u3225', // ㈥
	"sixinferior":                             '\u2086', // ₆
	"sixmonospace":                            '\uff16', // ６
	"sixoldstyle":                             '\uf736',
	"sixparen":                                '\u2479', // ⑹
	"sixperiod":                               '\u248d', // ⒍
	"sixpersian":                              '\u06f6', // ۶
	"sixroman":                                '\u2175', // ⅵ
	"sixsuperior":                             '\u2076', // ⁶
	"sixteencircle":                           '\u246f', // ⑯
	"sixteencurrencydenominatorbengali":       '\u09f9', // ৹
	"sixteenparen":                            '\u2483', // ⒃
	"sixteenperiod":                           '\u2497', // ⒗
	"sixthai":                                 '\u0e56', // ๖
	"slash":                                   '/',      // /
	"slashmonospace":                          '\uff0f', // ／
	"slong":                                   '\u017f', // ſ
	"slongdotaccent":                          '\u1e9b', // ẛ
	"smileface":                               '\u263a', // ☺
	"smonospace":                              '\uff53', // ｓ
	"sofpasuqhebrew":                          '\u05c3', // ׃
	"softhyphen":                              '\u00ad',
	"softsigncyrillic":                        '\u044c', // ь
	"sohiragana":                              '\u305d', // そ
	"sokatakana":                              '\u30bd', // ソ
	"sokatakanahalfwidth":                     '\uff7f', // ｿ
	"soliduslongoverlaycmb":                   '\u0338', // ̸
	"solidusshortoverlaycmb":                  '\u0337', // ̷
	"sorusithai":                              '\u0e29', // ษ
	"sosalathai":                              '\u0e28', // ศ
	"sosothai":                                '\u0e0b', // ซ
	"sosuathai":                               '\u0e2a', // ส
	"space":                                   ' ',      //
	"spacehackarabic":                         ' ',      //
	"spade":                                   '\u2660', // ♠
	"spadesuitblack":                          '\u2660', // ♠
	"spadesuitwhite":                          '\u2664', // ♤
	"sparen":                                  '\u24ae', // ⒮
	"squarebelowcmb":                          '\u033b', // ̻
	"squarecc":                                '\u33c4', // ㏄
	"squarecm":                                '\u339d', // ㎝
	"squarediagonalcrosshatchfill":            '\u25a9', // ▩
	"squarehorizontalfill":                    '\u25a4', // ▤
	"squarekg":                                '\u338f', // ㎏
	"squarekm":                                '\u339e', // ㎞
	"squarekmcapital":                         '\u33ce', // ㏎
	"squareln":                                '\u33d1', // ㏑
	"squarelog":                               '\u33d2', // ㏒
	"squaremg":                                '\u338e', // ㎎
	"squaremil":                               '\u33d5', // ㏕
	"squaremm":                                '\u339c', // ㎜
	"squaremsquared":                          '\u33a1', // ㎡
	"squareorthogonalcrosshatchfill":          '\u25a6', // ▦
	"squareupperlefttolowerrightfill":         '\u25a7', // ▧
	"squareupperrighttolowerleftfill":         '\u25a8', // ▨
	"squareverticalfill":                      '\u25a5', // ▥
	"squarewhitewithsmallblack":               '\u25a3', // ▣
	"srsquare":                                '\u33db', // ㏛
	"ssabengali":                              '\u09b7', // ষ
	"ssadeva":                                 '\u0937', // ष
	"ssagujarati":                             '\u0ab7', // ષ
	"ssangcieuckorean":                        '\u3149', // ㅉ
	"ssanghieuhkorean":                        '\u3185', // ㆅ
	"ssangieungkorean":                        '\u3180', // ㆀ
	"ssangkiyeokkorean":                       '\u3132', // ㄲ
	"ssangnieunkorean":                        '\u3165', // ㅥ
	"ssangpieupkorean":                        '\u3143', // ㅃ
	"ssangsioskorean":                         '\u3146', // ㅆ
	"ssangtikeutkorean":                       '\u3138', // ㄸ
	"ssuperior":                               '\uf6f2',
	"sterling":                                '\u00a3', // £
	"sterlingmonospace":                       '\uffe1', // ￡
	"strokelongoverlaycmb":                    '\u0336', // ̶
	"strokeshortoverlaycmb":                   '\u0335', // ̵
	"subset":                                  '\u2282', // ⊂
	"subsetnotequal":                          '\u228a', // ⊊
	"subsetorequal":                           '\u2286', // ⊆
	"succeeds":                                '\u227b', // ≻
	"suchthat":                                '\u220b', // ∋
	"suhiragana":                              '\u3059', // す
	"sukatakana":                              '\u30b9', // ス
	"sukatakanahalfwidth":                     '\uff7d', // ｽ
	"sukunarabic":                             '\u0652', // ْ
	"summation":                               '\u2211', // ∑
	"sun":                                     '\u263c', // ☼
	"superset":                                '\u2283', // ⊃
	"supersetnotequal":                        '\u228b', // ⊋
	"supersetorequal":                         '\u2287', // ⊇
	"svsquare":                                '\u33dc', // ㏜
	"syouwaerasquare":                         '\u337c', // ㍼
	"t":                                       't',      // t
	"tabengali":                               '\u09a4', // ত
	"tackdown":                                '\u22a4', // ⊤
	"tackleft":                                '\u22a3', // ⊣
	"tadeva":                                  '\u0924', // त
	"tagujarati":                              '\u0aa4', // ત
	"tagurmukhi":                              '\u0a24', // ਤ
	"taharabic":                               '\u0637', // ط
	"tahfinalarabic":                          '\ufec2', // ﻂ
	"tahinitialarabic":                        '\ufec3', // ﻃ
	"tahiragana":                              '\u305f', // た
	"tahmedialarabic":                         '\ufec4', // ﻄ
	"taisyouerasquare":                        '\u337d', // ㍽
	"takatakana":                              '\u30bf', // タ
	"takatakanahalfwidth":                     '\uff80', // ﾀ
	"tatweelarabic":                           '\u0640', // ـ
	"tau":                                     '\u03c4', // τ
	"tav":                                     '\u05ea', // ת
	"tavdages":                                '\ufb4a', // תּ
	"tavdagesh":                               '\ufb4a', // תּ
	"tavdageshhebrew":                         '\ufb4a', // תּ
	"tavhebrew":                               '\u05ea', // ת
	"tbar":                                    '\u0167', // ŧ
	"tbopomofo":                               '\u310a', // ㄊ
	"tcaron":                                  '\u0165', // ť
	"tccurl":                                  '\u02a8', // ʨ
	"tcedilla":                                '\u0163', // ţ
	"tcheharabic":                             '\u0686', // چ
	"tchehfinalarabic":                        '\ufb7b', // ﭻ
	"tchehinitialarabic":                      '\ufb7c', // ﭼ
	"tchehmedialarabic":                       '\ufb7d', // ﭽ
	"tchehmeeminitialarabic":                  '\ufb7c', // ﭼ
	"tcircle":                                 '\u24e3', // ⓣ
	"tcircumflexbelow":                        '\u1e71', // ṱ
	"tcommaaccent":                            '\u0163', // ţ
	"tdieresis":                               '\u1e97', // ẗ
	"tdotaccent":                              '\u1e6b', // ṫ
	"tdotbelow":                               '\u1e6d', // ṭ
	"tecyrillic":                              '\u0442', // т
	"tedescendercyrillic":                     '\u04ad', // ҭ
	"teharabic":                               '\u062a', // ت
	"tehfinalarabic":                          '\ufe96', // ﺖ
	"tehhahinitialarabic":                     '\ufca2', // ﲢ
	"tehhahisolatedarabic":                    '\ufc0c', // ﰌ
	"tehinitialarabic":                        '\ufe97', // ﺗ
	"tehiragana":                              '\u3066', // て
	"tehjeeminitialarabic":                    '\ufca1', // ﲡ
	"tehjeemisolatedarabic":                   '\ufc0b', // ﰋ
	"tehmarbutaarabic":                        '\u0629', // ة
	"tehmarbutafinalarabic":                   '\ufe94', // ﺔ
	"tehmedialarabic":                         '\ufe98', // ﺘ
	"tehmeeminitialarabic":                    '\ufca4', // ﲤ
	"tehmeemisolatedarabic":                   '\ufc0e', // ﰎ
	"tehnoonfinalarabic":                      '\ufc73', // ﱳ
	"tekatakana":                              '\u30c6', // テ
	"tekatakanahalfwidth":                     '\uff83', // ﾃ
	"telephone":                               '\u2121', // ℡
	"telephoneblack":                          '\u260e', // ☎
	"telishagedolahebrew":                     '\u05a0', // ֠
	"telishaqetanahebrew":                     '\u05a9', // ֩
	"tencircle":                               '\u2469', // ⑩
	"tenideographicparen":                     '\u3229', // ㈩
	"tenparen":                                '\u247d', // ⑽
	"tenperiod":                               '\u2491', // ⒑
	"tenroman":                                '\u2179', // ⅹ
	"tesh":                                    '\u02a7', // ʧ
	"tet":                                     '\u05d8', // ט
	"tetdagesh":                               '\ufb38', // טּ
	"tetdageshhebrew":                         '\ufb38', // טּ
	"tethebrew":                               '\u05d8', // ט
	"tetsecyrillic":                           '\u04b5', // ҵ
	"tevirhebrew":                             '\u059b', // ֛
	"tevirlefthebrew":                         '\u059b', // ֛
	"thabengali":                              '\u09a5', // থ
	"thadeva":                                 '\u0925', // थ
	"thagujarati":                             '\u0aa5', // થ
	"thagurmukhi":                             '\u0a25', // ਥ
	"thalarabic":                              '\u0630', // ذ
	"thalfinalarabic":                         '\ufeac', // ﺬ
	"thanthakhatlowleftthai":                  '\uf898',
	"thanthakhatlowrightthai":                 '\uf897',
	"thanthakhatthai":                         '\u0e4c', // ์
	"thanthakhatupperleftthai":                '\uf896',
	"theharabic":                              '\u062b', // ث
	"thehfinalarabic":                         '\ufe9a', // ﺚ
	"thehinitialarabic":                       '\ufe9b', // ﺛ
	"thehmedialarabic":                        '\ufe9c', // ﺜ
	"thereexists":                             '\u2203', // ∃
	"therefore":                               '\u2234', // ∴
	"theta":                                   '\u03b8', // θ
	"theta1":                                  '\u03d1', // ϑ
	"thetasymbolgreek":                        '\u03d1', // ϑ
	"thieuthacirclekorean":                    '\u3279', // ㉹
	"thieuthaparenkorean":                     '\u3219', // ㈙
	"thieuthcirclekorean":                     '\u326b', // ㉫
	"thieuthkorean":                           '\u314c', // ㅌ
	"thieuthparenkorean":                      '\u320b', // ㈋
	"thirteencircle":                          '\u246c', // ⑬
	"thirteenparen":                           '\u2480', // ⒀
	"thirteenperiod":                          '\u2494', // ⒔
	"thonangmonthothai":                       '\u0e11', // ฑ
	"thook":                                   '\u01ad', // ƭ
	"thophuthaothai":                          '\u0e12', // ฒ
	"thorn":                                   '\u00fe', // þ
	"thothahanthai":                           '\u0e17', // ท
	"thothanthai":                             '\u0e10', // ฐ
	"thothongthai":                            '\u0e18', // ธ
	"thothungthai":                            '\u0e16', // ถ
	"thousandcyrillic":                        '\u0482', // ҂
	"thousandsseparatorarabic":                '\u066c', // ٬
	"thousandsseparatorpersian":               '\u066c', // ٬
	"three":                                   '3',      // 3
	"threearabic":                             '\u0663', // ٣
	"threebengali":                            '\u09e9', // ৩
	"threecircle":                             '\u2462', // ③
	"threecircleinversesansserif":             '\u278c', // ➌
	"threedeva":                               '\u0969', // ३
	"threeeighths":                            '\u215c', // ⅜
	"threegujarati":                           '\u0ae9', // ૩
	"threegurmukhi":                           '\u0a69', // ੩
	"threehackarabic":                         '\u0663', // ٣
	"threehangzhou":                           '\u3023', // 〣
	"threeideographicparen":                   '\u3222', // ㈢
	"threeinferior":                           '\u2083', // ₃
	"threemonospace":                          '\uff13', // ３
	"threenumeratorbengali":                   '\u09f6', // ৶
	"threeoldstyle":                           '\uf733',
	"threeparen":                              '\u2476', // ⑶
	"threeperiod":                             '\u248a', // ⒊
	"threepersian":                            '\u06f3', // ۳
	"threequarters":                           '\u00be', // ¾
	"threequartersemdash":                     '\uf6de',
	"threeroman":                              '\u2172', // ⅲ
	"threesuperior":                           '\u00b3', // ³
	"threethai":                               '\u0e53', // ๓
	"thzsquare":                               '\u3394', // ㎔
	"tihiragana":                              '\u3061', // ち
	"tikatakana":                              '\u30c1', // チ
	"tikatakanahalfwidth":                     '\uff81', // ﾁ
	"tikeutacirclekorean":                     '\u3270', // ㉰
	"tikeutaparenkorean":                      '\u3210', // ㈐
	"tikeutcirclekorean":                      '\u3262', // ㉢
	"tikeutkorean":                            '\u3137', // ㄷ
	"tikeutparenkorean":                       '\u3202', // ㈂
	"tilde":                                   '\u02dc', // ˜
	"tildebelowcmb":                           '\u0330', // ̰
	"tildecmb":                                '\u0303', // ̃
	"tildecomb":                               '\u0303', // ̃
	"tildedoublecmb":                          '\u0360', // ͠
	"tildeoperator":                           '\u223c', // ∼
	"tildeoverlaycmb":                         '\u0334', // ̴
	"tildeverticalcmb":                        '\u033e', // ̾
	"timescircle":                             '\u2297', // ⊗
	"tipehahebrew":                            '\u0596', // ֖
	"tipehalefthebrew":                        '\u0596', // ֖
	"tippigurmukhi":                           '\u0a70', // ੰ
	"titlocyrilliccmb":                        '\u0483', // ҃
	"tiwnarmenian":                            '\u057f', // տ
	"tlinebelow":                              '\u1e6f', // ṯ
	"tmonospace":                              '\uff54', // ｔ
	"toarmenian":                              '\u0569', // թ
	"tohiragana":                              '\u3068', // と
	"tokatakana":                              '\u30c8', // ト
	"tokatakanahalfwidth":                     '\uff84', // ﾄ
	"tonebarextrahighmod":                     '\u02e5', // ˥
	"tonebarextralowmod":                      '\u02e9', // ˩
	"tonebarhighmod":                          '\u02e6', // ˦
	"tonebarlowmod":                           '\u02e8', // ˨
	"tonebarmidmod":                           '\u02e7', // ˧
	"tonefive":                                '\u01bd', // ƽ
	"tonesix":                                 '\u0185', // ƅ
	"tonetwo":                                 '\u01a8', // ƨ
	"tonos":                                   '\u0384', // ΄
	"tonsquare":                               '\u3327', // ㌧
	"topatakthai":                             '\u0e0f', // ฏ
	"tortoiseshellbracketleft":                '\u3014', // 〔
	"tortoiseshellbracketleftsmall":           '\ufe5d', // ﹝
	"tortoiseshellbracketleftvertical":        '\ufe39', // ︹
	"tortoiseshellbracketright":               '\u3015', // 〕
	"tortoiseshellbracketrightsmall":          '\ufe5e', // ﹞
	"tortoiseshellbracketrightvertical":       '\ufe3a', // ︺
	"totaothai":                               '\u0e15', // ต
	"tpalatalhook":                            '\u01ab', // ƫ
	"tparen":                                  '\u24af', // ⒯
	"trademark":                               '\u2122', // ™
	"trademarksans":                           '\uf8ea',
	"trademarkserif":                          '\uf6db',
	"tretroflexhook":                          '\u0288', // ʈ
	"triagdn":                                 '\u25bc', // ▼
	"triaglf":                                 '\u25c4', // ◄
	"triagrt":                                 '\u25ba', // ►
	"triagup":                                 '\u25b2', // ▲
	"ts":                                      '\u02a6', // ʦ
	"tsadi":                                   '\u05e6', // צ
	"tsadidagesh":                             '\ufb46', // צּ
	"tsadidageshhebrew":                       '\ufb46', // צּ
	"tsadihebrew":                             '\u05e6', // צ
	"tsecyrillic":                             '\u0446', // ц
	"tsere":                                   '\u05b5', // ֵ
	"tsere12":                                 '\u05b5', // ֵ
	"tsere1e":                                 '\u05b5', // ֵ
	"tsere2b":                                 '\u05b5', // ֵ
	"tserehebrew":                             '\u05b5', // ֵ
	"tserenarrowhebrew":                       '\u05b5', // ֵ
	"tserequarterhebrew":                      '\u05b5', // ֵ
	"tserewidehebrew":                         '\u05b5', // ֵ
	"tshecyrillic":                            '\u045b', // ћ
	"tsuperior":                               '\uf6f3',
	"ttabengali":                              '\u099f', // ট
	"ttadeva":                                 '\u091f', // ट
	"ttagujarati":                             '\u0a9f', // ટ
	"ttagurmukhi":                             '\u0a1f', // ਟ
	"tteharabic":                              '\u0679', // ٹ
	"ttehfinalarabic":                         '\ufb67', // ﭧ
	"ttehinitialarabic":                       '\ufb68', // ﭨ
	"ttehmedialarabic":                        '\ufb69', // ﭩ
	"tthabengali":                             '\u09a0', // ঠ
	"tthadeva":                                '\u0920', // ठ
	"tthagujarati":                            '\u0aa0', // ઠ
	"tthagurmukhi":                            '\u0a20', // ਠ
	"tturned":                                 '\u0287', // ʇ
	"tuhiragana":                              '\u3064', // つ
	"tukatakana":                              '\u30c4', // ツ
	"tukatakanahalfwidth":                     '\uff82', // ﾂ
	"tusmallhiragana":                         '\u3063', // っ
	"tusmallkatakana":                         '\u30c3', // ッ
	"tusmallkatakanahalfwidth":                '\uff6f', // ｯ
	"twelvecircle":                            '\u246b', // ⑫
	"twelveparen":                             '\u247f', // ⑿
	"twelveperiod":                            '\u2493', // ⒓
	"twelveroman":                             '\u217b', // ⅻ
	"twentycircle":                            '\u2473', // ⑳
	"twentyhangzhou":                          '\u5344', // 卄
	"twentyparen":                             '\u2487', // ⒇
	"twentyperiod":                            '\u249b', // ⒛
	"two":                                     '2',      // 2
	"twoarabic":                               '\u0662', // ٢
	"twobengali":                              '\u09e8', // ২
	"twocircle":                               '\u2461', // ②
	"twocircleinversesansserif":               '\u278b', // ➋
	"twodeva":                                 '\u0968', // २
	"twodotenleader":                          '\u2025', // ‥
	"twodotleader":                            '\u2025', // ‥
	"twodotleadervertical":                    '\ufe30', // ︰
	"twogujarati":                             '\u0ae8', // ૨
	"twogurmukhi":                             '\u0a68', // ੨
	"twohackarabic":                           '\u0662', // ٢
	"twohangzhou":                             '\u3022', // 〢
	"twoideographicparen":                     '\u3221', // ㈡
	"twoinferior":                             '\u2082', // ₂
	"twomonospace":                            '\uff12', // ２
	"twonumeratorbengali":                     '\u09f5', // ৵
	"twooldstyle":                             '\uf732',
	"twoparen":                                '\u2475', // ⑵
	"twoperiod":                               '\u2489', // ⒉
	"twopersian":                              '\u06f2', // ۲
	"tworoman":                                '\u2171', // ⅱ
	"twostroke":                               '\u01bb', // ƻ
	"twosuperior":                             '\u00b2', // ²
	"twothai":                                 '\u0e52', // ๒
	"twothirds":                               '\u2154', // ⅔
	"u":                                       'u',      // u
	"uacute":                                  '\u00fa', // ú
	"ubar":                                    '\u0289', // ʉ
	"ubengali":                                '\u0989', // উ
	"ubopomofo":                               '\u3128', // ㄨ
	"ubreve":                                  '\u016d', // ŭ
	"ucaron":                                  '\u01d4', // ǔ
	"ucircle":                                 '\u24e4', // ⓤ
	"ucircumflex":                             '\u00fb', // û
	"ucircumflexbelow":                        '\u1e77', // ṷ
	"ucyrillic":                               '\u0443', // у
	"udattadeva":                              '\u0951', // ॑
	"udblacute":                               '\u0171', // ű
	"udblgrave":                               '\u0215', // ȕ
	"udeva":                                   '\u0909', // उ
	"udieresis":                               '\u00fc', // ü
	"udieresisacute":                          '\u01d8', // ǘ
	"udieresisbelow":                          '\u1e73', // ṳ
	"udieresiscaron":                          '\u01da', // ǚ
	"udieresiscyrillic":                       '\u04f1', // ӱ
	"udieresisgrave":                          '\u01dc', // ǜ
	"udieresismacron":                         '\u01d6', // ǖ
	"udotbelow":                               '\u1ee5', // ụ
	"ugrave":                                  '\u00f9', // ù
	"ugujarati":                               '\u0a89', // ઉ
	"ugurmukhi":                               '\u0a09', // ਉ
	"uhiragana":                               '\u3046', // う
	"uhookabove":                              '\u1ee7', // ủ
	"uhorn":                                   '\u01b0', // ư
	"uhornacute":                              '\u1ee9', // ứ
	"uhorndotbelow":                           '\u1ef1', // ự
	"uhorngrave":                              '\u1eeb', // ừ
	"uhornhookabove":                          '\u1eed', // ử
	"uhorntilde":                              '\u1eef', // ữ
	"uhungarumlaut":                           '\u0171', // ű
	"uhungarumlautcyrillic":                   '\u04f3', // ӳ
	"uinvertedbreve":                          '\u0217', // ȗ
	"ukatakana":                               '\u30a6', // ウ
	"ukatakanahalfwidth":                      '\uff73', // ｳ
	"ukcyrillic":                              '\u0479', // ѹ
	"ukorean":                                 '\u315c', // ㅜ
	"umacron":                                 '\u016b', // ū
	"umacroncyrillic":                         '\u04ef', // ӯ
	"umacrondieresis":                         '\u1e7b', // ṻ
	"umatragurmukhi":                          '\u0a41', // ੁ
	"umonospace":                              '\uff55', // ｕ
	"underscore":                              '_',      // _
	"underscoredbl":                           '\u2017', // ‗
	"underscoremonospace":                     '\uff3f', // ＿
	"underscorevertical":                      '\ufe33', // ︳
	"underscorewavy":                          '\ufe4f', // ﹏
	"union":                                   '\u222a', // ∪
	"universal":                               '\u2200', // ∀
	"uogonek":                                 '\u0173', // ų
	"uparen":                                  '\u24b0', // ⒰
	"upblock":                                 '\u2580', // ▀
	"upperdothebrew":                          '\u05c4', // ׄ
	"upsilon":                                 '\u03c5', // υ
	"upsilondieresis":                         '\u03cb', // ϋ
	"upsilondieresistonos":                    '\u03b0', // ΰ
	"upsilonlatin":                            '\u028a', // ʊ
	"upsilontonos":                            '\u03cd', // ύ
	"uptackbelowcmb":                          '\u031d', // ̝
	"uptackmod":                               '\u02d4', // ˔
	"uragurmukhi":                             '\u0a73', // ੳ
	"uring":                                   '\u016f', // ů
	"ushortcyrillic":                          '\u045e', // ў
	"usmallhiragana":                          '\u3045', // ぅ
	"usmallkatakana":                          '\u30a5', // ゥ
	"usmallkatakanahalfwidth":                 '\uff69', // ｩ
	"ustraightcyrillic":                       '\u04af', // ү
	"ustraightstrokecyrillic":                 '\u04b1', // ұ
	"utilde":                                  '\u0169', // ũ
	"utildeacute":                             '\u1e79', // ṹ
	"utildebelow":                             '\u1e75', // ṵ
	"uubengali":                               '\u098a', // ঊ
	"uudeva":                                  '\u090a', // ऊ
	"uugujarati":                              '\u0a8a', // ઊ
	"uugurmukhi":                              '\u0a0a', // ਊ
	"uumatragurmukhi":                         '\u0a42', // ੂ
	"uuvowelsignbengali":                      '\u09c2', // ূ
	"uuvowelsigndeva":                         '\u0942', // ू
	"uuvowelsigngujarati":                     '\u0ac2', // ૂ
	"uvowelsignbengali":                       '\u09c1', // ু
	"uvowelsigndeva":                          '\u0941', // ु
	"uvowelsigngujarati":                      '\u0ac1', // ુ
	"v":                                       'v',      // v
	"vadeva":                                  '\u0935', // व
	"vagujarati":                              '\u0ab5', // વ
	"vagurmukhi":                              '\u0a35', // ਵ
	"vakatakana":                              '\u30f7', // ヷ
	"vav":                                     '\u05d5', // ו
	"vavdagesh":                               '\ufb35', // וּ
	"vavdagesh65":                             '\ufb35', // וּ
	"vavdageshhebrew":                         '\ufb35', // וּ
	"vavhebrew":                               '\u05d5', // ו
	"vavholam":                                '\ufb4b', // וֹ
	"vavholamhebrew":                          '\ufb4b', // וֹ
	"vavvavhebrew":                            '\u05f0', // װ
	"vavyodhebrew":                            '\u05f1', // ױ
	"vcircle":                                 '\u24e5', // ⓥ
	"vdotbelow":                               '\u1e7f', // ṿ
	"vecyrillic":                              '\u0432', // в
	"veharabic":                               '\u06a4', // ڤ
	"vehfinalarabic":                          '\ufb6b', // ﭫ
	"vehinitialarabic":                        '\ufb6c', // ﭬ
	"vehmedialarabic":                         '\ufb6d', // ﭭ
	"vekatakana":                              '\u30f9', // ヹ
	"venus":                                   '\u2640', // ♀
	"verticalbar":                             '|',      // |
	"verticallineabovecmb":                    '\u030d', // ̍
	"verticallinebelowcmb":                    '\u0329', // ̩
	"verticallinelowmod":                      '\u02cc', // ˌ
	"verticallinemod":                         '\u02c8', // ˈ
	"vewarmenian":                             '\u057e', // վ
	"vhook":                                   '\u028b', // ʋ
	"vikatakana":                              '\u30f8', // ヸ
	"viramabengali":                           '\u09cd', // ্
	"viramadeva":                              '\u094d', // ्
	"viramagujarati":                          '\u0acd', // ્
	"visargabengali":                          '\u0983', // ঃ
	"visargadeva":                             '\u0903', // ः
	"visargagujarati":                         '\u0a83', // ઃ
	"vmonospace":                              '\uff56', // ｖ
	"voarmenian":                              '\u0578', // ո
	"voicediterationhiragana":                 '\u309e', // ゞ
	"voicediterationkatakana":                 '\u30fe', // ヾ
	"voicedmarkkana":                          '\u309b', // ゛
	"voicedmarkkanahalfwidth":                 '\uff9e', // ﾞ
	"vokatakana":                              '\u30fa', // ヺ
	"vparen":                                  '\u24b1', // ⒱
	"vtilde":                                  '\u1e7d', // ṽ
	"vturned":                                 '\u028c', // ʌ
	"vuhiragana":                              '\u3094', // ゔ
	"vukatakana":                              '\u30f4', // ヴ
	"w":                                       'w',      // w
	"wacute":                                  '\u1e83', // ẃ
	"waekorean":                               '\u3159', // ㅙ
	"wahiragana":                              '\u308f', // わ
	"wakatakana":                              '\u30ef', // ワ
	"wakatakanahalfwidth":                     '\uff9c', // ﾜ
	"wakorean":                                '\u3158', // ㅘ
	"wasmallhiragana":                         '\u308e', // ゎ
	"wasmallkatakana":                         '\u30ee', // ヮ
	"wattosquare":                             '\u3357', // ㍗
	"wavedash":                                '\u301c', // 〜
	"wavyunderscorevertical":                  '\ufe34', // ︴
	"wawarabic":                               '\u0648', // و
	"wawfinalarabic":                          '\ufeee', // ﻮ
	"wawhamzaabovearabic":                     '\u0624', // ؤ
	"wawhamzaabovefinalarabic":                '\ufe86', // ﺆ
	"wbsquare":                                '\u33dd', // ㏝
	"wcircle":                                 '\u24e6', // ⓦ
	"wcircumflex":                             '\u0175', // ŵ
	"wdieresis":                               '\u1e85', // ẅ
	"wdotaccent":                              '\u1e87', // ẇ
	"wdotbelow":                               '\u1e89', // ẉ
	"wehiragana":                              '\u3091', // ゑ
	"weierstrass":                             '\u2118', // ℘
	"wekatakana":                              '\u30f1', // ヱ
	"wekorean":                                '\u315e', // ㅞ
	"weokorean":                               '\u315d', // ㅝ
	"wgrave":                                  '\u1e81', // ẁ
	"whitebullet":                             '\u25e6', // ◦
	"whitecircle":                             '\u25cb', // ○
	"whitecircleinverse":                      '\u25d9', // ◙
	"whitecornerbracketleft":                  '\u300e', // 『
	"whitecornerbracketleftvertical":          '\ufe43', // ﹃
	"whitecornerbracketright":                 '\u300f', // 』
	"whitecornerbracketrightvertical":         '\ufe44', // ﹄
	"whitediamond":                            '\u25c7', // ◇
	"whitediamondcontainingblacksmalldiamond": '\u25c8', // ◈
	"whitedownpointingsmalltriangle":          '\u25bf', // ▿
	"whitedownpointingtriangle":               '\u25bd', // ▽
	"whiteleftpointingsmalltriangle":          '\u25c3', // ◃
	"whiteleftpointingtriangle":               '\u25c1', // ◁
	"whitelenticularbracketleft":              '\u3016', // 〖
	"whitelenticularbracketright":             '\u3017', // 〗
	"whiterightpointingsmalltriangle":         '\u25b9', // ▹
	"whiterightpointingtriangle":              '\u25b7', // ▷
	"whitesmallsquare":                        '\u25ab', // ▫
	"whitesmilingface":                        '\u263a', // ☺
	"whitesquare":                             '\u25a1', // □
	"whitestar":                               '\u2606', // ☆
	"whitetelephone":                          '\u260f', // ☏
	"whitetortoiseshellbracketleft":           '\u3018', // 〘
	"whitetortoiseshellbracketright":          '\u3019', // 〙
	"whiteuppointingsmalltriangle":            '\u25b5', // ▵
	"whiteuppointingtriangle":                 '\u25b3', // △
	"wihiragana":                              '\u3090', // ゐ
	"wikatakana":                              '\u30f0', // ヰ
	"wikorean":                                '\u315f', // ㅟ
	"wmonospace":                              '\uff57', // ｗ
	"wohiragana":                              '\u3092', // を
	"wokatakana":                              '\u30f2', // ヲ
	"wokatakanahalfwidth":                     '\uff66', // ｦ
	"won":                        '\u20a9', // ₩
	"wonmonospace":               '\uffe6', // ￦
	"wowaenthai":                 '\u0e27', // ว
	"wparen":                     '\u24b2', // ⒲
	"wring":                      '\u1e98', // ẘ
	"wsuperior":                  '\u02b7', // ʷ
	"wturned":                    '\u028d', // ʍ
	"wynn":                       '\u01bf', // ƿ
	"x":                          'x',      // x
	"xabovecmb":                  '\u033d', // ̽
	"xbopomofo":                  '\u3112', // ㄒ
	"xcircle":                    '\u24e7', // ⓧ
	"xdieresis":                  '\u1e8d', // ẍ
	"xdotaccent":                 '\u1e8b', // ẋ
	"xeharmenian":                '\u056d', // խ
	"xi":                         '\u03be', // ξ
	"xmonospace":                 '\uff58', // ｘ
	"xparen":                     '\u24b3', // ⒳
	"xsuperior":                  '\u02e3', // ˣ
	"y":                          'y',      // y
	"yaadosquare":                '\u334e', // ㍎
	"yabengali":                  '\u09af', // য
	"yacute":                     '\u00fd', // ý
	"yadeva":                     '\u092f', // य
	"yaekorean":                  '\u3152', // ㅒ
	"yagujarati":                 '\u0aaf', // ય
	"yagurmukhi":                 '\u0a2f', // ਯ
	"yahiragana":                 '\u3084', // や
	"yakatakana":                 '\u30e4', // ヤ
	"yakatakanahalfwidth":        '\uff94', // ﾔ
	"yakorean":                   '\u3151', // ㅑ
	"yamakkanthai":               '\u0e4e', // ๎
	"yasmallhiragana":            '\u3083', // ゃ
	"yasmallkatakana":            '\u30e3', // ャ
	"yasmallkatakanahalfwidth":   '\uff6c', // ｬ
	"yatcyrillic":                '\u0463', // ѣ
	"ycircle":                    '\u24e8', // ⓨ
	"ycircumflex":                '\u0177', // ŷ
	"ydieresis":                  '\u00ff', // ÿ
	"ydotaccent":                 '\u1e8f', // ẏ
	"ydotbelow":                  '\u1ef5', // ỵ
	"yeharabic":                  '\u064a', // ي
	"yehbarreearabic":            '\u06d2', // ے
	"yehbarreefinalarabic":       '\ufbaf', // ﮯ
	"yehfinalarabic":             '\ufef2', // ﻲ
	"yehhamzaabovearabic":        '\u0626', // ئ
	"yehhamzaabovefinalarabic":   '\ufe8a', // ﺊ
	"yehhamzaaboveinitialarabic": '\ufe8b', // ﺋ
	"yehhamzaabovemedialarabic":  '\ufe8c', // ﺌ
	"yehinitialarabic":           '\ufef3', // ﻳ
	"yehmedialarabic":            '\ufef4', // ﻴ
	"yehmeeminitialarabic":       '\ufcdd', // ﳝ
	"yehmeemisolatedarabic":      '\ufc58', // ﱘ
	"yehnoonfinalarabic":         '\ufc94', // ﲔ
	"yehthreedotsbelowarabic":    '\u06d1', // ۑ
	"yekorean":                   '\u3156', // ㅖ
	"yen":                        '\u00a5', // ¥
	"yenmonospace":               '\uffe5', // ￥
	"yeokorean":                  '\u3155', // ㅕ
	"yeorinhieuhkorean":          '\u3186', // ㆆ
	"yerahbenyomohebrew":         '\u05aa', // ֪
	"yerahbenyomolefthebrew":     '\u05aa', // ֪
	"yericyrillic":               '\u044b', // ы
	"yerudieresiscyrillic":       '\u04f9', // ӹ
	"yesieungkorean":             '\u3181', // ㆁ
	"yesieungpansioskorean":      '\u3183', // ㆃ
	"yesieungsioskorean":         '\u3182', // ㆂ
	"yetivhebrew":                '\u059a', // ֚
	"ygrave":                     '\u1ef3', // ỳ
	"yhook":                      '\u01b4', // ƴ
	"yhookabove":                 '\u1ef7', // ỷ
	"yiarmenian":                 '\u0575', // յ
	"yicyrillic":                 '\u0457', // ї
	"yikorean":                   '\u3162', // ㅢ
	"yinyang":                    '\u262f', // ☯
	"yiwnarmenian":               '\u0582', // ւ
	"ymonospace":                 '\uff59', // ｙ
	"yod":                        '\u05d9', // י
	"yoddagesh":                  '\ufb39', // יּ
	"yoddageshhebrew":            '\ufb39', // יּ
	"yodhebrew":                  '\u05d9', // י
	"yodyodhebrew":               '\u05f2', // ײ
	"yodyodpatahhebrew":          '\ufb1f', // ײַ
	"yohiragana":                 '\u3088', // よ
	"yoikorean":                  '\u3189', // ㆉ
	"yokatakana":                 '\u30e8', // ヨ
	"yokatakanahalfwidth":        '\uff96', // ﾖ
	"yokorean":                   '\u315b', // ㅛ
	"yosmallhiragana":            '\u3087', // ょ
	"yosmallkatakana":            '\u30e7', // ョ
	"yosmallkatakanahalfwidth":   '\uff6e', // ｮ
	"yotgreek":                   '\u03f3', // ϳ
	"yoyaekorean":                '\u3188', // ㆈ
	"yoyakorean":                 '\u3187', // ㆇ
	"yoyakthai":                  '\u0e22', // ย
	"yoyingthai":                 '\u0e0d', // ญ
	"yparen":                     '\u24b4', // ⒴
	"ypogegrammeni":              '\u037a', // ͺ
	"ypogegrammenigreekcmb":      '\u0345', // ͅ
	"yr":                        '\u01a6', // Ʀ
	"yring":                     '\u1e99', // ẙ
	"ysuperior":                 '\u02b8', // ʸ
	"ytilde":                    '\u1ef9', // ỹ
	"yturned":                   '\u028e', // ʎ
	"yuhiragana":                '\u3086', // ゆ
	"yuikorean":                 '\u318c', // ㆌ
	"yukatakana":                '\u30e6', // ユ
	"yukatakanahalfwidth":       '\uff95', // ﾕ
	"yukorean":                  '\u3160', // ㅠ
	"yusbigcyrillic":            '\u046b', // ѫ
	"yusbigiotifiedcyrillic":    '\u046d', // ѭ
	"yuslittlecyrillic":         '\u0467', // ѧ
	"yuslittleiotifiedcyrillic": '\u0469', // ѩ
	"yusmallhiragana":           '\u3085', // ゅ
	"yusmallkatakana":           '\u30e5', // ュ
	"yusmallkatakanahalfwidth":  '\uff6d', // ｭ
	"yuyekorean":                '\u318b', // ㆋ
	"yuyeokorean":               '\u318a', // ㆊ
	"yyabengali":                '\u09df', // য়
	"yyadeva":                   '\u095f', // य़
	"z":                         'z',      // z
	"zaarmenian":                '\u0566', // զ
	"zacute":                    '\u017a', // ź
	"zadeva":                    '\u095b', // ज़
	"zagurmukhi":                '\u0a5b', // ਜ਼
	"zaharabic":                 '\u0638', // ظ
	"zahfinalarabic":            '\ufec6', // ﻆ
	"zahinitialarabic":          '\ufec7', // ﻇ
	"zahiragana":                '\u3056', // ざ
	"zahmedialarabic":           '\ufec8', // ﻈ
	"zainarabic":                '\u0632', // ز
	"zainfinalarabic":           '\ufeb0', // ﺰ
	"zakatakana":                '\u30b6', // ザ
	"zaqefgadolhebrew":          '\u0595', // ֕
	"zaqefqatanhebrew":          '\u0594', // ֔
	"zarqahebrew":               '\u0598', // ֘
	"zayin":                     '\u05d6', // ז
	"zayindagesh":               '\ufb36', // זּ
	"zayindageshhebrew":         '\ufb36', // זּ
	"zayinhebrew":               '\u05d6', // ז
	"zbopomofo":                 '\u3117', // ㄗ
	"zcaron":                    '\u017e', // ž
	"zcircle":                   '\u24e9', // ⓩ
	"zcircumflex":               '\u1e91', // ẑ
	"zcurl":                     '\u0291', // ʑ
	"zdot":                      '\u017c', // ż
	"zdotaccent":                '\u017c', // ż
	"zdotbelow":                 '\u1e93', // ẓ
	"zecyrillic":                '\u0437', // з
	"zedescendercyrillic":       '\u0499', // ҙ
	"zedieresiscyrillic":        '\u04df', // ӟ
	"zehiragana":                '\u305c', // ぜ
	"zekatakana":                '\u30bc', // ゼ
	"zero":                      '0',      // 0
	"zeroarabic":                '\u0660', // ٠
	"zerobengali":               '\u09e6', // ০
	"zerodeva":                  '\u0966', // ०
	"zerogujarati":              '\u0ae6', // ૦
	"zerogurmukhi":              '\u0a66', // ੦
	"zerohackarabic":            '\u0660', // ٠
	"zeroinferior":              '\u2080', // ₀
	"zeromonospace":             '\uff10', // ０
	"zerooldstyle":              '\uf730',
	"zeropersian":               '\u06f0', // ۰
	"zerosuperior":              '\u2070', // ⁰
	"zerothai":                  '\u0e50', // ๐
	"zerowidthjoiner":           '\ufeff',
	"zerowidthnonjoiner":        '\u200c',
	"zerowidthspace":            '\u200b',
	"zeta":                      '\u03b6', // ζ
	"zhbopomofo":                '\u3113', // ㄓ
	"zhearmenian":               '\u056a', // ժ
	"zhebrevecyrillic":          '\u04c2', // ӂ
	"zhecyrillic":               '\u0436', // ж
	"zhedescendercyrillic":      '\u0497', // җ
	"zhedieresiscyrillic":       '\u04dd', // ӝ
	"zihiragana":                '\u3058', // じ
	"zikatakana":                '\u30b8', // ジ
	"zinorhebrew":               '\u05ae', // ֮
	"zlinebelow":                '\u1e95', // ẕ
	"zmonospace":                '\uff5a', // ｚ
	"zohiragana":                '\u305e', // ぞ
	"zokatakana":                '\u30be', // ゾ
	"zparen":                    '\u24b5', // ⒵
	"zretroflexhook":            '\u0290', // ʐ
	"zstroke":                   '\u01b6', // ƶ
	"zuhiragana":                '\u305a', // ず
	"zukatakana":                '\u30ba', // ズ
}

var glyphlistRuneToGlyphMap = map[rune]string{ // 4281 entries
	'A':      "A",        // A
	'\u00c6': "AE",       // Æ
	'\u01fc': "AEacute",  // Ǽ
	'\u01e2': "AEmacron", // Ǣ
	'\uf7e6': "AEsmall",
	'\u00c1': "Aacute", // Á
	'\uf7e1': "Aacutesmall",
	'\u0102': "Abreve",               // Ă
	'\u1eae': "Abreveacute",          // Ắ
	'\u04d0': "Abrevecyrillic",       // Ӑ
	'\u1eb6': "Abrevedotbelow",       // Ặ
	'\u1eb0': "Abrevegrave",          // Ằ
	'\u1eb2': "Abrevehookabove",      // Ẳ
	'\u1eb4': "Abrevetilde",          // Ẵ
	'\u01cd': "Acaron",               // Ǎ
	'\u24b6': "Acircle",              // Ⓐ
	'\u00c2': "Acircumflex",          // Â
	'\u1ea4': "Acircumflexacute",     // Ấ
	'\u1eac': "Acircumflexdotbelow",  // Ậ
	'\u1ea6': "Acircumflexgrave",     // Ầ
	'\u1ea8': "Acircumflexhookabove", // Ẩ
	'\uf7e2': "Acircumflexsmall",
	'\u1eaa': "Acircumflextilde", // Ẫ
	'\uf6c9': "Acute",
	'\uf7b4': "Acutesmall",
	'\u0410': "Acyrillic",         // А
	'\u0200': "Adblgrave",         // Ȁ
	'\u00c4': "Adieresis",         // Ä
	'\u04d2': "Adieresiscyrillic", // Ӓ
	'\u01de': "Adieresismacron",   // Ǟ
	'\uf7e4': "Adieresissmall",
	'\u1ea0': "Adotbelow",  // Ạ
	'\u01e0': "Adotmacron", // Ǡ
	'\u00c0': "Agrave",     // À
	'\uf7e0': "Agravesmall",
	'\u1ea2': "Ahookabove",     // Ả
	'\u04d4': "Aiecyrillic",    // Ӕ
	'\u0202': "Ainvertedbreve", // Ȃ
	'\u0391': "Alpha",          // Α
	'\u0386': "Alphatonos",     // Ά
	'\u0100': "Amacron",        // Ā
	'\uff21': "Amonospace",     // Ａ
	'\u0104': "Aogonek",        // Ą
	'\u00c5': "Aring",          // Å
	'\u01fa': "Aringacute",     // Ǻ
	'\u1e00': "Aringbelow",     // Ḁ
	'\uf7e5': "Aringsmall",
	'\uf761': "Asmall",
	'\u00c3': "Atilde", // Ã
	'\uf7e3': "Atildesmall",
	'\u0531': "Aybarmenian", // Ա
	'B':      "B",           // B
	'\u24b7': "Bcircle",     // Ⓑ
	'\u1e02': "Bdotaccent",  // Ḃ
	'\u1e04': "Bdotbelow",   // Ḅ
	'\u0411': "Becyrillic",  // Б
	'\u0532': "Benarmenian", // Բ
	'\u0392': "Beta",        // Β
	'\u0181': "Bhook",       // Ɓ
	'\u1e06': "Blinebelow",  // Ḇ
	'\uff22': "Bmonospace",  // Ｂ
	'\uf6f4': "Brevesmall",
	'\uf762': "Bsmall",
	'\u0182': "Btopbar",    // Ƃ
	'C':      "C",          // C
	'\u053e': "Caarmenian", // Ծ
	'\u0106': "Cacute",     // Ć
	'\uf6ca': "Caron",
	'\uf6f5': "Caronsmall",
	'\u010c': "Ccaron",        // Č
	'\u00c7': "Ccedilla",      // Ç
	'\u1e08': "Ccedillaacute", // Ḉ
	'\uf7e7': "Ccedillasmall",
	'\u24b8': "Ccircle",     // Ⓒ
	'\u0108': "Ccircumflex", // Ĉ
	'\u010a': "Cdot",        // Ċ
	// '\u010a':    "Cdotaccent", // Ċ -- duplicate
	'\uf7b8': "Cedillasmall",
	'\u0549': "Chaarmenian",                   // Չ
	'\u04bc': "Cheabkhasiancyrillic",          // Ҽ
	'\u0427': "Checyrillic",                   // Ч
	'\u04be': "Chedescenderabkhasiancyrillic", // Ҿ
	'\u04b6': "Chedescendercyrillic",          // Ҷ
	'\u04f4': "Chedieresiscyrillic",           // Ӵ
	'\u0543': "Cheharmenian",                  // Ճ
	'\u04cb': "Chekhakassiancyrillic",         // Ӌ
	'\u04b8': "Cheverticalstrokecyrillic",     // Ҹ
	'\u03a7': "Chi",                           // Χ
	'\u0187': "Chook",                         // Ƈ
	'\uf6f6': "Circumflexsmall",
	'\uff23': "Cmonospace", // Ｃ
	'\u0551': "Coarmenian", // Ց
	'\uf763': "Csmall",
	'D':      "D",                // D
	'\u01f1': "DZ",               // Ǳ
	'\u01c4': "DZcaron",          // Ǆ
	'\u0534': "Daarmenian",       // Դ
	'\u0189': "Dafrican",         // Ɖ
	'\u010e': "Dcaron",           // Ď
	'\u1e10': "Dcedilla",         // Ḑ
	'\u24b9': "Dcircle",          // Ⓓ
	'\u1e12': "Dcircumflexbelow", // Ḓ
	'\u0110': "Dcroat",           // Đ
	'\u1e0a': "Ddotaccent",       // Ḋ
	'\u1e0c': "Ddotbelow",        // Ḍ
	'\u0414': "Decyrillic",       // Д
	'\u03ee': "Deicoptic",        // Ϯ
	'\u2206': "Delta",            // ∆
	'\u0394': "Deltagreek",       // Δ
	'\u018a': "Dhook",            // Ɗ
	'\uf6cb': "Dieresis",
	'\uf6cc': "DieresisAcute",
	'\uf6cd': "DieresisGrave",
	'\uf7a8': "Dieresissmall",
	'\u03dc': "Digammagreek", // Ϝ
	'\u0402': "Djecyrillic",  // Ђ
	'\u1e0e': "Dlinebelow",   // Ḏ
	'\uff24': "Dmonospace",   // Ｄ
	'\uf6f7': "Dotaccentsmall",
	// '\u0110':    "Dslash", // Đ -- duplicate
	'\uf764': "Dsmall",
	'\u018b': "Dtopbar",              // Ƌ
	'\u01f2': "Dz",                   // ǲ
	'\u01c5': "Dzcaron",              // ǅ
	'\u04e0': "Dzeabkhasiancyrillic", // Ӡ
	'\u0405': "Dzecyrillic",          // Ѕ
	'\u040f': "Dzhecyrillic",         // Џ
	'E':      "E",                    // E
	'\u00c9': "Eacute",               // É
	'\uf7e9': "Eacutesmall",
	'\u0114': "Ebreve",               // Ĕ
	'\u011a': "Ecaron",               // Ě
	'\u1e1c': "Ecedillabreve",        // Ḝ
	'\u0535': "Echarmenian",          // Ե
	'\u24ba': "Ecircle",              // Ⓔ
	'\u00ca': "Ecircumflex",          // Ê
	'\u1ebe': "Ecircumflexacute",     // Ế
	'\u1e18': "Ecircumflexbelow",     // Ḙ
	'\u1ec6': "Ecircumflexdotbelow",  // Ệ
	'\u1ec0': "Ecircumflexgrave",     // Ề
	'\u1ec2': "Ecircumflexhookabove", // Ể
	'\uf7ea': "Ecircumflexsmall",
	'\u1ec4': "Ecircumflextilde", // Ễ
	'\u0404': "Ecyrillic",        // Є
	'\u0204': "Edblgrave",        // Ȅ
	'\u00cb': "Edieresis",        // Ë
	'\uf7eb': "Edieresissmall",
	'\u0116': "Edot", // Ė
	// '\u0116':    "Edotaccent", // Ė -- duplicate
	'\u1eb8': "Edotbelow",  // Ẹ
	'\u0424': "Efcyrillic", // Ф
	'\u00c8': "Egrave",     // È
	'\uf7e8': "Egravesmall",
	'\u0537': "Eharmenian",          // Է
	'\u1eba': "Ehookabove",          // Ẻ
	'\u2167': "Eightroman",          // Ⅷ
	'\u0206': "Einvertedbreve",      // Ȇ
	'\u0464': "Eiotifiedcyrillic",   // Ѥ
	'\u041b': "Elcyrillic",          // Л
	'\u216a': "Elevenroman",         // Ⅺ
	'\u0112': "Emacron",             // Ē
	'\u1e16': "Emacronacute",        // Ḗ
	'\u1e14': "Emacrongrave",        // Ḕ
	'\u041c': "Emcyrillic",          // М
	'\uff25': "Emonospace",          // Ｅ
	'\u041d': "Encyrillic",          // Н
	'\u04a2': "Endescendercyrillic", // Ң
	'\u014a': "Eng",                 // Ŋ
	'\u04a4': "Enghecyrillic",       // Ҥ
	'\u04c7': "Enhookcyrillic",      // Ӈ
	'\u0118': "Eogonek",             // Ę
	'\u0190': "Eopen",               // Ɛ
	'\u0395': "Epsilon",             // Ε
	'\u0388': "Epsilontonos",        // Έ
	'\u0420': "Ercyrillic",          // Р
	'\u018e': "Ereversed",           // Ǝ
	'\u042d': "Ereversedcyrillic",   // Э
	'\u0421': "Escyrillic",          // С
	'\u04aa': "Esdescendercyrillic", // Ҫ
	'\u01a9': "Esh",                 // Ʃ
	'\uf765': "Esmall",
	'\u0397': "Eta",        // Η
	'\u0538': "Etarmenian", // Ը
	'\u0389': "Etatonos",   // Ή
	'\u00d0': "Eth",        // Ð
	'\uf7f0': "Ethsmall",
	'\u1ebc': "Etilde",       // Ẽ
	'\u1e1a': "Etildebelow",  // Ḛ
	'\u20ac': "Euro",         // €
	'\u01b7': "Ezh",          // Ʒ
	'\u01ee': "Ezhcaron",     // Ǯ
	'\u01b8': "Ezhreversed",  // Ƹ
	'F':      "F",            // F
	'\u24bb': "Fcircle",      // Ⓕ
	'\u1e1e': "Fdotaccent",   // Ḟ
	'\u0556': "Feharmenian",  // Ֆ
	'\u03e4': "Feicoptic",    // Ϥ
	'\u0191': "Fhook",        // Ƒ
	'\u0472': "Fitacyrillic", // Ѳ
	'\u2164': "Fiveroman",    // Ⅴ
	'\uff26': "Fmonospace",   // Ｆ
	'\u2163': "Fourroman",    // Ⅳ
	'\uf766': "Fsmall",
	'G':      "G",            // G
	'\u3387': "GBsquare",     // ㎇
	'\u01f4': "Gacute",       // Ǵ
	'\u0393': "Gamma",        // Γ
	'\u0194': "Gammaafrican", // Ɣ
	'\u03ea': "Gangiacoptic", // Ϫ
	'\u011e': "Gbreve",       // Ğ
	'\u01e6': "Gcaron",       // Ǧ
	'\u0122': "Gcedilla",     // Ģ
	'\u24bc': "Gcircle",      // Ⓖ
	'\u011c': "Gcircumflex",  // Ĝ
	// '\u0122':    "Gcommaaccent", // Ģ -- duplicate
	'\u0120': "Gdot", // Ġ
	// '\u0120':    "Gdotaccent", // Ġ -- duplicate
	'\u0413': "Gecyrillic",            // Г
	'\u0542': "Ghadarmenian",          // Ղ
	'\u0494': "Ghemiddlehookcyrillic", // Ҕ
	'\u0492': "Ghestrokecyrillic",     // Ғ
	'\u0490': "Gheupturncyrillic",     // Ґ
	'\u0193': "Ghook",                 // Ɠ
	'\u0533': "Gimarmenian",           // Գ
	'\u0403': "Gjecyrillic",           // Ѓ
	'\u1e20': "Gmacron",               // Ḡ
	'\uff27': "Gmonospace",            // Ｇ
	'\uf6ce': "Grave",
	'\uf760': "Gravesmall",
	'\uf767': "Gsmall",
	'\u029b': "Gsmallhook",          // ʛ
	'\u01e4': "Gstroke",             // Ǥ
	'H':      "H",                   // H
	'\u25cf': "H18533",              // ●
	'\u25aa': "H18543",              // ▪
	'\u25ab': "H18551",              // ▫
	'\u25a1': "H22073",              // □
	'\u33cb': "HPsquare",            // ㏋
	'\u04a8': "Haabkhasiancyrillic", // Ҩ
	'\u04b2': "Hadescendercyrillic", // Ҳ
	'\u042a': "Hardsigncyrillic",    // Ъ
	'\u0126': "Hbar",                // Ħ
	'\u1e2a': "Hbrevebelow",         // Ḫ
	'\u1e28': "Hcedilla",            // Ḩ
	'\u24bd': "Hcircle",             // Ⓗ
	'\u0124': "Hcircumflex",         // Ĥ
	'\u1e26': "Hdieresis",           // Ḧ
	'\u1e22': "Hdotaccent",          // Ḣ
	'\u1e24': "Hdotbelow",           // Ḥ
	'\uff28': "Hmonospace",          // Ｈ
	'\u0540': "Hoarmenian",          // Հ
	'\u03e8': "Horicoptic",          // Ϩ
	'\uf768': "Hsmall",
	'\uf6cf': "Hungarumlaut",
	'\uf6f8': "Hungarumlautsmall",
	'\u3390': "Hzsquare",   // ㎐
	'I':      "I",          // I
	'\u042f': "IAcyrillic", // Я
	'\u0132': "IJ",         // Ĳ
	'\u042e': "IUcyrillic", // Ю
	'\u00cd': "Iacute",     // Í
	'\uf7ed': "Iacutesmall",
	'\u012c': "Ibreve",      // Ĭ
	'\u01cf': "Icaron",      // Ǐ
	'\u24be': "Icircle",     // Ⓘ
	'\u00ce': "Icircumflex", // Î
	'\uf7ee': "Icircumflexsmall",
	'\u0406': "Icyrillic",         // І
	'\u0208': "Idblgrave",         // Ȉ
	'\u00cf': "Idieresis",         // Ï
	'\u1e2e': "Idieresisacute",    // Ḯ
	'\u04e4': "Idieresiscyrillic", // Ӥ
	'\uf7ef': "Idieresissmall",
	'\u0130': "Idot", // İ
	// '\u0130':    "Idotaccent", // İ -- duplicate
	'\u1eca': "Idotbelow",       // Ị
	'\u04d6': "Iebrevecyrillic", // Ӗ
	'\u0415': "Iecyrillic",      // Е
	'\u2111': "Ifraktur",        // ℑ
	'\u00cc': "Igrave",          // Ì
	'\uf7ec': "Igravesmall",
	'\u1ec8': "Ihookabove",      // Ỉ
	'\u0418': "Iicyrillic",      // И
	'\u020a': "Iinvertedbreve",  // Ȋ
	'\u0419': "Iishortcyrillic", // Й
	'\u012a': "Imacron",         // Ī
	'\u04e2': "Imacroncyrillic", // Ӣ
	'\uff29': "Imonospace",      // Ｉ
	'\u053b': "Iniarmenian",     // Ի
	'\u0401': "Iocyrillic",      // Ё
	'\u012e': "Iogonek",         // Į
	'\u0399': "Iota",            // Ι
	'\u0196': "Iotaafrican",     // Ɩ
	'\u03aa': "Iotadieresis",    // Ϊ
	'\u038a': "Iotatonos",       // Ί
	'\uf769': "Ismall",
	'\u0197': "Istroke",                 // Ɨ
	'\u0128': "Itilde",                  // Ĩ
	'\u1e2c': "Itildebelow",             // Ḭ
	'\u0474': "Izhitsacyrillic",         // Ѵ
	'\u0476': "Izhitsadblgravecyrillic", // Ѷ
	'J':      "J",                       // J
	'\u0541': "Jaarmenian",              // Ձ
	'\u24bf': "Jcircle",                 // Ⓙ
	'\u0134': "Jcircumflex",             // Ĵ
	'\u0408': "Jecyrillic",              // Ј
	'\u054b': "Jheharmenian",            // Ջ
	'\uff2a': "Jmonospace",              // Ｊ
	'\uf76a': "Jsmall",
	'K':      "K",                        // K
	'\u3385': "KBsquare",                 // ㎅
	'\u33cd': "KKsquare",                 // ㏍
	'\u04a0': "Kabashkircyrillic",        // Ҡ
	'\u1e30': "Kacute",                   // Ḱ
	'\u041a': "Kacyrillic",               // К
	'\u049a': "Kadescendercyrillic",      // Қ
	'\u04c3': "Kahookcyrillic",           // Ӄ
	'\u039a': "Kappa",                    // Κ
	'\u049e': "Kastrokecyrillic",         // Ҟ
	'\u049c': "Kaverticalstrokecyrillic", // Ҝ
	'\u01e8': "Kcaron",                   // Ǩ
	'\u0136': "Kcedilla",                 // Ķ
	'\u24c0': "Kcircle",                  // Ⓚ
	// '\u0136':    "Kcommaaccent", // Ķ -- duplicate
	'\u1e32': "Kdotbelow",     // Ḳ
	'\u0554': "Keharmenian",   // Ք
	'\u053f': "Kenarmenian",   // Կ
	'\u0425': "Khacyrillic",   // Х
	'\u03e6': "Kheicoptic",    // Ϧ
	'\u0198': "Khook",         // Ƙ
	'\u040c': "Kjecyrillic",   // Ќ
	'\u1e34': "Klinebelow",    // Ḵ
	'\uff2b': "Kmonospace",    // Ｋ
	'\u0480': "Koppacyrillic", // Ҁ
	'\u03de': "Koppagreek",    // Ϟ
	'\u046e': "Ksicyrillic",   // Ѯ
	'\uf76b': "Ksmall",
	'L':      "L",  // L
	'\u01c7': "LJ", // Ǉ
	'\uf6bf': "LL",
	'\u0139': "Lacute",           // Ĺ
	'\u039b': "Lambda",           // Λ
	'\u013d': "Lcaron",           // Ľ
	'\u013b': "Lcedilla",         // Ļ
	'\u24c1': "Lcircle",          // Ⓛ
	'\u1e3c': "Lcircumflexbelow", // Ḽ
	// '\u013b':    "Lcommaaccent", // Ļ -- duplicate
	'\u013f': "Ldot", // Ŀ
	// '\u013f':    "Ldotaccent", // Ŀ -- duplicate
	'\u1e36': "Ldotbelow",       // Ḷ
	'\u1e38': "Ldotbelowmacron", // Ḹ
	'\u053c': "Liwnarmenian",    // Լ
	'\u01c8': "Lj",              // ǈ
	'\u0409': "Ljecyrillic",     // Љ
	'\u1e3a': "Llinebelow",      // Ḻ
	'\uff2c': "Lmonospace",      // Ｌ
	'\u0141': "Lslash",          // Ł
	'\uf6f9': "Lslashsmall",
	'\uf76c': "Lsmall",
	'M':      "M",        // M
	'\u3386': "MBsquare", // ㎆
	'\uf6d0': "Macron",
	'\uf7af': "Macronsmall",
	'\u1e3e': "Macute",      // Ḿ
	'\u24c2': "Mcircle",     // Ⓜ
	'\u1e40': "Mdotaccent",  // Ṁ
	'\u1e42': "Mdotbelow",   // Ṃ
	'\u0544': "Menarmenian", // Մ
	'\uff2d': "Mmonospace",  // Ｍ
	'\uf76d': "Msmall",
	'\u019c': "Mturned",          // Ɯ
	'\u039c': "Mu",               // Μ
	'N':      "N",                // N
	'\u01ca': "NJ",               // Ǌ
	'\u0143': "Nacute",           // Ń
	'\u0147': "Ncaron",           // Ň
	'\u0145': "Ncedilla",         // Ņ
	'\u24c3': "Ncircle",          // Ⓝ
	'\u1e4a': "Ncircumflexbelow", // Ṋ
	// '\u0145':    "Ncommaaccent", // Ņ -- duplicate
	'\u1e44': "Ndotaccent",  // Ṅ
	'\u1e46': "Ndotbelow",   // Ṇ
	'\u019d': "Nhookleft",   // Ɲ
	'\u2168': "Nineroman",   // Ⅸ
	'\u01cb': "Nj",          // ǋ
	'\u040a': "Njecyrillic", // Њ
	'\u1e48': "Nlinebelow",  // Ṉ
	'\uff2e': "Nmonospace",  // Ｎ
	'\u0546': "Nowarmenian", // Ն
	'\uf76e': "Nsmall",
	'\u00d1': "Ntilde", // Ñ
	'\uf7f1': "Ntildesmall",
	'\u039d': "Nu", // Ν
	'O':      "O",  // O
	'\u0152': "OE", // Œ
	'\uf6fa': "OEsmall",
	'\u00d3': "Oacute", // Ó
	'\uf7f3': "Oacutesmall",
	'\u04e8': "Obarredcyrillic",         // Ө
	'\u04ea': "Obarreddieresiscyrillic", // Ӫ
	'\u014e': "Obreve",                  // Ŏ
	'\u01d1': "Ocaron",                  // Ǒ
	'\u019f': "Ocenteredtilde",          // Ɵ
	'\u24c4': "Ocircle",                 // Ⓞ
	'\u00d4': "Ocircumflex",             // Ô
	'\u1ed0': "Ocircumflexacute",        // Ố
	'\u1ed8': "Ocircumflexdotbelow",     // Ộ
	'\u1ed2': "Ocircumflexgrave",        // Ồ
	'\u1ed4': "Ocircumflexhookabove",    // Ổ
	'\uf7f4': "Ocircumflexsmall",
	'\u1ed6': "Ocircumflextilde",  // Ỗ
	'\u041e': "Ocyrillic",         // О
	'\u0150': "Odblacute",         // Ő
	'\u020c': "Odblgrave",         // Ȍ
	'\u00d6': "Odieresis",         // Ö
	'\u04e6': "Odieresiscyrillic", // Ӧ
	'\uf7f6': "Odieresissmall",
	'\u1ecc': "Odotbelow", // Ọ
	'\uf6fb': "Ogoneksmall",
	'\u00d2': "Ograve", // Ò
	'\uf7f2': "Ogravesmall",
	'\u0555': "Oharmenian",     // Օ
	'\u2126': "Ohm",            // Ω
	'\u1ece': "Ohookabove",     // Ỏ
	'\u01a0': "Ohorn",          // Ơ
	'\u1eda': "Ohornacute",     // Ớ
	'\u1ee2': "Ohorndotbelow",  // Ợ
	'\u1edc': "Ohorngrave",     // Ờ
	'\u1ede': "Ohornhookabove", // Ở
	'\u1ee0': "Ohorntilde",     // Ỡ
	// '\u0150':    "Ohungarumlaut", // Ő -- duplicate
	'\u01a2': "Oi",             // Ƣ
	'\u020e': "Oinvertedbreve", // Ȏ
	'\u014c': "Omacron",        // Ō
	'\u1e52': "Omacronacute",   // Ṓ
	'\u1e50': "Omacrongrave",   // Ṑ
	// '\u2126':    "Omega", // Ω -- duplicate
	'\u0460': "Omegacyrillic",      // Ѡ
	'\u03a9': "Omegagreek",         // Ω
	'\u047a': "Omegaroundcyrillic", // Ѻ
	'\u047c': "Omegatitlocyrillic", // Ѽ
	'\u038f': "Omegatonos",         // Ώ
	'\u039f': "Omicron",            // Ο
	'\u038c': "Omicrontonos",       // Ό
	'\uff2f': "Omonospace",         // Ｏ
	'\u2160': "Oneroman",           // Ⅰ
	'\u01ea': "Oogonek",            // Ǫ
	'\u01ec': "Oogonekmacron",      // Ǭ
	'\u0186': "Oopen",              // Ɔ
	'\u00d8': "Oslash",             // Ø
	'\u01fe': "Oslashacute",        // Ǿ
	'\uf7f8': "Oslashsmall",
	'\uf76f': "Osmall",
	// '\u01fe':    "Ostrokeacute", // Ǿ -- duplicate
	'\u047e': "Otcyrillic",     // Ѿ
	'\u00d5': "Otilde",         // Õ
	'\u1e4c': "Otildeacute",    // Ṍ
	'\u1e4e': "Otildedieresis", // Ṏ
	'\uf7f5': "Otildesmall",
	'P':      "P",                    // P
	'\u1e54': "Pacute",               // Ṕ
	'\u24c5': "Pcircle",              // Ⓟ
	'\u1e56': "Pdotaccent",           // Ṗ
	'\u041f': "Pecyrillic",           // П
	'\u054a': "Peharmenian",          // Պ
	'\u04a6': "Pemiddlehookcyrillic", // Ҧ
	'\u03a6': "Phi",                  // Φ
	'\u01a4': "Phook",                // Ƥ
	'\u03a0': "Pi",                   // Π
	'\u0553': "Piwrarmenian",         // Փ
	'\uff30': "Pmonospace",           // Ｐ
	'\u03a8': "Psi",                  // Ψ
	'\u0470': "Psicyrillic",          // Ѱ
	'\uf770': "Psmall",
	'Q':      "Q",          // Q
	'\u24c6': "Qcircle",    // Ⓠ
	'\uff31': "Qmonospace", // Ｑ
	'\uf771': "Qsmall",
	'R':      "R",          // R
	'\u054c': "Raarmenian", // Ռ
	'\u0154': "Racute",     // Ŕ
	'\u0158': "Rcaron",     // Ř
	'\u0156': "Rcedilla",   // Ŗ
	'\u24c7': "Rcircle",    // Ⓡ
	// '\u0156':    "Rcommaaccent", // Ŗ -- duplicate
	'\u0210': "Rdblgrave",       // Ȑ
	'\u1e58': "Rdotaccent",      // Ṙ
	'\u1e5a': "Rdotbelow",       // Ṛ
	'\u1e5c': "Rdotbelowmacron", // Ṝ
	'\u0550': "Reharmenian",     // Ր
	'\u211c': "Rfraktur",        // ℜ
	'\u03a1': "Rho",             // Ρ
	'\uf6fc': "Ringsmall",
	'\u0212': "Rinvertedbreve", // Ȓ
	'\u1e5e': "Rlinebelow",     // Ṟ
	'\uff32': "Rmonospace",     // Ｒ
	'\uf772': "Rsmall",
	'\u0281': "Rsmallinverted",         // ʁ
	'\u02b6': "Rsmallinvertedsuperior", // ʶ
	'S':      "S",                      // S
	'\u250c': "SF010000",               // ┌
	'\u2514': "SF020000",               // └
	'\u2510': "SF030000",               // ┐
	'\u2518': "SF040000",               // ┘
	'\u253c': "SF050000",               // ┼
	'\u252c': "SF060000",               // ┬
	'\u2534': "SF070000",               // ┴
	'\u251c': "SF080000",               // ├
	'\u2524': "SF090000",               // ┤
	'\u2500': "SF100000",               // ─
	'\u2502': "SF110000",               // │
	'\u2561': "SF190000",               // ╡
	'\u2562': "SF200000",               // ╢
	'\u2556': "SF210000",               // ╖
	'\u2555': "SF220000",               // ╕
	'\u2563': "SF230000",               // ╣
	'\u2551': "SF240000",               // ║
	'\u2557': "SF250000",               // ╗
	'\u255d': "SF260000",               // ╝
	'\u255c': "SF270000",               // ╜
	'\u255b': "SF280000",               // ╛
	'\u255e': "SF360000",               // ╞
	'\u255f': "SF370000",               // ╟
	'\u255a': "SF380000",               // ╚
	'\u2554': "SF390000",               // ╔
	'\u2569': "SF400000",               // ╩
	'\u2566': "SF410000",               // ╦
	'\u2560': "SF420000",               // ╠
	'\u2550': "SF430000",               // ═
	'\u256c': "SF440000",               // ╬
	'\u2567': "SF450000",               // ╧
	'\u2568': "SF460000",               // ╨
	'\u2564': "SF470000",               // ╤
	'\u2565': "SF480000",               // ╥
	'\u2559': "SF490000",               // ╙
	'\u2558': "SF500000",               // ╘
	'\u2552': "SF510000",               // ╒
	'\u2553': "SF520000",               // ╓
	'\u256b': "SF530000",               // ╫
	'\u256a': "SF540000",               // ╪
	'\u015a': "Sacute",                 // Ś
	'\u1e64': "Sacutedotaccent",        // Ṥ
	'\u03e0': "Sampigreek",             // Ϡ
	'\u0160': "Scaron",                 // Š
	'\u1e66': "Scarondotaccent",        // Ṧ
	'\uf6fd': "Scaronsmall",
	'\u015e': "Scedilla",              // Ş
	'\u018f': "Schwa",                 // Ə
	'\u04d8': "Schwacyrillic",         // Ә
	'\u04da': "Schwadieresiscyrillic", // Ӛ
	'\u24c8': "Scircle",               // Ⓢ
	'\u015c': "Scircumflex",           // Ŝ
	'\u0218': "Scommaaccent",          // Ș
	'\u1e60': "Sdotaccent",            // Ṡ
	'\u1e62': "Sdotbelow",             // Ṣ
	'\u1e68': "Sdotbelowdotaccent",    // Ṩ
	'\u054d': "Seharmenian",           // Ս
	'\u2166': "Sevenroman",            // Ⅶ
	'\u0547': "Shaarmenian",           // Շ
	'\u0428': "Shacyrillic",           // Ш
	'\u0429': "Shchacyrillic",         // Щ
	'\u03e2': "Sheicoptic",            // Ϣ
	'\u04ba': "Shhacyrillic",          // Һ
	'\u03ec': "Shimacoptic",           // Ϭ
	'\u03a3': "Sigma",                 // Σ
	'\u2165': "Sixroman",              // Ⅵ
	'\uff33': "Smonospace",            // Ｓ
	'\u042c': "Softsigncyrillic",      // Ь
	'\uf773': "Ssmall",
	'\u03da': "Stigmagreek",      // Ϛ
	'T':      "T",                // T
	'\u03a4': "Tau",              // Τ
	'\u0166': "Tbar",             // Ŧ
	'\u0164': "Tcaron",           // Ť
	'\u0162': "Tcedilla",         // Ţ
	'\u24c9': "Tcircle",          // Ⓣ
	'\u1e70': "Tcircumflexbelow", // Ṱ
	// '\u0162':    "Tcommaaccent", // Ţ -- duplicate
	'\u1e6a': "Tdotaccent",          // Ṫ
	'\u1e6c': "Tdotbelow",           // Ṭ
	'\u0422': "Tecyrillic",          // Т
	'\u04ac': "Tedescendercyrillic", // Ҭ
	'\u2169': "Tenroman",            // Ⅹ
	'\u04b4': "Tetsecyrillic",       // Ҵ
	'\u0398': "Theta",               // Θ
	'\u01ac': "Thook",               // Ƭ
	'\u00de': "Thorn",               // Þ
	'\uf7fe': "Thornsmall",
	'\u2162': "Threeroman", // Ⅲ
	'\uf6fe': "Tildesmall",
	'\u054f': "Tiwnarmenian",   // Տ
	'\u1e6e': "Tlinebelow",     // Ṯ
	'\uff34': "Tmonospace",     // Ｔ
	'\u0539': "Toarmenian",     // Թ
	'\u01bc': "Tonefive",       // Ƽ
	'\u0184': "Tonesix",        // Ƅ
	'\u01a7': "Tonetwo",        // Ƨ
	'\u01ae': "Tretroflexhook", // Ʈ
	'\u0426': "Tsecyrillic",    // Ц
	'\u040b': "Tshecyrillic",   // Ћ
	'\uf774': "Tsmall",
	'\u216b': "Twelveroman", // Ⅻ
	'\u2161': "Tworoman",    // Ⅱ
	'U':      "U",           // U
	'\u00da': "Uacute",      // Ú
	'\uf7fa': "Uacutesmall",
	'\u016c': "Ubreve",           // Ŭ
	'\u01d3': "Ucaron",           // Ǔ
	'\u24ca': "Ucircle",          // Ⓤ
	'\u00db': "Ucircumflex",      // Û
	'\u1e76': "Ucircumflexbelow", // Ṷ
	'\uf7fb': "Ucircumflexsmall",
	'\u0423': "Ucyrillic",         // У
	'\u0170': "Udblacute",         // Ű
	'\u0214': "Udblgrave",         // Ȕ
	'\u00dc': "Udieresis",         // Ü
	'\u01d7': "Udieresisacute",    // Ǘ
	'\u1e72': "Udieresisbelow",    // Ṳ
	'\u01d9': "Udieresiscaron",    // Ǚ
	'\u04f0': "Udieresiscyrillic", // Ӱ
	'\u01db': "Udieresisgrave",    // Ǜ
	'\u01d5': "Udieresismacron",   // Ǖ
	'\uf7fc': "Udieresissmall",
	'\u1ee4': "Udotbelow", // Ụ
	'\u00d9': "Ugrave",    // Ù
	'\uf7f9': "Ugravesmall",
	'\u1ee6': "Uhookabove",     // Ủ
	'\u01af': "Uhorn",          // Ư
	'\u1ee8': "Uhornacute",     // Ứ
	'\u1ef0': "Uhorndotbelow",  // Ự
	'\u1eea': "Uhorngrave",     // Ừ
	'\u1eec': "Uhornhookabove", // Ử
	'\u1eee': "Uhorntilde",     // Ữ
	// '\u0170':    "Uhungarumlaut", // Ű -- duplicate
	'\u04f2': "Uhungarumlautcyrillic",          // Ӳ
	'\u0216': "Uinvertedbreve",                 // Ȗ
	'\u0478': "Ukcyrillic",                     // Ѹ
	'\u016a': "Umacron",                        // Ū
	'\u04ee': "Umacroncyrillic",                // Ӯ
	'\u1e7a': "Umacrondieresis",                // Ṻ
	'\uff35': "Umonospace",                     // Ｕ
	'\u0172': "Uogonek",                        // Ų
	'\u03a5': "Upsilon",                        // Υ
	'\u03d2': "Upsilon1",                       // ϒ
	'\u03d3': "Upsilonacutehooksymbolgreek",    // ϓ
	'\u01b1': "Upsilonafrican",                 // Ʊ
	'\u03ab': "Upsilondieresis",                // Ϋ
	'\u03d4': "Upsilondieresishooksymbolgreek", // ϔ
	// '\u03d2':    "Upsilonhooksymbol", // ϒ -- duplicate
	'\u038e': "Upsilontonos",   // Ύ
	'\u016e': "Uring",          // Ů
	'\u040e': "Ushortcyrillic", // Ў
	'\uf775': "Usmall",
	'\u04ae': "Ustraightcyrillic",       // Ү
	'\u04b0': "Ustraightstrokecyrillic", // Ұ
	'\u0168': "Utilde",                  // Ũ
	'\u1e78': "Utildeacute",             // Ṹ
	'\u1e74': "Utildebelow",             // Ṵ
	'V':      "V",                       // V
	'\u24cb': "Vcircle",                 // Ⓥ
	'\u1e7e': "Vdotbelow",               // Ṿ
	'\u0412': "Vecyrillic",              // В
	'\u054e': "Vewarmenian",             // Վ
	'\u01b2': "Vhook",                   // Ʋ
	'\uff36': "Vmonospace",              // Ｖ
	'\u0548': "Voarmenian",              // Ո
	'\uf776': "Vsmall",
	'\u1e7c': "Vtilde",      // Ṽ
	'W':      "W",           // W
	'\u1e82': "Wacute",      // Ẃ
	'\u24cc': "Wcircle",     // Ⓦ
	'\u0174': "Wcircumflex", // Ŵ
	'\u1e84': "Wdieresis",   // Ẅ
	'\u1e86': "Wdotaccent",  // Ẇ
	'\u1e88': "Wdotbelow",   // Ẉ
	'\u1e80': "Wgrave",      // Ẁ
	'\uff37': "Wmonospace",  // Ｗ
	'\uf777': "Wsmall",
	'X':      "X",           // X
	'\u24cd': "Xcircle",     // Ⓧ
	'\u1e8c': "Xdieresis",   // Ẍ
	'\u1e8a': "Xdotaccent",  // Ẋ
	'\u053d': "Xeharmenian", // Խ
	'\u039e': "Xi",          // Ξ
	'\uff38': "Xmonospace",  // Ｘ
	'\uf778': "Xsmall",
	'Y':      "Y",      // Y
	'\u00dd': "Yacute", // Ý
	'\uf7fd': "Yacutesmall",
	'\u0462': "Yatcyrillic", // Ѣ
	'\u24ce': "Ycircle",     // Ⓨ
	'\u0176': "Ycircumflex", // Ŷ
	'\u0178': "Ydieresis",   // Ÿ
	'\uf7ff': "Ydieresissmall",
	'\u1e8e': "Ydotaccent",           // Ẏ
	'\u1ef4': "Ydotbelow",            // Ỵ
	'\u042b': "Yericyrillic",         // Ы
	'\u04f8': "Yerudieresiscyrillic", // Ӹ
	'\u1ef2': "Ygrave",               // Ỳ
	'\u01b3': "Yhook",                // Ƴ
	'\u1ef6': "Yhookabove",           // Ỷ
	'\u0545': "Yiarmenian",           // Յ
	'\u0407': "Yicyrillic",           // Ї
	'\u0552': "Yiwnarmenian",         // Ւ
	'\uff39': "Ymonospace",           // Ｙ
	'\uf779': "Ysmall",
	'\u1ef8': "Ytilde",                    // Ỹ
	'\u046a': "Yusbigcyrillic",            // Ѫ
	'\u046c': "Yusbigiotifiedcyrillic",    // Ѭ
	'\u0466': "Yuslittlecyrillic",         // Ѧ
	'\u0468': "Yuslittleiotifiedcyrillic", // Ѩ
	'Z':      "Z",                         // Z
	'\u0536': "Zaarmenian",                // Զ
	'\u0179': "Zacute",                    // Ź
	'\u017d': "Zcaron",                    // Ž
	'\uf6ff': "Zcaronsmall",
	'\u24cf': "Zcircle",     // Ⓩ
	'\u1e90': "Zcircumflex", // Ẑ
	'\u017b': "Zdot",        // Ż
	// '\u017b':    "Zdotaccent", // Ż -- duplicate
	'\u1e92': "Zdotbelow",            // Ẓ
	'\u0417': "Zecyrillic",           // З
	'\u0498': "Zedescendercyrillic",  // Ҙ
	'\u04de': "Zedieresiscyrillic",   // Ӟ
	'\u0396': "Zeta",                 // Ζ
	'\u053a': "Zhearmenian",          // Ժ
	'\u04c1': "Zhebrevecyrillic",     // Ӂ
	'\u0416': "Zhecyrillic",          // Ж
	'\u0496': "Zhedescendercyrillic", // Җ
	'\u04dc': "Zhedieresiscyrillic",  // Ӝ
	'\u1e94': "Zlinebelow",           // Ẕ
	'\uff3a': "Zmonospace",           // Ｚ
	'\uf77a': "Zsmall",
	'\u01b5': "Zstroke",                  // Ƶ
	'a':      "a",                        // a
	'\u0986': "aabengali",                // আ
	'\u00e1': "aacute",                   // á
	'\u0906': "aadeva",                   // आ
	'\u0a86': "aagujarati",               // આ
	'\u0a06': "aagurmukhi",               // ਆ
	'\u0a3e': "aamatragurmukhi",          // ਾ
	'\u3303': "aarusquare",               // ㌃
	'\u09be': "aavowelsignbengali",       // া
	'\u093e': "aavowelsigndeva",          // ा
	'\u0abe': "aavowelsigngujarati",      // ા
	'\u055f': "abbreviationmarkarmenian", // ՟
	'\u0970': "abbreviationsigndeva",     // ॰
	'\u0985': "abengali",                 // অ
	'\u311a': "abopomofo",                // ㄚ
	'\u0103': "abreve",                   // ă
	'\u1eaf': "abreveacute",              // ắ
	'\u04d1': "abrevecyrillic",           // ӑ
	'\u1eb7': "abrevedotbelow",           // ặ
	'\u1eb1': "abrevegrave",              // ằ
	'\u1eb3': "abrevehookabove",          // ẳ
	'\u1eb5': "abrevetilde",              // ẵ
	'\u01ce': "acaron",                   // ǎ
	'\u24d0': "acircle",                  // ⓐ
	'\u00e2': "acircumflex",              // â
	'\u1ea5': "acircumflexacute",         // ấ
	'\u1ead': "acircumflexdotbelow",      // ậ
	'\u1ea7': "acircumflexgrave",         // ầ
	'\u1ea9': "acircumflexhookabove",     // ẩ
	'\u1eab': "acircumflextilde",         // ẫ
	'\u00b4': "acute",                    // ´
	'\u0317': "acutebelowcmb",            // ̗
	'\u0301': "acutecmb",                 // ́
	// '\u0301':    "acutecomb", // ́ -- duplicate
	'\u0954': "acutedeva",         // ॔
	'\u02cf': "acutelowmod",       // ˏ
	'\u0341': "acutetonecmb",      // ́
	'\u0430': "acyrillic",         // а
	'\u0201': "adblgrave",         // ȁ
	'\u0a71': "addakgurmukhi",     // ੱ
	'\u0905': "adeva",             // अ
	'\u00e4': "adieresis",         // ä
	'\u04d3': "adieresiscyrillic", // ӓ
	'\u01df': "adieresismacron",   // ǟ
	'\u1ea1': "adotbelow",         // ạ
	'\u01e1': "adotmacron",        // ǡ
	'\u00e6': "ae",                // æ
	'\u01fd': "aeacute",           // ǽ
	'\u3150': "aekorean",          // ㅐ
	'\u01e3': "aemacron",          // ǣ
	'\u2015': "afii00208",         // ―
	'\u20a4': "afii08941",         // ₤
	// '\u0410':    "afii10017", // А -- duplicate
	// '\u0411':    "afii10018", // Б -- duplicate
	// '\u0412':    "afii10019", // В -- duplicate
	// '\u0413':    "afii10020", // Г -- duplicate
	// '\u0414':    "afii10021", // Д -- duplicate
	// '\u0415':    "afii10022", // Е -- duplicate
	// '\u0401':    "afii10023", // Ё -- duplicate
	// '\u0416':    "afii10024", // Ж -- duplicate
	// '\u0417':    "afii10025", // З -- duplicate
	// '\u0418':    "afii10026", // И -- duplicate
	// '\u0419':    "afii10027", // Й -- duplicate
	// '\u041a':    "afii10028", // К -- duplicate
	// '\u041b':    "afii10029", // Л -- duplicate
	// '\u041c':    "afii10030", // М -- duplicate
	// '\u041d':    "afii10031", // Н -- duplicate
	// '\u041e':    "afii10032", // О -- duplicate
	// '\u041f':    "afii10033", // П -- duplicate
	// '\u0420':    "afii10034", // Р -- duplicate
	// '\u0421':    "afii10035", // С -- duplicate
	// '\u0422':    "afii10036", // Т -- duplicate
	// '\u0423':    "afii10037", // У -- duplicate
	// '\u0424':    "afii10038", // Ф -- duplicate
	// '\u0425':    "afii10039", // Х -- duplicate
	// '\u0426':    "afii10040", // Ц -- duplicate
	// '\u0427':    "afii10041", // Ч -- duplicate
	// '\u0428':    "afii10042", // Ш -- duplicate
	// '\u0429':    "afii10043", // Щ -- duplicate
	// '\u042a':    "afii10044", // Ъ -- duplicate
	// '\u042b':    "afii10045", // Ы -- duplicate
	// '\u042c':    "afii10046", // Ь -- duplicate
	// '\u042d':    "afii10047", // Э -- duplicate
	// '\u042e':    "afii10048", // Ю -- duplicate
	// '\u042f':    "afii10049", // Я -- duplicate
	// '\u0490':    "afii10050", // Ґ -- duplicate
	// '\u0402':    "afii10051", // Ђ -- duplicate
	// '\u0403':    "afii10052", // Ѓ -- duplicate
	// '\u0404':    "afii10053", // Є -- duplicate
	// '\u0405':    "afii10054", // Ѕ -- duplicate
	// '\u0406':    "afii10055", // І -- duplicate
	// '\u0407':    "afii10056", // Ї -- duplicate
	// '\u0408':    "afii10057", // Ј -- duplicate
	// '\u0409':    "afii10058", // Љ -- duplicate
	// '\u040a':    "afii10059", // Њ -- duplicate
	// '\u040b':    "afii10060", // Ћ -- duplicate
	// '\u040c':    "afii10061", // Ќ -- duplicate
	// '\u040e':    "afii10062", // Ў -- duplicate
	'\uf6c4': "afii10063",
	'\uf6c5': "afii10064",
	// '\u0430':    "afii10065", // а -- duplicate
	'\u0431': "afii10066", // б
	'\u0432': "afii10067", // в
	'\u0433': "afii10068", // г
	'\u0434': "afii10069", // д
	'\u0435': "afii10070", // е
	'\u0451': "afii10071", // ё
	'\u0436': "afii10072", // ж
	'\u0437': "afii10073", // з
	'\u0438': "afii10074", // и
	'\u0439': "afii10075", // й
	'\u043a': "afii10076", // к
	'\u043b': "afii10077", // л
	'\u043c': "afii10078", // м
	'\u043d': "afii10079", // н
	'\u043e': "afii10080", // о
	'\u043f': "afii10081", // п
	'\u0440': "afii10082", // р
	'\u0441': "afii10083", // с
	'\u0442': "afii10084", // т
	'\u0443': "afii10085", // у
	'\u0444': "afii10086", // ф
	'\u0445': "afii10087", // х
	'\u0446': "afii10088", // ц
	'\u0447': "afii10089", // ч
	'\u0448': "afii10090", // ш
	'\u0449': "afii10091", // щ
	'\u044a': "afii10092", // ъ
	'\u044b': "afii10093", // ы
	'\u044c': "afii10094", // ь
	'\u044d': "afii10095", // э
	'\u044e': "afii10096", // ю
	'\u044f': "afii10097", // я
	'\u0491': "afii10098", // ґ
	'\u0452': "afii10099", // ђ
	'\u0453': "afii10100", // ѓ
	'\u0454': "afii10101", // є
	'\u0455': "afii10102", // ѕ
	'\u0456': "afii10103", // і
	'\u0457': "afii10104", // ї
	'\u0458': "afii10105", // ј
	'\u0459': "afii10106", // љ
	'\u045a': "afii10107", // њ
	'\u045b': "afii10108", // ћ
	'\u045c': "afii10109", // ќ
	'\u045e': "afii10110", // ў
	// '\u040f':    "afii10145", // Џ -- duplicate
	// '\u0462':    "afii10146", // Ѣ -- duplicate
	// '\u0472':    "afii10147", // Ѳ -- duplicate
	// '\u0474':    "afii10148", // Ѵ -- duplicate
	'\uf6c6': "afii10192",
	'\u045f': "afii10193", // џ
	'\u0463': "afii10194", // ѣ
	'\u0473': "afii10195", // ѳ
	'\u0475': "afii10196", // ѵ
	'\uf6c7': "afii10831",
	'\uf6c8': "afii10832",
	'\u04d9': "afii10846", // ә
	'\u200e': "afii299",
	'\u200f': "afii300",
	'\u200d': "afii301",
	'\u066a': "afii57381", // ٪
	'\u060c': "afii57388", // ،
	'\u0660': "afii57392", // ٠
	'\u0661': "afii57393", // ١
	'\u0662': "afii57394", // ٢
	'\u0663': "afii57395", // ٣
	'\u0664': "afii57396", // ٤
	'\u0665': "afii57397", // ٥
	'\u0666': "afii57398", // ٦
	'\u0667': "afii57399", // ٧
	'\u0668': "afii57400", // ٨
	'\u0669': "afii57401", // ٩
	'\u061b': "afii57403", // ؛
	'\u061f': "afii57407", // ؟
	'\u0621': "afii57409", // ء
	'\u0622': "afii57410", // آ
	'\u0623': "afii57411", // أ
	'\u0624': "afii57412", // ؤ
	'\u0625': "afii57413", // إ
	'\u0626': "afii57414", // ئ
	'\u0627': "afii57415", // ا
	'\u0628': "afii57416", // ب
	'\u0629': "afii57417", // ة
	'\u062a': "afii57418", // ت
	'\u062b': "afii57419", // ث
	'\u062c': "afii57420", // ج
	'\u062d': "afii57421", // ح
	'\u062e': "afii57422", // خ
	'\u062f': "afii57423", // د
	'\u0630': "afii57424", // ذ
	'\u0631': "afii57425", // ر
	'\u0632': "afii57426", // ز
	'\u0633': "afii57427", // س
	'\u0634': "afii57428", // ش
	'\u0635': "afii57429", // ص
	'\u0636': "afii57430", // ض
	'\u0637': "afii57431", // ط
	'\u0638': "afii57432", // ظ
	'\u0639': "afii57433", // ع
	'\u063a': "afii57434", // غ
	'\u0640': "afii57440", // ـ
	'\u0641': "afii57441", // ف
	'\u0642': "afii57442", // ق
	'\u0643': "afii57443", // ك
	'\u0644': "afii57444", // ل
	'\u0645': "afii57445", // م
	'\u0646': "afii57446", // ن
	'\u0648': "afii57448", // و
	'\u0649': "afii57449", // ى
	'\u064a': "afii57450", // ي
	'\u064b': "afii57451", // ً
	'\u064c': "afii57452", // ٌ
	'\u064d': "afii57453", // ٍ
	'\u064e': "afii57454", // َ
	'\u064f': "afii57455", // ُ
	'\u0650': "afii57456", // ِ
	'\u0651': "afii57457", // ّ
	'\u0652': "afii57458", // ْ
	'\u0647': "afii57470", // ه
	'\u06a4': "afii57505", // ڤ
	'\u067e': "afii57506", // پ
	'\u0686': "afii57507", // چ
	'\u0698': "afii57508", // ژ
	'\u06af': "afii57509", // گ
	'\u0679': "afii57511", // ٹ
	'\u0688': "afii57512", // ڈ
	'\u0691': "afii57513", // ڑ
	'\u06ba': "afii57514", // ں
	'\u06d2': "afii57519", // ے
	'\u06d5': "afii57534", // ە
	'\u20aa': "afii57636", // ₪
	'\u05be': "afii57645", // ־
	'\u05c3': "afii57658", // ׃
	'\u05d0': "afii57664", // א
	'\u05d1': "afii57665", // ב
	'\u05d2': "afii57666", // ג
	'\u05d3': "afii57667", // ד
	'\u05d4': "afii57668", // ה
	'\u05d5': "afii57669", // ו
	'\u05d6': "afii57670", // ז
	'\u05d7': "afii57671", // ח
	'\u05d8': "afii57672", // ט
	'\u05d9': "afii57673", // י
	'\u05da': "afii57674", // ך
	'\u05db': "afii57675", // כ
	'\u05dc': "afii57676", // ל
	'\u05dd': "afii57677", // ם
	'\u05de': "afii57678", // מ
	'\u05df': "afii57679", // ן
	'\u05e0': "afii57680", // נ
	'\u05e1': "afii57681", // ס
	'\u05e2': "afii57682", // ע
	'\u05e3': "afii57683", // ף
	'\u05e4': "afii57684", // פ
	'\u05e5': "afii57685", // ץ
	'\u05e6': "afii57686", // צ
	'\u05e7': "afii57687", // ק
	'\u05e8': "afii57688", // ר
	'\u05e9': "afii57689", // ש
	'\u05ea': "afii57690", // ת
	'\ufb2a': "afii57694", // שׁ
	'\ufb2b': "afii57695", // שׂ
	'\ufb4b': "afii57700", // וֹ
	'\ufb1f': "afii57705", // ײַ
	'\u05f0': "afii57716", // װ
	'\u05f1': "afii57717", // ױ
	'\u05f2': "afii57718", // ײ
	'\ufb35': "afii57723", // וּ
	'\u05b4': "afii57793", // ִ
	'\u05b5': "afii57794", // ֵ
	'\u05b6': "afii57795", // ֶ
	'\u05bb': "afii57796", // ֻ
	'\u05b8': "afii57797", // ָ
	'\u05b7': "afii57798", // ַ
	'\u05b0': "afii57799", // ְ
	'\u05b2': "afii57800", // ֲ
	'\u05b1': "afii57801", // ֱ
	'\u05b3': "afii57802", // ֳ
	'\u05c2': "afii57803", // ׂ
	'\u05c1': "afii57804", // ׁ
	'\u05b9': "afii57806", // ֹ
	'\u05bc': "afii57807", // ּ
	'\u05bd': "afii57839", // ֽ
	'\u05bf': "afii57841", // ֿ
	'\u05c0': "afii57842", // ׀
	'\u02bc': "afii57929", // ʼ
	'\u2105': "afii61248", // ℅
	'\u2113': "afii61289", // ℓ
	'\u2116': "afii61352", // №
	'\u202c': "afii61573",
	'\u202d': "afii61574",
	'\u202e': "afii61575",
	'\u200c': "afii61664",
	'\u066d': "afii63167",       // ٭
	'\u02bd': "afii64937",       // ʽ
	'\u00e0': "agrave",          // à
	'\u0a85': "agujarati",       // અ
	'\u0a05': "agurmukhi",       // ਅ
	'\u3042': "ahiragana",       // あ
	'\u1ea3': "ahookabove",      // ả
	'\u0990': "aibengali",       // ঐ
	'\u311e': "aibopomofo",      // ㄞ
	'\u0910': "aideva",          // ऐ
	'\u04d5': "aiecyrillic",     // ӕ
	'\u0a90': "aigujarati",      // ઐ
	'\u0a10': "aigurmukhi",      // ਐ
	'\u0a48': "aimatragurmukhi", // ੈ
	// '\u0639':    "ainarabic", // ع -- duplicate
	'\ufeca': "ainfinalarabic",      // ﻊ
	'\ufecb': "aininitialarabic",    // ﻋ
	'\ufecc': "ainmedialarabic",     // ﻌ
	'\u0203': "ainvertedbreve",      // ȃ
	'\u09c8': "aivowelsignbengali",  // ৈ
	'\u0948': "aivowelsigndeva",     // ै
	'\u0ac8': "aivowelsigngujarati", // ૈ
	'\u30a2': "akatakana",           // ア
	'\uff71': "akatakanahalfwidth",  // ｱ
	'\u314f': "akorean",             // ㅏ
	// '\u05d0':    "alef", // א -- duplicate
	// '\u0627':    "alefarabic", // ا -- duplicate
	'\ufb30': "alefdageshhebrew", // אּ
	'\ufe8e': "aleffinalarabic",  // ﺎ
	// '\u0623':    "alefhamzaabovearabic", // أ -- duplicate
	'\ufe84': "alefhamzaabovefinalarabic", // ﺄ
	// '\u0625':    "alefhamzabelowarabic", // إ -- duplicate
	'\ufe88': "alefhamzabelowfinalarabic", // ﺈ
	// '\u05d0':    "alefhebrew", // א -- duplicate
	'\ufb4f': "aleflamedhebrew", // ﭏ
	// '\u0622':    "alefmaddaabovearabic", // آ -- duplicate
	'\ufe82': "alefmaddaabovefinalarabic", // ﺂ
	// '\u0649':    "alefmaksuraarabic", // ى -- duplicate
	'\ufef0': "alefmaksurafinalarabic",   // ﻰ
	'\ufef3': "alefmaksurainitialarabic", // ﻳ
	'\ufef4': "alefmaksuramedialarabic",  // ﻴ
	'\ufb2e': "alefpatahhebrew",          // אַ
	'\ufb2f': "alefqamatshebrew",         // אָ
	'\u2135': "aleph",                    // ℵ
	'\u224c': "allequal",                 // ≌
	'\u03b1': "alpha",                    // α
	'\u03ac': "alphatonos",               // ά
	'\u0101': "amacron",                  // ā
	'\uff41': "amonospace",               // ａ
	'&':      "ampersand",                // &
	'\uff06': "ampersandmonospace",       // ＆
	'\uf726': "ampersandsmall",
	'\u33c2': "amsquare",                  // ㏂
	'\u3122': "anbopomofo",                // ㄢ
	'\u3124': "angbopomofo",               // ㄤ
	'\u0e5a': "angkhankhuthai",            // ๚
	'\u2220': "angle",                     // ∠
	'\u3008': "anglebracketleft",          // 〈
	'\ufe3f': "anglebracketleftvertical",  // ︿
	'\u3009': "anglebracketright",         // 〉
	'\ufe40': "anglebracketrightvertical", // ﹀
	'\u2329': "angleleft",                 // 〈
	'\u232a': "angleright",                // 〉
	'\u212b': "angstrom",                  // Å
	'\u0387': "anoteleia",                 // ·
	'\u0952': "anudattadeva",              // ॒
	'\u0982': "anusvarabengali",           // ং
	'\u0902': "anusvaradeva",              // ं
	'\u0a82': "anusvaragujarati",          // ં
	'\u0105': "aogonek",                   // ą
	'\u3300': "apaatosquare",              // ㌀
	'\u249c': "aparen",                    // ⒜
	'\u055a': "apostrophearmenian",        // ՚
	// '\u02bc':    "apostrophemod", // ʼ -- duplicate
	'\uf8ff': "apple",
	'\u2250': "approaches",         // ≐
	'\u2248': "approxequal",        // ≈
	'\u2252': "approxequalorimage", // ≒
	'\u2245': "approximatelyequal", // ≅
	'\u318e': "araeaekorean",       // ㆎ
	'\u318d': "araeakorean",        // ㆍ
	'\u2312': "arc",                // ⌒
	'\u1e9a': "arighthalfring",     // ẚ
	'\u00e5': "aring",              // å
	'\u01fb': "aringacute",         // ǻ
	'\u1e01': "aringbelow",         // ḁ
	'\u2194': "arrowboth",          // ↔
	'\u21e3': "arrowdashdown",      // ⇣
	'\u21e0': "arrowdashleft",      // ⇠
	'\u21e2': "arrowdashright",     // ⇢
	'\u21e1': "arrowdashup",        // ⇡
	'\u21d4': "arrowdblboth",       // ⇔
	'\u21d3': "arrowdbldown",       // ⇓
	'\u21d0': "arrowdblleft",       // ⇐
	'\u21d2': "arrowdblright",      // ⇒
	'\u21d1': "arrowdblup",         // ⇑
	'\u2193': "arrowdown",          // ↓
	'\u2199': "arrowdownleft",      // ↙
	'\u2198': "arrowdownright",     // ↘
	'\u21e9': "arrowdownwhite",     // ⇩
	'\u02c5': "arrowheaddownmod",   // ˅
	'\u02c2': "arrowheadleftmod",   // ˂
	'\u02c3': "arrowheadrightmod",  // ˃
	'\u02c4': "arrowheadupmod",     // ˄
	'\uf8e7': "arrowhorizex",
	'\u2190': "arrowleft", // ←
	// '\u21d0':    "arrowleftdbl", // ⇐ -- duplicate
	'\u21cd': "arrowleftdblstroke",  // ⇍
	'\u21c6': "arrowleftoverright",  // ⇆
	'\u21e6': "arrowleftwhite",      // ⇦
	'\u2192': "arrowright",          // →
	'\u21cf': "arrowrightdblstroke", // ⇏
	'\u279e': "arrowrightheavy",     // ➞
	'\u21c4': "arrowrightoverleft",  // ⇄
	'\u21e8': "arrowrightwhite",     // ⇨
	'\u21e4': "arrowtableft",        // ⇤
	'\u21e5': "arrowtabright",       // ⇥
	'\u2191': "arrowup",             // ↑
	'\u2195': "arrowupdn",           // ↕
	'\u21a8': "arrowupdnbse",        // ↨
	// '\u21a8':    "arrowupdownbase", // ↨ -- duplicate
	'\u2196': "arrowupleft",       // ↖
	'\u21c5': "arrowupleftofdown", // ⇅
	'\u2197': "arrowupright",      // ↗
	'\u21e7': "arrowupwhite",      // ⇧
	'\uf8e6': "arrowvertex",
	'^':      "asciicircum",             // ^
	'\uff3e': "asciicircummonospace",    // ＾
	'~':      "asciitilde",              // ~
	'\uff5e': "asciitildemonospace",     // ～
	'\u0251': "ascript",                 // ɑ
	'\u0252': "ascriptturned",           // ɒ
	'\u3041': "asmallhiragana",          // ぁ
	'\u30a1': "asmallkatakana",          // ァ
	'\uff67': "asmallkatakanahalfwidth", // ｧ
	'*':      "asterisk",                // *
	// '\u066d':    "asteriskaltonearabic", // ٭ -- duplicate
	// '\u066d':    "asteriskarabic", // ٭ -- duplicate
	'\u2217': "asteriskmath",      // ∗
	'\uff0a': "asteriskmonospace", // ＊
	'\ufe61': "asterisksmall",     // ﹡
	'\u2042': "asterism",          // ⁂
	'\uf6e9': "asuperior",
	'\u2243': "asymptoticallyequal", // ≃
	'@':      "at",                  // @
	'\u00e3': "atilde",              // ã
	'\uff20': "atmonospace",         // ＠
	'\ufe6b': "atsmall",             // ﹫
	'\u0250': "aturned",             // ɐ
	'\u0994': "aubengali",           // ঔ
	'\u3120': "aubopomofo",          // ㄠ
	'\u0914': "audeva",              // औ
	'\u0a94': "augujarati",          // ઔ
	'\u0a14': "augurmukhi",          // ਔ
	'\u09d7': "aulengthmarkbengali", // ৗ
	'\u0a4c': "aumatragurmukhi",     // ੌ
	'\u09cc': "auvowelsignbengali",  // ৌ
	'\u094c': "auvowelsigndeva",     // ौ
	'\u0acc': "auvowelsigngujarati", // ૌ
	'\u093d': "avagrahadeva",        // ऽ
	'\u0561': "aybarmenian",         // ա
	// '\u05e2':    "ayin", // ע -- duplicate
	'\ufb20': "ayinaltonehebrew", // ﬠ
	// '\u05e2':    "ayinhebrew", // ע -- duplicate
	'b':      "b",                    // b
	'\u09ac': "babengali",            // ব
	'\\':     "backslash",            // \\
	'\uff3c': "backslashmonospace",   // ＼
	'\u092c': "badeva",               // ब
	'\u0aac': "bagujarati",           // બ
	'\u0a2c': "bagurmukhi",           // ਬ
	'\u3070': "bahiragana",           // ば
	'\u0e3f': "bahtthai",             // ฿
	'\u30d0': "bakatakana",           // バ
	'|':      "bar",                  // |
	'\uff5c': "barmonospace",         // ｜
	'\u3105': "bbopomofo",            // ㄅ
	'\u24d1': "bcircle",              // ⓑ
	'\u1e03': "bdotaccent",           // ḃ
	'\u1e05': "bdotbelow",            // ḅ
	'\u266c': "beamedsixteenthnotes", // ♬
	'\u2235': "because",              // ∵
	// '\u0431':    "becyrillic", // б -- duplicate
	// '\u0628':    "beharabic", // ب -- duplicate
	'\ufe90': "behfinalarabic",        // ﺐ
	'\ufe91': "behinitialarabic",      // ﺑ
	'\u3079': "behiragana",            // べ
	'\ufe92': "behmedialarabic",       // ﺒ
	'\ufc9f': "behmeeminitialarabic",  // ﲟ
	'\ufc08': "behmeemisolatedarabic", // ﰈ
	'\ufc6d': "behnoonfinalarabic",    // ﱭ
	'\u30d9': "bekatakana",            // ベ
	'\u0562': "benarmenian",           // բ
	// '\u05d1':    "bet", // ב -- duplicate
	'\u03b2': "beta",            // β
	'\u03d0': "betasymbolgreek", // ϐ
	'\ufb31': "betdagesh",       // בּ
	// '\ufb31':    "betdageshhebrew", // בּ -- duplicate
	// '\u05d1':    "bethebrew", // ב -- duplicate
	'\ufb4c': "betrafehebrew", // בֿ
	'\u09ad': "bhabengali",    // ভ
	'\u092d': "bhadeva",       // भ
	'\u0aad': "bhagujarati",   // ભ
	'\u0a2d': "bhagurmukhi",   // ਭ
	'\u0253': "bhook",         // ɓ
	'\u3073': "bihiragana",    // び
	'\u30d3': "bikatakana",    // ビ
	'\u0298': "bilabialclick", // ʘ
	'\u0a02': "bindigurmukhi", // ਂ
	'\u3331': "birusquare",    // ㌱
	// '\u25cf':    "blackcircle", // ● -- duplicate
	'\u25c6': "blackdiamond",                        // ◆
	'\u25bc': "blackdownpointingtriangle",           // ▼
	'\u25c4': "blackleftpointingpointer",            // ◄
	'\u25c0': "blackleftpointingtriangle",           // ◀
	'\u3010': "blacklenticularbracketleft",          // 【
	'\ufe3b': "blacklenticularbracketleftvertical",  // ︻
	'\u3011': "blacklenticularbracketright",         // 】
	'\ufe3c': "blacklenticularbracketrightvertical", // ︼
	'\u25e3': "blacklowerlefttriangle",              // ◣
	'\u25e2': "blacklowerrighttriangle",             // ◢
	'\u25ac': "blackrectangle",                      // ▬
	'\u25ba': "blackrightpointingpointer",           // ►
	'\u25b6': "blackrightpointingtriangle",          // ▶
	// '\u25aa':    "blacksmallsquare", // ▪ -- duplicate
	'\u263b': "blacksmilingface",             // ☻
	'\u25a0': "blacksquare",                  // ■
	'\u2605': "blackstar",                    // ★
	'\u25e4': "blackupperlefttriangle",       // ◤
	'\u25e5': "blackupperrighttriangle",      // ◥
	'\u25b4': "blackuppointingsmalltriangle", // ▴
	'\u25b2': "blackuppointingtriangle",      // ▲
	'\u2423': "blank",                        // ␣
	'\u1e07': "blinebelow",                   // ḇ
	'\u2588': "block",                        // █
	'\uff42': "bmonospace",                   // ｂ
	'\u0e1a': "bobaimaithai",                 // บ
	'\u307c': "bohiragana",                   // ぼ
	'\u30dc': "bokatakana",                   // ボ
	'\u249d': "bparen",                       // ⒝
	'\u33c3': "bqsquare",                     // ㏃
	'\uf8f4': "braceex",
	'{':      "braceleft", // {
	'\uf8f3': "braceleftbt",
	'\uf8f2': "braceleftmid",
	'\uff5b': "braceleftmonospace", // ｛
	'\ufe5b': "braceleftsmall",     // ﹛
	'\uf8f1': "bracelefttp",
	'\ufe37': "braceleftvertical", // ︷
	'}':      "braceright",        // }
	'\uf8fe': "bracerightbt",
	'\uf8fd': "bracerightmid",
	'\uff5d': "bracerightmonospace", // ｝
	'\ufe5c': "bracerightsmall",     // ﹜
	'\uf8fc': "bracerighttp",
	'\ufe38': "bracerightvertical", // ︸
	'[':      "bracketleft",        // [
	'\uf8f0': "bracketleftbt",
	'\uf8ef': "bracketleftex",
	'\uff3b': "bracketleftmonospace", // ［
	'\uf8ee': "bracketlefttp",
	']':      "bracketright", // ]
	'\uf8fb': "bracketrightbt",
	'\uf8fa': "bracketrightex",
	'\uff3d': "bracketrightmonospace", // ］
	'\uf8f9': "bracketrighttp",
	'\u02d8': "breve",                  // ˘
	'\u032e': "brevebelowcmb",          // ̮
	'\u0306': "brevecmb",               // ̆
	'\u032f': "breveinvertedbelowcmb",  // ̯
	'\u0311': "breveinvertedcmb",       // ̑
	'\u0361': "breveinverteddoublecmb", // ͡
	'\u032a': "bridgebelowcmb",         // ̪
	'\u033a': "bridgeinvertedbelowcmb", // ̺
	'\u00a6': "brokenbar",              // ¦
	'\u0180': "bstroke",                // ƀ
	'\uf6ea': "bsuperior",
	'\u0183': "btopbar",             // ƃ
	'\u3076': "buhiragana",          // ぶ
	'\u30d6': "bukatakana",          // ブ
	'\u2022': "bullet",              // •
	'\u25d8': "bulletinverse",       // ◘
	'\u2219': "bulletoperator",      // ∙
	'\u25ce': "bullseye",            // ◎
	'c':      "c",                   // c
	'\u056e': "caarmenian",          // ծ
	'\u099a': "cabengali",           // চ
	'\u0107': "cacute",              // ć
	'\u091a': "cadeva",              // च
	'\u0a9a': "cagujarati",          // ચ
	'\u0a1a': "cagurmukhi",          // ਚ
	'\u3388': "calsquare",           // ㎈
	'\u0981': "candrabindubengali",  // ঁ
	'\u0310': "candrabinducmb",      // ̐
	'\u0901': "candrabindudeva",     // ँ
	'\u0a81': "candrabindugujarati", // ઁ
	'\u21ea': "capslock",            // ⇪
	// '\u2105':    "careof", // ℅ -- duplicate
	'\u02c7': "caron",          // ˇ
	'\u032c': "caronbelowcmb",  // ̬
	'\u030c': "caroncmb",       // ̌
	'\u21b5': "carriagereturn", // ↵
	'\u3118': "cbopomofo",      // ㄘ
	'\u010d': "ccaron",         // č
	'\u00e7': "ccedilla",       // ç
	'\u1e09': "ccedillaacute",  // ḉ
	'\u24d2': "ccircle",        // ⓒ
	'\u0109': "ccircumflex",    // ĉ
	'\u0255': "ccurl",          // ɕ
	'\u010b': "cdot",           // ċ
	// '\u010b':    "cdotaccent", // ċ -- duplicate
	'\u33c5': "cdsquare",   // ㏅
	'\u00b8': "cedilla",    // ¸
	'\u0327': "cedillacmb", // ̧
	'\u00a2': "cent",       // ¢
	'\u2103': "centigrade", // ℃
	'\uf6df': "centinferior",
	'\uffe0': "centmonospace", // ￠
	'\uf7a2': "centoldstyle",
	'\uf6e0': "centsuperior",
	'\u0579': "chaarmenian",          // չ
	'\u099b': "chabengali",           // ছ
	'\u091b': "chadeva",              // छ
	'\u0a9b': "chagujarati",          // છ
	'\u0a1b': "chagurmukhi",          // ਛ
	'\u3114': "chbopomofo",           // ㄔ
	'\u04bd': "cheabkhasiancyrillic", // ҽ
	'\u2713': "checkmark",            // ✓
	// '\u0447':    "checyrillic", // ч -- duplicate
	'\u04bf': "chedescenderabkhasiancyrillic", // ҿ
	'\u04b7': "chedescendercyrillic",          // ҷ
	'\u04f5': "chedieresiscyrillic",           // ӵ
	'\u0573': "cheharmenian",                  // ճ
	'\u04cc': "chekhakassiancyrillic",         // ӌ
	'\u04b9': "cheverticalstrokecyrillic",     // ҹ
	'\u03c7': "chi",                           // χ
	'\u3277': "chieuchacirclekorean",          // ㉷
	'\u3217': "chieuchaparenkorean",           // ㈗
	'\u3269': "chieuchcirclekorean",           // ㉩
	'\u314a': "chieuchkorean",                 // ㅊ
	'\u3209': "chieuchparenkorean",            // ㈉
	'\u0e0a': "chochangthai",                  // ช
	'\u0e08': "chochanthai",                   // จ
	'\u0e09': "chochingthai",                  // ฉ
	'\u0e0c': "chochoethai",                   // ฌ
	'\u0188': "chook",                         // ƈ
	'\u3276': "cieucacirclekorean",            // ㉶
	'\u3216': "cieucaparenkorean",             // ㈖
	'\u3268': "cieuccirclekorean",             // ㉨
	'\u3148': "cieuckorean",                   // ㅈ
	'\u3208': "cieucparenkorean",              // ㈈
	'\u321c': "cieucuparenkorean",             // ㈜
	'\u25cb': "circle",                        // ○
	'\u2297': "circlemultiply",                // ⊗
	'\u2299': "circleot",                      // ⊙
	'\u2295': "circleplus",                    // ⊕
	'\u3036': "circlepostalmark",              // 〶
	'\u25d0': "circlewithlefthalfblack",       // ◐
	'\u25d1': "circlewithrighthalfblack",      // ◑
	'\u02c6': "circumflex",                    // ˆ
	'\u032d': "circumflexbelowcmb",            // ̭
	'\u0302': "circumflexcmb",                 // ̂
	'\u2327': "clear",                         // ⌧
	'\u01c2': "clickalveolar",                 // ǂ
	'\u01c0': "clickdental",                   // ǀ
	'\u01c1': "clicklateral",                  // ǁ
	'\u01c3': "clickretroflex",                // ǃ
	'\u2663': "club",                          // ♣
	// '\u2663':    "clubsuitblack", // ♣ -- duplicate
	'\u2667': "clubsuitwhite",   // ♧
	'\u33a4': "cmcubedsquare",   // ㎤
	'\uff43': "cmonospace",      // ｃ
	'\u33a0': "cmsquaredsquare", // ㎠
	'\u0581': "coarmenian",      // ց
	':':      "colon",           // :
	'\u20a1': "colonmonetary",   // ₡
	'\uff1a': "colonmonospace",  // ：
	// '\u20a1':    "colonsign", // ₡ -- duplicate
	'\ufe55': "colonsmall",             // ﹕
	'\u02d1': "colontriangularhalfmod", // ˑ
	'\u02d0': "colontriangularmod",     // ː
	',':      "comma",                  // ,
	'\u0313': "commaabovecmb",          // ̓
	'\u0315': "commaaboverightcmb",     // ̕
	'\uf6c3': "commaaccent",
	// '\u060c':    "commaarabic", // ، -- duplicate
	'\u055d': "commaarmenian", // ՝
	'\uf6e1': "commainferior",
	'\uff0c': "commamonospace",        // ，
	'\u0314': "commareversedabovecmb", // ̔
	// '\u02bd':    "commareversedmod", // ʽ -- duplicate
	'\ufe50': "commasmall", // ﹐
	'\uf6e2': "commasuperior",
	'\u0312': "commaturnedabovecmb", // ̒
	'\u02bb': "commaturnedmod",      // ʻ
	'\u263c': "compass",             // ☼
	// '\u2245':    "congruent", // ≅ -- duplicate
	'\u222e': "contourintegral", // ∮
	'\u2303': "control",         // ⌃
	'\x06':   "controlACK",
	'\a':     "controlBEL",
	'\b':     "controlBS",
	'\x18':   "controlCAN",
	'\r':     "controlCR",
	'\x11':   "controlDC1",
	'\x12':   "controlDC2",
	'\x13':   "controlDC3",
	'\x14':   "controlDC4",
	'\u007f': "controlDEL",
	'\x10':   "controlDLE",
	'\x19':   "controlEM",
	'\x05':   "controlENQ",
	'\x04':   "controlEOT",
	'\x1b':   "controlESC",
	'\x17':   "controlETB",
	'\x03':   "controlETX",
	'\f':     "controlFF",
	'\x1c':   "controlFS",
	'\x1d':   "controlGS",
	'\t':     "controlHT",
	'\n':     "controlLF",
	'\x15':   "controlNAK",
	'\x1e':   "controlRS",
	'\x0f':   "controlSI",
	'\x0e':   "controlSO",
	'\x02':   "controlSOT",
	'\x01':   "controlSTX",
	'\x1a':   "controlSUB",
	'\x16':   "controlSYN",
	'\x1f':   "controlUS",
	'\v':     "controlVT",
	'\u00a9': "copyright", // ©
	'\uf8e9': "copyrightsans",
	'\uf6d9': "copyrightserif",
	'\u300c': "cornerbracketleft",           // 「
	'\uff62': "cornerbracketlefthalfwidth",  // ｢
	'\ufe41': "cornerbracketleftvertical",   // ﹁
	'\u300d': "cornerbracketright",          // 」
	'\uff63': "cornerbracketrighthalfwidth", // ｣
	'\ufe42': "cornerbracketrightvertical",  // ﹂
	'\u337f': "corporationsquare",           // ㍿
	'\u33c7': "cosquare",                    // ㏇
	'\u33c6': "coverkgsquare",               // ㏆
	'\u249e': "cparen",                      // ⒞
	'\u20a2': "cruzeiro",                    // ₢
	'\u0297': "cstretched",                  // ʗ
	'\u22cf': "curlyand",                    // ⋏
	'\u22ce': "curlyor",                     // ⋎
	'\u00a4': "currency",                    // ¤
	'\uf6d1': "cyrBreve",
	'\uf6d2': "cyrFlex",
	'\uf6d4': "cyrbreve",
	'\uf6d5': "cyrflex",
	'd':      "d",          // d
	'\u0564': "daarmenian", // դ
	'\u09a6': "dabengali",  // দ
	// '\u0636':    "dadarabic", // ض -- duplicate
	'\u0926': "dadeva",           // द
	'\ufebe': "dadfinalarabic",   // ﺾ
	'\ufebf': "dadinitialarabic", // ﺿ
	'\ufec0': "dadmedialarabic",  // ﻀ
	// '\u05bc':    "dagesh", // ּ -- duplicate
	// '\u05bc':    "dageshhebrew", // ּ -- duplicate
	'\u2020': "dagger",     // †
	'\u2021': "daggerdbl",  // ‡
	'\u0aa6': "dagujarati", // દ
	'\u0a26': "dagurmukhi", // ਦ
	'\u3060': "dahiragana", // だ
	'\u30c0': "dakatakana", // ダ
	// '\u062f':    "dalarabic", // د -- duplicate
	// '\u05d3':    "dalet", // ד -- duplicate
	'\ufb33': "daletdagesh", // דּ
	// '\ufb33':    "daletdageshhebrew", // דּ -- duplicate
	// '\u05d3':    "dalethatafpatah", // ד -- duplicate
	// '\u05d3':    "dalethatafpatahhebrew", // ד -- duplicate
	// '\u05d3':    "dalethatafsegol", // ד -- duplicate
	// '\u05d3':    "dalethatafsegolhebrew", // ד -- duplicate
	// '\u05d3':    "dalethebrew", // ד -- duplicate
	// '\u05d3':    "dalethiriq", // ד -- duplicate
	// '\u05d3':    "dalethiriqhebrew", // ד -- duplicate
	// '\u05d3':    "daletholam", // ד -- duplicate
	// '\u05d3':    "daletholamhebrew", // ד -- duplicate
	// '\u05d3':    "daletpatah", // ד -- duplicate
	// '\u05d3':    "daletpatahhebrew", // ד -- duplicate
	// '\u05d3':    "daletqamats", // ד -- duplicate
	// '\u05d3':    "daletqamatshebrew", // ד -- duplicate
	// '\u05d3':    "daletqubuts", // ד -- duplicate
	// '\u05d3':    "daletqubutshebrew", // ד -- duplicate
	// '\u05d3':    "daletsegol", // ד -- duplicate
	// '\u05d3':    "daletsegolhebrew", // ד -- duplicate
	// '\u05d3':    "daletsheva", // ד -- duplicate
	// '\u05d3':    "daletshevahebrew", // ד -- duplicate
	// '\u05d3':    "dalettsere", // ד -- duplicate
	// '\u05d3':    "dalettserehebrew", // ד -- duplicate
	'\ufeaa': "dalfinalarabic", // ﺪ
	// '\u064f':    "dammaarabic", // ُ -- duplicate
	// '\u064f':    "dammalowarabic", // ُ -- duplicate
	// '\u064c':    "dammatanaltonearabic", // ٌ -- duplicate
	// '\u064c':    "dammatanarabic", // ٌ -- duplicate
	'\u0964': "danda",       // ।
	'\u05a7': "dargahebrew", // ֧
	// '\u05a7':    "dargalefthebrew", // ֧ -- duplicate
	'\u0485': "dasiapneumatacyrilliccmb", // ҅
	'\uf6d3': "dblGrave",
	'\u300a': "dblanglebracketleft",          // 《
	'\ufe3d': "dblanglebracketleftvertical",  // ︽
	'\u300b': "dblanglebracketright",         // 》
	'\ufe3e': "dblanglebracketrightvertical", // ︾
	'\u032b': "dblarchinvertedbelowcmb",      // ̫
	// '\u21d4':    "dblarrowleft", // ⇔ -- duplicate
	// '\u21d2':    "dblarrowright", // ⇒ -- duplicate
	'\u0965': "dbldanda", // ॥
	'\uf6d6': "dblgrave",
	'\u030f': "dblgravecmb",             // ̏
	'\u222c': "dblintegral",             // ∬
	'\u2017': "dbllowline",              // ‗
	'\u0333': "dbllowlinecmb",           // ̳
	'\u033f': "dbloverlinecmb",          // ̿
	'\u02ba': "dblprimemod",             // ʺ
	'\u2016': "dblverticalbar",          // ‖
	'\u030e': "dblverticallineabovecmb", // ̎
	'\u3109': "dbopomofo",               // ㄉ
	'\u33c8': "dbsquare",                // ㏈
	'\u010f': "dcaron",                  // ď
	'\u1e11': "dcedilla",                // ḑ
	'\u24d3': "dcircle",                 // ⓓ
	'\u1e13': "dcircumflexbelow",        // ḓ
	'\u0111': "dcroat",                  // đ
	'\u09a1': "ddabengali",              // ড
	'\u0921': "ddadeva",                 // ड
	'\u0aa1': "ddagujarati",             // ડ
	'\u0a21': "ddagurmukhi",             // ਡ
	// '\u0688':    "ddalarabic", // ڈ -- duplicate
	'\ufb89': "ddalfinalarabic",        // ﮉ
	'\u095c': "dddhadeva",              // ड़
	'\u09a2': "ddhabengali",            // ঢ
	'\u0922': "ddhadeva",               // ढ
	'\u0aa2': "ddhagujarati",           // ઢ
	'\u0a22': "ddhagurmukhi",           // ਢ
	'\u1e0b': "ddotaccent",             // ḋ
	'\u1e0d': "ddotbelow",              // ḍ
	'\u066b': "decimalseparatorarabic", // ٫
	// '\u066b':    "decimalseparatorpersian", // ٫ -- duplicate
	// '\u0434':    "decyrillic", // д -- duplicate
	'\u00b0': "degree",                              // °
	'\u05ad': "dehihebrew",                          // ֭
	'\u3067': "dehiragana",                          // で
	'\u03ef': "deicoptic",                           // ϯ
	'\u30c7': "dekatakana",                          // デ
	'\u232b': "deleteleft",                          // ⌫
	'\u2326': "deleteright",                         // ⌦
	'\u03b4': "delta",                               // δ
	'\u018d': "deltaturned",                         // ƍ
	'\u09f8': "denominatorminusonenumeratorbengali", // ৸
	'\u02a4': "dezh",                                // ʤ
	'\u09a7': "dhabengali",                          // ধ
	'\u0927': "dhadeva",                             // ध
	'\u0aa7': "dhagujarati",                         // ધ
	'\u0a27': "dhagurmukhi",                         // ਧ
	'\u0257': "dhook",                               // ɗ
	'\u0385': "dialytikatonos",                      // ΅
	'\u0344': "dialytikatonoscmb",                   // ̈́
	'\u2666': "diamond",                             // ♦
	'\u2662': "diamondsuitwhite",                    // ♢
	'\u00a8': "dieresis",                            // ¨
	'\uf6d7': "dieresisacute",
	'\u0324': "dieresisbelowcmb", // ̤
	'\u0308': "dieresiscmb",      // ̈
	'\uf6d8': "dieresisgrave",
	// '\u0385':    "dieresistonos", // ΅ -- duplicate
	'\u3062': "dihiragana",    // ぢ
	'\u30c2': "dikatakana",    // ヂ
	'\u3003': "dittomark",     // 〃
	'\u00f7': "divide",        // ÷
	'\u2223': "divides",       // ∣
	'\u2215': "divisionslash", // ∕
	// '\u0452':    "djecyrillic", // ђ -- duplicate
	'\u2593': "dkshade",    // ▓
	'\u1e0f': "dlinebelow", // ḏ
	'\u3397': "dlsquare",   // ㎗
	// '\u0111':    "dmacron", // đ -- duplicate
	'\uff44': "dmonospace",  // ｄ
	'\u2584': "dnblock",     // ▄
	'\u0e0e': "dochadathai", // ฎ
	'\u0e14': "dodekthai",   // ด
	'\u3069': "dohiragana",  // ど
	'\u30c9': "dokatakana",  // ド
	'$':      "dollar",      // $
	'\uf6e3': "dollarinferior",
	'\uff04': "dollarmonospace", // ＄
	'\uf724': "dollaroldstyle",
	'\ufe69': "dollarsmall", // ﹩
	'\uf6e4': "dollarsuperior",
	'\u20ab': "dong",         // ₫
	'\u3326': "dorusquare",   // ㌦
	'\u02d9': "dotaccent",    // ˙
	'\u0307': "dotaccentcmb", // ̇
	'\u0323': "dotbelowcmb",  // ̣
	// '\u0323':    "dotbelowcomb", // ̣ -- duplicate
	'\u30fb': "dotkatakana", // ・
	'\u0131': "dotlessi",    // ı
	'\uf6be': "dotlessj",
	'\u0284': "dotlessjstrokehook", // ʄ
	'\u22c5': "dotmath",            // ⋅
	'\u25cc': "dottedcircle",       // ◌
	// '\ufb1f':    "doubleyodpatah", // ײַ -- duplicate
	// '\ufb1f':    "doubleyodpatahhebrew", // ײַ -- duplicate
	'\u031e': "downtackbelowcmb", // ̞
	'\u02d5': "downtackmod",      // ˕
	'\u249f': "dparen",           // ⒟
	'\uf6eb': "dsuperior",
	'\u0256': "dtail",                // ɖ
	'\u018c': "dtopbar",              // ƌ
	'\u3065': "duhiragana",           // づ
	'\u30c5': "dukatakana",           // ヅ
	'\u01f3': "dz",                   // ǳ
	'\u02a3': "dzaltone",             // ʣ
	'\u01c6': "dzcaron",              // ǆ
	'\u02a5': "dzcurl",               // ʥ
	'\u04e1': "dzeabkhasiancyrillic", // ӡ
	// '\u0455':    "dzecyrillic", // ѕ -- duplicate
	// '\u045f':    "dzhecyrillic", // џ -- duplicate
	'e':      "e",                        // e
	'\u00e9': "eacute",                   // é
	'\u2641': "earth",                    // ♁
	'\u098f': "ebengali",                 // এ
	'\u311c': "ebopomofo",                // ㄜ
	'\u0115': "ebreve",                   // ĕ
	'\u090d': "ecandradeva",              // ऍ
	'\u0a8d': "ecandragujarati",          // ઍ
	'\u0945': "ecandravowelsigndeva",     // ॅ
	'\u0ac5': "ecandravowelsigngujarati", // ૅ
	'\u011b': "ecaron",                   // ě
	'\u1e1d': "ecedillabreve",            // ḝ
	'\u0565': "echarmenian",              // ե
	'\u0587': "echyiwnarmenian",          // և
	'\u24d4': "ecircle",                  // ⓔ
	'\u00ea': "ecircumflex",              // ê
	'\u1ebf': "ecircumflexacute",         // ế
	'\u1e19': "ecircumflexbelow",         // ḙ
	'\u1ec7': "ecircumflexdotbelow",      // ệ
	'\u1ec1': "ecircumflexgrave",         // ề
	'\u1ec3': "ecircumflexhookabove",     // ể
	'\u1ec5': "ecircumflextilde",         // ễ
	// '\u0454':    "ecyrillic", // є -- duplicate
	'\u0205': "edblgrave", // ȅ
	'\u090f': "edeva",     // ए
	'\u00eb': "edieresis", // ë
	'\u0117': "edot",      // ė
	// '\u0117':    "edotaccent", // ė -- duplicate
	'\u1eb9': "edotbelow",       // ẹ
	'\u0a0f': "eegurmukhi",      // ਏ
	'\u0a47': "eematragurmukhi", // ੇ
	// '\u0444':    "efcyrillic", // ф -- duplicate
	'\u00e8': "egrave",     // è
	'\u0a8f': "egujarati",  // એ
	'\u0567': "eharmenian", // է
	'\u311d': "ehbopomofo", // ㄝ
	'\u3048': "ehiragana",  // え
	'\u1ebb': "ehookabove", // ẻ
	'\u311f': "eibopomofo", // ㄟ
	'8':      "eight",      // 8
	// '\u0668':    "eightarabic", // ٨ -- duplicate
	'\u09ee': "eightbengali",                // ৮
	'\u2467': "eightcircle",                 // ⑧
	'\u2791': "eightcircleinversesansserif", // ➑
	'\u096e': "eightdeva",                   // ८
	'\u2471': "eighteencircle",              // ⑱
	'\u2485': "eighteenparen",               // ⒅
	'\u2499': "eighteenperiod",              // ⒙
	'\u0aee': "eightgujarati",               // ૮
	'\u0a6e': "eightgurmukhi",               // ੮
	// '\u0668':    "eighthackarabic", // ٨ -- duplicate
	'\u3028': "eighthangzhou",         // 〨
	'\u266b': "eighthnotebeamed",      // ♫
	'\u3227': "eightideographicparen", // ㈧
	'\u2088': "eightinferior",         // ₈
	'\uff18': "eightmonospace",        // ８
	'\uf738': "eightoldstyle",
	'\u247b': "eightparen",         // ⑻
	'\u248f': "eightperiod",        // ⒏
	'\u06f8': "eightpersian",       // ۸
	'\u2177': "eightroman",         // ⅷ
	'\u2078': "eightsuperior",      // ⁸
	'\u0e58': "eightthai",          // ๘
	'\u0207': "einvertedbreve",     // ȇ
	'\u0465': "eiotifiedcyrillic",  // ѥ
	'\u30a8': "ekatakana",          // エ
	'\uff74': "ekatakanahalfwidth", // ｴ
	'\u0a74': "ekonkargurmukhi",    // ੴ
	'\u3154': "ekorean",            // ㅔ
	// '\u043b':    "elcyrillic", // л -- duplicate
	'\u2208': "element",          // ∈
	'\u246a': "elevencircle",     // ⑪
	'\u247e': "elevenparen",      // ⑾
	'\u2492': "elevenperiod",     // ⒒
	'\u217a': "elevenroman",      // ⅺ
	'\u2026': "ellipsis",         // …
	'\u22ee': "ellipsisvertical", // ⋮
	'\u0113': "emacron",          // ē
	'\u1e17': "emacronacute",     // ḗ
	'\u1e15': "emacrongrave",     // ḕ
	// '\u043c':    "emcyrillic", // м -- duplicate
	'\u2014': "emdash",               // —
	'\ufe31': "emdashvertical",       // ︱
	'\uff45': "emonospace",           // ｅ
	'\u055b': "emphasismarkarmenian", // ՛
	'\u2205': "emptyset",             // ∅
	'\u3123': "enbopomofo",           // ㄣ
	// '\u043d':    "encyrillic", // н -- duplicate
	'\u2013': "endash",              // –
	'\ufe32': "endashvertical",      // ︲
	'\u04a3': "endescendercyrillic", // ң
	'\u014b': "eng",                 // ŋ
	'\u3125': "engbopomofo",         // ㄥ
	'\u04a5': "enghecyrillic",       // ҥ
	'\u04c8': "enhookcyrillic",      // ӈ
	'\u2002': "enspace",
	'\u0119': "eogonek",             // ę
	'\u3153': "eokorean",            // ㅓ
	'\u025b': "eopen",               // ɛ
	'\u029a': "eopenclosed",         // ʚ
	'\u025c': "eopenreversed",       // ɜ
	'\u025e': "eopenreversedclosed", // ɞ
	'\u025d': "eopenreversedhook",   // ɝ
	'\u24a0': "eparen",              // ⒠
	'\u03b5': "epsilon",             // ε
	'\u03ad': "epsilontonos",        // έ
	'=':      "equal",               // =
	'\uff1d': "equalmonospace",      // ＝
	'\ufe66': "equalsmall",          // ﹦
	'\u207c': "equalsuperior",       // ⁼
	'\u2261': "equivalence",         // ≡
	'\u3126': "erbopomofo",          // ㄦ
	// '\u0440':    "ercyrillic", // р -- duplicate
	'\u0258': "ereversed", // ɘ
	// '\u044d':    "ereversedcyrillic", // э -- duplicate
	// '\u0441':    "escyrillic", // с -- duplicate
	'\u04ab': "esdescendercyrillic",     // ҫ
	'\u0283': "esh",                     // ʃ
	'\u0286': "eshcurl",                 // ʆ
	'\u090e': "eshortdeva",              // ऎ
	'\u0946': "eshortvowelsigndeva",     // ॆ
	'\u01aa': "eshreversedloop",         // ƪ
	'\u0285': "eshsquatreversed",        // ʅ
	'\u3047': "esmallhiragana",          // ぇ
	'\u30a7': "esmallkatakana",          // ェ
	'\uff6a': "esmallkatakanahalfwidth", // ｪ
	'\u212e': "estimated",               // ℮
	'\uf6ec': "esuperior",
	'\u03b7': "eta",                // η
	'\u0568': "etarmenian",         // ը
	'\u03ae': "etatonos",           // ή
	'\u00f0': "eth",                // ð
	'\u1ebd': "etilde",             // ẽ
	'\u1e1b': "etildebelow",        // ḛ
	'\u0591': "etnahtafoukhhebrew", // ֑
	// '\u0591':    "etnahtafoukhlefthebrew", // ֑ -- duplicate
	// '\u0591':    "etnahtahebrew", // ֑ -- duplicate
	// '\u0591':    "etnahtalefthebrew", // ֑ -- duplicate
	'\u01dd': "eturned",  // ǝ
	'\u3161': "eukorean", // ㅡ
	// '\u20ac':    "euro", // € -- duplicate
	'\u09c7': "evowelsignbengali",  // ে
	'\u0947': "evowelsigndeva",     // े
	'\u0ac7': "evowelsigngujarati", // ે
	'!':      "exclam",             // !
	'\u055c': "exclamarmenian",     // ՜
	'\u203c': "exclamdbl",          // ‼
	'\u00a1': "exclamdown",         // ¡
	'\uf7a1': "exclamdownsmall",
	'\uff01': "exclammonospace", // ！
	'\uf721': "exclamsmall",
	'\u2203': "existential", // ∃
	'\u0292': "ezh",         // ʒ
	'\u01ef': "ezhcaron",    // ǯ
	'\u0293': "ezhcurl",     // ʓ
	'\u01b9': "ezhreversed", // ƹ
	'\u01ba': "ezhtail",     // ƺ
	'f':      "f",           // f
	'\u095e': "fadeva",      // फ़
	'\u0a5e': "fagurmukhi",  // ਫ਼
	'\u2109': "fahrenheit",  // ℉
	// '\u064e':    "fathaarabic", // َ -- duplicate
	// '\u064e':    "fathalowarabic", // َ -- duplicate
	// '\u064b':    "fathatanarabic", // ً -- duplicate
	'\u3108': "fbopomofo",  // ㄈ
	'\u24d5': "fcircle",    // ⓕ
	'\u1e1f': "fdotaccent", // ḟ
	// '\u0641':    "feharabic", // ف -- duplicate
	'\u0586': "feharmenian",      // ֆ
	'\ufed2': "fehfinalarabic",   // ﻒ
	'\ufed3': "fehinitialarabic", // ﻓ
	'\ufed4': "fehmedialarabic",  // ﻔ
	'\u03e5': "feicoptic",        // ϥ
	'\u2640': "female",           // ♀
	'\ufb00': "ff",               // ﬀ
	'\ufb03': "ffi",              // ﬃ
	'\ufb04': "ffl",              // ﬄ
	'\ufb01': "fi",               // ﬁ
	'\u246e': "fifteencircle",    // ⑮
	'\u2482': "fifteenparen",     // ⒂
	'\u2496': "fifteenperiod",    // ⒖
	'\u2012': "figuredash",       // ‒
	// '\u25a0':    "filledbox", // ■ -- duplicate
	// '\u25ac':    "filledrect", // ▬ -- duplicate
	// '\u05da':    "finalkaf", // ך -- duplicate
	'\ufb3a': "finalkafdagesh", // ךּ
	// '\ufb3a':    "finalkafdageshhebrew", // ךּ -- duplicate
	// '\u05da':    "finalkafhebrew", // ך -- duplicate
	// '\u05da':    "finalkafqamats", // ך -- duplicate
	// '\u05da':    "finalkafqamatshebrew", // ך -- duplicate
	// '\u05da':    "finalkafsheva", // ך -- duplicate
	// '\u05da':    "finalkafshevahebrew", // ך -- duplicate
	// '\u05dd':    "finalmem", // ם -- duplicate
	// '\u05dd':    "finalmemhebrew", // ם -- duplicate
	// '\u05df':    "finalnun", // ן -- duplicate
	// '\u05df':    "finalnunhebrew", // ן -- duplicate
	// '\u05e3':    "finalpe", // ף -- duplicate
	// '\u05e3':    "finalpehebrew", // ף -- duplicate
	// '\u05e5':    "finaltsadi", // ץ -- duplicate
	// '\u05e5':    "finaltsadihebrew", // ץ -- duplicate
	'\u02c9': "firsttonechinese", // ˉ
	'\u25c9': "fisheye",          // ◉
	// '\u0473':    "fitacyrillic", // ѳ -- duplicate
	'5': "five", // 5
	// '\u0665':    "fivearabic", // ٥ -- duplicate
	'\u09eb': "fivebengali",                // ৫
	'\u2464': "fivecircle",                 // ⑤
	'\u278e': "fivecircleinversesansserif", // ➎
	'\u096b': "fivedeva",                   // ५
	'\u215d': "fiveeighths",                // ⅝
	'\u0aeb': "fivegujarati",               // ૫
	'\u0a6b': "fivegurmukhi",               // ੫
	// '\u0665':    "fivehackarabic", // ٥ -- duplicate
	'\u3025': "fivehangzhou",         // 〥
	'\u3224': "fiveideographicparen", // ㈤
	'\u2085': "fiveinferior",         // ₅
	'\uff15': "fivemonospace",        // ５
	'\uf735': "fiveoldstyle",
	'\u2478': "fiveparen",    // ⑸
	'\u248c': "fiveperiod",   // ⒌
	'\u06f5': "fivepersian",  // ۵
	'\u2174': "fiveroman",    // ⅴ
	'\u2075': "fivesuperior", // ⁵
	'\u0e55': "fivethai",     // ๕
	'\ufb02': "fl",           // ﬂ
	'\u0192': "florin",       // ƒ
	'\uff46': "fmonospace",   // ｆ
	'\u3399': "fmsquare",     // ㎙
	'\u0e1f': "fofanthai",    // ฟ
	'\u0e1d': "fofathai",     // ฝ
	'\u0e4f': "fongmanthai",  // ๏
	'\u2200': "forall",       // ∀
	'4':      "four",         // 4
	// '\u0664':    "fourarabic", // ٤ -- duplicate
	'\u09ea': "fourbengali",                // ৪
	'\u2463': "fourcircle",                 // ④
	'\u278d': "fourcircleinversesansserif", // ➍
	'\u096a': "fourdeva",                   // ४
	'\u0aea': "fourgujarati",               // ૪
	'\u0a6a': "fourgurmukhi",               // ੪
	// '\u0664':    "fourhackarabic", // ٤ -- duplicate
	'\u3024': "fourhangzhou",         // 〤
	'\u3223': "fourideographicparen", // ㈣
	'\u2084': "fourinferior",         // ₄
	'\uff14': "fourmonospace",        // ４
	'\u09f7': "fournumeratorbengali", // ৷
	'\uf734': "fouroldstyle",
	'\u2477': "fourparen",         // ⑷
	'\u248b': "fourperiod",        // ⒋
	'\u06f4': "fourpersian",       // ۴
	'\u2173': "fourroman",         // ⅳ
	'\u2074': "foursuperior",      // ⁴
	'\u246d': "fourteencircle",    // ⑭
	'\u2481': "fourteenparen",     // ⒁
	'\u2495': "fourteenperiod",    // ⒕
	'\u0e54': "fourthai",          // ๔
	'\u02cb': "fourthtonechinese", // ˋ
	'\u24a1': "fparen",            // ⒡
	'\u2044': "fraction",          // ⁄
	'\u20a3': "franc",             // ₣
	'g':      "g",                 // g
	'\u0997': "gabengali",         // গ
	'\u01f5': "gacute",            // ǵ
	'\u0917': "gadeva",            // ग
	// '\u06af':    "gafarabic", // گ -- duplicate
	'\ufb93': "gaffinalarabic",   // ﮓ
	'\ufb94': "gafinitialarabic", // ﮔ
	'\ufb95': "gafmedialarabic",  // ﮕ
	'\u0a97': "gagujarati",       // ગ
	'\u0a17': "gagurmukhi",       // ਗ
	'\u304c': "gahiragana",       // が
	'\u30ac': "gakatakana",       // ガ
	'\u03b3': "gamma",            // γ
	'\u0263': "gammalatinsmall",  // ɣ
	'\u02e0': "gammasuperior",    // ˠ
	'\u03eb': "gangiacoptic",     // ϫ
	'\u310d': "gbopomofo",        // ㄍ
	'\u011f': "gbreve",           // ğ
	'\u01e7': "gcaron",           // ǧ
	'\u0123': "gcedilla",         // ģ
	'\u24d6': "gcircle",          // ⓖ
	'\u011d': "gcircumflex",      // ĝ
	// '\u0123':    "gcommaaccent", // ģ -- duplicate
	'\u0121': "gdot", // ġ
	// '\u0121':    "gdotaccent", // ġ -- duplicate
	// '\u0433':    "gecyrillic", // г -- duplicate
	'\u3052': "gehiragana",            // げ
	'\u30b2': "gekatakana",            // ゲ
	'\u2251': "geometricallyequal",    // ≑
	'\u059c': "gereshaccenthebrew",    // ֜
	'\u05f3': "gereshhebrew",          // ׳
	'\u059d': "gereshmuqdamhebrew",    // ֝
	'\u00df': "germandbls",            // ß
	'\u059e': "gershayimaccenthebrew", // ֞
	'\u05f4': "gershayimhebrew",       // ״
	'\u3013': "getamark",              // 〓
	'\u0998': "ghabengali",            // ঘ
	'\u0572': "ghadarmenian",          // ղ
	'\u0918': "ghadeva",               // घ
	'\u0a98': "ghagujarati",           // ઘ
	'\u0a18': "ghagurmukhi",           // ਘ
	// '\u063a':    "ghainarabic", // غ -- duplicate
	'\ufece': "ghainfinalarabic",      // ﻎ
	'\ufecf': "ghaininitialarabic",    // ﻏ
	'\ufed0': "ghainmedialarabic",     // ﻐ
	'\u0495': "ghemiddlehookcyrillic", // ҕ
	'\u0493': "ghestrokecyrillic",     // ғ
	// '\u0491':    "gheupturncyrillic", // ґ -- duplicate
	'\u095a': "ghhadeva",     // ग़
	'\u0a5a': "ghhagurmukhi", // ਗ਼
	'\u0260': "ghook",        // ɠ
	'\u3393': "ghzsquare",    // ㎓
	'\u304e': "gihiragana",   // ぎ
	'\u30ae': "gikatakana",   // ギ
	'\u0563': "gimarmenian",  // գ
	// '\u05d2':    "gimel", // ג -- duplicate
	'\ufb32': "gimeldagesh", // גּ
	// '\ufb32':    "gimeldageshhebrew", // גּ -- duplicate
	// '\u05d2':    "gimelhebrew", // ג -- duplicate
	// '\u0453':    "gjecyrillic", // ѓ -- duplicate
	'\u01be': "glottalinvertedstroke",       // ƾ
	'\u0294': "glottalstop",                 // ʔ
	'\u0296': "glottalstopinverted",         // ʖ
	'\u02c0': "glottalstopmod",              // ˀ
	'\u0295': "glottalstopreversed",         // ʕ
	'\u02c1': "glottalstopreversedmod",      // ˁ
	'\u02e4': "glottalstopreversedsuperior", // ˤ
	'\u02a1': "glottalstopstroke",           // ʡ
	'\u02a2': "glottalstopstrokereversed",   // ʢ
	'\u1e21': "gmacron",                     // ḡ
	'\uff47': "gmonospace",                  // ｇ
	'\u3054': "gohiragana",                  // ご
	'\u30b4': "gokatakana",                  // ゴ
	'\u24a2': "gparen",                      // ⒢
	'\u33ac': "gpasquare",                   // ㎬
	'\u2207': "gradient",                    // ∇
	'`':      "grave",                       // `
	'\u0316': "gravebelowcmb",               // ̖
	'\u0300': "gravecmb",                    // ̀
	// '\u0300':    "gravecomb", // ̀ -- duplicate
	'\u0953': "gravedeva",           // ॓
	'\u02ce': "gravelowmod",         // ˎ
	'\uff40': "gravemonospace",      // ｀
	'\u0340': "gravetonecmb",        // ̀
	'>':      "greater",             // >
	'\u2265': "greaterequal",        // ≥
	'\u22db': "greaterequalorless",  // ⋛
	'\uff1e': "greatermonospace",    // ＞
	'\u2273': "greaterorequivalent", // ≳
	'\u2277': "greaterorless",       // ≷
	'\u2267': "greateroverequal",    // ≧
	'\ufe65': "greatersmall",        // ﹥
	'\u0261': "gscript",             // ɡ
	'\u01e5': "gstroke",             // ǥ
	'\u3050': "guhiragana",          // ぐ
	'\u00ab': "guillemotleft",       // «
	'\u00bb': "guillemotright",      // »
	'\u2039': "guilsinglleft",       // ‹
	'\u203a': "guilsinglright",      // ›
	'\u30b0': "gukatakana",          // グ
	'\u3318': "guramusquare",        // ㌘
	'\u33c9': "gysquare",            // ㏉
	'h':      "h",                   // h
	'\u04a9': "haabkhasiancyrillic", // ҩ
	'\u06c1': "haaltonearabic",      // ہ
	'\u09b9': "habengali",           // হ
	'\u04b3': "hadescendercyrillic", // ҳ
	'\u0939': "hadeva",              // ह
	'\u0ab9': "hagujarati",          // હ
	'\u0a39': "hagurmukhi",          // ਹ
	// '\u062d':    "haharabic", // ح -- duplicate
	'\ufea2': "hahfinalarabic",      // ﺢ
	'\ufea3': "hahinitialarabic",    // ﺣ
	'\u306f': "hahiragana",          // は
	'\ufea4': "hahmedialarabic",     // ﺤ
	'\u332a': "haitusquare",         // ㌪
	'\u30cf': "hakatakana",          // ハ
	'\uff8a': "hakatakanahalfwidth", // ﾊ
	'\u0a4d': "halantgurmukhi",      // ੍
	// '\u0621':    "hamzaarabic", // ء -- duplicate
	// '\u0621':    "hamzadammaarabic", // ء -- duplicate
	// '\u0621':    "hamzadammatanarabic", // ء -- duplicate
	// '\u0621':    "hamzafathaarabic", // ء -- duplicate
	// '\u0621':    "hamzafathatanarabic", // ء -- duplicate
	// '\u0621':    "hamzalowarabic", // ء -- duplicate
	// '\u0621':    "hamzalowkasraarabic", // ء -- duplicate
	// '\u0621':    "hamzalowkasratanarabic", // ء -- duplicate
	// '\u0621':    "hamzasukunarabic", // ء -- duplicate
	'\u3164': "hangulfiller", // ㅤ
	// '\u044a':    "hardsigncyrillic", // ъ -- duplicate
	'\u21bc': "harpoonleftbarbup",  // ↼
	'\u21c0': "harpoonrightbarbup", // ⇀
	'\u33ca': "hasquare",           // ㏊
	// '\u05b2':    "hatafpatah", // ֲ -- duplicate
	// '\u05b2':    "hatafpatah16", // ֲ -- duplicate
	// '\u05b2':    "hatafpatah23", // ֲ -- duplicate
	// '\u05b2':    "hatafpatah2f", // ֲ -- duplicate
	// '\u05b2':    "hatafpatahhebrew", // ֲ -- duplicate
	// '\u05b2':    "hatafpatahnarrowhebrew", // ֲ -- duplicate
	// '\u05b2':    "hatafpatahquarterhebrew", // ֲ -- duplicate
	// '\u05b2':    "hatafpatahwidehebrew", // ֲ -- duplicate
	// '\u05b3':    "hatafqamats", // ֳ -- duplicate
	// '\u05b3':    "hatafqamats1b", // ֳ -- duplicate
	// '\u05b3':    "hatafqamats28", // ֳ -- duplicate
	// '\u05b3':    "hatafqamats34", // ֳ -- duplicate
	// '\u05b3':    "hatafqamatshebrew", // ֳ -- duplicate
	// '\u05b3':    "hatafqamatsnarrowhebrew", // ֳ -- duplicate
	// '\u05b3':    "hatafqamatsquarterhebrew", // ֳ -- duplicate
	// '\u05b3':    "hatafqamatswidehebrew", // ֳ -- duplicate
	// '\u05b1':    "hatafsegol", // ֱ -- duplicate
	// '\u05b1':    "hatafsegol17", // ֱ -- duplicate
	// '\u05b1':    "hatafsegol24", // ֱ -- duplicate
	// '\u05b1':    "hatafsegol30", // ֱ -- duplicate
	// '\u05b1':    "hatafsegolhebrew", // ֱ -- duplicate
	// '\u05b1':    "hatafsegolnarrowhebrew", // ֱ -- duplicate
	// '\u05b1':    "hatafsegolquarterhebrew", // ֱ -- duplicate
	// '\u05b1':    "hatafsegolwidehebrew", // ֱ -- duplicate
	'\u0127': "hbar",        // ħ
	'\u310f': "hbopomofo",   // ㄏ
	'\u1e2b': "hbrevebelow", // ḫ
	'\u1e29': "hcedilla",    // ḩ
	'\u24d7': "hcircle",     // ⓗ
	'\u0125': "hcircumflex", // ĥ
	'\u1e27': "hdieresis",   // ḧ
	'\u1e23': "hdotaccent",  // ḣ
	'\u1e25': "hdotbelow",   // ḥ
	// '\u05d4':    "he", // ה -- duplicate
	'\u2665': "heart", // ♥
	// '\u2665':    "heartsuitblack", // ♥ -- duplicate
	'\u2661': "heartsuitwhite", // ♡
	'\ufb34': "hedagesh",       // הּ
	// '\ufb34':    "hedageshhebrew", // הּ -- duplicate
	// '\u06c1':    "hehaltonearabic", // ہ -- duplicate
	// '\u0647':    "heharabic", // ه -- duplicate
	// '\u05d4':    "hehebrew", // ה -- duplicate
	'\ufba7': "hehfinalaltonearabic", // ﮧ
	'\ufeea': "hehfinalalttwoarabic", // ﻪ
	// '\ufeea':    "hehfinalarabic", // ﻪ -- duplicate
	'\ufba5': "hehhamzaabovefinalarabic",    // ﮥ
	'\ufba4': "hehhamzaaboveisolatedarabic", // ﮤ
	'\ufba8': "hehinitialaltonearabic",      // ﮨ
	'\ufeeb': "hehinitialarabic",            // ﻫ
	'\u3078': "hehiragana",                  // へ
	'\ufba9': "hehmedialaltonearabic",       // ﮩ
	'\ufeec': "hehmedialarabic",             // ﻬ
	'\u337b': "heiseierasquare",             // ㍻
	'\u30d8': "hekatakana",                  // ヘ
	'\uff8d': "hekatakanahalfwidth",         // ﾍ
	'\u3336': "hekutaarusquare",             // ㌶
	'\u0267': "henghook",                    // ɧ
	'\u3339': "herutusquare",                // ㌹
	// '\u05d7':    "het", // ח -- duplicate
	// '\u05d7':    "hethebrew", // ח -- duplicate
	'\u0266': "hhook",               // ɦ
	'\u02b1': "hhooksuperior",       // ʱ
	'\u327b': "hieuhacirclekorean",  // ㉻
	'\u321b': "hieuhaparenkorean",   // ㈛
	'\u326d': "hieuhcirclekorean",   // ㉭
	'\u314e': "hieuhkorean",         // ㅎ
	'\u320d': "hieuhparenkorean",    // ㈍
	'\u3072': "hihiragana",          // ひ
	'\u30d2': "hikatakana",          // ヒ
	'\uff8b': "hikatakanahalfwidth", // ﾋ
	// '\u05b4':    "hiriq", // ִ -- duplicate
	// '\u05b4':    "hiriq14", // ִ -- duplicate
	// '\u05b4':    "hiriq21", // ִ -- duplicate
	// '\u05b4':    "hiriq2d", // ִ -- duplicate
	// '\u05b4':    "hiriqhebrew", // ִ -- duplicate
	// '\u05b4':    "hiriqnarrowhebrew", // ִ -- duplicate
	// '\u05b4':    "hiriqquarterhebrew", // ִ -- duplicate
	// '\u05b4':    "hiriqwidehebrew", // ִ -- duplicate
	'\u1e96': "hlinebelow",          // ẖ
	'\uff48': "hmonospace",          // ｈ
	'\u0570': "hoarmenian",          // հ
	'\u0e2b': "hohipthai",           // ห
	'\u307b': "hohiragana",          // ほ
	'\u30db': "hokatakana",          // ホ
	'\uff8e': "hokatakanahalfwidth", // ﾎ
	// '\u05b9':    "holam", // ֹ -- duplicate
	// '\u05b9':    "holam19", // ֹ -- duplicate
	// '\u05b9':    "holam26", // ֹ -- duplicate
	// '\u05b9':    "holam32", // ֹ -- duplicate
	// '\u05b9':    "holamhebrew", // ֹ -- duplicate
	// '\u05b9':    "holamnarrowhebrew", // ֹ -- duplicate
	// '\u05b9':    "holamquarterhebrew", // ֹ -- duplicate
	// '\u05b9':    "holamwidehebrew", // ֹ -- duplicate
	'\u0e2e': "honokhukthai",  // ฮ
	'\u0309': "hookabovecomb", // ̉
	// '\u0309':    "hookcmb", // ̉ -- duplicate
	'\u0321': "hookpalatalizedbelowcmb", // ̡
	'\u0322': "hookretroflexbelowcmb",   // ̢
	'\u3342': "hoonsquare",              // ㍂
	'\u03e9': "horicoptic",              // ϩ
	// '\u2015':    "horizontalbar", // ― -- duplicate
	'\u031b': "horncmb",             // ̛
	'\u2668': "hotsprings",          // ♨
	'\u2302': "house",               // ⌂
	'\u24a3': "hparen",              // ⒣
	'\u02b0': "hsuperior",           // ʰ
	'\u0265': "hturned",             // ɥ
	'\u3075': "huhiragana",          // ふ
	'\u3333': "huiitosquare",        // ㌳
	'\u30d5': "hukatakana",          // フ
	'\uff8c': "hukatakanahalfwidth", // ﾌ
	'\u02dd': "hungarumlaut",        // ˝
	'\u030b': "hungarumlautcmb",     // ̋
	'\u0195': "hv",                  // ƕ
	'-':      "hyphen",              // -
	'\uf6e5': "hypheninferior",
	'\uff0d': "hyphenmonospace", // －
	'\ufe63': "hyphensmall",     // ﹣
	'\uf6e6': "hyphensuperior",
	'\u2010': "hyphentwo", // ‐
	'i':      "i",         // i
	'\u00ed': "iacute",    // í
	// '\u044f':    "iacyrillic", // я -- duplicate
	'\u0987': "ibengali",    // ই
	'\u3127': "ibopomofo",   // ㄧ
	'\u012d': "ibreve",      // ĭ
	'\u01d0': "icaron",      // ǐ
	'\u24d8': "icircle",     // ⓘ
	'\u00ee': "icircumflex", // î
	// '\u0456':    "icyrillic", // і -- duplicate
	'\u0209': "idblgrave",                      // ȉ
	'\u328f': "ideographearthcircle",           // ㊏
	'\u328b': "ideographfirecircle",            // ㊋
	'\u323f': "ideographicallianceparen",       // ㈿
	'\u323a': "ideographiccallparen",           // ㈺
	'\u32a5': "ideographiccentrecircle",        // ㊥
	'\u3006': "ideographicclose",               // 〆
	'\u3001': "ideographiccomma",               // 、
	'\uff64': "ideographiccommaleft",           // ､
	'\u3237': "ideographiccongratulationparen", // ㈷
	'\u32a3': "ideographiccorrectcircle",       // ㊣
	'\u322f': "ideographicearthparen",          // ㈯
	'\u323d': "ideographicenterpriseparen",     // ㈽
	'\u329d': "ideographicexcellentcircle",     // ㊝
	'\u3240': "ideographicfestivalparen",       // ㉀
	'\u3296': "ideographicfinancialcircle",     // ㊖
	'\u3236': "ideographicfinancialparen",      // ㈶
	'\u322b': "ideographicfireparen",           // ㈫
	'\u3232': "ideographichaveparen",           // ㈲
	'\u32a4': "ideographichighcircle",          // ㊤
	'\u3005': "ideographiciterationmark",       // 々
	'\u3298': "ideographiclaborcircle",         // ㊘
	'\u3238': "ideographiclaborparen",          // ㈸
	'\u32a7': "ideographicleftcircle",          // ㊧
	'\u32a6': "ideographiclowcircle",           // ㊦
	'\u32a9': "ideographicmedicinecircle",      // ㊩
	'\u322e': "ideographicmetalparen",          // ㈮
	'\u322a': "ideographicmoonparen",           // ㈪
	'\u3234': "ideographicnameparen",           // ㈴
	'\u3002': "ideographicperiod",              // 。
	'\u329e': "ideographicprintcircle",         // ㊞
	'\u3243': "ideographicreachparen",          // ㉃
	'\u3239': "ideographicrepresentparen",      // ㈹
	'\u323e': "ideographicresourceparen",       // ㈾
	'\u32a8': "ideographicrightcircle",         // ㊨
	'\u3299': "ideographicsecretcircle",        // ㊙
	'\u3242': "ideographicselfparen",           // ㉂
	'\u3233': "ideographicsocietyparen",        // ㈳
	'\u3000': "ideographicspace",
	'\u3235': "ideographicspecialparen",   // ㈵
	'\u3231': "ideographicstockparen",     // ㈱
	'\u323b': "ideographicstudyparen",     // ㈻
	'\u3230': "ideographicsunparen",       // ㈰
	'\u323c': "ideographicsuperviseparen", // ㈼
	'\u322c': "ideographicwaterparen",     // ㈬
	'\u322d': "ideographicwoodparen",      // ㈭
	'\u3007': "ideographiczero",           // 〇
	'\u328e': "ideographmetalcircle",      // ㊎
	'\u328a': "ideographmooncircle",       // ㊊
	'\u3294': "ideographnamecircle",       // ㊔
	'\u3290': "ideographsuncircle",        // ㊐
	'\u328c': "ideographwatercircle",      // ㊌
	'\u328d': "ideographwoodcircle",       // ㊍
	'\u0907': "ideva",                     // इ
	'\u00ef': "idieresis",                 // ï
	'\u1e2f': "idieresisacute",            // ḯ
	'\u04e5': "idieresiscyrillic",         // ӥ
	'\u1ecb': "idotbelow",                 // ị
	'\u04d7': "iebrevecyrillic",           // ӗ
	// '\u0435':    "iecyrillic", // е -- duplicate
	'\u3275': "ieungacirclekorean", // ㉵
	'\u3215': "ieungaparenkorean",  // ㈕
	'\u3267': "ieungcirclekorean",  // ㉧
	'\u3147': "ieungkorean",        // ㅇ
	'\u3207': "ieungparenkorean",   // ㈇
	'\u00ec': "igrave",             // ì
	'\u0a87': "igujarati",          // ઇ
	'\u0a07': "igurmukhi",          // ਇ
	'\u3044': "ihiragana",          // い
	'\u1ec9': "ihookabove",         // ỉ
	'\u0988': "iibengali",          // ঈ
	// '\u0438':    "iicyrillic", // и -- duplicate
	'\u0908': "iideva",          // ई
	'\u0a88': "iigujarati",      // ઈ
	'\u0a08': "iigurmukhi",      // ਈ
	'\u0a40': "iimatragurmukhi", // ੀ
	'\u020b': "iinvertedbreve",  // ȋ
	// '\u0439':    "iishortcyrillic", // й -- duplicate
	'\u09c0': "iivowelsignbengali",        // ী
	'\u0940': "iivowelsigndeva",           // ी
	'\u0ac0': "iivowelsigngujarati",       // ી
	'\u0133': "ij",                        // ĳ
	'\u30a4': "ikatakana",                 // イ
	'\uff72': "ikatakanahalfwidth",        // ｲ
	'\u3163': "ikorean",                   // ㅣ
	'\u02dc': "ilde",                      // ˜
	'\u05ac': "iluyhebrew",                // ֬
	'\u012b': "imacron",                   // ī
	'\u04e3': "imacroncyrillic",           // ӣ
	'\u2253': "imageorapproximatelyequal", // ≓
	'\u0a3f': "imatragurmukhi",            // ਿ
	'\uff49': "imonospace",                // ｉ
	// '\u2206':    "increment", // ∆ -- duplicate
	'\u221e': "infinity",       // ∞
	'\u056b': "iniarmenian",    // ի
	'\u222b': "integral",       // ∫
	'\u2321': "integralbottom", // ⌡
	// '\u2321':    "integralbt", // ⌡ -- duplicate
	'\uf8f5': "integralex",
	'\u2320': "integraltop", // ⌠
	// '\u2320':    "integraltp", // ⌠ -- duplicate
	'\u2229': "intersection", // ∩
	'\u3305': "intisquare",   // ㌅
	// '\u25d8':    "invbullet", // ◘ -- duplicate
	'\u25d9': "invcircle", // ◙
	// '\u263b':    "invsmileface", // ☻ -- duplicate
	// '\u0451':    "iocyrillic", // ё -- duplicate
	'\u012f': "iogonek",                 // į
	'\u03b9': "iota",                    // ι
	'\u03ca': "iotadieresis",            // ϊ
	'\u0390': "iotadieresistonos",       // ΐ
	'\u0269': "iotalatin",               // ɩ
	'\u03af': "iotatonos",               // ί
	'\u24a4': "iparen",                  // ⒤
	'\u0a72': "irigurmukhi",             // ੲ
	'\u3043': "ismallhiragana",          // ぃ
	'\u30a3': "ismallkatakana",          // ィ
	'\uff68': "ismallkatakanahalfwidth", // ｨ
	'\u09fa': "issharbengali",           // ৺
	'\u0268': "istroke",                 // ɨ
	'\uf6ed': "isuperior",
	'\u309d': "iterationhiragana", // ゝ
	'\u30fd': "iterationkatakana", // ヽ
	'\u0129': "itilde",            // ĩ
	'\u1e2d': "itildebelow",       // ḭ
	'\u3129': "iubopomofo",        // ㄩ
	// '\u044e':    "iucyrillic", // ю -- duplicate
	'\u09bf': "ivowelsignbengali",  // ি
	'\u093f': "ivowelsigndeva",     // ि
	'\u0abf': "ivowelsigngujarati", // િ
	// '\u0475':    "izhitsacyrillic", // ѵ -- duplicate
	'\u0477': "izhitsadblgravecyrillic", // ѷ
	'j':      "j",                       // j
	'\u0571': "jaarmenian",              // ձ
	'\u099c': "jabengali",               // জ
	'\u091c': "jadeva",                  // ज
	'\u0a9c': "jagujarati",              // જ
	'\u0a1c': "jagurmukhi",              // ਜ
	'\u3110': "jbopomofo",               // ㄐ
	'\u01f0': "jcaron",                  // ǰ
	'\u24d9': "jcircle",                 // ⓙ
	'\u0135': "jcircumflex",             // ĵ
	'\u029d': "jcrossedtail",            // ʝ
	'\u025f': "jdotlessstroke",          // ɟ
	// '\u0458':    "jecyrillic", // ј -- duplicate
	// '\u062c':    "jeemarabic", // ج -- duplicate
	'\ufe9e': "jeemfinalarabic",   // ﺞ
	'\ufe9f': "jeeminitialarabic", // ﺟ
	'\ufea0': "jeemmedialarabic",  // ﺠ
	// '\u0698':    "jeharabic", // ژ -- duplicate
	'\ufb8b': "jehfinalarabic",    // ﮋ
	'\u099d': "jhabengali",        // ঝ
	'\u091d': "jhadeva",           // झ
	'\u0a9d': "jhagujarati",       // ઝ
	'\u0a1d': "jhagurmukhi",       // ਝ
	'\u057b': "jheharmenian",      // ջ
	'\u3004': "jis",               // 〄
	'\uff4a': "jmonospace",        // ｊ
	'\u24a5': "jparen",            // ⒥
	'\u02b2': "jsuperior",         // ʲ
	'k':      "k",                 // k
	'\u04a1': "kabashkircyrillic", // ҡ
	'\u0995': "kabengali",         // ক
	'\u1e31': "kacute",            // ḱ
	// '\u043a':    "kacyrillic", // к -- duplicate
	'\u049b': "kadescendercyrillic", // қ
	'\u0915': "kadeva",              // क
	// '\u05db':    "kaf", // כ -- duplicate
	// '\u0643':    "kafarabic", // ك -- duplicate
	'\ufb3b': "kafdagesh", // כּ
	// '\ufb3b':    "kafdageshhebrew", // כּ -- duplicate
	'\ufeda': "kaffinalarabic", // ﻚ
	// '\u05db':    "kafhebrew", // כ -- duplicate
	'\ufedb': "kafinitialarabic",         // ﻛ
	'\ufedc': "kafmedialarabic",          // ﻜ
	'\ufb4d': "kafrafehebrew",            // כֿ
	'\u0a95': "kagujarati",               // ક
	'\u0a15': "kagurmukhi",               // ਕ
	'\u304b': "kahiragana",               // か
	'\u04c4': "kahookcyrillic",           // ӄ
	'\u30ab': "kakatakana",               // カ
	'\uff76': "kakatakanahalfwidth",      // ｶ
	'\u03ba': "kappa",                    // κ
	'\u03f0': "kappasymbolgreek",         // ϰ
	'\u3171': "kapyeounmieumkorean",      // ㅱ
	'\u3184': "kapyeounphieuphkorean",    // ㆄ
	'\u3178': "kapyeounpieupkorean",      // ㅸ
	'\u3179': "kapyeounssangpieupkorean", // ㅹ
	'\u330d': "karoriisquare",            // ㌍
	// '\u0640':    "kashidaautoarabic", // ـ -- duplicate
	// '\u0640':    "kashidaautonosidebearingarabic", // ـ -- duplicate
	'\u30f5': "kasmallkatakana", // ヵ
	'\u3384': "kasquare",        // ㎄
	// '\u0650':    "kasraarabic", // ِ -- duplicate
	// '\u064d':    "kasratanarabic", // ٍ -- duplicate
	'\u049f': "kastrokecyrillic",             // ҟ
	'\uff70': "katahiraprolongmarkhalfwidth", // ｰ
	'\u049d': "kaverticalstrokecyrillic",     // ҝ
	'\u310e': "kbopomofo",                    // ㄎ
	'\u3389': "kcalsquare",                   // ㎉
	'\u01e9': "kcaron",                       // ǩ
	'\u0137': "kcedilla",                     // ķ
	'\u24da': "kcircle",                      // ⓚ
	// '\u0137':    "kcommaaccent", // ķ -- duplicate
	'\u1e33': "kdotbelow",           // ḳ
	'\u0584': "keharmenian",         // ք
	'\u3051': "kehiragana",          // け
	'\u30b1': "kekatakana",          // ケ
	'\uff79': "kekatakanahalfwidth", // ｹ
	'\u056f': "kenarmenian",         // կ
	'\u30f6': "kesmallkatakana",     // ヶ
	'\u0138': "kgreenlandic",        // ĸ
	'\u0996': "khabengali",          // খ
	// '\u0445':    "khacyrillic", // х -- duplicate
	'\u0916': "khadeva",     // ख
	'\u0a96': "khagujarati", // ખ
	'\u0a16': "khagurmukhi", // ਖ
	// '\u062e':    "khaharabic", // خ -- duplicate
	'\ufea6': "khahfinalarabic",      // ﺦ
	'\ufea7': "khahinitialarabic",    // ﺧ
	'\ufea8': "khahmedialarabic",     // ﺨ
	'\u03e7': "kheicoptic",           // ϧ
	'\u0959': "khhadeva",             // ख़
	'\u0a59': "khhagurmukhi",         // ਖ਼
	'\u3278': "khieukhacirclekorean", // ㉸
	'\u3218': "khieukhaparenkorean",  // ㈘
	'\u326a': "khieukhcirclekorean",  // ㉪
	'\u314b': "khieukhkorean",        // ㅋ
	'\u320a': "khieukhparenkorean",   // ㈊
	'\u0e02': "khokhaithai",          // ข
	'\u0e05': "khokhonthai",          // ฅ
	'\u0e03': "khokhuatthai",         // ฃ
	'\u0e04': "khokhwaithai",         // ค
	'\u0e5b': "khomutthai",           // ๛
	'\u0199': "khook",                // ƙ
	'\u0e06': "khorakhangthai",       // ฆ
	'\u3391': "khzsquare",            // ㎑
	'\u304d': "kihiragana",           // き
	'\u30ad': "kikatakana",           // キ
	'\uff77': "kikatakanahalfwidth",  // ｷ
	'\u3315': "kiroguramusquare",     // ㌕
	'\u3316': "kiromeetorusquare",    // ㌖
	'\u3314': "kirosquare",           // ㌔
	'\u326e': "kiyeokacirclekorean",  // ㉮
	'\u320e': "kiyeokaparenkorean",   // ㈎
	'\u3260': "kiyeokcirclekorean",   // ㉠
	'\u3131': "kiyeokkorean",         // ㄱ
	'\u3200': "kiyeokparenkorean",    // ㈀
	'\u3133': "kiyeoksioskorean",     // ㄳ
	// '\u045c':    "kjecyrillic", // ќ -- duplicate
	'\u1e35': "klinebelow",                      // ḵ
	'\u3398': "klsquare",                        // ㎘
	'\u33a6': "kmcubedsquare",                   // ㎦
	'\uff4b': "kmonospace",                      // ｋ
	'\u33a2': "kmsquaredsquare",                 // ㎢
	'\u3053': "kohiragana",                      // こ
	'\u33c0': "kohmsquare",                      // ㏀
	'\u0e01': "kokaithai",                       // ก
	'\u30b3': "kokatakana",                      // コ
	'\uff7a': "kokatakanahalfwidth",             // ｺ
	'\u331e': "kooposquare",                     // ㌞
	'\u0481': "koppacyrillic",                   // ҁ
	'\u327f': "koreanstandardsymbol",            // ㉿
	'\u0343': "koroniscmb",                      // ̓
	'\u24a6': "kparen",                          // ⒦
	'\u33aa': "kpasquare",                       // ㎪
	'\u046f': "ksicyrillic",                     // ѯ
	'\u33cf': "ktsquare",                        // ㏏
	'\u029e': "kturned",                         // ʞ
	'\u304f': "kuhiragana",                      // く
	'\u30af': "kukatakana",                      // ク
	'\uff78': "kukatakanahalfwidth",             // ｸ
	'\u33b8': "kvsquare",                        // ㎸
	'\u33be': "kwsquare",                        // ㎾
	'l':      "l",                               // l
	'\u09b2': "labengali",                       // ল
	'\u013a': "lacute",                          // ĺ
	'\u0932': "ladeva",                          // ल
	'\u0ab2': "lagujarati",                      // લ
	'\u0a32': "lagurmukhi",                      // ਲ
	'\u0e45': "lakkhangyaothai",                 // ๅ
	'\ufefc': "lamaleffinalarabic",              // ﻼ
	'\ufef8': "lamalefhamzaabovefinalarabic",    // ﻸ
	'\ufef7': "lamalefhamzaaboveisolatedarabic", // ﻷ
	'\ufefa': "lamalefhamzabelowfinalarabic",    // ﻺ
	'\ufef9': "lamalefhamzabelowisolatedarabic", // ﻹ
	'\ufefb': "lamalefisolatedarabic",           // ﻻ
	'\ufef6': "lamalefmaddaabovefinalarabic",    // ﻶ
	'\ufef5': "lamalefmaddaaboveisolatedarabic", // ﻵ
	// '\u0644':    "lamarabic", // ل -- duplicate
	'\u03bb': "lambda",       // λ
	'\u019b': "lambdastroke", // ƛ
	// '\u05dc':    "lamed", // ל -- duplicate
	'\ufb3c': "lameddagesh", // לּ
	// '\ufb3c':    "lameddageshhebrew", // לּ -- duplicate
	// '\u05dc':    "lamedhebrew", // ל -- duplicate
	// '\u05dc':    "lamedholam", // ל -- duplicate
	// '\u05dc':    "lamedholamdagesh", // ל -- duplicate
	// '\u05dc':    "lamedholamdageshhebrew", // ל -- duplicate
	// '\u05dc':    "lamedholamhebrew", // ל -- duplicate
	'\ufede': "lamfinalarabic",          // ﻞ
	'\ufcca': "lamhahinitialarabic",     // ﳊ
	'\ufedf': "laminitialarabic",        // ﻟ
	'\ufcc9': "lamjeeminitialarabic",    // ﳉ
	'\ufccb': "lamkhahinitialarabic",    // ﳋ
	'\ufdf2': "lamlamhehisolatedarabic", // ﷲ
	'\ufee0': "lammedialarabic",         // ﻠ
	'\ufd88': "lammeemhahinitialarabic", // ﶈ
	'\ufccc': "lammeeminitialarabic",    // ﳌ
	// '\ufedf':    "lammeemjeeminitialarabic", // ﻟ -- duplicate
	// '\ufedf':    "lammeemkhahinitialarabic", // ﻟ -- duplicate
	'\u25ef': "largecircle",      // ◯
	'\u019a': "lbar",             // ƚ
	'\u026c': "lbelt",            // ɬ
	'\u310c': "lbopomofo",        // ㄌ
	'\u013e': "lcaron",           // ľ
	'\u013c': "lcedilla",         // ļ
	'\u24db': "lcircle",          // ⓛ
	'\u1e3d': "lcircumflexbelow", // ḽ
	// '\u013c':    "lcommaaccent", // ļ -- duplicate
	'\u0140': "ldot", // ŀ
	// '\u0140':    "ldotaccent", // ŀ -- duplicate
	'\u1e37': "ldotbelow",          // ḷ
	'\u1e39': "ldotbelowmacron",    // ḹ
	'\u031a': "leftangleabovecmb",  // ̚
	'\u0318': "lefttackbelowcmb",   // ̘
	'<':      "less",               // <
	'\u2264': "lessequal",          // ≤
	'\u22da': "lessequalorgreater", // ⋚
	'\uff1c': "lessmonospace",      // ＜
	'\u2272': "lessorequivalent",   // ≲
	'\u2276': "lessorgreater",      // ≶
	'\u2266': "lessoverequal",      // ≦
	'\ufe64': "lesssmall",          // ﹤
	'\u026e': "lezh",               // ɮ
	'\u258c': "lfblock",            // ▌
	'\u026d': "lhookretroflex",     // ɭ
	// '\u20a4':    "lira", // ₤ -- duplicate
	'\u056c': "liwnarmenian", // լ
	'\u01c9': "lj",           // ǉ
	// '\u0459':    "ljecyrillic", // љ -- duplicate
	'\uf6c0': "ll",
	'\u0933': "lladeva",                   // ळ
	'\u0ab3': "llagujarati",               // ળ
	'\u1e3b': "llinebelow",                // ḻ
	'\u0934': "llladeva",                  // ऴ
	'\u09e1': "llvocalicbengali",          // ৡ
	'\u0961': "llvocalicdeva",             // ॡ
	'\u09e3': "llvocalicvowelsignbengali", // ৣ
	'\u0963': "llvocalicvowelsigndeva",    // ॣ
	'\u026b': "lmiddletilde",              // ɫ
	'\uff4c': "lmonospace",                // ｌ
	'\u33d0': "lmsquare",                  // ㏐
	'\u0e2c': "lochulathai",               // ฬ
	'\u2227': "logicaland",                // ∧
	'\u00ac': "logicalnot",                // ¬
	'\u2310': "logicalnotreversed",        // ⌐
	'\u2228': "logicalor",                 // ∨
	'\u0e25': "lolingthai",                // ล
	'\u017f': "longs",                     // ſ
	'\ufe4e': "lowlinecenterline",         // ﹎
	'\u0332': "lowlinecmb",                // ̲
	'\ufe4d': "lowlinedashed",             // ﹍
	'\u25ca': "lozenge",                   // ◊
	'\u24a7': "lparen",                    // ⒧
	'\u0142': "lslash",                    // ł
	// '\u2113':    "lsquare", // ℓ -- duplicate
	'\uf6ee': "lsuperior",
	'\u2591': "ltshade",                  // ░
	'\u0e26': "luthai",                   // ฦ
	'\u098c': "lvocalicbengali",          // ঌ
	'\u090c': "lvocalicdeva",             // ऌ
	'\u09e2': "lvocalicvowelsignbengali", // ৢ
	'\u0962': "lvocalicvowelsigndeva",    // ॢ
	'\u33d3': "lxsquare",                 // ㏓
	'm':      "m",                        // m
	'\u09ae': "mabengali",                // ম
	'\u00af': "macron",                   // ¯
	'\u0331': "macronbelowcmb",           // ̱
	'\u0304': "macroncmb",                // ̄
	'\u02cd': "macronlowmod",             // ˍ
	'\uffe3': "macronmonospace",          // ￣
	'\u1e3f': "macute",                   // ḿ
	'\u092e': "madeva",                   // म
	'\u0aae': "magujarati",               // મ
	'\u0a2e': "magurmukhi",               // ਮ
	'\u05a4': "mahapakhhebrew",           // ֤
	// '\u05a4':    "mahapakhlefthebrew", // ֤ -- duplicate
	'\u307e': "mahiragana", // ま
	'\uf895': "maichattawalowleftthai",
	'\uf894': "maichattawalowrightthai",
	'\u0e4b': "maichattawathai", // ๋
	'\uf893': "maichattawaupperleftthai",
	'\uf88c': "maieklowleftthai",
	'\uf88b': "maieklowrightthai",
	'\u0e48': "maiekthai", // ่
	'\uf88a': "maiekupperleftthai",
	'\uf884': "maihanakatleftthai",
	'\u0e31': "maihanakatthai", // ั
	'\uf889': "maitaikhuleftthai",
	'\u0e47': "maitaikhuthai", // ็
	'\uf88f': "maitholowleftthai",
	'\uf88e': "maitholowrightthai",
	'\u0e49': "maithothai", // ้
	'\uf88d': "maithoupperleftthai",
	'\uf892': "maitrilowleftthai",
	'\uf891': "maitrilowrightthai",
	'\u0e4a': "maitrithai", // ๊
	'\uf890': "maitriupperleftthai",
	'\u0e46': "maiyamokthai",        // ๆ
	'\u30de': "makatakana",          // マ
	'\uff8f': "makatakanahalfwidth", // ﾏ
	'\u2642': "male",                // ♂
	'\u3347': "mansyonsquare",       // ㍇
	// '\u05be':    "maqafhebrew", // ־ -- duplicate
	// '\u2642':    "mars", // ♂ -- duplicate
	'\u05af': "masoracirclehebrew", // ֯
	'\u3383': "masquare",           // ㎃
	'\u3107': "mbopomofo",          // ㄇ
	'\u33d4': "mbsquare",           // ㏔
	'\u24dc': "mcircle",            // ⓜ
	'\u33a5': "mcubedsquare",       // ㎥
	'\u1e41': "mdotaccent",         // ṁ
	'\u1e43': "mdotbelow",          // ṃ
	// '\u0645':    "meemarabic", // م -- duplicate
	'\ufee2': "meemfinalarabic",        // ﻢ
	'\ufee3': "meeminitialarabic",      // ﻣ
	'\ufee4': "meemmedialarabic",       // ﻤ
	'\ufcd1': "meemmeeminitialarabic",  // ﳑ
	'\ufc48': "meemmeemisolatedarabic", // ﱈ
	'\u334d': "meetorusquare",          // ㍍
	'\u3081': "mehiragana",             // め
	'\u337e': "meizierasquare",         // ㍾
	'\u30e1': "mekatakana",             // メ
	'\uff92': "mekatakanahalfwidth",    // ﾒ
	// '\u05de':    "mem", // מ -- duplicate
	'\ufb3e': "memdagesh", // מּ
	// '\ufb3e':    "memdageshhebrew", // מּ -- duplicate
	// '\u05de':    "memhebrew", // מ -- duplicate
	'\u0574': "menarmenian",        // մ
	'\u05a5': "merkhahebrew",       // ֥
	'\u05a6': "merkhakefulahebrew", // ֦
	// '\u05a6':    "merkhakefulalefthebrew", // ֦ -- duplicate
	// '\u05a5':    "merkhalefthebrew", // ֥ -- duplicate
	'\u0271': "mhook",                      // ɱ
	'\u3392': "mhzsquare",                  // ㎒
	'\uff65': "middledotkatakanahalfwidth", // ･
	'\u00b7': "middot",                     // ·
	'\u3272': "mieumacirclekorean",         // ㉲
	'\u3212': "mieumaparenkorean",          // ㈒
	'\u3264': "mieumcirclekorean",          // ㉤
	'\u3141': "mieumkorean",                // ㅁ
	'\u3170': "mieumpansioskorean",         // ㅰ
	'\u3204': "mieumparenkorean",           // ㈄
	'\u316e': "mieumpieupkorean",           // ㅮ
	'\u316f': "mieumsioskorean",            // ㅯ
	'\u307f': "mihiragana",                 // み
	'\u30df': "mikatakana",                 // ミ
	'\uff90': "mikatakanahalfwidth",        // ﾐ
	'\u2212': "minus",                      // −
	'\u0320': "minusbelowcmb",              // ̠
	'\u2296': "minuscircle",                // ⊖
	'\u02d7': "minusmod",                   // ˗
	'\u2213': "minusplus",                  // ∓
	'\u2032': "minute",                     // ′
	'\u334a': "miribaarusquare",            // ㍊
	'\u3349': "mirisquare",                 // ㍉
	'\u0270': "mlonglegturned",             // ɰ
	'\u3396': "mlsquare",                   // ㎖
	'\u33a3': "mmcubedsquare",              // ㎣
	'\uff4d': "mmonospace",                 // ｍ
	'\u339f': "mmsquaredsquare",            // ㎟
	'\u3082': "mohiragana",                 // も
	'\u33c1': "mohmsquare",                 // ㏁
	'\u30e2': "mokatakana",                 // モ
	'\uff93': "mokatakanahalfwidth",        // ﾓ
	'\u33d6': "molsquare",                  // ㏖
	'\u0e21': "momathai",                   // ม
	'\u33a7': "moverssquare",               // ㎧
	'\u33a8': "moverssquaredsquare",        // ㎨
	'\u24a8': "mparen",                     // ⒨
	'\u33ab': "mpasquare",                  // ㎫
	'\u33b3': "mssquare",                   // ㎳
	'\uf6ef': "msuperior",
	'\u026f': "mturned", // ɯ
	'\u00b5': "mu",      // µ
	// '\u00b5':    "mu1", // µ -- duplicate
	'\u3382': "muasquare",           // ㎂
	'\u226b': "muchgreater",         // ≫
	'\u226a': "muchless",            // ≪
	'\u338c': "mufsquare",           // ㎌
	'\u03bc': "mugreek",             // μ
	'\u338d': "mugsquare",           // ㎍
	'\u3080': "muhiragana",          // む
	'\u30e0': "mukatakana",          // ム
	'\uff91': "mukatakanahalfwidth", // ﾑ
	'\u3395': "mulsquare",           // ㎕
	'\u00d7': "multiply",            // ×
	'\u339b': "mumsquare",           // ㎛
	'\u05a3': "munahhebrew",         // ֣
	// '\u05a3':    "munahlefthebrew", // ֣ -- duplicate
	'\u266a': "musicalnote", // ♪
	// '\u266b':    "musicalnotedbl", // ♫ -- duplicate
	'\u266d': "musicflatsign",  // ♭
	'\u266f': "musicsharpsign", // ♯
	'\u33b2': "mussquare",      // ㎲
	'\u33b6': "muvsquare",      // ㎶
	'\u33bc': "muwsquare",      // ㎼
	'\u33b9': "mvmegasquare",   // ㎹
	'\u33b7': "mvsquare",       // ㎷
	'\u33bf': "mwmegasquare",   // ㎿
	'\u33bd': "mwsquare",       // ㎽
	'n':      "n",              // n
	'\u09a8': "nabengali",      // ন
	// '\u2207':    "nabla", // ∇ -- duplicate
	'\u0144': "nacute",              // ń
	'\u0928': "nadeva",              // न
	'\u0aa8': "nagujarati",          // ન
	'\u0a28': "nagurmukhi",          // ਨ
	'\u306a': "nahiragana",          // な
	'\u30ca': "nakatakana",          // ナ
	'\uff85': "nakatakanahalfwidth", // ﾅ
	'\u0149': "napostrophe",         // ŉ
	'\u3381': "nasquare",            // ㎁
	'\u310b': "nbopomofo",           // ㄋ
	'\u00a0': "nbspace",
	'\u0148': "ncaron",           // ň
	'\u0146': "ncedilla",         // ņ
	'\u24dd': "ncircle",          // ⓝ
	'\u1e4b': "ncircumflexbelow", // ṋ
	// '\u0146':    "ncommaaccent", // ņ -- duplicate
	'\u1e45': "ndotaccent",          // ṅ
	'\u1e47': "ndotbelow",           // ṇ
	'\u306d': "nehiragana",          // ね
	'\u30cd': "nekatakana",          // ネ
	'\uff88': "nekatakanahalfwidth", // ﾈ
	// '\u20aa':    "newsheqelsign", // ₪ -- duplicate
	'\u338b': "nfsquare",            // ㎋
	'\u0999': "ngabengali",          // ঙ
	'\u0919': "ngadeva",             // ङ
	'\u0a99': "ngagujarati",         // ઙ
	'\u0a19': "ngagurmukhi",         // ਙ
	'\u0e07': "ngonguthai",          // ง
	'\u3093': "nhiragana",           // ん
	'\u0272': "nhookleft",           // ɲ
	'\u0273': "nhookretroflex",      // ɳ
	'\u326f': "nieunacirclekorean",  // ㉯
	'\u320f': "nieunaparenkorean",   // ㈏
	'\u3135': "nieuncieuckorean",    // ㄵ
	'\u3261': "nieuncirclekorean",   // ㉡
	'\u3136': "nieunhieuhkorean",    // ㄶ
	'\u3134': "nieunkorean",         // ㄴ
	'\u3168': "nieunpansioskorean",  // ㅨ
	'\u3201': "nieunparenkorean",    // ㈁
	'\u3167': "nieunsioskorean",     // ㅧ
	'\u3166': "nieuntikeutkorean",   // ㅦ
	'\u306b': "nihiragana",          // に
	'\u30cb': "nikatakana",          // ニ
	'\uff86': "nikatakanahalfwidth", // ﾆ
	'\uf899': "nikhahitleftthai",
	'\u0e4d': "nikhahitthai", // ํ
	'9':      "nine",         // 9
	// '\u0669':    "ninearabic", // ٩ -- duplicate
	'\u09ef': "ninebengali",                // ৯
	'\u2468': "ninecircle",                 // ⑨
	'\u2792': "ninecircleinversesansserif", // ➒
	'\u096f': "ninedeva",                   // ९
	'\u0aef': "ninegujarati",               // ૯
	'\u0a6f': "ninegurmukhi",               // ੯
	// '\u0669':    "ninehackarabic", // ٩ -- duplicate
	'\u3029': "ninehangzhou",         // 〩
	'\u3228': "nineideographicparen", // ㈨
	'\u2089': "nineinferior",         // ₉
	'\uff19': "ninemonospace",        // ９
	'\uf739': "nineoldstyle",
	'\u247c': "nineparen",      // ⑼
	'\u2490': "nineperiod",     // ⒐
	'\u06f9': "ninepersian",    // ۹
	'\u2178': "nineroman",      // ⅸ
	'\u2079': "ninesuperior",   // ⁹
	'\u2472': "nineteencircle", // ⑲
	'\u2486': "nineteenparen",  // ⒆
	'\u249a': "nineteenperiod", // ⒚
	'\u0e59': "ninethai",       // ๙
	'\u01cc': "nj",             // ǌ
	// '\u045a':    "njecyrillic", // њ -- duplicate
	'\u30f3': "nkatakana",           // ン
	'\uff9d': "nkatakanahalfwidth",  // ﾝ
	'\u019e': "nlegrightlong",       // ƞ
	'\u1e49': "nlinebelow",          // ṉ
	'\uff4e': "nmonospace",          // ｎ
	'\u339a': "nmsquare",            // ㎚
	'\u09a3': "nnabengali",          // ণ
	'\u0923': "nnadeva",             // ण
	'\u0aa3': "nnagujarati",         // ણ
	'\u0a23': "nnagurmukhi",         // ਣ
	'\u0929': "nnnadeva",            // ऩ
	'\u306e': "nohiragana",          // の
	'\u30ce': "nokatakana",          // ノ
	'\uff89': "nokatakanahalfwidth", // ﾉ
	// '\u00a0':    "nonbreakingspace",  -- duplicate
	'\u0e13': "nonenthai", // ณ
	'\u0e19': "nonuthai",  // น
	// '\u0646':    "noonarabic", // ن -- duplicate
	'\ufee6': "noonfinalarabic", // ﻦ
	// '\u06ba':    "noonghunnaarabic", // ں -- duplicate
	'\ufb9f': "noonghunnafinalarabic", // ﮟ
	'\ufee7': "noonhehinitialarabic",  // ﻧ
	// '\ufee7':    "nooninitialarabic", // ﻧ -- duplicate
	'\ufcd2': "noonjeeminitialarabic",  // ﳒ
	'\ufc4b': "noonjeemisolatedarabic", // ﱋ
	'\ufee8': "noonmedialarabic",       // ﻨ
	'\ufcd5': "noonmeeminitialarabic",  // ﳕ
	'\ufc4e': "noonmeemisolatedarabic", // ﱎ
	'\ufc8d': "noonnoonfinalarabic",    // ﲍ
	'\u220c': "notcontains",            // ∌
	'\u2209': "notelement",             // ∉
	// '\u2209':    "notelementof", // ∉ -- duplicate
	'\u2260': "notequal",              // ≠
	'\u226f': "notgreater",            // ≯
	'\u2271': "notgreaternorequal",    // ≱
	'\u2279': "notgreaternorless",     // ≹
	'\u2262': "notidentical",          // ≢
	'\u226e': "notless",               // ≮
	'\u2270': "notlessnorequal",       // ≰
	'\u2226': "notparallel",           // ∦
	'\u2280': "notprecedes",           // ⊀
	'\u2284': "notsubset",             // ⊄
	'\u2281': "notsucceeds",           // ⊁
	'\u2285': "notsuperset",           // ⊅
	'\u0576': "nowarmenian",           // ն
	'\u24a9': "nparen",                // ⒩
	'\u33b1': "nssquare",              // ㎱
	'\u207f': "nsuperior",             // ⁿ
	'\u00f1': "ntilde",                // ñ
	'\u03bd': "nu",                    // ν
	'\u306c': "nuhiragana",            // ぬ
	'\u30cc': "nukatakana",            // ヌ
	'\uff87': "nukatakanahalfwidth",   // ﾇ
	'\u09bc': "nuktabengali",          // ়
	'\u093c': "nuktadeva",             // ़
	'\u0abc': "nuktagujarati",         // ઼
	'\u0a3c': "nuktagurmukhi",         // ਼
	'#':      "numbersign",            // #
	'\uff03': "numbersignmonospace",   // ＃
	'\ufe5f': "numbersignsmall",       // ﹟
	'\u0374': "numeralsigngreek",      // ʹ
	'\u0375': "numeralsignlowergreek", // ͵
	// '\u2116':    "numero", // № -- duplicate
	// '\u05e0':    "nun", // נ -- duplicate
	'\ufb40': "nundagesh", // נּ
	// '\ufb40':    "nundageshhebrew", // נּ -- duplicate
	// '\u05e0':    "nunhebrew", // נ -- duplicate
	'\u33b5': "nvsquare",                 // ㎵
	'\u33bb': "nwsquare",                 // ㎻
	'\u099e': "nyabengali",               // ঞ
	'\u091e': "nyadeva",                  // ञ
	'\u0a9e': "nyagujarati",              // ઞ
	'\u0a1e': "nyagurmukhi",              // ਞ
	'o':      "o",                        // o
	'\u00f3': "oacute",                   // ó
	'\u0e2d': "oangthai",                 // อ
	'\u0275': "obarred",                  // ɵ
	'\u04e9': "obarredcyrillic",          // ө
	'\u04eb': "obarreddieresiscyrillic",  // ӫ
	'\u0993': "obengali",                 // ও
	'\u311b': "obopomofo",                // ㄛ
	'\u014f': "obreve",                   // ŏ
	'\u0911': "ocandradeva",              // ऑ
	'\u0a91': "ocandragujarati",          // ઑ
	'\u0949': "ocandravowelsigndeva",     // ॉ
	'\u0ac9': "ocandravowelsigngujarati", // ૉ
	'\u01d2': "ocaron",                   // ǒ
	'\u24de': "ocircle",                  // ⓞ
	'\u00f4': "ocircumflex",              // ô
	'\u1ed1': "ocircumflexacute",         // ố
	'\u1ed9': "ocircumflexdotbelow",      // ộ
	'\u1ed3': "ocircumflexgrave",         // ồ
	'\u1ed5': "ocircumflexhookabove",     // ổ
	'\u1ed7': "ocircumflextilde",         // ỗ
	// '\u043e':    "ocyrillic", // о -- duplicate
	'\u0151': "odblacute",         // ő
	'\u020d': "odblgrave",         // ȍ
	'\u0913': "odeva",             // ओ
	'\u00f6': "odieresis",         // ö
	'\u04e7': "odieresiscyrillic", // ӧ
	'\u1ecd': "odotbelow",         // ọ
	'\u0153': "oe",                // œ
	'\u315a': "oekorean",          // ㅚ
	'\u02db': "ogonek",            // ˛
	'\u0328': "ogonekcmb",         // ̨
	'\u00f2': "ograve",            // ò
	'\u0a93': "ogujarati",         // ઓ
	'\u0585': "oharmenian",        // օ
	'\u304a': "ohiragana",         // お
	'\u1ecf': "ohookabove",        // ỏ
	'\u01a1': "ohorn",             // ơ
	'\u1edb': "ohornacute",        // ớ
	'\u1ee3': "ohorndotbelow",     // ợ
	'\u1edd': "ohorngrave",        // ờ
	'\u1edf': "ohornhookabove",    // ở
	'\u1ee1': "ohorntilde",        // ỡ
	// '\u0151':    "ohungarumlaut", // ő -- duplicate
	'\u01a3': "oi",                 // ƣ
	'\u020f': "oinvertedbreve",     // ȏ
	'\u30aa': "okatakana",          // オ
	'\uff75': "okatakanahalfwidth", // ｵ
	'\u3157': "okorean",            // ㅗ
	'\u05ab': "olehebrew",          // ֫
	'\u014d': "omacron",            // ō
	'\u1e53': "omacronacute",       // ṓ
	'\u1e51': "omacrongrave",       // ṑ
	'\u0950': "omdeva",             // ॐ
	'\u03c9': "omega",              // ω
	'\u03d6': "omega1",             // ϖ
	'\u0461': "omegacyrillic",      // ѡ
	'\u0277': "omegalatinclosed",   // ɷ
	'\u047b': "omegaroundcyrillic", // ѻ
	'\u047d': "omegatitlocyrillic", // ѽ
	'\u03ce': "omegatonos",         // ώ
	'\u0ad0': "omgujarati",         // ૐ
	'\u03bf': "omicron",            // ο
	'\u03cc': "omicrontonos",       // ό
	'\uff4f': "omonospace",         // ｏ
	'1':      "one",                // 1
	// '\u0661':    "onearabic", // ١ -- duplicate
	'\u09e7': "onebengali",                // ১
	'\u2460': "onecircle",                 // ①
	'\u278a': "onecircleinversesansserif", // ➊
	'\u0967': "onedeva",                   // १
	'\u2024': "onedotenleader",            // ․
	'\u215b': "oneeighth",                 // ⅛
	'\uf6dc': "onefitted",
	'\u0ae7': "onegujarati", // ૧
	'\u0a67': "onegurmukhi", // ੧
	// '\u0661':    "onehackarabic", // ١ -- duplicate
	'\u00bd': "onehalf",             // ½
	'\u3021': "onehangzhou",         // 〡
	'\u3220': "oneideographicparen", // ㈠
	'\u2081': "oneinferior",         // ₁
	'\uff11': "onemonospace",        // １
	'\u09f4': "onenumeratorbengali", // ৴
	'\uf731': "oneoldstyle",
	'\u2474': "oneparen",                // ⑴
	'\u2488': "oneperiod",               // ⒈
	'\u06f1': "onepersian",              // ۱
	'\u00bc': "onequarter",              // ¼
	'\u2170': "oneroman",                // ⅰ
	'\u00b9': "onesuperior",             // ¹
	'\u0e51': "onethai",                 // ๑
	'\u2153': "onethird",                // ⅓
	'\u01eb': "oogonek",                 // ǫ
	'\u01ed': "oogonekmacron",           // ǭ
	'\u0a13': "oogurmukhi",              // ਓ
	'\u0a4b': "oomatragurmukhi",         // ੋ
	'\u0254': "oopen",                   // ɔ
	'\u24aa': "oparen",                  // ⒪
	'\u25e6': "openbullet",              // ◦
	'\u2325': "option",                  // ⌥
	'\u00aa': "ordfeminine",             // ª
	'\u00ba': "ordmasculine",            // º
	'\u221f': "orthogonal",              // ∟
	'\u0912': "oshortdeva",              // ऒ
	'\u094a': "oshortvowelsigndeva",     // ॊ
	'\u00f8': "oslash",                  // ø
	'\u01ff': "oslashacute",             // ǿ
	'\u3049': "osmallhiragana",          // ぉ
	'\u30a9': "osmallkatakana",          // ォ
	'\uff6b': "osmallkatakanahalfwidth", // ｫ
	// '\u01ff':    "ostrokeacute", // ǿ -- duplicate
	'\uf6f0': "osuperior",
	'\u047f': "otcyrillic",         // ѿ
	'\u00f5': "otilde",             // õ
	'\u1e4d': "otildeacute",        // ṍ
	'\u1e4f': "otildedieresis",     // ṏ
	'\u3121': "oubopomofo",         // ㄡ
	'\u203e': "overline",           // ‾
	'\ufe4a': "overlinecenterline", // ﹊
	'\u0305': "overlinecmb",        // ̅
	'\ufe49': "overlinedashed",     // ﹉
	'\ufe4c': "overlinedblwavy",    // ﹌
	'\ufe4b': "overlinewavy",       // ﹋
	// '\u00af':    "overscore", // ¯ -- duplicate
	'\u09cb': "ovowelsignbengali",         // ো
	'\u094b': "ovowelsigndeva",            // ो
	'\u0acb': "ovowelsigngujarati",        // ો
	'p':      "p",                         // p
	'\u3380': "paampssquare",              // ㎀
	'\u332b': "paasentosquare",            // ㌫
	'\u09aa': "pabengali",                 // প
	'\u1e55': "pacute",                    // ṕ
	'\u092a': "padeva",                    // प
	'\u21df': "pagedown",                  // ⇟
	'\u21de': "pageup",                    // ⇞
	'\u0aaa': "pagujarati",                // પ
	'\u0a2a': "pagurmukhi",                // ਪ
	'\u3071': "pahiragana",                // ぱ
	'\u0e2f': "paiyannoithai",             // ฯ
	'\u30d1': "pakatakana",                // パ
	'\u0484': "palatalizationcyrilliccmb", // ҄
	'\u04c0': "palochkacyrillic",          // Ӏ
	'\u317f': "pansioskorean",             // ㅿ
	'\u00b6': "paragraph",                 // ¶
	'\u2225': "parallel",                  // ∥
	'(':      "parenleft",                 // (
	'\ufd3e': "parenleftaltonearabic",     // ﴾
	'\uf8ed': "parenleftbt",
	'\uf8ec': "parenleftex",
	'\u208d': "parenleftinferior",  // ₍
	'\uff08': "parenleftmonospace", // （
	'\ufe59': "parenleftsmall",     // ﹙
	'\u207d': "parenleftsuperior",  // ⁽
	'\uf8eb': "parenlefttp",
	'\ufe35': "parenleftvertical",      // ︵
	')':      "parenright",             // )
	'\ufd3f': "parenrightaltonearabic", // ﴿
	'\uf8f8': "parenrightbt",
	'\uf8f7': "parenrightex",
	'\u208e': "parenrightinferior",  // ₎
	'\uff09': "parenrightmonospace", // ）
	'\ufe5a': "parenrightsmall",     // ﹚
	'\u207e': "parenrightsuperior",  // ⁾
	'\uf8f6': "parenrighttp",
	'\ufe36': "parenrightvertical", // ︶
	'\u2202': "partialdiff",        // ∂
	// '\u05c0':    "paseqhebrew", // ׀ -- duplicate
	'\u0599': "pashtahebrew", // ֙
	'\u33a9': "pasquare",     // ㎩
	// '\u05b7':    "patah", // ַ -- duplicate
	// '\u05b7':    "patah11", // ַ -- duplicate
	// '\u05b7':    "patah1d", // ַ -- duplicate
	// '\u05b7':    "patah2a", // ַ -- duplicate
	// '\u05b7':    "patahhebrew", // ַ -- duplicate
	// '\u05b7':    "patahnarrowhebrew", // ַ -- duplicate
	// '\u05b7':    "patahquarterhebrew", // ַ -- duplicate
	// '\u05b7':    "patahwidehebrew", // ַ -- duplicate
	'\u05a1': "pazerhebrew", // ֡
	'\u3106': "pbopomofo",   // ㄆ
	'\u24df': "pcircle",     // ⓟ
	'\u1e57': "pdotaccent",  // ṗ
	// '\u05e4':    "pe", // פ -- duplicate
	// '\u043f':    "pecyrillic", // п -- duplicate
	'\ufb44': "pedagesh", // פּ
	// '\ufb44':    "pedageshhebrew", // פּ -- duplicate
	'\u333b': "peezisquare",         // ㌻
	'\ufb43': "pefinaldageshhebrew", // ףּ
	// '\u067e':    "peharabic", // پ -- duplicate
	'\u057a': "peharmenian", // պ
	// '\u05e4':    "pehebrew", // פ -- duplicate
	'\ufb57': "pehfinalarabic",       // ﭗ
	'\ufb58': "pehinitialarabic",     // ﭘ
	'\u307a': "pehiragana",           // ぺ
	'\ufb59': "pehmedialarabic",      // ﭙ
	'\u30da': "pekatakana",           // ペ
	'\u04a7': "pemiddlehookcyrillic", // ҧ
	'\ufb4e': "perafehebrew",         // פֿ
	'%':      "percent",              // %
	// '\u066a':    "percentarabic", // ٪ -- duplicate
	'\uff05': "percentmonospace", // ％
	'\ufe6a': "percentsmall",     // ﹪
	'.':      "period",           // .
	'\u0589': "periodarmenian",   // ։
	// '\u00b7':    "periodcentered", // · -- duplicate
	'\uff61': "periodhalfwidth", // ｡
	'\uf6e7': "periodinferior",
	'\uff0e': "periodmonospace", // ．
	'\ufe52': "periodsmall",     // ﹒
	'\uf6e8': "periodsuperior",
	'\u0342': "perispomenigreekcmb",  // ͂
	'\u22a5': "perpendicular",        // ⊥
	'\u2030': "perthousand",          // ‰
	'\u20a7': "peseta",               // ₧
	'\u338a': "pfsquare",             // ㎊
	'\u09ab': "phabengali",           // ফ
	'\u092b': "phadeva",              // फ
	'\u0aab': "phagujarati",          // ફ
	'\u0a2b': "phagurmukhi",          // ਫ
	'\u03c6': "phi",                  // φ
	'\u03d5': "phi1",                 // ϕ
	'\u327a': "phieuphacirclekorean", // ㉺
	'\u321a': "phieuphaparenkorean",  // ㈚
	'\u326c': "phieuphcirclekorean",  // ㉬
	'\u314d': "phieuphkorean",        // ㅍ
	'\u320c': "phieuphparenkorean",   // ㈌
	'\u0278': "philatin",             // ɸ
	'\u0e3a': "phinthuthai",          // ฺ
	// '\u03d5':    "phisymbolgreek", // ϕ -- duplicate
	'\u01a5': "phook",                 // ƥ
	'\u0e1e': "phophanthai",           // พ
	'\u0e1c': "phophungthai",          // ผ
	'\u0e20': "phosamphaothai",        // ภ
	'\u03c0': "pi",                    // π
	'\u3273': "pieupacirclekorean",    // ㉳
	'\u3213': "pieupaparenkorean",     // ㈓
	'\u3176': "pieupcieuckorean",      // ㅶ
	'\u3265': "pieupcirclekorean",     // ㉥
	'\u3172': "pieupkiyeokkorean",     // ㅲ
	'\u3142': "pieupkorean",           // ㅂ
	'\u3205': "pieupparenkorean",      // ㈅
	'\u3174': "pieupsioskiyeokkorean", // ㅴ
	'\u3144': "pieupsioskorean",       // ㅄ
	'\u3175': "pieupsiostikeutkorean", // ㅵ
	'\u3177': "pieupthieuthkorean",    // ㅷ
	'\u3173': "pieuptikeutkorean",     // ㅳ
	'\u3074': "pihiragana",            // ぴ
	'\u30d4': "pikatakana",            // ピ
	// '\u03d6':    "pisymbolgreek", // ϖ -- duplicate
	'\u0583': "piwrarmenian", // փ
	'+':      "plus",         // +
	'\u031f': "plusbelowcmb", // ̟
	// '\u2295':    "pluscircle", // ⊕ -- duplicate
	'\u00b1': "plusminus",                // ±
	'\u02d6': "plusmod",                  // ˖
	'\uff0b': "plusmonospace",            // ＋
	'\ufe62': "plussmall",                // ﹢
	'\u207a': "plussuperior",             // ⁺
	'\uff50': "pmonospace",               // ｐ
	'\u33d8': "pmsquare",                 // ㏘
	'\u307d': "pohiragana",               // ぽ
	'\u261f': "pointingindexdownwhite",   // ☟
	'\u261c': "pointingindexleftwhite",   // ☜
	'\u261e': "pointingindexrightwhite",  // ☞
	'\u261d': "pointingindexupwhite",     // ☝
	'\u30dd': "pokatakana",               // ポ
	'\u0e1b': "poplathai",                // ป
	'\u3012': "postalmark",               // 〒
	'\u3020': "postalmarkface",           // 〠
	'\u24ab': "pparen",                   // ⒫
	'\u227a': "precedes",                 // ≺
	'\u211e': "prescription",             // ℞
	'\u02b9': "primemod",                 // ʹ
	'\u2035': "primereversed",            // ‵
	'\u220f': "product",                  // ∏
	'\u2305': "projective",               // ⌅
	'\u30fc': "prolongedkana",            // ー
	'\u2318': "propellor",                // ⌘
	'\u2282': "propersubset",             // ⊂
	'\u2283': "propersuperset",           // ⊃
	'\u2237': "proportion",               // ∷
	'\u221d': "proportional",             // ∝
	'\u03c8': "psi",                      // ψ
	'\u0471': "psicyrillic",              // ѱ
	'\u0486': "psilipneumatacyrilliccmb", // ҆
	'\u33b0': "pssquare",                 // ㎰
	'\u3077': "puhiragana",               // ぷ
	'\u30d7': "pukatakana",               // プ
	'\u33b4': "pvsquare",                 // ㎴
	'\u33ba': "pwsquare",                 // ㎺
	'q':      "q",                        // q
	'\u0958': "qadeva",                   // क़
	'\u05a8': "qadmahebrew",              // ֨
	// '\u0642':    "qafarabic", // ق -- duplicate
	'\ufed6': "qaffinalarabic",   // ﻖ
	'\ufed7': "qafinitialarabic", // ﻗ
	'\ufed8': "qafmedialarabic",  // ﻘ
	// '\u05b8':    "qamats", // ָ -- duplicate
	// '\u05b8':    "qamats10", // ָ -- duplicate
	// '\u05b8':    "qamats1a", // ָ -- duplicate
	// '\u05b8':    "qamats1c", // ָ -- duplicate
	// '\u05b8':    "qamats27", // ָ -- duplicate
	// '\u05b8':    "qamats29", // ָ -- duplicate
	// '\u05b8':    "qamats33", // ָ -- duplicate
	// '\u05b8':    "qamatsde", // ָ -- duplicate
	// '\u05b8':    "qamatshebrew", // ָ -- duplicate
	// '\u05b8':    "qamatsnarrowhebrew", // ָ -- duplicate
	// '\u05b8':    "qamatsqatanhebrew", // ָ -- duplicate
	// '\u05b8':    "qamatsqatannarrowhebrew", // ָ -- duplicate
	// '\u05b8':    "qamatsqatanquarterhebrew", // ָ -- duplicate
	// '\u05b8':    "qamatsqatanwidehebrew", // ָ -- duplicate
	// '\u05b8':    "qamatsquarterhebrew", // ָ -- duplicate
	// '\u05b8':    "qamatswidehebrew", // ָ -- duplicate
	'\u059f': "qarneyparahebrew", // ֟
	'\u3111': "qbopomofo",        // ㄑ
	'\u24e0': "qcircle",          // ⓠ
	'\u02a0': "qhook",            // ʠ
	'\uff51': "qmonospace",       // ｑ
	// '\u05e7':    "qof", // ק -- duplicate
	'\ufb47': "qofdagesh", // קּ
	// '\ufb47':    "qofdageshhebrew", // קּ -- duplicate
	// '\u05e7':    "qofhatafpatah", // ק -- duplicate
	// '\u05e7':    "qofhatafpatahhebrew", // ק -- duplicate
	// '\u05e7':    "qofhatafsegol", // ק -- duplicate
	// '\u05e7':    "qofhatafsegolhebrew", // ק -- duplicate
	// '\u05e7':    "qofhebrew", // ק -- duplicate
	// '\u05e7':    "qofhiriq", // ק -- duplicate
	// '\u05e7':    "qofhiriqhebrew", // ק -- duplicate
	// '\u05e7':    "qofholam", // ק -- duplicate
	// '\u05e7':    "qofholamhebrew", // ק -- duplicate
	// '\u05e7':    "qofpatah", // ק -- duplicate
	// '\u05e7':    "qofpatahhebrew", // ק -- duplicate
	// '\u05e7':    "qofqamats", // ק -- duplicate
	// '\u05e7':    "qofqamatshebrew", // ק -- duplicate
	// '\u05e7':    "qofqubuts", // ק -- duplicate
	// '\u05e7':    "qofqubutshebrew", // ק -- duplicate
	// '\u05e7':    "qofsegol", // ק -- duplicate
	// '\u05e7':    "qofsegolhebrew", // ק -- duplicate
	// '\u05e7':    "qofsheva", // ק -- duplicate
	// '\u05e7':    "qofshevahebrew", // ק -- duplicate
	// '\u05e7':    "qoftsere", // ק -- duplicate
	// '\u05e7':    "qoftserehebrew", // ק -- duplicate
	'\u24ac': "qparen",      // ⒬
	'\u2669': "quarternote", // ♩
	// '\u05bb':    "qubuts", // ֻ -- duplicate
	// '\u05bb':    "qubuts18", // ֻ -- duplicate
	// '\u05bb':    "qubuts25", // ֻ -- duplicate
	// '\u05bb':    "qubuts31", // ֻ -- duplicate
	// '\u05bb':    "qubutshebrew", // ֻ -- duplicate
	// '\u05bb':    "qubutsnarrowhebrew", // ֻ -- duplicate
	// '\u05bb':    "qubutsquarterhebrew", // ֻ -- duplicate
	// '\u05bb':    "qubutswidehebrew", // ֻ -- duplicate
	'?': "question", // ?
	// '\u061f':    "questionarabic", // ؟ -- duplicate
	'\u055e': "questionarmenian", // ՞
	'\u00bf': "questiondown",     // ¿
	'\uf7bf': "questiondownsmall",
	'\u037e': "questiongreek",     // ;
	'\uff1f': "questionmonospace", // ？
	'\uf73f': "questionsmall",
	'"':      "quotedbl",              // "
	'\u201e': "quotedblbase",          // „
	'\u201c': "quotedblleft",          // “
	'\uff02': "quotedblmonospace",     // ＂
	'\u301e': "quotedblprime",         // 〞
	'\u301d': "quotedblprimereversed", // 〝
	'\u201d': "quotedblright",         // ”
	'\u2018': "quoteleft",             // ‘
	'\u201b': "quoteleftreversed",     // ‛
	// '\u201b':    "quotereversed", // ‛ -- duplicate
	'\u2019': "quoteright", // ’
	// '\u0149':    "quoterightn", // ŉ -- duplicate
	'\u201a': "quotesinglbase",       // ‚
	'\'':     "quotesingle",          // \'
	'\uff07': "quotesinglemonospace", // ＇
	'r':      "r",                    // r
	'\u057c': "raarmenian",           // ռ
	'\u09b0': "rabengali",            // র
	'\u0155': "racute",               // ŕ
	'\u0930': "radeva",               // र
	'\u221a': "radical",              // √
	'\uf8e5': "radicalex",
	'\u33ae': "radoverssquare",        // ㎮
	'\u33af': "radoverssquaredsquare", // ㎯
	'\u33ad': "radsquare",             // ㎭
	// '\u05bf':    "rafe", // ֿ -- duplicate
	// '\u05bf':    "rafehebrew", // ֿ -- duplicate
	'\u0ab0': "ragujarati",              // ર
	'\u0a30': "ragurmukhi",              // ਰ
	'\u3089': "rahiragana",              // ら
	'\u30e9': "rakatakana",              // ラ
	'\uff97': "rakatakanahalfwidth",     // ﾗ
	'\u09f1': "ralowerdiagonalbengali",  // ৱ
	'\u09f0': "ramiddlediagonalbengali", // ৰ
	'\u0264': "ramshorn",                // ɤ
	'\u2236': "ratio",                   // ∶
	'\u3116': "rbopomofo",               // ㄖ
	'\u0159': "rcaron",                  // ř
	'\u0157': "rcedilla",                // ŗ
	'\u24e1': "rcircle",                 // ⓡ
	// '\u0157':    "rcommaaccent", // ŗ -- duplicate
	'\u0211': "rdblgrave",       // ȑ
	'\u1e59': "rdotaccent",      // ṙ
	'\u1e5b': "rdotbelow",       // ṛ
	'\u1e5d': "rdotbelowmacron", // ṝ
	'\u203b': "referencemark",   // ※
	'\u2286': "reflexsubset",    // ⊆
	'\u2287': "reflexsuperset",  // ⊇
	'\u00ae': "registered",      // ®
	'\uf8e8': "registersans",
	'\uf6da': "registerserif",
	// '\u0631':    "reharabic", // ر -- duplicate
	'\u0580': "reharmenian",    // ր
	'\ufeae': "rehfinalarabic", // ﺮ
	'\u308c': "rehiragana",     // れ
	// '\u0631':    "rehyehaleflamarabic", // ر -- duplicate
	'\u30ec': "rekatakana",          // レ
	'\uff9a': "rekatakanahalfwidth", // ﾚ
	// '\u05e8':    "resh", // ר -- duplicate
	'\ufb48': "reshdageshhebrew", // רּ
	// '\u05e8':    "reshhatafpatah", // ר -- duplicate
	// '\u05e8':    "reshhatafpatahhebrew", // ר -- duplicate
	// '\u05e8':    "reshhatafsegol", // ר -- duplicate
	// '\u05e8':    "reshhatafsegolhebrew", // ר -- duplicate
	// '\u05e8':    "reshhebrew", // ר -- duplicate
	// '\u05e8':    "reshhiriq", // ר -- duplicate
	// '\u05e8':    "reshhiriqhebrew", // ר -- duplicate
	// '\u05e8':    "reshholam", // ר -- duplicate
	// '\u05e8':    "reshholamhebrew", // ר -- duplicate
	// '\u05e8':    "reshpatah", // ר -- duplicate
	// '\u05e8':    "reshpatahhebrew", // ר -- duplicate
	// '\u05e8':    "reshqamats", // ר -- duplicate
	// '\u05e8':    "reshqamatshebrew", // ר -- duplicate
	// '\u05e8':    "reshqubuts", // ר -- duplicate
	// '\u05e8':    "reshqubutshebrew", // ר -- duplicate
	// '\u05e8':    "reshsegol", // ר -- duplicate
	// '\u05e8':    "reshsegolhebrew", // ר -- duplicate
	// '\u05e8':    "reshsheva", // ר -- duplicate
	// '\u05e8':    "reshshevahebrew", // ר -- duplicate
	// '\u05e8':    "reshtsere", // ר -- duplicate
	// '\u05e8':    "reshtserehebrew", // ר -- duplicate
	'\u223d': "reversedtilde", // ∽
	'\u0597': "reviahebrew",   // ֗
	// '\u0597':    "reviamugrashhebrew", // ֗ -- duplicate
	// '\u2310':    "revlogicalnot", // ⌐ -- duplicate
	'\u027e': "rfishhook",              // ɾ
	'\u027f': "rfishhookreversed",      // ɿ
	'\u09dd': "rhabengali",             // ঢ়
	'\u095d': "rhadeva",                // ढ़
	'\u03c1': "rho",                    // ρ
	'\u027d': "rhook",                  // ɽ
	'\u027b': "rhookturned",            // ɻ
	'\u02b5': "rhookturnedsuperior",    // ʵ
	'\u03f1': "rhosymbolgreek",         // ϱ
	'\u02de': "rhotichookmod",          // ˞
	'\u3271': "rieulacirclekorean",     // ㉱
	'\u3211': "rieulaparenkorean",      // ㈑
	'\u3263': "rieulcirclekorean",      // ㉣
	'\u3140': "rieulhieuhkorean",       // ㅀ
	'\u313a': "rieulkiyeokkorean",      // ㄺ
	'\u3169': "rieulkiyeoksioskorean",  // ㅩ
	'\u3139': "rieulkorean",            // ㄹ
	'\u313b': "rieulmieumkorean",       // ㄻ
	'\u316c': "rieulpansioskorean",     // ㅬ
	'\u3203': "rieulparenkorean",       // ㈃
	'\u313f': "rieulphieuphkorean",     // ㄿ
	'\u313c': "rieulpieupkorean",       // ㄼ
	'\u316b': "rieulpieupsioskorean",   // ㅫ
	'\u313d': "rieulsioskorean",        // ㄽ
	'\u313e': "rieulthieuthkorean",     // ㄾ
	'\u316a': "rieultikeutkorean",      // ㅪ
	'\u316d': "rieulyeorinhieuhkorean", // ㅭ
	// '\u221f':    "rightangle", // ∟ -- duplicate
	'\u0319': "righttackbelowcmb",     // ̙
	'\u22bf': "righttriangle",         // ⊿
	'\u308a': "rihiragana",            // り
	'\u30ea': "rikatakana",            // リ
	'\uff98': "rikatakanahalfwidth",   // ﾘ
	'\u02da': "ring",                  // ˚
	'\u0325': "ringbelowcmb",          // ̥
	'\u030a': "ringcmb",               // ̊
	'\u02bf': "ringhalfleft",          // ʿ
	'\u0559': "ringhalfleftarmenian",  // ՙ
	'\u031c': "ringhalfleftbelowcmb",  // ̜
	'\u02d3': "ringhalfleftcentered",  // ˓
	'\u02be': "ringhalfright",         // ʾ
	'\u0339': "ringhalfrightbelowcmb", // ̹
	'\u02d2': "ringhalfrightcentered", // ˒
	'\u0213': "rinvertedbreve",        // ȓ
	'\u3351': "rittorusquare",         // ㍑
	'\u1e5f': "rlinebelow",            // ṟ
	'\u027c': "rlongleg",              // ɼ
	'\u027a': "rlonglegturned",        // ɺ
	'\uff52': "rmonospace",            // ｒ
	'\u308d': "rohiragana",            // ろ
	'\u30ed': "rokatakana",            // ロ
	'\uff9b': "rokatakanahalfwidth",   // ﾛ
	'\u0e23': "roruathai",             // ร
	'\u24ad': "rparen",                // ⒭
	'\u09dc': "rrabengali",            // ড়
	'\u0931': "rradeva",               // ऱ
	'\u0a5c': "rragurmukhi",           // ੜ
	// '\u0691':    "rreharabic", // ڑ -- duplicate
	'\ufb8d': "rrehfinalarabic",            // ﮍ
	'\u09e0': "rrvocalicbengali",           // ৠ
	'\u0960': "rrvocalicdeva",              // ॠ
	'\u0ae0': "rrvocalicgujarati",          // ૠ
	'\u09c4': "rrvocalicvowelsignbengali",  // ৄ
	'\u0944': "rrvocalicvowelsigndeva",     // ॄ
	'\u0ac4': "rrvocalicvowelsigngujarati", // ૄ
	'\uf6f1': "rsuperior",
	'\u2590': "rtblock",             // ▐
	'\u0279': "rturned",             // ɹ
	'\u02b4': "rturnedsuperior",     // ʴ
	'\u308b': "ruhiragana",          // る
	'\u30eb': "rukatakana",          // ル
	'\uff99': "rukatakanahalfwidth", // ﾙ
	'\u09f2': "rupeemarkbengali",    // ৲
	'\u09f3': "rupeesignbengali",    // ৳
	'\uf6dd': "rupiah",
	'\u0e24': "ruthai",                    // ฤ
	'\u098b': "rvocalicbengali",           // ঋ
	'\u090b': "rvocalicdeva",              // ऋ
	'\u0a8b': "rvocalicgujarati",          // ઋ
	'\u09c3': "rvocalicvowelsignbengali",  // ৃ
	'\u0943': "rvocalicvowelsigndeva",     // ृ
	'\u0ac3': "rvocalicvowelsigngujarati", // ૃ
	's':      "s",                         // s
	'\u09b8': "sabengali",                 // স
	'\u015b': "sacute",                    // ś
	'\u1e65': "sacutedotaccent",           // ṥ
	// '\u0635':    "sadarabic", // ص -- duplicate
	'\u0938': "sadeva",                          // स
	'\ufeba': "sadfinalarabic",                  // ﺺ
	'\ufebb': "sadinitialarabic",                // ﺻ
	'\ufebc': "sadmedialarabic",                 // ﺼ
	'\u0ab8': "sagujarati",                      // સ
	'\u0a38': "sagurmukhi",                      // ਸ
	'\u3055': "sahiragana",                      // さ
	'\u30b5': "sakatakana",                      // サ
	'\uff7b': "sakatakanahalfwidth",             // ｻ
	'\ufdfa': "sallallahoualayhewasallamarabic", // ﷺ
	// '\u05e1':    "samekh", // ס -- duplicate
	'\ufb41': "samekhdagesh", // סּ
	// '\ufb41':    "samekhdageshhebrew", // סּ -- duplicate
	// '\u05e1':    "samekhhebrew", // ס -- duplicate
	'\u0e32': "saraaathai",         // า
	'\u0e41': "saraaethai",         // แ
	'\u0e44': "saraaimaimalaithai", // ไ
	'\u0e43': "saraaimaimuanthai",  // ใ
	'\u0e33': "saraamthai",         // ำ
	'\u0e30': "saraathai",          // ะ
	'\u0e40': "saraethai",          // เ
	'\uf886': "saraiileftthai",
	'\u0e35': "saraiithai", // ี
	'\uf885': "saraileftthai",
	'\u0e34': "saraithai", // ิ
	'\u0e42': "saraothai", // โ
	'\uf888': "saraueeleftthai",
	'\u0e37': "saraueethai", // ื
	'\uf887': "saraueleftthai",
	'\u0e36': "sarauethai",      // ึ
	'\u0e38': "sarauthai",       // ุ
	'\u0e39': "sarauuthai",      // ู
	'\u3119': "sbopomofo",       // ㄙ
	'\u0161': "scaron",          // š
	'\u1e67': "scarondotaccent", // ṧ
	'\u015f': "scedilla",        // ş
	'\u0259': "schwa",           // ə
	// '\u04d9':    "schwacyrillic", // ә -- duplicate
	'\u04db': "schwadieresiscyrillic", // ӛ
	'\u025a': "schwahook",             // ɚ
	'\u24e2': "scircle",               // ⓢ
	'\u015d': "scircumflex",           // ŝ
	'\u0219': "scommaaccent",          // ș
	'\u1e61': "sdotaccent",            // ṡ
	'\u1e63': "sdotbelow",             // ṣ
	'\u1e69': "sdotbelowdotaccent",    // ṩ
	'\u033c': "seagullbelowcmb",       // ̼
	'\u2033': "second",                // ″
	'\u02ca': "secondtonechinese",     // ˊ
	'\u00a7': "section",               // §
	// '\u0633':    "seenarabic", // س -- duplicate
	'\ufeb2': "seenfinalarabic",   // ﺲ
	'\ufeb3': "seeninitialarabic", // ﺳ
	'\ufeb4': "seenmedialarabic",  // ﺴ
	// '\u05b6':    "segol", // ֶ -- duplicate
	// '\u05b6':    "segol13", // ֶ -- duplicate
	// '\u05b6':    "segol1f", // ֶ -- duplicate
	// '\u05b6':    "segol2c", // ֶ -- duplicate
	// '\u05b6':    "segolhebrew", // ֶ -- duplicate
	// '\u05b6':    "segolnarrowhebrew", // ֶ -- duplicate
	// '\u05b6':    "segolquarterhebrew", // ֶ -- duplicate
	'\u0592': "segoltahebrew", // ֒
	// '\u05b6':    "segolwidehebrew", // ֶ -- duplicate
	'\u057d': "seharmenian",         // ս
	'\u305b': "sehiragana",          // せ
	'\u30bb': "sekatakana",          // セ
	'\uff7e': "sekatakanahalfwidth", // ｾ
	';':      "semicolon",           // ;
	// '\u061b':    "semicolonarabic", // ؛ -- duplicate
	'\uff1b': "semicolonmonospace",          // ；
	'\ufe54': "semicolonsmall",              // ﹔
	'\u309c': "semivoicedmarkkana",          // ゜
	'\uff9f': "semivoicedmarkkanahalfwidth", // ﾟ
	'\u3322': "sentisquare",                 // ㌢
	'\u3323': "sentosquare",                 // ㌣
	'7':      "seven",                       // 7
	// '\u0667':    "sevenarabic", // ٧ -- duplicate
	'\u09ed': "sevenbengali",                // ৭
	'\u2466': "sevencircle",                 // ⑦
	'\u2790': "sevencircleinversesansserif", // ➐
	'\u096d': "sevendeva",                   // ७
	'\u215e': "seveneighths",                // ⅞
	'\u0aed': "sevengujarati",               // ૭
	'\u0a6d': "sevengurmukhi",               // ੭
	// '\u0667':    "sevenhackarabic", // ٧ -- duplicate
	'\u3027': "sevenhangzhou",         // 〧
	'\u3226': "sevenideographicparen", // ㈦
	'\u2087': "seveninferior",         // ₇
	'\uff17': "sevenmonospace",        // ７
	'\uf737': "sevenoldstyle",
	'\u247a': "sevenparen",      // ⑺
	'\u248e': "sevenperiod",     // ⒎
	'\u06f7': "sevenpersian",    // ۷
	'\u2176': "sevenroman",      // ⅶ
	'\u2077': "sevensuperior",   // ⁷
	'\u2470': "seventeencircle", // ⑰
	'\u2484': "seventeenparen",  // ⒄
	'\u2498': "seventeenperiod", // ⒘
	'\u0e57': "seventhai",       // ๗
	'\u00ad': "sfthyphen",
	'\u0577': "shaarmenian", // շ
	'\u09b6': "shabengali",  // শ
	// '\u0448':    "shacyrillic", // ш -- duplicate
	// '\u0651':    "shaddaarabic", // ّ -- duplicate
	'\ufc61': "shaddadammaarabic",    // ﱡ
	'\ufc5e': "shaddadammatanarabic", // ﱞ
	'\ufc60': "shaddafathaarabic",    // ﱠ
	// '\u0651':    "shaddafathatanarabic", // ّ -- duplicate
	'\ufc62': "shaddakasraarabic",    // ﱢ
	'\ufc5f': "shaddakasratanarabic", // ﱟ
	'\u2592': "shade",                // ▒
	// '\u2593':    "shadedark", // ▓ -- duplicate
	// '\u2591':    "shadelight", // ░ -- duplicate
	// '\u2592':    "shademedium", // ▒ -- duplicate
	'\u0936': "shadeva",          // श
	'\u0ab6': "shagujarati",      // શ
	'\u0a36': "shagurmukhi",      // ਸ਼
	'\u0593': "shalshelethebrew", // ֓
	'\u3115': "shbopomofo",       // ㄕ
	// '\u0449':    "shchacyrillic", // щ -- duplicate
	// '\u0634':    "sheenarabic", // ش -- duplicate
	'\ufeb6': "sheenfinalarabic",   // ﺶ
	'\ufeb7': "sheeninitialarabic", // ﺷ
	'\ufeb8': "sheenmedialarabic",  // ﺸ
	'\u03e3': "sheicoptic",         // ϣ
	// '\u20aa':    "sheqel", // ₪ -- duplicate
	// '\u20aa':    "sheqelhebrew", // ₪ -- duplicate
	// '\u05b0':    "sheva", // ְ -- duplicate
	// '\u05b0':    "sheva115", // ְ -- duplicate
	// '\u05b0':    "sheva15", // ְ -- duplicate
	// '\u05b0':    "sheva22", // ְ -- duplicate
	// '\u05b0':    "sheva2e", // ְ -- duplicate
	// '\u05b0':    "shevahebrew", // ְ -- duplicate
	// '\u05b0':    "shevanarrowhebrew", // ְ -- duplicate
	// '\u05b0':    "shevaquarterhebrew", // ְ -- duplicate
	// '\u05b0':    "shevawidehebrew", // ְ -- duplicate
	'\u04bb': "shhacyrillic", // һ
	'\u03ed': "shimacoptic",  // ϭ
	// '\u05e9':    "shin", // ש -- duplicate
	'\ufb49': "shindagesh", // שּ
	// '\ufb49':    "shindageshhebrew", // שּ -- duplicate
	'\ufb2c': "shindageshshindot", // שּׁ
	// '\ufb2c':    "shindageshshindothebrew", // שּׁ -- duplicate
	'\ufb2d': "shindageshsindot", // שּׂ
	// '\ufb2d':    "shindageshsindothebrew", // שּׂ -- duplicate
	// '\u05c1':    "shindothebrew", // ׁ -- duplicate
	// '\u05e9':    "shinhebrew", // ש -- duplicate
	// '\ufb2a':    "shinshindot", // שׁ -- duplicate
	// '\ufb2a':    "shinshindothebrew", // שׁ -- duplicate
	// '\ufb2b':    "shinsindot", // שׂ -- duplicate
	// '\ufb2b':    "shinsindothebrew", // שׂ -- duplicate
	'\u0282': "shook",  // ʂ
	'\u03c3': "sigma",  // σ
	'\u03c2': "sigma1", // ς
	// '\u03c2':    "sigmafinal", // ς -- duplicate
	'\u03f2': "sigmalunatesymbolgreek", // ϲ
	'\u3057': "sihiragana",             // し
	'\u30b7': "sikatakana",             // シ
	'\uff7c': "sikatakanahalfwidth",    // ｼ
	// '\u05bd':    "siluqhebrew", // ֽ -- duplicate
	// '\u05bd':    "siluqlefthebrew", // ֽ -- duplicate
	'\u223c': "similar", // ∼
	// '\u05c2':    "sindothebrew", // ׂ -- duplicate
	'\u3274': "siosacirclekorean", // ㉴
	'\u3214': "siosaparenkorean",  // ㈔
	'\u317e': "sioscieuckorean",   // ㅾ
	'\u3266': "sioscirclekorean",  // ㉦
	'\u317a': "sioskiyeokkorean",  // ㅺ
	'\u3145': "sioskorean",        // ㅅ
	'\u317b': "siosnieunkorean",   // ㅻ
	'\u3206': "siosparenkorean",   // ㈆
	'\u317d': "siospieupkorean",   // ㅽ
	'\u317c': "siostikeutkorean",  // ㅼ
	'6':      "six",               // 6
	// '\u0666':    "sixarabic", // ٦ -- duplicate
	'\u09ec': "sixbengali",                // ৬
	'\u2465': "sixcircle",                 // ⑥
	'\u278f': "sixcircleinversesansserif", // ➏
	'\u096c': "sixdeva",                   // ६
	'\u0aec': "sixgujarati",               // ૬
	'\u0a6c': "sixgurmukhi",               // ੬
	// '\u0666':    "sixhackarabic", // ٦ -- duplicate
	'\u3026': "sixhangzhou",         // 〦
	'\u3225': "sixideographicparen", // ㈥
	'\u2086': "sixinferior",         // ₆
	'\uff16': "sixmonospace",        // ６
	'\uf736': "sixoldstyle",
	'\u2479': "sixparen",                          // ⑹
	'\u248d': "sixperiod",                         // ⒍
	'\u06f6': "sixpersian",                        // ۶
	'\u2175': "sixroman",                          // ⅵ
	'\u2076': "sixsuperior",                       // ⁶
	'\u246f': "sixteencircle",                     // ⑯
	'\u09f9': "sixteencurrencydenominatorbengali", // ৹
	'\u2483': "sixteenparen",                      // ⒃
	'\u2497': "sixteenperiod",                     // ⒗
	'\u0e56': "sixthai",                           // ๖
	'/':      "slash",                             // /
	'\uff0f': "slashmonospace",                    // ／
	// '\u017f':    "slong", // ſ -- duplicate
	'\u1e9b': "slongdotaccent", // ẛ
	'\u263a': "smileface",      // ☺
	'\uff53': "smonospace",     // ｓ
	// '\u05c3':    "sofpasuqhebrew", // ׃ -- duplicate
	// '\u00ad':    "softhyphen",  -- duplicate
	// '\u044c':    "softsigncyrillic", // ь -- duplicate
	'\u305d': "sohiragana",             // そ
	'\u30bd': "sokatakana",             // ソ
	'\uff7f': "sokatakanahalfwidth",    // ｿ
	'\u0338': "soliduslongoverlaycmb",  // ̸
	'\u0337': "solidusshortoverlaycmb", // ̷
	'\u0e29': "sorusithai",             // ษ
	'\u0e28': "sosalathai",             // ศ
	'\u0e0b': "sosothai",               // ซ
	'\u0e2a': "sosuathai",              // ส
	' ':      "space",                  //
	// ' ': "spacehackarabic", //   -- duplicate
	'\u2660': "spade", // ♠
	// '\u2660':    "spadesuitblack", // ♠ -- duplicate
	'\u2664': "spadesuitwhite",                  // ♤
	'\u24ae': "sparen",                          // ⒮
	'\u033b': "squarebelowcmb",                  // ̻
	'\u33c4': "squarecc",                        // ㏄
	'\u339d': "squarecm",                        // ㎝
	'\u25a9': "squarediagonalcrosshatchfill",    // ▩
	'\u25a4': "squarehorizontalfill",            // ▤
	'\u338f': "squarekg",                        // ㎏
	'\u339e': "squarekm",                        // ㎞
	'\u33ce': "squarekmcapital",                 // ㏎
	'\u33d1': "squareln",                        // ㏑
	'\u33d2': "squarelog",                       // ㏒
	'\u338e': "squaremg",                        // ㎎
	'\u33d5': "squaremil",                       // ㏕
	'\u339c': "squaremm",                        // ㎜
	'\u33a1': "squaremsquared",                  // ㎡
	'\u25a6': "squareorthogonalcrosshatchfill",  // ▦
	'\u25a7': "squareupperlefttolowerrightfill", // ▧
	'\u25a8': "squareupperrighttolowerleftfill", // ▨
	'\u25a5': "squareverticalfill",              // ▥
	'\u25a3': "squarewhitewithsmallblack",       // ▣
	'\u33db': "srsquare",                        // ㏛
	'\u09b7': "ssabengali",                      // ষ
	'\u0937': "ssadeva",                         // ष
	'\u0ab7': "ssagujarati",                     // ષ
	'\u3149': "ssangcieuckorean",                // ㅉ
	'\u3185': "ssanghieuhkorean",                // ㆅ
	'\u3180': "ssangieungkorean",                // ㆀ
	'\u3132': "ssangkiyeokkorean",               // ㄲ
	'\u3165': "ssangnieunkorean",                // ㅥ
	'\u3143': "ssangpieupkorean",                // ㅃ
	'\u3146': "ssangsioskorean",                 // ㅆ
	'\u3138': "ssangtikeutkorean",               // ㄸ
	'\uf6f2': "ssuperior",
	'\u00a3': "sterling",              // £
	'\uffe1': "sterlingmonospace",     // ￡
	'\u0336': "strokelongoverlaycmb",  // ̶
	'\u0335': "strokeshortoverlaycmb", // ̵
	// '\u2282':    "subset", // ⊂ -- duplicate
	'\u228a': "subsetnotequal", // ⊊
	// '\u2286':    "subsetorequal", // ⊆ -- duplicate
	'\u227b': "succeeds",            // ≻
	'\u220b': "suchthat",            // ∋
	'\u3059': "suhiragana",          // す
	'\u30b9': "sukatakana",          // ス
	'\uff7d': "sukatakanahalfwidth", // ｽ
	// '\u0652':    "sukunarabic", // ْ -- duplicate
	'\u2211': "summation", // ∑
	// '\u263c':    "sun", // ☼ -- duplicate
	// '\u2283':    "superset", // ⊃ -- duplicate
	'\u228b': "supersetnotequal", // ⊋
	// '\u2287':    "supersetorequal", // ⊇ -- duplicate
	'\u33dc': "svsquare",        // ㏜
	'\u337c': "syouwaerasquare", // ㍼
	't':      "t",               // t
	'\u09a4': "tabengali",       // ত
	'\u22a4': "tackdown",        // ⊤
	'\u22a3': "tackleft",        // ⊣
	'\u0924': "tadeva",          // त
	'\u0aa4': "tagujarati",      // ત
	'\u0a24': "tagurmukhi",      // ਤ
	// '\u0637':    "taharabic", // ط -- duplicate
	'\ufec2': "tahfinalarabic",      // ﻂ
	'\ufec3': "tahinitialarabic",    // ﻃ
	'\u305f': "tahiragana",          // た
	'\ufec4': "tahmedialarabic",     // ﻄ
	'\u337d': "taisyouerasquare",    // ㍽
	'\u30bf': "takatakana",          // タ
	'\uff80': "takatakanahalfwidth", // ﾀ
	// '\u0640':    "tatweelarabic", // ـ -- duplicate
	'\u03c4': "tau", // τ
	// '\u05ea':    "tav", // ת -- duplicate
	'\ufb4a': "tavdages", // תּ
	// '\ufb4a':    "tavdagesh", // תּ -- duplicate
	// '\ufb4a':    "tavdageshhebrew", // תּ -- duplicate
	// '\u05ea':    "tavhebrew", // ת -- duplicate
	'\u0167': "tbar",      // ŧ
	'\u310a': "tbopomofo", // ㄊ
	'\u0165': "tcaron",    // ť
	'\u02a8': "tccurl",    // ʨ
	'\u0163': "tcedilla",  // ţ
	// '\u0686':    "tcheharabic", // چ -- duplicate
	'\ufb7b': "tchehfinalarabic",   // ﭻ
	'\ufb7c': "tchehinitialarabic", // ﭼ
	'\ufb7d': "tchehmedialarabic",  // ﭽ
	// '\ufb7c':    "tchehmeeminitialarabic", // ﭼ -- duplicate
	'\u24e3': "tcircle",          // ⓣ
	'\u1e71': "tcircumflexbelow", // ṱ
	// '\u0163':    "tcommaaccent", // ţ -- duplicate
	'\u1e97': "tdieresis",  // ẗ
	'\u1e6b': "tdotaccent", // ṫ
	'\u1e6d': "tdotbelow",  // ṭ
	// '\u0442':    "tecyrillic", // т -- duplicate
	'\u04ad': "tedescendercyrillic", // ҭ
	// '\u062a':    "teharabic", // ت -- duplicate
	'\ufe96': "tehfinalarabic",        // ﺖ
	'\ufca2': "tehhahinitialarabic",   // ﲢ
	'\ufc0c': "tehhahisolatedarabic",  // ﰌ
	'\ufe97': "tehinitialarabic",      // ﺗ
	'\u3066': "tehiragana",            // て
	'\ufca1': "tehjeeminitialarabic",  // ﲡ
	'\ufc0b': "tehjeemisolatedarabic", // ﰋ
	// '\u0629':    "tehmarbutaarabic", // ة -- duplicate
	'\ufe94': "tehmarbutafinalarabic", // ﺔ
	'\ufe98': "tehmedialarabic",       // ﺘ
	'\ufca4': "tehmeeminitialarabic",  // ﲤ
	'\ufc0e': "tehmeemisolatedarabic", // ﰎ
	'\ufc73': "tehnoonfinalarabic",    // ﱳ
	'\u30c6': "tekatakana",            // テ
	'\uff83': "tekatakanahalfwidth",   // ﾃ
	'\u2121': "telephone",             // ℡
	'\u260e': "telephoneblack",        // ☎
	'\u05a0': "telishagedolahebrew",   // ֠
	'\u05a9': "telishaqetanahebrew",   // ֩
	'\u2469': "tencircle",             // ⑩
	'\u3229': "tenideographicparen",   // ㈩
	'\u247d': "tenparen",              // ⑽
	'\u2491': "tenperiod",             // ⒑
	'\u2179': "tenroman",              // ⅹ
	'\u02a7': "tesh",                  // ʧ
	// '\u05d8':    "tet", // ט -- duplicate
	'\ufb38': "tetdagesh", // טּ
	// '\ufb38':    "tetdageshhebrew", // טּ -- duplicate
	// '\u05d8':    "tethebrew", // ט -- duplicate
	'\u04b5': "tetsecyrillic", // ҵ
	'\u059b': "tevirhebrew",   // ֛
	// '\u059b':    "tevirlefthebrew", // ֛ -- duplicate
	'\u09a5': "thabengali",  // থ
	'\u0925': "thadeva",     // थ
	'\u0aa5': "thagujarati", // થ
	'\u0a25': "thagurmukhi", // ਥ
	// '\u0630':    "thalarabic", // ذ -- duplicate
	'\ufeac': "thalfinalarabic", // ﺬ
	'\uf898': "thanthakhatlowleftthai",
	'\uf897': "thanthakhatlowrightthai",
	'\u0e4c': "thanthakhatthai", // ์
	'\uf896': "thanthakhatupperleftthai",
	// '\u062b':    "theharabic", // ث -- duplicate
	'\ufe9a': "thehfinalarabic",   // ﺚ
	'\ufe9b': "thehinitialarabic", // ﺛ
	'\ufe9c': "thehmedialarabic",  // ﺜ
	// '\u2203':    "thereexists", // ∃ -- duplicate
	'\u2234': "therefore", // ∴
	'\u03b8': "theta",     // θ
	'\u03d1': "theta1",    // ϑ
	// '\u03d1':    "thetasymbolgreek", // ϑ -- duplicate
	'\u3279': "thieuthacirclekorean",     // ㉹
	'\u3219': "thieuthaparenkorean",      // ㈙
	'\u326b': "thieuthcirclekorean",      // ㉫
	'\u314c': "thieuthkorean",            // ㅌ
	'\u320b': "thieuthparenkorean",       // ㈋
	'\u246c': "thirteencircle",           // ⑬
	'\u2480': "thirteenparen",            // ⒀
	'\u2494': "thirteenperiod",           // ⒔
	'\u0e11': "thonangmonthothai",        // ฑ
	'\u01ad': "thook",                    // ƭ
	'\u0e12': "thophuthaothai",           // ฒ
	'\u00fe': "thorn",                    // þ
	'\u0e17': "thothahanthai",            // ท
	'\u0e10': "thothanthai",              // ฐ
	'\u0e18': "thothongthai",             // ธ
	'\u0e16': "thothungthai",             // ถ
	'\u0482': "thousandcyrillic",         // ҂
	'\u066c': "thousandsseparatorarabic", // ٬
	// '\u066c':    "thousandsseparatorpersian", // ٬ -- duplicate
	'3': "three", // 3
	// '\u0663':    "threearabic", // ٣ -- duplicate
	'\u09e9': "threebengali",                // ৩
	'\u2462': "threecircle",                 // ③
	'\u278c': "threecircleinversesansserif", // ➌
	'\u0969': "threedeva",                   // ३
	'\u215c': "threeeighths",                // ⅜
	'\u0ae9': "threegujarati",               // ૩
	'\u0a69': "threegurmukhi",               // ੩
	// '\u0663':    "threehackarabic", // ٣ -- duplicate
	'\u3023': "threehangzhou",         // 〣
	'\u3222': "threeideographicparen", // ㈢
	'\u2083': "threeinferior",         // ₃
	'\uff13': "threemonospace",        // ３
	'\u09f6': "threenumeratorbengali", // ৶
	'\uf733': "threeoldstyle",
	'\u2476': "threeparen",    // ⑶
	'\u248a': "threeperiod",   // ⒊
	'\u06f3': "threepersian",  // ۳
	'\u00be': "threequarters", // ¾
	'\uf6de': "threequartersemdash",
	'\u2172': "threeroman",          // ⅲ
	'\u00b3': "threesuperior",       // ³
	'\u0e53': "threethai",           // ๓
	'\u3394': "thzsquare",           // ㎔
	'\u3061': "tihiragana",          // ち
	'\u30c1': "tikatakana",          // チ
	'\uff81': "tikatakanahalfwidth", // ﾁ
	'\u3270': "tikeutacirclekorean", // ㉰
	'\u3210': "tikeutaparenkorean",  // ㈐
	'\u3262': "tikeutcirclekorean",  // ㉢
	'\u3137': "tikeutkorean",        // ㄷ
	'\u3202': "tikeutparenkorean",   // ㈂
	// '\u02dc':    "tilde", // ˜ -- duplicate
	'\u0330': "tildebelowcmb", // ̰
	'\u0303': "tildecmb",      // ̃
	// '\u0303':    "tildecomb", // ̃ -- duplicate
	'\u0360': "tildedoublecmb", // ͠
	// '\u223c':    "tildeoperator", // ∼ -- duplicate
	'\u0334': "tildeoverlaycmb",  // ̴
	'\u033e': "tildeverticalcmb", // ̾
	// '\u2297':    "timescircle", // ⊗ -- duplicate
	'\u0596': "tipehahebrew", // ֖
	// '\u0596':    "tipehalefthebrew", // ֖ -- duplicate
	'\u0a70': "tippigurmukhi",                     // ੰ
	'\u0483': "titlocyrilliccmb",                  // ҃
	'\u057f': "tiwnarmenian",                      // տ
	'\u1e6f': "tlinebelow",                        // ṯ
	'\uff54': "tmonospace",                        // ｔ
	'\u0569': "toarmenian",                        // թ
	'\u3068': "tohiragana",                        // と
	'\u30c8': "tokatakana",                        // ト
	'\uff84': "tokatakanahalfwidth",               // ﾄ
	'\u02e5': "tonebarextrahighmod",               // ˥
	'\u02e9': "tonebarextralowmod",                // ˩
	'\u02e6': "tonebarhighmod",                    // ˦
	'\u02e8': "tonebarlowmod",                     // ˨
	'\u02e7': "tonebarmidmod",                     // ˧
	'\u01bd': "tonefive",                          // ƽ
	'\u0185': "tonesix",                           // ƅ
	'\u01a8': "tonetwo",                           // ƨ
	'\u0384': "tonos",                             // ΄
	'\u3327': "tonsquare",                         // ㌧
	'\u0e0f': "topatakthai",                       // ฏ
	'\u3014': "tortoiseshellbracketleft",          // 〔
	'\ufe5d': "tortoiseshellbracketleftsmall",     // ﹝
	'\ufe39': "tortoiseshellbracketleftvertical",  // ︹
	'\u3015': "tortoiseshellbracketright",         // 〕
	'\ufe5e': "tortoiseshellbracketrightsmall",    // ﹞
	'\ufe3a': "tortoiseshellbracketrightvertical", // ︺
	'\u0e15': "totaothai",                         // ต
	'\u01ab': "tpalatalhook",                      // ƫ
	'\u24af': "tparen",                            // ⒯
	'\u2122': "trademark",                         // ™
	'\uf8ea': "trademarksans",
	'\uf6db': "trademarkserif",
	'\u0288': "tretroflexhook", // ʈ
	// '\u25bc':    "triagdn", // ▼ -- duplicate
	// '\u25c4':    "triaglf", // ◄ -- duplicate
	// '\u25ba':    "triagrt", // ► -- duplicate
	// '\u25b2':    "triagup", // ▲ -- duplicate
	'\u02a6': "ts", // ʦ
	// '\u05e6':    "tsadi", // צ -- duplicate
	'\ufb46': "tsadidagesh", // צּ
	// '\ufb46':    "tsadidageshhebrew", // צּ -- duplicate
	// '\u05e6':    "tsadihebrew", // צ -- duplicate
	// '\u0446':    "tsecyrillic", // ц -- duplicate
	// '\u05b5':    "tsere", // ֵ -- duplicate
	// '\u05b5':    "tsere12", // ֵ -- duplicate
	// '\u05b5':    "tsere1e", // ֵ -- duplicate
	// '\u05b5':    "tsere2b", // ֵ -- duplicate
	// '\u05b5':    "tserehebrew", // ֵ -- duplicate
	// '\u05b5':    "tserenarrowhebrew", // ֵ -- duplicate
	// '\u05b5':    "tserequarterhebrew", // ֵ -- duplicate
	// '\u05b5':    "tserewidehebrew", // ֵ -- duplicate
	// '\u045b':    "tshecyrillic", // ћ -- duplicate
	'\uf6f3': "tsuperior",
	'\u099f': "ttabengali",  // ট
	'\u091f': "ttadeva",     // ट
	'\u0a9f': "ttagujarati", // ટ
	'\u0a1f': "ttagurmukhi", // ਟ
	// '\u0679':    "tteharabic", // ٹ -- duplicate
	'\ufb67': "ttehfinalarabic",          // ﭧ
	'\ufb68': "ttehinitialarabic",        // ﭨ
	'\ufb69': "ttehmedialarabic",         // ﭩ
	'\u09a0': "tthabengali",              // ঠ
	'\u0920': "tthadeva",                 // ठ
	'\u0aa0': "tthagujarati",             // ઠ
	'\u0a20': "tthagurmukhi",             // ਠ
	'\u0287': "tturned",                  // ʇ
	'\u3064': "tuhiragana",               // つ
	'\u30c4': "tukatakana",               // ツ
	'\uff82': "tukatakanahalfwidth",      // ﾂ
	'\u3063': "tusmallhiragana",          // っ
	'\u30c3': "tusmallkatakana",          // ッ
	'\uff6f': "tusmallkatakanahalfwidth", // ｯ
	'\u246b': "twelvecircle",             // ⑫
	'\u247f': "twelveparen",              // ⑿
	'\u2493': "twelveperiod",             // ⒓
	'\u217b': "twelveroman",              // ⅻ
	'\u2473': "twentycircle",             // ⑳
	'\u5344': "twentyhangzhou",           // 卄
	'\u2487': "twentyparen",              // ⒇
	'\u249b': "twentyperiod",             // ⒛
	'2':      "two",                      // 2
	// '\u0662':    "twoarabic", // ٢ -- duplicate
	'\u09e8': "twobengali",                // ২
	'\u2461': "twocircle",                 // ②
	'\u278b': "twocircleinversesansserif", // ➋
	'\u0968': "twodeva",                   // २
	'\u2025': "twodotenleader",            // ‥
	// '\u2025':    "twodotleader", // ‥ -- duplicate
	'\ufe30': "twodotleadervertical", // ︰
	'\u0ae8': "twogujarati",          // ૨
	'\u0a68': "twogurmukhi",          // ੨
	// '\u0662':    "twohackarabic", // ٢ -- duplicate
	'\u3022': "twohangzhou",         // 〢
	'\u3221': "twoideographicparen", // ㈡
	'\u2082': "twoinferior",         // ₂
	'\uff12': "twomonospace",        // ２
	'\u09f5': "twonumeratorbengali", // ৵
	'\uf732': "twooldstyle",
	'\u2475': "twoparen",         // ⑵
	'\u2489': "twoperiod",        // ⒉
	'\u06f2': "twopersian",       // ۲
	'\u2171': "tworoman",         // ⅱ
	'\u01bb': "twostroke",        // ƻ
	'\u00b2': "twosuperior",      // ²
	'\u0e52': "twothai",          // ๒
	'\u2154': "twothirds",        // ⅔
	'u':      "u",                // u
	'\u00fa': "uacute",           // ú
	'\u0289': "ubar",             // ʉ
	'\u0989': "ubengali",         // উ
	'\u3128': "ubopomofo",        // ㄨ
	'\u016d': "ubreve",           // ŭ
	'\u01d4': "ucaron",           // ǔ
	'\u24e4': "ucircle",          // ⓤ
	'\u00fb': "ucircumflex",      // û
	'\u1e77': "ucircumflexbelow", // ṷ
	// '\u0443':    "ucyrillic", // у -- duplicate
	'\u0951': "udattadeva",        // ॑
	'\u0171': "udblacute",         // ű
	'\u0215': "udblgrave",         // ȕ
	'\u0909': "udeva",             // उ
	'\u00fc': "udieresis",         // ü
	'\u01d8': "udieresisacute",    // ǘ
	'\u1e73': "udieresisbelow",    // ṳ
	'\u01da': "udieresiscaron",    // ǚ
	'\u04f1': "udieresiscyrillic", // ӱ
	'\u01dc': "udieresisgrave",    // ǜ
	'\u01d6': "udieresismacron",   // ǖ
	'\u1ee5': "udotbelow",         // ụ
	'\u00f9': "ugrave",            // ù
	'\u0a89': "ugujarati",         // ઉ
	'\u0a09': "ugurmukhi",         // ਉ
	'\u3046': "uhiragana",         // う
	'\u1ee7': "uhookabove",        // ủ
	'\u01b0': "uhorn",             // ư
	'\u1ee9': "uhornacute",        // ứ
	'\u1ef1': "uhorndotbelow",     // ự
	'\u1eeb': "uhorngrave",        // ừ
	'\u1eed': "uhornhookabove",    // ử
	'\u1eef': "uhorntilde",        // ữ
	// '\u0171':    "uhungarumlaut", // ű -- duplicate
	'\u04f3': "uhungarumlautcyrillic", // ӳ
	'\u0217': "uinvertedbreve",        // ȗ
	'\u30a6': "ukatakana",             // ウ
	'\uff73': "ukatakanahalfwidth",    // ｳ
	'\u0479': "ukcyrillic",            // ѹ
	'\u315c': "ukorean",               // ㅜ
	'\u016b': "umacron",               // ū
	'\u04ef': "umacroncyrillic",       // ӯ
	'\u1e7b': "umacrondieresis",       // ṻ
	'\u0a41': "umatragurmukhi",        // ੁ
	'\uff55': "umonospace",            // ｕ
	'_':      "underscore",            // _
	// '\u2017':    "underscoredbl", // ‗ -- duplicate
	'\uff3f': "underscoremonospace", // ＿
	'\ufe33': "underscorevertical",  // ︳
	'\ufe4f': "underscorewavy",      // ﹏
	'\u222a': "union",               // ∪
	// '\u2200':    "universal", // ∀ -- duplicate
	'\u0173': "uogonek",              // ų
	'\u24b0': "uparen",               // ⒰
	'\u2580': "upblock",              // ▀
	'\u05c4': "upperdothebrew",       // ׄ
	'\u03c5': "upsilon",              // υ
	'\u03cb': "upsilondieresis",      // ϋ
	'\u03b0': "upsilondieresistonos", // ΰ
	'\u028a': "upsilonlatin",         // ʊ
	'\u03cd': "upsilontonos",         // ύ
	'\u031d': "uptackbelowcmb",       // ̝
	'\u02d4': "uptackmod",            // ˔
	'\u0a73': "uragurmukhi",          // ੳ
	'\u016f': "uring",                // ů
	// '\u045e':    "ushortcyrillic", // ў -- duplicate
	'\u3045': "usmallhiragana",          // ぅ
	'\u30a5': "usmallkatakana",          // ゥ
	'\uff69': "usmallkatakanahalfwidth", // ｩ
	'\u04af': "ustraightcyrillic",       // ү
	'\u04b1': "ustraightstrokecyrillic", // ұ
	'\u0169': "utilde",                  // ũ
	'\u1e79': "utildeacute",             // ṹ
	'\u1e75': "utildebelow",             // ṵ
	'\u098a': "uubengali",               // ঊ
	'\u090a': "uudeva",                  // ऊ
	'\u0a8a': "uugujarati",              // ઊ
	'\u0a0a': "uugurmukhi",              // ਊ
	'\u0a42': "uumatragurmukhi",         // ੂ
	'\u09c2': "uuvowelsignbengali",      // ূ
	'\u0942': "uuvowelsigndeva",         // ू
	'\u0ac2': "uuvowelsigngujarati",     // ૂ
	'\u09c1': "uvowelsignbengali",       // ু
	'\u0941': "uvowelsigndeva",          // ु
	'\u0ac1': "uvowelsigngujarati",      // ુ
	'v':      "v",                       // v
	'\u0935': "vadeva",                  // व
	'\u0ab5': "vagujarati",              // વ
	'\u0a35': "vagurmukhi",              // ਵ
	'\u30f7': "vakatakana",              // ヷ
	// '\u05d5':    "vav", // ו -- duplicate
	// '\ufb35':    "vavdagesh", // וּ -- duplicate
	// '\ufb35':    "vavdagesh65", // וּ -- duplicate
	// '\ufb35':    "vavdageshhebrew", // וּ -- duplicate
	// '\u05d5':    "vavhebrew", // ו -- duplicate
	// '\ufb4b':    "vavholam", // וֹ -- duplicate
	// '\ufb4b':    "vavholamhebrew", // וֹ -- duplicate
	// '\u05f0':    "vavvavhebrew", // װ -- duplicate
	// '\u05f1':    "vavyodhebrew", // ױ -- duplicate
	'\u24e5': "vcircle",   // ⓥ
	'\u1e7f': "vdotbelow", // ṿ
	// '\u0432':    "vecyrillic", // в -- duplicate
	// '\u06a4':    "veharabic", // ڤ -- duplicate
	'\ufb6b': "vehfinalarabic",   // ﭫ
	'\ufb6c': "vehinitialarabic", // ﭬ
	'\ufb6d': "vehmedialarabic",  // ﭭ
	'\u30f9': "vekatakana",       // ヹ
	// '\u2640':    "venus", // ♀ -- duplicate
	// '|': "verticalbar", // | -- duplicate
	'\u030d': "verticallineabovecmb",    // ̍
	'\u0329': "verticallinebelowcmb",    // ̩
	'\u02cc': "verticallinelowmod",      // ˌ
	'\u02c8': "verticallinemod",         // ˈ
	'\u057e': "vewarmenian",             // վ
	'\u028b': "vhook",                   // ʋ
	'\u30f8': "vikatakana",              // ヸ
	'\u09cd': "viramabengali",           // ্
	'\u094d': "viramadeva",              // ्
	'\u0acd': "viramagujarati",          // ્
	'\u0983': "visargabengali",          // ঃ
	'\u0903': "visargadeva",             // ः
	'\u0a83': "visargagujarati",         // ઃ
	'\uff56': "vmonospace",              // ｖ
	'\u0578': "voarmenian",              // ո
	'\u309e': "voicediterationhiragana", // ゞ
	'\u30fe': "voicediterationkatakana", // ヾ
	'\u309b': "voicedmarkkana",          // ゛
	'\uff9e': "voicedmarkkanahalfwidth", // ﾞ
	'\u30fa': "vokatakana",              // ヺ
	'\u24b1': "vparen",                  // ⒱
	'\u1e7d': "vtilde",                  // ṽ
	'\u028c': "vturned",                 // ʌ
	'\u3094': "vuhiragana",              // ゔ
	'\u30f4': "vukatakana",              // ヴ
	'w':      "w",                       // w
	'\u1e83': "wacute",                  // ẃ
	'\u3159': "waekorean",               // ㅙ
	'\u308f': "wahiragana",              // わ
	'\u30ef': "wakatakana",              // ワ
	'\uff9c': "wakatakanahalfwidth",     // ﾜ
	'\u3158': "wakorean",                // ㅘ
	'\u308e': "wasmallhiragana",         // ゎ
	'\u30ee': "wasmallkatakana",         // ヮ
	'\u3357': "wattosquare",             // ㍗
	'\u301c': "wavedash",                // 〜
	'\ufe34': "wavyunderscorevertical",  // ︴
	// '\u0648':    "wawarabic", // و -- duplicate
	'\ufeee': "wawfinalarabic", // ﻮ
	// '\u0624':    "wawhamzaabovearabic", // ؤ -- duplicate
	'\ufe86': "wawhamzaabovefinalarabic", // ﺆ
	'\u33dd': "wbsquare",                 // ㏝
	'\u24e6': "wcircle",                  // ⓦ
	'\u0175': "wcircumflex",              // ŵ
	'\u1e85': "wdieresis",                // ẅ
	'\u1e87': "wdotaccent",               // ẇ
	'\u1e89': "wdotbelow",                // ẉ
	'\u3091': "wehiragana",               // ゑ
	'\u2118': "weierstrass",              // ℘
	'\u30f1': "wekatakana",               // ヱ
	'\u315e': "wekorean",                 // ㅞ
	'\u315d': "weokorean",                // ㅝ
	'\u1e81': "wgrave",                   // ẁ
	// '\u25e6':    "whitebullet", // ◦ -- duplicate
	// '\u25cb':    "whitecircle", // ○ -- duplicate
	// '\u25d9':    "whitecircleinverse", // ◙ -- duplicate
	'\u300e': "whitecornerbracketleft",                  // 『
	'\ufe43': "whitecornerbracketleftvertical",          // ﹃
	'\u300f': "whitecornerbracketright",                 // 』
	'\ufe44': "whitecornerbracketrightvertical",         // ﹄
	'\u25c7': "whitediamond",                            // ◇
	'\u25c8': "whitediamondcontainingblacksmalldiamond", // ◈
	'\u25bf': "whitedownpointingsmalltriangle",          // ▿
	'\u25bd': "whitedownpointingtriangle",               // ▽
	'\u25c3': "whiteleftpointingsmalltriangle",          // ◃
	'\u25c1': "whiteleftpointingtriangle",               // ◁
	'\u3016': "whitelenticularbracketleft",              // 〖
	'\u3017': "whitelenticularbracketright",             // 〗
	'\u25b9': "whiterightpointingsmalltriangle",         // ▹
	'\u25b7': "whiterightpointingtriangle",              // ▷
	// '\u25ab':    "whitesmallsquare", // ▫ -- duplicate
	// '\u263a':    "whitesmilingface", // ☺ -- duplicate
	// '\u25a1':    "whitesquare", // □ -- duplicate
	'\u2606': "whitestar",                      // ☆
	'\u260f': "whitetelephone",                 // ☏
	'\u3018': "whitetortoiseshellbracketleft",  // 〘
	'\u3019': "whitetortoiseshellbracketright", // 〙
	'\u25b5': "whiteuppointingsmalltriangle",   // ▵
	'\u25b3': "whiteuppointingtriangle",        // △
	'\u3090': "wihiragana",                     // ゐ
	'\u30f0': "wikatakana",                     // ヰ
	'\u315f': "wikorean",                       // ㅟ
	'\uff57': "wmonospace",                     // ｗ
	'\u3092': "wohiragana",                     // を
	'\u30f2': "wokatakana",                     // ヲ
	'\uff66': "wokatakanahalfwidth",            // ｦ
	'\u20a9': "won",                            // ₩
	'\uffe6': "wonmonospace",                   // ￦
	'\u0e27': "wowaenthai",                     // ว
	'\u24b2': "wparen",                         // ⒲
	'\u1e98': "wring",                          // ẘ
	'\u02b7': "wsuperior",                      // ʷ
	'\u028d': "wturned",                        // ʍ
	'\u01bf': "wynn",                           // ƿ
	'x':      "x",                              // x
	'\u033d': "xabovecmb",                      // ̽
	'\u3112': "xbopomofo",                      // ㄒ
	'\u24e7': "xcircle",                        // ⓧ
	'\u1e8d': "xdieresis",                      // ẍ
	'\u1e8b': "xdotaccent",                     // ẋ
	'\u056d': "xeharmenian",                    // խ
	'\u03be': "xi",                             // ξ
	'\uff58': "xmonospace",                     // ｘ
	'\u24b3': "xparen",                         // ⒳
	'\u02e3': "xsuperior",                      // ˣ
	'y':      "y",                              // y
	'\u334e': "yaadosquare",                    // ㍎
	'\u09af': "yabengali",                      // য
	'\u00fd': "yacute",                         // ý
	'\u092f': "yadeva",                         // य
	'\u3152': "yaekorean",                      // ㅒ
	'\u0aaf': "yagujarati",                     // ય
	'\u0a2f': "yagurmukhi",                     // ਯ
	'\u3084': "yahiragana",                     // や
	'\u30e4': "yakatakana",                     // ヤ
	'\uff94': "yakatakanahalfwidth",            // ﾔ
	'\u3151': "yakorean",                       // ㅑ
	'\u0e4e': "yamakkanthai",                   // ๎
	'\u3083': "yasmallhiragana",                // ゃ
	'\u30e3': "yasmallkatakana",                // ャ
	'\uff6c': "yasmallkatakanahalfwidth",       // ｬ
	// '\u0463':    "yatcyrillic", // ѣ -- duplicate
	'\u24e8': "ycircle",     // ⓨ
	'\u0177': "ycircumflex", // ŷ
	'\u00ff': "ydieresis",   // ÿ
	'\u1e8f': "ydotaccent",  // ẏ
	'\u1ef5': "ydotbelow",   // ỵ
	// '\u064a':    "yeharabic", // ي -- duplicate
	// '\u06d2':    "yehbarreearabic", // ے -- duplicate
	'\ufbaf': "yehbarreefinalarabic", // ﮯ
	'\ufef2': "yehfinalarabic",       // ﻲ
	// '\u0626':    "yehhamzaabovearabic", // ئ -- duplicate
	'\ufe8a': "yehhamzaabovefinalarabic",   // ﺊ
	'\ufe8b': "yehhamzaaboveinitialarabic", // ﺋ
	'\ufe8c': "yehhamzaabovemedialarabic",  // ﺌ
	// '\ufef3':    "yehinitialarabic", // ﻳ -- duplicate
	// '\ufef4':    "yehmedialarabic", // ﻴ -- duplicate
	'\ufcdd': "yehmeeminitialarabic",    // ﳝ
	'\ufc58': "yehmeemisolatedarabic",   // ﱘ
	'\ufc94': "yehnoonfinalarabic",      // ﲔ
	'\u06d1': "yehthreedotsbelowarabic", // ۑ
	'\u3156': "yekorean",                // ㅖ
	'\u00a5': "yen",                     // ¥
	'\uffe5': "yenmonospace",            // ￥
	'\u3155': "yeokorean",               // ㅕ
	'\u3186': "yeorinhieuhkorean",       // ㆆ
	'\u05aa': "yerahbenyomohebrew",      // ֪
	// '\u05aa':    "yerahbenyomolefthebrew", // ֪ -- duplicate
	// '\u044b':    "yericyrillic", // ы -- duplicate
	'\u04f9': "yerudieresiscyrillic",  // ӹ
	'\u3181': "yesieungkorean",        // ㆁ
	'\u3183': "yesieungpansioskorean", // ㆃ
	'\u3182': "yesieungsioskorean",    // ㆂ
	'\u059a': "yetivhebrew",           // ֚
	'\u1ef3': "ygrave",                // ỳ
	'\u01b4': "yhook",                 // ƴ
	'\u1ef7': "yhookabove",            // ỷ
	'\u0575': "yiarmenian",            // յ
	// '\u0457':    "yicyrillic", // ї -- duplicate
	'\u3162': "yikorean",     // ㅢ
	'\u262f': "yinyang",      // ☯
	'\u0582': "yiwnarmenian", // ւ
	'\uff59': "ymonospace",   // ｙ
	// '\u05d9':    "yod", // י -- duplicate
	'\ufb39': "yoddagesh", // יּ
	// '\ufb39':    "yoddageshhebrew", // יּ -- duplicate
	// '\u05d9':    "yodhebrew", // י -- duplicate
	// '\u05f2':    "yodyodhebrew", // ײ -- duplicate
	// '\ufb1f':    "yodyodpatahhebrew", // ײַ -- duplicate
	'\u3088': "yohiragana",                // よ
	'\u3189': "yoikorean",                 // ㆉ
	'\u30e8': "yokatakana",                // ヨ
	'\uff96': "yokatakanahalfwidth",       // ﾖ
	'\u315b': "yokorean",                  // ㅛ
	'\u3087': "yosmallhiragana",           // ょ
	'\u30e7': "yosmallkatakana",           // ョ
	'\uff6e': "yosmallkatakanahalfwidth",  // ｮ
	'\u03f3': "yotgreek",                  // ϳ
	'\u3188': "yoyaekorean",               // ㆈ
	'\u3187': "yoyakorean",                // ㆇ
	'\u0e22': "yoyakthai",                 // ย
	'\u0e0d': "yoyingthai",                // ญ
	'\u24b4': "yparen",                    // ⒴
	'\u037a': "ypogegrammeni",             // ͺ
	'\u0345': "ypogegrammenigreekcmb",     // ͅ
	'\u01a6': "yr",                        // Ʀ
	'\u1e99': "yring",                     // ẙ
	'\u02b8': "ysuperior",                 // ʸ
	'\u1ef9': "ytilde",                    // ỹ
	'\u028e': "yturned",                   // ʎ
	'\u3086': "yuhiragana",                // ゆ
	'\u318c': "yuikorean",                 // ㆌ
	'\u30e6': "yukatakana",                // ユ
	'\uff95': "yukatakanahalfwidth",       // ﾕ
	'\u3160': "yukorean",                  // ㅠ
	'\u046b': "yusbigcyrillic",            // ѫ
	'\u046d': "yusbigiotifiedcyrillic",    // ѭ
	'\u0467': "yuslittlecyrillic",         // ѧ
	'\u0469': "yuslittleiotifiedcyrillic", // ѩ
	'\u3085': "yusmallhiragana",           // ゅ
	'\u30e5': "yusmallkatakana",           // ュ
	'\uff6d': "yusmallkatakanahalfwidth",  // ｭ
	'\u318b': "yuyekorean",                // ㆋ
	'\u318a': "yuyeokorean",               // ㆊ
	'\u09df': "yyabengali",                // য়
	'\u095f': "yyadeva",                   // य़
	'z':      "z",                         // z
	'\u0566': "zaarmenian",                // զ
	'\u017a': "zacute",                    // ź
	'\u095b': "zadeva",                    // ज़
	'\u0a5b': "zagurmukhi",                // ਜ਼
	// '\u0638':    "zaharabic", // ظ -- duplicate
	'\ufec6': "zahfinalarabic",   // ﻆ
	'\ufec7': "zahinitialarabic", // ﻇ
	'\u3056': "zahiragana",       // ざ
	'\ufec8': "zahmedialarabic",  // ﻈ
	// '\u0632':    "zainarabic", // ز -- duplicate
	'\ufeb0': "zainfinalarabic",  // ﺰ
	'\u30b6': "zakatakana",       // ザ
	'\u0595': "zaqefgadolhebrew", // ֕
	'\u0594': "zaqefqatanhebrew", // ֔
	'\u0598': "zarqahebrew",      // ֘
	// '\u05d6':    "zayin", // ז -- duplicate
	'\ufb36': "zayindagesh", // זּ
	// '\ufb36':    "zayindageshhebrew", // זּ -- duplicate
	// '\u05d6':    "zayinhebrew", // ז -- duplicate
	'\u3117': "zbopomofo",   // ㄗ
	'\u017e': "zcaron",      // ž
	'\u24e9': "zcircle",     // ⓩ
	'\u1e91': "zcircumflex", // ẑ
	'\u0291': "zcurl",       // ʑ
	'\u017c': "zdot",        // ż
	// '\u017c':    "zdotaccent", // ż -- duplicate
	'\u1e93': "zdotbelow", // ẓ
	// '\u0437':    "zecyrillic", // з -- duplicate
	'\u0499': "zedescendercyrillic", // ҙ
	'\u04df': "zedieresiscyrillic",  // ӟ
	'\u305c': "zehiragana",          // ぜ
	'\u30bc': "zekatakana",          // ゼ
	'0':      "zero",                // 0
	// '\u0660':    "zeroarabic", // ٠ -- duplicate
	'\u09e6': "zerobengali",  // ০
	'\u0966': "zerodeva",     // ०
	'\u0ae6': "zerogujarati", // ૦
	'\u0a66': "zerogurmukhi", // ੦
	// '\u0660':    "zerohackarabic", // ٠ -- duplicate
	'\u2080': "zeroinferior",  // ₀
	'\uff10': "zeromonospace", // ０
	'\uf730': "zerooldstyle",
	'\u06f0': "zeropersian",  // ۰
	'\u2070': "zerosuperior", // ⁰
	'\u0e50': "zerothai",     // ๐
	'\ufeff': "zerowidthjoiner",
	// '\u200c':    "zerowidthnonjoiner",  -- duplicate
	'\u200b': "zerowidthspace",
	'\u03b6': "zeta",             // ζ
	'\u3113': "zhbopomofo",       // ㄓ
	'\u056a': "zhearmenian",      // ժ
	'\u04c2': "zhebrevecyrillic", // ӂ
	// '\u0436':    "zhecyrillic", // ж -- duplicate
	'\u0497': "zhedescendercyrillic", // җ
	'\u04dd': "zhedieresiscyrillic",  // ӝ
	'\u3058': "zihiragana",           // じ
	'\u30b8': "zikatakana",           // ジ
	'\u05ae': "zinorhebrew",          // ֮
	'\u1e95': "zlinebelow",           // ẕ
	'\uff5a': "zmonospace",           // ｚ
	'\u305e': "zohiragana",           // ぞ
	'\u30be': "zokatakana",           // ゾ
	'\u24b5': "zparen",               // ⒵
	'\u0290': "zretroflexhook",       // ʐ
	'\u01b6': "zstroke",              // ƶ
	'\u305a': "zuhiragana",           // ず
	'\u30ba': "zukatakana",           // ズ
}

var texGlyphlistGlyphToStringMap = map[string]string{ // 285 entries
	"Dbar":                 "\u0110",     // Đ
	"Delta":                "\u2206",     // ∆
	"Digamma":              "\U0001d7cb", // 𝟋
	"FFIsmall":             "\uf766\uf766\uf769",
	"FFLsmall":             "\uf766\uf766\uf76c",
	"FFsmall":              "\uf766\uf766",
	"FIsmall":              "\uf766\uf769",
	"FLsmall":              "\uf766\uf76c",
	"Finv":                 "\u2132", // Ⅎ
	"Germandbls":           "SS",     // SS
	"Germandblssmall":      "\uf773\uf773",
	"Gmir":                 "\u2141", // ⅁
	"Ifractur":             "\u2111", // ℑ
	"Ng":                   "\u014a", // Ŋ
	"Omega":                "\u2126", // Ω
	"Omegainv":             "\u2127", // ℧
	"Rfractur":             "\u211c", // ℜ
	"SS":                   "SS",     // SS
	"SSsmall":              "\uf773\uf773",
	"Yen":                  "\u00a5", // ¥
	"altselector":          "\ufffd", // �
	"angbracketleft":       "\u27e8", // ⟨
	"angbracketright":      "\u27e9", // ⟩
	"anticlockwise":        "\u27f2", // ⟲
	"approxorequal":        "\u224a", // ≊
	"archleftdown":         "\u21b6", // ↶
	"archrightdown":        "\u21b7", // ↷
	"arrowbothv":           "\u2195", // ↕
	"arrowdblbothv":        "\u21d5", // ⇕
	"arrowleftbothalf":     "\u21bd", // ↽
	"arrowlefttophalf":     "\u21bc", // ↼
	"arrownortheast":       "\u2197", // ↗
	"arrownorthwest":       "\u2196", // ↖
	"arrowparrleftright":   "\u21c6", // ⇆
	"arrowparrrightleft":   "\u21c4", // ⇄
	"arrowrightbothalf":    "\u21c1", // ⇁
	"arrowrighttophalf":    "\u21c0", // ⇀
	"arrowsoutheast":       "\u2198", // ↘
	"arrowsouthwest":       "\u2199", // ↙
	"arrowtailleft":        "\u21a2", // ↢
	"arrowtailright":       "\u21a3", // ↣
	"arrowtripleleft":      "\u21da", // ⇚
	"arrowtripleright":     "\u21db", // ⇛
	"ascendercompwordmark": "\ufffd", // �
	"asteriskcentered":     "\u2217", // ∗
	"bardbl":               "\u2225", // ∥
	"beth":                 "\u2136", // ℶ
	"between":              "\u226c", // ≬
	"capitalcompwordmark":  "\ufffd", // �
	"ceilingleft":          "\u2308", // ⌈
	"ceilingright":         "\u2309", // ⌉
	"check":                "\u2713", // ✓
	"circleR":              "\u00ae", // ®
	"circleS":              "\u24c8", // Ⓢ
	"circleasterisk":       "\u229b", // ⊛
	"circlecopyrt":         "\u20dd", // ⃝
	"circledivide":         "\u2298", // ⊘
	"circledot":            "\u2299", // ⊙
	"circleequal":          "\u229c", // ⊜
	"circleminus":          "\u2296", // ⊖
	"circlering":           "\u229a", // ⊚
	"clockwise":            "\u27f3", // ⟳
	"complement":           "\u2201", // ∁
	"compwordmark":         "\u200c",
	"coproduct":            "\u2a3f", // ⨿
	"ct":                   "ct",     // ct
	"curlyleft":            "\u21ab", // ↫
	"curlyright":           "\u21ac", // ↬
	"cwm":                  "\u200c",
	"daleth":               "\u2138",       // ℸ
	"dbar":                 "\u0111",       // đ
	"dblarrowdwn":          "\u21ca",       // ⇊
	"dblarrowheadleft":     "\u219e",       // ↞
	"dblarrowheadright":    "\u21a0",       // ↠
	"dblarrowup":           "\u21c8",       // ⇈
	"dblbracketleft":       "\u27e6",       // ⟦
	"dblbracketright":      "\u27e7",       // ⟧
	"defines":              "\u225c",       // ≜
	"diamond":              "\u2662",       // ♢
	"diamondmath":          "\u22c4",       // ⋄
	"diamondsolid":         "\u2666",       // ♦
	"difference":           "\u224f",       // ≏
	"dividemultiply":       "\u22c7",       // ⋇
	"dotlessj":             "\u0237",       // ȷ
	"dotplus":              "\u2214",       // ∔
	"downfall":             "\u22ce",       // ⋎
	"downslope":            "\u29f9",       // ⧹
	"emptyset":             "\u2205",       // ∅
	"emptyslot":            "\ufffd",       // �
	"epsilon1":             "\u03f5",       // ϵ
	"epsiloninv":           "\u03f6",       // ϶
	"equaldotleftright":    "\u2252",       // ≒
	"equaldotrightleft":    "\u2253",       // ≓
	"equalorfollows":       "\u22df",       // ⋟
	"equalorgreater":       "\u2a96",       // ⪖
	"equalorless":          "\u2a95",       // ⪕
	"equalorprecedes":      "\u22de",       // ⋞
	"equalorsimilar":       "\u2242",       // ≂
	"equalsdots":           "\u2251",       // ≑
	"equivasymptotic":      "\u224d",       // ≍
	"flat":                 "\u266d",       // ♭
	"floorleft":            "\u230a",       // ⌊
	"floorright":           "\u230b",       // ⌋
	"follownotdbleqv":      "\u2aba",       // ⪺
	"follownotslnteql":     "\u2ab6",       // ⪶
	"followornoteqvlnt":    "\u22e9",       // ⋩
	"follows":              "\u227b",       // ≻
	"followsequal":         "\u2ab0",       // ⪰
	"followsorcurly":       "\u227d",       // ≽
	"followsorequal":       "\u227f",       // ≿
	"forces":               "\u22a9",       // ⊩
	"forcesbar":            "\u22aa",       // ⊪
	"fork":                 "\u22d4",       // ⋔
	"frown":                "\u2322",       // ⌢
	"geomequivalent":       "\u224e",       // ≎
	"greaterdbleqlless":    "\u2a8c",       // ⪌
	"greaterdblequal":      "\u2267",       // ≧
	"greaterdot":           "\u22d7",       // ⋗
	"greaterlessequal":     "\u22db",       // ⋛
	"greatermuch":          "\u226b",       // ≫
	"greaternotdblequal":   "\u2a8a",       // ⪊
	"greaternotequal":      "\u2a88",       // ⪈
	"greaterorapproxeql":   "\u2a86",       // ⪆
	"greaterorequalslant":  "\u2a7e",       // ⩾
	"greaterornotdbleql":   "\u2269",       // ≩
	"greaterornotequal":    "\u2269",       // ≩
	"greaterorsimilar":     "\u2273",       // ≳
	"harpoondownleft":      "\u21c3",       // ⇃
	"harpoondownright":     "\u21c2",       // ⇂
	"harpoonleftright":     "\u21cc",       // ⇌
	"harpoonrightleft":     "\u21cb",       // ⇋
	"harpoonupleft":        "\u21bf",       // ↿
	"harpoonupright":       "\u21be",       // ↾
	"heart":                "\u2661",       // ♡
	"hyphenchar":           "-",            // -
	"integerdivide":        "\u2216",       // ∖
	"intercal":             "\u22ba",       // ⊺
	"interrobang":          "\u203d",       // ‽
	"interrobangdown":      "\u2e18",       // ⸘
	"intersectiondbl":      "\u22d2",       // ⋒
	"intersectionsq":       "\u2293",       // ⊓
	"latticetop":           "\u22a4",       // ⊤
	"lessdbleqlgreater":    "\u2a8b",       // ⪋
	"lessdblequal":         "\u2266",       // ≦
	"lessdot":              "\u22d6",       // ⋖
	"lessequalgreater":     "\u22da",       // ⋚
	"lessmuch":             "\u226a",       // ≪
	"lessnotdblequal":      "\u2a89",       // ⪉
	"lessnotequal":         "\u2a87",       // ⪇
	"lessorapproxeql":      "\u2a85",       // ⪅
	"lessorequalslant":     "\u2a7d",       // ⩽
	"lessornotdbleql":      "\u2268",       // ≨
	"lessornotequal":       "\u2268",       // ≨
	"lessorsimilar":        "\u2272",       // ≲
	"longdbls":             "\u017f\u017f", // ſſ
	"longsh":               "\u017fh",      // ſh
	"longsi":               "\u017fi",      // ſi
	"longsl":               "\u017fl",      // ſl
	"longst":               "\ufb05",       // ﬅ
	"lscript":              "\u2113",       // ℓ
	"maltesecross":         "\u2720",       // ✠
	"measuredangle":        "\u2221",       // ∡
	"multicloseleft":       "\u22c9",       // ⋉
	"multicloseright":      "\u22ca",       // ⋊
	"multimap":             "\u22b8",       // ⊸
	"multiopenleft":        "\u22cb",       // ⋋
	"multiopenright":       "\u22cc",       // ⋌
	"nand":                 "\u22bc",       // ⊼
	"natural":              "\u266e",       // ♮
	"negationslash":        "\u0338",       // ̸
	"ng":                   "\u014b",       // ŋ
	"notapproxequal":       "\u2247",       // ≇
	"notarrowboth":         "\u21ae",       // ↮
	"notarrowleft":         "\u219a",       // ↚
	"notarrowright":        "\u219b",       // ↛
	"notbar":               "\u2224",       // ∤
	"notdblarrowboth":      "\u21ce",       // ⇎
	"notdblarrowleft":      "\u21cd",       // ⇍
	"notdblarrowright":     "\u21cf",       // ⇏
	"notexistential":       "\u2204",       // ∄
	"notfollows":           "\u2281",       // ⊁
	"notfollowsoreql":      "\u2ab0\u0338", // ⪰̸
	"notforces":            "\u22ae",       // ⊮
	"notforcesextra":       "\u22af",       // ⊯
	"notgreaterdblequal":   "\u2267\u0338", // ≧̸
	"notgreaterequal":      "\u2271",       // ≱
	"notgreaterorslnteql":  "\u2a7e\u0338", // ⩾̸
	"notlessdblequal":      "\u2266\u0338", // ≦̸
	"notlessequal":         "\u2270",       // ≰
	"notlessorslnteql":     "\u2a7d\u0338", // ⩽̸
	"notprecedesoreql":     "\u2aaf\u0338", // ⪯̸
	"notsatisfies":         "\u22ad",       // ⊭
	"notsimilar":           "\u2241",       // ≁
	"notsubseteql":         "\u2288",       // ⊈
	"notsubsetordbleql":    "\u2ac5\u0338", // ⫅̸
	"notsubsetoreql":       "\u228a",       // ⊊
	"notsuperseteql":       "\u2289",       // ⊉
	"notsupersetordbleql":  "\u2ac6\u0338", // ⫆̸
	"notsupersetoreql":     "\u228b",       // ⊋
	"nottriangeqlleft":     "\u22ec",       // ⋬
	"nottriangeqlright":    "\u22ed",       // ⋭
	"nottriangleleft":      "\u22ea",       // ⋪
	"nottriangleright":     "\u22eb",       // ⋫
	"notturnstile":         "\u22ac",       // ⊬
	"orunderscore":         "\u22bb",       // ⊻
	"owner":                "\u220b",       // ∋
	"perpcorrespond":       "\u2a5e",       // ⩞
	"pertenthousand":       "\u2031",       // ‱
	"phi":                  "\u03d5",       // ϕ
	"phi1":                 "\u03c6",       // φ
	"pi1":                  "\u03d6",       // ϖ
	"planckover2pi":        "\u210f",       // ℏ
	"planckover2pi1":       "\u210f",       // ℏ
	"precedenotdbleqv":     "\u2ab9",       // ⪹
	"precedenotslnteql":    "\u2ab5",       // ⪵
	"precedeornoteqvlnt":   "\u22e8",       // ⋨
	"precedesequal":        "\u2aaf",       // ⪯
	"precedesorcurly":      "\u227c",       // ≼
	"precedesorequal":      "\u227e",       // ≾
	"prime":                "\u2032",       // ′
	"primereverse":         "\u2035",       // ‵
	"punctdash":            "\u2014",       // —
	"rangedash":            "\u2013",       // –
	"revasymptequal":       "\u22cd",       // ⋍
	"revsimilar":           "\u223d",       // ∽
	"rho1":                 "\u03f1",       // ϱ
	"rightanglene":         "\u231d",       // ⌝
	"rightanglenw":         "\u231c",       // ⌜
	"rightanglese":         "\u231f",       // ⌟
	"rightanglesw":         "\u231e",       // ⌞
	"ringfitted":           "\ufffd",       // �
	"ringinequal":          "\u2256",       // ≖
	"satisfies":            "\u22a8",       // ⊨
	"sharp":                "\u266f",       // ♯
	"shiftleft":            "\u21b0",       // ↰
	"shiftright":           "\u21b1",       // ↱
	"similarequal":         "\u2243",       // ≃
	"slurabove":            "\u2322",       // ⌢
	"slurbelow":            "\u2323",       // ⌣
	"smile":                "\u2323",       // ⌣
	"sphericalangle":       "\u2222",       // ∢
	"square":               "\u25a1",       // □
	"squaredot":            "\u22a1",       // ⊡
	"squareimage":          "\u228f",       // ⊏
	"squareminus":          "\u229f",       // ⊟
	"squaremultiply":       "\u22a0",       // ⊠
	"squareoriginal":       "\u2290",       // ⊐
	"squareplus":           "\u229e",       // ⊞
	"squaresolid":          "\u25a0",       // ■
	"squiggleleftright":    "\u21ad",       // ↭
	"squiggleright":        "\u21dd",       // ⇝
	"st":                   "\ufb06",       // ﬆ
	"star":                 "\u22c6",       // ⋆
	"subsetdbl":            "\u22d0",       // ⋐
	"subsetdblequal":       "\u2ac5",       // ⫅
	"subsetnoteql":         "\u228a",       // ⊊
	"subsetornotdbleql":    "\u2acb",       // ⫋
	"subsetsqequal":        "\u2291",       // ⊑
	"supersetdbl":          "\u22d1",       // ⋑
	"supersetdblequal":     "\u2ac6",       // ⫆
	"supersetnoteql":       "\u228b",       // ⊋
	"supersetornotdbleql":  "\u2acc",       // ⫌
	"supersetsqequal":      "\u2292",       // ⊒
	"triangle":             "\u25b3",       // △
	"triangledownsld":      "\u25bc",       // ▼
	"triangleinv":          "\u25bd",       // ▽
	"triangleleft":         "\u25c1",       // ◁
	"triangleleftequal":    "\u22b4",       // ⊴
	"triangleleftsld":      "\u25c0",       // ◀
	"triangleright":        "\u25b7",       // ▷
	"trianglerightequal":   "\u22b5",       // ⊵
	"trianglerightsld":     "\u25b6",       // ▶
	"trianglesolid":        "\u25b2",       // ▲
	"turnstileleft":        "\u22a2",       // ⊢
	"turnstileright":       "\u22a3",       // ⊣
	"twelveudash":          "\ufffd",       // �
	"uniondbl":             "\u22d3",       // ⋓
	"unionmulti":           "\u228e",       // ⊎
	"unionsq":              "\u2294",       // ⊔
	"uprise":               "\u22cf",       // ⋏
	"upslope":              "\u29f8",       // ⧸
	"vector":               "\u20d7",       // ⃗
	"visiblespace":         "\u2423",       // ␣
	"visualspace":          "\u2423",       // ␣
	"wreathproduct":        "\u2240",       // ≀
}

var additionalGlyphlistGlyphToRuneMap = map[string]rune{ // 120 entries
	"angbracketleft":        '\u3008', // 〈
	"angbracketleftBig":     '\u2329', // 〈
	"angbracketleftBigg":    '\u2329', // 〈
	"angbracketleftbig":     '\u2329', // 〈
	"angbracketleftbigg":    '\u2329', // 〈
	"angbracketright":       '\u3009', // 〉
	"angbracketrightBig":    '\u232a', // 〉
	"angbracketrightBigg":   '\u232a', // 〉
	"angbracketrightbig":    '\u232a', // 〉
	"angbracketrightbigg":   '\u232a', // 〉
	"arrowhookleft":         '\u21aa', // ↪
	"arrowhookright":        '\u21a9', // ↩
	"arrowleftbothalf":      '\u21bd', // ↽
	"arrowlefttophalf":      '\u21bc', // ↼
	"arrownortheast":        '\u2197', // ↗
	"arrownorthwest":        '\u2196', // ↖
	"arrowrightbothalf":     '\u21c1', // ⇁
	"arrowrighttophalf":     '\u21c0', // ⇀
	"arrowsoutheast":        '\u2198', // ↘
	"arrowsouthwest":        '\u2199', // ↙
	"backslashBig":          '\u2216', // ∖
	"backslashBigg":         '\u2216', // ∖
	"backslashbig":          '\u2216', // ∖
	"backslashbigg":         '\u2216', // ∖
	"bardbl":                '\u2016', // ‖
	"bracehtipdownleft":     '\ufe37', // ︷
	"bracehtipdownright":    '\ufe37', // ︷
	"bracehtipupleft":       '\ufe38', // ︸
	"bracehtipupright":      '\ufe38', // ︸
	"braceleftBig":          '{',      // {
	"braceleftBigg":         '{',      // {
	"braceleftbig":          '{',      // {
	"braceleftbigg":         '{',      // {
	"bracerightBig":         '}',      // }
	"bracerightBigg":        '}',      // }
	"bracerightbig":         '}',      // }
	"bracerightbigg":        '}',      // }
	"bracketleftBig":        '[',      // [
	"bracketleftBigg":       '[',      // [
	"bracketleftbig":        '[',      // [
	"bracketleftbigg":       '[',      // [
	"bracketrightBig":       ']',      // ]
	"bracketrightBigg":      ']',      // ]
	"bracketrightbig":       ']',      // ]
	"bracketrightbigg":      ']',      // ]
	"ceilingleftBig":        '\u2308', // ⌈
	"ceilingleftBigg":       '\u2308', // ⌈
	"ceilingleftbig":        '\u2308', // ⌈
	"ceilingleftbigg":       '\u2308', // ⌈
	"ceilingrightBig":       '\u2309', // ⌉
	"ceilingrightBigg":      '\u2309', // ⌉
	"ceilingrightbig":       '\u2309', // ⌉
	"ceilingrightbigg":      '\u2309', // ⌉
	"circlecopyrt":          '\u00a9', // ©
	"circledotdisplay":      '\u2299', // ⊙
	"circledottext":         '\u2299', // ⊙
	"circlemultiplydisplay": '\u2297', // ⊗
	"circlemultiplytext":    '\u2297', // ⊗
	"circleplusdisplay":     '\u2295', // ⊕
	"circleplustext":        '\u2295', // ⊕
	"contintegraldisplay":   '\u222e', // ∮
	"contintegraltext":      '\u222e', // ∮
	"controlNULL":           '\x00',
	"coproductdisplay":      '\u2210', // ∐
	"coproducttext":         '\u2210', // ∐
	"floorleftBig":          '\u230a', // ⌊
	"floorleftBigg":         '\u230a', // ⌊
	"floorleftbig":          '\u230a', // ⌊
	"floorleftbigg":         '\u230a', // ⌊
	"floorrightBig":         '\u230b', // ⌋
	"floorrightBigg":        '\u230b', // ⌋
	"floorrightbig":         '\u230b', // ⌋
	"floorrightbigg":        '\u230b', // ⌋
	"hatwide":               '\u0302', // ̂
	"hatwider":              '\u0302', // ̂
	"hatwidest":             '\u0302', // ̂
	"integraldisplay":       '\u222b', // ∫
	"integraltext":          '\u222b', // ∫
	"intercal":              '\u1d40', // ᵀ
	"intersectiondisplay":   '\u22c2', // ⋂
	"intersectiontext":      '\u22c2', // ⋂
	"logicalanddisplay":     '\u2227', // ∧
	"logicalandtext":        '\u2227', // ∧
	"logicalordisplay":      '\u2228', // ∨
	"logicalortext":         '\u2228', // ∨
	"parenleftBig":          '(',      // (
	"parenleftBigg":         '(',      // (
	"parenleftbig":          '(',      // (
	"parenleftbigg":         '(',      // (
	"parenrightBig":         ')',      // )
	"parenrightBigg":        ')',      // )
	"parenrightbig":         ')',      // )
	"parenrightbigg":        ')',      // )
	"prime":                 '\u2032', // ′
	"productdisplay":        '\u220f', // ∏
	"producttext":           '\u220f', // ∏
	"radicalBig":            '\u221a', // √
	"radicalBigg":           '\u221a', // √
	"radicalbig":            '\u221a', // √
	"radicalbigg":           '\u221a', // √
	"radicalbt":             '\u221a', // √
	"radicaltp":             '\u221a', // √
	"radicalvertex":         '\u221a', // √
	"slashBig":              '/',      // /
	"slashBigg":             '/',      // /
	"slashbig":              '/',      // /
	"slashbigg":             '/',      // /
	"summationdisplay":      '\u2211', // ∑
	"summationtext":         '\u2211', // ∑
	"tildewide":             '\u02dc', // ˜
	"tildewider":            '\u02dc', // ˜
	"tildewidest":           '\u02dc', // ˜
	"uniondisplay":          '\u22c3', // ⋃
	"unionmultidisplay":     '\u228e', // ⊎
	"unionmultitext":        '\u228e', // ⊎
	"unionsqdisplay":        '\u2294', // ⊔
	"unionsqtext":           '\u2294', // ⊔
	"uniontext":             '\u22c3', // ⋃
	"vextenddouble":         '\u2225', // ∥
	"vextendsingle":         '\u2223', // ∣
}
