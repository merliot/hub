var state
var conn
var overlay = document.getElementById("overlay")
var overlay = document.getElementById("overlay")

var opts = {
	angle: -0.2, // The span of the gauge arc
	lineWidth: 0.2, // The line thickness
	radiusScale: 1, // Relative radius
	pointer: {
		length: 0.6, // // Relative to gauge radius
		strokeWidth: 0.035, // The thickness
		color: '#000000' // Fill color
	},
	limitMax: true,      // If false, max value increases automatically if value > maxValue
	limitMin: false,     // If true, the min value of the gauge will be fixed
	highDpiSupport: true,     // High resolution support
	staticZones: [
		{strokeStyle: "#30B32D", min:   0, max: 350}, // Green
		{strokeStyle: "#FFDD00", min: 350, max: 380}, // Yellow
		{strokeStyle: "#F03E3E", min: 380, max: 500}  // Red
	],
}
var bh1750= document.getElementById('bh1750')
var gauge = new Gauge(bh1750).setOptions(opts)

gauge.maxValue = 500
gauge.setMinValue(0)
gauge.animationSpeed = 32
gauge.set(0)

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "ID:      " + state.Identity.Id + "\r\n"
	system.value += "Model:   " + state.Identity.Model + "\r\n"
	system.value += "Name:    " + state.Identity.Name
}

function showLux() {
	gauge.set(state.Lux)
}

function offline() {
	overlay.style.display = "block"
	clearInterval(pingID)
}

function ping() {
	conn.send("ping")
}

function online() {
	showSystem()
	showLux()
	overlay.style.display = "none"
	// for Koyeb work-around
	pingID = setInterval(ping, 1500)
}

function run(ws) {

	console.log('[sense]', 'connecting...')
	conn = new WebSocket(ws)

	conn.onopen = function(evt) {
		console.log('[sense]', 'open')
		conn.send(JSON.stringify({Path: "get/state"}))
	}

	conn.onclose = function(evt) {
		console.log('[sense]', 'close')
		offline()
		setTimeout(run(ws), 1000)
	}

	conn.onerror = function(err) {
		console.log('[sense]', 'error', err)
		conn.close()
	}

	conn.onmessage = function(evt) {
		msg = JSON.parse(evt.data)
		console.log('[sense]', msg)

		switch(msg.Path) {
		case "state":
			state = msg
			online()
			break
		case "update":
			state.Lux = msg.Lux
			showLux()
			break
		}
	}
}
