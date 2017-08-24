package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const datapath = "/tmp/data/data.zip"
const listenAddr = ":80"

func main() {
	var store Store
	store = NewMemoryStore()

	if err := loadData(store, datapath); err != nil {
		log.Fatal(err)
	}
	runtime.GC()
	printMemoryStats()

	go warmUp()

	srv := NewServer(store)
	log.Infof("Start listening on address %s", listenAddr)
	log.Fatal(srv.Listen(listenAddr))
}

func loadData(store Store, filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		log.Info("No data to load")
		return nil
	}

	log.Infof("Load data from file %s", filepath)
	start := time.Now()

	if err := store.Clear(); err != nil {
		return fmt.Errorf("Failed to clear database: %v", err)
	}

	r, err := zip.OpenReader(filepath)
	if err != nil {
		return err
	}
	defer r.Close()

	var files []*zip.File
	for _, f := range r.File {
		// process only .json files
		if !strings.HasSuffix(f.Name, ".json") {
			continue
		}
		files = append(files, f)
	}

	sort.SliceStable(files, func(i, j int) bool {
		order := func(idx int) int {
			base := path.Base(files[idx].Name)
			if strings.HasPrefix(base, "users") {
				return 0
			} else if strings.HasPrefix(base, "locations") {
				return 1
			} else if strings.HasPrefix(base, "visits") {
				return 2
			}
			return 3
		}
		return order(i) < order(j)
	})

	for _, f := range files {
		log.Infof("Processing file %s", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Warnf("Failed to open data file %s: %v", f.Name, err)
			continue
		}
		buf, err := ioutil.ReadAll(rc)
		rc.Close()
		if err != nil {
			log.Warnf("Failed to read data file %s: %v", f.Name, err)
			continue
		}

		data := struct {
			Users     []User
			Visits    []Visit
			Locations []Location
		}{}
		if err := json.Unmarshal(buf, &data); err != nil {
			log.Warnf("Failed to read data: %v", err)
			continue
		}

		if len(data.Users) > 0 {
			log.Infof("Import %d users", len(data.Users))
			if err := store.CreateUsers(data.Users); err != nil {
				log.Warnf("Import error %v", err)
			}
			log.Infof("Done")
		}
		if len(data.Visits) > 0 {
			log.Infof("Import %d visits", len(data.Visits))
			if err := store.CreateVisits(data.Visits); err != nil {
				log.Warnf("Import error %v", err)
			}
			log.Infof("Done")
		}
		if len(data.Locations) > 0 {
			log.Infof("Import %d locations", len(data.Locations))
			if err := store.CreateLocations(data.Locations); err != nil {
				log.Warnf("Import error %v", err)
			}
			log.Infof("Done")
		}
	}

	log.Infof("Done in %v", time.Now().Sub(start))

	return nil
}

func printMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Infof("Memory stats:\nAlloc = %vM\nTotalAlloc = %vM\nSys = %vM\nNumGC = %v",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

func warmUp() {
	time.Sleep(1 * time.Second)
	log.Info("Start warm up")
	start := time.Now()

	// Stage 1 - Get users, locations, visits from 1 to 5000 each
	for k := 0; k < 5; k++ {
		for i := 1; i <= 5000; i++ {
			request(fmt.Sprintf("/users/%d", i))
			request(fmt.Sprintf("/locations/%d", i))
			request(fmt.Sprintf("/visits/%d", i))
		}
	}
	stage1 := time.Now()
	log.Infof("Stage 1 complete in %v", stage1.Sub(start))

	// Stage 2 - Get user visits from 1 to 15000
	for k := 0; k < 5; k++ {
		for i := 1; i <= 15000; i++ {
			request(fmt.Sprintf("/users/%d/visits", i))
		}
	}
	stage2 := time.Now()
	log.Infof("Stage 2 complete in %v", stage2.Sub(stage1))

	// Stage 3 - Get locations avg from 1 to 15000
	for k := 0; k < 5; k++ {
		for i := 1; i <= 15000; i++ {
			request(fmt.Sprintf("/locations/%d/avg", i))
		}
	}
	stage3 := time.Now()
	log.Infof("Stage 3 complete in %v", stage3.Sub(stage2))

	runtime.GC()
	log.Infof("Done warm up in %v", time.Now().Sub(start))
	printMemoryStats()
}

func request(path string) {
	if _, _, err := fasthttp.Get(nil, "http://localhost"+listenAddr+path); err != nil {
		log.Errorf("Request '%s' error: %v", path, err)
	}
}
