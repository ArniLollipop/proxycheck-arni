package main

import (
	"bufio"
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type handler struct{}

func (h handler) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	switch {
	case rq.URL.Path == "/":
		tpl, err := template.ParseFiles("resources/templates/index.html")
		if err != nil {
			log.Println(err)
			return
		}
		settings = SettingsDefault(dbSettings)
		tpl.Execute(rw, map[string]interface{}{
			"settings": settings,
		})

	case strings.HasPrefix(rq.URL.Path, "/resources"):
		http.ServeFile(rw, rq, "./"+rq.URL.Path)
		return

	case strings.HasPrefix(rq.URL.Path, "/ws"):
		var err error
		client.Connect, err = upgrader.Upgrade(rw, rq, nil)
		if err != nil {
			log.Println(err)
			return
		}

		go client.Listen()
		return

	case strings.HasPrefix(rq.URL.Path, "/import"):
		file, _, err := rq.FormFile("0")
		if err != nil {
			log.Println(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			proxy := &Proxy{}
			err := proxy.Save(dbProxy, line)
			if err != nil {
				log.Println(err)
				continue
			}
			wsMutex.Lock()
			client.Connect.WriteJSON(&Message{
				Cmd:   "add",
				Value: proxy,
			})
			wsMutex.Unlock()
		}
	case strings.HasPrefix(rq.URL.Path, "/export_all"):
		buff := new(bytes.Buffer)
		keys, err := dbProxy.Keys(nil, 0, 0, true)
		if err != nil {
			log.Println(err)
			return
		}
		for _, key := range keys {
			var p Proxy
			err := dbProxy.Get(key, &p)
			if err != nil {
				log.Println(err)
				return
			}
			proxy := p.String()
			_, err = buff.WriteString(proxy + "\r\n")
			if err != nil {
				log.Println(err)
				return
			}
		}
		b := bytes.TrimRight(buff.Bytes(), "\r\n")
		buff.Truncate(0)
		_, err = buff.Write(b)
		if err != nil {
			log.Println(err)
			return
		}
		size := strconv.Itoa(buff.Len())
		rw.Header().Set("Content-Disposition", "attachment; filename=proxy.txt")
		rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		rw.Header().Set("Content-Length", size)
		io.Copy(rw, buff)

	case strings.HasPrefix(rq.URL.Path, "/export_selected"):
		buff := new(bytes.Buffer)
		list := rq.FormValue("list")
		listSplit := strings.Split(list, ",")
		listInt := sliceStrToIntConvert(listSplit)
		for _, key := range listInt {
			var p Proxy
			err := dbProxy.Get(key, &p)
			if err != nil {
				log.Println(err)
				return
			}
			proxy := p.String()
			_, err = buff.WriteString(proxy + "\r\n")
			if err != nil {
				log.Println(err)
				return
			}
		}
		b := bytes.TrimRight(buff.Bytes(), "\r\n")
		buff.Truncate(0)
		_, err := buff.Write(b)
		if err != nil {
			log.Println(err)
			return
		}
		size := strconv.Itoa(buff.Len())
		rw.Header().Set("Content-Disposition", "attachment; filename=proxy.txt")
		rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		rw.Header().Set("Content-Length", size)
		io.Copy(rw, buff)

	case strings.HasPrefix(rq.URL.Path, "/settings"):
		settingsMutex.Lock()
		settings = SettingsDefault(dbSettings)
		val := rq.PostFormValue("value")
		var valInt int
		var err error
		if rq.PostFormValue("name") != "url" {
			valInt, err = strconv.Atoi(val)
			if err != nil {
				log.Println(err)
			}
		}
		if val == "" {
			return
		}
		switch rq.PostFormValue("name") {
		case "url":
			settings.Url = val
		case "timeout":
			settings.Timeout = valInt
		case "threads":
			settings.Threads = valInt
		case "repeat":
			settings.Repeat = valInt
		}

		err = dbSettings.Set("settings", settings)
		if err != nil {
			log.Println(err)
		}
		settingsMutex.Unlock()

	case strings.HasPrefix(rq.URL.Path, "/verify"):
		id := rq.PostFormValue("id")
		var proxy Proxy
		proxy.Get(dbProxy, id)
		latency, err := Ping(settings, &proxy)
		if err != nil {
			log.Println(err)
		}
		if latency == 0 {
			proxy.Failures += 1
			proxy.LastLatency = 0
			proxy.LastStatus = 2
		} else {
			proxy.LastLatency = latency
			proxy.LastStatus = 1
		}
		proxy.RealIP, proxy.RealCountry = RealIp(settings, &proxy)
		err = proxy.Update(dbProxy)
		if err != nil {
			log.Println(err)
			return
		}
		wsMutex.Lock()
		client.Connect.WriteJSON(&Message{
			Cmd:   "update",
			Value: proxy,
		})
		wsMutex.Unlock()

	case strings.HasPrefix(rq.URL.Path, "/reset_one"):
		id := rq.PostFormValue("id")
		var proxy Proxy
		proxy.Get(dbProxy, id)
		proxy.Failures = 0
		err := proxy.Update(dbProxy)
		if err != nil {
			log.Println(err)
			return
		}

	case strings.HasPrefix(rq.URL.Path, "/delete_one"):
		id := rq.PostFormValue("id")

		idInt, err := strconv.Atoi(id)
		if err != nil {
			log.Println(err)
			return
		}
		err = dbProxy.Delete(idInt)
		if err != nil {
			log.Println(err)
			return
		}

	case strings.HasPrefix(rq.URL.Path, "/change"):
		id := rq.PostFormValue("id")

		idInt, err := strconv.Atoi(id)
		if err != nil {
			log.Println(err)
			return
		}
		var proxy Proxy
		err = dbProxy.Get(idInt, &proxy)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = rw.Write([]byte(proxy.String()))
		if err != nil {
			log.Println(err)
			return
		}
	case strings.HasPrefix(rq.URL.Path, "/save"):
		id := rq.PostFormValue("id")
		val := rq.PostFormValue("value")
		var proxy Proxy
		idInt, err := strconv.Atoi(id)
		if err != nil {
			log.Println(err)
			return
		}

		err = dbProxy.Get(idInt, &proxy)
		if err != nil {
			log.Println(err)
			return
		}
		proxy.Parse(val)
		err = dbProxy.Set(proxy.Id, &proxy)
		if err != nil {
			log.Println(err)
			return
		}
		wsMutex.Lock()
		client.Connect.WriteJSON(&Message{
			Cmd:   "update",
			Value: proxy,
		})
		wsMutex.Unlock()

	case strings.HasPrefix(rq.URL.Path, "/delete_selected"):
		list := rq.PostFormValue("list")
		listSplit := strings.Split(list, ",")
		listInt := sliceStrToIntConvert(listSplit)
		for _, key := range listInt {
			err := dbProxy.Delete(key)
			if err != nil {
				log.Println(err)
				return
			}
		}

	case strings.HasPrefix(rq.URL.Path, "/reset_failures"):
		keys, err := dbProxy.Keys(nil, 0, 0, true)
		if err != nil {
			log.Println(err)
			return
		}
		for _, key := range keys {
			var p Proxy
			err := dbProxy.Get(key, &p)
			if err != nil {
				log.Println(err)
				return
			}
			p.Failures = 0
			err = dbProxy.Set(p.Id, &p)
			if err != nil {
				log.Println(err)
				return
			}
		}

	case strings.HasPrefix(rq.URL.Path, "/bench"):
		tpl, err := template.ParseFiles("resources/templates/bench.html")
		if err != nil {
			log.Println(err)
			return
		}
		settings := BenchSettingsDefault(dbSettings)
		err = tpl.Execute(rw, map[string]interface{}{
			"settings": settings,
			"status":   isBench,
		})
		if err != nil {
			log.Println(err)
		}

	case strings.HasPrefix(rq.URL.Path, "/bimport"):
		file, _, err := rq.FormFile("0")
		if err != nil {
			log.Println(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			arLine := strings.Split(line, ":")
			if len(arLine) < 2 {
				continue
			}
			addr := arLine[0] + ":" + arLine[1]

			bench := Bench{
				Addr: addr,
			}

			bench.ID, err = dbSettings.Counter("benchID", 1)
			if err != nil {
				log.Println("bench db counter: ", err)
				return
			}

			dbBench.mu.Lock()
			dbBench.benches[bench.ID] = bench
			dbBench.mu.Unlock()

			brsp := BenchResponse{
				ID:           bench.ID,
				Addr:         bench.Addr,
				RequestCount: bench.RequestCount,
				Latency:      bench.Latency,
				LastTime:     bench.LastTime,
				LastStatus:   bench.LastStatus,
			}

			wsMutex.Lock()
			client.Connect.WriteJSON(&Message{
				Cmd:   "badd",
				Value: brsp,
			})
			wsMutex.Unlock()
		}

	case strings.HasPrefix(rq.URL.Path, "/bsettings"):
		settingsMutex.Lock()
		defer settingsMutex.Unlock()
		benchSettings = BenchSettingsDefault(dbSettings)
		val := rq.PostFormValue("value")
		var valInt int
		var err error

		if val == "" {
			return
		}

		valInt, err = strconv.Atoi(val)
		if err != nil {
			log.Println(err)
		}

		switch rq.PostFormValue("name") {
		case "bench_timeout":
			benchSettings.Timeout = valInt
		case "bench_interval":
			benchSettings.Interval = valInt
		case "bench_reset":
			benchSettings.Reset = valInt
		}

		err = dbSettings.Set("benchSettings", benchSettings)
		if err != nil {
			log.Println(err)
		}

	case strings.HasPrefix(rq.URL.Path, "/breset"):
		isBench = false

		for _, b := range dbBench.benches {
			b.RequestCount = 0
			b.Latency = 0
			b.LatencySum = 0
			b.LastTime = ""
			b.LastStatus = 0
			b.Results = []BenchResult{}

			dbBench.mu.Lock()
			dbBench.benches[b.ID] = b
			dbBench.mu.Unlock()
		}

	case strings.HasPrefix(rq.URL.Path, "/bdelete"):
		isBench = false

		dbBench.mu.Lock()
		dbBench.benches = make(map[int64]Bench)
		dbBench.mu.Unlock()

	case strings.HasPrefix(rq.URL.Path, "/bexport"):
		isBench = false

		f := excelize.NewFile()

		idx := 0

		for _, b := range dbBench.benches {

			col, err := excelize.ColumnNumberToName(idx + 1)
			if err != nil {
				log.Println(err)
			}

			colNext, err := excelize.ColumnNumberToName(idx + 2)
			if err != nil {
				log.Println(err)
			}

			f.SetColWidth("Sheet1", col, colNext, 10)

			f.MergeCell("Sheet", col+"1", colNext+"1")
			f.SetCellValue("Sheet1", col+"1", b.Addr)
			for i, v := range b.Results {
				iStr := strconv.Itoa(i + 2)
				f.SetCellValue("Sheet1", col+iStr, v.Start)
				f.SetCellValue("Sheet1", colNext+iStr, v.End)
			}

			idx += 2
		}

		buff, err := f.WriteToBuffer()
		if err != nil {
			log.Println(err)
			return
		}

		size := strconv.Itoa(buff.Len())
		rw.Header().Set("Content-Disposition", "attachment; filename=bench.xlsx")
		rw.Header().Set("Content-Type", "application/octet-stream")
		rw.Header().Set("Content-Transfer-Encoding", "binary")
		rw.Header().Set("Content-Length", size)
		io.Copy(rw, buff)
	}
}
