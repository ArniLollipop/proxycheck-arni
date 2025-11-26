package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type handler struct {
	db            *gorm.DB
	settings      *Settings
	geoIPClient   *GeoIPClient
	restartSignal chan<- struct{}
}

func (h handler) ProxyList(c *gin.Context) {
	var p Proxy
	list, err := p.List(h.db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

type ProxyRequest struct {
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Contacts string `json:"contacts"`
	Phone    string `json:"phone"`
	Name     string `json:"name"`
}

// createAndCheckProxy - вспомогательная функция для создания и проверки прокси
func (h handler) createAndCheckProxy(p *Proxy) error {
	latency, err := Ping(h.settings, p)
	if err != nil {
		log.Println(err)
		p.LastStatus = 2 // 2 - failed
		p.Failures = 1
	} else {
		p.LastStatus = 1 // 1 - success
	}
	if err != nil {
		log.Println(err)
		p.LastStatus = 2 // 2 - failed
		p.Failures = 1
	}
	p.LastLatency = latency

	realIp, realCountry, realOperator, _ := RealIp(h.settings, p, h.db, h.geoIPClient)

	p.RealIP = realIp
	p.RealCountry = realCountry
	p.Operator = realOperator

	return p.Save(h.db)
}

func (h handler) CreateProxy(c *gin.Context) {
	var req ProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := Proxy{
		Id:       uuid.NewString(),
		Ip:       req.Ip,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Contacts: req.Contacts,
		Phone:    req.Phone,
		Name:     req.Name,
	}
	err := h.createAndCheckProxy(&p)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})

}

func (h handler) UpdateProxy(c *gin.Context) {
	id := c.Param("id")
	var req ProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p Proxy
	if err := h.db.First(&p, "id = ?", id).Error; err != nil {
		log.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	// Обновляем поля
	p.Ip = req.Ip
	p.Port = req.Port
	p.Username = req.Username
	p.Password = req.Password
	p.Contacts = req.Contacts
	p.Phone = req.Phone
	p.Name = req.Name

	p.Save(h.db);

	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h handler) Verify(c *gin.Context) {
	id := c.Param("id")
	var p Proxy
	err := p.Get(h.db, id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}
	latency, err := Ping(h.settings, &p)
	if err != nil {
		log.Println(err)
		p.LastStatus = 2
		p.Failures += 1
	}
	speed, upload, err := CheckSpeed(h.settings, &p, h.db)
	if err != nil {
		log.Println(err)
	} else {
		p.LastStatus = 1
	}
	p.Speed = int(speed)
	p.Upload = int(upload)
	p.LastLatency = latency

	realIp, realCountry, realOperator, err := RealIp(h.settings, &p, h.db, h.geoIPClient)

	if err != nil {
		log.Println(err);
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return;
	}

	p.RealIP = realIp
	p.RealCountry = realCountry
	p.Operator = realOperator


	err = p.Save(h.db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h handler) VerifyBatch(c *gin.Context) {
	ids := strings.Split(c.Query("ids"), ",")
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids parameter is required"})
		return
	}

	w := c.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.String(http.StatusInternalServerError, "Streaming not supported")
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(": ping\n\n"))
	flusher.Flush()

	for i, id := range ids {
		id = strings.TrimSpace(id)

		var p Proxy
		if err := p.Get(h.db, id); err != nil {
			continue
		}

		// START
		startJSON, _ := json.Marshal(gin.H{"id": id, "current": i + 1, "total": len(ids)})
		w.Write([]byte(fmt.Sprintf("event:start\ndata:%s\n\n", startJSON)))
		flusher.Flush()
		log.Printf("✅ START for ID: %s", id)

		// Ваши проверки...
		latency, _ := Ping(h.settings, &p)
		p.LastLatency = latency
		speed, upload, _ := CheckSpeed(h.settings, &p, h.db)
		p.Speed = int(speed)
		p.Upload = int(upload)
		realIp, realCountry, realOperator, err := RealIp(h.settings, &p, h.db, h.geoIPClient)

		if err != nil {
			log.Println(err);
			continue;
		}

		p.RealIP = realIp
		p.RealCountry = realCountry
		p.Operator = realOperator

		p.Save(h.db)

		// PROGRESS
		progressJSON, err := json.Marshal(p)
		if err != nil {
			log.Println("failed to marshal proxy:", err)
		} else {
			w.Write([]byte(fmt.Sprintf("event:progress\ndata:%s\n\n", progressJSON)))
			flusher.Flush()
		}
		log.Printf("✅ PROGRESS for ID: %s", id)
	}

	// COMPLETE
	completeJSON, _ := json.Marshal(gin.H{"message": "done", "total": len(ids)})
	w.Write([]byte(fmt.Sprintf("event:complete\ndata:%s\n\n", completeJSON)))
	flusher.Flush()
}

func (h handler) ImportProxies(c *gin.Context) (message string){
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return ""
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return ""
	}
	defer openedFile.Close()

	scanner := bufio.NewScanner(openedFile)
	importedCount := 0
	failedLines := 0
	skippedDuplicates := 0

	var existingProxies []Proxy
	h.db.Select("ip, port, username").Find(&existingProxies)
	
	existingMap := make(map[string]bool)
	for _, ep := range existingProxies {
		key := fmt.Sprintf("%s:%s:%s", ep.Ip, ep.Port, ep.Username)
		existingMap[key] = true
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		proxyLine := parts[0]
		proxyName := ""
		proxyContacts := ""

		if len(parts) > 1 {
			proxyName = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			proxyContacts = strings.TrimSpace(parts[2])
		}

		p := Proxy{}
		p.Parse(proxyLine)
		
		if proxyName != "" {
			p.Name = proxyName
		}
		if proxyContacts != "" {
			p.Contacts = proxyContacts
		}

		key := fmt.Sprintf("%s:%s:%s", p.Ip, p.Port, p.Username)
		if existingMap[key] {
			log.Printf("Skipping duplicate proxy %s:%s", p.Ip, p.Port)
			skippedDuplicates++
			continue
		}

		p.Id = uuid.NewString()

		if err := p.Save(h.db); err != nil {
			log.Printf("Failed to import proxy %s:%s - %v", p.Ip, p.Port, err)
			failedLines++
		} else {
			importedCount++
			existingMap[key] = true 
			
			log.Printf("Imported proxy %s:%s with name: %s, contacts: %s", p.Ip, p.Port, proxyName, proxyContacts)
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Println("Error reading file for import:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
		return
	}

	msg := fmt.Sprintf("Import finished. Imported: %d, Skipped: %d, Failed: %d", 
			importedCount, skippedDuplicates, failedLines)

	c.JSON(http.StatusOK, gin.H{
		"message": msg,
		"importedCount": importedCount,
		"skippedCount":  skippedDuplicates,
		"failedCount":   failedLines,
	})

	return msg
}

func (h handler) Delete(c *gin.Context) {
	id := c.Param("id")
	var p Proxy
	err := p.Get(h.db, id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	err = p.Delete(h.db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "Proxy deleted"})
}

func (h handler) ExportAll(c *gin.Context) {
	var p Proxy
	list, err := p.List(h.db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve proxies"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=proxies.txt")
	c.Header("Content-Type", "text/plain")

	for _, proxy := range list {
		if _, err := c.Writer.WriteString(proxy.String() + "\n"); err != nil {
			log.Println("Error writing proxy to response:", err)
			return
		}
	}
}

func (h handler) ExportSelected(c *gin.Context) {
	idsQuery := c.Query("ids")
	if idsQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids query parameter is required"})
		return
	}
	ids := strings.Split(idsQuery, ",")

	var proxies []Proxy
	if err := h.db.Where("id IN ?", ids).Find(&proxies).Error; err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve selected proxies"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=selected_proxies.txt")
	c.Header("Content-Type", "text/plain")

	for _, proxy := range proxies {
		if _, err := c.Writer.WriteString(proxy.String() + "\n"); err != nil {
			log.Println("Error writing proxy to response:", err)
			// Прерываем, так как дальнейшая запись, скорее всего, не удастся
			return
		}
	}
}

func (h handler) GetSettings(c *gin.Context) {
	var s Settings
	settings, err := s.Get(h.db)
	if err != nil {
		log.Println("Error getting settings:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h handler) UpdateSettings(c *gin.Context) {
	var req Settings
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Invalid settings format:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем в базу данных
	if err := req.Save(h.db); err != nil {
		log.Println("Error saving settings:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save settings"})
		return
	}

	// Обновляем настройки в текущем экземпляре обработчика
	// Это полезно для немедленного ответа, но главное - последующий перезапуск
	*h.settings = req

	c.JSON(http.StatusOK, gin.H{"data": h.settings})

	// // Отправляем сигнал на перезапуск приложения
}

func (h handler) GetSpeedLogs(c *gin.Context) {
	var filters ProxySpeedLogFilters

	// Parse query parameters
	filters.ProxyId = c.Query("proxy_id")
	filters.SortField = c.Query("sort_field")

	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 0 {
			filters.Page = page
		} else {
			filters.Page = 1
		}
	} else {
		filters.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSize > 0 {
			filters.PageSize = pageSize
		} else {
			filters.PageSize = 20 // Default page size
		}
	} else {
		filters.PageSize = 20
	}

	const layout = "2006-01-02"
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		t, err := time.Parse(layout, startDateStr)
		if err == nil {
			filters.StartDate = t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD."})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		t, err := time.Parse(layout, endDateStr)
		if err == nil {
			filters.EndDate = t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD."})
			return
		}
	}

	var psl ProxySpeedLog
	logs, total, err := psl.List(filters, h.db)
	if err != nil {
		log.Println("Error fetching speed logs:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve speed logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
	})
}

func (h handler) GetProxyIPLogs(c *gin.Context) {
	var filters ProxyIPLogFilters

	// Parse query parameters
	filters.ProxyId = c.Query("proxy_id")
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		filters.Page = page
	} else {
		filters.Page = 1
	}
	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "15")); err == nil {
		filters.PageSize = pageSize
	} else {
		filters.PageSize = 15
	}

	const layout = "2006-01-02"
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		t, err := time.Parse(layout, startDateStr)
		if err == nil {
			filters.StartDate = t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD."})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		t, err := time.Parse(layout, endDateStr)
		if err == nil {
			filters.EndDate = t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD."})
			return
		}
	}

	filters.SortField = c.Query("sort_field")

	var p ProxyIPLog
	logs, count, err := p.List(filters, h.db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get IP logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": count,
	})
}

func (h handler) CreateProxyVisitLog(c *gin.Context) {
	var visitLogs []ProxyVisitLogs
	if err := c.ShouldBindJSON(&visitLogs); err != nil {
		log.Println("Invalid visit log format:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use CreateInBatches to avoid "too many SQL variables" error with SQLite
	batchSize := 100 // Insert 100 records at a time
	if err := h.db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&visitLogs, batchSize).Error; err != nil {
		log.Println("Error saving visit logs:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save visit logs"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": visitLogs})
}

func (h handler) GetProxyVisitLogs(c *gin.Context) {
	var filters ProxyVisitLogsFilters
	var err error

	filters.ProxyId = c.Query("proxy_id")
	if filters.ProxyId != "" {
		proxy := &Proxy{}
		if err = h.db.Where("id =?", filters.ProxyId).First(proxy).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
			return
		}
		filters.ProxyId = proxy.Username
	}
	filters.SourceIP = c.Query("source_ip")
	filters.TargetIP = c.Query("target_ip")
	filters.Domain = c.Query("domain")
	filters.SortField = c.Query("sort_field")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}
	filters.Page = page

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
		return
	}
	filters.PageSize = pageSize
	const layout = "2006-01-02"

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		filters.StartDate, err = time.Parse(layout, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use RFC3339 format (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		filters.EndDate, err = time.Parse(layout, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use RFC3339 format (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	}

	var proxyVisitLogs ProxyVisitLogs
	logs, count, err := proxyVisitLogs.List(filters, h.db)
	if err != nil {
		log.Println("Error fetching visit logs:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch visit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": count,
	})
}
