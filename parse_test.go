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
	trs, err := root.FindAll(`table[id="proxylisttable"].tbody.tr`)
	if err != nil {
		log.Fatal(err)
	}
	for _, tr := range trs {
		fmt.Println(tr.Name)
	}
	fmt.Println(len(trs))
}
