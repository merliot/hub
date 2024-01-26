import { WebSocketController } from './common.js'

export function run(prefix, url) {
	const prime = new Prime(prefix, url)
}

class Prime extends WebSocketController {

	constructor(prefix, url) {
		super(prefix, url)
		this.view = document.getElementById("view")
	}

	open() {
		super.open()
	}
}
