package nbsoup

import (
	"fmt"
	"io/ioutil"
	"log"
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
	divList, err := FindAll(root, `span[class="inline-hed" & @content*="code"]`)
	if err != nil {
		log.Fatal(err)
	}
	for _, div := range divList {
		fmt.Println(div.Content)
	}
}
