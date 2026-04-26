package certbootstrap

import (
	"errors"
	"fmt"
	"os"
	"runtime"
)

var ErrSkipBootstrap = errors.New("skip bootstrap")

func Preflight(args []string) error {
	fmt.Fprintf(os.Stderr, "[certbootstrap] preflight entered goos=%s args=%q\n", runtime.GOOS, args)

	if runtime.GOOS != "windows" {
		fmt.Fprintln(os.Stderr, "[certbootstrap] skip: non-windows")
		return ErrSkipBootstrap
	}

	targetURL, err := ResolveTargetURL(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[certbootstrap] resolve url error: %v\n", err)
		return err
	}
	if targetURL == "" {
		fmt.Fprintln(os.Stderr, "[certbootstrap] skip: no target url")
		return ErrSkipBootstrap
	}

	fmt.Fprintf(os.Stderr, "[certbootstrap] resolved target url=%s\n", targetURL)

	chain, trusted, err := GetVerifiedChainIfWindowsTrusts(targetURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[certbootstrap] verified probe error: %v\n", err)
		return err
	}
	if !trusted {
		fmt.Fprintln(os.Stderr, "[certbootstrap] target not trusted by windows/go")
		return ErrSkipBootstrap
	}

	fmt.Fprintf(os.Stderr, "[certbootstrap] verified chain length=%d\n", len(chain))
	for i, cert := range chain {
		fmt.Fprintf(
			os.Stderr,
			"[certbootstrap] chain[%d] subject=%s issuer=%s isCA=%v\n",
			i,
			cert.Subject.String(),
			cert.Issuer.String(),
			cert.IsCA,
		)
	}


jfrogHome := ResolveJfrogHomeDir()
	fmt.Fprintf(os.Stderr, "[certbootstrap] jfrog home=%s\n", jfrogHome)

	written, err := PersistVerifiedChain(jfrogHome, chain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[certbootstrap] persist error: %v\n", err)
		return err
	}

	fmt.Fprintf(os.Stderr, "[certbootstrap] persisted cert count=%d\n", written)
	return ErrSkipBootstrap
}