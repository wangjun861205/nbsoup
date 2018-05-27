# nbmq

## Overview
This is a light weight HTML parser which provide some function to search node in HTML file.

## Usage
1. Parse HTML:
  ```root, err := Parse(html)```
  ```func Parse(html []byte) (*Node, error)``` receive a []byte and return a ```*Node```, if something wrong, it will return a ```nil```
  and a non-nil ```error```.
2. Search Node:
  ```n, err := FindAll(root, `div[id="app" | class*="bb"]`)```
  ```func FindAll(n *Node, queryStr string) ([]*Node, error)``` receive a ```*Node``` as start position, a query string and return
  a ```[]*Node``` if success, else it will return a ```nil``` and a ```error```.
  Note: if no ```*Node``` match your query, the ```[]*Node``` returned will be ```nil```.This is only convenience for check result.

## Query String

### Attribute Query
1. ```div[class="your class"]```
  This will find all ```div``` nodes which class **is equal to** "your class".
2. ```div[class="your class" & id="your id"]
  This will find all ```div``` nodes which class **is equal to** "your class" **and** id **is equal to** "your id"
3. ```div[class="your class" | id="your id"]
  This will find all ```div``` nodes which class **is equal to** "your class" **or** id **is equal to** "your id"
4. ```div[class!="your class"]```
  This will find all ```div``` nodes which class **is not equal to** "your class"
5. ```div[class*="your class"]```
  This will find all ```div``` nodes which class **contains** "your class"
6. ```div[class!*="your class"]```
  This will find all ```div``` nodes which class **not contains** "your class"
7. ```div[class%="your regexp"]```
  This will find all ```div``` nodes which class **match** "your regexp"
8. ```div[class="your class"].h1[id="your id"]```
  This will find all ```h1``` nodes which ```id``` **is equal to** "your id" and which parent node is a ```div``` which class
  **is equal to** "your class"

### Content Query
Content query is all the same as attribute query except the attribute name must be ```@content```.
For example ```div[@content*="your content"]```
