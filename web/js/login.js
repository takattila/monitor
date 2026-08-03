function setCookie(cname, cvalue, exdays) {
    const d = new Date();
    if (exdays === Infinity) {
        d.setFullYear(9999, 11, 31);
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

function hexToRgb(hex) {
    hex = hex.replace('#', '');
    if (hex.length === 3) {
        hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
    }
    var num = parseInt(hex, 16);
    return { r: (num >> 16) & 255, g: (num >> 8) & 255, b: num & 255 };
}

function rgbToHsl(c) {
    var r = c.r / 255, g = c.g / 255, b = c.b / 255;
    var max = Math.max(r, g, b), min = Math.min(r, g, b);
    var h = 0, s = 0, l = (max + min) / 2;
    if (max !== min) {
        var d = max - min;
        s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
        if (max === r) h = (g - b) / d + (g < b ? 6 : 0);
        else if (max === g) h = (b - r) / d + 2;
        else h = (r - g) / d + 4;
        h /= 6;
    }
    return [h * 360, s, l];
}

function hslToRgb(h, s, l) {
    h /= 360;
    var r, g, b;
    if (s === 0) {
        r = g = b = l;
    } else {
        var hue2rgb = function(p, q, t) {
            if (t < 0) t += 1;
            if (t > 1) t -= 1;
            if (t < 1 / 6) return p + (q - p) * 6 * t;
            if (t < 1 / 2) return q;
            if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
            return p;
        };
        var q = l < 0.5 ? l * (1 + s) : l + s - l * s;
        var p = 2 * l - q;
        r = hue2rgb(p, q, h + 1 / 3);
        g = hue2rgb(p, q, h);
        b = hue2rgb(p, q, h - 1 / 3);
    }
    return { r: Math.round(r * 255), g: Math.round(g * 255), b: Math.round(b * 255) };
}

function rgbToHex(c) {
    var to2 = function(v) { return ('0' + v.toString(16)).slice(-2); };
    return '#' + to2(c.r) + to2(c.g) + to2(c.b);
}

function warmOrCoolPastel() {
    var NEUTRAL = '#e8d6b2';
    var accent = getComputedStyle(document.documentElement).getPropertyValue('--cp-accent').trim();
    if (accent === '') {
        accent = '#666666';
    }
    var hsl = rgbToHsl(hexToRgb(accent));
    var hue = hsl[0], sat = hsl[1];

    if (sat < 0.08 || (hue >= 70 && hue < 150)) {
        return NEUTRAL;
    }

    if (hue < 55) {
        hue = Math.min(50, hue + 12);
    }

    var pastel = hslToRgb(hue, Math.min(0.7, Math.max(0.2, sat * 0.5)), 0.82);
    return rgbToHex(pastel);
}

function applyLightBg() {
    document.documentElement.style.setProperty('--cp-light-bg', warmOrCoolPastel());
}

function setLightSkin() {
    $('header').attr('data-click-state', 0);
    $('footer').attr('data-click-state', 0);
    $('.w3-dark').addClass('w3-white').removeClass('w3-dark');
    $('.w3-dark-grey').addClass('w3-light-grey').removeClass('w3-dark-grey');
    $('.w3-text-light-grey').addClass('w3-text-grey').removeClass('w3-text-light-grey');
    $('body').addClass('light-mode');
    applyLightBg();
    setCookie("skin", "light", Infinity);
}

function setDarkSkin() {
    $('header').attr('data-click-state', 1);
    $('footer').attr('data-click-state', 1);
    $('.w3-white').addClass('w3-dark').removeClass('w3-white');
    $('.w3-light-grey').addClass('w3-dark-grey').removeClass('w3-light-grey');
    $('.w3-text-grey').addClass('w3-text-light-grey').removeClass('w3-text-grey');
    $('body').removeClass('light-mode');
    setCookie("skin", "dark", Infinity);
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
    newlink.setAttribute("href", ROUTE_WEB + "/css/" + skin + ".css?v=" + VERSION);
    newlink.onload = function() {
        applyLightBg();
    };

    oldlink.replaceWith(newlink);
}

function loadCssFromCookie() {
    css = getCookie("css");
    if (css) {
        loadCSS(css);
    } else {
        loadCSS("github_purple");
    }
}

function applySkin() {
    skin = getCookie("skin");

    if (skin === "") {
        skin = "dark"; 
        setCookie("skin", "dark", Infinity);
    }

    if (skin == "dark") {
        setDarkSkin();
    } else {
        setLightSkin();
    }
}

function loader() {
    $("body").animate({opacity: 1}, 800);
}

$(document).ready(function() {
    loader();
    toggleThemeOnHeaderOrFooterClick();
    loadLogoFromCookie();
    loadCssFromCookie();
    applySkin();
});
