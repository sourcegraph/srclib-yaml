ifeq (${OS},Windows_NT)
	EXE := .bin/srclib-yaml.exe
else
	EXE := .bin/srclib-yaml
endif

.PHONY: install clean

install: ${EXE}

clean:
	rm -f ${EXE}

${EXE}:
	go build -o ${EXE}
