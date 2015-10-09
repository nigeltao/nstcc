package nstcc

import (
	"reflect"
	"testing"
)

func TestNext(t *testing.T) {
	// TODO: s/XXX/42/
	c := newCompiler(nil, nil, []byte("int main(int argc, char** argv) { return XXX; }"))

	ident := func(s string) token {
		ts, err := c.idents.byStr([]byte(s))
		if err != nil {
			t.Fatal(err)
		}
		return ts.tok
	}

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

	want := []token{
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
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\ngot  %v\nwant %v", got, want)
	}
}
