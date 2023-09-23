function setCookie(cname, cvalue, exdays) {
    const d = new Date();
    d.setTime(d.getTime() + (exdays*24*60*60*1000));
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

function setLightSkin() {
    $('header').attr('data-click-state', 0);
    $('footer').attr('data-click-state', 0);
    $('.w3-dark').addClass('w3-white').removeClass('w3-dark');
    $('.w3-dark-grey').addClass('w3-light-grey').removeClass('w3-dark-grey');
    $('.w3-text-light-grey').addClass('w3-text-grey').removeClass('w3-text-light-grey');
    setCookie("skin", "light", 30);
}

function setDarkSkin() {
    $('header').attr('data-click-state', 1);
    $('footer').attr('data-click-state', 1);
    $('.w3-white').addClass('w3-dark').removeClass('w3-white');
    $('.w3-light-grey').addClass('w3-dark-grey').removeClass('w3-light-grey');
    $('.w3-text-grey').addClass('w3-text-light-grey').removeClass('w3-text-grey');
    setCookie("skin", "dark", 30);
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

    oldlink.replaceWith(newlink);
}

function loadCssFromCookie() {
    css = getCookie("css");
    if (css) {
        loadCSS(css);
    } else {
        loadCSS("rpi");
    }
}

function applySkin() {
    skin = getCookie("skin");

    if (skin == "dark") {
        setDarkSkin();
    } else {
        setLightSkin();
    }
}

function loader() {
    $("body").fadeIn(800)
}

$(document).ready(function() {
    loader();
    toggleThemeOnHeaderOrFooterClick();
    loadLogoFromCookie();
    loadCssFromCookie();
    applySkin();
});
