package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestRunServerHTTPS(t *testing.T) {
	tlsConfig, err := newTLSConfig("127.0.0.1:0")
	if err != nil {
		t.Fatalf("newTLSConfig() error = %v", err)
	}
	certificate, err := x509.ParseCertificate(tlsConfig.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(certificate)

	addresses := make(chan string, 1)
	server := &http.Server{
		Addr:      "127.0.0.1:0",
		TLSConfig: tlsConfig,
		BaseContext: func(listener net.Listener) context.Context {
			addresses <- listener.Addr().String()
			return context.Background()
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil {
				t.Error("request has no TLS connection state")
			}
			w.WriteHeader(http.StatusNoContent)
		}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Cleanup(func() { _ = server.Close() })
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- runServer(ctx, server) }()

	var address string
	select {
	case address = <-addresses:
	case err := <-serverErrors:
		t.Fatalf("runServer() exited before startup: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("server startup timed out")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	response, err := client.Get("https://" + address)
	if err != nil {
		t.Fatalf("HTTPS request: %v", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}
	cancel()
	select {
	case err := <-serverErrors:
		if err != nil {
			t.Fatalf("runServer() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server shutdown timed out")
	}
}

func TestNewTLSConfigInvalidAddress(t *testing.T) {
	if _, err := newTLSConfig("invalid"); err == nil {
		t.Fatal("newTLSConfig() error = nil, want error")
	}
}
