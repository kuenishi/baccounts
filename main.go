package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"crypto/rand"
	"math/big"
	"net/url"
	"strings"

	"github.com/google/subcommands"
	"github.com/kuenishi/baccounts/pkg"
)

var defaultMail = "who@example.com"
var defaultName = "john smith"

type listCmd struct {
}

func (*listCmd) Name() string {
	return "list"
}
func (*listCmd) Synopsis() string {
	return "List all profiles"
}
func (*listCmd) Usage() string {
	return `list`
}
func (l *listCmd) SetFlags(f *flag.FlagSet) {
}
func (l *listCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {
	var b = (argv[0]).(*Baccount)
	fmt.Println("default mail:", b.DefaultMail)
	for _, acc := range b.Profiles {
		fmt.Printf("%s default=%v\n", acc.Name, acc.Default)
		for dom := range acc.Sites {
			fmt.Printf("\t%s:\t%s\t%s\n", dom, acc.Sites[dom].Url, acc.Sites[dom].Mail)
		}
	}
	return subcommands.ExitSuccess
}

type addProfileCmd struct {
	name string
}

func (*addProfileCmd) Name() string {
	return "add-profile"
}
func (*addProfileCmd) Synopsis() string {
	return "Add a new profile"
}
func (*addProfileCmd) Usage() string {
	return `add-profile --name name
`
}
func (a *addProfileCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&a.name, "name", defaultName, "Name of a new profile")
}
func (a *addProfileCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {

	var b = (argv[0]).(*Baccount)
	var datafile = (argv[1]).(string)

	if a.name == defaultName {
		fmt.Println("No name provided")
		return subcommands.ExitFailure
	}

	p, e := b.GetProfile(a.name)
	if p != nil || e == nil {
		fmt.Println("Profile already exists:", p)
		return subcommands.ExitFailure
	}

	fmt.Println("Adding a profile:", a.name)
	dflt := (len(b.Profiles) == 0)
	b.Profiles = append(b.Profiles, NewProfile(a.name, dflt))

	b.updateConfigFile(datafile)
	return subcommands.ExitSuccess
}

type generateCmd struct {
	url  string
	name string
	mail string
	len  int
}

func (*generateCmd) Name() string {
	return "generate"
}
func (*generateCmd) Synopsis() string {
	return "Generates and save password for the site"
}
func (*generateCmd) Usage() string {
	return `generate --url url --name name --mail mail --len 16
`
}
func (g *generateCmd) SetFlags(f *flag.FlagSet) {
	defaultName := os.ExpandEnv("$USER")
	f.StringVar(&g.url, "url", "https://example.com", "URL of the site")
	f.StringVar(&g.name, "name", defaultName, "Profile of the site")
	f.StringVar(&g.mail, "mail", defaultMail, "Mail address")
	f.IntVar(&g.len, "len", 16, "Length of the pass")
}
func (g *generateCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {
	fmt.Printf("Generate profiles: %s @ %s\n", g.name, g.url)
	var b = (argv[0]).(*Baccount)
	var datafile = (argv[1]).(string)

	if g.url == "https://example.com" {
		fmt.Println("URL Required")
		return subcommands.ExitFailure
	}

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	max := big.NewInt(int64(len(letters)))

	p, e := b.GetProfile(g.name)
	if e != nil {
		fmt.Println("Error:", e)
		return subcommands.ExitFailure
	}

	u, e := url.Parse(g.url)
	if e != nil {
		fmt.Println("Error:", e)
		return subcommands.ExitFailure
	}
	fmt.Println("host:", u.Host)

	bytes := make([]rune, g.len)
	for i := 0; i < g.len; i++ {
		j, e := rand.Int(rand.Reader, max)
		if e != nil {
			fmt.Println("Error:", e)
			return subcommands.ExitFailure
		}
		bytes[i] = letters[j.Int64()]
	}

	fmt.Println(string(bytes))

	coder := baccounts.NewCoder()
	// TODO: check we already have same site
	encpass, err := coder.Encode(string(bytes), 0)

	if err != nil {
		fmt.Println("Can't decode pass:", err)
		return subcommands.ExitFailure
	}

	e = p.AddSite(u.Host, g.url, g.name, encpass, g.mail)
	if e != nil {
		fmt.Println("Error:", e)
		return subcommands.ExitFailure
	}

	b.updateConfigFile(datafile)
	return subcommands.ExitSuccess
}

type showCmd struct {
	name string
	site string
}

func (*showCmd) Name() string {
	return "show"
}
func (*showCmd) Synopsis() string {
	return "Show password for the site"
}
func (*showCmd) Usage() string {
	return `show -site example.com -mail mail -name name
`
}
func (g *showCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&g.site, "site", "example.com", "Site name of the acc")
	f.StringVar(&g.name, "name", defaultName, "Profile")
}

func (g *showCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {

	var b = (argv[0]).(*Baccount)

	if g.site == "example.com" {
		fmt.Println("Site lacking")
		return subcommands.ExitFailure
	}

	p, e := b.GetProfile(g.name)
	if e != nil {
		fmt.Println("Cannot get profile:", e)
		return subcommands.ExitFailure
	}

	u, e := url.Parse(g.site)
	if e != nil {
		fmt.Println("Cannot parse URL", g.site, "as an URL:", e)
		return subcommands.ExitFailure
	}
	if u.Host == "" {
		var word = g.site

		var count = 0
		var one_site = Site{}
		for host, site := range p.Sites {
			if strings.Contains(host, word) {
				count += 1
				one_site = site
				url := strings.Replace(site.Url, word, "\x1b[31m"+word+"\x1b[0m", -1)
				fmt.Printf("Match: %s\n", url)
			}
		}
		if count > 1 {
			fmt.Printf("%d, more than 2 site matched for keyword '%s'\n", count, word)
			return subcommands.ExitFailure
		} else if count == 0 {
			fmt.Printf("No site matching '%s' found", word)
			return subcommands.ExitFailure
		} else {
			fmt.Printf("One site matched for %s\n", one_site.Name)
			return b.show(one_site)
		}
	}

	site, ok := p.Sites[u.Host]
	if !ok {
		fmt.Println("site not found:", u.Host, p.Sites)
		return subcommands.ExitFailure
	}
	return b.show(site)
}

type setDefaultCmd struct {
	name string
}

func (*setDefaultCmd) Name() string {
	return "set-default"
}
func (*setDefaultCmd) Synopsis() string {
	return "set default profile"
}
func (*setDefaultCmd) Usage() string {
	return `set-default --name name
`
}
func (g *setDefaultCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&g.name, "name", defaultName, "Profile name to set default")
}
func (g *setDefaultCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {
	var b = (argv[0]).(*Baccount)
	var datafile = (argv[1]).(string)

	p, e := b.GetProfile(g.name)
	if e != nil {
		fmt.Println("Error:", e)
		return subcommands.ExitFailure
	}

	for _, tmp := range b.Profiles {
		tmp.SetDefault(false)
	}

	p.SetDefault(true)
	b.updateConfigFile(datafile)

	fmt.Println("Set default:", p.Name, p.Default)
	return subcommands.ExitSuccess
}

func main() {
	datafile := os.ExpandEnv("$HOME/.baccounts")
	b, _ := LoadKeys(datafile)
	if b != nil {
		fmt.Printf("baccounts version %s - data file version: %s\n", version, b.Version)
	}

	subcommands.Register(subcommands.HelpCommand(), "meta")
	subcommands.Register(subcommands.FlagsCommand(), "meta")
	subcommands.Register(subcommands.CommandsCommand(), "meta")
	subcommands.Register(&testCmd{}, "meta")
	subcommands.Register(&listKeysCmd{}, "meta")

	// profiles
	subcommands.Register(&listCmd{}, "profile")
	subcommands.Register(&addProfileCmd{}, "profile")
	// deleteMail deletes mail only when it has no sites
	subcommands.Register(&generateCmd{}, "profile")
	// delete deletes site info
	subcommands.Register(&showCmd{}, "profile")
	subcommands.Register(&setDefaultCmd{}, "profile")
	// backup --path ~/Dropbox/
	// restore --path ~/Dropbox/
	subcommands.Register(&exportCmd{}, "compat")

	flag.Parse()
	ctx := context.Background()
	ret := int(subcommands.Execute(ctx, b, datafile))

	os.Exit(ret)
}
