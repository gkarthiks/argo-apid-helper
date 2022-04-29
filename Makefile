CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist

apid-build:
	go build -o ${DIST_DIR}/apid .

apid-build-linux:
	make clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make apid-build

apid-build-mac:
	make clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 make apid-build

clean:
	rm -f ${DIST_DIR}/apid