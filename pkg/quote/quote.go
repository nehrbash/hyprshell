package quote

import (
	"bufio"
	"embed"
	"encoding/csv"
	"io"
	"math/rand"
	"strings"
	"time"
)

//go:embed dist/quotes.csv
var quotesFile embed.FS

func openQuotesFile() (io.ReadSeekCloser, error) {
	file, err := quotesFile.Open("dist/quotes.csv")
	if err != nil {
		return nil, err
	}
	return file.(io.ReadSeekCloser), nil
}

func countLines(file io.Reader) (int, error) {
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	return lineCount, scanner.Err()
}

func getRandomLine(file io.ReadSeeker, lineCount int) (string, error) {
	rand.Seed(time.Now().UnixNano())
	randomLineIndex := rand.Intn(lineCount)

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	lineIndex := 0
	for scanner.Scan() {
		if lineIndex == randomLineIndex {
			break
		}
		lineIndex++
	}

	return scanner.Text(), scanner.Err()
}

func parseLine(line string) (string, string, error) {
	reader := csv.NewReader(strings.NewReader(line))
	record, err := reader.Read()
	if err != nil {
		return "", "", err
	}
	return record[0], record[1], nil
}

func GetRandomQuote() (string, string, error) {
	file, err := openQuotesFile()
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	lineCount, err := countLines(file)
	if err != nil {
		return "", "", err
	}

	randomLine, err := getRandomLine(file, lineCount)
	if err != nil {
		return "", "", err
	}

	author, quote, err := parseLine(randomLine)
	if err != nil {
		return "", "", err
	}

	return quote, author, nil
}
