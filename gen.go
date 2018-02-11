//+build ignore

package main

import (
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	head, body1 := makeType("Uint64", "uint64")
	_, body2 := makeType("Pointer", "unsafe.Pointer")
	contents := "// CODE GENERATED; DO NOT EDIT\n\n" +
		head + "\n" + body1 + "\n" + body2
	fi, err := os.Stat("chan.go")
	must(err)
	must(ioutil.WriteFile("chan_types.go", []byte(contents), fi.Mode()))
}

func makeType(sig, typ string) (head, body string) {
	data, err := ioutil.ReadFile("chan.go")
	must(err)
	s := string(data)
	s = strings.Replace(s, "Chan", "Chan"+sig, -1)
	s = strings.Replace(s, "sleepN", "sleep"+sig, -1)
	s = strings.Replace(s, "nodeT", "node"+sig, -1)
	s = strings.Replace(s, "interface{}", typ, -1)
	var tgen bool
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "//go:generate") {
			tgen = true
		} else {
			if tgen {
				body += line + "\n"
			} else {
				head += line + "\n"
			}
		}
	}
	return head, body
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
