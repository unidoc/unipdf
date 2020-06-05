/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

// The follow constant configure debugging.
const (
	verbose      = false
	verboseGeom  = false
	verbosePage  = false
	verbosePara  = false
	verboseTable = false
)

// The following constants control the approaches used in the code.
const (
	useTables = true
	doHyphens = true
	useEBBox  = false
)

// The following constants are the tuning parameter for text extracton
const (
	// Size of depth bins in points
	depthBinPoints = 6

	// Variation in line depth as a fraction of font size. +lineDepthR for subscripts, -lineDepthR for
	// superscripts
	lineDepthR = 0.5

	// All constants that end in R are relative to font size.

	// Max difference in font sizes allowed within a word.
	maxIntraWordFontTolR = 0.05

	// Maximum gap between a word and a para in the depth direction for which we pull the word
	// into the para, as a fraction of the font size.
	maxIntraDepthGapR = 1.0
	// Max diffrence in font size for word and para for the above case
	maxIntraDepthFontTolR = 0.05

	// Maximum gap between a word and a para in the reading direction for which we pull the word
	// into the para.
	maxIntraReadingGapR = 0.4
	// Max diffrence in font size for word and para for the above case
	maxIntraReadingFontTol = 0.6

	// Minimum spacing between paras in the reading direction.
	minInterReadingGapR = 1.0
	// Max diffrence in font size for word and para for the above case
	minInterReadingFontTol = 0.1

	// Maximum inter-word spacing.
	maxIntraWordGapR = 1.4

	// Maximum overlap between characters allowd within a line
	maxIntraLineOverlapR = 0.46

	// Maximum spacing between characters within a line.
	maxIntraLineGapR = 0.03
)
