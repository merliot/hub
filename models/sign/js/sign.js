var overlay = document.getElementById("overlay")

function init() {
}

function open() {
	state.Online ? online() : offline()
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
		break
	}
}

