
.PHONY: proto
proto: shared/proto/*.proto
	$(MAKE) -C broker proto
	$(MAKE) -C broker-control-cli proto
	$(MAKE) -C client/python proto
	$(MAKE) -C wb-controllers/erebus-robot-controller proto
	$(MAKE) -C wb-controllers/erebus-supervisor-controller proto

.PHONY: test
test:
	$(MAKE) -C broker test
	$(MAKE) -C client/python test
