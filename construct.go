package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/knakk/kbp/rdf"
	"github.com/knakk/kbp/sparql"
	"github.com/teris-io/shortid"
)

const queries = `
# tag: allByClass
SELECT DISTINCT ?uri WHERE {
	GRAPH <old_deichman> {
		?uri a <{{.Class}}> .
	}
} LIMIT 100

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
		<specification> ?specification .
} WHERE {
	{{.Old}} a deich:Place ;
	deich:prefLabel ?name .
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
	OPTIONAL { {{.Old}} deich:specification ?specification }
}

# tag: constructEvent
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	{{.New}} a <Event> ;
		<name> ?name ;
		<altName> ?altName ;
		<ordinal> ?ordinal ;
		<year> ?year ;
		<note> ?note ;
		<place> ?place ;
		<specification> ?specification .
} WHERE {
	{{.Old}} a deich:Event ;
	deich:prefLabel ?name .
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
	OPTIONAL { {{.Old}} deich:specification ?specification }
	OPTIONAL { {{.Old}} deich:date ?dateStr . FILTER REGEX(?dateStr, "^\\d\\d\\d\\d$") . BIND(xsd:integer(?dateStr) AS ?year) }
	OPTIONAL { {{.Old}} deich:ordinal ?ordinalStr . BIND(xsd:integer(?ordinalStr) AS ?ordinal)}
	OPTIONAL { {{.Old}} deich:place ?placeOld . BIND(IRI(STRAFTER(STR(?placeOld), "http://data.deichman.no/")) AS ?place) }
}

# tag: constructInstrument
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	{{.New}} a <Instrument> ;
		<name> ?name ;
		<altName> ?altName .
} WHERE {
	{{.Old}} a deich:Instrument ;
	deich:prefLabel ?name .
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
}

# tag: constructGenre
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	{{.New}} a <Genre> ;
		<name> ?name ;
		<altName> ?altName ;
		<specification> ?specification .
} WHERE {
	{{.Old}} a deich:Genre ;
	deich:prefLabel ?name .
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
	OPTIONAL { {{.Old}} deich:specification ?specification }
}

# tag: constructSubject
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	{{.New}} a <Subject> ;
		<name> ?name ;
		<altName> ?altName ;
		<specification> ?specification .
} WHERE {
	{{.Old}} a deich:Subject ;
	deich:prefLabel ?name .
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
	OPTIONAL { {{.Old}} deich:specification ?specification }
}

# tag: constructCorporation
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	{{.New}} a <Corporation> ;
		<name> ?name ;
		<altName> ?altName ;
		<nationality> ?nationality ;
		<subdivision> ?subdivision ;
		<specification> ?specification .
} WHERE {
	{{.Old}} a deich:Corporation ;
	deich:name ?name .
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
	OPTIONAL { {{.Old}} deich:specification ?specification }
	OPTIONAL { {{.Old}} deich:nationality ?nationality }
	OPTIONAL { {{.Old}} deich:subdivision ?subdivision }
}

# tag: constructCompositionType
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	{{.New}} a <CompositionType> ;
		<name> ?name ;
		<altName> ?altName ;
		<specification> ?specification .
} WHERE {
	{{.Old}} a deich:CompositionType ;
	deich:prefLabel ?name .
	OPTIONAL { {{.Old}} deich:alternativeName ?altName }
	OPTIONAL { {{.Old}} deich:specification ?specification }
}

# tag: constructPublication
PREFIX deich: <http://data.deichman.no/ontology#>
WITH <old_deichman>
CONSTRUCT {
	?np a <Publication> ;
		<title> ?title ;
		<subtitle> ?subtitle ;
		<partTitle> ?partTitle ;
		<partNumber> ?partNumber ;
		<isbn> ?isbn ;
		<manifestationOf> {{.New}} ;
		<contrib> ?pubContrib .
	{{.New}} a <Work> ;
		<subject> ?subject ;
		<genre> ?genre ;
		<contrib> ?workContrib .
	?workContrib a <Contribution>, ?workMainEntry ;
		<role> ?workContribRole ;
		<agent> ?workContribAgent .
	?pubContrib a <Contribution> ;
		<role> ?pubContribRole ;
		<agent> ?pubContribAgent .
} WHERE {
	{{.Old}} a deich:Publication ;
	deich:mainTitle ?name ;
	deich:publicationOf ?work .
	BIND(IRI(STRAFTER(STR({{.Old}}), "http://data.deichman.no/")) AS ?np)
	OPTIONAL {
		?work deich:contributor ?wc .
		?wc deich:agent ?wAgent ;
			deich:role ?wRole .
		OPTIONAL { ?wc a deich:MainEntry . BIND(IRI("MainEntry") AS ?workMainEntry)}
		BIND(IRI(CONCAT("contrib/", SHA1(CONCAT(?work, ?wAgent, ?wRole, ?workMainEntry)))) AS ?workContrib)
		BIND(IRI(CONCAT("role/", STRAFTER(STR(?wRole), "http://data.deichman.no/role#"))) AS ?workContribRole)
		BIND(IRI(STRAFTER(STR(?wAgent), "http://data.deichman.no/")) AS ?workContribAgent)
	}
	OPTIONAL {
		{{.Old}} deich:contributor ?pc .
		?pc deich:agent ?pAgent ;
			deich:role ?pRole .
		BIND(IRI(CONCAT("contrib/", SHA1(CONCAT({{.Old}}, ?pAgent, ?pRole)))) AS ?pubContrib)
		BIND(IRI(CONCAT("role/", STRAFTER(STR(?pRole), "http://data.deichman.no/role#"))) AS ?pubContribRole)
		BIND(IRI(STRAFTER(STR(?pAgent), "http://data.deichman.no/")) AS ?pubContribAgent)
	}
	OPTIONAL { {{.Old}} deich:subtitle ?subtitle }
	OPTIONAL { {{.Old}} deich:partTitle ?partTitle }
	OPTIONAL { {{.Old}} deich:partNumber ?partNumber }
	OPTIONAL { {{.Old}} deich:isbn ?isbn }
	OPTIONAL { ?work deich:subject ?subjectTemp . BIND(IRI(STRAFTER(STR(?subjectTemp), "http://data.deichman.no/")) AS ?subject)}
	OPTIONAL { ?work deich:genre ?genreTemp . BIND(IRI(STRAFTER(STR(?genreTemp), "http://data.deichman.no/")) AS ?genre)}
}
`

func main() {
	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	if err != nil {
		log.Fatal(err)
	}
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
		newURI := rdf.NewNamedNode(strings.TrimPrefix(oldURI.(rdf.NamedNode).Name(), "http://data.deichman.no/"))
		if class == "Publication" {
			id, err := sid.Generate()
			if err != nil {
				log.Fatal(err)
			}
			newURI = rdf.NewNamedNode("work/" + id)
		}
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

		fmt.Fprintf(os.Stderr, "%d/%d   \r", i, len(entities))
	}

}
