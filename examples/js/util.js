// Dynamically load CSS file into <head>
function loadCSS(href) {
	if (!document.querySelector(`link[href="${href}"]`)) {
		var link = document.createElement('link');
		link.rel = 'stylesheet';
		link.href = href;
		document.head.appendChild(link);
	}
}

// Dynamically load script file into <head>.  Callback is called once script is
// loaded.
function loadScript(src, callback) {
	if (!document.querySelector(`script[src="${src}"]`)) {
		var script = document.createElement('script');
		script.src = src;
		script.onload = callback;
		document.head.appendChild(script);
	} else {
		if (callback) callback();
	}
}

/*
	TODO: figure out how to this with 100% htmx:
	 - need to set/clear class "offline" on body on ws
	   connect/disconnect
*/

document.addEventListener("htmx:wsOpen", function(event) {
	document.getElementById("session").classList.remove("offline")
});
document.addEventListener("htmx:wsClose", function(event) {
	document.getElementById("session").classList.add("offline")
});
document.addEventListener("htmx:wsError", function(event) {
	document.getElementById("session").classList.add("offline")
});
