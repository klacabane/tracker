package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	maxHeight = 20

	marginY = 4
	marginX = 8
)

type Graph struct {
	height, width int

	values []float64
	labels []string

	abs map[int]float64
	ord map[int]string

	points      map[string]float64
	coordinates []coord

	offset int
}

type coord struct {
	x int
	y int
}

func NewGraph(labels []string, points map[string]float64) *Graph {
	g := &Graph{
		labels:      labels,
		values:      make([]float64, 0),
		ord:         make(map[int]string),
		abs:         make(map[int]float64),
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

	offset := strings.Repeat(" ", g.offset)
	for i := 0; i < g.height; i++ {
		last, penult, line := i == g.height-1, i == g.height-2, ""
		for j := 0; j < g.width; j++ {
			if last {
				if label, ok := g.ord[j]; ok {
					line += label
					j += len(label) - 1
				} else if j == g.offset {
					line += "|"
				} else {
					line += " "
				}
			} else if penult {
				if j == g.offset {
					line += "|"
				} else {
					line += "_"
				}
			} else if j == g.offset {
				line += "|"
			} else if j == 0 {
				if val, ok := g.abs[i]; ok {
					sval := fmt.Sprintf("%v", val)

					if diff := g.offset - len(sval); diff > 0 {
						line += strings.Repeat(" ", diff)
					}
					line += sval
				} else {
					line += offset
				}
			} else if g.hasPoint(j, i) {
				line += "+"
			} else if j > g.offset {
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

	go g.computeAbs(ch)
	g.computeOrd()

	<-ch
	g.addCoordinates()
}

func (g *Graph) computeOrd() {
	g.setOffset()

	g.width = g.offset + 1 /*border*/ + marginX/2
	for _, label := range g.labels {
		g.ord[g.width] = label
		g.width += len(label) + marginX
	}

	g.width -= marginX / 2
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
			if _, ok := g.abs[y]; ok {
				for ok {
					y++
					_, ok = g.abs[y]
				}
			}

			gap := y - g.height
			g.height += gap
		} else {
			// top value
			g.height += topMargin
		}
		g.abs[g.height] = val
	}
	g.height += marginY

	ch <- struct{}{}
}

func (g *Graph) addCoordinates() {
	var wg sync.WaitGroup
	wg.Add(len(g.points))

	for label, value := range g.points {
		go g.addCoordinate(label, value, &wg)
	}
	wg.Wait()
}

func (g *Graph) addCoordinate(label string, value float64, wg *sync.WaitGroup) {
	defer wg.Done()

	var c coord
	for pos, lab := range g.ord {
		if lab == label {
			c.x = pos + len(lab)/2
		}
	}
	for pos, val := range g.abs {
		if val == value {
			c.y = pos
		}
	}
	g.coordinates = append(g.coordinates, c)
}

func (g *Graph) setOffset() {
	for _, value := range g.values {
		if width := len(fmt.Sprintf("%v", value)); width > g.offset {
			g.offset = width
		}
	}
}

// helpers
func contains(sl []float64, val float64) bool {
	for _, v := range sl {
		if v == val {
			return true
		}
	}
	return false
}
