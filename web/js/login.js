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

function applySkin() {
    skin = getCookie("skin");

    if (skin == "dark") {
        setDarkSkin();
    } else {
        setLightSkin();
    }
}

$(document).ready(function() {
    toggleThemeOnHeaderOrFooterClick();
    applySkin();
});