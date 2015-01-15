package main

import (
	"fmt"
	"sort"
	"strings"
)

var (
	maxHeight = 20

	marginY = 4
	marginX = 8
)

type Graph struct {
	height, width int

	labels      []string
	values      []float64
	x           map[int]string
	y           map[int]float64
	points      map[string]float64
	coordinates []coord
	paddingLeft int
}

type coord struct {
	x int
	y int
}

func NewGraph(labels []string, points map[string]float64) *Graph {
	g := &Graph{
		labels:      labels,
		values:      make([]float64, 0),
		x:           make(map[int]string),
		y:           make(map[int]float64),
		coordinates: make([]coord, 0),
		points:      points,
	}

	for _, val := range g.points {
		if !contains(g.values, val) {
			g.values = append(g.values, val)
		}
	}

	return g
}

func (g *Graph) Print() {
	g.compute()

	spaces := spacestr(g.paddingLeft)
	for i := 0; i < g.height; i++ {
		last, penult, line := i == g.height-1, i == g.height-2, ""
		for j := 0; j < g.width; j++ {
			if last {
				if label, ok := g.x[j]; ok {
					line += label
					j += len(label) - 1
				} else if j == g.paddingLeft {
					line += "|"
				} else {
					line += " "
				}
			} else if penult {
				if j == g.paddingLeft {
					line += "|"
				} else {
					line += "_"
				}
			} else if j == g.paddingLeft {
				line += "|"
			} else if j == 0 {
				if val, ok := g.y[i]; ok {
					sval := fmt.Sprintf("%d", int(val))

					if diff := g.paddingLeft - len(sval); diff > 0 {
						line += strings.Repeat(" ", diff)
					}
					line += sval
				} else {
					line += spaces()
				}
			} else if g.hasPoint(j, i) {
				line += "+"
			} else if j > g.paddingLeft {
				line += " "
			}

		}
		fmt.Print(line, "\n")
	}
}

func (g *Graph) hasPoint(x, y int) bool {
	for _, c := range g.coordinates {
		if c.x == x && c.y == y {
			return true
		}
	}
	return false
}

func (g *Graph) compute() {
	ch := make(chan struct{}, 1)

	// abscissa
	go g.computeAbs(ch)

	g.setPaddingLeft()

	// ordinate
	x := g.paddingLeft + 1 + marginX/2
	for _, label := range g.labels {
		g.x[x] = label
		x += len(label) + marginX
	}

	g.width = x - marginX/2

	<-ch
	for label, value := range g.points {
		var c coord
		for pos, lab := range g.x {
			if lab == label {
				c.x = pos + len(lab)/2
			}
		}
		for pos, val := range g.y {
			if val == value {
				c.y = pos
			}
		}
		g.coordinates = append(g.coordinates, c)
	}
}

func spacestr(length int) func() string {
	spaces := ""
	for i := 0; i < length; i++ {
		spaces += " "
	}

	return func() string {
		return spaces
	}
}

func contains(sl []float64, val float64) bool {
	for _, v := range sl {
		if v == val {
			return true
		}
	}
	return false
}

func (g *Graph) computeAbs(ch chan<- struct{}) {
	sort.Float64s(g.values)

	max, min, topMargin := g.values[len(g.values)-1], g.values[0], marginY/2
	gap := int(max - min)

	// increase gap to nearest number % maxHeight == 0
	for gap%maxHeight > 0 {
		gap++
	}
	scale := gap / maxHeight

	for i := len(g.values) - 1; i >= 0; i-- {
		val := g.values[i]

		diff := int(max - val)
		if diff > 0 {
			y := (diff / scale) + topMargin
			if _, ok := g.y[y]; ok {
				for ok {
					y++
					_, ok = g.y[y]
				}
			}

			gap := y - g.height
			g.height += gap
		} else {
			g.height += topMargin
		}
		g.y[g.height] = val
	}
	g.height += marginY

	ch <- struct{}{}
}

func (g *Graph) setPaddingLeft() {
	for _, value := range g.values {
		if width := len(fmt.Sprintf("%d", int(value))); width > g.paddingLeft {
			g.paddingLeft = width
		}
	}
}
