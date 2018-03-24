package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/knakk/kbp/rdf"
	"github.com/knakk/kbp/sparql"
)

const queries = `
# tag: allByClass
SELECT DISTINCT ?uri WHERE {
	GRAPH <old_deichman> {
		?uri a <{{.Class}}> .
	}
}

# tag: renameURI
DELETE { ?s ?p <{{.Old}}> }
INSERT { ?s ?o <{{.New}}> }
WHERE  { ?s ?p <{{.Old}}> }
;
DELETE { <{{.Old}}> ?p ?o }
INSERT { <{{.New}}> ?p ?p }
WHERE  { <{{.Old}}> ?p ?o }

# tag: constructPerson
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	{{.New}} a <Person> ;
		<name> ?name ;
		<gender> ?gender ;
		<altName> ?altName ;
		<ordinal> ?ordinal ;
		<specification> ?specification ;
		<birthYear> ?birthYear ;
		<deathYear> ?deathYear ;
		<nationality> ?nationality .
} WHERE {
	{{.Old}} a deich:Person ;
	deich:name ?name .
	OPTIONAL { {{.Old}} deich:gender ?genderStr
			   VALUES (?genderStr ?gender) {
					  ("Mann"   <gender/male>)
					  ("m"      <gender/male>)
					  ("M"      <gender/male>)
					  ("mann"   <gender/male>)
					  ("Mann."  <gender/male>)
					  ("Kv."    <gender/male>)
					  ("Kvinne" <gender/female>)
					  ("F"      <gender/female>)
					  ("kvinne" <gender/female>)
			   }
			  }
	OPTIONAL { {{.Old}} deich:ordinal ?ordinal }
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
	OPTIONAL { {{.Old}} deich:specification ?specification }
	OPTIONAL { {{.Old}} deich:birthYear ?birthYear }
	OPTIONAL { {{.Old}} deich:deathYear ?deathYear }
	OPTIONAL { {{.Old}} deich:nationality ?nationality }
}


# tag: constructPlace
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	{{.New}} a <Place> ;
		<name> ?name ;
		<altName> ?altName ;
		<ordinal> ?ordinal ;
		<specification> ?specification .
} WHERE {
	{{.Old}} a deich:Place ;
	deich:prefLabel ?name .
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
	OPTIONAL { {{.Old}} deich:specification ?specification }

}
`

var idcount int32

func newID() string {
	// base32 encoding alphabet
	const base32 = "0123456789abcdefghjkmnpqrstvwxyz"

	// Base32-encoding of timestamp taken from github.com/oklog/ulid

	// Get current time in Unix milliseconds:
	now := time.Now().UTC()
	ms := uint64(now.Unix())*1000 + uint64(now.Nanosecond()/int(time.Millisecond))

	id := make([]byte, 6)
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)

	dst := make([]byte, 12)

	dst[0] = base32[(id[0]&224)>>5]
	dst[1] = base32[id[0]&31]
	dst[2] = base32[(id[1]&248)>>3]
	dst[3] = base32[((id[1]&7)<<2)|((id[2]&192)>>6)]
	dst[4] = base32[(id[2]&62)>>1]
	dst[5] = base32[((id[2]&1)<<4)|((id[3]&240)>>4)]
	dst[6] = base32[((id[3]&15)<<1)|((id[4]&128)>>7)]
	dst[7] = base32[(id[4]&124)>>2]
	dst[8] = base32[((id[4]&3)<<3)|((id[5]&224)>>5)]
	dst[9] = base32[id[5]&31]

	n := atomic.AddInt32(&idcount, 1) % 1024
	//dst[10] = base32[byte(((n>>16)&124)>>3)]
	dst[10] = base32[byte(((n>>8)&3<<3)|(n&224>>5))]
	dst[11] = base32[byte(n)&31]

	return string(dst)

}

func main() {
	qbank := sparql.LoadBank(bytes.NewBufferString(queries))

	repo, err := sparql.NewRepo("http://192.168.1.39:8890/sparql-auth", sparql.DigestAuth("dba", "dba"))
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 {
		log.Fatal("usage: construt <class>")
	}
	class := strings.Title(os.Args[1])

	q, err := qbank.Prepare(
		"allByClass", struct{ Class string }{"http://data.deichman.no/ontology#" + class})
	if err != nil {
		log.Fatal(err)
	}
	res, err := repo.Select(q)
	if err != nil {
		log.Fatal(err)
	}
	entities := res.Bindings()["uri"]
	for i, oldURI := range entities {
		tries := 0
	retry:
		//newP := rdf.NewNamedNode("person/" + newID())
		newURI := rdf.NewNamedNode(strings.TrimPrefix(oldURI.(rdf.NamedNode).Name(), "http://data.deichman.no/"))
		q, err := qbank.Prepare(
			"construct"+class, struct{ Old, New rdf.Node }{oldURI, newURI})
		if err != nil {
			time.Sleep(1)
			if tries < 3 {
				tries++
				goto retry
			}
			log.Fatal(err)
		}
		g, err := repo.Construct(q)
		if err != nil {
			log.Fatal(err)
		}
		g.EncodeNTriples(os.Stdout)
		//fmt.Println(g.Dot(newP, memory.DotOptions{}))

		fmt.Fprintf(os.Stderr, "%d/%d\r", i, len(entities))
	}

}
