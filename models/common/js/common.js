class WebSocketController {

	constructor() {
		this.state = null;
		this.conn = null;
		this.pingID = null;
		this.pingAlive = false;
		this.pingSent = null;
		this.stat = document.getElementById("status");
	}

	run(prefix, ws) {

		const url = new URL(ws);
		const params = new URLSearchParams(url.search);
		const pingPeriod = params.get("ping-period") * 1000;

		console.log(prefix, 'connecting...');
		this.conn = new WebSocket(ws);

		this.conn.onopen = (evt) => {
			console.log(prefix, 'open');
			this.conn.send(JSON.stringify({Path: "get/state"}));
			this.pingAlive = true;
			this.pingID = setInterval(() => this.ping(prefix), pingPeriod)
		};

		this.conn.onclose = (evt) => {
			console.log(prefix, 'close');
			this.close();
			clearInterval(this.pingID);
			setTimeout(() => this.run(prefix, ws), 1000); // Reconnecting after 1 second
		};

		this.conn.onerror = (err) => {
			console.log(prefix, 'error', err);
			this.conn.close();
		};

		this.conn.onmessage = (evt) => {

			if (evt.data == "pong") {
				//console.log(prefix, "pong", new Date() - this.pingSent)
				this.pingAlive = true
				return
			}

			var msg = JSON.parse(evt.data)
			console.log(prefix, msg)

			switch(msg.Path) {
				case "state":
					this.state = msg
					this.open()
					break
				case "online":
					this.state.Online = true
					this.online()
					break
				case "offline":
					this.state.Online = false
					this.offline()
					break
				default:
					this.handle(msg)
					break
			}
		};
	}

	ping(prefix) {
		if (!this.pingAlive) {
			console.log(prefix, "NOT ALIVE", new Date() - this.pingSent)
			// This waits for an ACK from server, but the server
			// may be gone, it may take a bit to close the websocket
			this.conn.close()
			clearInterval(this.pingID)
			return
		}
		this.pingAlive = false
		this.conn.send("ping")
		this.pingSent = new Date()
	}

	open() {
		this.state.Online ? this.online() : this.offline()
	}

	close() {
		this.offline()
	}

	online() {
		if (this.stat !== null) {
			this.stat.innerHTML = ""
			this.stat.style.border = "none"
			this.stat.style.color = "none"
		}
	}

	offline() {
		if (this.stat !== null) {
			this.stat.innerHTML = "Offline"
			this.stat.style.border = "solid"
			this.stat.style.color = "red"
		}
	}

	handle(msg) {
		// drop msg
	}
}

export { WebSocketController };
