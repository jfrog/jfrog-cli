package main

import "fmt"

func _pocVetFail() {
	fmt.Printf("%d", "Sahar, MixBanana test.") // go vet should complain about %d with string
}
