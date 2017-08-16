package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"

	log "github.com/sirupsen/logrus"
)

const datapath = "/tmp/data/data.zip"
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

	if err := loadData(store, datapath); err != nil {
		log.Fatal(err)
	}

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
			for _, u := range data.Users {
				if err := store.CreateUser(&u); err != nil {
					log.Warnf("Import error %v", err)
				}
			}
			log.Infof("Done")
		}
		if len(data.Visits) > 0 {
			log.Infof("Import %d visits", len(data.Visits))
			for _, v := range data.Visits {
				if err := store.CreateVisit(&v); err != nil {
					log.Warnf("Import error %v", err)
				}
			}
			log.Infof("Done")
		}
		if len(data.Locations) > 0 {
			log.Infof("Import %d locations", len(data.Locations))
			for _, l := range data.Locations {
				if err := store.CreateLocation(&l); err != nil {
					log.Warnf("Import error %v", err)
				}
			}
			log.Infof("Done")
		}
	}

	log.Infof("Done in %v", time.Now().Sub(start))

	return nil
}
