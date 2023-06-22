var state
var conn
var online = false

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
		{strokeStyle: "#30B32D", min:      0, max:  650000}, // Green
		{strokeStyle: "#FFDD00", min: 650000, max:  700000}, // Yellow
		{strokeStyle: "#F03E3E", min: 700000, max: 1000000}  // Red
	],
}
var bh1750= document.getElementById('bh1750')
var gauge = new Gauge(bh1750).setOptions(opts)

gauge.maxValue = 1000000
gauge.setMinValue(0)
gauge.animationSpeed = 32
gauge.set(0)

function showTemp() {
	let tempc = document.getElementById("tempc")
	tempc.value = ""
	tempc.value = "Temperature:     " + state.TempC + "(C)"
}

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "CPU Frequency:   " + state.CPUFreq + "Mhz\r\n"
	system.value += "MAC Address:     " + state.Mac + "\r\n"
	system.value += "IP Address:      " + state.Ip
}

function showBH1750() {
	gauge.set(state.Lux)
}

function showRelay() {
	let relay = document.getElementById("relay")
	if (650000 <= state.Lux && state.Lux <= 700000) {
		relay.src = "images/relay-on.svg"
	} else {
		relay.src = "images/relay-off.svg"
	}
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showSystem()
	showTemp()
	showBH1750()
	showRelay()
}

function reset() {
	state.Lux = 0
	showBH1750()
	showRelay()
	conn.send(JSON.stringify({Path: "lux", Lux: 0}))
	conn.send(JSON.stringify({Path: "reset"}))
}

function run(ws) {

	console.log('connecting...')
	conn = new WebSocket(ws)

	conn.onopen = function(evt) {
		console.log("open")
		conn.send(JSON.stringify({Path: "get/state"}))
	}

	conn.onclose = function(evt) {
		console.log("close")
		online = false
		show()
		setTimeout(run(ws), 1000)
	}

	conn.onerror = function(err) {
		console.log("error", err)
		conn.close()
	}

	conn.onmessage = function(evt) {
		msg = JSON.parse(evt.data)
		console.log('connect', msg)

		switch(msg.Path) {
		case "state":
			online = true
			// fall-thru
		case "update":
			state = msg
			show()
			break
		case "lux":
			state.Lux = msg.Lux
			showBH1750()
			showRelay()
			break
		case "tempc":
			state.TempC = msg.TempC
			showTemp()
			break
		}
	}
}

