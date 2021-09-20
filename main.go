package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type APIBoxList []struct {
	Box struct {
		ID int `json:"id"`
	} `json:"storagebox"`
}

type APIBoxDetail struct {
	Box Storagebox `json:"storagebox"`
}

type APIError struct {
	Error struct {
		Status int    `json:"status"`
		Code   string `json:"code"`
	} `json:"error"`
}

type Storagebox struct {
	ID                   int     `json:"id"`
	Login                string  `json:"login"`
	Name                 string  `json:"name"`
	Product              string  `json:"product"`
	Cancelled            bool    `json:"cancelled"`
	Location             string  `json:"location"`
	Locked               bool    `json:"locked"`
	LinkedServer         int     `json:"linked_server"`
	PaidUntil            string  `json:"paid_until"`
	DiskQuota            float64 `json:"disk_quota"`
	DiskUsage            float64 `json:"disk_usage"`
	DiskUsageData        float64 `json:"disk_usage_data"`
	DiskUsageSnapshots   float64 `json:"disk_usage_snapshots"`
	Webdav               bool    `json:"webdav"`
	Samba                bool    `json:"samba"`
	SSH                  bool    `json:"ssh"`
	BackupService        bool    `json:"backup_service"`
	ExternalReachability bool    `json:"external_reachability"`
	Zfs                  bool    `json:"zfs"`
	Server               string  `json:"server"`
	HostSystem           string  `json:"host_system"`
}

var (
	hetznerUsername string
	hetznerPassword string
	boxes           []Storagebox
	labels          = []string{"id", "name", "product", "server", "location", "host"}
	diskQuota       = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "storagebox",
			Name:      "disk_quota",
			Help:      "Total diskspace in MB",
		},
		labels,
	)
	diskUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "storagebox",
			Name:      "disk_usage",
			Help:      "Total used diskspace in MB",
		},
		labels,
	)
	diskUsageData = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "storagebox",
			Name:      "disk_usage_data",
			Help:      "Used diskspace by files in MB",
		},
		labels,
	)
	diskUsageSnapshots = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "storagebox",
			Name:      "disk_usage_snapshots",
			Help:      "Used diskspace by snapshots in MB",
		},
		labels,
	)
	locationRepresentation = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "storagebox",
			Name:      "location_hash",
			Help:      "Number representation of the location short name",
		},
		labels,
	)
	hostRepresentation = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "storagebox",
			Name:      "host_system_hash",
			Help:      "Number representation of the location short name",
		},
		labels,
	)
)

func updateBoxes() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://robot-ws.your-server.de/storagebox", nil)
	req.SetBasicAuth(hetznerUsername, hetznerPassword)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	bodyText, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		var apiErr APIError
		err = json.Unmarshal(bodyText, &apiErr)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("API Error: %d - %s", apiErr.Error.Status, apiErr.Error.Code)
		return
	}

	var apiResponse APIBoxList
	err = json.Unmarshal(bodyText, &apiResponse)
	if err != nil {
		log.Fatal(err)
	}

	boxes = nil
	for _, entry := range apiResponse {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://robot-ws.your-server.de/storagebox/%d", entry.Box.ID), nil)
		req.SetBasicAuth(hetznerUsername, hetznerPassword)
		resp, err = client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		bodyText, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			var apiErr APIError
			err = json.Unmarshal(bodyText, &apiErr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("API Error: %d - %s", apiErr.Error.Status, apiErr.Error.Code)
			return
		}

		var box APIBoxDetail
		err = json.Unmarshal(bodyText, &box)
		if err != nil {
			log.Fatal(err)
		}
		boxes = append(boxes, box.Box)
	}
}

func updateMetrics() {
	for {
		updateBoxes()
		for _, box := range boxes {
			diskQuota.With(prometheus.Labels{
				"id":       strconv.Itoa(box.ID),
				"name":     box.Name,
				"product":  box.Product,
				"server":   box.Server,
				"location": box.Location,
				"host":     box.HostSystem,
			}).Set(box.DiskQuota)

			diskUsage.With(prometheus.Labels{
				"id":       strconv.Itoa(box.ID),
				"name":     box.Name,
				"product":  box.Product,
				"server":   box.Server,
				"location": box.Location,
				"host":     box.HostSystem,
			}).Set(box.DiskUsage)

			diskUsageData.With(prometheus.Labels{
				"id":       strconv.Itoa(box.ID),
				"name":     box.Name,
				"product":  box.Product,
				"server":   box.Server,
				"location": box.Location,
				"host":     box.HostSystem,
			}).Set(box.DiskUsageData)

			diskUsageSnapshots.With(prometheus.Labels{
				"id":       strconv.Itoa(box.ID),
				"name":     box.Name,
				"product":  box.Product,
				"server":   box.Server,
				"location": box.Location,
				"host":     box.HostSystem,
			}).Set(box.DiskUsageSnapshots)

			locationRepresentation.With(prometheus.Labels{
				"id":       strconv.Itoa(box.ID),
				"name":     box.Name,
				"product":  box.Product,
				"server":   box.Server,
				"location": box.Location,
				"host":     box.HostSystem,
			}).Set(hash(box.Location))

			hostRepresentation.With(prometheus.Labels{
				"id":       strconv.Itoa(box.ID),
				"name":     box.Name,
				"product":  box.Product,
				"server":   box.Server,
				"location": box.Location,
				"host":     box.HostSystem,
			}).Set(hash(box.HostSystem))
		}

		// Try to avoid rate limiting
		// Limit is 200req / 1h
		// 200 requests / (box count + 2 to space out in case of restarts)
		requestsPerBox := 200 / (len(boxes) + 5)
		waitMinutes := 60 / requestsPerBox
		// minimum wait time are 5 minutes
		if waitMinutes < 5 {
			waitMinutes = 5
		}
		waitTime := time.Duration(waitMinutes) * time.Minute
		fmt.Println(fmt.Sprintf("Waiting for %s minutes to avoid ratelimting", waitTime))
		time.Sleep(waitTime)
	}
}

func hash(s string) float64 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return float64(h.Sum32())
}

var listenAddr string
var path string

func main() {

	hetznerUsername = os.Getenv("HETZNER_USER")
	hetznerPassword = os.Getenv("HETZNER_PASS")

	if hetznerUsername == "" || hetznerPassword == "" {
		log.Fatal("Please provide HETZNER_USER and HETZNER_PASS as environment variables")
	}
	flag.StringVar(&listenAddr, "listen", ":9509", "exporter listen port ':9509' or 'localhost:9509'")
	flag.StringVar(&path, "path", "/metrics", "exporter path default: '/metrics'")
	flag.Parse()
	prometheus.MustRegister(diskQuota)
	prometheus.MustRegister(diskUsage)
	prometheus.MustRegister(diskUsageData)
	prometheus.MustRegister(diskUsageSnapshots)
	prometheus.MustRegister(locationRepresentation)
	prometheus.MustRegister(hostRepresentation)

	go updateMetrics()

	fmt.Printf("Listening on %q", listenAddr)
	http.Handle(path, promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}
