import { WebSocketController } from './common.js'

export function run(prefix, url) {
	const prime = new Prime(prefix, url)
}

class Prime extends WebSocketController {

	constructor(prefix, url) {
		super(prefix, url)
		this.view = document.getElementById("view")
		this.lastPath = ""
	}

	open() {
		super.open()
		document.title = this.state.Child.Model + " - " + this.state.Child.Name
		const path = "/" + this.state.Child.Id + "/";
		if (this.lastPath !== path) {
			this.lastPath = path
			this.view.data = path;
		}

	}
}
