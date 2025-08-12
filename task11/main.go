package main

import (
	"fmt"
	"slices"
	"strings"
)

var words = []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}

func main() {
	fmt.Println(words)

	anagrams := findAnagrams(words)

	for key, anagram := range anagrams {
		fmt.Println(key+":", anagram)
	}
}

func findAnagrams(words []string) map[string][]string {
	anagramGroups := make(map[string][]string)
	signatureM := make(map[string]string)

	for _, word := range words {
		lowWord := strings.ToLower(word)
		signature := findSignature(lowWord)

		if firstWord, exist := signatureM[signature]; exist {
			anagramGroups[firstWord] = append(anagramGroups[firstWord], lowWord)
		} else {
			signatureM[signature] = lowWord
			anagramGroups[lowWord] = []string{lowWord}
		}
	}

	sortResMapAndCleaning(anagramGroups)

	return anagramGroups
}

func findSignature(lowWord string) string {
	stringToRune := []rune(lowWord)
	slices.Sort(stringToRune)
	return string(stringToRune)
}

func sortResMapAndCleaning(anagramGroups map[string][]string) {
	for key, value := range anagramGroups {
		if len(value) == 1 {
			delete(anagramGroups, key)
		} else {
			slices.Sort(value)
		}
	}
}
