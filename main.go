package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/duckfullstop/blinkybeacon/pkg/fsbeacon"
)

// there can be only one beacon...
var beaconMu sync.Mutex

func main() {
	http.HandleFunc("/spin", spinReqHandler)
	http.HandleFunc("/strobe", strobeReqHandler)

	log.Println("Server running on :1337")
	log.Fatal(http.ListenAndServe(":1337", nil))
}

func spinReqHandler(w http.ResponseWriter, r *http.Request) {
	err := handleSpin()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Err"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func strobeReqHandler(w http.ResponseWriter, r *http.Request) {
	err := handleStrobeBeacon()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Err"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleSpin() error {
	beaconMu.Lock()
	defer beaconMu.Unlock()

	runtime := 1 * time.Second

	d, err := fsbeacon.OpenFarmBeacon()
	if err != nil {
		return err
	}
	defer d.Close()

	fmt.Printf("Spinning beacon for %s.\n", runtime)

	err = d.Spin()
	if err != nil {
		return err
	}

	time.Sleep(runtime)

	err = d.Stop()
	if err != nil {
		return err
	}

	// Wait for the worker goroutine to finish writing stop bytes
	d.Mutex.Lock()
	d.Mutex.Unlock()

	return nil
}

func handleStrobeBeacon() error {
	beaconMu.Lock()
	defer beaconMu.Unlock()

	runtime := 5 * time.Second

	d, err := fsbeacon.OpenFarmBeacon()
	if err != nil {
		return err
	}
	defer d.Close()

	fmt.Printf("Strobing beacon for %s.\n", runtime)

	err = d.Flash()
	if err != nil {
		return err
	}

	time.Sleep(runtime)

	err = d.Stop()
	if err != nil {
		return err
	}

	// Wait for the worker goroutine to finish writing stop bytes
	d.Mutex.Lock()
	d.Mutex.Unlock()

	return nil
}
