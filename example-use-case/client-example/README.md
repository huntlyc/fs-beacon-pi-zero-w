# Client example

The example in this folder will check a specified git repo every 5 minutes and alert when there are incoming changes to pull down.

## Edit the shell script
Edit the `alarm-repo-check.sh` to look at your repo of choice and then place the script into a path accessible folder like `/usr/local/bin`

Run a `chmod +x /usr/local/bin/alarm-repo-check.sh` to make sure it's executable.

Create a service and timer combo via systemd (or use crontab if you're old skool)


## Systemd

Copy the service and timer files into `~/.config/systemd/user/`

Then run the following commands to setup.


```sh
systemctl --user daemon-reload
systemctl --user enable --now alarm-repo-check.timer
```


## Crontab

```sh
crontab -e

# then add the following line, tweak the time to suit
*/5 * * * * /usr/local/bin/alarm-repo-check.sh
```
