class DeviceHub extends DeviceBase {

	constructor(container, view, assets) {
		super(container, view, assets)

		this.menu = document.getElementById("menu");

		this.menuIcon = document.getElementById("menu-icon")
		this.menuIcon.onclick = () => { this.toggleMenu() }

		this.menuItems = document.querySelectorAll(".menu-item");
		this.menuItems.forEach(item => {
			item.addEventListener("click", () => {
				this.menuItemClick(item);
			});
		});

	}

	toggleMenu() {
		menu.style.display = (menu.style.display === "block") ? "none" : "block";
	}

	downloadDevices() {
		var a = document.createElement('a');
		a.href = '/devices';
		a.download = 'devices.json';
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
	}

	showAboutDialog() {
		var aboutDialog = document.getElementById("about-dialog")
		aboutDialog.showModal()
		var close = document.getElementById("about-close")
		close.onclick = (event) => aboutDialog.close()
		close.focus()
	}

	menuItemClick(item) {
		menu.style.display = "none"
		switch (item.textContent) {
		case "Download Devices":
			this.downloadDevices()
			break;
		case "About":
			this.showAboutDialog()
			break;
		}
	}
}
