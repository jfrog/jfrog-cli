package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"math/big"
	"net"
	"os"
	"time"
)

const CertFile = "naive_proxy_cert.pem"
const KeyFile = "naive_proxy_key.pem"

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

func CreateNewCert(absPathCert, absPathKey string) (err error) {
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if errorutils.CheckError(err) != nil {
		return err
	}
	certTemplate := createCertTemplate()
	derBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &rootKey.PublicKey, rootKey)
	if errorutils.CheckError(err) != nil {
		return err
	}
	certOut, err := os.Create(absPathCert)
	if errorutils.CheckError(err) != nil {
		return err
	}
	defer func() {
		e := certOut.Close()
		if err == nil {
			err = e
		}
	}()
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if errorutils.CheckError(err) != nil {
		return err
	}

	keyOut, err := os.OpenFile(absPathKey, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if errorutils.CheckError(err) != nil {
		return err
	}
	defer func() {
		e := keyOut.Close()
		if err == nil {
			err = e
		}
	}()
	return pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rootKey)})
}
