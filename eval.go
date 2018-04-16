package main

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

type node struct {
	Char      rune
	Children  map[rune]*node
	Frequency uint64
}

func (root node) Eval(prefix string) bool {
	pre := []rune(prefix)
	for i := range pre {
		head := &root
		var j int
		for j = i; j < len(pre); j++ {
			var exists bool
			head, exists = head.Children[pre[j]]
			if !exists {
				break
			}
		}
		if j == len(pre) {
			var maxkey rune
			var maxfreq uint64
			for key, val := range head.Children {
				if val.Frequency > maxfreq {
					maxkey = key
					maxfreq = val.Frequency
				}
			}
			if maxfreq > 0 {
				return maxkey == ' '
			}
		}
	}
	return false
}

func reverseStr(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func main() {
	fileName := os.Args[1]
	modelFileName := os.Args[2]
	reverse := false
	if len(os.Args) > 3 {
		reverse = true
	}
	modelFile, err := os.Open(modelFileName)
	if err != nil {
		log.Fatalf("Opening %s: %v", modelFileName, err)
	}
	defer modelFile.Close()
	decoder := gob.NewDecoder(modelFile)
	root := node{Children: map[rune]*node{}}
	for {
		var key rune
		var val node
		err = decoder.Decode(&key)
		if err != nil {
			break
		}
		err = decoder.Decode(&val)
		if err != nil {
			log.Fatalf("Decoding: %v", err)
		}
		root.Children[key] = &val
	}
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Cannot read %s: %v", fileName, err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var line []string
	count := 0
	precision := 0.0
	recall := 0.0
	for line, err = reader.Read(); err == nil; line, err = reader.Read() {
		count++
		if count%100000 == 0 {
			log.Printf("%d %f %f", count, precision/float64(count), recall/float64(count))
		}
		wins := 0
		fails := 0
		real := 0
		prefix := ""
		str := line[1]
		if reverse {
			str = reverseStr(str)
		}
		result := make([]bool, 1, len(str))
		for i, c := range str {
			if i == 0 {
				prefix += string(c)
				continue
			}
			split := root.Eval(prefix)
			result = append(result, split)
			if split {
				if c == ' ' {
					wins++
				} else {
					fails++
				}
			}
			if c == ' ' {
				real++
			}
			if split && c != ' ' {
				prefix += " "
			}
			prefix += string(c)
		}
		if reverse {
			for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
				result[i], result[j] = result[j], result[i]
			}
		}
		for _, v := range result {
			var iv int
			if v {
				iv = 1
			}
			fmt.Printf("%d ", iv)
		}
		fmt.Printf("\n")
		if wins+fails > 0 {
			precision += float64(wins) / float64(wins+fails)
		}
		if real > 0 {
			recall += float64(wins) / float64(real)
		}
	}
	fmt.Printf("Precision: %f\n", precision/float64(count))
	fmt.Printf("Recall: %f\n", recall/float64(count))
}
