# cstatus

I was looking for Claude Code statusline options and discovered most of the popular ones are all written in JavaScript and run via `npx`. I guess nothing is inherently wrong with this, but this felt like the perfect use-case for using a language that could be compiled into a minimal binary. This project is the result of me making my own in Go.

## Installation

```bash
go install github.com/CS-5/cstatus@latest
cstatus install
```