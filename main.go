package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

type (
	pkg struct {
		name      string
		wikiTitle string
		aur       bool
	}
	category struct {
		name string
		pkgs []pkg
	}
	ghost struct {
		categories []*category
	}
)

func (p *pkg) archWikiURL() string {
	if p.wikiTitle == "" {
		return ""
	}
	return "https://wiki.archlinux.org/title/" + p.wikiTitle
}

func (p *pkg) archLinuxURL() string {
	var s string
	if p.aur {
		s = "https://aur.archlinux.org/packages?K="
	} else {
		s = "https://archlinux.org/packages/?q="
	}
	return s + p.name
}

func (g *ghost) writeMD(w io.Writer) {
	for _, cat := range g.categories {
		fmt.Fprintf(w, "## %s:\n\n", cat.name)
		for _, pkg := range cat.pkgs {
			fmt.Fprintf(w, "[%s](%s)", pkg.name, pkg.archLinuxURL())
			if pkg.aur {
				fmt.Fprint(w, " (aur)")
			}
			if pkg.wikiTitle != "" {
				fmt.Fprintf(w, " - [%s](%s)", pkg.wikiTitle, pkg.archWikiURL())
			}
			fmt.Fprintln(w)
			fmt.Fprintln(w)
		}
	}
}

var (
	categoryRe = regexp.MustCompile(`^(.+):$`)
	packageRe  = regexp.MustCompile(`^(\+)?(\S+)(?: @(\S+))?$`)
)

func newGhostFromFile(name string) (*ghost, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return newGhost(f)
}

func newGhost(r io.Reader) (*ghost, error) {
	scanner := bufio.NewScanner(r)
	g := ghost{
		categories: make([]*category, 0, 1),
	}
	var curCat *category
	for scanner.Scan() {
		line := scanner.Text()
		if catMatch := categoryRe.FindStringSubmatch(line); catMatch != nil {
			if curCat != nil {
				g.categories = append(g.categories, curCat)
			}
			curCat = &category{
				name: catMatch[1],
				pkgs: make([]pkg, 0, 1),
			}
		} else if pkgMatch := packageRe.FindStringSubmatch(line); pkgMatch != nil {
			if curCat == nil {
				continue
			}
			curCat.pkgs = append(curCat.pkgs, pkg{
				name:      pkgMatch[2],
				wikiTitle: pkgMatch[3],
				aur:       pkgMatch[1] != "",
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if curCat != nil {
		g.categories = append(g.categories, curCat)
	}
	return &g, nil
}

func main() {
	for _, arg := range os.Args[1:] {
		g, err := newGhostFromFile(arg)
		if err != nil {
			panic(err)
		}
		g.writeMD(os.Stdout)
	}
}
