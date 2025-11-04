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
	"strings"
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

var (
	// Regex for format 1: 2025-10-29 10:00:01 PROXY.2315 ... 65.109.18.254:35366 85.198.79.24:443 ... CONNECT static-basket-01.wbbasket.ru:443 ...
	logFmt1 = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2})\s(PROXY\.\d+)\s+.*\s([\d\.]+):\d+\s([\d\.]+):\d+.*\sCONNECT\s([\w\.-]+):\d+`)
	// Regex for format 2: 94.130.134.120 - DiamondBlond [28/Oct/2025:00:00:04 +0300] "CONNECT adsmanager-graph.facebook.com:443" ... SOCK5/185.60.218.19:443
	logFmt2 = regexp.MustCompile(`^([\d\.]+)\s-\s(\w+)\s\[(.*?)\]\s"CONNECT\s([\w\.-]+):\d+"\s+.*\sSOCK5/([\d\.]+):\d+`)
)

func main() {
	logPath := flag.String("log-path", ".", "Path to the directory with log files")
	apiHost := flag.String("api-host", "http://localhost:8080", "API host URL")
	proxyId := flag.String("proxy-id", "", "Proxy ID to assign to log entries")
	flag.Parse()

	if *proxyId == "" {
		log.Fatal("proxy-id flag is required")
	}

	log.Println("Agent started...")
	log.Printf("Log directory: %s", *logPath)
	log.Printf("API Host: %s", *apiHost)
	log.Printf("Proxy ID: %s", *proxyId)

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

	logs, newLastTimestamp, err := parseLogFile(logFile, lastTimestamp, *proxyId)
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

func parseLogFile(filePath string, lastTimestamp time.Time, proxyId string) ([]ProxyVisitLogs, time.Time, error) {
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
		logEntry, err := parseLogLine(line, proxyId)
		if err != nil {
			// log.Printf("Skipping unparsable line: %s", line)
			continue
		}

		if logEntry.Timestamp.After(lastTimestamp) {
			logs = append(logs, *logEntry)
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

func parseLogLine(line string, proxyId string) (*ProxyVisitLogs, error) {
	// Try format 1 first
	matches1 := logFmt1.FindStringSubmatch(line)
	if len(matches1) >= 6 {
		timestamp, err := time.Parse("2006-01-02 15:04:05", matches1[1])
		if err != nil {
			return nil, err
		}
		return &ProxyVisitLogs{
			Id:        generateId(line),
			ProxyId:   proxyId,
			Timestamp: timestamp,
			SourceIP:  matches1[3],
			TargetIP:  matches1[4],
			Domain:    matches1[5],
		}, nil
	}

	// Try format 2 second
	matches2 := logFmt2.FindStringSubmatch(line)
	if len(matches2) >= 6 {
		// Parse timestamp from format like "28/Oct/2025:00:00:04 +0300"
		timestampStr := matches2[3]
		// Remove timezone part for parsing
		timestampStr = strings.Split(timestampStr, " ")[0]
		timestamp, err := time.Parse("02/Jan/2006:15:04:05", timestampStr)
		if err != nil {
			return nil, err
		}
		return &ProxyVisitLogs{
			Id:        generateId(line),
			ProxyId:   proxyId,
			Timestamp: timestamp,
			SourceIP:  matches2[1],
			TargetIP:  matches2[5],
			Domain:    matches2[4],
		}, nil
	}

	return nil, fmt.Errorf("unparsable log line")
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
