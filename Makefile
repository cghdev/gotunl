TARGET=./build
ARCHS=amd64
LDFLAGS="-s -w"

current:
	@go build -ldflags=${LDFLAGS} -o ./gotunl; \
	echo "Done."

windows:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for windows $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gotunl-windows-$${GOARCH} ; \
		GOOS=windows GOARCH=$${GOARCH} go build -trimpath -ldflags=${LDFLAGS} -o ${TARGET}/gotunl-windows-$${GOARCH}/gotunl.exe ; \
	done; \
	echo "Done."

linux:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for linux $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gotunl-linux-$${GOARCH} ; \
		GOOS=linux GOARCH=$${GOARCH} go build -trimpath -ldflags=${LDFLAGS} -o ${TARGET}/gotunl-linux-$${GOARCH}/gotunl ; \
	done; \
	echo "Done."

darwin:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for darwin $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gotunl-darwin-$${GOARCH} ; \
		GOOS=darwin GOARCH=$${GOARCH} go build -trimpath -ldflags=${LDFLAGS} -o ${TARGET}/gotunl-darwin-$${GOARCH}/gotunl ; \
	done; \
	echo "Done."

all: darwin linux windows

test:
	@go test -v -race ./... ; \
	echo "Done."

clean:
	@rm -rf ${TARGET}/* ; \
	echo "Done."
