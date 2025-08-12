package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const fileArg = 0

func main() {
	fieldsFlag := flag.String("f", "", "select number fields(columns)")
	delimiterFlag := flag.String("d", "\t", "select field delimiter")
	separatedFlag := flag.Bool("s", false, "only lines with delimiter")
	flag.Parse()

	fields, err := parseFields(*fieldsFlag)
	if err != nil {
		fmt.Printf("Error parsing fields: %v (-f)\n", err)
		os.Exit(1)
	}

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

	if err := processInput(input, output, fields, *delimiterFlag, *separatedFlag); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFields(fieldsStr string) ([]int, error) {
	if fieldsStr == "" {
		return nil, fmt.Errorf("fields flag is required")
	}

	var fields []int
	field := strings.Split(fieldsStr, ",")
	for _, part := range field {
		if strings.Contains(part, "-") {
			fieldPart := strings.Split(part, "-")
			if len(fieldPart) != 2 {
				return nil, fmt.Errorf("invalid field part: %s", part)
			}
			firstNum, err := strconv.Atoi(fieldPart[0])
			if err != nil || firstNum < 1 {
				return nil, fmt.Errorf("invalid start num for cut: %s", fieldPart[0])
			}

			secondNum, err := strconv.Atoi(fieldPart[1])
			if err != nil || secondNum < 1 {
				return nil, fmt.Errorf("invalid end num for cut: %s", fieldPart[1])
			}

			if secondNum < firstNum {
				return nil, fmt.Errorf("end num less start num: %s", part)
			}

			for i := firstNum; i <= secondNum; i++ {
				fields = append(fields, i)
			}
		} else {
			fieldWithOutRange, err := strconv.Atoi(part)
			if err != nil || fieldWithOutRange < 1 {
				return nil, fmt.Errorf("invalid field with out range: %s", part)
			}
			fields = append(fields, fieldWithOutRange)
		}
	}

	return fields, nil
}

func processInput(r io.Reader, w io.Writer, fields []int, delimiter string, separated bool) error {
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)
	defer func() {
		if err := writer.Flush(); err != nil {
			fmt.Printf("Error by writer flush: %v\n", err)
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()
		var parts []string
		if delimiter == "\t" {
			parts = strings.Fields(line)
		} else {
			parts = strings.Split(line, delimiter)
		}

		if separated && len(parts) <= 1 {
			continue
		}

		var output []string
		for _, field := range fields {
			if field > 0 && field <= len(parts) {
				output = append(output, parts[field-1])
			}
		}

		outDelimiter := delimiter
		if delimiter == " " || delimiter == "\t" {
			outDelimiter = " "
		}

		if len(output) > 0 {
			result := strings.Join(output, outDelimiter)
			if _, err := fmt.Fprintln(writer, result); err != nil {
				return err
			}
			if err := writer.Flush(); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}
