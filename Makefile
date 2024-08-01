.PHONY: test
test:
	cd test && go run main.go -dir=${DIR} -run=${RUN}
