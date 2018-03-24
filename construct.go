package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/knakk/kbp/sparql"
)

const queries = `
# tag: allByClass
SELECT DISTINCT ?uri WHERE {
	GRAPH <old_deichman> {
		?uri a <{{.Type}}> .
	}
}

`

func main() {
	qbank := sparql.LoadBank(bytes.NewBufferString(queries))

	repo, err := sparql.NewRepo("http://192.168.1.39:8890/sparql-auth", sparql.DigestAuth("dba", "dba"))
	if err != nil {
		log.Fatal(err)
	}

	q, err := qbank.Prepare(
		"allByClass", struct{ Type string }{"http://data.deichman.no/ontology#Person"})
	if err != nil {
		log.Fatal(err)
	}
	res, err := repo.Query(q)
	if err != nil {
		log.Fatal(err)
	}

	persons := res.Bindings()["uri"]
	for _, p := range persons[:100] {
		fmt.Println(p)
	}
	fmt.Println(len(persons))

}
