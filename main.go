package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode"
)

type wordFreq struct {
	word  string
	count int
}

func cleanWord(word string) string {
	var sb strings.Builder
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			sb.WriteRune(unicode.ToLower(r))
		}
	}
	return sb.String()
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
		cleanedWord := cleanWord(word)
		if cleanedWord != "" {
			wordCount[cleanedWord]++
			wordchan <- cleanedWord
		}
	}

	// fmt.Println(filename, wordCount)
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
	fmt.Println(sortedWords[:10])
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
