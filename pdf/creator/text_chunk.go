/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// TextChunk represents a chunk of text along with a particular style.
type TextChunk struct {
	// The text that is being rendered in the PDF.
	Text string

	// The style of the text being rendered.
	Style TextStyle
}
