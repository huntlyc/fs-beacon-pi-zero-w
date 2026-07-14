# Server Setup

After compiling the project using either `make` or `make build-pi` depending on your target, move the compiled binary to `sudo mv fsbeacon /usr/local/bin/beacon`.
Also ensure it's executable by running `sudo chmod +x /usr/local/bin/beacon`

Then put the beacon.service file into `/etc/systemd/system/`.

Then run the following commands to enable the service:

```sh
sudo systemctl daemon-reload
sudo systemctl enable beacon
sudo systemctl start beacon
```
