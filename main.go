package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"crypto/rand"
	"math/big"
	"net/url"

	"github.com/google/subcommands"
	"github.com/kuenishi/baccounts/pkg"
)

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
	var b = (argv[0]).(*baccounts.Baccount)
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
	f.StringVar(&a.name, "name", "", "Name of a new profile")
}
func (a *addProfileCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {

	var b = (argv[0]).(*baccounts.Baccount)
	var datafile = (argv[1]).(string)

	p, e := b.GetProfile(a.name)
	if p != nil || e == nil {
		fmt.Println("Profile already exists:", p)
		return subcommands.ExitFailure
	}

	fmt.Println("Adding a profile:", a.name)
	dflt := (len(b.Profiles) == 0)
	b.Profiles = append(b.Profiles, baccounts.NewProfile(a.name, dflt))

	b.UpdateConfigFile(datafile)
	return subcommands.ExitSuccess
}

type updateCmd struct {
	site string
	name string
	new  string
}

func (*updateCmd) Name() string {
	return "update"
}
func (*updateCmd) Synopsis() string {
	return "Update password for the site"
}
func (*updateCmd) Usage() string {
	return `update -name name -site site
`
}
func (g *updateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&g.site, "site", "", "Profile of the site (required)")
	f.StringVar(&g.name, "name", "", "Profile name")
}
func (g *updateCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {
	fmt.Printf("Update profile: %s @ %s\n", g.name, g.site)
	var b = (argv[0]).(*baccounts.Baccount)
	var datafile = (argv[1]).(string)

	if g.site == "" {
		fmt.Println("-site cannot be empty")
		return subcommands.ExitFailure
	}

	p, e := b.GetProfile(g.name)
	if e != nil {
		fmt.Println("Error:", e)
		return subcommands.ExitFailure
	}

	if _, err := p.FindSite(g.site); err != nil {
		fmt.Printf("Cannot find site: %v\n", g.site, err)
		return subcommands.ExitFailure
	}

	pass, err := baccounts.ReadPassword("New Password: ")
	if err != nil {
		return subcommands.ExitFailure
	}
	pass2, err := baccounts.ReadPassword("Input Again: ")
	if err != nil {
		return subcommands.ExitFailure
	}
	if pass != pass2 {
		fmt.Printf("Password inputs don't match.\n")
		return subcommands.ExitFailure
	}

	if len(pass) < 8 {
		fmt.Printf("New password should be longer than 8 chars (%d)\n", len(pass))
		return subcommands.ExitFailure
	}

	coder := baccounts.NewCoder()
	encpass, err := coder.Encode(pass, 0)
	if err != nil {
		fmt.Println("Can't encode pass:", err)
		return subcommands.ExitFailure
	}

	if err := p.UpdateSite(g.site, encpass); err != nil {
		fmt.Println("Can't update profile:", err)
		return subcommands.ExitFailure
	}
	b.UpdateConfigFile(datafile)
	fmt.Println("Password successfully updated.")
	return subcommands.ExitSuccess
}

type generateCmd struct {
	url  string
	name string
	mail string
	len  int
	num  bool
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
	f.StringVar(&g.url, "url", "https://example.com", "URL of the site")
	f.StringVar(&g.name, "name", "", "Profile of the site")
	f.StringVar(&g.mail, "mail", "", "Mail address")
	f.IntVar(&g.len, "len", 16, "Length of the pass")
	f.BoolVar(&g.num, "num", false, "Num-only")
}
func (g *generateCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {
	fmt.Printf("Generate profiles: %s @ %s\n", g.name, g.url)
	var b = (argv[0]).(*baccounts.Baccount)
	var datafile = (argv[1]).(string)

	if g.url == "https://example.com" {
		fmt.Println("URL Required")
		return subcommands.ExitFailure
	}

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	var nums = []rune("0123456789")

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
		if g.num {
			j, e := rand.Int(rand.Reader, big.NewInt(int64(len(nums))))
			if e != nil {
				fmt.Println("Error:", e)
				return subcommands.ExitFailure
			}
			bytes[i] = nums[j.Int64()]
		} else {
			j, e := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
			if e != nil {
				fmt.Println("Error:", e)
				return subcommands.ExitFailure
			}
			bytes[i] = letters[j.Int64()]
		}
	}

	fmt.Println(string(bytes))

	coder := baccounts.NewCoder()
	// TODO: check we already have same site
	encpass, err := coder.Encode(string(bytes), 0)

	if err != nil {
		fmt.Println("Can't decode pass:", err)
		return subcommands.ExitFailure
	}

	e = p.AddSite(u.Host, g.url, p.Name, encpass, g.mail)
	if e != nil {
		fmt.Println("Error:", e)
		return subcommands.ExitFailure
	}

	b.UpdateConfigFile(datafile)
	return subcommands.ExitSuccess
}

type showCmd struct {
	name        string
	site        string
	interactive bool
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
	f.StringVar(&g.site, "site", "", "Site name of the acc")
	f.StringVar(&g.name, "name", "", "Profile")
	f.BoolVar(&g.interactive, "i", false, "Interactive mode")
}

func (g *showCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {

	var b = (argv[0]).(*baccounts.Baccount)

	p, e := b.GetProfile(g.name)
	if e != nil {
		fmt.Println("Cannot get profile:", e)
		return subcommands.ExitFailure
	}

	if g.interactive {
		// Interactive mode
		if g.site != "" {
			fmt.Println("-site not needed in interactive mode")
			return subcommands.ExitFailure
		}

		site, err := p.FindSiteInteractive()
		if err != nil {
			return subcommands.ExitFailure
		}
		fmt.Println("Selected: %s", site.Url)
		return b.Show(site)
	}

	// Noniteractive mode
	if g.site == "" {
		fmt.Println("Site lacking")
		return subcommands.ExitFailure
	}

	site, err := p.FindSite(g.site)
	if err != nil {
		return subcommands.ExitFailure
	}
	return b.Show(site)
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
	f.StringVar(&g.name, "name", "", "Profile name to set default")
}
func (g *setDefaultCmd) Execute(_ context.Context, f *flag.FlagSet, argv ...interface{}) subcommands.ExitStatus {
	var b = (argv[0]).(*baccounts.Baccount)
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
	b.UpdateConfigFile(datafile)

	fmt.Println("Set default:", p.Name, p.Default)
	return subcommands.ExitSuccess
}

func main() {
	datafile := os.ExpandEnv("$HOME/.baccounts")
	b, err := baccounts.LoadKeys(datafile)
	if err != nil {
		fmt.Printf("baccounts version %s - data file version: %s\n", baccounts.Version, b.Version)
		return
	}

	subcommands.Register(subcommands.HelpCommand(), "meta")
	subcommands.Register(subcommands.FlagsCommand(), "meta")
	subcommands.Register(subcommands.CommandsCommand(), "meta")
	subcommands.Register(&testCmd{}, "meta")
	subcommands.Register(&listKeysCmd{}, "meta")

	// profiles
	subcommands.Register(&listCmd{}, "profile")
	subcommands.Register(&addProfileCmd{}, "profile")
	subcommands.Register(&updateCmd{}, "profile")
	// deleteMail deletes mail only when it has no sites
	subcommands.Register(&generateCmd{}, "profile")
	// delete deletes site info
	subcommands.Register(&showCmd{}, "profile")
	subcommands.Register(&setDefaultCmd{}, "profile")
	subcommands.Register(&exportCmd{}, "compat")

	flag.Parse()
	ctx := context.Background()
	ret := int(subcommands.Execute(ctx, b, datafile))

	os.Exit(ret)
}
