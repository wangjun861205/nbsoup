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
	makeModel, err := root.FindAll(`font[@content*="Make Model"]`)
	if err != nil {
		log.Fatal(err)
	}
	tableNode := makeModel[0]
	for tableNode != nil && tableNode.Name != "table" {
		tableNode = tableNode.Parent
	}
	trs, _ := FindAll(tableNode, `tr`)
	for _, tr := range trs {
		tds, _ := tr.FindAll(`td`)
		fmt.Println(len(tds))
	}

}
