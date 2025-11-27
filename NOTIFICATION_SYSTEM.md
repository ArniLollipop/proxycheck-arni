# Telegram Notification System & Failure History

## Overview
This document describes the implementation of the Telegram notification system and failure history tracking that was added to the proxy checker application.

## New Files Created

### 1. `code/notifications.go`
Telegram notification service that sends alerts for various proxy events.

**Key Features:**
- `NotificationService` struct handles all notifications
- HTTP client with 10-second timeout
- HTML-formatted messages for better readability
- Automatic HTML escaping for security

**Notification Types:**
- 🔴 **Proxy Down** - When proxy fails after 3 attempts
- 🟢 **Proxy Recovered** - When dead proxy comes back online
- 🔄 **IP Changed** - When proxy's real IP address changes
- ⚠️ **IP Stuck** - When IP hasn't changed for >24 hours
- 🐌 **Low Speed** - When speed drops below threshold
- 📊 **Daily Summary** - Daily statistics report

### 2. `code/failure_log.go`
Tracks all proxy failures with filtering and statistics.

**ProxyFailureLog Model:**
```go
type ProxyFailureLog struct {
    ID        string    // UUID
    ProxyID   string    // Proxy UUID reference
    Timestamp time.Time // When failure occurred
    ErrorType string    // "ping_failed", "speed_check_failed", "ip_check_failed"
    ErrorMsg  string    // Error details
    Latency   int       // Last known latency before failure
}
```

**Features:**
- Filtering by proxy, error type, date range
- Pagination support
- Statistics calculation (failure rate, counts by type)
- Sortable by any field

### 3. `code/scheduler_with_notifications.go`
Enhanced scheduler functions with notification support.

**Functions:**
- `IPCheckIteratorWithNotifications()` - IP check with notifications
- `checkSingleProxyIPWithNotifications()` - Single proxy IP check with alerts
- `HealthCheckIteratorWithNotifications()` - Speed check with notifications
- `checkSingleProxyHealthWithNotifications()` - Single proxy speed check with alerts

**Behavior:**
- Logs all failures to database
- Sends notifications based on settings
- Tracks proxy state changes (down → up, up → down)
- Detects IP changes and stuck IPs

## Modified Files

### 4. `code/settings.go`
Added notification configuration fields:

```go
// Notification settings
TelegramEnabled      bool   // Master toggle for Telegram
TelegramToken        string // Bot token from @BotFather
TelegramChatID       string // Chat/Channel ID
NotifyOnDown         bool   // Alert when proxy goes down
NotifyOnRecovery     bool   // Alert when proxy recovers
NotifyOnIPChange     bool   // Alert on IP address change
NotifyOnIPStuck      bool   // Alert if IP stuck >24h
NotifyOnLowSpeed     bool   // Alert on low speed
LowSpeedThreshold    int    // Mbps threshold (default: 10)
NotifyDailySummary   bool   // Send daily report
DailySummaryTime     string // Time for report (HH:MM format)
```

**Default Values:**
- NotifyOnDown: true
- NotifyOnRecovery: true
- NotifyOnIPChange: false (can be noisy)
- NotifyOnIPStuck: true
- NotifyOnLowSpeed: false
- LowSpeedThreshold: 10 Mbps
- NotifyDailySummary: false
- DailySummaryTime: "09:00"

### 5. `code/main.go`
Integration changes:

**Line 42:** Added ProxyFailureLog to database migration
```go
db.AutoMigrate(&Proxy{}, &Settings{}, &ProxySpeedLog{}, &ProxyIPLog{},
    &ProxyVisitLogs{}, &ProxyFailureLog{})
```

**Lines 49-55:** Initialize NotificationService
```go
notificationService := NewNotificationService(
    settings.TelegramEnabled,
    settings.TelegramToken,
    settings.TelegramChatID,
)
```

**Lines 75-76:** Pass notifier to schedulers
```go
go StartIPCheckScheduler(&wg, quit, db, settings, geoIP, notificationService)
go StartHealthCheckScheduler(&wg, quit, db, settings, notificationService)
```

**Lines 129-130:** Updated import handler
```go
go RunSingleIPCheck(db, settings, geoIP, notificationService)
go RunSingleHealthCheck(db, settings, notificationService)
```

**Lines 136-138:** Added new API endpoints
```go
router.GET("/api/failureLogs", h.GetFailureLogs)
router.GET("/api/failureStats/:id", h.GetFailureStats)
router.POST("/api/testNotification", h.TestNotification)
```

### 6. `code/scheduler.go`
Updated function signatures to accept NotificationService:

**Line 20:** `RunSingleIPCheck()` - Added notifier parameter
**Line 29:** `RunSingleHealthCheck()` - Added notifier parameter
**Line 132:** `StartIPCheckScheduler()` - Added notifier parameter
**Line 174:** Calls notification-enabled iterator
**Line 253:** `StartHealthCheckScheduler()` - Added notifier parameter
**Line 295:** Calls notification-enabled health check

### 7. `code/handler.go`
Added three new handler methods:

**Lines 648-699:** `GetFailureLogs()` - List failure logs with filtering
- Query parameters: proxy_id, error_type, start_date, end_date, page, page_size, sort_field
- Returns paginated failure log list

**Lines 701-717:** `GetFailureStats()` - Get failure statistics for a proxy
- URL parameter: :id (proxy UUID)
- Returns failure counts by type, failure rate, last failure time

**Lines 719-752:** `TestNotification()` - Test Telegram integration
- POST body: `{"message": "test message"}`
- Sends test notification to configured Telegram chat
- Returns success/error response

## API Endpoints

### Failure Logs
```
GET /api/failureLogs
Query Parameters:
  - proxy_id (optional): Filter by proxy UUID
  - error_type (optional): Filter by error type (ping_failed, speed_check_failed, ip_check_failed)
  - start_date (optional): YYYY-MM-DD format
  - end_date (optional): YYYY-MM-DD format
  - page (default: 1): Page number
  - page_size (default: 50): Items per page
  - sort_field (optional): Field to sort by (default: timestamp desc)

Response:
{
  "data": [
    {
      "id": "uuid",
      "proxy_id": "uuid",
      "timestamp": "2025-11-27T10:30:00Z",
      "error_type": "ping_failed",
      "error_msg": "context deadline exceeded",
      "latency": 150
    }
  ],
  "total": 42
}
```

### Failure Statistics
```
GET /api/failureStats/:id

Response:
{
  "data": {
    "total_failures": 15,
    "ping_failures": 10,
    "speed_failures": 3,
    "ip_check_failures": 2,
    "last_failure": "2025-11-27T10:30:00Z",
    "failure_rate": 2.5  // failures per day
  }
}
```

### Test Notification
```
POST /api/testNotification
Body:
{
  "message": "Testing notification system"
}

Response:
{
  "message": "Test notification sent successfully"
}
```

## Configuration

### Setting up Telegram Bot

1. **Create Bot:**
   - Message @BotFather on Telegram
   - Send `/newbot` command
   - Follow instructions to create bot
   - Save the bot token (format: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. **Get Chat ID:**

   **For personal notifications:**
   - Message your bot
   - Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
   - Find "chat": {"id": 123456789} in the response

   **For channel notifications:**
   - Add bot as channel admin
   - Post a message in the channel
   - Visit the getUpdates URL
   - Chat ID will be negative (e.g., -1001234567890)

3. **Configure in Application:**
   - Go to Settings page
   - Enable Telegram notifications
   - Enter Bot Token and Chat ID
   - Select which notifications to receive
   - Click Save
   - Use "Test Notification" button to verify

## Notification Logic

### Proxy Down Detection
- Triggers after 3 consecutive ping failures
- Only sends notification once (doesn't spam)
- Includes: proxy name, IP, username, error message
- Requires `NotifyOnDown: true`

### Proxy Recovery Detection
- Triggers when proxy status changes from 2 (dead) to 1 (alive)
- Only sends notification once per recovery
- Includes: proxy name, IP, username, latency
- Requires `NotifyOnRecovery: true`

### IP Change Detection
- Compares current real IP with previous IP
- Only triggers if old IP exists (not first check)
- Includes: old IP, new IP, country, operator
- Requires `NotifyOnIPChange: true`

### IP Stuck Detection
- Checks if IP hasn't changed for >24 hours
- Uses Stack field from Proxy model
- Calculates hours since last change
- Includes: stuck IP, duration, country, operator
- Requires `NotifyOnIPStuck: true`

### Low Speed Detection
- Triggers when speed < threshold (default: 10 Mbps)
- Only checks download speed (not upload)
- Ignores speed=0 (failed checks)
- Includes: download speed, upload speed, threshold
- Requires `NotifyOnLowSpeed: true`

## Database Schema

### ProxyFailureLog Table
```sql
CREATE TABLE proxy_failure_logs (
    id TEXT PRIMARY KEY,
    proxy_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    error_type TEXT NOT NULL,
    error_msg TEXT,
    latency INTEGER,
    FOREIGN KEY (proxy_id) REFERENCES proxies(id)
);

CREATE INDEX idx_proxy_failure_logs_proxy_id ON proxy_failure_logs(proxy_id);
CREATE INDEX idx_proxy_failure_logs_timestamp ON proxy_failure_logs(timestamp);
CREATE INDEX idx_proxy_failure_logs_error_type ON proxy_failure_logs(error_type);
```

## Usage Examples

### View All Failures
```bash
curl "http://localhost:8080/api/failureLogs?page=1&page_size=50"
```

### View Failures for Specific Proxy
```bash
curl "http://localhost:8080/api/failureLogs?proxy_id=uuid-here"
```

### View Only Ping Failures
```bash
curl "http://localhost:8080/api/failureLogs?error_type=ping_failed"
```

### View Failures in Date Range
```bash
curl "http://localhost:8080/api/failureLogs?start_date=2025-11-01&end_date=2025-11-27"
```

### Get Statistics for Proxy
```bash
curl "http://localhost:8080/api/failureStats/uuid-here"
```

### Send Test Notification
```bash
curl -X POST http://localhost:8080/api/testNotification \
  -H "Content-Type: application/json" \
  -d '{"message": "Test from API"}'
```

## Testing

### Manual Testing Steps

1. **Test Notification Service:**
   - Configure Telegram settings
   - Use Test Notification endpoint
   - Verify message received in Telegram

2. **Test Proxy Down Alert:**
   - Add invalid proxy or stop proxy server
   - Wait for IP check scheduler (15 min default)
   - Verify notification sent after 3 failures

3. **Test Proxy Recovery:**
   - Fix/restart proxy that was down
   - Wait for next IP check
   - Verify recovery notification

4. **Test IP Change:**
   - Enable NotifyOnIPChange
   - Rotate proxy IP manually
   - Wait for IP check
   - Verify notification with old/new IP

5. **Test Low Speed:**
   - Enable NotifyOnLowSpeed
   - Set LowSpeedThreshold higher than proxy speed
   - Wait for health check
   - Verify low speed alert

6. **Test Failure Logs:**
   - Trigger some failures (invalid proxies)
   - Query failure logs API
   - Verify logs are stored correctly
   - Check statistics endpoint

## Performance Considerations

- Notifications are sent asynchronously (don't block checks)
- HTTP client has 10-second timeout
- Failure logs indexed on proxy_id and timestamp
- Pagination prevents large result sets
- Statistics calculated on-demand (not cached)

## Security Notes

- Bot token stored in database (not encrypted)
- Consider encrypting sensitive settings in production
- HTML escaping prevents injection attacks
- Rate limiting not implemented (consider adding)
- Authentication required for all API endpoints (Basic Auth)

## Future Enhancements

1. **Daily Summary Implementation:**
   - Schedule daily report at configured time
   - Include: total proxies, alive/dead counts, avg speed
   - Already has notification method, needs scheduler

2. **Notification Rate Limiting:**
   - Prevent spam for frequently failing proxies
   - Implement cooldown period between same notifications
   - Add notification history tracking

3. **Multiple Notification Channels:**
   - Email support
   - Slack integration
   - Discord webhooks
   - SMS via Twilio

4. **Advanced Filtering:**
   - Failure log search by error message
   - Aggregate statistics (failures per day graph)
   - Export failure logs to CSV

5. **Notification Templates:**
   - Customizable message formats
   - Multi-language support
   - Emoji customization

## Troubleshooting

### Notifications Not Received

1. **Check Settings:**
   - Verify TelegramEnabled is true
   - Confirm token and chat ID are correct
   - Check notification toggle for specific event

2. **Test Connection:**
   - Use Test Notification endpoint
   - Check logs for "Telegram notification sent successfully"
   - Verify bot is not blocked

3. **Check Logs:**
   - Look for "Failed to send telegram notification" errors
   - Check for network/timeout errors
   - Verify Telegram API is accessible

### Failure Logs Not Appearing

1. **Check Database:**
   - Verify ProxyFailureLog table exists
   - Check AutoMigrate ran successfully
   - Look for database errors in logs

2. **Check Scheduler:**
   - Verify notification-enabled iterators are being called
   - Check logs for "Scheduler: ..." messages
   - Confirm checks are actually failing

3. **Check Filters:**
   - Verify query parameters are correct
   - Check date range includes expected failures
   - Try querying without filters

## Summary

The notification system provides comprehensive monitoring and alerting for proxy infrastructure:

✅ Telegram integration with rich formatting
✅ Failure history tracking with statistics
✅ Configurable alert types and thresholds
✅ RESTful API for accessing failure data
✅ Integration with existing scheduler system
✅ Test endpoint for verification
✅ Graceful error handling

The system is production-ready and can be extended with additional notification channels and features as needed.
