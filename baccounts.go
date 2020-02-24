package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/atotto/clipboard"
	"github.com/google/subcommands"
)

type Site struct {
	Url         string
	Name        string
	EncodedPass string
	Mail        string
}

type Profile struct {
	// This mail address corresponds to login info
	Name    string
	Sites   map[string]Site
	Default bool
}

func NewProfile(mail string, dflt bool) *Profile {
	return &Profile{mail, make(map[string]Site), dflt}
}

func (profile *Profile) AddSite(domain, url, name, encpass, mail string) error {
	_, ok := profile.Sites[domain]
	if ok {
		return errors.New("Site already exists: " + domain + " (currently no pass update implemented; TODO)")
	}
	profile.Sites[domain] = Site{url, name, encpass, mail}
	return nil
}

func (profile *Profile) SetDefault(b bool) {
	profile.Default = b
}

var version = "0.0.1-SNAPSHOT"

type Baccount struct {
	Profiles    []*Profile
	DefaultMail string // Used for private key seek
	Version     string
}

func (b *Baccount) GetDefault() (*Profile, error) {
	for _, p := range b.Profiles {
		if p.Default {
			return p, nil
		}
	}
	return nil, errors.New("Default profile not found")
}

func (b *Baccount) GetProfile(name string) (*Profile, error) {
	if name == defaultName {
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

func (b *Baccount) updateConfigFile(dest string) error {
	tmpfile := os.ExpandEnv("$HOME/.baccounts-temp")
	b.save(tmpfile)
	e := os.Rename(tmpfile, dest)
	if e != nil {
		fmt.Printf("Error on saving profiles: %v\n", e)
		os.Exit(-1)
	}
	return nil
}

func (b *Baccount) save(file string) error {
	json, err := b.toJson()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, []byte(json), 0600)
}

func (b *Baccount) show(site Site) subcommands.ExitStatus {
	coder.setPassphrase()
	pass, err := coder.decode(site.EncodedPass)
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

func LoadKeysFromJson(js string) (*Baccount, error) {
	var b Baccount
	err := json.Unmarshal([]byte(js), &b)
	if err != nil {
		fmt.Printf("Invalid format: %v\n", err)
		return nil, err
	}
	return &b, nil
}

func LoadKeys(datafile string) (*Baccount, error) {
	file, err := os.Open(datafile)
	if err != nil {
		fmt.Printf("No such file as %s. Will create a new one\n", datafile)
		// return &Baccount{make([]*Profile, 0, 16), nil, version}, nil
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("no content")
	}
	return LoadKeysFromJson(string(bytes))
}
