// Package main implements the staticlint multichecker for linkshrtservice
//
// # Analyzers
//
// The multichecker includes:
//
//   - the standard golang.org/x/tools/go/analysis passes used by go vet to
//     detect common correctness, API usage, concurrency, formatting, and
//     compatibility errors
//   - every Staticcheck SA analyzer for detecting correctness problems
//   - S1000 for suggesting simpler and more idiomatic Go constructs
//   - ST1005 for checking the format of error strings
//   - bodyclose for finding HTTP response bodies that are not closed
//   - nilerr for finding code that returns nil after checking a non-nil error
//   - noosexit, the project analyzer that forbids os.Exit calls in main
//
// # Running
//
// Run the checker for every package in the project from the repository root
//
//	go run ./cmd/staticlint ./...
//
// The checker can also be built and passed to go vet
//
//	go build -o staticlint ./cmd/staticlint
//	go vet -vettool=./staticlint ./...
package main
