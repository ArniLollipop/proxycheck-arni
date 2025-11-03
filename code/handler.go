package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type handler struct {
	db       *gorm.DB
	settings *Settings
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
	p.LastLatency = latency
	p.RealIP, p.RealCountry = RealIp(h.settings, p)

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

	// Перепроверяем прокси после обновления данных
	err := h.createAndCheckProxy(&p)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	}
	p.LastLatency = latency
	p.LastStatus = 1
	p.RealIP, p.RealCountry = RealIp(h.settings, &p)

	err = p.Save(h.db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h handler) VerifyBatch(c *gin.Context) {
	ids := c.Query("ids")
	idsArray := strings.Split(ids, ",")
	for _, id := range idsArray {
		var p Proxy
		err := p.Get(h.db, id)
		if err != nil {
			log.Println(err)
			continue
		}
		latency, err := Ping(h.settings, &p)
		if err != nil {
			log.Println(err)
		}
		p.LastLatency = latency
		p.LastStatus = 1
		p.RealIP, p.RealCountry = RealIp(h.settings, &p)
		err = p.Save(h.db)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": "Proxies verified"})
}

func (h handler) ImportProxies(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer openedFile.Close()

	scanner := bufio.NewScanner(openedFile)
	importedCount := 0
	failedLines := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			failedLines++
			continue
		}

		p := Proxy{
			Id:       uuid.NewString(),
			Ip:       parts[0],
			Port:     parts[1],
			Username: parts[2],
			Password: parts[3],
		}

		if len(parts) > 4 {
			p.Name = parts[4]
		}

		if err := h.createAndCheckProxy(&p); err != nil {
			log.Printf("Failed to import proxy %s:%s - %v", p.Ip, p.Port, err)
			failedLines++
		} else {
			importedCount++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading file for import:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Import finished. Imported: %d, Failed: %d", importedCount, failedLines),
		"importedCount": importedCount,
		"failedCount":   failedLines,
	})
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

	c.Header("Content-Disposition", "attachment; filename=proxies.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := []string{"ip", "port", "username", "password", "last_latency", "last_status", "failures", "real_ip", "real_country", "tag", "name", "contacts", "phone"}
	if err := writer.Write(headers); err != nil {
		log.Println("Error writing CSV header:", err)
		return
	}

	for _, proxy := range list {
		row := []string{
			proxy.Ip,
			proxy.Port,
			proxy.Username,
			proxy.Password,
			fmt.Sprint(proxy.LastLatency),
			strconv.Itoa(proxy.LastStatus),
			strconv.Itoa(proxy.Failures),
			proxy.RealIP,
			proxy.RealCountry,
			proxy.Tag,
			proxy.Name,
			proxy.Contacts,
			proxy.Phone,
		}
		if err := writer.Write(row); err != nil {
			log.Println("Error writing CSV row:", err)
			continue
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

	c.Header("Content-Disposition", "attachment; filename=selected_proxies.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := []string{"ip", "port", "username", "password", "last_latency", "last_status", "failures", "real_ip", "real_country", "tag", "name", "contacts", "phone"}
	if err := writer.Write(headers); err != nil {
		log.Println("Error writing CSV header:", err)
		return
	}

	for _, proxy := range proxies {
		row := []string{
			proxy.Ip,
			proxy.Port,
			proxy.Username,
			proxy.Password,
			fmt.Sprint(proxy.LastLatency),
			strconv.Itoa(proxy.LastStatus),
			strconv.Itoa(proxy.Failures),
			proxy.RealIP,
			proxy.RealCountry,
			proxy.Tag,
			proxy.Name,
			proxy.Contacts,
			proxy.Phone,
		}
		if err := writer.Write(row); err != nil {
			log.Println("Error writing CSV row:", err)
			continue
		}
	}
}
