package ui

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gchaincl/mempool/client"
	"github.com/jroimartin/gocui"
)

const (
	BLOCK_WIDTH = 22
)

type state struct {
	loaded          bool
	blocks          []client.Block
	mempool         []client.MempoolBlock
	vBytesPerSecond int
	info            *client.MempoolInfo
	tracking        *client.TrackTx
}

type UI struct {
	client *client.Client
	gui    *gocui.Gui
	fd     *FeeDistribution
	ts     *TXSearch
	state  state
}

func New() (*UI, error) {
	gui, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		return nil, err
	}

	ui := &UI{gui: gui}
	ui.fd = NewFeeDistribution(gui)
	ui.ts = NewTXSearch(gui)
	gui.SetManager(ui, ui.fd, ui.ts)

	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	ui.ts.SetKeybinding()

	gui.Mouse = true
	gui.Highlight = true
	gui.InputEsc = true
	gui.SelFgColor = gocui.ColorWhite

	go func() {
		c, err := client.New()
		if err != nil {
			log.Fatal(err)
		}
		if err := c.Want(); err != nil {
			log.Fatal(err)
		}
		ui.client = c
		ui.ts.Callback(func(txId string) error {
			return c.Track(txId)
		})

		for {
			resp, err := c.Read()
			if err != nil {
				log.Fatal(err)
			}
			ui.Render(resp)
		}

	}()

	return ui, nil
}

func quit(*gocui.Gui, *gocui.View) error { return gocui.ErrQuit }

func (ui *UI) Close() { ui.gui.Close() }

func (ui *UI) Loop() error {
	err := ui.gui.MainLoop()
	// Mask ErrQuit
	if err == gocui.ErrQuit {
		return nil
	}
	return err
}

func (ui *UI) Render(resp *client.Response) {
	ui.state.loaded = true

	ui.state.vBytesPerSecond = resp.VBytesPerSecond

	nBlocks := len(resp.Blocks)
	blocks := make([]client.Block, nBlocks)
	for i := 0; i < nBlocks; i++ {
		blocks[i] = resp.Blocks[len(resp.Blocks)-1-i]
	}

	if bs := blocks; len(bs) != 0 {
		ui.state.blocks = bs
	}

	if bs := resp.MempoolBlocks; len(bs) != 0 {
		ui.state.mempool = bs
	}

	if b := resp.Block; b != nil {
		ui.state.blocks = append([]client.Block{*b}, ui.state.blocks...)
	}

	if info := resp.MempoolInfo; info != nil {
		ui.state.info = info
	}

	// Update tracking info
	ui.state.tracking = &resp.TrackTx

	// delete all the views
	for _, v := range ui.gui.Views() {
		ui.gui.DeleteView(v.Name())
	}

	ui.gui.Update(ui.Layout)
}

func (ui *UI) Layout(g *gocui.Gui) error {
	x, y := g.Size()

	if !ui.state.loaded {
		return ui.loading(g, x, y)
	}
	g.DeleteView("loading")

	// vertical layout is used if 8 blocks don't fit on the screen
	// when in vertical layout the mempool is shown in the top
	// and the blockchain in the bottom
	vertical := BLOCK_WIDTH*6 > x

	track := ui.state.tracking

	// draw projected blocks (mempool)
	for i, _ := range ui.state.mempool {
		name := fmt.Sprintf("mempool-block-%d", i)
		var x0, x1, y0, y1 int
		if vertical {
			x0 = x - (BLOCK_WIDTH+2)*(i+1)
			x1 = x0 + BLOCK_WIDTH
			y0 = (y / 2) - 12
			y1 = (y / 2) - 2
		} else {
			x0 = x/2 - BLOCK_WIDTH*(i+1)
			x1 = x0 + BLOCK_WIDTH - 2
			y0 = (y / 2) - 5
			y1 = (y / 2) + 5
		}

		v, err := g.SetView(name, x0, y0, x1, y1)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.BgColor = gocui.ColorBlack
			g.SetKeybinding(v.Name(), gocui.MouseLeft, gocui.ModNone, ui.onBlockClick)

			if track.Tracking && !track.TX.Status.Confirmed {
				if track.BlockHeight == i {
					v.SelBgColor = gocui.ColorRed
					v.SelFgColor = gocui.ColorRed
					g.SetCurrentView(v.Name())
				}
			}
		}

		v.Clear()
		if _, err := v.Write(ui.printProjectedBlock(i, x1-x0, y1-y0)); err != nil {
			return err
		}
	}

	if err := ui.separator(g, x, y, vertical); err != nil {
		return err
	}

	// draw blockchain blocks
	for i, block := range ui.state.blocks {
		name := fmt.Sprintf("block-%d", block.Height)
		var x0, x1, y0, y1 int
		if vertical {
			x0 = x - (BLOCK_WIDTH+2)*(i+1)
			x1 = x0 + BLOCK_WIDTH
			y0 = (y / 2) + 2
			y1 = (y / 2) + 12
		} else {
			x0 = (x / 2) + (BLOCK_WIDTH*i + 1) + 1
			x1 = x0 + BLOCK_WIDTH - 2
			y0 = (y / 2) - 5
			y1 = (y / 2) + 5
		}

		v, err := g.SetView(name, x0, y0, x1, y1)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.BgColor = gocui.ColorBlack
			g.SetKeybinding(v.Name(), gocui.MouseLeft, gocui.ModNone, ui.onBlockClick)

		}

		v.Title = fmt.Sprintf("#%d", block.Height)
		if track.Tracking && track.TX.Status.Confirmed {
			if track.BlockHeight == block.Height {
				v.SelBgColor = gocui.ColorRed
				v.SelFgColor = gocui.ColorRed
				g.SetCurrentView(v.Name())
			}
		}
		v.Clear()
		if _, err := v.Write(ui.printBlock(i, x1-x0, y1-y0)); err != nil {
			return err
		}
	}

	if err := ui.info(g, x, y); err != nil {
		return err
	}

	return nil
}

func (ui *UI) loading(g *gocui.Gui, x, y int) error {
	v, err := g.SetView("loading", x/2-10, y/2-1, x/2+10, y/2+1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	v.Clear()
	fmt.Fprintf(v, "Loading blocks ...")
	return nil
}

func (ui *UI) printProjectedBlock(n int, x, y int) []byte {
	b := ui.state.mempool[n]
	return MempoolBlock(b).Print(n, x, y)
}

func (ui *UI) printBlock(n int, x, y int) []byte {
	b := ui.state.blocks[n]
	return Block(b).Print(n, x, y)
}

func (ui *UI) separator(g *gocui.Gui, x, y int, vertical bool) error {
	var x0, x1, y0, y1 int
	if vertical {
		x0, x1 = 0, x
		y0, y1 = y/2-1, y/2+1
	} else {
		x0, x1 = x/2-1, x/2+1
		y0, y1 = 0, y
	}

	v, err := g.SetView("separator", x0, y0, x1, y1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Wrap = true
	}

	v.Clear()
	if vertical {
		fmt.Fprintf(v, strings.Repeat("-", x))
	} else {
		fmt.Fprintf(v, strings.Repeat("|", y))
	}

	return nil
}

func (ui *UI) info(g *gocui.Gui, x, y int) error {
	var (
		white = color.New(color.FgWhite).SprintfFunc()
		red   = color.New(color.FgRed).SprintfFunc()
		blue  = color.New(color.FgBlue).SprintfFunc()
	)

	v, err := g.SetView("info", 0, y-2, x, y)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.BgColor = gocui.ColorBlack
	}
	v.Clear()
	info := ui.state.info
	if info == nil {
		return nil
	}

	var mSize int
	for _, b := range ui.state.mempool {
		mSize += b.BlockSize
	}

	// Compute the total number of blocks on the mempool
	// We use the total BlockWeight / 4mm
	var w float64
	for _, b := range ui.state.mempool {
		w += float64(b.BlockWeight)
	}

	fmt.Fprintf(v, "%s %s, %s %s, %s %s",
		red("Unconfirmed Txs: "), white("%d", info.Size),
		blue("Mempool size"), white("%s (%d blocks)", fmtSize(mSize), ceil(w/4_000_000)),
		blue("Tx weight per second"), red("%d vBytes/s", ui.state.vBytesPerSecond),
	)
	return nil
}

func (ui *UI) onBlockClick(g *gocui.Gui, v *gocui.View) error {
	name := v.Name()
	if strings.HasPrefix(name, "projected-block-") {
		id := strings.TrimPrefix(name, "projected-block-")
		n, _ := strconv.Atoi(id)
		return ui.fd.FetchProjection(n)
	}

	if strings.HasPrefix(name, "block-") {
		id := strings.TrimPrefix(name, "block-")
		n, _ := strconv.Atoi(id)
		return ui.fd.FetchBlock(n)
	}

	return nil
}
