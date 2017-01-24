package main

import (
	"bytes"
	"golang.org/x/crypto/openpgp"

	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

type Coder struct {
	gpgDir     string
	passphrase string
}

func NewTestCoder() *Coder {
	return &Coder{"./keys/", "baccounts"}
}

func NewCoder() *Coder {
	return &Coder{os.ExpandEnv("$HOME/.gnupg/"), "null"}
}

func (coder *Coder) setPassphrase() {
	fmt.Printf("Passphrase of your GPG key:")
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		log.Fatal("Can't read password:", err)
	}
	coder.passphrase = string(bytes)
}

func (coder *Coder) hasPubKey(id int) bool {
	publicKeyring := coder.gpgDir + "pubring.gpg"
	fmt.Println("Public keyring:", publicKeyring)

	keyringFileBuffer, _ := os.Open(publicKeyring)
	defer keyringFileBuffer.Close()
	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		return false
	}
	return (id > 0 && len(entityList) > id)
}
func (coder *Coder) encode(txt string, id int) (string, error) {
	//publicKeyring := os.ExpandEnv("$HOME/.gnupg/pubring.gpg")
	//publicKeyring := "./keys/pubring.gpg"
	publicKeyring := coder.gpgDir + "pubring.gpg"
	fmt.Println("Public keyring:", publicKeyring)

	keyringFileBuffer, _ := os.Open(publicKeyring)
	defer keyringFileBuffer.Close()
	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		return "fail", err
	}
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, entityList[id:], nil, nil, nil)
	if err != nil {
		return "fail2", err
	}

	_, e := w.Write([]byte(txt))
	if e != nil {
		return "fail3", e
	}
	e = w.Close()
	if e != nil {
		return "fail4", e
	}

	bytes, _ := ioutil.ReadAll(buf)
	encStr := base64.StdEncoding.EncodeToString(bytes)
	//fmt.Println("Encrypted secret:", encStr)
	return encStr, nil
}

func (coder *Coder) decode(txt string) (string, error) {
	//secretKeyring := os.ExpandEnv("$HOME/.gnupg/secring.gpg")
	// secretKeyring := "./keys/secring.gpg"
	secretKeyring := coder.gpgDir + "secring.gpg"
	passphrase := coder.passphrase
	fmt.Println("Secret Keyring:", secretKeyring)

	// init some vars
	var entity *openpgp.Entity
	var entityList openpgp.EntityList

	// Open the private key file
	keyringFileBuffer, err := os.Open(secretKeyring)
	if err != nil {
		return "", err
	}
	defer keyringFileBuffer.Close()
	entityList, err = openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		return "", err
	}
	entity = entityList[0]

	// Get the passphrase and read the private key.
	// Have not touched the encrypted string yet
	passphraseByte := []byte(passphrase)
	// fmt.Println("Decrypting private key using passphrase")
	entity.PrivateKey.Decrypt(passphraseByte)
	for _, subkey := range entity.Subkeys {
		subkey.PrivateKey.Decrypt(passphraseByte)
	}

	// Decode the base64 string
	// fmt.Println("Decoding:", txt)
	dec, err := base64.StdEncoding.DecodeString(txt)
	if err != nil {
		return "", err
	}

	// Decrypt it with the contents of the private key
	md, err := openpgp.ReadMessage(bytes.NewBuffer(dec), entityList, nil, nil)
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return "", err
	}
	decStr := string(bytes)

	return decStr, nil
}

func ShowKeys(keyfile string) error {
	keyringFileBuffer, _ := os.Open(keyfile)
	defer keyringFileBuffer.Close()

	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		return err
	}

	for i, entity := range entityList {
		fmt.Println("Entity:", i)
		if entity.PrimaryKey != nil {
			fmt.Println("  Primary key:", entity.PrimaryKey.CreationTime)
		}
		if entity.PrivateKey != nil {
			fmt.Println("  Private key:", entity.PrivateKey.PublicKey.CreationTime)
		}
		// fmt.Printf("(%v, %v) \n", entity.Revocations, entity.Subkeys)
		for key := range entity.Identities {
			id := entity.Identities[key].UserId
			fmt.Printf("  identity: %v => %v: (%s, %s, %s)\n", key,
				entity.Identities[key].Name, id.Name, id.Comment, id.Email)
		}
	}
	return nil
}
