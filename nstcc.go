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
	c := newCompiler(ctx, dst, src)
	c.parseFlags = parseFlagPreprocess | parseFlagLineFeed | parseFlagAsmComments | parseFlagSpaces
	for {
		if err := c.next(); err != nil {
			return err
		}
		if c.tok == tokEOF {
			break
		}
	}
	return nil
}

type tokenSym struct {
	hashNext *tokenSym

	symDefine     *sym
	symLabel      *sym
	symStruct     *sym
	symIdentifier *sym

	tok token
	str []byte
}

const symFirstAnom = 0x10000000 // First anonymous sym.

type sym struct {
	tok token
	// TODO: asmLabel []byte
	// TODO: r for register.
	// TODO: c for associated number.
	defineTokStr []tokenValue
	// TODO: cType.
	// TODO: jnext for jump-next.
	next      *sym
	stackPrev *sym
	tokPrev   *sym
}
