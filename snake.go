/* This snake game implementation was in large parts AI generated. */

package main

import (
	"fmt"
	"iter"
	"math/rand"
	"slices"
	"strings"
	"time"
)

const (
	WIDTH  = 30
	HEIGHT = 20
	SLEEP  = 50 * time.Millisecond
)

const (
	EMPTY = " "
	HEAD  = "O"
	BODY  = "o"
	TAIL  = "."
	FOOD  = "x"
)

type Point struct{ X, Y int }
type PointSet map[Point]struct{}

var DIRS = []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

func (s PointSet) containsPoint(p Point) bool {
	_, ok := s[p]
	return ok
}

func hideCursor() string    { return "\x1b[?25l" }
func showCursor() string    { return "\x1b[?25h" }
func moveCursorTop() string { return "\x1b[H" }

func drawScreen(snake []Point, food Point, score int) []byte {
	snakeSet := make(PointSet)
	for _, p := range snake {
		snakeSet[p] = struct{}{}
	}

	var screen strings.Builder

	screen.WriteString(moveCursorTop())
	screen.WriteString("\n\n~~~~~~~~~~~~ SNAKE ~~~~~~~~~~~~~\n\n")

	screen.WriteString("┌")
	screen.WriteString(strings.Repeat("─", WIDTH))
	screen.WriteString("┐\n")

	for y := range HEIGHT {
		screen.WriteString("│")
		for x := range WIDTH {
			p := Point{x, y}
			ch := EMPTY
			switch {
			case p == food:
				ch = FOOD
			case p == snake[0]:
				ch = HEAD
			case p == snake[len(snake)-1]:
				ch = TAIL
			case snakeSet.containsPoint(p):
				ch = BODY
			}
			screen.WriteString(ch)
		}
		screen.WriteString("│\n")
	}

	screen.WriteString("└")
	screen.WriteString(strings.Repeat("─", WIDTH))
	screen.WriteString("┘\n\n")
	fmt.Fprintf(&screen, "Score: %d\n\n", score)
	return []byte(screen.String())
}

func randomFood(snake []Point) Point {
	snakeSet := make(PointSet)
	for _, p := range snake {
		snakeSet[p] = struct{}{}
	}

	for {
		p := Point{rand.Intn(WIDTH), rand.Intn(HEIGHT)}
		if !snakeSet.containsPoint(p) {
			return p
		}
	}
}

func isInBound(p Point) bool {
	return 0 <= p.X && p.X < WIDTH && 0 <= p.Y && p.Y < HEIGHT
}

func bfs(start, goal Point, snake []Point) []Point {
	body := make(PointSet)
	for _, p := range snake[:len(snake)-1] {
		body[p] = struct{}{}
	}

	queue := []Point{start}
	prev := make(map[Point]Point)
	prev[start] = start

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == goal {
			break
		}
		for _, d := range DIRS {
			nxt := Point{cur.X + d.X, cur.Y + d.Y}
			if isInBound(nxt) && !body.containsPoint(nxt) {
				if _, ok := prev[nxt]; !ok {
					prev[nxt] = cur
					queue = append(queue, nxt)
				}
			}
		}
	}

	if _, ok := prev[goal]; !ok {
		return nil
	}

	path := []Point{}
	for cur := goal; cur != start; {
		path = append(path, cur)
		cur = prev[cur]
	}
	slices.Reverse(path)
	return path
}

func chooseMove(snake []Point, food Point) *Point {
	head := snake[0]
	tail := snake[len(snake)-1]

	pathFood := bfs(head, food, snake)
	if len(pathFood) > 0 {
		return &pathFood[0]
	}

	pathTail := bfs(head, tail, snake)
	if len(pathTail) > 0 {
		return &pathTail[0]
	}

	body := make(PointSet)
	for _, p := range snake[:len(snake)-1] {
		body[p] = struct{}{}
	}
	for _, d := range DIRS {
		nxt := Point{head.X + d.X, head.Y + d.Y}
		if isInBound(nxt) && !body.containsPoint(nxt) {
			return &nxt
		}
	}

	return nil
}

func SnakeGame() iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		snake := []Point{{WIDTH / 2, HEIGHT / 2}}
		food := randomFood(snake)
		score := 0

		if !yield([]byte("\x1b[2J" + hideCursor())) {
			return
		}

		for {
			move := chooseMove(snake, food)
			if move == nil {
				data := drawScreen(snake, food, score)
				data = append(data, []byte("Game Over!!\n\n")...)
				if !yield(data) {
					return
				}
				break
			}

			snake = append([]Point{*move}, snake...)
			if *move == food {
				score++
				food = randomFood(snake)
			} else {
				snake = snake[:len(snake)-1]
			}

			if !yield(drawScreen(snake, food, score)) {
				return
			}
		}

		if !yield([]byte(showCursor())) {
			return
		}
	}
}
