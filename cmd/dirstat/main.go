package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileInfo struct {
	Path string
	Size int64
	Dir  bool
	Depth int
}

func usage() {
	fmt.Print(`dirstat — Directory size analyzer

Usage: dirstat [options] [path]

Options:
  -n int       Show top N items (default 20)
  -d int       Max depth (default unlimited)
  -t           Show tree view
  -a           Include hidden files
  -s string    Sort by: size, name, count (default size)
  -h           Human-readable sizes

Examples:
  dirstat                      # Analyze current directory
  dirstat -n 10 /var/log       # Top 10 items in /var/log
  dirstat -t -d 2 ~/projects   # Tree view, depth 2
  dirstat -h -s name .         # Human sizes, sorted by name
`)
}

func main() {
	var (
		topN     = flag.Int("n", 20, "Show top N items")
		maxDepth = flag.Int("d", -1, "Max depth (-1 = unlimited)")
		treeView = flag.Bool("t", false, "Tree view")
		allFiles = flag.Bool("a", false, "Include hidden files")
		sortBy   = flag.String("s", "size", "Sort by: size, name, count")
		human    = flag.Bool("h", false, "Human-readable sizes")
	)
	flag.Usage = usage
	flag.Parse()

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	info, err := os.Stat(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintln(os.Stderr, "Not a directory:", root)
		os.Exit(1)
	}

	if *treeView {
		showTree(root, *maxDepth, *allFiles, *human)
	} else {
		showTop(root, *topN, *maxDepth, *allFiles, *sortBy, *human)
	}
}

func showTree(root string, maxDepth int, allFiles, humanSizes bool) {
	printTree(root, "", maxDepth, 0, allFiles, humanSizes)
}

func printTree(path, prefix string, maxDepth, depth int, allFiles, humanSizes bool) {
	if maxDepth >= 0 && depth > maxDepth {
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	// Sort entries: dirs first, then by name
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	// Filter hidden
	if !allFiles {
		var filtered []os.DirEntry
		for _, e := range entries {
			if !strings.HasPrefix(e.Name(), ".") {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	for i, entry := range entries {
		isLast := i == len(entries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		name := entry.Name()
		fullPath := filepath.Join(path, name)

		if entry.IsDir() {
			size, _ := dirSize(fullPath)
			if humanSizes {
				fmt.Printf("%s%s%s/ %s\n", prefix, connector, name, humanSize(size))
			} else {
				fmt.Printf("%s%s%s/ %d\n", prefix, connector, name, size)
			}

			nextPrefix := prefix
			if isLast {
				nextPrefix += "    "
			} else {
				nextPrefix += "│   "
			}
			printTree(fullPath, nextPrefix, maxDepth, depth+1, allFiles, humanSizes)
		} else {
			info, _ := entry.Info()
			if info != nil {
				if humanSizes {
					fmt.Printf("%s%s%s %s\n", prefix, connector, name, humanSize(info.Size()))
				} else {
					fmt.Printf("%s%s%s %d\n", prefix, connector, name, info.Size())
				}
			}
		}
	}
}

func showTop(root string, topN, maxDepth int, allFiles bool, sortBy string, humanSizes bool) {
	var items []FileInfo

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == root {
			return nil
		}

		// Skip hidden
		if !allFiles && strings.HasPrefix(filepath.Base(path), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check depth
		rel, _ := filepath.Rel(root, path)
		depth := strings.Count(rel, string(os.PathSeparator))
		if maxDepth >= 0 && depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		var size int64
		if d.IsDir() {
			size, _ = dirSize(path)
		} else {
			info, _ := d.Info()
			if info != nil {
				size = info.Size()
			}
		}

		items = append(items, FileInfo{
			Path:  rel,
			Size:  size,
			Dir:   d.IsDir(),
			Depth: depth,
		})

		return nil
	})

	// Sort
	switch sortBy {
	case "name":
		sort.Slice(items, func(i, j int) bool {
			return items[i].Path < items[j].Path
		})
	case "count":
		// Sort by number of items in dir (approximate by keeping dirs first)
		sort.Slice(items, func(i, j int) bool {
			if items[i].Dir != items[j].Dir {
				return items[i].Dir
			}
			return items[i].Size > items[j].Size
		})
	default: // size
		sort.Slice(items, func(i, j int) bool {
			return items[i].Size > items[j].Size
		})
	}

	// Show top N
	if topN > 0 && len(items) > topN {
		items = items[:topN]
	}

	// Print header
	fmt.Printf("%-8s %-7s %s\n", "SIZE", "TYPE", "PATH")
	fmt.Println(strings.Repeat("-", 70))

	for _, item := range items {
		var typeStr string
		if item.Dir {
			typeStr = "dir"
		} else {
			typeStr = "file"
		}

		var sizeStr string
		if humanSizes {
			sizeStr = humanSize(item.Size)
		} else {
			sizeStr = fmt.Sprintf("%d", item.Size)
		}

		fmt.Printf("%-8s %-7s %s\n", sizeStr, typeStr, item.Path)
	}

	// Total
	var total int64
	for _, item := range items {
		total += item.Size
	}
	fmt.Println(strings.Repeat("-", 70))
	if humanSizes {
		fmt.Printf("Total: %s (showing %d of %d items)\n", humanSize(total), len(items), len(items))
	} else {
		fmt.Printf("Total: %d (showing %d items)\n", total, len(items))
	}
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, _ := d.Info()
			if info != nil {
				size += info.Size()
			}
		}
		return nil
	})
	return size, err
}

func humanSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fT", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fM", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fK", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
