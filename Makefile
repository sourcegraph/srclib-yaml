ifeq (${OS},Windows_NT)
	EXE := .bin/srclib-yaml.exe
else
	EXE := .bin/srclib-yaml
endif

.PHONY: install clean

default: govendor install

install: ${EXE}

clean:
	rm -f ${EXE}

govendor:
	go get github.com/karidanos/govendor
	govendor sync

${EXE}: $(shell /usr/bin/find . -type f -and -name '*.go' -not -path './vendor/*')
	go build -o ${EXE}
