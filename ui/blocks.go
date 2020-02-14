package ui

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gchaincl/mempool/client"
)

type Box struct {
	x     int
	lines []string
}

func (b *Box) Printf(c color.Attribute, f string, args ...interface{}) *Box {
	s := b.center(fmt.Sprintf(f, args...))
	s = color.New(c).Sprint(s)
	b.lines = append(b.lines, s)
	return b
}

func (b *Box) Append(s string) *Box {
	b.lines = append(b.lines, b.center(s))
	return b
}

func (b *Box) center(line string) string {
	l := len(line)
	if l > b.x {
		return line
	}
	offset := strings.Repeat(" ", (b.x-l)/2)
	return fmt.Sprintf("%s%s%s", offset, line, offset)
}

func (b *Box) Render(full int, bg color.Attribute) []byte {
	buf := bytes.NewBuffer(nil)
	fn := color.New(bg).SprintfFunc()
	for i, s := range b.lines {
		if 9-full <= i {
			s = fn("%s", s)
		}
		fmt.Fprintln(buf, s)
	}
	return buf.Bytes()
}

type ProjectedBlock client.ProjectedBlock

func (b ProjectedBlock) Print(n int, x, _y int) []byte {
	var footer string
	// Attach ETA to the first 3 blocks
	if n < 3 {
		footer = fmt.Sprintf("in ~%d minutes", (n+1)*10)
	} else {
		n := ceil(float64(b.BlockWeight) / 4000000.0)
		footer = fmt.Sprintf("+%d blocks", n)
	}

	box := &Box{x: x}
	box.Printf(color.FgWhite, "~%d sat/vB", ceil(b.MedianFee)).
		Printf(color.FgYellow, "%d-%d sat/vB", ceil(b.MinFee), ceil(b.MaxFee)).
		Append("").
		Printf(color.FgWhite, "%.2f MB", float64(b.BlockSize)/(1000*1000)).
		Printf(color.FgWhite, "%4d transactions", b.NTx).
		Append("").
		Append("").
		Append("").
		Printf(color.FgWhite, footer)

	// calculate how full is the block
	var full int
	if n < 3 {
		full = int(
			float64(b.BlockWeight) / 4000000 * 10,
		)
	}

	return box.Render(full, color.BgRed)
}

type Block client.Block

func (b Block) Print(n int, x, _y int) []byte {
	ago := time.Now().Unix() - int64(b.Time)
	box := &Box{x: x}

	box.Printf(color.FgWhite, "~%d sat/Vb", ceil(b.MedianFee)).
		Printf(color.FgYellow, "%d-%d sat/vB", ceil(b.MinFee), ceil(b.MaxFee)).
		Append("").
		Printf(color.FgWhite, "%.2f MB", float64(b.Size)/(1000*1000)).
		Printf(color.FgWhite, " %4d transactions", b.NTx).
		Append("").
		Append("").
		Append("").
		Printf(color.FgWhite, "%s ago", fmtSeconds(ago))

	full := int(
		float64(b.Weight) / 4000000 * 10,
	)

	return box.Render(full+4, color.BgBlue)
}
