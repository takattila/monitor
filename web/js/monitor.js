// Defined in: web/html/monitor.html
// - let ROUTE_SYSTEMCTL = "{{.RouteSystemCtl}}";
// - let ROUTE_POWER = "{{.RoutePower}}";
// - let ROUTE_TOGGLE = "{{.RouteToggle}}";
// - let ROUTE_LOGOUT = "{{.RouteLogout}}";
// - let ROUTE_API = "{{.RouteApi}}";
// - let INTERVAL_SECONDS = "{{.QuerySeconds}}";

var loop = null;
var stdoutLoop;
var header = document.getElementById("model_name");
var sticky = header.offsetTop;
var autoScroll = true;

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

function systemctl(params) {
    var args = params.split(",");
    var action = args[0];
    var service = args[1];

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

function logout() {
    window.location.replace(ROUTE_LOGOUT);
}

function power(action) {
    logout();
    var params = {
        type: "POST",
        url: ROUTE_POWER.replace("{action}", action),
        async: true
    };

    return $.ajax(params).responseText;
}

function kill(pid) {
    var params = {
        type: "POST",
        url: ROUTE_KILL.replace("{pid}", pid),
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

function logoutIfSessionEnded() {
    if (!getCookie("session")) {
        logout();
    }
}

function confirmSystemCtlAction(action, service) {
    dialog({
        id: "confirm", 
        title: "Confirm", 
        content: 'Are you sure you want to <b class="w3-red">[&nbsp;' + action + '&nbsp;]</b> the "' + service + '" service?', 
        cancelBtnText: "NO", 
        okFunc: systemctl, 
        okFuncParam: [action, service], 
        okBtnText: "YES"
    });
}

function confirmPowerAction(action) {
    dialog({
        id: "confirm", 
        title: "Confirm", 
        content: 'Are you sure you want to <b class="w3-red">[&nbsp;' + action + '&nbsp;]</b> the computer?', 
        cancelBtnText: "NO", 
        okFunc: power, 
        okFuncParam: action, 
        okBtnText: "YES"
    });
}

function dialog({
        id, 
        title, 
        content, 
        cancelFunc, 
        cancelFuncParam, 
        cancelBtnText, 
        okFunc, 
        okFuncParam, 
        okBtnText
    } = {}) {

    var infoModal = `
    <div id="dialog_` + id + `" class="w3-modal modal-open scroll-hidden">
        <div id="dialog_box_` + id + `" class="w3-modal-content w3-animate-left w3-white w3-card dialog-open">
            <header class="w3-container w3-red"> 
                <span onclick="dialogCancel({closeId: '` + id + `'})" class="w3-button w3-display-topright modal-header-close-font">&times;</span>
                <h2 id="dialog_header_` + id + `" data-click-state="1" class="modal-header-font">` + title + `</h2>
            </header>
            <div class="w3-container w3-margin-bottom w3-center w3-margin-left">

                <div id="dialog_loader_` + id + `" class="w3-display-middle w3-medium">
                    <i class="fa fa-spinner w3-spin" class="modal-loader-duration"></i> Loading data...
                </div>
                
                <div id="dialog_content_` + id + `" class="w3-medium custom-scrollbar modal-content-scroll">
                <p>`;

    infoModal += content.trim();

    infoModal += `
                </p>
                <p>
                  <table class="w3-table">
                    <tr>`;

    if (okFunc) {
        infoModal += `<td class="service-td"><button type="button" onclick="dialogOk({functionToExecute: ` + okFunc.name + `, funcParam: '` + okFuncParam + `', closeId: '` + id + `'})" class="service-button w3-button w3-green">` + okBtnText + `</button></td>`;
    }

    if (cancelFunc) {
        infoModal += `<td class="service-td"><button type="button" onclick="dialogCancel({functionToExecute: ` + cancelFunc.name + `, funcParam: '` + cancelFuncParam + `', closeId: '` + id + `'})" class="service-button w3-button w3-red">` + cancelBtnText + `</button></td>`;
    } else {
        infoModal += `<td class="service-td"><button type="button" onclick="dialogCancel({closeId: '` + id + `'})" class="service-button w3-button w3-red">` + cancelBtnText + `</button></td>`
    }

    infoModal += `
                    </tr>
                  </table>
                </p>
            </div>
          </div>
        </div>
    </div>`;

    $('#dialog_container').html(infoModal + '<p></p>');
    
    skin = getCookie("skin");
    
    if (skin == "light") {
        $('#dialog_box_'+ id).addClass('w3-white').removeClass('w3-dark');
    } else {
        $('#dialog_box_'+ id).addClass('w3-dark').removeClass('w3-white');
    }

    $('#dialog_' + id).css('display', "block");

    if ($('#dialog_header_' + id).attr('data-click-state') == 1) {
        $('#dialog_' + id).css("z-index", "9999999");
        $('#dialog_loader_' + id).css("display", "none");
        $('#dialog_content_' + id).css("display", "block");
        $('#dialog_content_' + id).css("max-height", ($(window).height() - 100) + "px");
        $('#dialog_header_' + id).attr('data-click-state', 0);
    } else {
        $('#dialog_header_' + id).attr('data-click-state', 1);
    }
}

function dialogOk({functionToExecute, funcParam, closeId} = {}) {
    if (functionToExecute) {
        functionToExecute(funcParam);
    }
    $('#dialog_' + closeId).remove();
}

function dialogCancel({functionToExecute, funcParam, closeId} = {}) {
    if (functionToExecute) {
        functionToExecute(funcParam);
    }
    $('#dialog_' + closeId).remove();
}

function killProcess(pid, cmd) {
    dialog({
        id: "confirm",
        title: "Confirm",
        content: 'Are you sure you want to kill the process?<br><br><b class="w3-red">PID:</b> [&nbsp;' + pid + '&nbsp;]<br><b class="w3-red">Command:</b> ' + cmd.substring(0, 50) + "...",
        cancelBtnText: "NO",
        okFunc: kill,
        okFuncParam: pid,
        okBtnText: "YES"
    });
}

function copyProcessContent(id) {
    content = $('#' + id).text();
    content = content.replaceAll("&nbsp;", "");
    content = content.replaceAll(/<\/?[^>]+(>|$)/gi, "");
    content = content.replace(/\s+/g, ' ').trim();

    const element = document.createElement("textarea");
    element.value = content;
    document.body.appendChild(element)
    element.select();
    document.execCommand("copy");

    dialog({
        id: "info",
        title: "Info",
        content: "Content copied to the clipboard!",
        cancelBtnText: "OK"
    });

    document.body.removeChild(element);
}

function tail(id) {
    if (autoScroll) {
        let window = $(id);
        if (window.length > 0) {
            const height = window.get(0).scrollHeight;
            window.animate({
                scrollTop: height + 20
            }, 100);
        }
    }
}

function startLoopStdout(id) {
    stdoutLoop = setInterval(function() {
        var stdout = $.ajax({
            type: "GET",
            url: ROUTE_RUN.replace("{action}", "stdout").replace("{name}", id),
            dataType: 'text',
            timeout: 500,
            cache: false,
            async: true
        });

        stdout.done(function(stdout_response) {
            if (stdout_response) {
                tail('#modal_content_' + id);
                $('#modal_loader_' + id).css("display", "none");
                $('#modal_content_' + id).css("height", ($('#modal_' + id).height() - 80) + "px");
                $('#modal_content_' + id).css("display", "block");
                if (stdout_response.indexOf('~x~o(f)o~x~') >= 0) {
                    stopLoopStdout();
                    autoScroll = false;
                }
                $('#modal_data_' + id).text(stdout_response.replace("~x~o(f)o~x~", "").split("\r").join("\n"));
            }
        });

    }, INTERVAL_SECONDS * 1000);
}

function stopLoopStdout() {
    clearInterval(stdoutLoop);
}

function confirmModalOpen(id) {
    dialog({
        id: "confirm", 
        title: "Confirm", 
        content: 'Are you sure you want to run the <span class="w3-red">[&nbsp;' + id + '&nbsp;]</span> command?', 
        cancelBtnText: "NO", 
        okFunc: modalOpen, 
        okFuncParam: id, 
        okBtnText: "YES"
    });
}

function modalOpen(id) {
    stop();

    autoScroll = true;

    $('#modal_' + id).css('display', "block");
    skin = getCookie("skin");
    if (skin == "light") {
        $('#modal_box_'+ id).addClass('w3-white').removeClass('w3-dark');
    } else {
        $('#modal_box_'+ id).addClass('w3-dark').removeClass('w3-white');
    }

    var run = $.ajax({
        type: "GET",
        url: ROUTE_RUN.replace("{action}", "exec").replace("{name}", id)
    });

    run.done(function() {
        startLoopStdout(id);
    });

    $('#modal_header_' + id).on('click', function() {
        if ($(this).attr('data-click-state') == 1) {
            stopLoopStdout();
            autoScroll = false;
            $(this).attr('data-click-state', 0);
        } else {
            startLoopStdout(id);
            autoScroll = true;
            $(this).attr('data-click-state', 1);
        }
    });
}

function copyContent(id) {
    var aux = document.createElement("div");

    aux.setAttribute("contentEditable", true);
    aux.innerHTML = document.getElementById('modal_content_' + id).innerHTML;
    aux.setAttribute("onfocus", "document.execCommand('selectAll',false,null)"); 
    document.body.appendChild(aux);
    aux.focus();
    document.execCommand("copy");
    document.body.removeChild(aux);

    dialog({
        id: "info", 
        title: "Info", 
        content: "Content copied to the clipboard!", 
        cancelBtnText: "OK"
    });
}

function modalClose(id) {
    start();

    $('#modal_' + id).css('display', "none");
    $('#modal_loader_' + id).css("display", "block");
    $('#modal_content_' + id).css("display", "none");

    stopLoopStdout();
}

function reload() {
    window.location.reload();
}

function loadLogoPng(logo) {
    var oldlink = $('#logo_png');
    var newlink = document.createElement("link");
    newlink.setAttribute("rel", "icon");
    newlink.setAttribute("type", "image/png");
    newlink.setAttribute("href", ROUTE_WEB + "/img/" + logo + ".png?v=" + VERSION);

    oldlink.replaceWith(newlink);
}

function loadLogoSvg(logo) {
    var img = ROUTE_WEB + "/img/" + logo + ".svg?v=" + VERSION
    $('#logo_svg').attr("src", img);
}

function setLogoToCookies(logo) {
    setCookie("logo", logo, 30);
    reload();
}

function loadLogoFromCookie() {
    logo = getCookie("logo");
    if (logo) {
        loadLogoPng(logo);
        loadLogoSvg(logo);
    }
}

function loadCSS(skin) {
    var oldlink = $('#css');
    var newlink = document.createElement("link");
    newlink.setAttribute("rel", "stylesheet");
    newlink.setAttribute("type", "text/css");
    newlink.setAttribute("href", "/monitor/web/css/" + skin + ".css?v=" + VERSION);

    oldlink.replaceWith(newlink);
}

function setCssToCookies(skin) {
    setCookie("css", skin, 30);
    reload();
}

function loadCssFromCookie() {
    css = getCookie("css");
    if (css) {
        loadCSS(css);
    } else {
        loadCSS("rpi");
    }
}

function toggleRun(id) {
    var container = "#" + id + "_container";
    var run = id + "_run"
    var toggle = getCookie(run);

    if (toggle == "1") {
        $(container).hide(0);
        setCookie(run, "0", 0.0003472222);
    } else {
        $(container).show(0);
        setCookie(run, "1", 0.0003472222);
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
            // Calculate the width of the Memory section dynamically...
            if (window.innerWidth > 600) {
                $("#cpu_section").css("max-width", "520px")
                cpuSectionWidth = $('#cpu_section').width();
                fullWidth = $('#page_container').width();;
                newMemorySectionWidth = fullWidth - cpuSectionWidth - 32;
                $('#memory_section').css('width', newMemorySectionWidth + "px");
            } else {
                $('#memory_section').css('width', "");
            }

            // Header section: write model name
            $('#model_name').text(data.model_name);

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
                heightCpuTempBlockNew = heightMemBlock - heightCpuUsageBlock - heightCpuLoadBlock - 53;
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
                        <td id="` + obj.pid + `_kill" onclick="killProcess('` + obj.pid + `', '` + obj.cmd.replaceAll("'","") + `')">
                            <h4 class="w3-light-gray round-left process-padding-left w3-red">&times;</h4>
                            <b>PID:</b> <br>
                            <b class="w3-text-red">USER:</b> <br>
                            <b>MEM:</b> <br>
                            <b class="w3-text-red">CPU:</b> <br>
                            <b>CMD:</b>
                        </td>
                        <td id="` + obj.pid + `_content" class="word-wrap" onclick="copyProcessContent('` + obj.pid + `_content')">
                            <h4 class="w3-light-gray round-right w3-green">&nbsp;` + id + `.</h4>
                            ` + obj.pid + ` <br>
                            <span class="w3-text-red">` + obj.user + ` </span><br>
                            ` + obj.mem + `% </span><br>
                            <span class="w3-text-red">` + obj.cpu + `% </span><br>
                            ` + obj.cmd + `
                        </td>
                    </tr>
                    `;
                }
            }

            var processTable = `<table class="w3-table cursor-hand" id="processTable">` + processHtml + `</table>`
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

            // Run section
            var runList = data.run_list;
            var runModal = '';
            var runHtml = '';

            for (var id in runList) {
                if (runList.hasOwnProperty(id)) {
                    var obj = runList[id];
                    var toggle = getCookie(id + '_run');
                    var style = ""

                    if (toggle != "1") {
                        style = `style="display: none"`;
                    }

                    runHtml += `<h3 id="` + id + `" onclick="toggleRun('` + id + `')" class="cursor-hand">`;
                    runHtml += `<i class="fa fa-terminal fa-fw w3-margin-right"></i>` + id;
                    runHtml += `</h3>`;

                    runHtml += `<div id="` + id + `_container" ` + style + `>`;

                    runHtml += `<pre class="w3-medium w3-card w3-panel w3-padding-16 run-list-pre" >`;
                    runHtml += obj.trim()
                    runHtml += `</pre>`;

                    runHtml +=`<button onclick="confirmModalOpen('` + id + `');" class="service-button w3-button w3-red round-left">run</button>`;
                    runHtml += `<br><br>`;

                    runHtml += `</div>`;

                    runModal += `
                    <div id="modal_` + id + `" class="w3-modal modal-open scroll-hidden">
                        <div id="modal_box_` + id + `" class="w3-modal-content w3-animate-top w3-white w3-card modal-ninetynine">
                            <header class="w3-container w3-red"> 
                                <span onclick="modalClose('` + id + `')" class="w3-button w3-display-topright modal-header-close-font">&times;</span>
                                <h2 id="modal_header_` + id + `" data-click-state="1" class="modal-header-font">Running: "` + id + `"</h2>
                            </header>
                            <div class="w3-container w3-margin-bottom">

                                <div id="modal_loader_` + id + `" class="w3-display-middle w3-medium">
                                    <i class="fa fa-spinner w3-spin" class="modal-loader-duration"></i> Loading data...
                                </div>
                                
                                <div id="modal_content_` + id + `" class="w3-medium custom-scrollbar modal-content-scroll">
                                    <pre id="modal_data_` + id + `" class="w3-medium w3-panel w3-padding-16" ondblclick="copyContent('` + id + `')">
                                        -= CONTENT =-
                                    </pre>
                                </div>

                            </div>
                        </div>
                    </div>
                    `;
                }
            }

            $('#modal_container').html(runModal + '<p></p>');
            $('#run_container').html(runHtml + '<p></p>');

            // Settings section
            var skinHtml = '<div class="w3-row-padding w3-margin-bottom"><h3>Set skin</h3>';
            var skins = data.skins;

            for (let i = 0; i < skins.length; i++) {
                skinHtml += `
                <div class="w3-half w3-card w3-padding w3-margin-top cursor-hand" onclick="setCssToCookies('` + skins[i] + `');">
                    <i class="fa fa-angle-right"></i> ` + skins[i] + `
                </div>
                `;
            }

            skinHtml += `</div>`
            
            $('#settings_container').html(skinHtml);

            var logoHtml = '<div class="w3-row-padding w3-margin-bottom"><h3>Set logo</h3>';
            var logos = data.logos;

            for (let i = 0; i < logos.length; i++) {
                logoHtml += `
                <div class="w3-half w3-card w3-padding w3-margin-top cursor-hand" onclick="setLogoToCookies('` + logos[i] + `');">
                    <i class="fa fa-angle-right"></i> ` + logos[i] + `
                </div>
                `;
            }

            logoHtml += `</div>`

            var settingsHtml = skinHtml+logoHtml
            $('#settings_container').html(settingsHtml);

            // Uptime section
            $('#uptime_info').text(data.uptime_info);
        }
    });
}

function toggleSection() {
    $('h2').on('click', function() {
        var id = $(this).attr('id');
        var container = "#" + id + "_container";
        if ($(this).attr('data-click-state') == 1) {
            $(this).attr('data-click-state', 0);
            $(container).slideUp(200);
            section = $(this).text().replace(/\s+/g, '').trim();
            toggleStatus(section, false);
        } else {
            $(this).attr('data-click-state', 1);
            $('#' + id + '_loader').css("margin-top", "-32px").fadeIn(200, function() {
                $(container).fadeIn(1000);
            }).fadeOut(600);
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
            $('#cpu_usage_wrapper').fadeIn(200);
            $('#cpu_load_container').fadeIn(200);
            $('#vertical_progress_container').fadeIn(200);
        }
    });
}

function toggleSectionMemory() {
    let memoryVisible = false
    let cpuVisible = false

    $('#memory').on('click', function() {
        if (window.innerWidth > 600) {
            if ($('#memory').attr('data-click-state') == 1) {
                $('#memory').attr('data-click-state', 1);
                memoryVisible = true
            } else {
                $('#memory').attr('data-click-state', 0);
                memoryVisible = false
            }

            if ($('#cpu').attr('data-click-state') == 1) {
                cpuVisible = false
            } else {
                cpuVisible = true
            }

            if (!memoryVisible && cpuVisible) {
                $('#cpu').attr('data-click-state', 1);
                $('#cpu_usage_wrapper').fadeOut(500);
                $('#cpu_load_container').fadeOut(500);
                $('#vertical_progress_container').fadeOut(500);
            }

            if (memoryVisible && !cpuVisible) {
                $('#cpu').attr('data-click-state', 0);
                $('#cpu_usage_wrapper').fadeIn(500);
                $('#cpu_load_container').fadeIn(500);
                $('#vertical_progress_container').fadeIn(500);
            }
        }
    });
}

function setLightSkin() {
    $('header').attr('data-click-state', 0);
    $('footer').attr('data-click-state', 0);
    $('#model_name').attr('data-click-state', 0);
    $('.w3-dark').addClass('w3-white').removeClass('w3-dark');
    $('.w3-dark-grey').addClass('w3-light-grey').removeClass('w3-dark-grey');
    $('.w3-text-light-grey').addClass('w3-text-grey').removeClass('w3-text-light-grey');
    setCookie("skin", "light", 30);
    skin = getCookie("skin");
}

function setDarkSkin() {
    $('header').attr('data-click-state', 1);
    $('footer').attr('data-click-state', 1);
    $('#model_name').attr('data-click-state', 1);
    $('.w3-white').addClass('w3-dark').removeClass('w3-white');
    $('.w3-light-grey').addClass('w3-dark-grey').removeClass('w3-light-grey');
    $('.w3-text-grey').addClass('w3-text-light-grey').removeClass('w3-text-grey');
    setCookie("skin", "dark", 30);
    skin = getCookie("skin");
}

function toggleThemeOnHeaderOrFooterClick() {
    $('header, footer, #model_name').on('click', function() {
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
    $('#run').click();
    $('#settings').click();
    $('#power').click();
    $('#logout').click();
}

function sticyHeader() {
    if (window.pageYOffset > sticky) {
        header.classList.add("sticky");
    } else {
        header.classList.remove("sticky");
    }
}

function loader() {
    $("body").fadeIn(800);
}

function start() {
    if (!loop) {
        monitor();
        loop = setInterval(function() {
            logoutIfSessionEnded();
            monitor();
        }, INTERVAL_SECONDS * 1000);
        console.log("started setInterval");
    }
}

function stop() {
    clearInterval(loop);
    loop = null;
    console.log("stopped setInterval");
}

window.onscroll = function() { sticyHeader() };

$(document).ready(function() {
    loader();
    logoutIfSessionEnded();
    start();
    applySkin();
    loadCssFromCookie();
    loadLogoFromCookie();
    toggleSection();
    toggleSectionCpu();
    toggleSectionMemory();
    toggleThemeOnHeaderOrFooterClick();
    collapseSectionsExceptCpu();
});
