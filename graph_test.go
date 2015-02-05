package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var graph *Graph

func TestMain(m *testing.M) {
	TRACKER_DIR, _ = os.Getwd()

	labels := []string{"1", "2000", "3", "4"}
	points := map[string]float64{
		"1":    210,
		"2000": 100,
		"3":    1000,
		"4":    100,
	}
	graph = NewGraph(labels, points)

	m.Run()
}

func TestComputeOrd(t *testing.T) {
	assert.Equal(t, 3, len(graph.values))

	graph.setOffset()
	assert.Equal(t, 4, graph.offset)

	graph.computeOrd()
	assert.Equal(t, "1", graph.ord[5+marginX/2])
	assert.Equal(t, "2000", graph.ord[5+marginX/2+1+marginX])
	assert.Equal(t, 5+marginX/2+7+4*marginX-marginX/2, graph.width)
}

func TestComputeAbs(t *testing.T) {
	ch := make(chan struct{}, 1)
	graph.computeAbs(ch)
	<-ch

	assert.Equal(t, 100, graph.values[0])
	assert.Equal(t, 1000, graph.abs[marginY/2])
	assert.Equal(t, maxHeight+marginY/2+marginY, graph.height)
}

func TestAddCoordinates(t *testing.T) {
	graph.addCoordinates()

	assert.Equal(t, 4, len(graph.coordinates))

	x, y := graph.width-(marginX/2+1+marginX+1), marginY/2
	assert.True(t, graph.hasPoint(x, y))
}
