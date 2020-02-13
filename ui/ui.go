package ui

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gchaincl/mempool/client"
	"github.com/jroimartin/gocui"
)

const (
	BLOCK_WIDTH = 22
)

type state struct {
	blocks    []client.Block
	projected []client.ProjectedBlock
	info      *client.MempoolInfo
}

type UI struct {
	gui   *gocui.Gui
	state state
}

func New() (*UI, error) {
	gui, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		return nil, err
	}

	ui := &UI{gui: gui}
	gui.SetManager(ui)
	gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	gui.SetKeybinding("", 'q', gocui.ModNone, quit)
	gui.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, ui.click)
	gui.Mouse = true
	gui.Highlight = true
	gui.SelFgColor = gocui.ColorWhite

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
	nBlocks := len(resp.Blocks)
	blocks := make([]client.Block, nBlocks)
	for i := 0; i < nBlocks; i++ {
		blocks[i] = resp.Blocks[len(resp.Blocks)-1-i]
	}

	if bs := blocks; len(bs) != 0 {
		ui.state.blocks = bs
	}

	if bs := resp.ProjectedBlocks; len(bs) != 0 {
		ui.state.projected = bs
	}

	if b := resp.Block; b != nil {
		ui.state.blocks = append(ui.state.blocks, *b)
	}

	if info := resp.MempoolInfo; info != nil {
		ui.state.info = info
	}

	// delete all the views
	for _, v := range ui.gui.Views() {
		ui.gui.DeleteView(v.Name())
	}

	ui.gui.Update(ui.Layout)
}

func (ui *UI) Layout(g *gocui.Gui) error {
	x, y := g.Size()

	// vertical layout is used if 8 blocks don't fit on the screen
	// when in vertical layout the mempool is shown in the top
	// and the blockchain in the bottom
	vertical := BLOCK_WIDTH*8 > x

	// draw projected blocks (mempool)
	for i, _ := range ui.state.projected {
		name := fmt.Sprintf("projected-block-%d", i)
		var x0, x1, y0, y1 int
		if vertical {
			x0 = x - (x/4)*(i+1)
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
		}
		v.Clear()
		if _, err := v.Write(ui.printProjectedBlock(i)); err != nil {
			return err
		}
	}

	if err := ui.separator(g, x, y, vertical); err != nil {
		return err
	}

	// draw blockchain blocks
	for i, block := range ui.state.blocks {
		name := fmt.Sprintf("block-%d", i)
		var x0, x1, y0, y1 int
		if vertical {
			x0 = x - (x/4)*(i+1)
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
		}
		v.Title = fmt.Sprintf("#%d", block.Height)
		v.Clear()
		if _, err := v.Write(ui.printBlock(i)); err != nil {
			return err
		}
	}

	if err := ui.info(g, x, y); err != nil {
		return err
	}

	return nil
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

	v, err := g.SetView("separtor", x0, y0, x1, y1)
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

var (
	white  = color.New(color.FgWhite).SprintfFunc()
	yellow = color.New(color.FgYellow).SprintfFunc()
	red    = color.New(color.FgRed).SprintfFunc()
	blue   = color.New(color.FgBlue).SprintfFunc()
)

func (ui *UI) info(g *gocui.Gui, x, y int) error {
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
	for _, b := range ui.state.projected {
		mSize += b.BlockSize
	}

	fmt.Fprintf(v, "%s %s, %s %s",
		red("Unconfirmed Txs: "),
		white("%d", info.Size),
		blue("Mempool size"),
		white("%s (%d block/s)", fmtSize(mSize), len(ui.state.projected)),
	)
	return nil
}

func ceil(f float64) int {
	return int(
		math.Ceil(f),
	)
}

func (ui *UI) printProjectedBlock(n int) []byte {
	block := ui.state.projected[n]

	lines := []string{
		white("   ~%d sat/vB       ", ceil(block.MedianFee)),
		yellow("  %d - %d sat/vB     ", ceil(block.MinFee), ceil(block.MaxFee)),
		"                       ",
		white("     %.2f MB              ", float64(block.BlockSize)/(1000*1000)),
		white(" %4d transactions     ", block.NTx),
		"                       ",
		"                       ",
		"                       ",
		"                       ",
		"                       ",
	}

	if n < 3 {
		lines[8] = white("   in ~%2d minutes     ", (n+1)*10)
		bg := color.New(color.BgRed).SprintfFunc()
		offset := 9 - int(
			float64(block.BlockWeight)/4000000.0*10,
		)
		for i := offset; i < len(lines); i++ {
			lines[i] = bg("%s", lines[i])
		}
	} else {
		bw := math.Ceil(float64(block.BlockWeight) / 4000000.0)
		lines[8] = white("    +%d blocks", int(bw))
	}

	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, strings.Join(lines, "\n"))
	return buf.Bytes()
}

func (ui *UI) printBlock(n int) []byte {
	block := ui.state.blocks[n]

	ago := time.Now().Unix() - int64(block.Time)
	lines := []string{
		white("    ~%d sat/Vb      ", ceil(block.MedianFee)),
		yellow("  %d - %d sat/vB     ", ceil(block.MinFee), ceil(block.MaxFee)),
		"                       ",
		white("     %.2f MB              ", float64(block.Size)/(1000*1000)),
		white(" %4d transactions     ", block.NTx),
		"                       ",
		"                       ",
		"                       ",
		white(" %s ago        ", fmtSeconds(ago)),
		"                       ",
	}

	bg := color.New(color.BgBlue).SprintfFunc()
	offset := 9 - int(
		float64(block.Weight)/4000000.0*10,
	)
	for i := offset; i < len(lines); i++ {
		lines[i] = bg("%s", lines[i])
	}

	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, strings.Join(lines, "\n"))
	return buf.Bytes()
}

func fmtSeconds(s int64) string {
	if s < 60 {
		return "< 1 minute"
	} else if s < 120 {
		return fmt.Sprintf("1 minute")
	} else if s < 3600 {
		return fmt.Sprintf("%d minutes", s/60)
	}
	return fmt.Sprintf("%d hours", s/3600)
}

func fmtSize(s int) string {
	if s < 1024*1024 {
		m := float64(s) / 1000.0
		return fmt.Sprintf("%dkB", ceil(m))
	}
	m := float64(s) / (1000.0 * 1000.0)
	return fmt.Sprintf("%fMB", ceil(m))
}

func (ui *UI) click(g *gocui.Gui, v *gocui.View) error {
	_, err := g.SetCurrentView(v.Name())
	return err
}
