var explorer = document.getElementById("explorer")
var view = document.getElementById("view")

function init() {
}

folder = "folder:"
device = "device:"

function clickedDev(id) {
	view.textContent = ''
	var elem = document.createElement("iframe")
	elem.src = "/" + encodeURIComponent(id) + "/"
	view.appendChild(elem)
}

function clickedFolder() {
	view.textContent = ''
}

function insertFolder(level, key) {
	var elem = document.createElement("div")
	elem.style.paddingLeft = level * 20 + "px"
	elem.onclick = function (){clickedFolder()}
	elem.appendChild(document.createTextNode(key))
	explorer.appendChild(elem)
}

function insertDevice(level, dev) {
	var elem = document.createElement("div")
	elem.style.paddingLeft = level * 20 + "px"
	elem.onclick = function (){clickedDev(dev["Id"])}
	elem.appendChild(document.createTextNode(dev["Name"]))
	explorer.appendChild(elem)
}

function doit(level, devs) {
	for (let key in devs) {
		if (key.startsWith(folder)) {
			insertFolder(level, key)
			doit(level+1, devs[key])
		} else if (key.startsWith(device)) {
			insertDevice(level, devs[key])
		}
	}
}

function show() {
	explorer.textContent = ''
	doit(0, state.Devices)
}

function hide() {
}

function update(child) {
}

function handle(msg) {
	switch(msg.Path) {
	case "connected":
	case "disconnected":
		update(msg)
		break
	}
}
