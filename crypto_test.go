package main

import (
	"log"
	"testing"
)

func TestEnc(t *testing.T) {
	coder := NewTestCoder()

	s := "simple secret word?"

	encStr, e := coder.encode(s)

	if e != nil {
		t.Fatal("Error:", e)
	}

	decStr, e := coder.decode(encStr)

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
