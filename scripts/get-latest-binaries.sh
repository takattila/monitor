#!/bin/bash

RED="\e[31m"
GREEN="\e[32m"
YELLOW="\e[1;93m"
ENDCOLOR="\e[0m"
ERROR="${RED}[ERROR]${ENDCOLOR}"
REQUIRED_PROGRAMS=(
    awk
    bash
    cat
    curl
    getconf
    grep
    hostnamectl
    python
    sed
    systemctl
    tr
    uname
    unzip
    wget
    xargs
)

function checkOS {
    if [[ "$OSTYPE" != "linux-gnu"* ]]; then
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


function checkProgramIsInstalled {
    local program=$1
    sudo which ${program} &> /dev/null
    echo $?
}

function checkAllProgramsInstalled {
    local shouldBeInstalled
    local check

    declare -A shouldBeInstalled

    echo -en "- Checking neccesary programs:\n"

    for p in ${REQUIRED_PROGRAMS[@]} ; do
        echo -en "  - ${YELLOW}${p}${ENDCOLOR}..."
        check=$(checkProgramIsInstalled "${p}")
        if [[ "$check" != "0" ]]; then
            shouldBeInstalled["${p}"]="$check"
            echo -e "${RED}[FAIL]${ENDCOLOR}"
        else
            echo -e "${GREEN}[PASS]${ENDCOLOR}"
        fi
    done

    echo

    if [[ "${#shouldBeInstalled[@]}" -gt 0 ]]; then
        echo -e "${ERROR} For a successful installation, the following programs must be installed:"
        for program in ${!shouldBeInstalled[@]}; do
            if [[ "${shouldBeInstalled[$program]}" = "1" ]]; then
                echo "- $program"
            fi
        done
        exit 1
    fi
}

function installServices {
    local url="$1"
    local basePath="/opt/"
    local programDir="monitor"
    local monitorPath="${basePath}${programDir}"
    local cfgBackupPath="${monitorPath}-cfg-backup"
    local totalSteps="11"
    local backupCfg="n"

    echo -e "- ${YELLOW}[1./${totalSteps}.] ${GREEN}Downloading...${ENDCOLOR}"
        sudo mkdir -p "${basePath}" >/dev/null 2>&1 || true
        cd "${basePath}"
        echo -e "  - ${GREEN}$url${ENDCOLOR}"
        echo -e "  - ${GREEN}to: ${basePath}...${ENDCOLOR}"
        sudo rm -f monitor-v*.zip 2>&1 || true
        sudo wget -q --show-progress "$url"

    if [[ -e "${monitorPath}" ]]; then
        echo -e "- ${YELLOW}[2./${totalSteps}.] ${GREEN}Backup existing configuration...${ENDCOLOR}"
            read -r -p $'  - '$(echo -e "${YELLOW}")'Do you want to keep your existing configuration?'$(echo -e "${ENDCOLOR}")' [y/N] ' backupCfg
            if [[ "$backupCfg" =~ ^([yY][eE][sS]|[yY])$ ]]; then
                echo -e "  - ${YELLOW}Creating backup...${ENDCOLOR}"
                sudo mkdir -p ${cfgBackupPath} >/dev/null 2>&1 || true
                sudo chown ${USER}:${USER} ${cfgBackupPath}
                sudo chown -R ${USER}:${USER} ${cfgBackupPath}
                sudo cp -f ${monitorPath}/configs/*.yaml ${cfgBackupPath} >/dev/null 2>&1 || true
                sudo rm -rf ${monitorPath} >/dev/null 2>&1 || true
        else
                echo -e "  - ${YELLOW}Backup skipped...${ENDCOLOR}"
        fi
    else
        echo -e "- ${YELLOW}[2./${totalSteps}.] ${GREEN}There is no existing configuration, backup skipped...${ENDCOLOR}"
    fi

    echo -e "- ${YELLOW}[3./${totalSteps}.] ${GREEN}Unzip monitor-v*.zip to ${basePath}...${ENDCOLOR}"
        sudo unzip -q -o monitor-v*.zip -d monitor
        sudo cp ${cfgBackupPath}/*.yaml ${monitorPath}/configs >/dev/null 2>&1 || true
        sudo rm -rf ${cfgBackupPath} >/dev/null 2>&1 || true
        sudo rm -f monitor-v*.zip 2>&1 || true

    echo -e "- ${YELLOW}[4./${totalSteps}.] ${GREEN}Change ownership of the ${monitorPath} directory to $USER...${ENDCOLOR}"
        sudo chown ${USER}:${USER} ${monitorPath}
        sudo chown -R ${USER}:${USER} ${monitorPath}

    echo -e "- ${YELLOW}[5./${totalSteps}.] ${GREEN}Change directory to: ${monitorPath}${ENDCOLOR}"
        cd "${monitorPath}"

    echo -e "- ${YELLOW}[6./${totalSteps}.] ${GREEN}Save your credentials${ENDCOLOR}"
        sudo ./cmd/credentials

    echo -e "- ${YELLOW}[7./${totalSteps}.] ${GREEN}Copy ${programDir}/tools/*.service to /etc/systemd/system...${ENDCOLOR}"
        sudo cp tools/*.service /etc/systemd/system
    
    echo -e "- ${YELLOW}[8./${totalSteps}.] ${GREEN}Reload daemon...${ENDCOLOR}"
        sudo systemctl daemon-reload

    echo -e "- ${YELLOW}[9./${totalSteps}.] ${GREEN}Enabling services...${ENDCOLOR}"
        sudo systemctl enable monitor-api.service monitor-web.service
        echo "  - monitor-api: $(sudo systemctl is-enabled monitor-api.service)"
        echo "  - monitor-web: $(sudo systemctl is-enabled monitor-web.service)"

    echo -e "- ${YELLOW}[10./${totalSteps}.] ${GREEN}Starting services...${ENDCOLOR}"
        sudo systemctl stop monitor-api.service monitor-web.service
        sudo systemctl start monitor-api.service monitor-web.service
        echo "  - monitor-api: $(sudo systemctl is-active monitor-api.service)"
        echo "  - monitor-web: $(sudo systemctl is-active monitor-web.service)"

    echo -e "- ${YELLOW}[11./${totalSteps}.] ${GREEN}Finished!${ENDCOLOR}"
        echo -e "  - $(cat /opt/monitor/VERSION.md | sed ':a;N;$!ba;s/\n/ /g')"
        echo -e "  - Web interface: ${YELLOW}http://$(getIP):$(getPort "${monitorPath}")$(getRoute "${monitorPath}")${ENDCOLOR}"
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
    checkAllProgramsInstalled
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
