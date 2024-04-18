module github.com/lsherman98/boot.dev/hellogo

go 1.22.1

replace "github.com/lsherman98/boot.dev/mystrings" v0.0.0 => "../mystrings"

require (
    "github.com/lsherman98/boot.dev/mystrings" v0.0.0
)