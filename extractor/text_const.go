/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

// The follow constant configure debugging.
const (
	verbose         = false
	verboseGeom     = false
	verbosePage     = false
	verbosePara     = false
	verboseParaLine = verbosePara && false
	verboseParaWord = verboseParaLine && false
	verboseTable    = false
)

// The following constants control the approaches used in the code.
const (
	doHyphens           = true
	doRemoveDuplicates  = true
	doCombineDiacritics = true
	useEBBox            = false
)

// The following constants are the tuning parameter for text extracton
const (
	// Change in angle of text in degrees that we treat as a different orientatiom/
	orientationGranularity = 10
	// Size of depth bins in points
	depthBinPoints = 6

	// Variation in line depth as a fraction of font size. +lineDepthR for subscripts, -lineDepthR for
	// superscripts
	lineDepthR = 0.5

	// All constants that end in R are relative to font size.

	maxWordAdvanceR = 0.11

	maxKerningR = 0.19
	maxLeadingR = 0.04

	// Max difference in font sizes allowed within a word.
	maxIntraWordFontTolR = 0.04

	// Maximum gap between a word and a para in the depth direction for which we pull the word
	// into the para, as a fraction of the font size.
	maxIntraDepthGapR = 1.0
	// Max diffrence in font size for word and para for the above case
	maxIntraDepthFontTolR = 0.04

	// Maximum gap between a word and a para in the reading direction for which we pull the word
	// into the para.
	maxIntraReadingGapR = 0.4
	// Max diffrence in font size for word and para for the above case
	maxIntraReadingFontTol = 0.7

	// Minimum spacing between paras in the reading direction.
	minInterReadingGapR = 1.0
	// Max difference in font size for word and para for the above case
	minInterReadingFontTol = 0.1

	// Maximum inter-word spacing.
	maxIntraWordGapR = 1.4

	// Maximum overlap between characters allowd within a line
	maxIntraLineOverlapR = 0.46

	// Maximum spacing between characters within a line.
	maxIntraLineGapR = 0.02

	// Maximum difference in coordinates of duplicated textWords.
	maxDuplicateWordR = 0.2

	// Maximum distance from a character to its diacritic marks as a fraction of the character size.
	diacriticRadiusR = 0.5

	// Minimum number of rumes in the first half of a hyphenated word
	minHyphenation = 4

	// The distance we look down from the top of a wordBag for the leftmost word.
	topWordRangeR = 4.0

	// Minimum number of cells in a textTable
	minTableParas = 6
)
