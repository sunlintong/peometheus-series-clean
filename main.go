package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Metrics struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

const (
	prometheusURL = "http://localhost:9090"
	concurrentNum = 10 // must less than your series count
	seriesPrefix  = "istio_agent_"
)

func main() {
	// 1. get all mereics
	// 2. delete series
	// 3. clean tombstones
	r, _ := http.Get(prometheusURL + "/api/v1/label/__name__/values")
	body, _ := io.ReadAll(r.Body)
	metrics := Metrics{}
	json.Unmarshal(body, &metrics)

	begin := time.Now()
	count := 0

	targets := []string{}
	for _, v := range metrics.Data {
		if strings.HasPrefix(v, seriesPrefix) {
			targets = append(targets, v)
		}
	}

	if len(targets) < concurrentNum {
		log.Fatalf("concurrentNum must less than your series count: %d", len(targets))
	}

	// separate
	sliceCap := len(targets) / (concurrentNum - 1)
	urls := [(concurrentNum)][]string{}
	for i := 0; i < (concurrentNum - 1); i++ {
		urls[i] = targets[i*sliceCap : (i+1)*sliceCap]
	}
	urls[concurrentNum-1] = targets[(concurrentNum-1)*sliceCap:]

	// do in batch
	wg := sync.WaitGroup{}
	wg.Add(concurrentNum)
	mu := sync.Mutex{}
	for i := 0; i < concurrentNum; i++ {
		go func(i int) {
			for _, v := range urls[i] {
				mu.Lock()
				count++
				mu.Unlock()

				start := time.Now()
				url := fmt.Sprintf("%s/api/v1/admin/tsdb/delete_series?match[]=%s", prometheusURL, v)
				resp, _ := http.Post(url, "", nil)
				log.Printf("goroutine %d: %s	%d %s %d", i+1, url, resp.StatusCode, time.Since(start).String(), count)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	log.Printf("finished %d in %s", count, time.Since(begin).String())

	t := time.Now()
	http.Post(prometheusURL+"/api/v1/admin/tsdb/clean_tombstones", "", nil)
	log.Printf("clean tombstones in %s", time.Since(t).String())
}
