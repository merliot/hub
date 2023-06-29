var state
var conn
var online = false

var map
var marker

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "ID:      " + state.Identity.Id + "\r\n"
	system.value += "Model:   " + state.Identity.Model + "\r\n"
	system.value += "Name:    " + state.Identity.Name
}

function showLocation() {
	marker.setLatLng([state.Lat, state.Long])
	map.panTo([state.Lat, state.Long])
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showSystem()
	showLocation()
}

function createMap() {

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

function run(ws) {

	createMap()

	console.log('[gps]', 'connecting...')
	conn = new WebSocket(ws)

	conn.onopen = function(evt) {
		console.log('[gps]', 'open')
		conn.send(JSON.stringify({Path: "get/state"}))
	}

	conn.onclose = function(evt) {
		console.log('[gps]', 'close')
		online = false
		show()
		setTimeout(run(ws), 1000)
	}

	conn.onerror = function(err) {
		console.log('[gps]', 'error', err)
		conn.close()
	}

	conn.onmessage = function(evt) {
		msg = JSON.parse(evt.data)
		console.log('[gps]', msg)

		switch(msg.Path) {
		case "state":
			online = true
			state = msg
			show()
			break
		case "update":
			state.Lat = msg.Lat
			state.Long = msg.Long
			showLocation()
			break
		}
	}
}
