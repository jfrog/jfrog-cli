package certbootstrap

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func ResolveJfrogHomeDir() string {
	if v := os.Getenv("JFROG_CLI_HOME_DIR"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".jfrog"
	}
	return filepath.Join(home, ".jfrog")
}

func PersistVerifiedChain(jfrogHome string, chain []*x509.Certificate) (int, error) {
	certsDir := filepath.Join(jfrogHome, "security", "certs")
	if err := os.MkdirAll(certsDir, 0o700); err != nil {
		return 0, err
	}

	fmt.Fprintf(os.Stderr, "[certbootstrap] persisting to cert dir=%s\n", certsDir)

	written := 0

	for i, cert := range chain {
		fmt.Fprintf(
			os.Stderr,
			"[certbootstrap] persist input chain[%d] subject=%s issuer=%s isCA=%v\n",
			i,
			cert.Subject.String(),
			cert.Issuer.String(),
			cert.IsCA,
		)

		var name string
		switch i {
		case 0:
			name = "leaf.pem"
		case 1:
			name = "intermediate.pem"
		case 2:
			name = "root.pem"
		default:
			name = fmt.Sprintf("cert-%d.pem", i)
		}

		path := filepath.Join(certsDir, name)

		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}
		data := pem.EncodeToMemory(block)

		if err := os.WriteFile(path, data, 0o600); err != nil {
			return written, err
		}

		written++
		fmt.Fprintf(os.Stderr, "[certbootstrap] wrote %s\n", path)
	}

	return written, nil
}