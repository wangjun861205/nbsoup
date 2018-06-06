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
	makeModel, _ := root.FindAll(`font[@content*="Make Model"]`)
	table := makeModel[0]
	for table != nil && table.Name != "table" {
		table = table.Parent
	}
	trs, _ := table.FindAll(`tr`)
	for _, tr := range trs {
		fonts, _ := tr.FindAll(`font`)
		var text string
		for _, font := range fonts {
			text += font.Content + ","
		}
		fmt.Println(text)
	}
}
