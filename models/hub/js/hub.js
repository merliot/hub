var explorer = document.getElementById("explorer")
var view = document.getElementById("view")
var btnDeploy = document.getElementById("deploy")
var btnCreate = document.getElementById("create")
var dialogDeploy = document.getElementById("deploy-dialog")
var dialogCreate = document.getElementById("create-dialog")

function init() {
}

function runDeploy(id) {
	dialogDeploy.close()
}

function clickDeploy(id) {
	var btnClose = document.getElementById("deploy-close")
	var btnDeploy = document.getElementById("deploy-deploy")
	btnClose.onclick = function(){dialogDeploy.close()}
	btnDeploy.onclick = function(){runDeploy(id)}
	dialogDeploy.showModal()
}

function clickDev(id) {
	view.textContent = ''
	var obj = document.createElement("object")
	obj.data = "/" + id + "/"
	view.appendChild(obj)
	btnDeploy.onclick = function(){clickDeploy(id)}
}

function insertDevice(id, dev) {
	var div = document.createElement("div")
	div.onclick = function (){clickDev(id)}
	div.appendChild(document.createTextNode(dev.Name))
	explorer.appendChild(div)
}

function stageCreate() {
	var btnClose = document.getElementById("create-close")
	btnClose.onclick = function(){dialogCreate.close()}
	btnCreate.onclick = function(){dialogCreate.showModal()}

	var createModels = document.getElementById("create-models")
	createModels.textContent = ''
	for (let i in state.Models) {
		var option = document.createElement("option")
		option.value = state.Models[i]
		option.text = state.Models[i]
		createModels.appendChild(option)
	}
}

function show() {
	explorer.textContent = ''
	for (let id in state.Devices) {
		insertDevice(id, state.Devices[id])
	}
	stageCreate()
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
