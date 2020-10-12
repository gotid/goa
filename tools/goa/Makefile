version := $(shell /bin/date "+%Y-%m-%d %H:%M")

mac:
	GOOS=darwin go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o goa goa.go
	command -v upx &> /dev/null && upx goa