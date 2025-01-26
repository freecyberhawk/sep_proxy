package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	appVersion   = "1.1.0"
	serverPort   = ":8080"
	bankDomain   = "https://sep.shaparak.ir"
	readTimeout  = 10 * time.Second
	writeTimeout = 15 * time.Second
	idleTimeout  = 30 * time.Second
)

func loadPublicKey(pubKeyPath string) (*rsa.PublicKey, error) {
	pubKeyBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	formattedPubKey := strings.TrimSpace(string(pubKeyBytes))
	block, _ := pem.Decode([]byte(formattedPubKey))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block")
	}

	if block.Type != "PUBLIC KEY" && block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA public key")
	}

	if publicKey.N == nil || publicKey.E == 0 {
		return nil, fmt.Errorf("invalid RSA public key")
	}

	return publicKey, nil
}

func verifySignature(publicKey *rsa.PublicKey, data string, signature string) error {
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid base64 signature: %w", err)
	}

	hash := sha256.New()
	hash.Write([]byte(data))
	hashedData := hash.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashedData, signatureBytes)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

func showServerInfo() {
	banner := figure.NewFigure("Free Cyber Hawk", "small", true)

	c := color.New(color.FgHiCyan).Add(color.Bold)
	c.Println(banner.String())

	fmt.Println("Application Version:", appVersion)

	hostInfo, err := host.Info()
	if err != nil {
		log.Printf("Failed to retrieve host information: %v", err)
		return
	}

	fmt.Println("\nServer Information:")
	fmt.Printf("  Hostname: %s\n", hostInfo.Hostname)
	fmt.Printf("  OS: %s %s\n", hostInfo.Platform, hostInfo.PlatformVersion)
	fmt.Printf("  Uptime: %d seconds\n", hostInfo.Uptime)

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Failed to retrieve memory information: %v", err)
		return
	}

	fmt.Println("\nMemory Information:")
	fmt.Printf("  Total: %.2f GB\n", float64(memInfo.Total)/1e9)
	fmt.Printf("  Used: %.2f GB\n", float64(memInfo.Used)/1e9)
	fmt.Printf("  Free: %.2f GB\n", float64(memInfo.Free)/1e9)

	hiCyan := color.New(color.FgHiCyan).Add(color.Bold)
	hiCyan.Printf("\nServer Status: Started\n")
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	publicKey, err := loadPublicKey("public_key.pem")
	if err != nil {
		http.Error(w, "Public key error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Parse the JSON into a generic map
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract "sec" and "secval" safely
	sec, okSec := data["sec"].(string)
	secval, okSecVal := data["secval"].(string)

	if !okSec || !okSecVal {
		http.Error(w, "sec and secval must be strings", http.StatusBadRequest)
		return
	}

	if sec == "" || secval == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// Verify the signature using the public key
	err = verifySignature(publicKey, secval, sec)
	if err != nil {
		http.Error(w, "Signature verification failed", http.StatusUnauthorized)
		return
	}

	// Remove sec and secval from the request data before forwarding
	delete(data, "sec")
	delete(data, "secval")

	// Convert the modified map back to JSON
	modifiedBody, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error generating modified body", http.StatusInternalServerError)
		return
	}

	// Construct target URL with proper validation
	targetURL := bankDomain + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Create new request with context and timeout, using modifiedBody
	ctx, cancel := context.WithTimeout(r.Context(), readTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method, targetURL, bytes.NewReader(modifiedBody))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Forward headers (excluding sensitive ones)
	for name, values := range r.Header {
		if !strings.HasPrefix(strings.ToLower(name), "x-") {
			continue // Only allow custom headers
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Configure HTTP client with timeout settings
	client := &http.Client{
		Timeout: readTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     idleTimeout,
			DisableCompression:  false,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}

	// Send the request to the target server
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach target server", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Forward response headers and body
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err = io.Copy(w, resp.Body); err != nil {
		log.Printf("Response body copy error: %v", err)
		http.Error(w, "Response transfer failed", http.StatusInternalServerError)
	}
}

func main() {
	// Configure and start the HTTP server
	server := &http.Server{
		Addr:         serverPort,
		Handler:      http.HandlerFunc(proxyHandler),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	showServerInfo()
	log.Printf("Proxy server running on %s", serverPort)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server startup failed: %v", err)
	}
}
