
.PHONY: proto
proto: shared/proto/*.proto
	$(MAKE) -C broker proto
	$(MAKE) -C broker-control-cli proto
	$(MAKE) -C client proto
