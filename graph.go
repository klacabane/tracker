package main

import (
	"fmt"
	"sort"
)

var (
	valueMargin = 4
	labelMargin = 8
)

type Graph struct {
	height, width int

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

func NewGraph(points map[string]float64) *Graph {
	g := &Graph{
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

	var (
		line         string
		penult, last bool

		spaces = spacestr(g.paddingLeft)
	)

	for i := 0; i < g.height; i++ {
		line = ""
		penult = i == g.height-2
		last = i == g.height-1

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
						for k := 0; k < diff; k++ {
							line += " "
						}
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
	// abscissa
	sort.Float64s(g.values)

	height := valueMargin / 2
	for i := len(g.values) - 1; i >= 0; i-- {
		val := g.values[i]

		if length := len(fmt.Sprintf("%d", int(val))); length > g.paddingLeft {
			g.paddingLeft = length
		}
		g.y[height] = val
		height += valueMargin
	}
	g.height = height + 2

	// ordinate
	current := g.paddingLeft + 1 + labelMargin
	for label, _ := range g.points {
		g.x[current] = label
		current += len(label) + labelMargin
	}

	g.width = current - labelMargin/2

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
