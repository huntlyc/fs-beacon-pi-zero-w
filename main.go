package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/duckfullstop/blinkybeacon/pkg/fsbeacon"
)

// there can be only one beacon...
var beaconMu sync.Mutex

func main() {
	http.HandleFunc("GET /favicon.ico", faviconHandler)

	// home
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request){
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))
	})

	// spin routes, time and default
	http.HandleFunc("GET /spin", spinReqHandler)
	http.HandleFunc("GET /spin/{time}/", spinReqHandler)

	// stobe routes, time and default
	http.HandleFunc("GET /strobe/{time}/", strobeReqHandler)
	http.HandleFunc("GET /strobe", strobeReqHandler)


	// catchall
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		w.WriteHeader(http.StatusNotFound);
		w.Write([]byte("Like my sanity, not found"))
	})

	log.Println("Server running on :1337")
	log.Fatal(http.ListenAndServe(":1337", nil))
}



func getTimeInt(timeStr string) (int, error) {
	var err error
	time := 1

	if timeStr != "" {
		time, err = strconv.Atoi(timeStr)
		if err == nil {
			if time > 10 {
				time = 10
			} else if time < 1 {
				time = 1
			}
		}
	}

	return time, err
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// Tell the browser to expect an SVG image
	w.Header().Set("Content-Type", "image/svg+xml")

	// The SVG string with the alarm emoji
	emojiSVG := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text y=".9em" font-size="90">🚨</text></svg>`

	w.Write([]byte(emojiSVG))
}

func spinReqHandler(w http.ResponseWriter, r *http.Request) {
	timeStr := r.PathValue("time")
	time, _ := getTimeInt(timeStr)

	fmt.Printf("Spinnning for %v\n", time);
	err := handleSpin(time)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Err"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func strobeReqHandler(w http.ResponseWriter, r *http.Request) {
	timeStr := r.PathValue("time")
	time, _ := getTimeInt(timeStr)
	fmt.Printf("stroe for %v\n", time);

	err := handleStrobeBeacon(time)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Err"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func handleSpin(t int) error {
	beaconMu.Lock()
	defer beaconMu.Unlock()

	runtime := time.Duration(t) * time.Second

	d, err := fsbeacon.OpenFarmBeacon()
	if err != nil {
		return err
	}
	defer d.Close()

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

func handleStrobeBeacon(t int) error {
	beaconMu.Lock()
	defer beaconMu.Unlock()

	runtime := time.Duration(t) * time.Second

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
