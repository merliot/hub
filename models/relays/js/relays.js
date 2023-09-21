var overlay = document.getElementById("overlay")

function init() {
	var relayImages = document.querySelectorAll(".relay-img");

	relayImages.forEach(function(img, index) {
		img.addEventListener("click", function() {
			var currentSrc = this.getAttribute("src");
			var relay = state.Relays[index]

			if (currentSrc === "images/relay-off.png") {
				relay.State = true
				this.setAttribute("src", "images/relay-on.png");
			} else {
				relay.State = false
				this.setAttribute("src", "images/relay-off.png");
			}
			conn.send(JSON.stringify({Path: "click", Relay: index, State: relay.State}))
		});
	});
}

function open() {
	state.Online ? online() : offline()
}

function close() {
	offline()
}

function saveClick(msg) {
	var relay = state.Relays[msg.Relay]
	var image = document.getElementById("relay" + msg.Relay + "-img")
	relay.State = msg.State
	if (relay.State) {
		image.src = "images/relay-on.png"
	} else {
		image.src = "images/relay-off.png"
	}
}

function online() {
	overlay.innerHTML = ""
	for (var i = 0; i < 4; i++) {
		div = document.getElementById("relay" + i)
		label = document.getElementById("relay" + i + "-name")
		image = document.getElementById("relay" + i + "-img")
		relay = state.Relays[i]
		if (relay.Name === "") {
			div.style.display = "none"
			label.textContent = "<unused>"
			image.src = "images/relay-off.png"
		} else {
			div.style.display = "flex"
			label.textContent = relay.Name
			if (relay.State) {
				image.src = "images/relay-on.png"
			} else {
				image.src = "images/relay-off.png"
			}
		}
	}
}

function offline() {
	overlay.innerHTML = "Offline"
}

function handle(msg) {
	switch(msg.Path) {
	case "click":
		saveClick(msg)
		break
	}
}
