// Skeleton Device
//
// Use this skeleton device to start a new device.  Rename instances of
// "skeleton" with your device model name.
//
// NOTE: Most of the comments in the code are instructional and can be removed
// or updated.

import { WebSocketController } from './common.js'

//
// run() is the main entry point for the device.
//

export function run(prefix, ws) {
	const bones = new Skeleton()
	bones.run(prefix, ws)
}

//
// Skeleton is a web socket controller.  It wil manage the web socket
// connection between the device and the client (browser).
//

class Skeleton extends WebSocketController {

	constructor() {
		super()

		//
		// Add device-specific initialization here...
		//
	}

	open() {
		super.open()

		//
		// Open is called when the websocket connects.
		//
		// Add any device-specific open code here...
		//
	}

	close() {
		//
		// Close is called when the websocket disconnects.
		//
		// Add any device-specific close code here...
		//

		super.open()
	}

	handle(msg) {

		//
		// Handle device messages here...
		//

		switch(msg.Path) {
		case "rattle":
			//
			// Handle rattle bones here...
			//
			break
		default:
			super.handle(msg)
		}
	}
}
