package nstcc

import (
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	testCases := []struct {
		desc string
		src  string
		want []interface{}
	}{{
		"hello world",
		"int main(int argc, char** argv) { return 42; }",
		[]interface{}{
			tokInt,
			"main",
			'(',
			tokInt,
			"argc",
			',',
			tokChar,
			'*',
			'*',
			"argv",
			')',
			'{',
			tokReturn,
			tokPPNum, // TODO: tokCInt.
			';',
			'}',
		},
	}, {
		"char",
		`x = 'a';`,
		[]interface{}{
			"x",
			'=',
			tokCChar,
			';',
		},
	}, {
		"string",
		`x = "abc";`,
		[]interface{}{
			"x",
			'=',
			tokStr,
			';',
		},
	}, {
		"slash star",
		"int /*foo*/ x;",
		[]interface{}{
			tokInt,
			"x",
			';',
		},
	}, {
		"slash slash",
		"int x;\n// int y;\nint z;\n",
		[]interface{}{
			tokInt,
			"x",
			';',
			tokInt,
			"z",
			';',
		},
	}, {
		"dot dot dot",
		"int printf(const char* format, ...);",
		[]interface{}{
			tokInt,
			"printf",
			'(',
			"const", // TODO: make this tokConst1.
			tokChar,
			'*',
			"format",
			',',
			tokDots,
			')',
			';',
		},
	}}

loop:
	for _, tc := range testCases {
		c := newCompiler(nil, nil, []byte(tc.src))
		var got []token
		for {
			if err := c.next(); err != nil {
				t.Errorf("parsing %q:\n%v", tc.desc, err)
				continue loop
			}
			if c.tok == tokEOF {
				break
			}
			got = append(got, c.tok)
		}

		// Replace the placeholder tokens with their real values.
		want := make([]token, len(tc.want))
		for i, x := range tc.want {
			switch x := x.(type) {
			case rune:
				want[i] = token(x)
			case string:
				ts, err := c.idents.byStr([]byte(x))
				if err != nil {
					t.Errorf("parsing %q:\n%v", tc.desc, err)
					continue loop
				}
				want[i] = ts.tok
			case token:
				want[i] = x
			default:
				t.Errorf("parsing %q:\ninvalid type %T", tc.desc, x)
				continue loop
			}
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("parsing %q:\ngot  %v\nwant %v", tc.desc, got, want)
		}
	}
}
