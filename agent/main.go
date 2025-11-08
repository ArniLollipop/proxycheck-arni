package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type ProxyVisitLogs struct {
	Id        string    `json:"id"`
	ProxyId   string    `json:"proxy_id"`
	Timestamp time.Time `json:"timestamp"`
	SourceIP  string    `json:"source_ip"`
	TargetIP  string    `json:"target_ip"`
	Domain    string    `json:"domain"`
}

type AgentState struct {
	LastTimestamp time.Time `json:"last_timestamp"`
}

const stateFileName = "agent_state.json"
const batchSize = 10000

func main() {
	logPath := flag.String("log-path", ".", "Path to the directory with log files")
	apiHost := flag.String("api-host", "http://localhost:8080", "API host URL")
	flag.Parse()
	log.Println("Agent started...")
	log.Printf("Log directory: %s", *logPath)
	log.Printf("API Host: %s", *apiHost)

	logFile, err := findLatestLogFile(*logPath)
	if err != nil {
		log.Fatalf("Error finding log file: %v", err)
	}
	log.Printf("Processing log file: %s", logFile)

	lastTimestamp, err := loadState()
	if err != nil {
		log.Printf("Could not load state, starting from scratch: %v", err)
	} else {
		log.Printf("Resuming from last timestamp: %s", lastTimestamp.Format(time.RFC3339))
	}

	logs, newLastTimestamp, err := parseLogFile(logFile, lastTimestamp)
	if err != nil {
		log.Fatalf("Error parsing log file: %v", err)
	}

	if len(logs) == 0 {
		log.Println("No new log entries to send.")
		return
	}

	log.Printf("Found %d new log entries to send.", len(logs))

	err = sendLogsInBatches(logs, *apiHost)
	if err != nil {
		log.Fatalf("Failed to send logs: %v", err)
	}

	err = saveState(newLastTimestamp)
	if err != nil {
		log.Fatalf("Failed to save state: %v", err)
	}

	log.Printf("Successfully sent %d logs. New state saved.", len(logs))
	log.Println("Agent finished.")
}

func findLatestLogFile(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var latestFile os.FileInfo
	var latestModTime time.Time

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		if latestFile == nil || info.ModTime().After(latestModTime) {
			latestFile = info
			latestModTime = info.ModTime()
		}
	}

	if latestFile == nil {
		return "", fmt.Errorf("no files found in directory %s", dir)
	}

	return filepath.Join(dir, latestFile.Name()), nil
}

func parseLogFile(filePath string, lastTimestamp time.Time) ([]ProxyVisitLogs, time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, lastTimestamp, err
	}
	defer file.Close()

	var logs []ProxyVisitLogs
	scanner := bufio.NewScanner(file)
	latestTimestamp := lastTimestamp

	for scanner.Scan() {
		line := scanner.Text()
		logEntry, err := ParseLog(line)
		if err != nil {
			continue
		}

		if logEntry.Timestamp.After(lastTimestamp) {
			logs = append(logs, ProxyVisitLogs{
				Id:        generateId(line),
				ProxyId:   logEntry.Username,
				Timestamp: logEntry.Timestamp,
				SourceIP:  logEntry.ClientIP,
				TargetIP:  logEntry.TargetIP,
				Domain:    logEntry.TargetHost,
			})
			if logEntry.Timestamp.After(latestTimestamp) {
				latestTimestamp = logEntry.Timestamp
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, lastTimestamp, err
	}

	return logs, latestTimestamp, nil
}

type LogEntry struct {
	Timestamp  time.Time
	Username   string
	ClientIP   string
	ClientPort string
	TargetHost string
	TargetIP   string
	TargetPort string
	Method     string
	Protocol   string
	StatusCode string
	BytesSent  int64
	BytesRecv  int64
	Raw        string
}

var (
	reSocks  = regexp.MustCompile(`^([\d.]+)\s+-\s+(\S+)\s+\[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2}\s+[+-]\d{4})\]\s+"CONNECT\s+([^"]+)"\s+(\d+)\s+(\d+)\s+(\d+)\s+SOCK5/([\d.]+):(\d+)$`)
	re3Proxy = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+\S+\s+(\d+)\s+(\S+)\s+([\d.]+):(\d+)\s+([\d.]+):(\d+)\s+[\d.]+:\d+\s+(\d+)\s+(\d+)\s+(\S+)\s+CONNECT\s+(\S+):(\d+)\s+HTTP/\d\.\d$`)
)

func ParseLog(line string) (*LogEntry, error) {
	if m := reSocks.FindStringSubmatch(line); m != nil {
		t, _ := time.Parse("02/Jan/2006:15:04:05 -0700", m[3])
		return &LogEntry{
			Timestamp:  t,
			Username:   m[2],
			ClientIP:   m[1],
			TargetHost: m[4],
			StatusCode: m[5],
			BytesSent:  parseInt(m[6]),
			BytesRecv:  parseInt(m[7]),
			Protocol:   "SOCKS5",
			TargetIP:   m[8],
			TargetPort: m[9],
			Method:     "CONNECT",
			Raw:        line,
		}, nil
	}

	if m := re3Proxy.FindStringSubmatch(line); m != nil {
		t, _ := time.Parse("2006-01-02 15:04:05", m[1])
		return &LogEntry{
			Timestamp:  t,
			Username:   m[3],
			ClientIP:   m[4],
			ClientPort: m[5],
			TargetIP:   m[6],
			TargetPort: m[7],
			StatusCode: m[2],
			BytesSent:  parseInt(m[8]),
			BytesRecv:  parseInt(m[9]),
			TargetHost: m[10],
			Method:     "CONNECT",
			Protocol:   "HTTP",
			Raw:        line,
		}, nil
	}

	return nil, fmt.Errorf("unknown log format")
}

func parseInt(s string) int64 {
	var v int64
	fmt.Sscan(s, &v)
	return v
}

func generateId(rawLine string) string {
	hash := sha256.Sum256([]byte(rawLine))
	return hex.EncodeToString(hash[:])
}

func loadState() (time.Time, error) {
	data, err := os.ReadFile(stateFileName)
	if err != nil {
		return time.Time{}, err
	}

	var state AgentState
	if err := json.Unmarshal(data, &state); err != nil {
		return time.Time{}, err
	}

	return state.LastTimestamp, nil
}

func saveState(timestamp time.Time) error {
	state := AgentState{
		LastTimestamp: timestamp,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFileName, data, 0644)
}

func sendLogsInBatches(logs []ProxyVisitLogs, apiHost string) error {
	for i := 0; i < len(logs); i += batchSize {
		end := i + batchSize
		if end > len(logs) {
			end = len(logs)
		}

		batch := logs[i:end]
		if err := sendBatch(batch, apiHost); err != nil {
			return err
		}
	}

	return nil
}

func sendBatch(batch []ProxyVisitLogs, apiHost string) error {
	data, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	resp, err := http.Post(apiHost+"/api/proxyVisits", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
