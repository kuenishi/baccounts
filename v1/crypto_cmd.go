package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/google/subcommands"
	"github.com/kuenishi/baccounts/pkg"
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
	coder := baccounts.NewCoder()
	enc, err := coder.Encode(msg, 0)
	if err != nil {
		slog.Error("Encode fail", "fail", enc, "err", err)
		return subcommands.ExitFailure
	}
	slog.Info("encode ok", "msg", msg, "encoded", enc)
	coder.SetPassphrase()
	decoded, err := coder.Decode(enc)
	if err != nil {
		slog.Error("Decode fail", "err", err)
		return subcommands.ExitFailure
	}
	if decoded != msg {
		fmt.Printf("%s != %s\n", msg, decoded)
		return subcommands.ExitFailure
	}
	slog.Info("decode ok", "message", msg, "decoded", decoded)
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
	coder := baccounts.NewCoder()
	publicKeyring := coder.PublicKeyringFile()
	secretKeyring := coder.SecretKeyringFile()
	fmt.Println("Public Keyring:", publicKeyring)
	baccounts.ShowKeys(publicKeyring)
	fmt.Println("Secret Keyring:", secretKeyring)
	baccounts.ShowKeys(secretKeyring)
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
}
func (g *exportCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {

	var b = (argv[0]).(*baccounts.Baccount)
	var datafile = (argv[1]).(string)

	if g.file == datafile {
		fmt.Println("No datafile destination: using STDOUT instead")
	}
	coder := baccounts.NewCoder()
	coder.SetPassphrase()
	fmt.Printf("Exporting to %s with key id = %d\n", g.file, g.key)

	profiles := make([]*baccounts.Profile, 0)
	for _, p := range b.Profiles {
		fmt.Println(p.Name)
		profile := baccounts.NewProfile(p.Name, p.Default)
		for key := range p.Sites {
			site := p.Sites[key]
			u, _ := url.Parse(site.Url)
			pass, _ := coder.Decode(site.EncodedPass)
			profile.AddSite(u.Host, site.Url, site.Name, pass, site.Mail)
		}
		profiles = append(profiles, profile)
	}
	c := &baccounts.Baccount{profiles, b.DefaultMail, b.Version, false}

	if err := c.UpdateConfigFile(g.file); err != nil {
		fmt.Println("Failed to save to", g.file)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
