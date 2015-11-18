package nstcc

import (
	"bufio"
	"io"
)

type arch uint32

const (
	archAMD64 arch = 0
	arch386   arch = 1
)

type macroType uint32

const (
	macroObj  macroType = 0
	macroFunc macroType = 1
)

type tokFlags uint32

const (
	tokFlagBOL   tokFlags = 0x01
	tokFlagBOF   tokFlags = 0x02
	tokFlagEndIf tokFlags = 0x04
	tokFlagEOF   tokFlags = 0x08
)

type parseFlags uint32

const (
	parseFlagPreprocess  parseFlags = 0x01
	parseFlagTokNum      parseFlags = 0x02
	parseFlagLineFeed    parseFlags = 0x04
	parseFlagAsmComments parseFlags = 0x08
	parseFlagSpaces      parseFlags = 0x10
)

type cType struct {
	typ macroType // TODO: s/macroType/token/?
	sym *sym
}

type cValue struct {
	int int64
	str []byte
}

type compiler struct {
	ctx *Context
	dst *bufio.Writer

	src []byte
	s   int

	macroPtr []tokenValue
	m        int

	tokFlags   tokFlags
	parseFlags parseFlags

	tok  token
	tokc cValue

	globalStack      *sym
	localStack       *sym
	localLabelStack  *sym
	globalLabelStack *sym
	defineStack      *sym

	idents idents
}

func newCompiler(ctx *Context, dst io.Writer, src []byte) *compiler {
	bw, ok := dst.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(dst)
	}
	c := &compiler{
		ctx:      ctx,
		dst:      bw,
		src:      src,
		tokFlags: tokFlagBOL | tokFlagBOF,
	}
	c.idents.init()
	return c
}
