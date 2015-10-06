package nstcc

import (
	"bufio"
	"io"
)

type parser struct {
	ctx *Context
	dst *bufio.Writer
	src []byte
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

func (p *parser) preprocess() error {
	return nil
}
