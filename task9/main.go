package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrInvalidString  = errors.New("invalid string: starts with digit")
	ErrInStrWithDigit = errors.New("invalid number format")
)

func main() {
	testData := []string{
		"a4bc2d5e",
		"abcd",
		"45",
		"",
		"qwe\\4\\5",
		"qwe\\45",
		"a0b0c2x0",
		"abc\\\\5",
		"Г1м0а10ф3\\4",
	}

	for ind, testOne := range testData {
		if ind == 0 {
		} else {
			fmt.Println()
		}
		result, err := unpackStr(testOne)
		if err != nil {
			fmt.Printf("Input: %s\nError: %s\n", testOne, err)
		} else {
			fmt.Printf("Input: %s\nOutput: %s\n", testOne, result)
		}
	}
}

func calculateUnpackStrLen(str string) (int, error) {
	var length int
	var escape bool

	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '\\' && !escape {
			escape = true
			continue
		}

		if escape {
			length++
			escape = false
			continue
		}

		if unicode.IsDigit(r) {
			if i == 0 {
				return 0, ErrInvalidString
			}

			numStr := string(r)
			j := i + 1
			for j < len(runes) && unicode.IsDigit(runes[j]) {
				numStr = numStr + string(runes[j])
				j++
			}

			repeat, err := strconv.Atoi(numStr)
			if err != nil {
				return 0, ErrInStrWithDigit
			}

			if repeat == 0 {
				length -= 1
			} else {
				length += repeat - 1
			}
			i = j - 1
			continue
		}
		length++
	}
	return length, nil
}

func unpackStr(str string) (string, error) {
	var strBuilder strings.Builder
	var escape bool

	/*lenUnpackStr, err := calculateUnpackStrLen(str) // Использовать если на выходе предлолагаются большие строки (a100b500)
	  if err != nil {
	    return "", err
	  }
	  strBuilder.Grow(lenUnpackStr)*/

	runes := []rune(str)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '\\' && !escape {
			escape = true
			continue
		}

		if escape {
			strBuilder.WriteRune(r)
			escape = false
			continue
		}

		if unicode.IsDigit(r) {
			if i == 0 {
				return "", ErrInvalidString
			}

			numStr := string(r)
			j := i + 1
			for j < len(runes) && unicode.IsDigit(runes[j]) {
				numStr = numStr + string(runes[j])
				j++
			}

			repeat, err := strconv.Atoi(numStr)
			if err != nil {
				return "", ErrInStrWithDigit
			}

			if repeat == 0 {
				temp := strBuilder.String()
				if len(temp) > 0 {
					_, size := utf8.DecodeLastRuneInString(temp)
					strBuilder.Reset()
					strBuilder.WriteString(temp[:len(temp)-size])
				}
			} else {
				lastRune := runes[i-1]
				strBuilder.WriteString(strings.Repeat(string(lastRune), repeat-1))
			}
			i = j - 1
			continue
		}
		strBuilder.WriteRune(r)
	}
	return strBuilder.String(), nil
}
