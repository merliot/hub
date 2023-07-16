var explorer = document.getElementById("explorer")
var view = document.getElementById("view")
var dialogCreate = document.getElementById("create-dialog")
var dialogDelete = document.getElementById("delete-dialog")
var selected

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

	if (typeof selected !== "undefined") {
		var seldiv = document.getElementById("device-" + selected)
		seldiv.style.background = "lightgrey"
	}
	
	selected = id
	var seldiv = document.getElementById("device-" + selected)
	seldiv.style.background = "white"
}

function create() {
	var model = document.getElementById("create-model")
	var id = document.getElementById("create-id")
	var name = document.getElementById("create-name")
	conn.send(JSON.stringify({Path: "create", Id: id.value, Model: model.value, Name: name.value}))
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
	apidev.data = "/" + selected + "/api
	dialogApi.showModal()
}

function stageApi() {
	var btnApi = document.getElementById("api")
	btnApi.onclick = function(){showApi()}
	var btnClose = document.getElementById("api-close")
	btnClose.onclick = function(){dialogApi.close()}
}

function deletef() {
	conn.send(JSON.stringify({Path: "delete", Id: selected}))
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
	selected = undefined
	view.textContent = ''
	explorer.removeChild(div)
	delete state.Devices[id]
}

function loadExplorer() {
	explorer.textContent = ''
	for (let id in state.Devices) {
		insertDevice(id, state.Devices[id])
	}
}

function show() {
	loadExplorer()
	stageCreate()
	stageApi()
	stageDelete()
	stageSave()
	stageDeploy()
}

function hide() {
}

function update(child) {
}

function createBad(msg) {
	var err = document.getElementById("create-err")
	err.innerText = msg.Err
}

function createGood(msg) {
	dialogCreate.close()
	insertDevice(msg.Id, msg)
	clickDev(msg.Id)
}

function deleteBad(msg) {
	var err = document.getElementById("delete-err")
	err.innerText = msg.Err
}

function deleteGood(msg) {
	dialogDelete.close()
	removeDevice(msg.Id)
}

function handle(msg) {
	switch(msg.Path) {
	case "connected":
	case "disconnected":
		update(msg)
		break
	case "create/bad":
		createBad(msg)
		break
	case "create/good":
		createGood(msg)
		break
	case "delete/bad":
		deleteBad(msg)
		break
	case "delete/good":
		deleteGood(msg)
		break
	}
}
