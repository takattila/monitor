// Defined in: web/html/monitor.html
// - let ROUTE_SYSTEMCTL = "{{.RouteSystemCtl}}";
// - let ROUTE_POWER = "{{.RoutePower}}";
// - let ROUTE_TOGGLE = "{{.RouteToggle}}";
// - let ROUTE_LOGOUT = "{{.RouteLogout}}";
// - let ROUTE_API = "{{.RouteApi}}";
// - let INTERVAL_SECONDS = "{{.QuerySeconds}}";

function setCookie(cname, cvalue, exdays) {
    const d = new Date();
    d.setTime(d.getTime() + (exdays*24*60*60*1000));
    let expires = "expires=" + d.toUTCString();
    document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}

function getCookie(cname) {
    var name = cname + "=";
    var decodedCookie = decodeURIComponent(document.cookie);
    var ca = decodedCookie.split(';');
    for (var i = 0; i < ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
        }
    }
    return "";
}

function systemctl(action, service) {
    var params = {
        type: "POST",
        url: ROUTE_SYSTEMCTL.replace("{action}", action).replace("{service}", service),
        async: false
    };

    if (action == "start" | action == "stop" | action == "restart" | action == "enable" | action == "disable") {
        params.async = true;
        return $.ajax(params).responseText;
    }

    return $.ajax(params).responseText;
}

function power(action) {
    var params = {
        type: "POST",
        url: ROUTE_POWER.replace("{action}", action),
        async: true
    };

    return $.ajax(params).responseText;
}

function toggleStatus(section, status) {
    var params = {
        type: "GET",
        url: ROUTE_TOGGLE.replace("{section}", section).replace("{status}", status),
        async: true
    };

    return $.ajax(params).responseText;
}

function logout() {
    window.location.replace(ROUTE_LOGOUT);
}

function logoutIfSessionEnded() {
    if (!getCookie("session")) {
        logout();
    }
}

function confirmSystemCtlAction(action, service) {
    if (confirm('Are you sure you want to [ ' + action + ' ] the "' + service + '" service?')) {
        systemctl(action, service);
    }
}

function confirmPowerAction(action) {
    if (confirm('Are you sure you want to [ ' + action + ' ] the computer?')) {
        logout();
        power(action);
    }
}

function copyTableRows() {
    var table = document.getElementById("processTable");
    var rows = table.getElementsByTagName("tr");
    for (i = 0; i < rows.length; i++) {
        var currentRow = table.rows[i];
        var createClickHandler = 
            function(row) {
                return function() { 
					var cell = row.getElementsByTagName("td")[1];
					var content = cell.innerHTML;
					content = content.replaceAll("&nbsp;", "");
					content = content.replaceAll(/<\/?[^>]+(>|$)/gi, "");
					content = content.replace(/\s+/g, ' ').trim();

                    const element = document.createElement("textarea");
                    element.value = content;
                    document.body.appendChild(element)
                    element.select();

                    document.execCommand("copy");
                    // navigator.clipboard.writeText(element.value);
					
                    alert("Copied to clipboard:\n\n" + content);
                    document.body.removeChild(element);
				 };
            };

        currentRow.onclick = createClickHandler(currentRow);
    }
}

function monitor() {
    var cpuUsage = new CircleProgress('#percent_cpu_usage_circle', {
        max: 100,
        value: 0,
        textFormat: 'percent',
    });

    var promise = $.ajax({
        type: "GET",
        url: ROUTE_API.replace("{statistics}", "all")
    });

    promise.done(function(response) {
        var data = $.parseJSON(response);

        // Parse JSON only if has a specific field...
        if (data.processor_info) {
            // CPU section
            var procInfo = data.processor_info;

            // CPU usage container
            cpuUsage.max = procInfo.usage.total;
            cpuUsage.value = procInfo.usage.actual;

            // CPU load container
            $('#cpu_load_01_minute_avg').text(Math.round(procInfo.load.min_01*100)/100);
            $('#cpu_load_05_minute_avg').text(Math.round(procInfo.load.min_05*100)/100);
            $('#cpu_load_15_minute_avg').text(Math.round(procInfo.load.min_15*100)/100);

            // CPU temperature container
            var cpuTempHtml = `
            <div class="w3-container">
                <p class="w3-large"></p>
                <div class="w3-light-red w3-large">
                    <div 
                        class="w3-container w3-center w3-red w3-large"
                        style="width:` + procInfo.temp.percent + `%;"
                        id="percent_cpu_temp">
                        ` + procInfo.temp.percent + `°C
                    </div>
                </div>
                <p></p>
            </div>
            `;

            $('#cpu_temp_container').html(cpuTempHtml + '<p></p>');

            // CPU temperature container -> responsive height
            heightMemBlock = $('#memory_section').height();
            heightCpuTempBlock = $('#cpu_temp_container').height();
            heightCpuUsageBlock = $('#cpu_usage_container').height();
            heightCpuLoadBlock = $('#cpu_load_container').height();

            if (window.innerWidth < 600) {
                // Portrait
                heightCpuTempBlockNew = 150;
            } else {
                // Landscape
                heightCpuTempBlockNew = heightMemBlock - heightCpuUsageBlock - heightCpuLoadBlock - 32;
            }

            if (heightMemBlock < 100 & window.innerWidth > 600) {
                $('#vertical_progress_span').hide();
            } else {
                $('#vertical_progress_span').show();
            }

            vertProgHeightWrap = heightCpuTempBlockNew - 36;
            vertProgHeightMask = (((100 - procInfo.temp.percent) * vertProgHeightWrap) / 100);
            vertProgHeightSpan = vertProgHeightWrap / 2;

            $('#vertical_progress_container').css('height', heightCpuTempBlockNew + "px");
            $('#vertical_progress_wrapper').css('height', vertProgHeightWrap + "px");
            $('#vertical_progress_mask').css('height', vertProgHeightMask + "px");
            $('#vertical_progress_span').css('top', vertProgHeightSpan + "px");
            $('#vertical_progress_span').text(procInfo.temp.percent + "°C");

            // Memory section
            var memInfo = data.memory_info;
            var memoryHtml = '';

            for (var id in memInfo) {
                if (memInfo.hasOwnProperty(id)) {
                    var obj = memInfo[id];

                    memoryHtml += `
                    <p class="w3-large">
                        <span class="capitalize">` + id + `</span><br>
                        <span class="w3-medium">
                            [Actual] <b><span class="w3-text-green">` + obj.actual + " " + obj.actual_unit + `</span></b>
                            / 
                            [Total] <b><span class="w3-text-green">` + obj.total + " " + obj.total_unit + `</span></b>
                        </span>
                    </p>
                    <div class="w3-light-green w3-large w3-round">
                        <div
                            class="w3-container w3-center w3-large w3-green w3-round"
                            style="width:` + obj.percent + `%">
                            ` + obj.percent + `%
                        </div>
                    </div>
                    `;
                }
            }

            $('#memory_container').html(memoryHtml + '<p></p>');

            // Services section
            var servicesInfo = data.services_info;
            var servicesHtml = '';

            $.each(servicesInfo, function(service, status) {
                var serviceStatusClass = function(status) {
                    if (status != undefined) {
                        status = status.replace(/\r?\n|\r/g, "");
                    }
                    if (status === "active") {
                        return "w3-text-green";
                    }
                    return "w3-text-red";
                };

                var serviceEnabledBtnClass = function(status) {
                    if (status != undefined) {
                        status = status.replace(/\r?\n|\r/g, "");
                    }
                    if (status === "enabled") {
                        return "w3-green";
                    }
                    return "w3-red";
                };

                var serviceEnabledBtnAction = function(status) {
                    if (status != undefined) {
                        status = status.replace(/\r?\n|\r/g, "");
                    }
                    if (status === "enabled") {
                        return "disable";
                    }
                    return "enable";
                };

                var enabledBtnAction = serviceEnabledBtnAction(status.is_enabled);
                var enabledBtnClass = serviceEnabledBtnClass(status.is_enabled);

                servicesHtml += `
                <thead>
                    <tr>
                        <th class="service-td 3-large" colspan="3">
                            <span class="` + serviceStatusClass(status.is_active) + `">[ ` + status.is_active + ` ]</span> ` + service + `
                        </th>
                    </tr>
                </thead>
                <tr>
                    <td class="service-td"><button onclick="confirmSystemCtlAction('start', '` + service + `')" class="service-button w3-button w3-green round-left">start</button></td>
                    <td class="service-td"><button onclick="confirmSystemCtlAction('stop', '` + service + `')" class="service-button w3-button w3-red">stop</button></td>
                    <td class="service-td"><button onclick="confirmSystemCtlAction('restart', '` + service + `')" class="service-button w3-button w3-blue round-right">restart</button></td>
                </tr>
                <tr>
                    <td class="service-td 3-large" colspan="3">
                        <button onclick="confirmSystemCtlAction('` + enabledBtnAction + `', '` + service + `')" class="service-button w3-button ` + enabledBtnClass + ` round">[ ` + status.is_enabled + ` ] -> ` + enabledBtnAction + ` service</button>
                    </td>
                </tr>
                <tr>
                    <td class="w3-medium" colspan="3"> </td>
                </tr>
                `;
            });

            var servicessTable = `<table class="w3-table">` + servicesHtml + `</table>`;
            $('#services_container').html(servicessTable + '<p></p>');

            // Process section
            var processInfo = data.process_info;
            var processHtml = '';

            for (var id in processInfo) {
                if (processInfo.hasOwnProperty(id)) {
                    var obj = processInfo[id];

                    processHtml += `
                    <tr>
                        <td>
                            <h4 class="w3-light-gray round-left" style="padding-left: 13px;">` + id + `. </h4>
                            <b>PID</b>: <br>
                            <b>USER</b>: <br>
                            <b>MEM</b>: <br>
                            <b>CPU</b>: <br>
                            <b>CMD</b>:
                        </td>
                        <td class="word-wrap">
                            <h4 class="w3-light-gray round-right">&nbsp;</h4>
                            ` + obj.pid + ` <br>
                            ` + obj.user + ` <br>
                            ` + obj.mem + `% <br>
                            ` + obj.cpu + `% <br>
                            ` + obj.cmd + `
                        </td>
                    </tr>
                    `;
                }
            }

            var processTable = `<table class="w3-table" id="processTable" style="cursor:pointer">` + processHtml + `</table>`
            $('#process_container').html(processTable + '<p></p>');

            // Network Traffic section
            var networkInfo = data.network_info;
            var networkHtml = '';

            var trafficInArray = [];
            var trafficOutArray = [];

            for (var id in networkInfo) {
                if (networkInfo.hasOwnProperty(id)) {
                    var obj = networkInfo[id];
                    trafficInArray.push(obj.in);
                    trafficOutArray.push(obj.out);
                }
            }

            var maxInTraffic = Math.max.apply(null, trafficInArray)
            var maxOutTraffic = Math.max.apply(null, trafficOutArray)

            for (var id in networkInfo) {
                if (networkInfo.hasOwnProperty(id)) {
                    var obj = networkInfo[id];

                    inPercent = (obj.in == 0 ? 0 : obj.in * 100 / maxInTraffic);
                    inPercent = Number((inPercent).toFixed(1));

                    outPercent = (obj.out == 0 ? 0 : obj.out * 100 / maxOutTraffic);
                    outPercent = Number((outPercent).toFixed(1));

                    networkHtml += `
                    <p>
                        <b>[ ` + id + ` ]</b> <i class="fas fa-angle-double-left w3-text-blue"></i> <b>in</b>
                    </p>
                    <div class="color-light-blue w3-large w3-round">
                        <div class="w3-container w3-center w3-large w3-blue w3-round" style="width: ` + inPercent + `%">
                            ` + obj.in + `&nbsp;KB/s
                        </div>
                    </div>
                    <p>
                        <b>[ ` + id + ` ]</b> <i class="fas fa-angle-double-right color-text-dark-blue"></i> <b>out</b>
                    </p>
                    <div class="color-light-blue w3-large w3-round">
                        <div class="w3-container w3-center w3-large color-dark-blue w3-round" style="width: ` + outPercent + `%">
                            ` + obj.out + `&nbsp;KB/s
                        </div>
                    </div>
                    `;
                }
            }

            $('#network_container').html(networkHtml + '<p></p>');

            // Storage section
            var devInfo = data.storage_info;
            var storageHtml = '';

            for (var id in devInfo) {
                if (devInfo.hasOwnProperty(id)) {
                    var obj = devInfo[id];

                    storageHtml += `
                    <p class="w3-large">
                        ` + id + `<br>
                        <span class="w3-medium">
                            - [Used] <b><span class="color-text-light-blue">` + obj.actual + " " + obj.actual_unit + `</span></b> <br>
                            - [Total] <b><span class="color-text-light-blue">` + obj.total + " " + obj.total_unit + `</span></b> <br>
                            - [Free] <b><span class="color-text-light-blue">` + obj.free + " " + obj.free_unit + `</span></b>
                        </span>
                    </p>
                    <div class="color-light-blue w3-large w3-round">
                        <div 
                            class="w3-container w3-center w3-large color-dark-blue w3-round"
                            style="width:` + obj.percent + `%">
                            ` + obj.percent + `%
                        </div>
                    </div>
                    `;
                }
            }

            $('#storage_container').html(storageHtml + '<p></p>');

            // Footer section: write model name
            $('#model_name').text(data.model_name);

            window.onload = copyTableRows();
        }
    });
}

function toggleSection() {
    $('h2').on('click', function() {
        var id = $(this).attr('id');
        var container = "#" + id + "_container";
        if ($(this).attr('data-click-state') == 1) {
            $(this).attr('data-click-state', 0);
            $(container).hide(200);
            section = $(this).text().replace(/\s+/g, '').trim();
            toggleStatus(section, false);
        } else {
            $(this).attr('data-click-state', 1);
            $(container).show(200);
            monitor();
            section = $(this).text().replace(/\s+/g, '').trim();
            toggleStatus(section, true);
        }
    });
}

function toggleSectionCpu() {
    $('#cpu').on('click', function() {
        if (window.innerWidth > 600) {
            $('#memory').click();
        }
        if ($('#cpu').attr('data-click-state') == 1) {
            $('#cpu').attr('data-click-state', 1);
            $('#cpu_usage_wrapper').hide(200);
            $('#cpu_load_container').hide(200);
            $('#vertical_progress_container').hide(200);
        } else {
            $('#cpu').attr('data-click-state', 0);
            $('#cpu_usage_wrapper').show(200);
            $('#cpu_load_container').show(200);
            $('#vertical_progress_container').show(200);
        }
    });
}

function setLightSkin() {
    $('header').attr('data-click-state', 0);
    $('footer').attr('data-click-state', 0);
    $('.w3-dark').addClass('w3-white').removeClass('w3-dark');
    $('.w3-dark-grey').addClass('w3-light-grey').removeClass('w3-dark-grey');
    $('.w3-text-light-grey').addClass('w3-text-grey').removeClass('w3-text-light-grey');
    setCookie("skin", "light", 30);
    skin = getCookie("skin");
}

function setDarkSkin() {
    $('header').attr('data-click-state', 1);
    $('footer').attr('data-click-state', 1);
    $('.w3-white').addClass('w3-dark').removeClass('w3-white');
    $('.w3-light-grey').addClass('w3-dark-grey').removeClass('w3-light-grey');
    $('.w3-text-grey').addClass('w3-text-light-grey').removeClass('w3-text-grey');
    setCookie("skin", "dark", 30);
    skin = getCookie("skin");
}

function toggleThemeOnHeaderOrFooterClick() {
    $('header, footer').on('click', function() {
        if ($(this).attr('data-click-state') == 1) {
            setLightSkin();
        } else {
            setDarkSkin();
        }
    });
}

function applySkin() {
    skin = getCookie("skin");

    if (skin == "dark") {
        setDarkSkin();
    } else {
        setLightSkin();
    }
}

function collapseSectionsExceptCpu() {
    toggleStatus("Memory", true);
    if (window.innerWidth < 600) {
        $('#memory').click();
    }
    $('#services').click();
    $('#process').click();
    $('#network').click();
    $('#storage').click();
    $('#power').click();
    $('#logout').click();
}

var loop;

function start() {
    monitor();
    loop = setInterval(function() {
        logoutIfSessionEnded();
        monitor();
    }, INTERVAL_SECONDS * 1000);
    console.log("started setInterval");
}

function stop() {
    clearInterval(loop);
    console.log("stopped setInterval");
}

$(document).ready(function() {
    logoutIfSessionEnded();
    start();
    toggleSection();
    toggleSectionCpu();
    toggleThemeOnHeaderOrFooterClick();
    collapseSectionsExceptCpu();
    applySkin();
});
