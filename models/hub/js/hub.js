var overlay = document.getElementById("overlay")
var explorer = document.getElementById("explorer")
var view = document.getElementById("view")
var dialogCreate = document.getElementById("create-dialog")
var dialogApi = document.getElementById("api-dialog")
var dialogDelete = document.getElementById("delete-dialog")
var selected // currently selected device ID

function init() {
}

function updateDeployLink() {
	var link = document.getElementById("deploy-link")
	var target = document.getElementById("deploy-target")
	var http = document.getElementById("deploy-http")
	var url = "https://sw-poc.merliot.net/deploy?id=" + selected +
		"&target=" + target.value +
		"&http=" + http.checked
	link.href = url
	link.innerHTML = url
}

function clickDeploy() {
	var dialogDeploy = document.getElementById("deploy-dialog")
	var btnClose = document.getElementById("deploy-close")
	var selectTarget = document.getElementById("deploy-target")
	var checkboxHttp = document.getElementById("deploy-http")
	btnClose.onclick = function(){dialogDeploy.close()}
	selectTarget.onchange = function(){updateDeployLink()}
	checkboxHttp.onchange = function(){updateDeployLink()}
	updateDeployLink()
	dialogDeploy.showModal()
}

function clickDev(id) {
	var obj = document.createElement("object")
	obj.data = "/" + id + "/"
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

	document.getElementById("delete").disabled = false
	document.getElementById("deploy").disabled = false
}

function unclickDev(prev) {
	document.getElementById("delete").disabled = true
	document.getElementById("deploy").disabled = true
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
	var err = document.getElementById("create-err")
	err.innerText = ""
	dialogCreate.showModal()
}

function stageCreate() {
	var btnCreate = document.getElementById("create")
	btnCreate.onclick = function(){showCreate()}

	var btnClose = document.getElementById("create-close")
	btnClose.onclick = function(){dialogCreate.close()}

	var btnCreate = document.getElementById("create-create")
	btnCreate.onclick = function(){create()}

	var createModel = document.getElementById("create-model")
	createModel.textContent = ''
	for (let i in state.Models) {
		var option = document.createElement("option")
		option.value = state.Models[i]
		option.text = state.Models[i]
		createModel.appendChild(option)
	}
}

function showApi() {
	var apihub = document.getElementById("api-hub")
	apihub.data = "/api"
	var apidev = document.getElementById("api-dev")
	apidev.data = "/" + selected + "/api"
	dialogApi.showModal()
}

function stageApi() {
	var btnApi = document.getElementById("api")
	btnApi.onclick = function(){showApi()}
	var btnClose = document.getElementById("api-close")
	btnClose.onclick = function(){dialogApi.close()}
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

function stageDelete() {
	var btn = document.getElementById("delete")
	btn.onclick = function(){showDelete()}

	var btnClose = document.getElementById("delete-close")
	btnClose.onclick = function(){dialogDelete.close()}

	var btnCreate = document.getElementById("delete-delete")
	btnCreate.onclick = function(){deletef()}
}

function stageSave() {
	var btn = document.getElementById("save")
	btn.onclick = function(){alert("not implemented")}
}

function stageDeploy() {
	var btn = document.getElementById("deploy")
	btn.onclick = function(){clickDeploy()}
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
	if (explorer.children.length == 1) {
		clickDev(id)
	}
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

function loadExplorer() {
	selected = undefined
	explorer.textContent = ''
	for (let id in state.Devices) {
		insertDevice(id, state.Devices[id])
	}
}

function open() {
	state.Online ? online() : offline()
	loadExplorer()
	stageCreate()
	stageApi()
	stageDelete()
	stageSave()
	stageDeploy()
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
