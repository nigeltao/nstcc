package nstcc

import (
	"bytes"
	"errors"
)

func dup(s []byte) []byte {
	t := make([]byte, len(s))
	copy(t, s)
	return t
}

const (
	hashInit = 1
	hashSize = 8192
)

func hashFunc(h uint32, c uint8) uint32 {
	return h*263 + uint32(c)
}

type idents struct {
	list []*tokenSym
	hash [hashSize]*tokenSym
}

func (m *idents) init() {
	i := uint32(0)
	for _, j := range builtInTokensLengths {
		m.byStr(builtInTokensStrings[i:j])
		i = j
	}
}

func (m *idents) byTok(t token) *tokenSym {
	return m.list[t-tokIdent]
}

// The TCC code calls this tok_alloc.
func (m *idents) byStr(s []byte) (*tokenSym, error) {
	h := uint32(hashInit)
	for _, c := range s {
		h = hashFunc(h, c)
	}
	p := &m.hash[h&(hashSize-1)]
	for {
		y := *p
		if y == nil {
			break
		}
		if bytes.Equal(s, y.str) {
			return y, nil
		}
		p = &y.hashNext
	}
	return m.alloc(p, s)
}

// The TCC code calls this tok_alloc_new.
func (m *idents) alloc(p **tokenSym, s []byte) (*tokenSym, error) {
	t := tokIdent + token(len(m.list))
	if t >= symFirstAnom {
		return nil, errors.New("nstcc: memory full")
	}
	y := &tokenSym{
		tok: t,
		str: dup(s),
	}
	m.list = append(m.list, y)
	*p = y
	return y, nil
}

func (m *idents) defineFind(t token) *sym {
	t -= tokIdent
	if t < 0 || len(m.list) <= int(t) {
		return nil
	}
	return m.list[t].symDefine
}

func (m *idents) defineUndef(s *sym) {
	t := s.tok - tokIdent
	if 0 <= t && int(t) < len(m.list) {
		m.list[t].symDefine = nil
	}
	s.tok = 0
}

type token int32

func (t token) isSpace() bool {
	return t == ' ' || t == '\t' || t == '\v' || t == '\f' || t == '\r'
}

type tokenValue struct {
	tok  token
	tokc cValue
}

const (
	tokEOF token = -1

	tokULT    token = 0x92
	tokUGE    token = 0x93
	tokEq     token = 0x94
	tokNE     token = 0x95
	tokULE    token = 0x96
	tokUGT    token = 0x97
	tokNset   token = 0x98
	tokNclear token = 0x99
	tokLT     token = 0x9c
	tokGE     token = 0x9d
	tokLE     token = 0x9e
	tokGT     token = 0x9f

	tokLAnd token = 0xa0
	tokLOr  token = 0xa1

	tokDec       token = 0xa2
	tokMid       token = 0xa3
	tokInc       token = 0xa4
	tokUDiv      token = 0xb0
	tokUMod      token = 0xb1
	tokPDiv      token = 0xb2
	tokCInt      token = 0xb3
	tokCChar     token = 0xb4
	tokStr       token = 0xb5
	tokTwoSharps token = 0xb6
	tokLChar     token = 0xb7
	tokLStr      token = 0xb8
	tokCFloat    token = 0xb9
	tokLineNum   token = 0xba
	tokCDouble   token = 0xc0
	tokCLDouble  token = 0xc1
	tokUMull     token = 0xc2
	tokAddC1     token = 0xc3
	tokAddC2     token = 0xc4
	tokSubC1     token = 0xc5
	tokSubC2     token = 0xc6
	tokCUint     token = 0xc8
	tokCLLong    token = 0xc9
	tokCULLong   token = 0xca
	tokArrow     token = 0xcb
	tokDots      token = 0xcc
	tokShR       token = 0xcd
	tokPPNum     token = 0xce
	tokNoSubst   token = 0xcf

	tokShL token = 0x01
	tokSAR token = 0x02

	tokAMod token = 0xa5
	tokAAnd token = 0xa6
	tokAMul token = 0xaa
	tokAAdd token = 0xab
	tokASub token = 0xad
	tokADiv token = 0xaf
	tokAXor token = 0xde
	tokAOr  token = 0xfc
	tokAShL token = 0x81
	tokASAR token = 0x82

	// tokIdent is the token value of the first identifier token. Token values
	// greater than or equal to tokIdent represent identifiers. Token values
	// less than tokIdent represent symbols.
	tokIdent token = 0x100
)
