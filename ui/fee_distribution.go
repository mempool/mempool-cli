package ui

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/fatih/color"
	"github.com/gchaincl/mempool/client"
	"github.com/jroimartin/gocui"
)

type FeeDistribution struct {
	gui *gocui.Gui

	m           sync.Mutex
	loading     bool
	isProjected bool
	cancelFn    context.CancelFunc
	fees        client.Fees
	title       string
}

func NewFeeDistribution(g *gocui.Gui) *FeeDistribution {
	return &FeeDistribution{gui: g}
}

func (fd *FeeDistribution) newCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	fd.cancelFn = cancel
	return ctx
}

func (fd *FeeDistribution) FetchProjection(n int) error {
	fn := func(ctx context.Context) (client.Fees, error) {
		return client.GetProjectedFee(ctx, n)
	}
	return fd.fetch(fn)
}

func (fd *FeeDistribution) FetchBlock(n int) error {
	fn := func(ctx context.Context) (client.Fees, error) {
		return client.GetBlockFee(ctx, n)
	}
	return fd.fetch(fn)
}

func (fd *FeeDistribution) fetch(fn func(ctx context.Context) (client.Fees, error)) error {
	fd.m.Lock()
	defer fd.m.Unlock()

	if fn := fd.cancelFn; fn != nil {
		fn()
	}
	fd.loading = true

	ctx := fd.newCtx()
	go func() {
		fees, err := fn(ctx)
		if err != nil {
			return
		}

		fd.fees = fees
		fd.loading = false
		fd.gui.Update(fd.Layout)
	}()

	return nil
}

func (fd *FeeDistribution) Layout(g *gocui.Gui) error {
	fd.m.Lock()
	defer fd.m.Unlock()

	if fd.loading == false && (fd.fees == nil || len(fd.fees) == 0) {
		return nil
	}

	x, y := g.Size()
	name := "fee_distribution"
	v, err := g.SetView(name, x/2-20, y/2-3, x/2+20, y/2+3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Fee distribution" + fd.title + "('esc' to close)"
		g.SetCurrentView(name)
		g.SetViewOnTop(name)
		g.SetKeybinding(name, gocui.KeyEsc, gocui.ModNone, fd.close)
	}

	v.Clear()

	if fd.loading == true {
		fmt.Fprint(v, "Loading...")
		return nil
	}

	sort.Sort(fd.fees)
	txs := len(fd.fees)
	min, max := fd.fees[0].FPV, fd.fees[txs-1].FPV

	var (
		white  = color.New(color.FgWhite).SprintfFunc()
		yellow = color.New(color.FgYellow).SprintfFunc()
	)
	fmt.Fprintf(v, white("Fee span:")+" %d - %d "+yellow("sat/vByte\n"), ceil(min), ceil(max))
	fmt.Fprintf(v, white("Tx count:")+" %d "+white("transactions\n"), txs)
	fmt.Fprintf(v, white("Median:  ")+" ~%d "+yellow("sat/vBytes\n"), ceil(fd.fees[txs/2].FPV))

	return nil
}

func (fd *FeeDistribution) close(g *gocui.Gui, v *gocui.View) error {
	fd.m.Lock()
	defer fd.m.Unlock()

	if fn := fd.cancelFn; fn != nil {
		fn()
	}

	fd.cancelFn = nil
	fd.loading = false
	fd.fees = nil
	g.DeleteView(v.Name())
	return nil
}
