package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mailru/easyjson"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const datapath = "/tmp/data/data.zip"
const optionspath = "/tmp/data/options.txt"
const listenAddr = ":80"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "warm-up" {
		warmUp()
		return
	}

	genTs, env := loadOptions(optionspath)
	log.Infof("Options: genTs=%d, env=%d", genTs, env)

	var store Store
	store = NewMemoryStore()

	if err := loadData(store, datapath); err != nil {
		log.Fatal(err)
	}
	runtime.GC()
	printMemoryStats()

	srv := NewServer(store)

	if env == 1 { // rating fire
		go runWarmUp(srv)
	}

	log.Infof("Start listening on address %s", listenAddr)
	log.Fatal(srv.Listen(listenAddr))
}

func loadOptions(filepath string) (ts int64, env int) {
	file, err := os.Open(filepath)
	ts = time.Now().Unix()
	env = 0
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() { // first line - data generation timestamp
		i, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err == nil {
			ts = i
		}
	}
	if scanner.Scan() { // second line - environment
		i, err := strconv.ParseInt(scanner.Text(), 10, 0)
		if err == nil {
			env = int(i)
		}
	}
	return
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

	for i, f := range files {
		log.Infof("Processing file %s", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Warnf("Failed to open data file %s: %v", f.Name, err)
			continue
		}

		var data FileData
		err = easyjson.UnmarshalFromReader(rc, &data)
		rc.Close()
		if err != nil {
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
		if i%10 == 0 {
			runtime.GC()
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

func runWarmUp(srv *Server) {
	cmd := exec.Command(os.Args[0], "warm-up")
	log.Infof("Start warm up")
	start := time.Now()
	err := cmd.Run()
	if err != nil {
		log.Warnf("Process finished with error: %v", err)
	}
	runtime.GC()
	log.Infof("Done warm up in %v", time.Now().Sub(start))
	printMemoryStats()
	srv.EnableStageGC()
}

func warmUp() {
	requests := []string{
		"/users/%d",
		"/locations/%d",
		"/visits/%d",
		"/users/%d/visits",
		"/users/%d/visits?fromDate=1203861305",
		"/users/%d/visits?toDate=1503861305",
		"/users/%d/visits?country=Russia",
		"/users/%d/visits?toDistance=4200",
		"/locations/%d/avg",
		"/locations/%d/avg?fromDate=1203861305",
		"/locations/%d/avg?toDate=1503861306",
		"/locations/%d/avg?fromAge=20",
		"/locations/%d/avg?toAge=60",
		"/locations/%d/avg?gender=f",
		"/locations/%d/avg?gender=m",
	}
	time.Sleep(1 * time.Second)
	log.Info("Start warm up")
	start := time.Now()
	rand.Seed(start.Unix())

	for i := 0; i < 500000; i++ {
		req := requests[rand.Intn(len(requests))]
		id := rand.Intn(1000000) + 1
		path := fmt.Sprintf(req, id)
		request(path)
	}

	runtime.GC()
	log.Infof("Done warm up in %v", time.Now().Sub(start))
	printMemoryStats()
}

func request(path string) {
	if _, _, err := fasthttp.Get(nil, "http://localhost"+listenAddr+path); err != nil {
		log.Errorf("Request '%s' error: %v", path, err)
	}
}
