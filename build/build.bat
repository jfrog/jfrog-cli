set CGO_ENABLED=0
go build -o jfrog.exe -ldflags "-w -extldflags -static" main.go
