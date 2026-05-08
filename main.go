package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/showwin/speedtest-go/speedtest"
)

var (
	mu sync.Mutex

	downloadMbps = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "homelab_speedtest_download_mbps",
		Help: "Última velocidade de download medida em Mbps.",
	})

	uploadMbps = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "homelab_speedtest_upload_mbps",
		Help: "Última velocidade de upload medida em Mbps.",
	})

	latencyMs = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "homelab_speedtest_latency_ms",
		Help: "Última latência medida em ms.",
	})

	durationSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "homelab_speedtest_duration_seconds",
		Help: "Duração do último teste em segundos.",
	})

	lastRunTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "homelab_speedtest_last_run_timestamp",
		Help: "Timestamp Unix da última execução.",
	})

	lastRunSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "homelab_speedtest_last_run_success",
		Help: "1 se o último teste funcionou, 0 se falhou.",
	})

	failuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "homelab_speedtest_failures_total",
		Help: "Total de falhas nos testes de velocidade.",
	})
)

func markFailure(start time.Time, err error) {
	log.Println("erro no speedtest:", err)

	lastRunSuccess.Set(0)
	lastRunTimestamp.Set(float64(time.Now().Unix()))
	durationSeconds.Set(time.Since(start).Seconds())
	failuresTotal.Inc()
}

func runSpeedtest() {
	mu.Lock()
	defer mu.Unlock()

	start := time.Now()

	log.Println("iniciando speedtest...")

	serverList, err := speedtest.FetchServers()
	if err != nil {
		markFailure(start, err)
		return
	}

	targets, err := serverList.FindServer([]int{})
	if err != nil {
		markFailure(start, err)
		return
	}

	if len(targets) == 0 {
		markFailure(start, err)
		return
	}

	server := targets[0]

	if err := server.PingTest(nil); err != nil {
		markFailure(start, err)
		return
	}

	if err := server.DownloadTest(); err != nil {
		markFailure(start, err)
		return
	}

	if err := server.UploadTest(); err != nil {
		markFailure(start, err)
		return
	}

	downloadMbps.Set(server.DLSpeed.Mbps())
	uploadMbps.Set(server.ULSpeed.Mbps())
	latencyMs.Set(float64(server.Latency.Milliseconds()))
	durationSeconds.Set(time.Since(start).Seconds())
	lastRunTimestamp.Set(float64(time.Now().Unix()))
	lastRunSuccess.Set(1)

	log.Printf(
		"resultado: download=%.2f Mbps upload=%.2f Mbps latency=%dms duration=%.2fs",
		server.DLSpeed.Mbps(),
		server.ULSpeed.Mbps(),
		server.Latency.Milliseconds(),
		time.Since(start).Seconds(),
	)
}

func main() {
	prometheus.MustRegister(downloadMbps)
	prometheus.MustRegister(uploadMbps)
	prometheus.MustRegister(latencyMs)
	prometheus.MustRegister(durationSeconds)
	prometheus.MustRegister(lastRunTimestamp)
	prometheus.MustRegister(lastRunSuccess)
	prometheus.MustRegister(failuresTotal)

	go func() {
		runSpeedtest()

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			runSpeedtest()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	log.Println("ouvindo em :9108")
	log.Fatal(http.ListenAndServe(":9108", nil))
}
