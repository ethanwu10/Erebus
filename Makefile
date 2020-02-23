
.PHONY: proto
proto: shared/proto/*.proto
	$(MAKE) -C broker proto
	$(MAKE) -C broker-control-cli proto
	$(MAKE) -C client/python proto
	$(MAKE) -C wb-controllers/erebus-robot-controller
