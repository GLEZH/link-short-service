// Package main implements the staticlint multichecker for linkshrtservice
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
