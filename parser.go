package nstcc

import (
	"errors"
	"fmt"
)

func (c *compiler) next() error {
	for {
		if c.parseFlags&parseFlagSpaces != 0 {
			if err := c.nextNoMacroSpace(); err != nil {
				return err
			}
		} else {
			if err := c.nextNoMacro(); err != nil {
				return err
			}
		}

		if true { // TODO: the TCC code says "if (!macro_ptr)".
			if c.tok >= tokIdent && c.parseFlags&parseFlagPreprocess != 0 {
				// TODO: if not reading from macro substituted string, then try
				// to substitute macros.
			}

		} else {
			// TODO: macro_ptr code path.
		}

		if c.tok == tokPPNum && c.parseFlags&parseFlagTokNum != 0 {
			// TODO: convert preprocessor tokens into C tokens.
		}

		return nil
	}
}

func (c *compiler) nextNoMacro() error {
	for {
		if err := c.nextNoMacroSpace(); err != nil {
			return err
		}
		if !c.tok.isSpace() {
			return nil
		}
	}
}

func (c *compiler) nextNoMacroSpace() error {
	// TODO: check what the TCC code calls macro_ptr.
	return c.nextNoMacro1()
}

func (c *compiler) nextNoMacro1() error {
redoNoStart:
	for {
		if c.s >= len(c.src) {
			c.tok = tokEOF
			return nil
		}

		switch t := c.src[c.s]; {
		default:
			return fmt.Errorf(`nstcc: unrecognized token '\x%02x'`, t)

		case t == ' ' || t == '\t':
			c.tok = token(t)
			c.s++
			return nil

		case t == '\f' || t == '\v' || t == '\r':
			c.s++
			continue redoNoStart

		case t == '\\':
			// TODO.

		case t == '\n':
			// TODO: file->line_num++
			c.tokFlags |= tokFlagBOL
			c.s++
			if c.parseFlags&parseFlagLineFeed == 0 {
				continue redoNoStart
			}
			c.tok = '\n'
			return nil

		case t == '#':
			// TODO.

		case t == 'L':
			// TODO: parse things like the wchar_t L"abc".
			fallthrough

		case isID[t]:
			s := c.s + 1
			for ; s < len(c.src); s++ {
				t = c.src[s]
				if !isIDNum[t] {
					break
				}
			}

			if t != '\\' {
				ts, err := c.idents.byStr(c.src[c.s:s])
				if err != nil {
					return err
				}
				c.tok = ts.tok
			} else {
				// TODO.
			}

			c.tokFlags = 0
			c.s = s
			return nil

		case t == '\'' || t == '"':
			c.s++
			return c.parseString(t, false)

		case t == '<':
			c.s++
			switch c.peekc() {
			case '=':
				c.s++
				c.tok = tokLE
			case '<':
				c.s++
				if c.peekc() == '=' {
					c.s++
					c.tok = tokAShL
				} else {
					c.tok = tokShL
				}
			default:
				c.tok = tokLT
			}
			c.tokFlags = 0
			return nil

		case t == '>':
			c.s++
			switch c.peekc() {
			case '=':
				c.s++
				c.tok = tokGE
			case '>':
				c.s++
				if c.peekc() == '=' {
					c.s++
					c.tok = tokASAR
				} else {
					c.tok = tokSAR
				}
			default:
				c.tok = tokGT
			}
			c.tokFlags = 0
			return nil

		case t == '&':
			c.s++
			switch c.peekc() {
			case '&':
				c.s++
				c.tok = tokLAnd
			case '=':
				c.s++
				c.tok = tokAAnd
			default:
				c.tok = '&'
			}
			c.tokFlags = 0
			return nil

		case t == '|':
			c.s++
			switch c.peekc() {
			case '|':
				c.s++
				c.tok = tokLOr
			case '=':
				c.s++
				c.tok = tokAOr
			default:
				c.tok = '|'
			}
			c.tokFlags = 0
			return nil

		case t == '+':
			c.s++
			switch c.peekc() {
			case '+':
				c.s++
				c.tok = tokInc
			case '=':
				c.s++
				c.tok = tokAAdd
			default:
				c.tok = '+'
			}
			c.tokFlags = 0
			return nil

		case t == '-':
			c.s++
			switch c.peekc() {
			case '-':
				c.s++
				c.tok = tokDec
			case '=':
				c.s++
				c.tok = tokASub
			case '>':
				c.s++
				c.tok = tokArrow
			default:
				c.tok = '-'
			}
			c.tokFlags = 0
			return nil

		case t == '!':
			c.s++
			if c.peekc() == '=' {
				c.s++
				c.tok = tokNE
			} else {
				c.tok = '!'
			}
			c.tokFlags = 0
			return nil

		case t == '=':
			c.s++
			if c.peekc() == '=' {
				c.s++
				c.tok = tokEq
			} else {
				c.tok = '='
			}
			c.tokFlags = 0
			return nil

		case t == '*':
			c.s++
			if c.peekc() == '=' {
				c.s++
				c.tok = tokAMul
			} else {
				c.tok = '*'
			}
			c.tokFlags = 0
			return nil

		case t == '%':
			c.s++
			if c.peekc() == '=' {
				c.s++
				c.tok = tokAMod
			} else {
				c.tok = '%'
			}
			c.tokFlags = 0
			return nil

		case t == '^':
			c.s++
			if c.peekc() == '=' {
				c.s++
				c.tok = tokAXor
			} else {
				c.tok = '^'
			}
			c.tokFlags = 0
			return nil

		case t == '/':
			c.s++
			switch c.peekc() {
			case '*':
				c.tok = ' '
				c.s++
				return c.parseSlashStarComment()
			case '/':
				c.tok = ' '
				c.s++
				c.parseSlashSlashComment()
				return nil
			case '=':
				c.tok = tokADiv
				c.tokFlags = 0
				c.s++
				return nil
			default:
				c.tok = '/'
				c.tokFlags = 0
				c.s++
				return nil
			}

		case isSimpleToken[t]:
			c.tok = token(t)
			c.tokFlags = 0
			c.s++
			return nil
		}
	}
}

func (c *compiler) peekc() token {
	if c.s >= len(c.src) {
		return tokEOF
	}
	// TODO: handle c == '\\'.
	return token(c.src[c.s])
}

func (c *compiler) parseString(sep byte, isLong bool) error {
	var str []byte
loop:
	for {
		if c.s >= len(c.src) {
			return errors.New("nstcc: unexpected end of file in string")
		}
		x := c.src[c.s]
		switch x {
		case sep:
			c.s++
			break loop
		case '\\':
			switch y := c.peekc(); y {
			default:
				c.s++
				str = append(str, x, byte(y))
			case tokEOF:
				return errors.New("nstcc: unexpected end of file in string")
			case '\n':
				// TODO: file->line_num++
				c.s++

				// TODO: case '\r':
			}
			continue loop
		case '\n':
			// TODO: file->line_num++

			// TODO: case '\r':
			// Note that it says PEEKC_EOB instead of PEEKC.
		}
		c.s++
		str = append(str, x)
	}
	println("got string", string(str)) // TODO
	return nil
}

func (c *compiler) parseSlashStarComment() error {
	star := false
	for c.s < len(c.src) {
		switch x := c.src[c.s]; x {
		default:
			c.s++
			star = x == '*'
		case '\n':
			// TODO: file->line_num++
			c.s++
			star = false
		case '/':
			c.s++
			if star {
				return nil
			}
			star = false

			// TODO: case '\\':
		}
	}
	return errors.New("nstcc: unexpected end of file in comment")
}

func (c *compiler) parseSlashSlashComment() {
	for c.s < len(c.src) {
		switch c.src[c.s] {
		default:
			c.s++
		case '\n':
			return
			// TODO: case '\\':
		}
	}
}
