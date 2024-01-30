package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image/docker"
)

func main() {
	user := flag.Bool("u", false, "show user")
	flag.Parse()
	name := flag.Arg(0)

	if name == "" {
		panic("specify image name to fetch")
	}

	engine := docker.NewResolverFromEngine()
	img, err := engine.Fetch(name)
	if err != nil {
		log.Fatal(err)
	}

	var w io.Writer
	w = os.Stdout
	// if *uniq {
	// 	drw := drw.NewWriter(os.Stdout, '\n', drw.NewMapCache(0))
	// 	w = drw
	// }

	nodes := map[string]*filetree.FileNode{}

	visitor := func(node *filetree.FileNode) error {
		nodes[node.Path()] = node

		// fmt.Fprintf(w, "%s\t%d\t%s\n", node.Data.FileInfo.Mode.String(), node.Size, node.Path())
		return nil
	}

	visitEvaluator := func(node *filetree.FileNode) bool {
		return node.IsLeaf()
	}

	for _, t := range img.Trees {
		err := t.VisitDepthChildFirst(visitor, visitEvaluator)
		if err != nil {
			log.Fatal(err)
		}
	}

	var listNodes []*filetree.FileNode
	for _, n := range nodes {
		listNodes = append(listNodes, n)
	}

	slices.SortFunc(listNodes, func(a, b *filetree.FileNode) int {
		return strings.Compare(a.Path(), b.Path())
	})

	for _, node := range listNodes {
		uinfo := ""
		if *user {
			uinfo = fmt.Sprintf("%d\t%d\t", node.Data.FileInfo.Uid, node.Data.FileInfo.Gid)
		}

		fmt.Fprintf(w, "%s\t%s%s\t%s\n",
			node.Data.FileInfo.Mode.String(),
			uinfo,
			formatSize(node.Data.FileInfo),
			formatPath(node))
	}
}

func formatSize(info filetree.FileInfo) string {
	if info.TypeFlag == tar.TypeDir {
		return "-"
	}
	if info.Size == 0 {
		return "0"
	}
	return strings.Replace(humanize.Bytes(uint64(info.Size)), " ", "", -1)
}

func formatPath(node *filetree.FileNode) string {
	ret := node.Path()
	switch node.Data.FileInfo.TypeFlag {
	case tar.TypeLink, tar.TypeSymlink:
		ret += fmt.Sprintf(" -> %s", node.Data.FileInfo.Linkname)
	}
	return ret
}
