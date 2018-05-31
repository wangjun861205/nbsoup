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
	fontList, err := FindAll(root, `font[@content="Make Model"]`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(fontList))
}
