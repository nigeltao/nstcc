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
		// TODO: s/XXX/42/
		"int main(int argc, char** argv) { return XXX; }",
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
			"XXX",
			';',
			'}',
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

		// Replace the placeholder tokens with their real values.
		want := make([]token, len(tc.want))
		for i, x := range tc.want {
			switch x := x.(type) {
			case rune:
				want[i] = token(x)
			case string:
				ts, err := c.idents.byStr([]byte(x))
				if err != nil {
					t.Fatal(err)
				}
				want[i] = ts.tok
			case token:
				want[i] = x
			default:
				t.Fatalf("invalid type %T", x)
			}
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("parsing %q:\ngot  %v\nwant %v", tc.desc, got, want)
		}
	}
}
