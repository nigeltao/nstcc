package nstcc

import (
	"bufio"
	"io"
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

type parser struct {
	ctx *Context
	dst *bufio.Writer
	src []byte

	tokFlags   tokFlags
	parseFlags parseFlags

	tok token
	s   int

	idents idents
}

func newParser(ctx *Context, dst io.Writer, src []byte) *parser {
	bw, ok := dst.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(dst)
	}
	p := &parser{
		ctx: ctx,
		dst: bw,
		src: src,
	}
	p.idents.init()
	return p
}

func (p *parser) next() error {
	p.tok = tokEOF
	return nil
}
