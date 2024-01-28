package baccounts

import (
	"log"
	"testing"
)

func TestEnc(t *testing.T) {
	coder := NewTestCoder()

	s := "simple secret word?"

	encStr, e := coder.Encode(s, 0)

	if e != nil {
		t.Fatal("Error:", e)
	}

	decStr, e := coder.Decode(encStr)

	if e != nil {
		t.Fatal("Decrypt fail:", e)
	}

	if s != decStr {
		log.Printf("%s => %s => %s\n", s, encStr, decStr)
		t.Fatal("no match")
	}
}

func TestShowKeys(t *testing.T) {
	log.Println("pubring")
	ShowKeys("./keys/pubring.gpg")
	log.Println("secring")
	ShowKeys("./keys/secring.gpg")
}
