//go:generate go run gen.go

package nstcc

import (
	"io"
)

type Context struct {
	Arch Arch

	// TODO: add something about how to resolve and read filenames like
	// "foo.h" and <stdio.h>.
}

func Compile(ctx *Context, dst io.Writer, src []byte) error {
	// TODO.
	return nil
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

const symFirstAnon = 0x10000000 // First anonymous sym.

type sym struct {
	tok token
	// TODO: asmLabel []byte
	// TODO: r for register.
	c            int32 // TODO: int64?
	defineTokStr []tokenValue
	cType        cType
	// TODO: jnext for jump-next.
	next      *sym
	stackPrev *sym
	tokPrev   *sym
}

func (c *compiler) symPush2(ps **sym, tok token, typ macroType, cc int32) *sym {
	if ps == &c.localStack {
		// TODO: look for incompatible types for redefinition.
	}
	s := &sym{
		tok: tok,
		cType: cType{
			typ: typ,
		},
		c:         cc,
		stackPrev: *ps,
	}
	*ps = s
	return s
}
