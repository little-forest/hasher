package cmd

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/morikuni/aec"
)

type HasherProgressViewer struct {
	NumOfWorkers int
	Verbose      bool
	total        int
}

func NewHasherProgressViewer(numOfWorkers int, verbose bool) *HasherProgressViewer {
	return &HasherProgressViewer{
		NumOfWorkers: numOfWorkers,
		Verbose:      verbose,
	}
}

func (p *HasherProgressViewer) SetTotal(total int) {
	if total >= 0 {
		p.total = total
	}
}

func (p HasherProgressViewer) IsVerbose() bool {
	return p.Verbose
}

func (p HasherProgressViewer) Setup() {
	if !p.Verbose {
		return
	}
	row := p.NumOfWorkers
	fmt.Print(strings.Repeat("\n", row))
	fmt.Print(aec.Hide)
	fmt.Print(aec.Up(uint(row)))
}

func (p HasherProgressViewer) Progress(workerId int, done int, total int, path string) {
	if !p.Verbose {
		return
	}

	if total < 0 && p.total >= 0 {
		total = p.total
	}

	path = chopPath(path)

	fmt.Print(aec.Down(uint(workerId)))
	fmt.Print("\x1b[0K") // delete line after cursor
	fmt.Printf("[Worker-%d] : %s", workerId, path)
	fmt.Print(aec.NextLine(uint(p.NumOfWorkers - workerId)))
	fmt.Printf("%d / %d", done, total)
	fmt.Print(aec.PreviousLine(uint(p.NumOfWorkers)))
}

func (p HasherProgressViewer) ShowError(msg string) {
	if p.Verbose {
		// Insert one line to bottom
		fmt.Print(aec.NextLine(uint(p.NumOfWorkers)))
		fmt.Printf("\n")
		fmt.Print(aec.PreviousLine(uint(p.NumOfWorkers + 1)))

		fmt.Print("\x1b[1L") // insert one line
	}
	// TODO: when stderr is redirected to file, display is broken
	fmt.Fprintln(os.Stderr, C_lred.Apply(msg))
}

func (p HasherProgressViewer) TearDown() {
	if !p.Verbose {
		return
	}
	row := p.NumOfWorkers + 1
	fmt.Print(aec.Down(uint(row)))
	fmt.Print(aec.Show)
}

func chopPath(path string) string {
	paths := strings.Split(path, string(os.PathSeparator))
	l := len(paths)
	if l > 2 {
		return filepath.Join("...", paths[l-2], paths[l-1])
	} else {
		return path
	}
}

// check implementation
var _ core.ProgressWatcher = &HasherProgressViewer{}
