package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type TXSearch struct {
	gui *gocui.Gui

	opened bool
	txid   string
	cb     func(string) error
}

func NewTXSearch(gui *gocui.Gui) *TXSearch {
	ts := &TXSearch{gui: gui}
	return ts
}

func (s *TXSearch) Callback(fn func(txId string) error) {
	s.cb = fn
}

func (s *TXSearch) SetKeybinding() {
	s.gui.SetKeybinding("", 'f', gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
		s.gui.DeleteKeybinding("", 'f', gocui.ModNone)
		s.Open()
		return nil
	})
}

func (s *TXSearch) Layout(g *gocui.Gui) error {
	name := "tx_search"
	if !s.opened {
		g.Cursor = false
		g.DeleteView(name)
		return nil
	}

	g.Cursor = true
	x, y := g.Size()
	v, err := g.SetView(name, x/2-35, y/2-1, x/2+35, y/2+1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Track transaction (txid)"
		v.Editable = true
		g.SetCurrentView(name)
		v.Editor = gocui.EditorFunc(s.editFn)
		v.Autoscroll = false
		fmt.Fprintf(v, "%s", s.txid)
		v.SetCursor(len(s.txid), 0)

		g.SetKeybinding(v.Name(), gocui.KeyEsc, gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
			s.Close()
			return nil
		})
	}

	return nil
}

func (s *TXSearch) editFn(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch key {
	case gocui.KeyEnter:
		if id := s.txid; id != "" {
			s.cb(id)
		}
		s.Close()
	case gocui.KeyArrowDown, gocui.KeyArrowUp:
		return
	}

	gocui.DefaultEditor.Edit(v, key, ch, mod)
	s.txid, _ = v.Line(0)
}

func (s *TXSearch) Open() {
	s.opened = true
}

func (s *TXSearch) Close() {
	s.SetKeybinding()
	s.opened = false
}
