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
	var banner = document.getElementById("banner")
	banner.style.width = state.Terminal.Width + 'ch'
	banner.style.height = state.Terminal.Height + 'em'
}

function offline() {
	overlay.innerHTML = "Offline"
}

function handle(msg) {
	switch(msg.Path) {
	}
}

