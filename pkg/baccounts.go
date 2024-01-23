package baccounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
	"github.com/google/subcommands"
)

const Version = "0.1.0"

type Baccount struct {
	Profiles    []*Profile
	DefaultMail string // Used for private key seek
	Version     string
	ReadOnly    bool
}

func (b *Baccount) List() error {
	fmt.Println("default mail:", b.DefaultMail)
	for _, acc := range b.Profiles {
		fmt.Printf("%s default=%v\n", acc.Name, acc.Default)
		for dom := range acc.Sites {
			fmt.Printf("\t%s:\t%s\t%s\n", dom, acc.Sites[dom].Url, acc.Sites[dom].Mail)
		}
	}
	return nil
}

func (b *Baccount) AddProfile(name, datafile string) error {
	p, e := b.GetProfile(name)
	if p != nil || e == nil {
		return fmt.Errorf("Profile already exists:", p)
	}

	fmt.Println("Adding a profile:", name)
	dflt := (len(b.Profiles) == 0)
	b.Profiles = append(b.Profiles, NewProfile(name, dflt))

	return b.UpdateConfigFile(datafile)
}

func (b *Baccount) Update(name, site, datafile string) error {
	fmt.Printf("Update profile: %s @ %s\n", name, site)

	if site == "" {
		return fmt.Errorf("-site cannot be empty")
	}

	p, e := b.GetProfile(name)
	if e != nil {
		return e
	}

	if _, err := p.FindSite(site); err != nil {
		return fmt.Errorf("Cannot find site: %v\n", site, err)
	}

	pass, err := ReadPassword("New Password: ")
	if err != nil {
		return err
	}
	pass2, err := ReadPassword("Input Again: ")
	if err != nil {
		return err
	}
	if pass != pass2 {
		return fmt.Errorf("Password inputs don't match.")
	}

	if len(pass) < 8 {
		return fmt.Errorf("New password should be longer than 8 chars (%d)\n", len(pass))
	}

	coder := NewCoder()
	encpass, err := coder.Encode(pass, 0)
	if err != nil {
		log.Printf("Can't encode pass. Encoded as %s: %v", encpass, err)
		return err
	}

	if err := p.UpdateSite(site, encpass); err != nil {
		log.Println("Can't update profile:", err)
		return err
	}
	if err := b.UpdateConfigFile(datafile); err != nil {
		log.Println("Cannot update password file")
		return err
	}
	fmt.Println("Password successfully updated.")
	return nil
}

// ===============================================

func (b *Baccount) GetDefault() (*Profile, error) {
	for _, p := range b.Profiles {
		if p.Default {
			return p, nil
		}
	}
	return nil, errors.New("Default profile not found")
}

func (b *Baccount) GetProfile(name string) (*Profile, error) {
	if name == "" {
		return b.GetDefault()
	}
	fmt.Println("profiles:", len(b.Profiles), name)
	for _, p := range b.Profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, errors.New("Profile not found:" + name)
}

func (b *Baccount) toJson() (string, error) {
	j, e := json.Marshal(b)
	if e != nil {
		return "error...", e
	}
	return string(j), nil
}

func (b *Baccount) UpdateConfigFile(dest string) error {
	if b.ReadOnly {
		return fmt.Errorf("read-only config: possibly, move it to $XDG_CONFIG_HOME/baccounts.json")
	}

	dir1, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	tmpfile := filepath.Join(dir1, "baccounts-temp.json")
	if err := b.save(tmpfile); err != nil {
		return err
	}
	if err := os.Rename(tmpfile, dest); err != nil {
		fmt.Printf("Error on saving profiles: %v\n", err)
		return err
	}
	return nil
}

func (b *Baccount) save(datafile string) error {

	fp, err := os.OpenFile(datafile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()

	encoder := json.NewEncoder(fp)
	return encoder.Encode(b)
}

func (b *Baccount) Show(site *Site) subcommands.ExitStatus {
	coder := NewCoder()
	coder.SetPassphrase()
	pass, err := coder.Decode(site.EncodedPass)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}

	e := clipboard.WriteAll(pass)
	if e != nil {
		fmt.Println("Failed to copy password to clipboard: ", e)
		return subcommands.ExitFailure
	}
	fmt.Printf("Pass for %s (%s) copied to clipboard\n", site.Url, site.Name)
	return subcommands.ExitSuccess

}

func LoadAccounts() (*Baccount, string, error) {
	// Try new config file
	dir1, err := os.UserConfigDir()
	if err == nil {
		datafile := filepath.Join(dir1, "baccounts.json")
		b, err := LoadKeys(datafile)
		if err != nil {
			log.Printf("Failed to load keys: %v", err)
		} else {
			b.ReadOnly = false
			return b, datafile, nil
		}
	}
	// Try old config file
	dir2, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("Failed to get home dir: %v", err)
	}
	datafile := filepath.Join(dir2, ".baccounts")
	log.Printf("Reading keys from %s", datafile)
	b, err := LoadKeys(datafile)
	if err != nil {
		return nil, datafile, err
	}
	b.ReadOnly = true
	return b, datafile, nil
}

func LoadKeys(datafile string) (*Baccount, error) {
	fp, err := os.Open(datafile)
	if err != nil {
		fmt.Printf("No such file as %s. Will create a new one\n", datafile)
		// return &Baccount{make([]*Profile, 0, 16), nil, version}, nil
		return nil, err
	}
	defer fp.Close()

	var b Baccount
	decoder := json.NewDecoder(fp)
	if err := decoder.Decode(&b); err != nil {
		return nil, fmt.Errorf("Unable to parse json")
	}
	return &b, nil
}
