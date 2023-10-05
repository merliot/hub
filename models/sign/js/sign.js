var overlay = document.getElementById("overlay")
var banner = document.getElementById("banner")
var saveBtn = document.getElementById("save")

function init() {
}

function open() {
	state.Online ? online() : offline()
}

function close() {
	offline()
}

function handleSave(msg) {
	banner.value = msg.Banner
}

function save() {
	conn.send(JSON.stringify({Path: "save", Banner: banner.value}))
}

function online() {
	overlay.innerHTML = ""

	let style = window.getComputedStyle(banner);
	let oneEm = parseFloat(style.fontSize);
	let heightValue = 1.2 * oneEm * state.Terminal.Height;

	banner.style.width = state.Terminal.Width + 'ch'
	banner.style.height = `${heightValue}px`;
	banner.value = state.Banner

	saveBtn.onclick = save
}

function offline() {
	overlay.innerHTML = "Offline"
}

function handle(msg) {
	switch(msg.Path) {
		case "save":
			handleSave(msg)
			break
	}
}

