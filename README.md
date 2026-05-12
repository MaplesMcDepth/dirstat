# dirstat

![CI](https://github.com/MaplesMcDepth/dirstat/actions/workflows/ci.yml/badge.svg)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)


Directory size analyzer — like `du` but readable.

## Install

```bash
go install github.com/MaplesMcDepth/dirstat/cmd/dirstat@latest
```

## Commands

### Top N largest items
```bash
dirstat                    # Top 20 in current dir
dirstat -n 10 /var/log     # Top 10 in /var/log
dirstat -h -n 5 .          # Human-readable sizes
```

### Tree view
```bash
dirstat -t ~/projects      # Tree with sizes
dirstat -t -d 2 -h .       # Tree, depth 2, human sizes
```

### Options

| Flag | Description |
|------|-------------|
| `-n int` | Show top N items (default 20) |
| `-d int` | Max depth (-1 = unlimited) |
| `-t` | Tree view |
| `-a` | Include hidden files |
| `-s string` | Sort by: size, name, count |
| `-h` | Human-readable sizes |
