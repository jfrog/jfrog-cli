// +build !windows

package utils

import (
	"crypto/x509"
)

func LoadSystemRoots() (*x509.CertPool, error) {
	return x509.SystemCertPool()
}