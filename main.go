package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"
)

func cleanWord(word string) string {
	var sb strings.Builder
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			sb.WriteRune(unicode.ToLower(r))
		}
	}
	return sb.String()
}

func readFile(filename string, wg *sync.WaitGroup) {
	defer wg.Done()

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	wordCount := make(map[string]int) // Don't think we need this
	words := strings.Fields(string(data))

	for _, word := range words {
		cleanedWord := cleanWord(word)
		if cleanedWord != "" {
			wordCount[cleanedWord]++
		}
	}

	fmt.Println(filename, wordCount)
}

func main() {
	filenames := []string{"file1.txt", "file2.txt", "file3.txt"}

	var wg sync.WaitGroup

	for _, filename := range filenames {
		wg.Add(1)
		go readFile(filename, &wg)
	}
	wg.Wait()
}
