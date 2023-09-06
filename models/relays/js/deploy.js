const sharedItems = ["Item 1", "Item 2", "Item 3", "Item 4", "Item 5", "Item 6", "Item 7", "Item 8"];

function populateDropdown(dropdown, exclude = []) {
	dropdown.innerHTML = '<option value="">Select Item</option>'; // Default option
	sharedItems.forEach(gpio => {
		if (!exclude.includes(gpio)) {
			let option = document.createElement('option');
			option.value = gpio;
			option.textContent = gpio;
			dropdown.appendChild(option);
		}
	});
}

function updateAllDropdowns() {
	const dropdowns = document.getElementsByClassName('gpio');
	const selectedValues = Array.from(dropdowns).map(d => d.value).filter(v => v !== ""); // Exclude empty selections

	for (const dropdown of dropdowns) {
		const currentValue = dropdown.value;
		const exclude = [...selectedValues];
		if(currentValue) {
			exclude.splice(exclude.indexOf(currentValue), 1); // Keep the current value in its dropdown
		}
		populateDropdown(dropdown, exclude);
		dropdown.value = currentValue; // Restore the selected value after repopulating
	}
}

function addRelay() {
	const relaysList = document.getElementById('relaysList');
	const existingNames = relaysList.getElementsByClassName('relayItem');

	if(existingNames.length < sharedItems.length) {
		const newItem = document.createElement('div');
		newItem.classList.add('relayItem', 'divFlexRow');

		const label = document.createElement('label');
		label.classList.add('relayLabel');
		label.textContent = 'Relay #' + (existingNames.length + 1);

		const input = document.createElement('input');
		input.type = 'text';
		input.placeholder = 'Name';
		input.name = 'relay_' + (existingNames.length + 1);

		const dropdown = document.createElement('select');
		dropdown.classList.add('gpio');
		dropdown.onchange = updateAllDropdowns;
		dropdown.name = 'gpio_' + (existingNames.length + 1);

		const trashIcon = document.createElement('span');
		trashIcon.classList.add('trashIcon');
		trashIcon.innerHTML = '&#x1F5D1;';
		trashIcon.onclick = function() {
			deleteRelay(trashIcon);
		};

		newItem.appendChild(label);
		newItem.appendChild(input);
		newItem.appendChild(dropdown);
		newItem.appendChild(trashIcon);

		relaysList.appendChild(newItem);
		updateAllDropdowns();
	} else {
		alert('Maximum relays reached.');
	}
}

function deleteRelay(deleteIconElement) {
	const relaysList = document.getElementById('relaysList');
	const relayItem = deleteIconElement.closest('.relayItem');
	relaysList.removeChild(relayItem);

	const existingNames = relaysList.getElementsByClassName('relayItem');
	if (existingNames.length === 0) {
		addRelay();
	} else {
		// Renumber the relays
		for (let i = 0; i < existingNames.length; i++) {
			existingNames[i].querySelector('.relayLabel').textContent = 'Relay #' + (i + 1);
		}
		updateAllDropdowns();
	}
}

// On page load, populate the initial dropdown
window.onload = function() {
	const initialDropdown = document.querySelector('.gpio');
	populateDropdown(initialDropdown);
}

