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

func displaySites(s tcell.Screen, offset int, sites []*Site) {
	_, h := s.Size()
	s.Clear()
	// style := tcell.StyleDefault.Foreground(tcell.ColorCadetBlue.TrueColor()).Background(tcell.ColorWhite)

	// qs := fmt.Sprintf("Query (offset=%d):", offset)
	qs := fmt.Sprintf("Query (offset=%d):", offset)
	emitStr(s, 0, 0, tcell.StyleDefault, qs)
	base := 0
	if offset >= h-2 {
		base -= (offset - h + 2)
	}
	for i, site := range sites {
		if base+i < 0 {
			continue
		}
		style := tcell.StyleDefault
		if i == offset {
			emitStr(s, 0, base+i+1, style, "*")
			style = style.Reverse(true)
		}
		line := fmt.Sprintf("%s", site.Url)
		emitStr(s, 2, base+i+1, style, line)
	}
	//emitStr(s, w/2-7, h/2, style, "Hello, World!")
	//emitStr(s, w/2-9, h/2+1, tcell.StyleDefault, "Press ESC to exit.")
	s.Show()
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

	defStyle := tcell.StyleDefault
	//Background(tcell.ColorBlack).
	//Foreground(tcell.ColorWhite)
	s.SetStyle(defStyle)

	sites := make([]*Site, 0, len(p.Sites))
	for _, site := range p.Sites {
		sites = append(sites, site)
		//		fmt.Printf("%s:\t%s\t%s\n", domain, site.Url, site.Mail)
	}
	pos := 0
	displaySites(s, pos, sites)

	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
			displaySites(s, pos, sites)
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				return nil, fmt.Errorf("Selection cancelled")
			case tcell.KeyEnter:
				return sites[pos], nil
			case tcell.KeyPgUp, tcell.KeyUp:
				if pos > 0 {
					pos -= 1
				}
			case tcell.KeyPgDn, tcell.KeyDown:
				if pos < len(sites)-1 {
					pos += 1
				}
			}
		}
		s.Sync()
		displaySites(s, pos, sites)
	}

	//return  no reach here
	return nil, fmt.Errorf("Bug: no reach here")
}
