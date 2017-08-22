package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"

	log "github.com/sirupsen/logrus"
)

const datapath = "/tmp/data/data.zip"
const optionspath = "/tmp/data/options.txt"
const listenAddr = ":80"
const mongodbURL = "mongodb://localhost/travels"

func main() {
	session, err := mgo.Dial(mongodbURL)
	if err != nil {
		log.Fatal(err)
	}
	store, err := NewMongoStore(session)
	if err != nil {
		log.Fatal(err)
	}

	datats := getDataTs(optionspath)
	log.Infof("Data timestamp: %v", datats)
	if err := loadData(store, datapath); err != nil {
		log.Fatal(err)
	}
	runtime.GC()

	srv := NewServer(store, datats)
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

	for _, f := range r.File {
		// process only .json files
		if !strings.HasSuffix(f.Name, ".json") {
			continue
		}
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

func getDataTs(filepath string) time.Time {
	file, err := os.Open(filepath)
	if err != nil {
		log.Errorf("Failed to open options file: %v", err)
		return time.Now()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		ts, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Errorf("Invalid timestamp '%s': %v", scanner.Text(), err)
			return time.Now()
		}
		return time.Unix(int64(ts), 0)
	}
	if err := scanner.Err(); err != nil {
		log.Errorf("Failed to read options file: %v", err)
	}
	return time.Now()
}
