package baccounts

import (
	"bytes"
	"golang.org/x/crypto/openpgp"

	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"os"

	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

type Coder struct {
	gpgDir     string
	passphrase string
	publicKeyring openpgp.EntityList
}

func NewTestCoder() *Coder {
	return &Coder{"../keys/", "baccounts", nil}
}

func NewCoder() *Coder {
	gpgDir := os.ExpandEnv("$HOME/.gnupg/")

	//publicKeyring := os.ExpandEnv("$HOME/.gnupg/pubring.gpg")
	//publicKeyring := "./keys/pubring.gpg"
	publicKeyring := gpgDir + "pubring.gpg"

	keyringFileBuffer, _ := os.Open(publicKeyring)
	defer keyringFileBuffer.Close()
	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		slog.Error("cant read public keyring", "err", err)
		panic("Can't load pubring")
	}

	coder := &Coder{
		gpgDir,
		"",
		entityList,
	}
	fmt.Println("Public keyring:", coder.PublicKeyringFile())
	fmt.Println("Secret Keyring:", coder.SecretKeyringFile())
	return coder
}

func (coder *Coder) PublicKeyringFile() string {
	return coder.gpgDir + "pubring.gpg"
}
func (coder *Coder) SecretKeyringFile() string {
	return coder.gpgDir + "secring.gpg"
}

func ReadPassword(msg string) (string, error) {
	fmt.Printf(msg)
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Printf("Can't read password: %v\n", err)
		return "", err
	}
	return string(bytes), nil
}

func (coder *Coder) SetPassphrase() {
	pass, err := ReadPassword("Passphrase of your GPG key:")
	if err != nil {
		log.Fatalf("Can't read password: %v", err)
	}
	coder.passphrase = pass
}

func (coder *Coder) HasPubKey(id int) bool {
	return (id > 0 && len(coder.publicKeyring) > id)
}
func (coder *Coder) Encode(txt string, id int) (string, error) {
	slog.Info("Encode", "key", coder.publicKeyring[id])
	keys := coder.publicKeyring[id:]
	
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, keys, nil, nil, nil)
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

func (coder *Coder) Decode(txt string) (string, error) {
	//secretKeyring := os.ExpandEnv("$HOME/.gnupg/secring.gpg")
	// secretKeyring := "./keys/secring.gpg"
	passphrase := coder.passphrase

	// init some vars
	var entity *openpgp.Entity
	var entityList openpgp.EntityList

	// Open the private key file
	keyringFileBuffer, err := os.Open(coder.SecretKeyringFile())
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
