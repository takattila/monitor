let loop = null;
let stdoutLoop;
let autoScroll = true;
let networkHistory = {};
const NETWORK_HISTORY_POINTS = 60;

function setCookie(cname, cvalue, exdays) {
    const d = new Date();
    if (exdays === Infinity) {
        d.setFullYear(9999, 12, 31);
    } else {
        d.setTime(d.getTime() + (exdays*24*60*60*1000));
    }
    let expires = "expires=" + d.toUTCString();
    let path = "path=" + ROUTE_INDEX + "/";
    let cookie = cname + "=" + cvalue + ";" + expires + ";" + path;
    document.cookie = cookie;
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
    $.ajax({
        type: "GET",
        url: ROUTE_SETTINGS,
        timeout: 5000,
        success: function(response) {
            try {
                var settings = $.parseJSON(response);
                if (!settings || typeof settings !== "object") {
                    logout();
                }
            } catch(e) {
                logout();
            }
        },
        error: function() {
            logout();
        }
    });
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

    stop();

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

function isAnyModalOpen() {
    var modalIsOpen = false
    var dialog = $('#dialog_container').text().trim();
    if (dialog) {
        modalIsOpen = true;
    }
    $('#modal_container').children('.w3-modal').each(function (index, item) {
        var closeId = $(item).attr("id");
        var display = $('#' + closeId).css("display");
        if (display == "block") {
            modalIsOpen = true;
        }
    });
    return modalIsOpen;
}

function dialogOk({functionToExecute, funcParam, closeId} = {}) {
    if (functionToExecute) {
        functionToExecute(funcParam);
    }
    var $box = $('#dialog_box_' + closeId);
    $box.addClass('animateleft-out').one('animationend', function() {
        $('#dialog_' + closeId).remove();
        if (isAnyModalOpen() == false) {
            start();
        }
    });
}

function dialogCancel({functionToExecute, funcParam, closeId} = {}) {
    if (functionToExecute) {
        functionToExecute(funcParam);
    }
    var $box = $('#dialog_box_' + closeId);
    $box.addClass('animateleft-out').one('animationend', function() {
        $('#dialog_' + closeId).remove();
        if (isAnyModalOpen() == false) {
            start();
        }
    });
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
    var $box = $('#modal_box_' + id);
    $box.addClass('animatetop-out').one('animationend', function() {
        $('#modal_' + id).css('display', "none");
        $('#modal_loader_' + id).css("display", "block");
        $('#modal_content_' + id).css("display", "none");
        stopLoopStdout();
        start();
    });
}

function setProgressPreset(preset) {
    $.ajax({
        type: "POST",
        url: ROUTE_SETTINGS,
        data: JSON.stringify({key: "preset", value: preset}),
        contentType: "application/json",
        success: function() {
            $('body').removeClass(function(_, cls) {
                return (cls.match(/(^|\s)preset-\S+/g) || []).join(' ');
            });
            document.body.classList.add('preset-' + preset);
        }
    });
}

function loadProgressPreset(preset) {
    if (!preset) {
        preset = "block";
    }
    $('body').removeClass(function(_, cls) {
        return (cls.match(/(^|\s)preset-\S+/g) || []).join(' ');
    });
    document.body.classList.add('preset-' + preset);
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

function setLogo(logo) {
    $.ajax({
        type: "POST",
        url: ROUTE_SETTINGS,
        data: JSON.stringify({key: "logo", value: logo}),
        contentType: "application/json",
        success: function() {
            loadLogoPng(logo);
            loadLogoSvg(logo);
        }
    });
}

function loadLogo(logo) {
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
    newlink.setAttribute("href", ROUTE_WEB + "/css/" + skin + ".css?v=" + VERSION);

    oldlink.replaceWith(newlink);
}

function setCss(skin) {
    $.ajax({
        type: "POST",
        url: ROUTE_SETTINGS,
        data: JSON.stringify({key: "css", value: skin}),
        contentType: "application/json",
        success: function() {
            loadCSS(skin);
        }
    });
}

function loadCss(css) {
    if (css) {
        loadCSS(css);
    } else {
        loadCSS("github_red");
    }
}

function toggleSubSection(id) {
    var container = "#" + id + "_container";
    var sub = id + "_sub"
    var toggle = getCookie(sub);

    if (toggle == "1") {
        $(container).slideUp(200);
        setCookie(sub, "0", 0.0003472222);
    } else {
        $(container).slideDown(200);
        setCookie(sub, "1", 0.0003472222);
    }
}

function getNetworkColors() {
    var inEl = document.querySelector('#network_container .net-arrow-in');
    var outEl = document.querySelector('#network_container .net-arrow-out');
    var cardEl = document.querySelector('#network_container');
    var card = cardEl ? cardEl.parentElement : null;

    var inColor = inEl ? getComputedStyle(inEl).color : '#00e5ff';
    var outColor = outEl ? getComputedStyle(outEl).color : '#ff9800';

    var light = false;
    if (card) {
        light = isLightColor(getComputedStyle(card).backgroundColor);
    }

    return {
        in: inColor,
        out: outColor,
        grid: light ? 'rgba(0,0,0,0.12)' : 'rgba(255,255,255,0.10)'
    };
}

function isLightColor(color) {
    var m = color.match(/rgba?\((\d+)[,\s]+(\d+)[,\s]+(\d+)/);
    if (!m) {
        return false;
    }
    var r = Number(m[1]);
    var g = Number(m[2]);
    var b = Number(m[3]);
    return (0.299 * r + 0.587 * g + 0.114 * b) > 128;
}

function drawNetworkChart(canvas, hist) {
    if (!hist || hist.in.length < 2) {
        return;
    }

    var rect = canvas.getBoundingClientRect();
    if (rect.width === 0) {
        return;
    }

    var dpr = window.devicePixelRatio || 1;
    canvas.width = Math.round(rect.width * dpr);
    canvas.height = Math.round(rect.height * dpr);

    var ctx = canvas.getContext('2d');
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    var colors = getNetworkColors();
    var width = rect.width;
    var height = rect.height;

    ctx.clearRect(0, 0, width, height);

    ctx.strokeStyle = colors.grid;
    ctx.lineWidth = 1;
    for (var g = 1; g <= 3; g++) {
        var gy = height - (height * g / 4);
        ctx.beginPath();
        ctx.moveTo(0, gy);
        ctx.lineTo(width, gy);
        ctx.stroke();
    }

    var maxVal = 1;
    for (var i = 0; i < hist.in.length; i++) {
        if (hist.in[i] > maxVal) {
            maxVal = hist.in[i];
        }
        if (hist.out[i] > maxVal) {
            maxVal = hist.out[i];
        }
    }

    drawNetworkSeries(ctx, hist.in, colors.in, width, height, maxVal);
    drawNetworkSeries(ctx, hist.out, colors.out, width, height, maxVal);
}

function traceSmoothPath(ctx, points) {
    for (var i = 0; i < points.length - 1; i++) {
        var p0 = points[i - 1] || points[i];
        var p1 = points[i];
        var p2 = points[i + 1];
        var p3 = points[i + 2] || p2;

        var cp1x = p1.x + (p2.x - p0.x) / 6;
        var cp1y = p1.y + (p2.y - p0.y) / 6;
        var cp2x = p2.x - (p3.x - p1.x) / 6;
        var cp2y = p2.y - (p3.y - p1.y) / 6;

        ctx.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, p2.x, p2.y);
    }
}

function drawNetworkSeries(ctx, values, color, width, height, maxVal) {
    var n = values.length;
    if (n < 2) {
        return;
    }

    var step = width / (NETWORK_HISTORY_POINTS - 1);
    var startX = width - (n - 1) * step;

    var points = [];
    for (var i = 0; i < n; i++) {
        points.push({
            x: startX + i * step,
            y: height - (values[i] / maxVal) * (height - 4) - 2
        });
    }

    ctx.beginPath();
    ctx.moveTo(points[0].x, height);
    ctx.lineTo(points[0].x, points[0].y);
    traceSmoothPath(ctx, points);
    ctx.lineTo(points[n - 1].x, height);
    ctx.closePath();
    ctx.globalAlpha = 0.15;
    ctx.fillStyle = color;
    ctx.fill();
    ctx.globalAlpha = 1;

    ctx.beginPath();
    ctx.moveTo(points[0].x, points[0].y);
    traceSmoothPath(ctx, points);
    ctx.strokeStyle = color;
    ctx.lineWidth = 1.5;
    ctx.stroke();
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
            var networkIds = [];

            for (var id in networkInfo) {
                if (networkInfo.hasOwnProperty(id)) {
                    var obj = networkInfo[id];
                    var inVal = Number(obj.in) || 0;
                    var outVal = Number(obj.out) || 0;

                    if (!networkHistory[id]) {
                        networkHistory[id] = { in: [], out: [] };
                    }
                    var hist = networkHistory[id];
                    hist.in.push(inVal);
                    hist.out.push(outVal);
                    if (hist.in.length > NETWORK_HISTORY_POINTS) {
                        hist.in.shift();
                        hist.out.shift();
                    }

                    networkIds.push(id);
                    networkHtml += `
                    <p>
                        <b>[ ` + id + ` ]</b>
                        <i class="fas fa-angle-double-left w3-text-blue net-arrow-in"></i> <b>in</b>
                        <span class="w3-text-blue">` + inVal.toFixed(2) + `&nbsp;KB/s</span>
                        &nbsp;&nbsp;
                        <i class="fas fa-angle-double-right color-text-dark-blue net-arrow-out"></i> <b>out</b>
                        <span class="color-text-dark-blue">` + outVal.toFixed(2) + `&nbsp;KB/s</span>
                    </p>
                    <canvas class="network-chart"></canvas>
                    `;
                }
            }

            // Remove the history of interfaces that no longer exist.
            for (var h in networkHistory) {
                if (!networkInfo.hasOwnProperty(h)) {
                    delete networkHistory[h];
                }
            }

            $('#network_container').html(networkHtml + '<p></p>');

            var chartIndex = 0;
            $('#network_container .network-chart').each(function() {
                drawNetworkChart(this, networkHistory[networkIds[chartIndex]]);
                chartIndex++;
            });

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
                    var toggle = getCookie(id + '_sub');
                    var style = ""

                    if (toggle != "1") {
                        style = `style="display: none"`;
                    }

                    runHtml += `<h3 id="` + id + `" onclick="toggleSubSection('` + id + `')" class="cursor-hand">`;
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
            var skinHtml = '';
            var skins = data.skins;
            var toggleSkin = getCookie('set_skin_sub');
            var styleSkin = `style="display: block"`;

            if (toggleSkin != "1") {
                styleSkin = `style="display: none"`;
            }

            skinHtml += `<div id="set_skin" onclick="toggleSubSection('set_skin')" class="w3-card w3-padding cursor-hand w3-margin-bottom">`;
            skinHtml += '<h3><i class="fa fa-wrench fa-fw w3-margin-right"></i> Skin</h3>';

            skinHtml += '<div id="set_skin_container" class="w3-row-padding" ' + styleSkin + '>';

            for (let i = 0; i < skins.length; i++) {
                skinHtml += `
                <div class="w3-half w3-card w3-padding w3-margin-bottom cursor-hand" onclick="setCss('` + skins[i] + `');">
                <i class="fa fa-angle-right"></i> ` + skins[i] + `
                </div>
                `;
            }

            skinHtml += '</div>';
            skinHtml += '</div>';

            var logoHtml = '';
            var logos = data.logos;
            var toggleLogo = getCookie('set_logo_sub');
            var styleLogo = `style="display: block"`;

            if (toggleLogo != "1") {
                styleLogo = `style="display: none"`;
            }

            logoHtml += `<div id="set_logo" onclick="toggleSubSection('set_logo')" class="w3-card w3-padding cursor-hand w3-margin-bottom">`;
            logoHtml += '<h3><i class="fa fa-wrench fa-fw w3-margin-right"></i> Logo</h3>';

            logoHtml += '<div id="set_logo_container" class="w3-row-padding" ' + styleLogo + '>';

            for (let i = 0; i < logos.length; i++) {
                logoHtml += `
                <div class="w3-half w3-card w3-padding w3-margin-bottom cursor-hand" onclick="setLogo('` + logos[i] + `');">
                <i class="fa fa-angle-right"></i> ` + logos[i] + `
                </div>
                `;
            }

            logoHtml += '</div>';
            logoHtml += `</div>`;

            var progressHtml = '';
            var presets = [
                ['thin', 'Thin'],
                ['dashed', 'Dashed'],
                ['rounded', 'Rounded'],
                ['jumbo', 'Jumbo'],
                ['elegant', 'Elegant'],
                ['block', 'Block'],
                ['dotted', 'Dotted'],
                ['classic', 'Classic'],
                ['pill', 'Pill'],
                ['hairline', 'Hairline'],
                ['mesh', 'Mesh'],
                ['bold', 'Bold']
            ];
            var toggleProgress = getCookie('set_progress_sub');
            var styleProgress = `style="display: block"`;

            if (toggleProgress != "1") {
                styleProgress = `style="display: none"`;
            }

            progressHtml += `<div id="set_progress" onclick="toggleSubSection('set_progress')" class="w3-card w3-padding cursor-hand w3-margin-bottom">`;
            progressHtml += '<h3><i class="fa fa-wrench fa-fw w3-margin-right"></i> Progress</h3>';

            progressHtml += '<div id="set_progress_container" class="w3-row-padding" ' + styleProgress + '>';

            for (let i = 0; i < presets.length; i++) {
                progressHtml += `
                <div class="w3-half w3-card w3-padding w3-margin-bottom cursor-hand" onclick="setProgressPreset('` + presets[i][0] + `');">
                <i class="fa fa-angle-right"></i> ` + presets[i][1] + `
                </div>
                `;
            }

            progressHtml += '</div>';
            progressHtml += '</div>';

            var settingsHtml = skinHtml + logoHtml + progressHtml;
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

let terminalSession = null;
let terminalWS = null;
let terminalResizeTimer = null;

function openTerminal() {
    var container = document.getElementById('terminal_container');
    if (terminalSession) {
        return;
    }
    if (typeof Terminal === "undefined") {
        container.innerHTML = '<p class="w3-text-red">xterm.js is not loaded.</p>';
        return;
    }
    var div = document.createElement('div');
    div.id = 'terminal_box';
    container.innerHTML = '';
    container.appendChild(div);

    terminalSession = new Terminal({
        convertEol: true,
        cursorBlink: true,
        fontSize: 14,
        fontFamily: '"Anonymice NF", "Anonymous Pro for Powerline", monospace',
        theme: { background: '#000000' }
    });
    var fitAddon = new FitAddon.FitAddon();
    terminalSession.loadAddon(fitAddon);
    terminalSession.open(div);

    var terminalOpened = false;

    var openWebSocket = function() {
        var protocol = location.protocol === 'https:' ? 'wss://' : 'ws://';
        var url = protocol + location.host + ROUTE_TERMINAL + '?cols=' + terminalSession.cols + '&rows=' + terminalSession.rows;
        terminalWS = new WebSocket(url);

        terminalWS.onopen = function() {
            fitAddon.fit();
            terminalWS.send(JSON.stringify({type: "resize", cols: terminalSession.cols, rows: terminalSession.rows}));
            terminalSession.focus();
        };

        terminalWS.onmessage = function(event) {
            if (event.data instanceof Blob) {
                event.data.arrayBuffer().then(function(buffer) {
                    terminalSession.write(new Uint8Array(buffer));
                });
            } else {
                terminalSession.write(event.data);
            }
        };

        terminalWS.onclose = function() {
            closeTerminal(true);
        };
    };

    (function waitForVisible() {
        if (terminalOpened || !terminalSession || !div.isConnected) {
            return;
        }
        if (div.offsetWidth > 0 || div.offsetHeight > 0) {
            terminalOpened = true;
            fitAddon.fit();
            openWebSocket();
        } else {
            setTimeout(waitForVisible, 50);
        }
    })();

    terminalSession.onData(function(data) {
        if (terminalWS && terminalWS.readyState === WebSocket.OPEN) {
            terminalWS.send(JSON.stringify({type: "input", data: data}));
        }
    });

    $(window).on('resize.terminal', function() {
        if (terminalSession && terminalWS) {
            clearTimeout(terminalResizeTimer);
            terminalResizeTimer = setTimeout(function() {
                fitAddon.fit();
                if (terminalWS.readyState === WebSocket.OPEN) {
                    terminalWS.send(JSON.stringify({type: "resize", cols: terminalSession.cols, rows: terminalSession.rows}));
                }
            }, 150);
        }
    });
}

function closeTerminal() {
    clearTimeout(terminalResizeTimer);
    if (terminalWS) {
        terminalWS.onclose = null;
        terminalWS.close();
        terminalWS = null;
    }
    if (terminalSession) {
        terminalSession.dispose();
        terminalSession = null;
    }
    $('#terminal_box').remove();
    $(window).off('resize.terminal');
}

function toggleTerminal() {
    $('#terminal').on('click', function() {
        if ($(this).attr('data-click-state') == 1) {
            openTerminal();
        } else {
            closeTerminal();
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

function setLightSkin(save) {
    $('header').attr('data-click-state', 0);
    $('footer').attr('data-click-state', 0);
    $('#model_name').attr('data-click-state', 0);
    $('.w3-dark').addClass('w3-white').removeClass('w3-dark');
    $('.w3-dark-grey').addClass('w3-light-grey').removeClass('w3-dark-grey');
    $('.w3-text-light-grey').addClass('w3-text-grey').removeClass('w3-text-light-grey');
    $('body').addClass('light-mode');
    skin = "light";
    if (save) {
        $.ajax({
            type: "POST",
            url: ROUTE_SETTINGS,
            data: JSON.stringify({key: "skin", value: "light"}),
            contentType: "application/json"
        });
    }
}

function setDarkSkin(save) {
    $('header').attr('data-click-state', 1);
    $('footer').attr('data-click-state', 1);
    $('#model_name').attr('data-click-state', 1);
    $('.w3-white').addClass('w3-dark').removeClass('w3-white');
    $('.w3-light-grey').addClass('w3-dark-grey').removeClass('w3-light-grey');
    $('.w3-text-grey').addClass('w3-text-light-grey').removeClass('w3-text-grey');
    $('body').removeClass('light-mode');
    skin = "dark";
    if (save) {
        $.ajax({
            type: "POST",
            url: ROUTE_SETTINGS,
            data: JSON.stringify({key: "skin", value: "dark"}),
            contentType: "application/json"
        });
    }
}

function toggleThemeOnHeaderOrFooterClick() {
    $('header, footer, #model_name').on('click', function() {
        if ($(this).attr('data-click-state') == 1) {
            setLightSkin(true);
        } else {
            setDarkSkin(true);
        }
    });
}

function loadSettings() {
    $.ajax({
        type: "GET",
        url: ROUTE_SETTINGS,
        timeout: 5000,
        success: function(response) {
            var settings = $.parseJSON(response);

            skin = settings.skin || "dark";
            if (skin == "dark") {
                setDarkSkin(false);
            } else {
                setLightSkin(false);
            }

            loadCss(settings.css);
            loadLogo(settings.logo);
            loadProgressPreset(settings.preset);
        },
        error: function() {
            skin = "dark";
            setDarkSkin(false);
            loadCss("");
            loadLogo("");
            loadProgressPreset("");
        },
        complete: function() {
            start();
        }
    });
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
    $('#terminal').click();
    $('#settings').click();
    $('#power').click();
    $('#logout').click();
}

function loader() {
    $("body").animate({opacity: 1}, 800);
}

var stickyOffset = null;

function sticyHeader() {
    var mn = document.getElementById('model_name');
    if (mn) {
        if (stickyOffset === null) {
            stickyOffset = mn.offsetTop;
        }
        if (window.pageYOffset > stickyOffset) {
            mn.classList.add("sticky");
        } else {
            mn.classList.remove("sticky");
        }
    }
}

window.addEventListener('scroll', sticyHeader);

function enkerKeyPressed() {
    $(document).on('keydown', function(event) {
        if (event.key == "Enter") {
            $("#dialog_container").find('button').each(function(index, item) {
                var dialog = $('#dialog_container').text().trim();
                if (dialog) {
                    var closeId = $('#dialog_container :first-child').attr('id');
                    $('#' + closeId).remove();
                }
                var text = $(item).text();
                if (text == "YES") {
                    var func = $(item).attr("onclick");
                    eval(func);
                }
            });

            if (isAnyModalOpen() == false) {
                start();
            }
        }
    });
}

function escKeyPressed() {
    $(document).on('keydown', function(event) {
        if (event.key == "Escape") {
            var dialog = $('#dialog_container').text().trim();
            if (dialog) {
                var closeId = $('#dialog_container :first-child').attr('id');
                $('#' + closeId).remove();
            }

            $('#modal_container').children('.w3-modal').each(function (index, item) {
                var closeId = $(item).attr("id"); display = $('#' + closeId).css("display");
                if (display == "block") {
                    modalClose(closeId);
                }
            });

            if (isAnyModalOpen() == false) {
                start();
            }
        }
    });
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

$(document).ready(function() {
    loader();
    logoutIfSessionEnded();
    loadSettings();
    toggleSection();
    toggleTerminal();
    toggleSectionCpu();
    toggleSectionMemory();
    toggleThemeOnHeaderOrFooterClick();
    collapseSectionsExceptCpu();
    enkerKeyPressed();
    escKeyPressed();
});
