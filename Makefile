TARGET=goyascm

all:: $(TARGET)

goyascm: parser.go yascm.go
	go build

parser.go: parser.y
	go tool yacc -p "scm" -o parser.go parser.y

clean::
	-rm -f parser.go
