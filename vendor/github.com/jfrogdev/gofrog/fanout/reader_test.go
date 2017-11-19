package fanout

import (
	"testing"
	"io"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"crypto/sha1"
	"errors"
)

const input = "yogreshobuddy!"
const sha1sum = "a967c390de10f37dab8eb33549c6304ded62e951"
const sha2sum = "72a0230d6e5eebb437a9069ebb390171284192e9a993938d02cb0aaae003fd1c"

var (
	inputBytes = []byte(input)
)

func TestFanoutRead(t *testing.T) {
	proc := func(r io.Reader) (interface{}, error) {
		hash := sha256.New()
		if _, err := io.Copy(hash, r); err != nil {
			t.Fatal(t)
		}
		return hash.Sum(nil), nil
	}

	//Using a closure argument instead of results
	var sum3 []byte
	proc1 := func(r io.Reader) (rt interface{}, er error) {
		hash := sha256.New()
		if _, err := io.Copy(hash, r); err != nil {
			t.Fatal(t)
		}
		sum3 = hash.Sum(nil)
		return
	}

	r := bytes.NewReader(inputBytes)
	fr := NewReadAllReader(r, ReadAllConsumerFunc(proc), ReadAllConsumerFunc(proc), ReadAllConsumerFunc(proc1))
	results, err := fr.ReadAll()

	if err != nil {
		t.Error(err)
	}
	sum1 := results[0].([]byte)
	sum2 := results[1].([]byte)

	sum1str := hex.EncodeToString(sum1)
	sum2str := hex.EncodeToString(sum2)
	sum3str := hex.EncodeToString(sum3)

	if !(sum1str == sum2str && sum1str == sum3str) {
		t.Errorf("Sum1 %s and sum2 %s and sum3 %s are not the same", sum1str, sum2str, sum3str)
	}

	if sum1str != sha2sum {
		t.Errorf("Checksum is not as expected: %s != %s", sum1str, sha2sum)
	}
}

func TestFanoutProgressiveRead(t *testing.T) {
	hash1 := sha1.New()
	proc1 := func(p []byte) (err error) {
		if _, err := hash1.Write(p); err != nil {
			t.Fatal(t)
		}
		return
	}

	hash2 := sha256.New()
	proc2 := func(p []byte) (err error) {
		if _, err := hash2.Write(p); err != nil {
			t.Fatal(t)
		}
		return
	}

	r := bytes.NewReader(inputBytes)
	pfr := NewReader(r, ConsumerFunc(proc1), ConsumerFunc(proc2))
	defer pfr.Close()

	_, err := ioutil.ReadAll(pfr)
	if err != nil {
		t.Error(err)
	}

	sum1 := hash1.Sum(nil)
	sum1str := hex.EncodeToString(sum1)
	if sum1str != sha1sum {
		t.Errorf("Sha1 is not as expected: %s != %s", sum1str, sha1sum)
	}
	sum2 := hash2.Sum(nil)
	sum2str := hex.EncodeToString(sum2)
	if sum2str != sha2sum {
		t.Errorf("Sha2 is not as expected: %s != %s", sum2str, sha2sum)
	}
}

func TestFanoutProgressiveReadError(t *testing.T) {
	const errmsg = "ERRSHA1"

	hash1 := sha1.New()
	proc1 := func(p []byte) (err error) {
		return errors.New(errmsg)
	}

	hash2 := sha256.New()
	proc2 := func(p []byte) (err error) {
		if _, err := hash2.Write(p); err != nil {
			t.Fatal(t)
		}
		return
	}

	r := bytes.NewReader(inputBytes)
	pfr := NewReader(r, ConsumerFunc(proc1), ConsumerFunc(proc2))
	defer pfr.Close()

	_, err := ioutil.ReadAll(pfr)
	if err == nil {
		t.Fatal("Expected a non-nil error")
	}
	if err.Error() != errmsg {
		t.Fatalf("Error message is different from: %s", errmsg)
	}

	sum1 := hash1.Sum(nil)
	sum1str := hex.EncodeToString(sum1)
	if sum1str == sha1sum {
		t.Errorf("Sha1 is not as expected: %s != %s", sum1str, sha1sum)
	}
	var sum2str string
	if err == nil {
		sum2 := hash2.Sum(nil)
		sum2str = hex.EncodeToString(sum2)
	}
	if sum2str == sha2sum {
		t.Error("Sha2 calculation should have terminated a head of time due to an error")
	}
}
