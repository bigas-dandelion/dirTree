package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Node interface {
	fmt.Stringer
}

type File struct {
	name string
	size int64
}

type Directory struct {
	name  string
	files []Node
}

func (f File) String() string {
	return f.name + ", " + strconv.FormatInt(f.size, 10)
}

func (d Directory) String() string {
	return d.name
}

func WalkDir(path string, dirSlice []Node) ([]Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return []Node{}, err
	}

	dirs, err := file.Readdir(0)
	if err != nil {
		return []Node{}, err
	}
	file.Close()

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name() < dirs[j].Name()
	})

	for _, el := range dirs {
		if !el.IsDir() {
			dirSlice = append(dirSlice, File{
				name: el.Name(),
				size: el.Size(),
			})
		} else {
			child, _ := WalkDir(filepath.Join(path, el.Name()), []Node{})
			dirSlice = append(dirSlice, Directory{
				files: child,
				name:  el.Name(),
			})
		}
	}

	return dirSlice, err
}

func printDir(out io.Writer, nodes []Node, prefixes []string) {
	if len(nodes) == 0 {
		return
	}

	fmt.Fprintf(out, "%s", strings.Join(prefixes, ""))

	node := nodes[0]

	if len(nodes) == 1 {
		fmt.Fprintf(out, "%s%s\n", "└───", node)
		if directory, ok := node.(Directory); ok {
			printDir(out, directory.files, append(prefixes, "\t"))
		}
		return
	}

	fmt.Fprintf(out, "%s%s\n", "├───", node)
	if directory, ok := node.(Directory); ok {
		printDir(out, directory.files, append(prefixes, "│\t"))
	}

	printDir(out, nodes[1:], prefixes)
}

func PrintDir(out io.Writer, nodes []Node, prefix []string) {
	for i, node := range nodes {
		fmt.Fprintf(out, "%s", strings.Join(prefix, ""))

		isLast := i == len(nodes)-1

		if isLast {
			fmt.Fprintf(out, "└───%s\n", node)
			if dir, ok := node.(Directory); ok {
				PrintDir(out, dir.files, append(prefix, "\t"))
			}
		} else {
			fmt.Fprintf(out, "├───%s\n", node)
			if dir, ok := node.(Directory); ok {
				PrintDir(out, dir.files, append(prefix, "│\t"))
			}
		}
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	files, err := WalkDir(path, []Node{})
	if err != nil {
		return err
	}

	PrintDir(out, files, []string{})
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
