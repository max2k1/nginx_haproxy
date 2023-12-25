package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const bufferSize int = 10 * 1024 * 1024
const line string = "01|02|03|04|05|06|07|08|09|0A|0B|0C|0D|0E|0F|10|11|12|13|14|15|16|17|18|19|1A|1B\n" +
	"--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--\n"

func logAccess(logger *slog.Logger, r *http.Request, msg string, startTime time.Time) {
	if logger == nil {
		return
	}
	logger.WithGroup("request").Info(msg,
		"src", r.RemoteAddr,
		"host", r.Host,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"time", startTime.Format(time.RFC3339Nano),
		"elapsed_ns", time.Now().Sub(startTime))
}

func buildLogger(filename string) *slog.Logger {
	if filename == "" {
		return nil
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	return slog.New(slog.NewJSONHandler(f, nil))
}

func main() {
	var filename string
	var port int
	flag.StringVar(&filename, "accesslog", "", "Filename to write access log into")
	flag.IntVar(&port, "port", 9990, "Port to listen to")
	flag.Parse()
	accessLogger := buildLogger(filename)

	var sb = strings.Builder{}
	sb.Grow(bufferSize)
	for idx := 0; idx < bufferSize; idx++ {
		sb.WriteByte(line[idx%len(line)])
	}
	var buffer string = sb.String()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logAccess(accessLogger, r, "rootHandler", time.Now())
		http.Error(w, "Unknown URI: "+r.URL.Path, http.StatusNotFound)
	})

	// Health-check handler
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		defer logAccess(accessLogger, r, "healthHandler", time.Now())
		_, _ = fmt.Fprintln(w, "alive")
	})

	// Should be called as "/generate/N", where N is number from 1 to 10M
	const handleName string = "/generate/"
	http.HandleFunc(handleName, func(w http.ResponseWriter, r *http.Request) {
		defer logAccess(accessLogger, r, handleName, time.Now())
		numStr := r.URL.Path[len(handleName):]
		num, err := strconv.Atoi(numStr)
		if err != nil || num < 1 || num > len(buffer) {
			http.Error(w, "Invalid argument was given", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprintln(w, buffer[:num-1])
	})

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
