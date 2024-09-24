package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type wordFreq struct {
	word  string
	count int
}

var wordRegex = regexp.MustCompile(`[a-zA-Z0-9]+`)

func cleanWord(word string) []string {
	matches := wordRegex.FindAllString(strings.ToLower(word), -1)
	return matches
}

func readFile(filename string, wg *sync.WaitGroup, wordchan chan string) {
	defer wg.Done()

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	wordCount := make(map[string]int) // Don't think we need this
	words := strings.Fields(string(data))

	for _, word := range words {
		cleanedWords := cleanWord(word)
		for _, cleanedWord := range cleanedWords {
			if cleanedWord != "" {
				wordCount[cleanedWord]++
				wordchan <- cleanedWord
			}
		}
	}
	// fmt.Println(filename,wordCount)
}

func countWords(totalWordCount map[string]int, wordchan chan string, done chan bool) {
	for word := range wordchan {
		totalWordCount[word]++
	}
	done <- true
}

func main() {
	filenames := []string{"file1.txt", "file2.txt", "file3.txt"}
	var wg sync.WaitGroup

	wordchan := make(chan string)
	totalWordCount := make(map[string]int)

	for _, filename := range filenames {
		wg.Add(1)
		go readFile(filename, &wg, wordchan)
	}

	go func() {
		wg.Wait()
		close(wordchan)
	}()

	done := make(chan bool)
	go countWords(totalWordCount, wordchan, done)
	<-done
	close(done)
	fmt.Println(totalWordCount)

	fmt.Println("####################")
	sortedWords := sortByFrequency(totalWordCount)
	for i := 0; i < 10 && i < len(sortedWords); i++ {
		fmt.Printf("%s: %d\n", sortedWords[i].word, sortedWords[i].count)
	}
}

func sortByFrequency(wordCount map[string]int) []wordFreq {
	var wordList []wordFreq
	for word, count := range wordCount {
		wordList = append(wordList, wordFreq{word, count})
	}
	sort.Slice(wordList, func(i, j int) bool {
		return wordList[i].count > wordList[j].count
	})
	return wordList
}
