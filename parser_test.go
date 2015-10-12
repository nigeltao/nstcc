package nstcc

import (
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	// Map identifier strings like "main" to (negative) placeholder token
	// values.
	t2s := map[token]string{}
	s2t := map[string]token{}
	ident := func(s string) token {
		x, ok := s2t[s]
		if !ok {
			x = token(^len(s2t))
			s2t[s] = x
			t2s[x] = s
		}
		return x
	}

	testCases := []struct {
		desc string
		src  string
		want []token
	}{{
		"hello world",
		// TODO: s/XXX/42/
		"int main(int argc, char** argv) { return XXX; }",
		[]token{
			tokInt,
			ident("main"),
			'(',
			tokInt,
			ident("argc"),
			',',
			tokChar,
			'*',
			'*',
			ident("argv"),
			')',
			'{',
			tokReturn,
			ident("XXX"),
			';',
			'}',
		},
	}, {
		"slash star",
		"int /*foo*/ x;",
		[]token{
			tokInt,
			ident("x"),
			';',
		},
	}, {
		"slash slash",
		"int x;\n// int y;\nint z;\n",
		[]token{
			tokInt,
			ident("x"),
			';',
			tokInt,
			ident("z"),
			';',
		},
	}}

	for _, tc := range testCases {
		c := newCompiler(nil, nil, []byte(tc.src))
		var got []token
		for {
			if err := c.next(); err != nil {
				t.Fatal(err)
			}
			if c.tok == tokEOF {
				break
			}
			got = append(got, c.tok)
		}

		// Replace the placeholder identifier tokens with their real values.
		for i, x := range tc.want {
			if x >= 0 {
				continue
			}
			ts, err := c.idents.byStr([]byte(t2s[x]))
			if err != nil {
				t.Fatal(err)
			}
			tc.want[i] = ts.tok
		}

		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("parsing %q:\ngot  %v\nwant %v", tc.desc, got, tc.want)
		}
	}
}
