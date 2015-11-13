package nstcc

import (
	"errors"
	"fmt"
	"reflect"
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

		if c.macroPtr == nil {
			if c.tok < tokIdent || c.parseFlags&parseFlagPreprocess == 0 {
				break
			}
			s := c.idents.defineFind(c.tok)
			if s == nil {
				break
			}
			ts, nestedList, ml := tokenString{}, (*sym)(nil), (*macroLevel)(nil)
			if c.macroSubstTok(&ts, &nestedList, s, &ml) == macroSubstTokNoSubstitute {
				break
			}
			c.macroPtr = ts.tokStr
		} else if c.tok == tokEOM {
			if false { // TODO: unget buffer.
			} else {
				c.macroPtr = nil
			}
		} else if c.tok != tokNoSubst {
			break
		}
	}

	if c.tok == tokPPNum && c.parseFlags&parseFlagTokNum != 0 {
		return c.parseNumber(c.tokc.str)
	}
	return nil
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
	if c.macroPtr == nil {
		return c.nextNoMacro1()
	}
	for {
		tv := c.macroPtr[0]
		c.macroPtr = c.macroPtr[1:]
		c.tok = tv.tok
		c.tokc = tv.tokc
		if c.tok != tokLineNum {
			return nil
		}
		// TODO: file->line_num = tokc.i;
	}
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

		case t == '\n' || t == '#':
			c.s++
			if t == '\n' {
				// TODO: file->line_num++
				c.tokFlags |= tokFlagBOL
			} else if c.tokFlags&tokFlagBOL == 0 || c.parseFlags&parseFlagPreprocess == 0 {
				if c.peekc() == '#' {
					c.s++
					c.tok = tokTwoSharps
				} else if false {
					// TODO: parseFlagAsmComments.
				} else {
					c.tok = '#'
				}
				c.tokFlags = 0
				return nil
			} else if err := c.preprocess(c.tokFlags&tokFlagBOF != 0); err != nil {
				return err
			}
			if c.parseFlags&parseFlagLineFeed == 0 {
				continue redoNoStart
			}
			c.tok = '\n'
			return nil

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

		case t == '.':
			c.s++
			switch c.peekc() {
			default:
				c.tok = '.'
				c.tokFlags = 0
				return nil
			case '.':
				c.s++
				if c.peekc() != '.' {
					return errors.New(`nstcc: incomplete "..." token`)
				}
				c.s++
				c.tok = tokDots
				c.tokFlags = 0
				return nil
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				// No-op, and fall through below.
			}
			c.s--
			fallthrough

		case isNum[t]:
			s := c.s + 1
			for ; s < len(c.src); s++ {
				x := c.src[s]
				if isNum[x] || isID[x] || x == '.' {
					continue
				}
				if x == '-' || x == '+' {
					switch c.src[s-1] {
					case 'E', 'P', 'e', 'p':
						continue
					}
				}
				break
			}
			c.tok = tokPPNum
			c.tokc.str = c.src[c.s:s]
			c.tokFlags = 0
			c.s = s
			return nil

		case t == '\'' || t == '"':
			c.s++
			str, err := c.parseString(t, false, false)
			if err != nil {
				return err
			}
			str, err = unescape(str)
			if err != nil {
				return err
			}
			if t == '\'' {
				if len(str) == 0 {
					return errors.New("nstcc: empty character constant")
				} else if len(str) > 1 {
					return errors.New("nstcc: multi-character character constant")
				}
				c.tok = tokCChar // TODO: map L'x' chars to tokLChar.
				c.tokc.int = int64(str[0])
			} else {
				c.tok = tokStr // TODO: map L"xxx" strings to tokLStr.
				c.tokc.str = str
			}
			return nil

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

func (c *compiler) preprocess(isBOF bool) error {
	// TODO: save/restore c.src and c.s as file->buf_ptr.

	savedParseFlags := c.parseFlags
	c.parseFlags = parseFlagPreprocess | parseFlagTokNum | parseFlagLineFeed
	defer func() {
		c.parseFlags = savedParseFlags
	}()

	if err := c.nextNoMacro(); err != nil {
		return err
	}

redo:
	for {
		switch c.tok {
		case tokDefine:
			if err := c.nextNoMacro(); err != nil {
				return err
			}
			if err := c.parseDefine(); err != nil {
				return err
			}

			// TODO: other cases.
		}

		break redo
	}

	for c.tok != '\n' {
		if c.tok == tokEOF {
			return errors.New("nstcc: unexpected end of file")
		}
		if err := c.nextNoMacro(); err != nil {
			return err
		}
	}
	return nil
}

type macroSubstTokResult bool

const (
	macroSubstTokSubstitute   macroSubstTokResult = false
	macroSubstTokNoSubstitute macroSubstTokResult = true
)

func (c *compiler) macroSubstTok(ts *tokenString, nestedList **sym, s *sym, canReadStream **macroLevel) macroSubstTokResult {
	// TODO.
	switch c.tok {
	case tok__LINE__:
		// TODO.
	case tok__FILE__:
		// TODO.
	case tok__DATE__, tok__TIME__:
		// TODO.
	default:
		mStr := s.defineTokStr
		if s.cType.typ == macroFunc {
			// TODO.
		}

		c.symPush2(nestedList, s.tok, 0, 0)
		c.macroSubst(ts, nestedList, mStr, canReadStream)
		*nestedList = (*nestedList).stackPrev
	}
	return macroSubstTokSubstitute
}

func (c *compiler) macroSubst(ts *tokenString, nestedList **sym, mStr []tokenValue, canReadStream **macroLevel) {
	// TODO: handle two-sharps.
	spc, forceBlank := false, false
	for m := mStr; len(m) > 0; {
		tv := m[0]
		m = m[1:]
		if tv.tok == tokNoSubst {
			ts.tokStr = append(ts.tokStr, tv)
			tv = m[0]
			m = m[1:]
		} else if s := c.idents.defineFind(tv.tok); s != nil {
			if symFind2(*nestedList, tv.tok) != nil {
				ts.tokStr = append(ts.tokStr, tokenValue{tok: tokNoSubst})
			} else {
				ml := macroLevel{tokStr: c.macroPtr}
				if canReadStream != nil {
					ml.prev = *canReadStream
					*canReadStream = &ml
				}
				c.macroPtr = m
				c.tok = tv.tok
				ret := c.macroSubstTok(ts, nestedList, s, canReadStream)
				m = c.macroPtr
				c.macroPtr = ml.tokStr
				if canReadStream != nil && *canReadStream == &ml {
					*canReadStream = ml.prev
				}
				if ret == macroSubstTokSubstitute {
					if c.parseFlags&parseFlagSpaces != 0 {
						forceBlank = true
					}
					continue
				}
			}
		}

		if forceBlank {
			ts.tokStr = append(ts.tokStr, tokenValue{tok: ' '})
			spc, forceBlank = true, false
		}
		if !checkSpace(tv.tok, &spc) {
			ts.tokStr = append(ts.tokStr, tv)
		}
	}
}

func symFind2(s *sym, t token) *sym {
	for ; s != nil; s = s.stackPrev {
		if s.tok == t {
			return s
		}
	}
	return nil
}

func checkSpace(t token, spc *bool) bool {
	if t.isSpace() {
		if *spc {
			return true
		}
		*spc = true
	} else {
		*spc = false
	}
	return false
}

func (c *compiler) parseDefine() error {
	name := c.tok
	if name < tokIdent {
		return errors.New("nstcc: invalid macro name")
	}

	if err := c.nextNoMacroSpace(); err != nil {
		return err
	}
	typ, tokStr, first := macroObj, []tokenValue(nil), (*sym)(nil)
	if c.tok == '(' {
		// TODO.
		typ = macroFunc
	}

	for c.tok != '\n' && c.tok != tokEOF {
		// TODO: remove spaces around ## and after '#'.
		tokStr = append(tokStr, tokenValue{tok: c.tok, tokc: c.tokc})
		if err := c.nextNoMacroSpace(); err != nil {
			return err
		}
	}
	// TODO: strip trailing space.
	tokStr = append(tokStr, tokenValue{tok: tokEOM})
	return c.definePush(name, typ, tokStr, first)
}

func (c *compiler) definePush(name token, typ macroType, tokStr []tokenValue, first *sym) error {
	s := c.idents.defineFind(name)
	if s != nil && !reflect.DeepEqual(s.defineTokStr, tokStr) {
		return fmt.Errorf("nstcc: %q redfined", c.idents.byTok(name).str)
	}
	s = c.symPush2(&c.defineStack, name, typ, 0)
	s.defineTokStr = tokStr
	s.next = first
	c.idents.byTok(name).symDefine = s
	return nil
}

func (c *compiler) parseString(sep byte, isLong bool, justSkip bool) (ret []byte, retErr error) {
loop:
	for {
		if c.s >= len(c.src) {
			return nil, errors.New("nstcc: unexpected end of file in string")
		}
		x := c.src[c.s]
		c.s++
		switch x {
		case sep:
			break loop
		case '\\':
			switch y := c.peekc(); y {
			default:
				c.s++
				if !justSkip {
					ret = append(ret, x, byte(y))
				}
			case tokEOF:
				return nil, errors.New("nstcc: unexpected end of file in string")
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
		if !justSkip {
			ret = append(ret, x)
		}
	}
	return ret, nil
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

func unescape(s []byte) ([]byte, error) {
	j := 0
	for i := 0; i < len(s); {
		x := s[i]
		if x != '\\' {
			s[j] = x
			i++
			j++
			continue
		}
		i++
		if i == len(s) {
			return nil, errors.New("nstcc: unexpected end of string after backslash")
		}
		x = s[i]
		i++
		switch x {
		default:
			return nil, fmt.Errorf(`nstcc: invalid escape sequence '\%c'`, x)
		case '0', '1', '2', '3', '4', '5', '6', '7':
			n := int(x) - '0'
			if i < len(s) && isOctal(s[i]) {
				n = 8*n + int(s[i]) - '0'
				i++
				if i < len(s) && isOctal(s[i]) {
					n = 8*n + int(s[i]) - '0'
					i++
				}
			}
			s[j] = byte(n)
			j++
		case 'x': // TODO: 'u', 'U'.
			// TODO: check that this is correct for something like "\xabc".
			// Does "\x" imply at most two hex digits afterwards?
			n := 0
		loop:
			for ; i < len(s); i++ {
				switch x := s[i]; {
				default:
					break loop
				case '0' <= x && x <= '9':
					n = 16*n + int(x) - '0'
				case 'A' <= x && x <= 'F':
					n = 16*n + int(x) - ('A' - 10)
				case 'a' <= x && x <= 'f':
					n = 16*n + int(x) - ('a' - 10)
				}
			}
			s[j] = byte(n)
			j++
		case 'a':
			s[j] = '\a'
			j++
		case 'b':
			s[j] = '\b'
			j++
		case 'f':
			s[j] = '\f'
			j++
		case 'n':
			s[j] = '\n'
			j++
		case 'r':
			s[j] = '\r'
			j++
		case 't':
			s[j] = '\t'
			j++
		case 'v':
			s[j] = '\v'
			j++
		case 'e':
			s[j] = '\x1b'
			j++
		case '\'', '"', '\\', '?':
			s[j] = x
			j++
		}
	}
	return s[:j], nil
}

func isOctal(x byte) bool {
	return '0' <= x && x <= '7'
}

func (c *compiler) parseNumber(s []byte) error {
	base := 10
	if s[0] == '.' {
		// TODO: goto float_frac_parse.
	}
	if s[0] == '0' && len(s) > 1 {
		switch s[1] {
		case 'B', 'b':
			base = 2
			s = s[2:]
		case 'X', 'x':
			base = 16
			s = s[2:]
		}
	}

	i := 0
	for ; i < len(s); i++ {
		x := s[i]
		switch {
		case '0' <= x && x <= '9':
			x -= '0'
		case 'A' <= x && x <= 'F':
			x -= 'A' - 10
		case 'a' <= x && x <= 'f':
			x -= 'a' - 10
		}
		if int(x) >= base {
			break
		}
	}

	if i < len(s) {
		x := s[i]
		if x == '.' ||
			((x == 'E' || x == 'e') && (base == 10)) ||
			((x == 'P' || x == 'p') && (base == 16 || base == 2)) {

			return fmt.Errorf("nstcc: TODO: parse floating point numbers")
		}
	}

	if base == 10 && s[0] == '0' {
		base = 8
		s = s[1:]
		i--
	}

	n := int64(0)
	for _, x := range s[:i] {
		switch {
		case '0' <= x && x <= '9':
			x -= '0'
		case 'A' <= x && x <= 'F':
			x -= 'A' - 10
		case 'a' <= x && x <= 'f':
			x -= 'a' - 10
		}
		if int(x) >= base {
			return fmt.Errorf("nstcc: invalid number")
		}
		n = n*int64(base) + int64(x)
		// TODO: detect overflow.
	}

	// TODO: recognize tokCULLong and tokCLLong.
	if n > 0x7fffffff {
		c.tok = tokCUint
	} else {
		c.tok = tokCInt
	}

	if i != len(s) {
		// TODO: deal with trailing Us and Ls in s[:i], such as in parsing
		// "123ULL". Note that floating point constants don't have Us and Ls.
		return fmt.Errorf("nstcc: invalid number")
	}

	c.tokc.int = n
	return nil
}
