var gauge

function init() {
	if (typeof guage !== 'undefined') {
		return
	}

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
	gauge = new Gauge(bh1750).setOptions(opts)

	gauge.maxValue = 500
	gauge.setMinValue(0)
	gauge.animationSpeed = 32
	gauge.set(0)
}

function show() {
	showDevice()
	gauge.set(state.Lux)
}

function hide() {
}

function handle(msg) {
	switch(msg.Path) {
	case "update":
		state.Lux = msg.Lux
		show()
		break
	}
}
