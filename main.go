package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/duckfullstop/blinkybeacon/pkg/fsbeacon"
	"github.com/huntlyc/beacon-pi/lcd"
)

// DefaultDuration is the duration of the beacon spin or strobe
// if time not specified.
const DefaultDuration = 3

// there can be only one beacon...
var beaconMu sync.Mutex
var displayMu sync.Mutex
var display *lcd.LCD

type RequestWithMsg struct {
	Msg string `json:"msg"`
}

func main() {
	http.HandleFunc("GET /favicon.ico", faviconHandler)

	// root path
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!doctype html>
			<html lang="en">
			<head>
				<title>The Beacon</title>
				<style>
					*{
						margin:0;
						padding:0
					}
					body{
						font-family:arial,sans-serif;
						font-size:20px
					}
					h1{
						font-size:2rem;
						margin-bottom:1em;
						text-align:center
					}
					.container{
						margin:0 auto;
						padding:4rem 2rem;
						max-width:calc(640px - 2rem)
					}
					p{
						margin-bottom:1em;
						line-height:1.4
					}
					ul{
						padding-left:1em;
						margin-bottom:3em;
					}
					li{
						margin-bottom:0.6em;
					}
					code{
						display: inline-block;
						padding: 0.3em 0.5em;
						font-family: monospace;
						font-size: 1rem;
						background: #eee;
					}
					a[data-ajax]{
						display: inline-block;
						padding: 0.5em 0.8em;
						background: red;
						color: white;
						text-decoration: none;
						box-shadow: 2px 2px 2px orange;
						transition: all 0.2s ease
					}
					a[data-ajax]:hover,
					a[data-ajax]:active{
						box-shadow: 0px 0px 0px transparent !important;
						transform: translate(2px,2px);
					}
					a[data-ajax][disabled]{
						opacity:0.2;
					}
					a[data-ajax][disabled]:hover,
					a[data-ajax][disabled]:active{
						cursor: not-allowed;
					}
				</style>
			</head>
			<body>
				<section class="container">
					<h1>🚨 Beacon ready! 🚨</h1>
					<p>To get started you can call one of these endpoints:</p>
					<ul>
						<li><code>/spin/</code></li>
						<li><code>/strobe/</code></li>
					</ul>
					<p>You can additionally add a time value of 1-10 seconds after the route.</p>
					<ul>
						<li><code>/spin/6/</code></li>
						<li><code>/strobe/9/</code></li>
					</ul>
					<p style="text-align:center">
						<a data-ajax="1" href="/spin/2">Take it for a spin!</a>
					</p>
				</section>
				<script>
					const initBtns = () => {
						const ajaxBtns = document.querySelectorAll('[data-ajax="1"]');
						if(ajaxBtns){
							ajaxBtns.forEach((btn) => {
								btn.addEventListener('click', (e) => {
									e.preventDefault();
									if(btn.getAttribute("disabled")) return false;

									btn.setAttribute("disabled","disabled");
									fetch(btn.href).then(() => {
										btn.removeAttribute("disabled");
									})
								});
							});
						}
					}
					document.addEventListener("DOMContentLoaded", initBtns);
				</script>
			</body>
		</html>`))
	})

	// spin routes
	http.HandleFunc("GET /spin", spinReqHandler)
	http.HandleFunc("GET /spin/{time}/", spinReqHandler)
	http.HandleFunc("POST /spin", spinReqHandler)
	http.HandleFunc("POST /spin/{time}/", spinReqHandler)

	// stobe routes
	http.HandleFunc("GET /strobe", strobeReqHandler)
	http.HandleFunc("GET /strobe/{time}/", strobeReqHandler)
	http.HandleFunc("POST /strobe", strobeReqHandler)
	http.HandleFunc("POST /strobe/{time}/", strobeReqHandler)

	// catch all else 404
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := makeBeaconSpinForDuration(DefaultDuration)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Err"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Like my sanity, not found"))
	})

	// fire up the lcd
	var err error
	display, err = initLCD()
	if err != nil {
		log.Fatal(err)
	}
	defer display.Close()

	log.Println("Server running on :1337")

	go printLCDMessage(display, "Beacon Ready")

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
	msg := "BEACON LIT"

	var reqPayload RequestWithMsg
	if r.Body != nil && r.ContentLength != 0 {
		err := json.NewDecoder(r.Body).Decode(&reqPayload)
		if err != nil {
			log.Printf("Failed to decode JSON: %v", err)
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if reqPayload.Msg != "" {
			msg = reqPayload.Msg
		}
	}
	go printLCDMessage(display, msg)

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

// Responds to a /strobe or /strobe/{time} request.
func strobeReqHandler(w http.ResponseWriter, r *http.Request) {
	timeStr := r.PathValue("time")
	time, _ := getTimeInt(timeStr)
	msg := "BEACON LIT"

	var reqPayload RequestWithMsg
	if r.Body != nil && r.ContentLength != 0 {
		err := json.NewDecoder(r.Body).Decode(&reqPayload)
		if err != nil {
			log.Printf("Failed to decode JSON: %v", err)
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if reqPayload.Msg != "" {
			msg = reqPayload.Msg
		}
	}
	go printLCDMessage(display, msg)

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

	// Wait for the beacon worker goroutine to finish writing stop bytes
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

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

	// Wait for the beacon worker goroutine to finish writing stop bytes
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	return nil
}

func initLCD() (*lcd.LCD, error) {
	return lcd.New(0x20, 1)
}

// Centers text, if possible, on the 16 char display line. e.g.
// 16 char line:
// "1234567890123456".
// To print "hello", would be:
// "hello67890123456" (where numbers are empty space).
// After pad it becomes:
// "12345hello123456" (where numbers are empty space).
func spacePad(s string) string {
	maxLen := 16 // 16 col wide display, can't be bothered implementing scrolling
	if len(s) < maxLen {
		spaces := maxLen - len(s)
		start := math.Floor(float64(spaces) / 2.0)
		s = strings.Repeat(" ", int(start)) + s
	} else if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

func printLCDMessage(display *lcd.LCD, msg string) {

	displayMu.Lock()

	display.Backlight(true)
	display.Clear()
	display.SetCursor(0, 0)

	parts := strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n")

	if len(parts) > 0 {
		display.Print(spacePad(parts[0]))
		if len(parts) > 1 {
			display.SetCursor(0, 1)
			display.Print(spacePad(parts[1]))
		}
	}
	displayMu.Unlock()

	time.Sleep(60 * time.Second)
	displayMu.Lock()
	display.Backlight(false)
	displayMu.Unlock()
}
