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

// The TCC code calls this tok_alloc.
func (m *idents) byStr(s []byte) (*tokenSym, error) {
	h := uint32(hashInit)
	for _, c := range s {
		h = hashFunc(h, c)
	}
	p := &m.hash[h&(hashSize-1)]
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
func (m *idents) alloc(p **tokenSym, s []byte) (*tokenSym, error) {
	tok := tokIdent + token(len(m.list))
	if tok >= symFirstAnom {
		return nil, errors.New("nstcc: memory full")
	}
	t := &tokenSym{
		tok: tok,
		str: dup(s),
	}
	m.list = append(m.list, t)
	*p = t
	return t, nil
}

// tokIdent is the token value of the first identifier token. Token values
// greater than or equal to tokIdent represent identifiers. Token values less
// than tokIdent represent symbols.
const tokIdent = 256

type token int32

type tokenSym struct {
	hashNext *tokenSym
	// TODO: symThis, symThat.
	tok token
	str []byte
}

const symFirstAnom = 0x10000000 // First anonymous sym.

type sym struct {
	// TODO.
}
