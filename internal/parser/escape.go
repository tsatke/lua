package parser

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"unicode"
)

var (
	escapeBytes [256]byte
)

func init() {
	escapeBytes['a'] = '\a'  // bell
	escapeBytes['b'] = '\b'  // backspace
	escapeBytes['f'] = '\f'  // form feed
	escapeBytes['n'] = '\n'  // newline
	escapeBytes['r'] = '\r'  // carriage return
	escapeBytes['t'] = '\t'  // horizontal tab
	escapeBytes['v'] = '\v'  // vertical tab
	escapeBytes['\\'] = '\\' // backslash
	escapeBytes['"'] = '"'   // double quote
	escapeBytes['\''] = '\'' // single quote
	escapeBytes['\n'] = '\n' // escaped newline is a new line
}

func unescape(s string) (string, error) {
	if !containsByte(s, '\\') {
		return s, nil
	}

	unescaped := make([]byte, 0, len(s))
	bytes := []byte(s)
	var escape bool
	var skipWhitespace bool

	var hexEscape bool
	var hexEscapeIndex int
	var hexEscapeBuffer [2]byte

	var decEscape bool
	var decEscapeIndex int
	var decEscapeBuffer [3]byte

bytes:
	for _, next := range bytes {
		if skipWhitespace {
			if unicode.IsSpace(rune(next)) {
				continue bytes
			} else {
				skipWhitespace = false
				escape = false
			}
		}

		if escape {
			if hexEscape {
				hexEscapeBuffer[hexEscapeIndex] = next
				hexEscapeIndex++
				if hexEscapeIndex == len(hexEscapeBuffer) {
					hexEscapeIndex = 0
					hexEscape = false
					escape = false
					decoded, err := hex.DecodeString(string(hexEscapeBuffer[:]))
					if err != nil {
						return "", fmt.Errorf("decode hex: %w", err)
					}
					unescaped = append(unescaped, decoded[0])
				}
				continue bytes
			} else if decEscape {
				isDigit := unicode.IsDigit(rune(next))
				if isDigit {
					decEscapeBuffer[decEscapeIndex] = next
					decEscapeIndex++
				}
				if !isDigit || decEscapeIndex == len(decEscapeBuffer) {
					decoded, err := strconv.Atoi(string(decEscapeBuffer[:decEscapeIndex]))
					if err != nil {
						return "", fmt.Errorf("decode dec: %w", err)
					}
					if decoded > 255 {
						return "", fmt.Errorf("decimal escape too large: %d (max 255)", decoded)
					}
					decEscapeIndex = 0
					decEscape = false
					escape = false
					unescaped = append(unescaped, byte(decoded))

					if !isDigit {
						if next == '\\' {
							escape = true
						} else {
							unescaped = append(unescaped, next)
						}
					}
				}
				continue bytes
			}

			switch next {
			case 'z':
				skipWhitespace = true
				continue bytes
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				decEscape = true
				decEscapeBuffer[0] = next
				decEscapeIndex = 1
				continue bytes
			case 'x':
				hexEscape = true
				continue bytes
			}
			if next == 'z' {
				skipWhitespace = true
				continue bytes
			}
			b := escapeBytes[next]
			if b == 0 {
				return "", fmt.Errorf("unknown escape sequence '\\%s'", string(next))
			}
			unescaped = append(unescaped, b)

			escape = false
		} else {
			if next == '\\' {
				escape = true
			} else {
				unescaped = append(unescaped, next)
			}
		}
	}

	// check for incomplete escapes
	if hexEscapeIndex != 0 {
		// index is reset at the end of escape, so if not reset,
		// the escape is incomplete
		return "", fmt.Errorf("incomplete hex escape at end of string")
	}
	if decEscapeIndex != 0 {
		// index is reset at the end of escape, so if not reset,
		// the escape is incomplete, but we can finish it
		decoded, err := strconv.Atoi(string(decEscapeBuffer[:decEscapeIndex]))
		if err != nil {
			return "", fmt.Errorf("decode dec: %w", err)
		}
		if decoded > 255 {
			return "", fmt.Errorf("decimal escape too large: %d (max 255)", decoded)
		}
		decEscapeIndex = 0
		decEscape = false
		escape = false
		unescaped = append(unescaped, byte(decoded))
	}
	if escape {
		return "", fmt.Errorf("unfinished escape at end of string")
	}

	return string(unescaped), nil
}

func containsByte(s string, b byte) bool {
	for _, sb := range []byte(s) {
		if sb == b {
			return true
		}
	}
	return false
}
