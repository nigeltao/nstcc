package nstcc

func (c *compiler) next() error {
	for {
		if c.parseFlags&parseFlagSpaces != 0 {
			if err := c.nextNoMacroSpace(); err != nil {
				return err
			}
		} else {
			if err := c.nextNoMacro(); err != nil {
				return err
			}
		}

		if true { // TODO: the TCC code says "if (!macro_ptr)".
			if c.tok >= tokIdent && c.parseFlags&parseFlagPreprocess != 0 {
				// TODO: if not reading from macro substituted string, then try
				// to substitute macros.
			}

		} else {
			// TODO: macro_ptr code path.
		}

		if c.tok == tokPPNum && c.parseFlags&parseFlagTokNum != 0 {
			// TODO: convert preprocessor tokens into C tokens.
		}

		return nil
	}
}

func (c *compiler) nextNoMacro() error {
	for {
		if err := c.nextNoMacroSpace(); err != nil {
			return err
		}
		if !c.tok.isSpace() {
			return nil
		}
	}
}

func (c *compiler) nextNoMacroSpace() error {
	// TODO: check what the TCC code calls macro_ptr.
	return c.nextNoMacro1()
}

func (c *compiler) nextNoMacro1() error {
	c.tok = tokEOF
	return nil
}
