/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

// IsWhiteSpace checks if byte represents a white space character.
func IsWhiteSpace(ch byte) bool {
	// Table 1 white-space characters (7.2.2 Character Set)
	// spaceCharacters := string([]byte{0x00, 0x09, 0x0A, 0x0C, 0x0D, 0x20})
	if (ch == 0x00) || (ch == 0x09) || (ch == 0x0A) || (ch == 0x0C) || (ch == 0x0D) || (ch == 0x20) {
		return true
	}
	return false
}

// IsFloatDigit checks if a character can be a part of a float number string.
func IsFloatDigit(c byte) bool {
	return ('0' <= c && c <= '9') || c == '.'
}

// IsDecimalDigit checks if the character is a part of a decimal number string.
func IsDecimalDigit(c byte) bool {
	if c >= '0' && c <= '9' {
		return true
	}

	return false
}

// IsOctalDigit checks if a character can be part of an octal digit string.
func IsOctalDigit(c byte) bool {
	if c >= '0' && c <= '7' {
		return true
	}

	return false
}

// IsPrintable checks if a character is printable.
// Regular characters that are outside the range EXCLAMATION MARK(21h)
// (!) to TILDE (7Eh) (~) should be written using the hexadecimal notation.
func IsPrintable(char byte) bool {
	if char < 0x21 || char > 0x7E {
		return false
	}
	return true
}

// IsDelimiter checks if a character represents a delimiter.
func IsDelimiter(char byte) bool {
	if char == '(' || char == ')' {
		return true
	}
	if char == '<' || char == '>' {
		return true
	}
	if char == '[' || char == ']' {
		return true
	}
	if char == '{' || char == '}' {
		return true
	}
	if char == '/' {
		return true
	}
	if char == '%' {
		return true
	}

	return false
}
