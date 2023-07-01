# Monitor Services

[![main test](https://github.com/takattila/monitor/actions/workflows/main-test.yaml/badge.svg)](https://github.com/takattila/monitor/actions/workflows/main-test.yaml)
[![Coverage Status](https://coveralls.io/repos/github/takattila/monitor/badge.svg?branch=master)](https://coveralls.io/github/takattila/monitor?branch=master)

[![main release](https://github.com/takattila/monitor/actions/workflows/main-release.yaml/badge.svg)](https://github.com/takattila/monitor/actions/workflows/main-release.yaml)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/takattila/monitor)](https://github.com/takattila/monitor/releases)

This repository is a service stack written in Go language **for monitoring a Raspberry PI device** (or any linux systems).

Currently, it is tested on: Raspberry PI 3 Model B, but it can be used on other devices as well.

The web interface is responsive, it has `desktop` and `mobile` modes.


![Monitor Service](assets/banner.png)

**It monitors:**

- `CPU`: usage, load, temperature
- `Memory`: total, used, free, cached, available, swap, video
- `Services`: listed in: `configs/api.yaml` under: `on_runtime.services_list` section
- `Top processes`
- `Network traffic`
- `Storage`
- `Uptime`

![Monitor Service](assets/features.png)

**Management features:**

- `restart` or `shutdown` the device
- `start`/`stop`/`restart` or `enable`/`disable` services

**Parts:**

- [API](#api-service)
- [Web](#web-service)

![Monitor Service](assets/footer.png)

## Installation

You can choose to [download the pre-built binaries](https://github.com/takattila/monitor/releases) or build the service yourself.

### By downloading the latest release

If you didn't install go, and you don't want to install it, you can download install the latest pre-built binaries with: `get-latest-binaries.sh`.

1. Install the latest release for your architecture

   ```bash
   bash -c "$(wget -q --no-check-certificate --no-cache --no-cookies -O- https://raw.githubusercontent.com/takattila/monitor/master/scripts/get-latest-binaries.sh)"
   ```

### By building the service from the sources

If you cannot find the correct OS/architecture in the [releases](https://github.com/takattila/monitor/releases), you can optionally build the services yourself. 

1. Download and install Go

   - The `go version` can be fetched from the [go.mod](go.mod) file.
   - Please follow the instructions here: https://go.dev/doc/install

1. Clone repository:

   ```
   sudo git clone git@github.com:takattila/monitor.git /opt/monitor
   ```

1. Change ownership of the /opt/monitor directory to $USER:

   ```
   sudo chown $USER:$USER /opt/monitor
   sudo chown -R $USER:$USER /opt/monitor
   ```

1. Change directory:

   ```
   cd /opt/monitor
   ```

1. Install dependencies:

   ```
   go mod tidy
   ```

1. Build programs:

   ```
   go build -o cmd internal/api/app/api.go
   go build -o cmd internal/web/app/web.go
   go build -o cmd internal/credentials/app/credentials.go
   ```
### Run the service

1. Save your credentials:

   ```
   sudo ./cmd/credentials
   ```

1. Copy service files:

   ```
   sudo cp tools/*.service /etc/systemd/system
   ```

1. Reload daemon:

   ```
   sudo systemctl daemon-reload
   ```

1. Enable services:

   ```
   sudo systemctl enable monitor-api.service monitor-web.service
   sudo systemctl is-enabled monitor-api.service monitor-web.service
   ```

1. Start services:

   ```
   sudo systemctl start monitor-api.service monitor-web.service
   sudo systemctl is-active monitor-api.service monitor-web.service
   ```

1. Open the web interface

   You can fetch the IP of the device with this command:

   ```
   hostname -I | awk '{print $1}'
   ```

   - `http://<IP-OF-THE-DEVICE>:8383/monitor`

# API Service

API service provides hardware statistics information from the Raspberry PI, by serving a JSON file.

## Run the service

```
go build -o cmd internal/api/app/api.go && ./cmd/api
```

## Reach the service

`http://<IP-OF-THE-DEVICE>:7070/all`

# Web Service

Web interface for monitoring the Raspberry PI with  management features:

- restart or shutdown the device
- start/stop/restart or enable/disable services

## Run the service

```
go build -o cmd internal/web/app/web.go && sudo ./cmd/web
```

## Reach the service

`http://<IP-OF-THE-DEVICE>:8383/monitor`

## Re-initialize the service

```
rm go.*
go mod init github.com/takattila/monitor
go mod tidy
```

Optional:
m

```
go mod vendor
```

# Run the service over HTTPS

1. Keep running the API service in the background
   
   Do not need to stop or restart the API service, let's keep running it in the background.

   You can check it's status by running:

   ```
   sudo systemctl status monitor-api.service
   ```

   If the service status is not "active", you can start it by running the command bellow:

   ```
   sudo systemctl start monitor-api.service
   ```

1. Edit the `configs/web.{raspbian|linux}.yaml`  configuration file:

   Replace the port to: 443 and add your domain:

   ```diff
    static:
   -  port: 8383
   -  domain: example.net
   +  port: 443
   +  domain: yourdomain.com
   ```

   If you want to use the base path as: `/` instead of: `/monitor`, you can simply change the routes as the followings:

   ```diff
      routes:
   -    index: /monitor
   -    login: /monitor/login
   -    logout: /monitor/logout
   -    internal: /monitor/internal
   -    api: /monitor/api/{statistics}
   -    systemctl: /monitor/systemctl/{action}/{service}
   -    power: /monitor/power/{action}
   -    toggle: /monitor/toggle/{section}/{status}
   -    web: /monitor/web
   +    index: /
   +    login: /login
   +    logout: /logout
   +    internal: /internal
   +    api: /api/{statistics}
   +    systemctl: /systemctl/{action}/{service}
   +    power: /power/{action}
   +    toggle: /toggle/{section}/{status}
   +    web: /web
   ```

   After that you can reach the web interface on: `http://<IP-OF-THE-DEVICE>:8383`

1. Rebuild the program

   ```
   sudo systemctl stop monitor-web.service
   go build -o cmd internal/web/app/web.go
   ```

1. Start the service

   ```
   sudo systemctl start monitor-web.service
   ```

# Run service with Caddy

[Caddy](https://caddyserver.com/) is an open source web server with automatic HTTPS.

If you installed it on your Raspberry PI, you can add a custom route, where the service can be reached.

Example Caddyfile:

```
yurdomain.com {
        route /monitor* {
              reverse_proxy http://127.0.0.1:8383 {
                  header_up X-Real-IP {remote}
              }
        }
}
```

# Directory structure

This project uses the directory structure as explained in: [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

# Easy configuration

You can find the configuration under the [configs](configs) directory.
Each `YAML` file is a configuration for each service.

## Configuration structure

Each configuration has two big sections: `on_startup` and `on_runtime`.

### On startup

The `on_startup` section means that all settings belonging to this section are applied when the service starts.

### On runtime

The `on_runtime` section means that all settings belonging to this section can be applied during the service running.

# Configuration explained

## api.yaml

```yaml
on_start:                                 # These settings can be applied only, when the service starts.
  port: 7070                              #  - The service can be reached under this port.
  routes:                                 #  - URL schema, which describe the interfaces for making requests to the service.
    all: /all                             #    - All hardware information merged into one JSON.
    model: /model                         #    - Provides a model name JSON.
    cpu: /cpu                             #    - Provides a cpu statistics JSON.
    memory: /memory                       #    - Provides a memory statistics JSON.
    processes: /processes                 #    - Provides a top 10 processes JSON.
    storages: /storages                   #    - Provides a storages JSON.
    services: /services                   #    - Provides a services list JSON.
    network: /network                     #    - Provides a network traffic JSON.
    toggle: /toggle/{section}/{status}    #    - The processes, storages, services, network JSON provision can be turned on or off.
  logger:                                 #  - Setup logging functionality.
    level: debug                          #    - From debug to none levels, the detail of the logging can be set.
    color: on                             #    - Colorizing the log output.
on_runtime:                               # - These settings can be applied during the service running.
  physical_memory: 1GB                    #   - Set the memory amount 'by hand'. It can be commented out, and the program will get the total memory.
  commands:                               #   - Commands to get hardware information.
    ...                                   # 
    processes:                            # 
      - dash                              #     - The Dash linux shell roughly 4x times faster than Bash.
      - -c                                # 
      - |                                 #     - We can use 'pipe' in YAML to write multi-line blocks...
        ps -ewwo pid,user,%mem,%cpu,cmd \ # 
          --sort=-%cpu --no-headers \     # 
          | head -n 10 \                  # 
          | tail -n 10                    # 
    ...                                   # 
  services_list:                          #   - List of services which we want to manage.
    - smbd                                #     - The service checks in the background, whether the service is
    - sshd                                #       - active or enabled
    - syslog                              #       - and also we can start, stop, restart, enable, disable it.
```

## web.yaml

```yaml
on_start:                                            # These settings can be applied only, when the service starts.
  port: 8383                                         # - The service can be reached under this port.
  domain: example.net                                # - If you want to run this service as a stand-alone web service, you can set your domain here.
  web_sources_directory: /web                        # - The source files of the web interface can be found under this directory.
  auth_file: /configs/auth.db                        # - Usernames and passwords are stored here.
  save_credentials: false                            # - Do we want to initialize the user credentials each time when the service starts?
  routes:                                            # - URL schema, which describe the interfaces for making requests to the service.
    index: /monitor                                  #   - Route to the index page.
    login: /monitor/login                            #   - Route to the login page. (Login required)
    logout: /monitor/logout                          #   - Route to the logout page. (Login required)
    internal: /monitor/internal                      #   - Route to the internal page. (Login required)
    api: /monitor/api/{statistics}                   #   - Route to the api page. (Login required)
    systemctl: /monitor/systemctl/{action}/{service} #   - Route to the systemctl page. (Login required)
    power: /monitor/power/{action}                   #   - Route to the power page. (Login required)
    toggle: /monitor/toggle/{section}/{status}       #   - Route to the toggle page. (Login required)
    web: /monitor/web                                #   - The files: html, js, css can be served under this route.
  pages:                                             # - HTML files path.
    login: /html/login.html                          #   - Index file path.
    internal: /html/monitor.html                     #   - The internal page file path.
  logger:                                            # - Setup logging functionality.
    level: debug                                     #   - From debug to none levels, the detail of the logging can be set.
    color: on                                        #   - Colorizing the log output.
on_runtime:                                          # - These settings can be applied during the service running.
  theme:                                             #   - Themes configuration.
    skin: suse                                       #     - Color scheme
    logo: rpi                                        #     - Image
  allowed_ip: 0.0.0.0                                #   - We can set the IP, from where the service can be reached.
                                                     #     - 0.0.0.0 -> means: any IP will be accepted.
                                                     #     - 10.1.1.34,10.3.4.5 -> means: multiple IP can be accepted.
  interval_seconds: 1                                #   - How many seconds are we want to query the API?
  api:                                               #   - API service related stuff.
    url: "http://127.0.0.1"                          #     - URL of the API.
    port: 7070                                       #     - Port of the API.
  commands:                                          #   - Commands for the device management.
    systemctl:                                       #     - Start, Stop, Restart, Enable, Disable a service
      - dash                                         #   
      - -c                                           #   
      - systemctl {action} {service}                 #   
    init:                                            #     - Restart or shutdown.
      - dash                                         # 
      - -c                                           # 
      - init {number}                                # 

```

# Switch between light and dark modes

You can **switch between light and dark modes** by clicking the `header` or the `footer`.

# Skin support

Both the `skin` and the `logo` of the web service can be modified in the [configs/web.yaml](configs/web.yaml).

```yaml
on_runtime:
  theme:
    skin: suse
    logo: rpi
```

## Available skins

- [centos](#centos)
- [fedora](#fedora)
- [github_blue](#github_blue)
- [github_green](#github_green)
- [github_purple](#github_purple)
- [github_red](#github_red)
- [github_yellow](#github_yellow)
- [manjaro](#manjaro)
- [mint](#mint)
- [opi](#opi)
- [redhat](#redhat)
- [rpi](#rpi)
- [suse](#suse)
- [ubuntu](#ubuntu)
- [vanilla](#vanilla)

## Available logos

- arch
- centos
- debian
- fedora
- tux
- manjaro
- mint
- opi
- pop
- redhat
- rpi
- suse
- ubuntu
- vanilla
- zorin

## Skins & logos location

You can find the available skins under [web/css](web/css) directory.
The logos can be found under the [web/img](web/img) directory.

- Each `CSS` file is a skin.
- Each `PNG` is a favicon.
- Each `SVG` file is a logo.

# Screenshots

## Base skin

<table align="center" border="0" cellpadding="1" cellspacing="1">
	<tbody>
		<tr>
			<td colspan="2" style="text-align:center">Dark</td>
		</tr>
		<tr>
			<td style="text-align:center">Desktop</td>
			<td style="text-align:center">Mobile</td>
		</tr>
		<tr bgcolor="#3f3f3f">
			<td style="text-align:center"><img src="assets/desktop-dark.png"></td>
			<td style="text-align:center"><img src="assets/mobile-dark.png"></td>
		</tr>
		<tr>
			<td colspan="2" style="text-align:center">Light</td>
		</tr>
		<tr>
			<td style="text-align:center">Desktop</td>
			<td style="text-align:center">Mobile</td>
		</tr>
		<tr bgcolor="#f1f1f1">
			<td style="text-align:center"><img src="assets/desktop-light.png"></td>
			<td style="text-align:center"><img src="assets/mobile-light.png"></td>
		</tr>
	</tbody>
</table>

### Services

![Monitor Service](assets/services.png)

### Top Processes

![Monitor Service](assets/processes.png)

### Network Traffic

![Monitor Service](assets/network.png)

### Storage

![Monitor Service](assets/storage.png)

### Power & Logout

![Monitor Service](assets/power-logout.png)

### Full

![Monitor Service](assets/desktop-full-light.png)

## Additinal skins

### CentOS

![Monitor Service](assets/desktop-centos-dark.png)
![Monitor Service](assets/desktop-centos-light.png)

### Fedora

![Monitor Service](assets/desktop-fedora-dark.png)
![Monitor Service](assets/desktop-fedora-light.png)

### github_blue

![Monitor Service](assets/desktop-github_blue-dark.png)
![Monitor Service](assets/desktop-github_blue-light.png)

### github_green

![Monitor Service](assets/desktop-github_green-dark.png)
![Monitor Service](assets/desktop-github_green-light.png)

### github_purple

![Monitor Service](assets/desktop-github_purple-dark.png)
![Monitor Service](assets/desktop-github_purple-light.png)

### github_red

![Monitor Service](assets/desktop-github_red-dark.png)
![Monitor Service](assets/desktop-github_red-light.png)

### github_yellow

![Monitor Service](assets/desktop-github_yellow-dark.png)
![Monitor Service](assets/desktop-github_yellow-light.png)

### Manjaro

![Monitor Service](assets/desktop-manjaro-dark.png)
![Monitor Service](assets/desktop-manjaro-light.png)

### Mint

![Monitor Service](assets/desktop-mint-dark.png)
![Monitor Service](assets/desktop-mint-light.png)

### OPI

![Monitor Service](assets/desktop-opi-dark.png)
![Monitor Service](assets/desktop-opi-light.png)

### Redhat

![Monitor Service](assets/desktop-redhat-dark.png)
![Monitor Service](assets/desktop-redhat-light.png)

### RPI

![Monitor Service](assets/desktop-dark.png)
![Monitor Service](assets/desktop-light.png)

### Suse

![Monitor Service](assets/desktop-suse-dark.png)
![Monitor Service](assets/desktop-suse-light.png)

### Ubuntu

![Monitor Service](assets/desktop-ubuntu-dark.png)
![Monitor Service](assets/desktop-ubuntu-light.png)

### Vanilla

![Monitor Service](assets/desktop-vanilla-dark.png)
![Monitor Service](assets/desktop-vanilla-light.png)

# Troubleshooting

## If the service crashes

If the service crashes, panics delete the `.cache/go-build` folder:

```
sudo rm -rf $HOME/.cache/go-build
```