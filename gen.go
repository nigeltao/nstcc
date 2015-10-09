// +build ignore

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var debug = flag.Bool("debug", false, "")

func main() {
	flag.Parse()

	f, err := os.Open("builtins.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var (
		tokens [][2]string
		values []byte
	)
	r := bufio.NewScanner(f)
	for r.Scan() {
		s := strings.TrimSpace(r.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		i := strings.IndexByte(s, '\t')
		if i < 0 {
			log.Fatalf("line %q does not contain a \t character", s)
		}
		tokens = append(tokens, [2]string{s[:i], s[i+1:]})
		values = append(values, s[i+1:]...)
	}
	if err := r.Err(); err != nil {
		log.Fatal(err)
	}

	w := new(bytes.Buffer)
	fmt.Fprintf(w, "// generated by \"go run gen.go\". DO NOT EDIT.\n\n"+
		"package nstcc\n\n"+
		"const (\n")
	for i, t := range tokens {
		fmt.Fprintf(w, "%s = tokIdent + %d // %s\n", t[0], i, t[1])
	}
	fmt.Fprintf(w, ")\n\n")
	fmt.Fprintf(w, "var builtInTokensStrings = []byte(%q)\n\n", values)
	fmt.Fprintf(w, "var builtInTokensLengths = [...]uint32{\n")
	sum := 0
	for _, t := range tokens {
		sum += len(t[1])
		fmt.Fprintf(w, "%d, // %s\n", sum, t[1])
	}
	fmt.Fprintf(w, "}\n\n")

	writeIsTable(w, "isID", func(c byte) bool {
		return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || c == '_'
	})
	writeIsTable(w, "isIDNum", func(c byte) bool {
		return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || c == '_' || ('0' <= c && c <= '9')
	})
	writeIsTable(w, "isNum", func(c byte) bool {
		return '0' <= c && c <= '9'
	})
	writeIsTable(w, "isSimpleToken", func(c byte) bool {
		return c == '(' || c == ')' || c == '[' || c == ']' || c == '{' || c == '}' || c == ',' ||
			c == ';' || c == ':' || c == '?' || c == '~' || c == '$' || c == '@'
	})

	if *debug {
		os.Stdout.Write(w.Bytes())
		return
	}
	out, err := format.Source(w.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("table.go", out, 0660); err != nil {
		log.Fatal(err)
	}
}

func writeIsTable(w *bytes.Buffer, name string, f func(c byte) bool) {
	fmt.Fprintf(w, "var %s = [256]bool{\n", name)
	for i := 0; i < 256; i++ {
		fmt.Fprintf(w, "%t,", f(byte(i)))
		if i%8 == 7 {
			fmt.Fprintln(w)
		}
	}
	fmt.Fprintf(w, "}\n\n")
}