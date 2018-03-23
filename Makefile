.PHONY: clean_virtuoso

all: transform

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

import: fusekidump.nt.gz
	docker cp ./fusekidump.nt.gz virtuoso:/data/
	$(call import_graph,fusekidump.nt.gz,old_deichman)

transform: import