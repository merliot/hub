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

function unclickDev(id) {
	document.getElementById("delete").disabled = true
	document.getElementById("deploy").disabled = true
	selected = undefined
	view.textContent = ''
}

function create() {
	var model = document.getElementById("create-model")
	var id = document.getElementById("create-id")
	var name = document.getElementById("create-name")
	conn.send(JSON.stringify({Path: "create/device", Id: id.value, Model: model.value, Name: name.value}))
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

function deletef() {
	conn.send(JSON.stringify({Path: "delete/device", Id: selected}))
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

function insertDevice(id, dev) {
	var div = document.createElement("div")
	div.onclick = function (){clickDev(id)}
	div.id = "device-" + id
	div.className = "devBtn"
	div.appendChild(document.createTextNode(dev.Name))
	explorer.appendChild(div)
	if (explorer.children.length == 1) {
		clickDev(id)
	}
}

function removeDevice(id) {
	var div = document.getElementById("device-" + id)
	explorer.removeChild(div)
	delete state.Devices[id]
	if (selected == id) {
		unclickDev(id)
	}
}

function loadExplorer() {
	explorer.textContent = ''
	for (let id in state.Devices) {
		insertDevice(id, state.Devices[id])
	}
}

function open() {
	loadExplorer()
	stageCreate()
	stageApi()
	stageDelete()
	stageSave()
	stageDeploy()
}

function close() {
}

function online() {
}

function offline() {
}

function update(child) {
}

function createDeviceResult(msg) {
	if (msg.Err == "") {
		dialogCreate.close()
		clickDev(msg.Id)
	} else {
		var err = document.getElementById("create-err")
		err.innerText = msg.Err
	}
}

function deleteDeviceResult(msg) {
	if (msg.Err == "") {
		dialogDelete.close()
	} else {
		var err = document.getElementById("delete-err")
		err.innerText = msg.Err
	}
}

function handle(msg) {
	switch(msg.Path) {
	case "connected":
	case "disconnected":
		update(msg)
		break
	case "create/device":
		insertDevice(msg.Id, msg)
		break
	case "create/device/result":
		createDeviceResult(msg)
		break
	case "delete/device":
		removeDevice(msg.Id)
		break
	case "delete/device/result":
		deleteDeviceResult(msg)
		break
	}
}
