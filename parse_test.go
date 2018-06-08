package nbsoup

import (
	"fmt"
	"io/ioutil"
	"log"
	_ "net/http/pprof"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
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
	makeModels, _ := root.FindAll(`font[@content*="Make Model"]`)
	makeModel := makeModels[0]
	tableNode := makeModel.Parent
	for tableNode != nil && tableNode.Name != "table" {
		tableNode = tableNode.Parent
	}
	trs, _ := tableNode.FindAll(`tr`)
	fmt.Println("row numbers:", len(trs))
	for _, tr := range trs {
		tds, _ := tr.FindAll("td")
		if len(tds) != 2 {
			continue
		}
		fmt.Println(tds[0].GetAllContent(), ":", tds[1].GetAllContent())
	}
}
