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

function mergeConfigFile {
    local oldFile="$1"
    local newFile="$2"
    local oldIndex
    local newIndex

    oldIndex="$(grep '^    index:' "${oldFile}" | head -n1 | awk '{print $2}')"
    newIndex="$(grep '^    index:' "${newFile}" | head -n1 | awk '{print $2}')"

    awk -v oldFile="${oldFile}" -v newFile="${newFile}" -v oldIndex="${oldIndex}" -v newIndex="${newIndex}" '
BEGIN {
    while ((getline line < oldFile) > 0) {
        oldCount++
        oldLines[oldCount] = line
    }
    close(oldFile)

    while ((getline line < newFile) > 0) {
        newCount++
        newLines[newCount] = line
    }
    close(newFile)

    buildPaths("old")
    buildPaths("new")
    collectInsertions()
    applyInsertions()

    for (i = 1; i <= outCount; i++) {
        print outLines[i]
    }
    exit
}

function indentOf(line,    n, c) {
    n = 0
    c = substr(line, n + 1, 1)
    while (c == " " || c == "\t") {
        n++
        c = substr(line, n + 1, 1)
    }
    return n
}

function isKeyLine(line) {
    return line ~ /^[ \t]*[A-Za-z0-9_.-]+:/
}

function extractKey(line,    k) {
    k = line
    sub(/^[ \t]*/, "", k)
    sub(/:.*$/, "", k)
    return k
}

function joinPath(stackName, depth,    i, p) {
    p = stackName[1]
    for (i = 2; i <= depth; i++) p = p "/" stackName[i]
    return p
}

function parentOf(path,    i, n) {
    n = length(path)
    i = n
    while (i > 0 && substr(path, i, 1) != "/") i--
    if (i == 0) return ""
    return substr(path, 1, i - 1)
}

function buildPaths(kind,    cnt, i, line, indent, key, path, depth, stackInd, stackName, stackPath) {
    depth = 0
    if (kind == "old") {
        cnt = oldCount
    } else {
        cnt = newCount
    }

    for (i = 1; i <= cnt; i++) {
        if (kind == "old") line = oldLines[i]; else line = newLines[i]
        if (line ~ /^[ \t]*$/ || line ~ /^[ \t]*#/) continue
        indent = indentOf(line)
        if (isKeyLine(line)) {
            while (depth > 0 && stackInd[depth] >= indent) {
                if (kind == "old") oldBlockEnd[stackPath[depth]] = i - 1
                else newBlockEnd[stackPath[depth]] = i - 1
                depth--
            }
            depth++
            stackInd[depth] = indent
            key = extractKey(line)
            stackName[depth] = key
            path = joinPath(stackName, depth)
            stackPath[depth] = path
            if (kind == "old") {
                oldBlockStart[path] = i
                oldPathSet[path] = 1
            } else {
                newBlockStart[path] = i
                newPaths[++newPathsCount] = path
            }
        }
    }
    for (i = depth; i >= 1; i--) {
        if (kind == "old") oldBlockEnd[stackPath[i]] = cnt
        else newBlockEnd[stackPath[i]] = cnt
    }
}

function routeTransformValue(path, value,    v) {
    if (newIndex == "" || oldIndex == "") return value
    if (path !~ /^on_start\/routes\//) return value
    v = value
    if (v == newIndex) return oldIndex
    if (index(v, newIndex "/") == 1) {
        return oldIndex substr(v, length(newIndex) + 1)
    }
    return value
}

function transformKeyLine(path, line,    i, keyPart, valPart) {
    if (newIndex == "" || oldIndex == "") return line
    if (path !~ /^on_start\/routes\//) return line
    i = index(line, ":")
    if (i == 0) return line
    keyPart = substr(line, 1, i)
    valPart = substr(line, i + 1)
    sub(/^[ \t]*/, "", valPart)
    sub(/[ \t]*$/, "", valPart)
    return keyPart " " routeTransformValue(path, valPart)
}

function collectInsertions(    n, i, path, parent) {
    n = 0
    for (i = 1; i <= newPathsCount; i++) {
        path = newPaths[i]
        if (path in oldPathSet) continue
        parent = parentOf(path)
        if (parent != "" && !(parent in oldPathSet)) continue
        n++
        insPos[n] = (parent == "" ? oldCount : oldBlockEnd[parent])
        insStart[n] = newBlockStart[path]
        insEnd[n] = newBlockEnd[path]
        insPath[n] = path
    }
    insCount = n
    for (i = 2; i <= n; i++) {
        keyPos = insPos[i]
        keyStart = insStart[i]
        keyEnd = insEnd[i]
        keyPath = insPath[i]
        j = i - 1
        while (j >= 1 && insPos[j] < keyPos) {
            insPos[j + 1] = insPos[j]
            insStart[j + 1] = insStart[j]
            insEnd[j + 1] = insEnd[j]
            insPath[j + 1] = insPath[j]
            j--
        }
        insPos[j + 1] = keyPos
        insStart[j + 1] = keyStart
        insEnd[j + 1] = keyEnd
        insPath[j + 1] = keyPath
    }
}

function applyInsertions(    i, k, pos, m, j, line, start) {
    for (i = 1; i <= oldCount; i++) outLines[i] = oldLines[i]
    outCount = oldCount
    for (k = 1; k <= insCount; k++) {
        pos = insPos[k]
        m = insEnd[k] - insStart[k] + 1
        for (i = outCount; i > pos; i--) outLines[i + m] = outLines[i]
        start = insStart[k]
        for (j = 1; j <= m; j++) {
            line = newLines[start + j - 1]
            if (j == 1) line = transformKeyLine(insPath[k], line)
            outLines[pos + j] = line
        }
        outCount += m
    }
}
'
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
                echo "  - $program"
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
    local totalSteps="12"
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
                sudo cp -f ${monitorPath}/configs/auth.db ${cfgBackupPath}/auth.db >/dev/null 2>&1 || true
                sudo rm -rf ${monitorPath} >/dev/null 2>&1 || true
                echo -e "  - ${YELLOW}Backup saved to: ${cfgBackupPath}${ENDCOLOR}"
            else
                echo -e "  - ${YELLOW}Backup skipped...${ENDCOLOR}"
            fi
    else
        echo -e "- ${YELLOW}[2./${totalSteps}.] ${GREEN}There is no existing configuration, backup skipped...${ENDCOLOR}"
    fi

    echo -e "- ${YELLOW}[3./${totalSteps}.] ${GREEN}Unzip monitor-v*.zip to ${basePath}...${ENDCOLOR}"
        sudo unzip -q -o monitor-v*.zip -d monitor

        if [[ "$backupCfg" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            echo -e "  - ${YELLOW}Restoring configuration and adding new options...${ENDCOLOR}"
            for cfg in "${monitorPath}"/configs/*.yaml; do
                local cfgName
                cfgName="$(basename "${cfg}")"
                if [[ -e "${cfgBackupPath}/${cfgName}" ]]; then
                    local mergedFile="${cfgBackupPath}/${cfgName}.merged"
                    echo -e "  - ${GREEN}${cfgName}${ENDCOLOR}"
                    mergeConfigFile "${cfgBackupPath}/${cfgName}" "${cfg}" > "${mergedFile}"
                    sudo cp -f "${mergedFile}" "${cfg}"
                    rm -f "${mergedFile}"
                fi
            done
            sudo cp ${cfgBackupPath}/auth.db ${monitorPath}/configs/auth.db >/dev/null 2>&1 || true
        else
            echo -e "  - ${YELLOW}Using the default configuration...${ENDCOLOR}"
        fi

        sudo rm -f monitor-v*.zip 2>&1 || true

    echo -e "- ${YELLOW}[4./${totalSteps}.] ${GREEN}Change ownership of the ${monitorPath} directory to $USER...${ENDCOLOR}"
        sudo chown ${USER}:${USER} ${monitorPath}
        sudo chown -R ${USER}:${USER} ${monitorPath}

    echo -e "- ${YELLOW}[5./${totalSteps}.] ${GREEN}Change directory to: ${monitorPath}${ENDCOLOR}"
        cd "${monitorPath}"

    echo -e "- ${YELLOW}[6./${totalSteps}.] ${GREEN}Set web terminal user...${ENDCOLOR}"
        local webConfig="${monitorPath}/configs/web.$(getWebConfigType).yaml"
        local currentTerminalUser
        currentTerminalUser="$(grep '^  terminal_user:' "${webConfig}" | head -n1 | sed 's/^  terminal_user:[[:space:]]*//; s/"//g' | xargs)"
        if [[ -n "${currentTerminalUser}" ]]; then
            echo -e "  - ${YELLOW}Keeping terminal_user: ${currentTerminalUser}${ENDCOLOR}"
        elif grep -q "^  terminal_user:" "${webConfig}"; then
            sudo sed -i "s|^  terminal_user:.*|  terminal_user: ${USER}|" "${webConfig}"
            echo -e "  - ${GREEN}terminal_user: ${USER}${ENDCOLOR}"
        else
            sudo sed -i "/^  save_credentials:/a\  terminal_user: ${USER}" "${webConfig}"
            echo -e "  - ${GREEN}terminal_user: ${USER}${ENDCOLOR}"
        fi

    echo -e "- ${YELLOW}[7./${totalSteps}.] ${GREEN}Save your credentials${ENDCOLOR}"
        if [[ "$backupCfg" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            echo -e "  - ${YELLOW}Using backup...${ENDCOLOR}"
            sudo chown root:root ${monitorPath}/configs/auth.db >/dev/null 2>&1 || true
        else
            sudo ./cmd/credentials
            sudo chown root:root ${monitorPath}/configs/auth.db >/dev/null 2>&1 || true
        fi

    echo -e "- ${YELLOW}[8./${totalSteps}.] ${GREEN}Copy ${programDir}/tools/*.service to /etc/systemd/system...${ENDCOLOR}"
        sudo cp tools/*.service /etc/systemd/system
    
    echo -e "- ${YELLOW}[9./${totalSteps}.] ${GREEN}Reload daemon...${ENDCOLOR}"
        sudo systemctl daemon-reload

    echo -e "- ${YELLOW}[10./${totalSteps}.] ${GREEN}Enabling services...${ENDCOLOR}"
        sudo systemctl enable monitor-api.service monitor-web.service
        echo "  - monitor-api: $(sudo systemctl is-enabled monitor-api.service)"
        echo "  - monitor-web: $(sudo systemctl is-enabled monitor-web.service)"

    echo -e "- ${YELLOW}[11./${totalSteps}.] ${GREEN}Starting services...${ENDCOLOR}"
        sudo systemctl stop monitor-api.service monitor-web.service
        sudo systemctl start monitor-api.service monitor-web.service
        echo "  - monitor-api: $(sudo systemctl is-active monitor-api.service)"
        echo "  - monitor-web: $(sudo systemctl is-active monitor-web.service)"

    echo -e "- ${YELLOW}[12./${totalSteps}.] ${GREEN}Finished!${ENDCOLOR}"
        echo -e "  - $(cat /opt/monitor/VERSION.md | sed ':a;N;$!ba;s/\n/ /g')"
        echo -e "  - Web interface: ${YELLOW}http://$(getIP):$(getPort "${monitorPath}")$(getRoute "${monitorPath}")${ENDCOLOR}"
}

function setRootPassword {
    sudo -p "$(
        echo
        echo -e "- Root password is required for installation."
        echo -e "  The following actions need root privileges:"
        echo -e "  - Creating directories and setting ownership"
        echo -e "  - Copying service files to /etc/systemd/system"
        echo -e "  - Reloading and enabling systemd services"
        echo -e "  - Setting file permissions for auth.db"
        echo -e "  Please enter the ${YELLOW}root password${ENDCOLOR}: "
    )" echo -n "" 2> /dev/null
}

function clearScreen {
    echo -ne '\e]11;#000000\e\\' # set default foreground to black
    echo -ne '\e]10;#ffffff\e\\' # set default background to white
  
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
