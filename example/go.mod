module github.com/haxii/log/example

go 1.22.0

require (
	github.com/haxii/log/v2 v2.6.4
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.32.0
)

replace github.com/haxii/log/v2 => ../

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.18.0 // indirect
)
