package nbsoup

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()
	f, err := os.Open("test.html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	html, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	root, err := Parse(html)
	if err != nil {
		log.Fatal(err)
	}
	fontList, err := root.FindAll(`div[align="center"]`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(fontList))
}
