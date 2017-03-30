// API influenced by Rob Pike.
package ast

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const eof rune = -1

const (
	digits    = "0123456789"
	hexDigits = "0123456789abcdefABCDEF"
	letters   = "abcedfghijklmnopqrstuvwxyzABCEDFGHIJKLMNOPQRSTUVWXYZ"
	symbols   = "+-*/\\^~=<>!?@#$%&|:`."
	name      = letters + digits + symbols + "'_"
)

// stateFn represents the state of the lexer
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	r        *bufio.Reader
	state    stateFn // next lexing function to enter
	readErr  error   // last error set by read (non-nil iff it returned eof)
	token    []byte  // pending input
	runeSize int     // size of the last rune read (zero if readErr != nil)
	tokens   []token // tokens read so far
}

func newLexer(r io.Reader) *lexer {
	return &lexer{r: bufio.NewReader(r)}
}

func (l *lexer) lex() ([]token, error) {
	for l.state = lexAny; l.state != nil; {
		l.readErr = nil
		l.state = l.state(l)
		if l.readErr != nil {
			if l.readErr == io.EOF {
				break
			}
			return nil, l.readErr
		}
	}
	return l.tokens, nil
}

func lexAny(l *lexer) stateFn {
	l.discardRun("\t ")
	r := l.read()
	switch {
	case containsRune(letters, r):
		return lexAtom
	case r == '(':
		l.emit(LPAREN)
		return lexAny
	case r == ')':
		l.emit(RPAREN)
		return lexAny
	case r == '_':
		l.emit(UNDERSCORE)
		return lexAny
	case r == '=':
		l.emit(EQUAL)
		return lexAny
	case r == '/':
		l.emit(SLASH)
		return lexAny
	case r == '.':
		l.emit(DOT)
		return lexAny
	case r == '$':
		return lexName
	case r == '"':
		return lexString
	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.unread()
		return lexNumber
	case r == '\n':
		return lexAny // no-op
	case r == eof:
		return nil
	default:
		return l.errorf("unexpected character: %#U", r)
	}
}

func lexRightDelim(l *lexer) stateFn {
	if !l.accept(" \n\t()") && l.peek() != eof {
		return l.errorf("unexpected character %#U, expected one of %q", l.peek(), " \\n\\t()")
	}
	l.unread()
	return lexAny
}

// lexAtom scans an atom.
// The first character has been scanned.
func lexAtom(l *lexer) stateFn {
	l.acceptRun(letters + digits + "_")
	switch {
	case bytes.HasSuffix(l.token, []byte("_s")):
		tok := bytes.TrimSuffix(l.token, []byte("_s"))
		if typ, ok := atom[string(tok)]; ok {
			l.tokens = append(l.tokens, token{typ: typ, text: tok})
			l.tokens = append(l.tokens, token{typ: UNDERSCORE, text: []byte("_")})
			l.tokens = append(l.tokens, token{typ: S, text: []byte("s")})
			return lexAny
		}
	case bytes.HasSuffix(l.token, []byte("_u")):
		tok := bytes.TrimSuffix(l.token, []byte("_u"))
		if typ, ok := atom[string(tok)]; ok {
			l.tokens = append(l.tokens, token{typ: typ, text: tok})
			l.tokens = append(l.tokens, token{typ: UNDERSCORE, text: []byte("_")})
			l.tokens = append(l.tokens, token{typ: U, text: []byte("u")})
			return lexAny
		}
	default:
		if typ, ok := atom[string(l.token)]; ok {
			l.emit(typ)
			return lexAny
		}
	}
	return l.errorf("unexpected token: %s", string(l.token))
}

// lexName scans a name literal.
// The $ has been scanned.
func lexName(l *lexer) stateFn {
	if !l.accept(name) {
		return l.errorf("unexpected character in name literal: %#U", l.peek())
	}
	l.acceptRun(name)
	l.emit(NAME)
	return lexRightDelim
}

// lexString scans a string literal.
// The first " has been scanned.
func lexString(l *lexer) stateFn {
	for l.readErr == nil {
		r := l.read()
		switch {
		case r == '"':
			l.emit(STRING)
			return lexRightDelim
		case r == '\\' && !l.accept(`nt\'"`) && !(l.accept(hexDigits) && l.accept(hexDigits)):
			return l.errorf("illegal escape in string literal: %#U", r)
		case r == '\n' || r == eof:
			return l.errorf("unclosed string literal")
		case 0x00 <= r && r <= 0x1f, r == 0x7f:
			return l.errorf("illegal control character in string literal: %#U", r)
		default: // no-op
		}
	}
	// l.readErr != nil
	return nil
}

// lexNumber scans an number literal.
// This is not a perfect number scanner, check its output via strconv.
// FIXME: match the spec.
func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("unexpected character in number literal: %#U", l.peek())
	}
	l.emit(NUMBER)
	return lexRightDelim
}

func (l *lexer) scanNumber() bool {
	// Optional leading sign
	l.accept("+-")
	// Is it hex?
	d := digits
	if l.accept("0") && l.accept("xX") {
		d = hexDigits
	}
	l.acceptRun(d)
	if l.accept(".") {
		l.acceptRun(d)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun(digits)
	}
	// Next thing must not be alphanumeric
	if isAlphaNumeric(l.peek()) {
		return false
	}
	return true
}

func (l *lexer) emit(typ tokenType) {
	l.tokens = append(l.tokens, token{typ: typ, text: l.token})
	l.token = nil
	l.runeSize = 0
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens = append(l.tokens, token{
		typ:  ERROR,
		text: []byte(fmt.Sprintf(format, args...)),
	})
	return nil
}

// read returns the next rune.
// On error, it returns eof and sets readErr.
func (l *lexer) read() rune {
	r, size, err := l.r.ReadRune()
	if err != nil {
		l.readErr = err
		l.runeSize = 0
		return eof
	}
	l.token = append(l.token, string(r)...)
	l.runeSize = size
	return r
}

// unread steps back one rune.
// It panics if called more than once per call of read.
// It has no effect if the previous read returned eof.
func (l *lexer) unread() {
	if l.readErr != nil { // eof
		return
	}
	if l.runeSize == 0 {
		panic("invalid use of unread")
	}
	l.r.UnreadRune() // erroneous cases guarded above
	l.token = l.token[:len(l.token)-l.runeSize]
	l.runeSize = 0
}

// peek returns but does not consume the next rune.
func (l *lexer) peek() rune {
	r := l.read()
	l.unread()
	return r
}

// discard skips the next rune if it is in the valid set.
func (l *lexer) discard(valid string) {
	if !containsRune(valid, l.read()) {
		l.unread()
		return
	}
	l.token = nil
}

// discardRun skips a run of runes from the valid set.
func (l *lexer) discardRun(valid string) {
	for containsRune(valid, l.read()) {
	}
	l.unread()
	l.token = nil
	return
}

// accept consumes the next rune if it is in the valid set.
func (l *lexer) accept(valid string) bool {
	if containsRune(valid, l.read()) {
		return true
	}
	l.unread()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for containsRune(valid, l.read()) {
	}
	l.unread()
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.token = nil
	l.runeSize = 0
}

// containsRune reports whether r is in s.
// It returns false if r is eof (useful for loops that must terminate on eof).
func containsRune(s string, r rune) bool {
	if r == eof {
		return false
	}
	return strings.ContainsRune(s, r)
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
