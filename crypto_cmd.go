package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"net/url"
)

type testCmd struct {
}

func (*testCmd) Name() string {
	return "test"
}
func (*testCmd) Synopsis() string {
	return "tests your GPG key"
}
func (*testCmd) Usage() string {
	return `test
`
}
func (g *testCmd) SetFlags(f *flag.FlagSet) {
}
func (g *testCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	msg := "you huh?"
	enc, err := coder.Encode(msg, 0)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Printf("%s\n => %s\n", msg, enc)
	coder.SetPassphrase()
	decoded, err := coder.Decode(enc)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	if decoded != msg {
		fmt.Printf("%s != %s\n", msg, decoded)
		return subcommands.ExitFailure
	}
	fmt.Println(" =>", decoded)
	return subcommands.ExitSuccess
}

type listKeysCmd struct {
}

func (*listKeysCmd) Name() string {
	return "list-keys"
}
func (*listKeysCmd) Synopsis() string {
	return "list all your GPG keys"
}
func (*listKeysCmd) Usage() string {
	return `list-keys
`
}
func (g *listKeysCmd) SetFlags(f *flag.FlagSet) {
}
func (g *listKeysCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	publicKeyring := coder.gpgDir + "pubring.gpg"
	secretKeyring := coder.gpgDir + "secring.gpg"
	fmt.Println("Public Keyring:", publicKeyring)
	ShowKeys(publicKeyring)
	fmt.Println("Secret Keyring:", secretKeyring)
	ShowKeys(secretKeyring)
	return subcommands.ExitSuccess
}

type exportCmd struct {
	file string
	key  int
}

func (*exportCmd) Name() string {
	return "export"
}
func (*exportCmd) Synopsis() string {
	return "export your accounts for another GPG key, for other computer"
}
func (*exportCmd) Usage() string {
	return `export --file filename --key <integer>
`
}
func (g *exportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&g.file, "file", "/tmp/baccounts.exported", "File name to export file")
	f.IntVar(&g.key, "key", 0, "PGP public key choice to export (default: 0)")
}
func (g *exportCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {

	var b = (argv[0]).(*Baccount)
	var datafile = (argv[1]).(string)

	if g.file == datafile {
		fmt.Println("No datafile destination: using STDOUT instead")
	}
	if !coder.HasPubKey(g.key) {
		fmt.Println("Invalid key selection")
		return subcommands.ExitFailure
	}

	fmt.Printf("Exporting to %s with key id = %d\n", g.file, g.key)

	// Get public key entity from public keyring and encrypt with it
	coder.SetPassphrase()

	profiles := make([]*Profile, 0)
	for _, p := range b.Profiles {
		fmt.Println(p.Name)
		profile := NewProfile(p.Name, p.Default)
		for key := range p.Sites {
			site := p.Sites[key]
			u, _ := url.Parse(site.Url)
			pass, _ := coder.Decode(site.EncodedPass)
			new, _ := coder.Encode(pass, g.key)
			fmt.Printf("%s => %s\n", site.EncodedPass, new)
			profile.AddSite(u.Host, site.Url, site.Name, new, site.Mail)
		}
		profiles = append(profiles, profile)
	}
	c := &Baccount{profiles, b.DefaultMail, b.Version}

	err := c.save(g.file)
	if err != nil {
		fmt.Println("Failed to save to", g.file)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
