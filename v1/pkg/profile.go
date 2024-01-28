package baccounts

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
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
	Sites   map[string]*Site
	Default bool
}

func NewProfile(mail string, dflt bool) *Profile {
	return &Profile{mail, make(map[string]*Site), dflt}
}

func (profile *Profile) AddSite(domain, url, name, encpass, mail string) error {
	_, ok := profile.Sites[domain]
	if ok {
		return errors.New("Site already exists: " + domain + " (currently no pass update implemented; TODO)")
	}
	profile.Sites[domain] = &Site{url, name, encpass, mail}
	return nil
}

func (p *Profile) FindSite(urlPattern string) (*Site, error) {
	u, e := url.Parse(urlPattern)
	if e != nil {
		fmt.Println("Cannot parse URL", urlPattern, "as an URL:", e)
		return nil, e
	}

	site, ok := p.Sites[u.Host]
	if ok {
		return site, nil
	}

	count := 0
	word := urlPattern

	for host, site := range p.Sites {
		if strings.Contains(host, word) {
			count += 1
			url := strings.Replace(site.Url, word, "\x1b[31m"+word+"\x1b[0m", -1)
			fmt.Printf("Match: %s\n", url)
		}
	}
	if count > 1 {
		return nil, fmt.Errorf("%d, more than 2 site matched for keyword '%s'\n", count, word)
	} else if count == 0 {
		return nil, fmt.Errorf("No site matching '%s' found", word)
	}
	for host, site := range p.Sites {
		if strings.Contains(host, word) {
			fmt.Printf("One site matched for %s\n", urlPattern)

			return site, nil
		}
	}
	log.Fatalf("Cannot reach here")
	return nil, nil
}

func (p *Profile) UpdateSite(urlPattern, encpass string) error {
	u, e := url.Parse(urlPattern)
	if e != nil {
		fmt.Println("Cannot parse URL", urlPattern, "as an URL:", e)
		return e
	}

	site, ok := p.Sites[u.Host]
	if ok {
		site.EncodedPass = encpass
		return nil
	}

	count := 0
	word := urlPattern

	for host, site := range p.Sites {
		if strings.Contains(host, word) {
			count += 1
			url := strings.Replace(site.Url, word, "\x1b[31m"+word+"\x1b[0m", -1)
			fmt.Printf("Match: %s\n", url)
		}
	}
	if count > 1 {
		return fmt.Errorf("%d, more than 2 site matched for keyword '%s'\n", count, word)
	} else if count == 0 {
		return fmt.Errorf("No site matching '%s' found", word)
	}
	for host, site := range p.Sites {
		if strings.Contains(host, word) {
			site.EncodedPass = encpass
			fmt.Printf("One site matched for %s\n", site.Name)
			return nil
		}
	}
	log.Fatalf("Cannot reach here")
	return nil

}

func (profile *Profile) SetDefault(b bool) {
	profile.Default = b
}
