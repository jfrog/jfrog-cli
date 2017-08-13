package certificate

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"time"
	"net"
	"crypto/rsa"
	"crypto/rand"
	"math/big"
	"os"
	"encoding/pem"
)

const CERT_FILE = "naive_proxy_cert.pem"
const KEY_FILE = "naive_proxy_key.pem"

func createCertTemplate() *x509.Certificate {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(err)
	}
	return &x509.Certificate{
		Subject:               pkix.Name{Organization: []string{"Test Inc."}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		SerialNumber:          serialNumber,
	}
}

func CreateNewCert() {
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	certTemplate := createCertTemplate()
	derBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		panic(err)
	}
	certOut, err := os.Create(CERT_FILE)
	if err != nil {
		panic(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(KEY_FILE, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rootKey)})
	keyOut.Close()
}