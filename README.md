# PI Blinkybeacon HTTP REST API

Messing about with [blinkybeacon](https://github.com/duckfullstop/blinkybeacon#) on a pi

**Note**: Needs to run as **root** for the HID access, but opens a simple http server on port 1337 with the following endpoints:

- /strobe
- /strobe/{time}
- /spin
- /spin/{time}

Note time is an numeric value, ranged between 1-10

## Permanent Instalation

Only tested on a Raspberry Pi Zero W v1.1!

To make a binary and have it so you can run it by `beacon` do the following steps in the project root to build and then place it so it's available system wide:

```sh
$> make build-pi
$> sudo mv fsbeacon /usr/local/bin/beacon && sudo chmod +x /usr/local/bin/beacon
```

### Bonus: daemon

Creating a daemon on your pi too:
Create a file called `/etc/systemd/system/beacon.service` with the following contents:

***Note:*** if you placed your binary somewhere else, update ExecStart to match!

```sh
[Unit]
Description=Beacon REST API
After=multi-user.target

[Service]
Type=simple
ExecStart=/usr/local/bin/beacon
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```


Then enable it:
```sh
$> sudo systemctl daemon-reload && sudo systemctl enable beacon && sudo systemctl start beacon
```


To test: reboot, ssh in and run `wget http://127.0.0.1:1337/spin`
