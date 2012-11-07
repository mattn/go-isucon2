ifeq ($(OS),Windows_NT)
EXT=.exe
else
EXT=
endif

all: isucon2$(EXT)

isucon2$(EXT) : isucon2.go
	go get
	go build -o isucon2$(EXT) isucon2.go
