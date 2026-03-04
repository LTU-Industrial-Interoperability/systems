package certgen

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"time"
)


func Generate(certPath, keyPath string, name pkix.Name, appURI string, duration time.Duration) error {
	if _, err := os.Stat(certPath); err == nil {
		return fmt.Errorf("certificate file already exists at %s", certPath)
	}
	if _, err := os.Stat(keyPath); err == nil {
		return fmt.Errorf("key file already exists at %s", keyPath)
	}

	u, _ := url.Parse(appURI)
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	
	cert := &x509.Certificate {
		SerialNumber: serial,
		Subject: name,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(duration),
		URIs: 				   []*url.URL{u},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment,
		BasicConstraintsValid: true,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &certPrivKey.PublicKey, certPrivKey)
	if err != nil {
		return err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	if err := os.WriteFile(certPath, certPEM.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", certPath, err)
	}
	if err := os.WriteFile(keyPath, certPrivKeyPEM.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", keyPath, err)
	}
	return nil
}

func Load(certPath, keyPath string) (cert []byte, key *rsa.PrivateKey, err error) {
	certPEM, err := os.ReadFile(certPath); 
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read certificate file: %v", err)
	}
	keyPEM, err := os.ReadFile(keyPath); 
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read key file: %v", err)
	}
	c, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load key pair: %v", err)
	}
	key, ok := c.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("failed to assert private key type")
	}

	return c.Certificate[0], key, nil
}


func EnsureExists(certPath, keyPath string, name pkix.Name, appURI string) (cert []byte, key *rsa.PrivateKey, err error) {
	cert, key, err = Load(certPath, keyPath)
	if err == nil {
		// Files exist — check expiry
		certParsed, err := x509.ParseCertificate(cert)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
		}
		if certParsed.NotAfter.After(time.Now()) {
			// Still valid, return as-is
			return cert, key, nil
		}
		// Expired — remove old files so Generate can overwrite
		os.Remove(certPath)
		os.Remove(keyPath)
	}

	// No files, or expired: generate fresh cert
	if err := Generate(certPath, keyPath, name, appURI, 365*24*time.Hour); err != nil {
		return nil, nil, fmt.Errorf("failed to generate certificate: %v", err)
	}
	return Load(certPath, keyPath)
}