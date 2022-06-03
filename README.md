# go_life: A Go implementation of Conway's Game of Life

[![Go Reference](https://pkg.go.dev/badge/github.com/kerrigan29a/go_life.svg)](https://pkg.go.dev/github.com/kerrigan29a/go_life)
[![Go Report Card](https://goreportcard.com/badge/github.com/kerrigan29a/go_life)](https://goreportcard.com/report/github.com/kerrigan29a/go_life)

![Demo](demo.gif)

# Keymap
- `ESC`, `Ctrl+C`, `q`: Exit
- `p`: Pause / Resume
- `c`: Redraw the screen
- `n`: (On pause) Next generation
- `Right click`: Turn ON all the 8 cells in the current position.
- `Any other click`: Turn OFF all the 8 cells in the current position.

# Why mouse clicks turn ON/OFF 8 cells?
This program uses [Braille characters](https://en.wikipedia.org/wiki/Braille_Patterns) to represent the cells so, when you click on the screen the program cannot differentiate which of the 8 cells you want to change.