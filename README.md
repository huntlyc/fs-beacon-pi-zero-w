# PI Blinkybeacon HTTP REST API

Messing about with [duckfullstop/blinkybeacon](https://github.com/duckfullstop/blinkybeacon#) on a pi

This project is a simple http server on port 1337 that will spin or strobe the beacon.

**Note**: Server needs to run as **root** for the HID access.

| Endpoint | Description |
| :--- | :--- |
| /strobe | 1s strobe |
| /strobe/{time} | strobe for *{time}* seconds, whole number between 1-10 |
| /spin | 1s spin |
| /spin/{time} | spin for *{time}* seconds, whole number between 1-10 |

## Uses

The possibilities are endless, but for an example that runs the server on a pi which gets called from another machine when a git repo has commits to pull down see [example-use-case](./example-use-case/)

