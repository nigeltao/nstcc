package nstcc

import (
	"bufio"
	"io"
)

type Arch uint32

const (
	ArchAMD64 Arch = 0
	Arch386   Arch = 1
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

	textSection     *section
	dataSection     *section
	bssSection      *section
	curTextSection  *section
	lastTextSection *section

	globalStack      *sym
	localStack       *sym
	localLabelStack  *sym
	globalLabelStack *sym
	defineStack      *sym

	rsym    int32
	anonSym int32
	ind     int32
	loc     int32

	funcName []byte

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
		anonSym:  symFirstAnon,
	}
	c.idents.init()
	return c
}
