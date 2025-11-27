package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// NotificationService handles sending notifications
type NotificationService struct {
	TelegramEnabled bool
	TelegramToken   string
	TelegramChatID  string
	client          *http.Client
}

// NewNotificationService creates a new notification service
func NewNotificationService(enabled bool, token, chatID string) *NotificationService {
	return &NotificationService{
		TelegramEnabled: enabled,
		TelegramToken:   token,
		TelegramChatID:  chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TelegramMessage represents a Telegram API message
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// SendTelegram sends a message to Telegram
func (n *NotificationService) SendTelegram(message string) error {
	if !n.TelegramEnabled || n.TelegramToken == "" || n.TelegramChatID == "" {
		return nil // Notifications disabled
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.TelegramToken)

	msg := TelegramMessage{
		ChatID:    n.TelegramChatID,
		Text:      message,
		ParseMode: "HTML",
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	log.Printf("Telegram notification sent successfully")
	return nil
}

// NotifyProxyDown sends notification when proxy goes down
func (n *NotificationService) NotifyProxyDown(proxy *Proxy, errorMsg string) {
	message := fmt.Sprintf(
		"🔴 <b>Proxy Down</b>\n\n"+
			"<b>Name:</b> %s\n"+
			"<b>IP:</b> %s:%s\n"+
			"<b>Username:</b> %s\n"+
			"<b>Failures:</b> %d\n"+
			"<b>Error:</b> %s\n"+
			"<b>Time:</b> %s",
		escapeHTML(proxy.Name),
		proxy.Ip,
		proxy.Port,
		escapeHTML(proxy.Username),
		proxy.Failures,
		escapeHTML(errorMsg),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if err := n.SendTelegram(message); err != nil {
		log.Printf("Failed to send telegram notification: %v", err)
	}
}

// NotifyProxyRecovered sends notification when proxy recovers
func (n *NotificationService) NotifyProxyRecovered(proxy *Proxy) {
	message := fmt.Sprintf(
		"🟢 <b>Proxy Recovered</b>\n\n"+
			"<b>Name:</b> %s\n"+
			"<b>IP:</b> %s:%s\n"+
			"<b>Username:</b> %s\n"+
			"<b>Latency:</b> %d ms\n"+
			"<b>Time:</b> %s",
		escapeHTML(proxy.Name),
		proxy.Ip,
		proxy.Port,
		escapeHTML(proxy.Username),
		proxy.LastLatency,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if err := n.SendTelegram(message); err != nil {
		log.Printf("Failed to send telegram notification: %v", err)
	}
}

// NotifyIPChanged sends notification when proxy IP changes
func (n *NotificationService) NotifyIPChanged(proxy *Proxy, oldIP, newIP string) {
	message := fmt.Sprintf(
		"🔄 <b>IP Changed</b>\n\n"+
			"<b>Name:</b> %s\n"+
			"<b>Proxy:</b> %s:%s\n"+
			"<b>Username:</b> %s\n"+
			"<b>Old IP:</b> %s\n"+
			"<b>New IP:</b> %s\n"+
			"<b>Country:</b> %s\n"+
			"<b>Operator:</b> %s\n"+
			"<b>Time:</b> %s",
		escapeHTML(proxy.Name),
		proxy.Ip,
		proxy.Port,
		escapeHTML(proxy.Username),
		oldIP,
		newIP,
		escapeHTML(proxy.RealCountry),
		escapeHTML(proxy.Operator),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if err := n.SendTelegram(message); err != nil {
		log.Printf("Failed to send telegram notification: %v", err)
	}
}

// NotifyIPStuck sends notification when IP is stuck (>24 hours)
func (n *NotificationService) NotifyIPStuck(proxy *Proxy, stuckIP string, hours int) {
	message := fmt.Sprintf(
		"⚠️ <b>IP Stuck</b>\n\n"+
			"<b>Name:</b> %s\n"+
			"<b>Proxy:</b> %s:%s\n"+
			"<b>Username:</b> %s\n"+
			"<b>Stuck IP:</b> %s\n"+
			"<b>Duration:</b> %d hours\n"+
			"<b>Country:</b> %s\n"+
			"<b>Operator:</b> %s\n"+
			"<b>Time:</b> %s",
		escapeHTML(proxy.Name),
		proxy.Ip,
		proxy.Port,
		escapeHTML(proxy.Username),
		stuckIP,
		hours,
		escapeHTML(proxy.RealCountry),
		escapeHTML(proxy.Operator),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if err := n.SendTelegram(message); err != nil {
		log.Printf("Failed to send telegram notification: %v", err)
	}
}

// NotifyLowSpeed sends notification when proxy speed is below threshold
func (n *NotificationService) NotifyLowSpeed(proxy *Proxy, threshold int) {
	message := fmt.Sprintf(
		"🐌 <b>Low Speed Detected</b>\n\n"+
			"<b>Name:</b> %s\n"+
			"<b>Proxy:</b> %s:%s\n"+
			"<b>Username:</b> %s\n"+
			"<b>Download:</b> %d Mbps\n"+
			"<b>Upload:</b> %d Mbps\n"+
			"<b>Threshold:</b> %d Mbps\n"+
			"<b>Time:</b> %s",
		escapeHTML(proxy.Name),
		proxy.Ip,
		proxy.Port,
		escapeHTML(proxy.Username),
		proxy.Speed,
		proxy.Upload,
		threshold,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if err := n.SendTelegram(message); err != nil {
		log.Printf("Failed to send telegram notification: %v", err)
	}
}

// NotifyDailySummary sends a daily summary of proxy status
func (n *NotificationService) NotifyDailySummary(totalProxies, aliveProxies, deadProxies int, avgSpeed float64) {
	message := fmt.Sprintf(
		"📊 <b>Daily Proxy Summary</b>\n\n"+
			"<b>Total Proxies:</b> %d\n"+
			"<b>Alive:</b> %d (%.1f%%)\n"+
			"<b>Dead:</b> %d (%.1f%%)\n"+
			"<b>Avg Speed:</b> %.1f Mbps\n"+
			"<b>Date:</b> %s",
		totalProxies,
		aliveProxies,
		float64(aliveProxies)/float64(totalProxies)*100,
		deadProxies,
		float64(deadProxies)/float64(totalProxies)*100,
		avgSpeed,
		time.Now().Format("2006-01-02"),
	)

	if err := n.SendTelegram(message); err != nil {
		log.Printf("Failed to send telegram notification: %v", err)
	}
}

// escapeHTML escapes HTML special characters for Telegram
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
