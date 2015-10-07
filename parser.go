package nstcc

func (c *compiler) next() error {
	c.tok = tokEOF
	return nil
}
