package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
)

type node struct {
	Char      rune
	Children  map[rune]*node
	Frequency uint64
}

func (parent *node) GetChild(at rune, alloc *allocator) *node {
	nd, exists := parent.Children[at]
	if !exists {
		nd = alloc.New(at)
		parent.Children[at] = nd
	}
	return nd
}

type allocatorChunk struct {
	Data []node
	Pos  int
}

func newAllocatorChunk() *allocatorChunk {
	return &allocatorChunk{Data: make([]node, 1<<16), Pos: 0}
}

type allocator struct {
	mem []*allocatorChunk
}

func (alloc *allocator) New(char rune) *node {
	if alloc.mem == nil {
		alloc.mem = []*allocatorChunk{newAllocatorChunk()}
	}
	chunk := alloc.mem[len(alloc.mem)-1]
	if chunk.Pos == len(chunk.Data) {
		chunk = newAllocatorChunk()
		alloc.mem = append(alloc.mem, chunk)
	}
	nd := &chunk.Data[chunk.Pos]
	nd.Char = char
	nd.Children = map[rune]*node{}
	chunk.Pos++
	return nd
}

func (alloc *allocator) Len() int64 {
	if alloc.mem == nil {
		return 0
	}
	return int64(len(alloc.mem)) * int64(len(alloc.mem[0].Data))
}

func applyString(text string, depth int, root *node, alloc *allocator) {
	runes := []rune(text)
	fullRun := false
	for currentDepth := 1; currentDepth <= depth; currentDepth++ {
		slided := false
		for i := range runes {
			start := i - currentDepth
			if start < 0 {
				continue
			}
			slided = true
			head := root
			for j := start; j <= i; j++ {
				nd := head.GetChild(runes[j], alloc)
				nd.Frequency++
				head = nd
			}
		}
		if !slided && !fullRun {
			fullRun = true
			head := root
			for j := 0; j < len(runes); j++ {
				nd := head.GetChild(runes[j], alloc)
				nd.Frequency++
				head = nd
			}
		}
	}
}

func main() {
	alloc := &allocator{}
	fileName := os.Args[1]
	depth, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Parsing os.Args[2]: %v", err)
	}
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Cannot read %s: %v", fileName, err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var line []string
	root := node{Children: map[rune]*node{}}
	i := 0
	for line, err = reader.Read(); err == nil; line, err = reader.Read() {
		applyString(line[1], depth, &root, alloc)
		i++
		if i%100000 == 0 {
			log.Printf("%d", i)
		}
	}
	log.Printf("Size: %d\n", alloc.Len())
	if err != io.EOF {
		log.Fatalf("Reading CSV: %v", err)
	}
	//encoder := gob.NewEncoder(os.Stdout)
	//encoder.Encode(root)
}
