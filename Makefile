.PHONY: force

streamdeck-goui: force
	go build --ldflags '-extldflags "-Wl,--allow-multiple-definition"'
