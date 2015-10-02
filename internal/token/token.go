package token

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

type Map struct {
	tableIdent []*TokenSym
	hashIdent  [hashSize]*TokenSym
}

// The TCC code calls this tok_alloc.
func (m *Map) Find(s []byte) (*TokenSym, error) {
	h := uint32(hashInit)
	for _, c := range s {
		h = hashFunc(h, c)
	}
	p := &m.hashIdent[h&(hashSize-1)]
	for {
		t := *p
		if t == nil {
			break
		}
		if bytes.Equal(s, t.str) {
			return t, nil
		}
		p = &t.hashNext
	}
	return m.alloc(p, s)
}

// The TCC code calls this tok_alloc_new.
func (m *Map) alloc(p **TokenSym, s []byte) (*TokenSym, error) {
	tok := TokIdent + Token(len(m.tableIdent))
	if tok >= symFirstAnom {
		return nil, errors.New("token: memory full")
	}
	t := &TokenSym{
		tok: tok,
		str: dup(s),
	}
	m.tableIdent = append(m.tableIdent, t)
	*p = t
	return t, nil
}

// TokIdent is the Token value of the first identifier token. Token values
// greater than or equal to TokIdent represent identifiers. Token values less
// than TokIdent represent symbols.
const TokIdent = 256

type Token int32

type TokenSym struct {
	hashNext *TokenSym
	// TODO: symThis, symThat.
	tok Token
	str []byte
}

const symFirstAnom = 0x10000000 // First anonymous sym.

type Sym struct {
	// TODO.
}
