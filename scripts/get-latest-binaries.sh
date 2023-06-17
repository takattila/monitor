#!/bin/bash

RED="\e[31m"
GREEN="\e[32m"
YELLOW="\e[1;93m"
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
    wget -q -O- \
        https://api.github.com/repos/takattila/monitor/releases/latest \
        | grep "tag_name" \
        | awk '{print $2}' \
        | tr -d '"' \
        | tr -d ','
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
    echo "$(hostname)"
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
    echo -e "- ${YELLOW}[1./10.] ${GREEN}Downloading...${ENDCOLOR}"
    echo -e "  - ${GREEN}$url${ENDCOLOR}"
    echo -e "  - ${GREEN}to: ${baseDir}...${ENDCOLOR}"
    sudo rm -f monitor-v*.zip
    sudo wget -q --show-progress "$url"
    
    echo -e "- ${YELLOW}[2./10.] ${GREEN}Unzip monitor-v*.zip to ${baseDir}...${ENDCOLOR}"
    sudo unzip -q -o monitor-v*.zip -d monitor

    echo -e "- ${YELLOW}[3./10.] ${GREEN}Change ownership of the ${baseDir}${monitorDir} directory to $USER...${ENDCOLOR}"
    sudo chown ${USER}:${USER} ${baseDir}${monitorDir}
    sudo chown -R ${USER}:${USER} ${baseDir}${monitorDir}

    echo -e "- ${YELLOW}[4./10.] ${GREEN}Change directory: $monitorDir${ENDCOLOR}"
    cd "$monitorDir"

    echo -e "- ${YELLOW}[5./10.] ${GREEN}Save your credentials${ENDCOLOR}"
    sudo ./cmd/credentials

    echo -e "- ${YELLOW}[6./10.] ${GREEN}Copy ${monitorDir}/tools/*.service to /etc/systemd/system...${ENDCOLOR}"
    sudo cp tools/*.service /etc/systemd/system
    
    echo -e "- ${YELLOW}[7./10.] ${GREEN}Reload daemon...${ENDCOLOR}"
    sudo systemctl daemon-reload

    echo -e "- ${YELLOW}[8./10.] ${GREEN}Enabling services...${ENDCOLOR}"
    sudo systemctl enable monitor-api.service monitor-web.service
    echo "  - monitor-api: $(sudo systemctl is-enabled monitor-api.service)"
    echo "  - monitor-web: $(sudo systemctl is-enabled monitor-web.service)"

    echo -e "- ${YELLOW}[9./10.] ${GREEN}Starting services...${ENDCOLOR}"
    sudo systemctl start monitor-api.service monitor-web.service
    echo "  - monitor-api: $(sudo systemctl is-active monitor-api.service)"
    echo "  - monitor-web: $(sudo systemctl is-active monitor-web.service)"

    echo -e "- ${YELLOW}[10./10.] ${GREEN}Finished!${ENDCOLOR}"
    echo -e "  - $(cat /opt/monitor/VERSION.md | sed ':a;N;$!ba;s/\n/ /g')"
    echo -e "  - Web interface: ${YELLOW}http://$(getIP):$(getPort "${baseDir}${monitorDir}")$(getRoute "${baseDir}${monitorDir}")${ENDCOLOR}"
}

function setRootPassword {
    sudo -p "$(
        echo
        echo -e "- A password is required for installation."
        echo -e "  Please enter the ${YELLOW}root password${ENDCOLOR}: "
    )" echo -n "" 2> /dev/null
}

function clearScreen {
    echo -ne '\e]11;#000000\e\\' # set default foreground to black
    echo -ne '\e]10;#ffffff\e\\' # set default background to #abcdef
  
    clear
}

function printLogo {
    printf "${YELLOW}"
cat <<-'EOF'
      _____                .__  __                   
     /     \   ____   ____ |__|/  |_  ___________    
    /  \ /  \ /  _ \ /    \|  \   __\/  _ \_  __ \   
   /    Y    (  <_> )   |  \  ||  | (  <_> )  | \/   
   \____|__  /\____/|___|  /__||__|  \____/|__|      
           \/            \/                          
  _________                  .__                     
 /   _____/ ______________  _|__| ____  ____   ______
 \_____  \_/ __ \_  __ \  \/ /  |/ ___\/ __ \ /  ___/
 /        \  ___/|  | \/\   /|  \  \__\  ___/ \___ \ 
/_______  /\___  >__|    \_/ |__|\___  >___  >____  >
        \/     \/                    \/    \/     \/ 

                 ...installation...

EOF
    printf "${ENDCOLOR}\n"

}

function main {
    local architecture
    local version
    local url

    clearScreen
    printLogo
    checkOS
    setRootPassword

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
