package baccounts

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
)

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

type display struct {
	screen tcell.Screen
	sites  []*Site
	offset int
	qs     string
}

func newDisplay(s tcell.Screen, sites []*Site) *display {
	return &display{
		screen: s,
		sites:  sites,
		offset: 0,
		qs:     "",
	}
}

func (d *display) current() *Site {
	return d.sites[d.offset]
}

func (d *display) up() {
	if d.offset > 0 {
		d.offset -= 1
	}
}

func (d *display) down() {
	if d.offset < len(d.sites)-1 {
		d.offset += 1
	}
}

func (d *display) inc(s string) {
	d.qs += s
}

func (d *display) dec() {
	//d.inc("bs")
	if d.qs != "" {
		d.qs = d.qs[:len(d.qs)-1]
	}
}

func (d *display) display() {
	_, h := d.screen.Size()
	d.screen.Clear()
	// style := tcell.StyleDefault.Foreground(tcell.ColorCadetBlue.TrueColor()).Background(tcell.ColorWhite)

	// qs := fmt.Sprintf("Query (offset=%d):", offset)
	qs := fmt.Sprintf("Query (offset=%d): %s", d.offset, d.qs)
	emitStr(d.screen, 0, 0, tcell.StyleDefault, qs)
	base := 0
	if d.offset >= h-2 {
		base -= (d.offset - h + 2)
	}
	for i, site := range d.sites {
		if base+i < 0 {
			continue
		}
		style := tcell.StyleDefault
		if i == d.offset {
			emitStr(d.screen, 0, base+i+1, style, "*")
			style = style.Reverse(true)
		}
		line := fmt.Sprintf("%s", site.Url)
		emitStr(d.screen, 2, base+i+1, style, line)
	}
	//emitStr(s, w/2-7, h/2, style, "Hello, World!")
	//emitStr(s, w/2-9, h/2+1, tcell.StyleDefault, "Press ESC to exit.")
	d.screen.Show()
}

func (p *Profile) FindSiteInteractive() (*Site, error) {
	encoding.Register()

	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		return nil, e
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		return nil, e
	}
	defer s.Fini()

	s.SetStyle(tcell.StyleDefault)

	sites := make([]*Site, 0, len(p.Sites))
	for _, site := range p.Sites {
		sites = append(sites, site)
		//		fmt.Printf("%s:\t%s\t%s\n", domain, site.Url, site.Mail)
	}

	d := newDisplay(s, sites)
	d.display()

	for {
		switch ev := s.PollEvent().(type) {
		//case *tcell.EventResize:
		//s.Sync()
		//d.display()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				return nil, fmt.Errorf("Selection cancelled")
			case tcell.KeyEnter:
				return d.current(), nil
			case tcell.KeyPgUp, tcell.KeyUp:
				d.up()
			case tcell.KeyPgDn, tcell.KeyDown, tcell.KeyCtrlSpace:
				d.down()
			case tcell.KeyRune:
				d.inc(string(ev.Rune()))
			case tcell.KeyBS, tcell.KeyDEL:
				d.dec()
			default:
				return nil, fmt.Errorf("Unused key pressed: %v", ev.Key())
			}
		}
		s.Sync()
		d.display()
	}

	//return  no reach here
	return nil, fmt.Errorf("Bug: no reach here")
}
