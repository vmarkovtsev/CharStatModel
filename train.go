package main

import (
	"encoding/csv"
	"encoding/gob"
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
	if parent.Children == nil {
		parent.Children = map[rune]*node{}
	}
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

func prune(root *node) int {
	stack := []*node{root}
	count := 0
	for {
		if len(stack) == 0 {
			return count
		}
		head := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		count++
		if head.Children != nil {
			leaves := true
			var maxFreq uint64
			var maxChild *node
			var maxRune rune
			for key, child := range head.Children {
				if child.Children != nil {
					leaves = false
					break
				}
				if child.Frequency > maxFreq {
					maxFreq = child.Frequency
					maxChild = child
					maxRune = key
				}
			}
			if leaves {
				head.Children = map[rune]*node{}
				head.Children[maxRune] = maxChild
				count++
			} else {
				for _, child := range head.Children {
					stack = append(stack, child)
				}
			}
		}
	}
}

func reverseStr(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
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
	reverse := false
	if len(os.Args) > 3 {
		reverse = true
	}
	for line, err = reader.Read(); err == nil; line, err = reader.Read() {
		str := line[1]
		if reverse {
			str = reverseStr(str)
		}
		applyString(str, depth, &root, alloc)
		i++
		if i%100000 == 0 {
			log.Printf("%d", i)
		}
	}
	log.Printf("Size: %d\n", alloc.Len())
	if err != io.EOF {
		log.Fatalf("Reading CSV: %v", err)
	}
	newSize := prune(&root)
	log.Printf("Pruned to %d - %f\n", newSize, float64(newSize)/float64(alloc.Len()))
	encoder := gob.NewEncoder(os.Stdout)
	for key, val := range root.Children {
		err = encoder.Encode(key)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		err = encoder.Encode(val)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}
