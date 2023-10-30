import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const sign = new Sign()
	sign.run(prefix, ws)
}

class Sign extends WebSocketController {

	constructor() {
		super()
		this.banner = document.getElementById("banner")
		this.clearBtn = document.getElementById("clear")
		this.saveBtn = document.getElementById("save")

		this.clearBtn.onclick = () => this.clear()
		this.saveBtn.onclick = () => this.save()
	}

	open() {
		super.open()
		this.showBanner()
	}

	handle(msg) {
		switch(msg.Path) {
		case "save":
			this.update(msg)
			break
		}
	}

	clear() {
		this.state.Banner = ""
		this.banner.value = ""
	}

	save() {
		this.state.Banner = this.banner.value
		this.conn.send(JSON.stringify({Path: "save", Banner: this.banner.value}))
	}

	update(msg) {
		this.state.Banner = msg.Banner
		this.banner.value = msg.Banner
	}

	showBanner() {
		let style = window.getComputedStyle(this.banner);
		let oneEm = parseFloat(style.fontSize);
		let heightValue = 1.2 * oneEm * this.state.Terminal.Height;

		this.banner.style.width = this.state.Terminal.Width + 'ch'
		this.banner.style.height = `${heightValue}px`;
		this.banner.value = this.state.Banner
		this.banner.style.display = "block"
	}
}
