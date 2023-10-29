import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const hub = new Hub()
	hub.run(prefix, ws)
}

class Hub extends WebSocketController {

	constructor() {
		super()

		this.createBtn = document.getElementById("create")
		this.deleteBtn = document.getElementById("delete")
		this.saveBtn = document.getElementById("save")

		this.createBtn.onclick = () => this.showCreate()
		this.deleteBtn.onclick = () => this.showDelete()
		this.saveBtn.onclick = () => this.save()

		this.toggleBtns = document.querySelectorAll('.toggle-btn');
		this.toggleBtns.forEach(btn => {
			btn.onclick = () => this.toggleBtn(btn)
		})

		this.deployBtn = document.getElementById("deploy")

		this.dialogCreate = document.getElementById("create-dialog")
		this.dialogDelete = document.getElementById("delete-dialog")

		this.explorer = document.getElementById("explorer")
		this.view = document.getElementById("view")

		this.selected = null
		this.subPage = ""
	}

	open() {
		this.state.Online ? this.online() : this.offline()
		this.loadExplorer()
		this.setBackup()
	}

	close() {
		this.offline()
		for (let id in this.state.Devices) {
			this.disconnected(id)
		}
	}

	handle(msg) {
		switch(msg.Path) {
		case "connected":
			this.connected(msg.Id)
			break
		case "disconnected":
			this.disconnected(msg.Id)
			break
		case "created/device":
			this.state.Devices[msg.Id] = {Model: msg.Model, Name: msg.Name, Online: false}
			this.insertDevice(msg.Id, msg)
			this.selectOne()
			break
		case "deleted/device":
			this.removeDevice(msg.Id)
			break
		}
	}

	connected(id) {
		this.setDeviceIcon(id, true)
	}

	disconnected(id) {
		this.setDeviceIcon(id, false)
	}

	setDeviceIcon(id, online) {
		var img = document.getElementById("device-" + id + "-status")
		if (img) {
			img.src = online ? "/images/online.png" : "/images/offline.png"
		}
	}

	insertDevice(id, dev) {
		var div = document.createElement("div")
		div.onclick = () => this.clickDev(id)
		div.id = "device-" + id
		div.className = "devBtn"

		var img = document.createElement("img")
		img.id = "device-" + id + "-status"
		img.className = "statusIcon"
		img.src = "/images/offline.png"
		div.appendChild(img)

		var text = document.createTextNode(dev.Name)
		div.appendChild(text)

		this.explorer.appendChild(div)
	}

	removeDevice(id) {
		var div = document.getElementById("device-" + id)
		var prev = div.previousElementSibling
		this.explorer.removeChild(div)
		delete this.state.Devices[id]
		if (this.selected == id) {
			this.unclickDev(prev)
		}
	}

	selectOne() {
		if (this.selected) {
			this.clickDev(this.selected)
		} else {
			if (Object.keys(this.state.Devices).length > 0) {
				this.clickDev(Object.keys(this.state.Devices)[0])
			}
		}
	}

	loadExplorer() {
		this.explorer.textContent = ''
		this.view.textContent = ''
		for (let id in this.state.Devices) {
			this.insertDevice(id, this.state.Devices[id])
		}
		this.selectOne()
	}

	clickDev(id) {
		var obj = document.createElement("object")
		obj.data = "/" + id + "/" + this.subPage
		this.view.textContent = ''
		this.view.appendChild(obj)

		// grey out previous selected tab
		if (this.selected) {
			var seldiv = document.getElementById("device-" + this.selected)
			seldiv.style.background = "lightgrey"
		}
		
		this.selected = id
		var seldiv = document.getElementById("device-" + this.selected)
		seldiv.style.background = "white"

		this.toggleBtns.forEach(button => { button.disabled = false })
	}

	unclickDev(prev) {
		this.toggleBtns.forEach(button => { button.disabled = true })

		this.selected = null
		this.view.textContent = ''
		if (prev) {
			parts = prev.id.split('-')
			prevId = parts[1]
			this.clickDev(prevId)
		}
	}

	setBackup() {
		if (this.state.BackupHub) {
			document.title = this.state.Model + " - " + this.state.Name + " (backup)"
		} else {
			this.createBtn.disabled = false
			this.deleteBtn.disabled = false
			this.saveBtn.disabled = false
			this.deployBtn.style.display = "block"
		}
	}

	showCreate() {
		var createClose = document.getElementById("create-close")
		var createCreate = document.getElementById("create-create")
		createClose.onclick = () => this.dialogCreate.close()
		createCreate.onclick = () => this.create()
		var createModel = document.getElementById("create-model")
		createModel.textContent = ''
		for (let i in this.state.Models) {
			var option = document.createElement("option")
			option.value = this.state.Models[i]
			option.text = this.state.Models[i]
			createModel.appendChild(option)
		}
		var err = document.getElementById("create-err")
		err.innerText = ""
		this.dialogCreate.showModal()
	}

	showDelete() {
		var deleteClose = document.getElementById("delete-close")
		var deleteCreate = document.getElementById("delete-delete")
		deleteClose.onclick = () => this.dialogDelete.close()
		deleteCreate.onclick = () => this.deletef()
		var err = document.getElementById("delete-err")
		err.innerText = ""
		var delprompt = document.getElementById("delete-prompt")
		delprompt.innerText = "Delete device ID " + this.selected + "?"
		this.dialogDelete.showModal()
	}

	async create() {
		var id = document.getElementById("create-id")
		var model = document.getElementById("create-model")
		var name = document.getElementById("create-name")

		let response = await fetch("/create?id=" + id.value + "&model=" + model.value + "&name=" + name.value)

		if (response.status == 201) {
			this.dialogCreate.close()
		} else {
			let data = await response.text()
			var err = document.getElementById("create-err")
			err.innerText = data
		}
	}

	async deletef() {
		let response = await fetch("/delete?id=" + this.selected)

		if (response.status == 200) {
			this.dialogDelete.close()
		} else {
			let data = await response.text()
			var err = document.getElementById("delete-err")
			err.innerText = data
		}
	}

	async save() {
		let response = await fetch("/save")
		let data = await response.text()
		alert(data)
	}

	toggleBtn(btn) {
		// Check if the clicked button is already pressed
		const wasPressed = btn.classList.contains('pressed')
		// Unpress all buttons
		this.toggleBtns.forEach(btn => btn.classList.remove('pressed'))
		// If the clicked button wasn't pressed, press it and set subPage
		if (!wasPressed) {
			btn.classList.add('pressed')
			this.subPage = btn.getAttribute('data-sub-value')
		} else {
			this.subPage = ""
		}
		this.clickDev(this.selected)
	}
}
