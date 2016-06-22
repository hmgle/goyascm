TARGET=goyascm

all:: $(TARGET)

goyascm: parser.go yascm.go
	go build

parser.go: parser.y
	go generate

clean::
	-rm -f parser.go goyascm
