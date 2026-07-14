package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/duckfullstop/blinkybeacon/pkg/fsbeacon"
)

// DefaultDuration is the duration of the beacon spin or strobe
// if time not specified.
const DefaultDuration = 3;

// there can be only one beacon...
var beaconMu sync.Mutex


func main() {
	http.HandleFunc("GET /favicon.ico", faviconHandler)

	// root path
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request){
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("🚨 Beacon ready! 🚨"))
	})

	// spin routes
	http.HandleFunc("GET /spin", spinReqHandler)
	http.HandleFunc("GET /spin/{time}/", spinReqHandler)

	// stobe routes
	http.HandleFunc("GET /strobe/{time}/", strobeReqHandler)
	http.HandleFunc("GET /strobe", strobeReqHandler)


	// catch all else 404
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		err := makeBeaconSpinForDuration(DefaultDuration)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Err"))
			return
		}
		w.WriteHeader(http.StatusNotFound);
		w.Write([]byte("Like my sanity, not found"))
	})

	log.Println("Server running on :1337")
	log.Fatal(http.ListenAndServe(":1337", nil))
}



// Returns a time between 1 and 10.
// If timeStr is not empty, attempt to convert to int
func getTimeInt(timeStr string) (int, error) {
	var err error
	time := DefaultDuration

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

// Gotta have a favicon.
// Returns an svg for it.
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// Tell the browser to expect an SVG image
	w.Header().Set("Content-Type", "image/svg+xml")

	// The SVG string with the alarm emoji
	emojiSVG := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text y=".9em" font-size="90">🚨</text></svg>`

	w.Write([]byte(emojiSVG))
}

// Responds to a /spin or /spin/{time} request.
func spinReqHandler(w http.ResponseWriter, r *http.Request) {
	timeStr := r.PathValue("time")
	time, _ := getTimeInt(timeStr)

	err := makeBeaconSpinForDuration(time)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Err"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// Responds to a /spin or /spin/{time} request.
func strobeReqHandler(w http.ResponseWriter, r *http.Request) {
	timeStr := r.PathValue("time")
	time, _ := getTimeInt(timeStr)

	err := makeBeaconStrobeForDuration(time)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Err"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// Makes the beacon spin for 't' seconds.
func makeBeaconSpinForDuration(t int) error {
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

// Makes the beacon strobe for 't' seconds.
func makeBeaconStrobeForDuration(t int) error {
	beaconMu.Lock()
	defer beaconMu.Unlock()

	runtime := time.Duration(t) * time.Second

	d, err := fsbeacon.OpenFarmBeacon()
	if err != nil {
		return err
	}
	defer d.Close()

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
