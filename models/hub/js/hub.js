var explorer = document.getElementById("explorer")
var view = document.getElementById("view")
var deploy = document.getElementById("deploy")
var dialogDeploy= document.getElementById("deploy-dialog")

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
	deploy.onclick = function(){clickDeploy(id)}
}

function insertDevice(id, dev) {
	var div = document.createElement("div")
	div.onclick = function (){clickDev(id)}
	div.appendChild(document.createTextNode(dev.Name))
	explorer.appendChild(div)
}

function show() {
	explorer.textContent = ''
	for (let id in state.Devices) {
		insertDevice(id, state.Devices[id])
	}
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
