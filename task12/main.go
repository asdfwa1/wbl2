package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strings"
)

const (
	templateArg = iota
	fileArg
)

var (
	afterFlag      = flag.Int("A", 0, "print N lines after match")
	beforeFlag     = flag.Int("B", 0, "print N lines before match")
	contextFlag    = flag.Int("C", 0, "print N lines around match")
	countFlag      = flag.Bool("c", false, "print only count of matching lines")
	ignoreCaseFlag = flag.Bool("i", false, "ignore register")
	invertFlag     = flag.Bool("v", false, "invert the filter")
	fixedFlag      = flag.Bool("F", false, "interpret pattern as fixed string")
	lineNumFlag    = flag.Bool("n", false, "print line number before each found line")
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Printf("Error: template is required")
		os.Exit(1)
	}

	template := flag.Arg(templateArg)
	var input io.Reader = os.Stdin

	if flag.NArg() > 1 {
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

	if *contextFlag > 0 && (*afterFlag > 0 || *beforeFlag > 0) {
		fmt.Printf("cannot use flag -C with flag -A or flag -B")
		os.Exit(1)
	}

	if *contextFlag > 0 {
		*afterFlag = *contextFlag
		*beforeFlag = *contextFlag
	}

	re, err := compileTemplate(template)
	if err != nil {
		fmt.Printf("Error: compiling regexp %v\n", err)
		os.Exit(1)
	}

	if *countFlag {
		countMatch(input, re, template)
	} else {
		printMatch(input, re, template)
	}
}

func compileTemplate(template string) (*regexp.Regexp, error) {
	if *fixedFlag {
		return nil, nil
	}

	if *ignoreCaseFlag {
		template = "(?i)" + template
	}

	return regexp.Compile(template)
}

func countMatch(input io.Reader, regexp *regexp.Regexp, template string) {
	count := 0
	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		line := scanner.Text()
		var match bool
		if *fixedFlag {
			if *ignoreCaseFlag {
				match = strings.EqualFold(line, template)
			} else {
				match = line == template
			}
		} else {
			match = regexp.MatchString(line)
		}

		if match && !(*invertFlag) {
			count++
		}
	}

	fmt.Print(count)
}

func printMatch(input io.Reader, regexp *regexp.Regexp, template string) {
	scanner := bufio.NewScanner(input)
	var lines []string
	var matches []int
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)

		var matched bool
		if *fixedFlag {
			if *ignoreCaseFlag {
				matched = strings.EqualFold(line, template)
			} else {
				matched = line == template
			}
		} else {
			matched = regexp.MatchString(line)
		}

		if matched != *invertFlag {
			matches = append(matches, lineNum)
		}
		lineNum++
	}

	printedResult := make(map[int]bool)
	for _, match := range matches {
		start := int(math.Max(float64(0), float64(match)-float64(*beforeFlag)))
		for i := start; i < match; i++ {
			if !printedResult[i] {
				printStr(lines[i], i+1, false)
				printedResult[i] = true
			}
		}

		if !printedResult[match] {
			printStr(lines[match], match+1, true)
			printedResult[match] = true
		}

		end := int(math.Min(float64(len(lines)-1), float64(match)+float64(*afterFlag)))
		for j := match + 1; j <= end; j++ {
			if !printedResult[j] {
				printStr(lines[j], j+1, false)
				printedResult[j] = true
			}
		}
	}
}

func printStr(str string, number int, isMatch bool) {
	if *lineNumFlag {
		prefix := fmt.Sprintf("%d:", number)
		if isMatch {
			fmt.Printf("[Совпадение в строке] %s %s\n", prefix, str)
		} else {
			fmt.Printf("%s %s\n", prefix, str)
		}
	} else {
		fmt.Println(str)
	}
}
