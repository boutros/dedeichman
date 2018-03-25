.PHONY: clean_virtuoso import transform

all: construct

define clear_graph
	docker exec -i virtuoso bash -c "cd /data && echo -e \"log_enable(3,1);\nSPARQL DROP SILENT GRAPH <$(1)>;\ncheckpoint;\" | isql"
endef

define import_graph
	docker exec -i virtuoso bash -c "cd /data && echo -e \"DELETE FROM DB.DBA.load_list;\nld_dir('/data', '$(1)', '$(2)');\nrdf_loader_run();\ncheckpoint;\" | isql"
endef

fusekidump.nt.gz:
	rm -f fusekidump.nt.gz
	wget https://static.deichman.no/fusekidump.nt.gz

clean_virtuoso:
	$(call clear_graph,http://www.openlinksw.com/schemas/virtrdf#)
	$(call clear_graph,http://www.w3.org/ns/ldp#)
	$(call clear_graph,http://localhost:8890/sparql)
	$(call clear_graph,http://localhost:8890/DAV/)
	$(call clear_graph,http://www.w3.org/2002/07/owl#)
	$(call clear_graph,old_deichman)
	$(call clear_graph,new_deichman)

all.nt.gz: fusekidump.nt.gz
	zcat fusekidump.nt.gz | grep -v "migration.deichman.no" | grep -v "deichman.no/raw" > all.nt
	gzip all.nt

import: all.nt.gz clean_virtuoso
	docker cp ./all.nt.gz virtuoso:/data/
	$(call import_graph,all.nt.gz,old_deichman)
	docker cp ./resources.ttl virtuoso:/data/
	$(call import_graph,resources.ttl,new_deichman)

CLASSES=place event genre subject instrument corporation person
construct:
	for class in $(CLASSES); do echo $$class; go run construct.go $$class > $$class.nt; done
