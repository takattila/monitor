#!/bin/bash

RED="\e[31m"
ENDCOLOR="\e[0m"
ERROR="${RED}[ERROR]${ENDCOLOR}"

function checkOS {
    if [[ "$(uname -s | awk '{print tolower($0)}')" != "linux" ]]; then
        echo -e "${ERROR} Sorry, there is no release for your OS for now."
        exit 1
    fi
}

function getArchitecture {
    local bits
    local arch
    local return

    bits="$(getconf LONG_BIT)"
    arch="$(uname -m)"
    return=""

    if [[ "$arch" == *"arm"* && "$bits" = "32" ]]; then
        return="arm"
    fi

    if [[ "$arch" == *"arm"* && "$bits" = "64" ]]; then
        return="arm64"
    fi

    if [[ "$arch" != *"arm"* && "$bits" = "32" ]]; then
        return="386"
    fi

    if [[ "$arch" != *"arm"* && "$bits" = "64" ]]; then
        return="amd64"
    fi

    echo "$return"
}

function getVersion {
    echo "$(wget -q -O- https://api.github.com/repos/takattila/monitor/releases/latest | jq -r '.tag_name')"
}

function getLatestReleaseURL {
    local version="$1"
    local architecture="$2"
    echo "https://github.com/takattila/monitor/releases/download/${version}/monitor-${version}-linux-${architecture}.zip"
}

function getWebConfigType {
    os="$(hostnamectl | grep Operating | awk -F: '{print $2}' | xargs | awk '{print $1}' | awk '{print tolower($0)}')"
    if [[ "$os" = "raspbian" ]]; then
        echo "raspbian"
    else
        echo "linux"
    fi
}

function getIP {
    echo "$(hostname -I | awk '{print $1}')"
}

function getPort {
    local monitorPath="$1"
    cat "${monitorPath}/configs/web.$(getWebConfigType).yaml" | grep "^  port:" | awk '{print $2}'
}

function getRoute {
    local monitorPath="$1"
    cat "${monitorPath}/configs/web.$(getWebConfigType).yaml" | grep "^    index:" | awk '{print $2}'
}

function installServices {
    local url="$1"
    local baseDir="/opt/"
    local monitorDir="monitor"

    cd "$baseDir"
    echo "=========================================="
    echo "[1./10.] Downloading $url to ${baseDir}..."
    echo "=========================================="
    sudo rm -f monitor-v*.zip
    sudo wget "$url"
    
    echo "=============================================="
    echo "[2./10.] Unzip monitor-v*.zip to ${baseDir}..."
    echo "=============================================="
    sudo unzip -o monitor-v*.zip -d monitor

    echo "=============================================================================="
    echo "[3./10.] Change ownership of the ${baseDir}${monitorDir} directory to $USER..."
    echo "=============================================================================="
    sudo chown ${USER}:${USER} ${baseDir}${monitorDir}
    sudo chown -R ${USER}:${USER} ${baseDir}${monitorDir}

    echo "======================================"
    echo "[4./10.] Change directory: $monitorDir"
    echo "======================================"
    cd "$monitorDir"

    echo "=============================="
    echo "[5./10.] Save your credentials"
    echo "=============================="
    sudo ./cmd/credentials

    echo "====================================================================="
    echo "[6./10.] Copy ${monitorDir}/tools/*.service to /etc/systemd/system..."
    echo "====================================================================="
    sudo cp tools/*.service /etc/systemd/system
    
    echo "========================="
    echo "[7./10.] Reload daemon..."
    echo "========================="
    sudo systemctl daemon-reload

    echo "============================="
    echo "[8./10.] Enabling services..."
    echo "============================="
    sudo systemctl enable monitor-api.service monitor-web.service
    sudo systemctl is-enabled monitor-api.service monitor-web.service

    echo "============================="
    echo "[9./10.] Starting services..."
    echo "============================="
    sudo systemctl start monitor-api.service monitor-web.service
    sudo systemctl is-active monitor-api.service monitor-web.service

    echo "==================="
    echo "[10./10.] Finished!"
    echo "==================="
    echo "Web interface:"
    echo "http://$(getIP):$(getPort "${baseDir}${monitorDir}")$(getRoute "${baseDir}${monitorDir}")"
}

function main {
    local architecture
    local version
    local url
  
    checkOS

    architecture="$(getArchitecture)"
    if [[ "$architecture" = "" ]]; then
        echo -e "${ERROR} Sorry, there is no release for your architecture for now."
        exit 1
    fi

    version="$(getVersion)"
    if [[ "$version" = "" ]]; then
        echo -e "${ERROR} Sorry, the latest release number cannot be fetched."
        exit 1
    fi

    url="$(getLatestReleaseURL "$version" "$architecture")"
    installServices "$url"
}

main
