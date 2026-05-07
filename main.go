package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/showwin/speedtest-go/speedtest"
)

var (
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

	lastRunTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "homelab_speedtest_last_run_timestamp",
		Help: "Timestamp Unix da última execução.",
	})

	lastRunSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "homelab_speedtest_last_run_success",
		Help: "1 se o último teste funcionou, 0 se falhou.",
	})
)

func runSpeedtest() {
	log.Println("iniciando speedtest...")

	user, err := speedtest.FetchUserInfo()
	if err != nil {
		log.Println("erro ao buscar user info:", err)
		lastRunSuccess.Set(0)
		lastRunTimestamp.Set(float64(time.Now().Unix()))
		return
	}

	serverList, err := speedtest.FetchServers(user)
	if err != nil {
		log.Println("erro ao buscar servidores:", err)
		lastRunSuccess.Set(0)
		lastRunTimestamp.Set(float64(time.Now().Unix()))
		return
	}

	targets, err := serverList.FindServer([]int{})
	if err != nil || len(targets) == 0 {
		log.Println("erro ao escolher servidor:", err)
		lastRunSuccess.Set(0)
		lastRunTimestamp.Set(float64(time.Now().Unix()))
		return
	}

	server := targets[0]

	if err := server.PingTest(nil); err != nil {
		log.Println("erro no ping:", err)
	}

	if err := server.DownloadTest(); err != nil {
		log.Println("erro no download:", err)
		lastRunSuccess.Set(0)
		lastRunTimestamp.Set(float64(time.Now().Unix()))
		return
	}

	if err := server.UploadTest(); err != nil {
		log.Println("erro no upload:", err)
		lastRunSuccess.Set(0)
		lastRunTimestamp.Set(float64(time.Now().Unix()))
		return
	}

	downloadMbps.Set(server.DLSpeed.Mbps())
	uploadMbps.Set(server.ULSpeed.Mbps())
	latencyMs.Set(float64(server.Latency.Milliseconds()))
	lastRunTimestamp.Set(float64(time.Now().Unix()))
	lastRunSuccess.Set(1)

	log.Printf("resultado: download=%.2f Mbps upload=%.2f Mbps latency=%dms",
		server.DLSpeed.Mbps(),
		server.ULSpeed.Mbps(),
		server.Latency.Milliseconds(),
	)
}

func main() {
	prometheus.MustRegister(downloadMbps)
	prometheus.MustRegister(uploadMbps)
	prometheus.MustRegister(latencyMs)
	prometheus.MustRegister(lastRunTimestamp)
	prometheus.MustRegister(lastRunSuccess)

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
