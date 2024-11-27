Mac/ Linux
go build -ldflags="-s -w" -o GoQu_Mac app.go

Windows
go build -ldflags -H=windowsgui -o GoQu_Windows app.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-s -w -H=windowsgui" -o GoQu_Windows.exe app.go
