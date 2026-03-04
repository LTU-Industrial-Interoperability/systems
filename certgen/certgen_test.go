package certgen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var testName = pkix.Name{
	CommonName:   "test-client",
	Organization: []string{"Test Org"},
}

const testAppURI = "urn:test:client"

// writeExpiredCert writes a cert/key pair whose NotAfter is in the past.
// Uses 2048-bit keys to keep test runtime reasonable.
func writeExpiredCert(t *testing.T, certPath, keyPath string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("writeExpiredCert: generate key: %v", err)
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	u, _ := url.Parse(testAppURI)
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      testName,
		NotBefore:    time.Now().Add(-2 * time.Hour),
		NotAfter:     time.Now().Add(-1 * time.Hour), // already expired
		URIs:         []*url.URL{u},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("writeExpiredCert: create cert: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("writeExpiredCert: write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0644); err != nil {
		t.Fatalf("writeExpiredCert: write key: %v", err)
	}
}

// ---------------- Generate ----------------

func TestGenerate_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	if err := Generate(certPath, keyPath, testName, testAppURI, 24*time.Hour); err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if _, err := os.Stat(certPath); err != nil {
		t.Errorf("cert file not created: %v", err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("key file not created: %v", err)
	}
}

func TestGenerate_CertHasCorrectFields(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	duration := 48 * time.Hour
	before := time.Now()
	if err := Generate(certPath, keyPath, testName, testAppURI, duration); err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	certDER, _, err := Load(certPath, keyPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	parsed, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("ParseCertificate: %v", err)
	}

	if parsed.Subject.CommonName != testName.CommonName {
		t.Errorf("CommonName: got %q, want %q", parsed.Subject.CommonName, testName.CommonName)
	}
	if parsed.IsCA {
		t.Error("IsCA should be false for a client cert")
	}
	if parsed.NotBefore.Before(before.Add(-time.Second)) {
		t.Errorf("NotBefore is too early: %v", parsed.NotBefore)
	}
	if parsed.NotAfter.Before(before.Add(duration - time.Second)) {
		t.Errorf("NotAfter is too early: %v", parsed.NotAfter)
	}
	if len(parsed.URIs) != 1 || parsed.URIs[0].String() != testAppURI {
		t.Errorf("URIs: got %v, want [%s]", parsed.URIs, testAppURI)
	}
	if len(parsed.ExtKeyUsage) != 1 || parsed.ExtKeyUsage[0] != x509.ExtKeyUsageClientAuth {
		t.Errorf("ExtKeyUsage: got %v, want [ClientAuth]", parsed.ExtKeyUsage)
	}
	if parsed.KeyUsage&x509.KeyUsageCertSign != 0 {
		t.Error("KeyUsageCertSign should not be set on a client cert")
	}
	if parsed.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		t.Error("KeyUsageDigitalSignature should be set")
	}
}

func TestGenerate_ErrorIfCertExists(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	if err := os.WriteFile(certPath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Generate(certPath, keyPath, testName, testAppURI, 24*time.Hour); err == nil {
		t.Error("expected error when cert file already exists, got nil")
	}
}

func TestGenerate_ErrorIfKeyExists(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	if err := os.WriteFile(keyPath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Generate(certPath, keyPath, testName, testAppURI, 24*time.Hour); err == nil {
		t.Error("expected error when key file already exists, got nil")
	}
}

// ---------------- Load ----------------

func TestLoad_Success(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	if err := Generate(certPath, keyPath, testName, testAppURI, 24*time.Hour); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	cert, key, err := Load(certPath, keyPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(cert) == 0 {
		t.Error("cert bytes are empty")
	}
	if key == nil {
		t.Error("key is nil")
	}
}

func TestLoad_MissingCert(t *testing.T) {
	dir := t.TempDir()
	_, _, err := Load(filepath.Join(dir, "missing.pem"), filepath.Join(dir, "missing.key"))
	if err == nil {
		t.Error("expected error for missing cert file, got nil")
	}
}

func TestLoad_MissingKey(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	if err := Generate(certPath, keyPath, testName, testAppURI, 24*time.Hour); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	os.Remove(keyPath)

	if _, _, err := Load(certPath, keyPath); err == nil {
		t.Error("expected error for missing key file, got nil")
	}
}

// ---------------- EnsureExists ----------------

func TestEnsureExists_CreatesWhenMissing(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	cert, key, err := EnsureExists(certPath, keyPath, testName, testAppURI)
	if err != nil {
		t.Fatalf("EnsureExists returned error: %v", err)
	}
	if len(cert) == 0 {
		t.Error("cert bytes are empty")
	}
	if key == nil {
		t.Error("key is nil")
	}
	if _, err := os.Stat(certPath); err != nil {
		t.Error("cert file was not created on disk")
	}
}

func TestEnsureExists_ReturnsCachedValidCert(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	cert1, _, err := EnsureExists(certPath, keyPath, testName, testAppURI)
	if err != nil {
		t.Fatalf("first EnsureExists: %v", err)
	}
	cert2, _, err := EnsureExists(certPath, keyPath, testName, testAppURI)
	if err != nil {
		t.Fatalf("second EnsureExists: %v", err)
	}

	p1, _ := x509.ParseCertificate(cert1)
	p2, _ := x509.ParseCertificate(cert2)
	if p1.SerialNumber.Cmp(p2.SerialNumber) != 0 {
		t.Error("second call returned a different cert - files were regenerated unexpectedly")
	}
}

func TestEnsureExists_RegeneratesExpiredCert(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	writeExpiredCert(t, certPath, keyPath)

	expiredDER, _, err := Load(certPath, keyPath)
	if err != nil {
		t.Fatalf("load expired cert: %v", err)
	}
	expiredParsed, _ := x509.ParseCertificate(expiredDER)

	cert, key, err := EnsureExists(certPath, keyPath, testName, testAppURI)
	if err != nil {
		t.Fatalf("EnsureExists with expired cert: %v", err)
	}
	if key == nil {
		t.Fatal("key is nil after regeneration")
	}

	newParsed, err := x509.ParseCertificate(cert)
	if err != nil {
		t.Fatalf("parse regenerated cert: %v", err)
	}
	if !newParsed.NotAfter.After(time.Now()) {
		t.Error("regenerated cert is still expired")
	}
	if newParsed.SerialNumber.Cmp(expiredParsed.SerialNumber) == 0 {
		t.Error("serial number unchanged - cert was not regenerated")
	}
}
