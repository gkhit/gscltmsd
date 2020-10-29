rmdir /S /Q bin
mkdir bin
go env -w GOPRIVATE=github.com/gkhit
go build -ldflags "-s -w" -o bin\gscltmsd.exe main.go
