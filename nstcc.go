//go:generate go run gen.go

package nstcc

import (
	"io"
)

type Context struct {
	// TODO: add something about how to resolve and read filenames like
	// "foo.h" and <stdio.h>.
}

func Preprocess(ctx *Context, dst io.Writer, src []byte) error {
	p := newParser(ctx, dst, src)
	p.tokFlags = tokFlagBOL | tokFlagBOF
	p.parseFlags = parseFlagPreprocess | parseFlagLineFeed | parseFlagAsmComments | parseFlagSpaces
	for {
		if err := p.next(); err != nil {
			return err
		}
		if p.tok == tokEOF {
			break
		}
	}
	return nil
}

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
