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
	    attribution: '© OpenStreetMap'
	}).addTo(map)

	<!-- Create a map marker with popup that has [Id, Model, Name] -- !>
	popup = "ID: {{.Id}}<br>Model: {{.Model}}<br>Name: {{.Name}}"
	marker = L.marker([0, 0]).addTo(map).bindPopup(popup);
}

function show() {
	marker.setLatLng([state.Lat, state.Long])
	map.panTo([state.Lat, state.Long])
}

function hide() {
}

function handle(msg) {
	switch(msg.Path) {
	case "update":
		state.Lat = msg.Lat
		state.Long = msg.Long
		show()
		break
	}
}
