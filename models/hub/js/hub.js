var overlay = document.getElementById("overlay")
var explorer = document.getElementById("explorer")
var view = document.getElementById("view")
var createBtn = document.getElementById("create")
var deleteBtn = document.getElementById("delete")
var saveBtn = document.getElementById("save")
var deployBtn = document.getElementById("deploy")
var dialogCreate = document.getElementById("create-dialog")
var dialogDelete = document.getElementById("delete-dialog")
var selected // currently selected device ID
var sub = ""

// Create button opens create model dialog

createBtn.onclick = function(){showCreate()}
document.getElementById("create-close").onclick = function(){dialogCreate.close()}
document.getElementById("create-create").onclick = function(){create()}

// Delete button open delete modal dialog

deleteBtn.onclick = function(){showDelete()}
document.getElementById("delete-close").onclick = function(){dialogDelete.close()}
document.getElementById("delete-delete").onclick = function(){deletef()}

// Save button saves devices.json to repo

saveBtn.onclick = function(){save()}

// Toggle buttons

const buttons = document.querySelectorAll('.toggle-btn');

buttons.forEach(button => {
	button.addEventListener('click', () => {
		// Check if the clicked button is already pressed
		const wasPressed = button.classList.contains('pressed')
		// Unpress all buttons
		buttons.forEach(btn => btn.classList.remove('pressed'))
		// If the clicked button wasn't pressed, press it and set sub
		if (!wasPressed) {
			button.classList.add('pressed')
			sub = button.getAttribute('data-sub-value')
		} else {
			sub = ""
		}
		clickDev(selected)
	})
})

function init() {
}

function clickDev(id) {
	var obj = document.createElement("object")
	obj.data = "/" + id + "/" + sub
	view.textContent = ''
	view.appendChild(obj)

	// grey out previous selected tab
	if (typeof selected !== "undefined") {
		var seldiv = document.getElementById("device-" + selected)
		seldiv.style.background = "lightgrey"
	}
	
	selected = id
	var seldiv = document.getElementById("device-" + selected)
	seldiv.style.background = "white"

	const buttons = document.querySelectorAll('.toggle-btn');
	buttons.forEach(button => { button.disabled = false })
}

function unclickDev(prev) {
	const buttons = document.querySelectorAll('.toggle-btn');
	buttons.forEach(button => { button.disabled = true })

	selected = undefined
	view.textContent = ''
	if (prev) {
		parts = prev.id.split('-')
		prevId = parts[1]
		clickDev(prevId)
	}
}

async function create() {
	var id = document.getElementById("create-id")
	var model = document.getElementById("create-model")
	var name = document.getElementById("create-name")

	let response = await fetch("/create?id=" + id.value + "&model=" + model.value + "&name=" + name.value)

	if (response.status == 201) {
		dialogCreate.close()
	} else {
		let data = await response.text()
		var err = document.getElementById("create-err")
		err.innerText = data
	}
}

function showCreate() {
	var createModel = document.getElementById("create-model")
	createModel.textContent = ''
	for (let i in state.Models) {
		var option = document.createElement("option")
		option.value = state.Models[i]
		option.text = state.Models[i]
		createModel.appendChild(option)
	}
	var err = document.getElementById("create-err")
	err.innerText = ""
	dialogCreate.showModal()
}

async function deletef() {
	let response = await fetch("/delete?id=" + selected)

	if (response.status == 200) {
		dialogDelete.close()
	} else {
		let data = await response.text()
		var err = document.getElementById("delete-err")
		err.innerText = data
	}
}

function showDelete() {
	var err = document.getElementById("delete-err")
	err.innerText = ""
	var delprompt = document.getElementById("delete-prompt")
	delprompt.innerText = "Delete device ID " + selected + "?"
	dialogDelete.showModal()
}

async function save() {
	let response = await fetch("/save")
	let data = await response.text()
	alert(data)
}

function setDeviceIcon(img, online) {
	if (img) {
		img.src = online ? "/images/online.png" : "/images/offline.png"
	}
}

function insertDevice(id, dev) {
	var div = document.createElement("div")
	div.onclick = function (){clickDev(id)}
	div.id = "device-" + id
	div.className = "devBtn"

	var img = document.createElement("img")
	img.id = "device-" + id + "-status"
	img.className = "statusIcon"
	setDeviceIcon(img, dev.Online)
	div.appendChild(img)

	var text = document.createTextNode(dev.Name)
	div.appendChild(text)

	explorer.appendChild(div)
}

function removeDevice(id) {
	var div = document.getElementById("device-" + id)
	var prev = div.previousElementSibling
	explorer.removeChild(div)
	delete state.Devices[id]
	if (selected == id) {
		unclickDev(prev)
	}
}

function setBackup() {
	if (state.BackupHub) {
		document.title = state.Model + " - " + state.Name + " (backup)"
	} else {
		createBtn.disabled = false
		deleteBtn.disabled = false
		saveBtn.disabled = false
		deployBtn.style.display = "block"
	}
}

function loadExplorer() {
	explorer.textContent = ''
	view.textContent = ''
	for (let id in state.Devices) {
		insertDevice(id, state.Devices[id])
		if (typeof selected !== "undefined") {
			if (id == selected) {
				clickDev(id)
			}
		}
	}
	if (typeof selected === "undefined") {
		if (Object.keys(state.Devices).length > 0) {
			clickDev(Object.keys(state.Devices)[0])
		}
	}
}

function open() {
	state.Online ? online() : offline()
	loadExplorer()
	setBackup()
}

function close() {
	offline()
	for (let id in state.Devices) {
		disconnected(id)
	}
}

function online() {
	overlay.innerHTML = ""
}

function offline() {
	overlay.innerHTML = "Offline"
}

function connected(id) {
	var img = document.getElementById("device-" + id + "-status")
	setDeviceIcon(img, true)
}

function disconnected(id) {
	var img = document.getElementById("device-" + id + "-status")
	setDeviceIcon(img, false)
}

function handle(msg) {
	switch(msg.Path) {
	case "connected":
		connected(msg.Id)
		break
	case "disconnected":
		disconnected(msg.Id)
		break
	case "created/device":
		insertDevice(msg.Id, msg)
		break
	case "deleted/device":
		removeDevice(msg.Id)
		break
	}
}
