package ui

import (
	"context"
	"fmt"
	"sort"
	"sync"

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
	fd.m.Lock()
	defer fd.m.Unlock()

	if fn := fd.cancelFn; fn != nil {
		fn()
	}
	fd.loading = true

	ctx := fd.newCtx()
	go func() {
		fees, err := client.GetProjectedFee(ctx, n)
		if err != nil {
			return
		}

		fd.fees = fees
		fd.loading = false
		fd.gui.Update(fd.Layout)
	}()

	return nil
}

func (fd *FeeDistribution) FetchBlock(n int) error {
	fd.m.Lock()
	defer fd.m.Unlock()

	if fn := fd.cancelFn; fn != nil {
		fn()
	}
	fd.loading = true

	ctx := fd.newCtx()
	go func() {
		fees, err := client.GetBlockFee(ctx, n)
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
	v, err := g.SetView(name, x/2-20, y/2-7, x/2+20, y/2+7)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Fee distribution"
		g.SetCurrentView(name)
		g.SetViewOnTop(name)
		g.SetKeybinding(name, gocui.KeyEsc, gocui.ModNone, fd.close)
	}

	v.Clear()

	if fd.loading == true {
		fmt.Fprint(v, "Loading...")
		return nil
	}

	min, max := 99999, 0
	for _, f := range fd.fees {
		fee := int(f.FPV)
		if fee < min {
			min = fee
		}
		if fee > max {
			max = fee
		}
	}
	fmt.Fprintf(v, "Fee span: %d - %d sat/vByte\n", min, max)

	fmt.Fprintf(v, "Tx count: %d transactions\n", len(fd.fees))

	sort.Sort(fd.fees)
	fmt.Fprintf(v, "Median: ~%d sat/vBytes", int(fd.fees[len(fd.fees)/2].FPV))

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
