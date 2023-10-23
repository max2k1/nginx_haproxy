package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const bufferSize int = 10 * 1024 * 1024
const line string = "01|02|03|04|05|06|07|08|09|0A|0B|0C|0D|0E|0F|10|11|12|13|14|15|16|17|18|19|1A|1B\n" +
	"--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--\n"

func main() {
	var sb = strings.Builder{}
	sb.Grow(bufferSize)
	for idx := 0; idx < bufferSize; idx++ {
		sb.WriteByte(line[idx%len(line)])
	}
	var buffer string = sb.String()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Unknown URI: "+r.URL.Path+"\n")
	})

	// Health-check handler
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "alive")
	})

	// Should be called as "/generate/N", where N is number from 1 to 10M
	const handleName string = "/generate/"
	http.HandleFunc(handleName, func(w http.ResponseWriter, r *http.Request) {
		numStr := r.URL.Path[len(handleName):]
		num, err := strconv.Atoi(numStr)
		if err != nil || num < 1 || num > len(buffer) {
			http.Error(w, "Invalid argument was given", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, buffer[:num-1])
	})

	http.ListenAndServe(":9990", nil)
}
