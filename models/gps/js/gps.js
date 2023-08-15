var overlay = document.getElementById("overlay")

var map
var marker

function init() {
	if (typeof map !== 'undefined') {
		return
	}

	<!-- Create a Leaflet map using OpenStreetMap -->
	map = L.map('map').setZoom(13)
	L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
		maxZoom: 19,
		attribution: '© OpenStreetMap',
	}).addTo(map)

	marker = L.marker([0, 0]).addTo(map)
}

function showMarker() {
	marker.setLatLng([state.Lat, state.Long])
	map.panTo([state.Lat, state.Long])
}

function open() {
	state.Online ? online() : offline()
	showMarker()
}

function close() {
	offline()
}

function online() {
	overlay.innerHTML = ""
}

function offline() {
	overlay.innerHTML = "Offline"
}

function handle(msg) {
	switch(msg.Path) {
	case "update":
		state.Lat = msg.Lat
		state.Long = msg.Long
		showMarker()
		break
	}
}
