package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const fileArg = 0

var (
	columnFlag           = flag.Int("k", 0, "select column to sort")
	delimiterFlag        = flag.String("d", "\t", "select field delimiter")
	numericFlag          = flag.Bool("n", false, "sort numerically")
	reverseFlag          = flag.Bool("r", false, "sort in reverse order")
	uniqueFlag           = flag.Bool("u", false, "output only unique lines")
	monthSortFlag        = flag.Bool("M", false, "sort by month names")
	ignoreBlanksFlag     = flag.Bool("b", false, "ignore trailing blanks")
	checkSortedFlag      = flag.Bool("c", false, "check if input is sorted")
	sortedWithSuffixFlag = flag.Bool("h", false, "compare human-readable numbers (K=kilobytes, M=megabytes...)")
)

func main() {
	flag.Parse()

	var input io.Reader = os.Stdin
	var output io.Writer = os.Stdout
	if flag.NArg() > 0 {
		file, err := os.Open(flag.Arg(fileArg))
		if err != nil {
			fmt.Printf("Error open file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Printf("Error close file: %v\n", err)
			}
		}()
		input = file
	}

	if *columnFlag > 0 {
		processingForColumn(input, output, *columnFlag)
	} else {
		processingNotColumn(input, output)
	}
}

func processingForColumn(r io.Reader, w io.Writer, column int) {
	var lines []string
	columnData := make(map[string]string)
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)
	defer func() {
		if err := writer.Flush(); err != nil {
			fmt.Printf("Error by writer flush: %v\n", err)
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()
		if *ignoreBlanksFlag {
			line = strings.TrimRight(line, " \t")
		}

		columns := strings.Split(line, *delimiterFlag)
		if len(columns) < column {
			continue
		}

		key := columns[column-1]
		if *uniqueFlag {
			if _, exist := columnData[key]; exist {
				continue
			}
		}
		columnData[key] = line
		lines = append(lines, key)
	}

	if *checkSortedFlag {
		if isSorted(lines) {
			_, _ = fmt.Fprintln(writer, "File is sorted")
		} else {
			_, _ = fmt.Fprintln(writer, "File is not sorted")
		}
		return
	}

	sortedKey(lines)

	for _, key := range lines {
		_, _ = fmt.Fprintln(writer, columnData[key])
	}
}

func processingNotColumn(r io.Reader, w io.Writer) {
	var lines []string
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)
	defer func() {
		if err := writer.Flush(); err != nil {
			fmt.Printf("Error by writer flush: %v\n", err)
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()
		if *ignoreBlanksFlag {
			line = strings.TrimRight(line, " \t")
		}

		if *uniqueFlag {
			if contains(lines, line) {
				continue
			}
		}
		lines = append(lines, line)
	}

	if *checkSortedFlag {
		if isSorted(lines) {
			_, _ = fmt.Fprintln(writer, "File is sorted")
		} else {
			_, _ = fmt.Fprintln(writer, "File is not sorted")
		}
		return
	}

	sortedKey(lines)

	for _, line := range lines {
		_, _ = fmt.Fprintln(writer, line)
	}
}

func isSorted(lines []string) bool {
	for i := 1; i < len(lines); i++ {
		compareResult := compareValues(lines[i-1], lines[i])
		if *reverseFlag {
			if compareResult < 0 {
				return false
			}
		} else {
			if compareResult > 0 {
				return false
			}
		}
	}
	return true
}

func sortedKey(lines []string) {
	sort.Slice(lines, func(i, j int) bool {
		compareResult := compareValues(lines[i], lines[j])
		if *reverseFlag {
			return compareResult > 0
		}
		return compareResult < 0
	})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func compareValues(a, b string) int {
	if *monthSortFlag {
		monthA := parseMonth(a)
		monthB := parseMonth(b)
		if monthA != 0 && monthB != 0 {
			return int(monthA - monthB)
		}
	}

	if *sortedWithSuffixFlag {
		numA := parseHumanNumber(a)
		numB := parseHumanNumber(b)
		if numA != 0 || numB != 0 {
			if numA > numB {
				return 1
			} else if numA < numB {
				return -1
			}
			return 0
		}
	}

	if *numericFlag {
		numA, errA := strconv.ParseFloat(a, 64)
		numB, errB := strconv.ParseFloat(b, 64)
		if errA == nil && errB == nil {
			if numA > numB {
				return 1
			} else if numA < numB {
				return -1
			}
			return 0
		}
	}

	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

func parseMonth(str string) time.Month {
	months := []string{
		"JAN", "FEB", "MAR", "APR", "MAY", "JUN",
		"JUL", "AUG", "SEP", "OCT", "NOV", "DEC",
	}
	str = strings.ToUpper(strings.TrimSpace(str))
	for i, month := range months {
		if strings.HasPrefix(str, month) {
			return time.Month(i + 1)
		}
	}
	return 0
}

func parseHumanNumber(s string) int {
	s = strings.ToUpper(strings.TrimSpace(s))
	multiplier := 1

	if strings.HasSuffix(s, "K") {
		multiplier = 1_000
		s = strings.TrimSuffix(s, "K")
	} else if strings.HasSuffix(s, "M") {
		multiplier = 1000_000
		s = strings.TrimSuffix(s, "M")
	} else if strings.HasSuffix(s, "G") {
		multiplier = 1000_000_000
		s = strings.TrimSuffix(s, "G")
	} else if strings.HasSuffix(s, "T") {
		multiplier = 1000_000_000_000
		s = strings.TrimSuffix(s, "T")
	}

	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int(num) * multiplier
}
