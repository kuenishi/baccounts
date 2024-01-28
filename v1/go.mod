module github.com/kuenishi/baccounts

go 1.19

require (
	github.com/atotto/clipboard v0.1.2
	github.com/google/subcommands v1.2.0
	golang.org/x/crypto v0.17.0
)

require (
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.15.0 // indirect
)

replace github.com/kuenishi/baccounts => ./v1
