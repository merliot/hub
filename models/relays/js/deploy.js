// Sample items list for the dropdowns
const items = ['Item 1', 'Item 2', 'Item 3', 'Item 4', 'Item 5', 'Item 6', 'Item 7', 'Item 8'];

// Function to add a relay
function addRelay() {
    const relaysList = document.getElementById('relaysList');
    const existingRelays = relaysList.getElementsByClassName('relayItem');

    // Check if all items are already added
    if (existingRelays.length >= items.length) {
        alert('Maximum relays added.');
        return;
    }

    const relayItem = document.createElement('div');
    relayItem.classList.add('relayItem');

    const label = document.createElement('label');
    label.classList.add('relayLabel');
    label.textContent = `Relay #${existingRelays.length + 1}`;
    relayItem.appendChild(label);

    const input = document.createElement('input');
    input.type = 'text';
    input.placeholder = 'Relay';
    input.name = 'relay_' + (existingRelays.length + 1);
    relayItem.appendChild(input);

    const dropdown = document.createElement('select');
    dropdown.classList.add('dropdown');
    dropdown.onchange = function() {
        updateAllDropdowns(dropdown);
    };
		dropdown.name = 'item_' + (existingRelays.length + 1);
    relayItem.appendChild(dropdown);

    const deleteIcon = document.createElement('span');
    deleteIcon.classList.add('icon', 'trashIcon');
    deleteIcon.innerHTML = '&#x1F5D1;';
    deleteIcon.onclick = function() {
        deleteRelay(deleteIcon);
    };
    relayItem.appendChild(deleteIcon);

    relaysList.appendChild(relayItem);

    updateAllDropdowns();
}

// Function to delete a relay
function deleteRelay(deleteIconElement) {
    const relaysList = document.getElementById('relaysList');
    const relayItem = deleteIconElement.closest('.relayItem');
    relaysList.removeChild(relayItem);

    const existingRelays = relaysList.getElementsByClassName('relayItem');
    for (let i = 0; i < existingRelays.length; i++) {
        existingRelays[i].querySelector('.relayLabel').textContent = `Relay #${i + 1}`;
        existingRelays[i].querySelector('input').name = 'relay_' + (i + 1);
        existingRelays[i].querySelector('select').name = 'item_' + (i + 1);
    }

    updateAllDropdowns();
}

// Function to update dropdown options based on selected items
function updateAllDropdowns(changedDropdown = null) {
    const relaysList = document.getElementById('relaysList');
    const dropdowns = relaysList.getElementsByClassName('dropdown');
    const selectedItems = Array.from(dropdowns).map(dropdown => dropdown.value).filter(value => value);

    for (const dropdown of dropdowns) {
        if (dropdown !== changedDropdown && dropdown.value) {
            // If it's not the changed dropdown and it has a value, continue without making changes
            continue;
        }

        const currentValue = dropdown.value;
        dropdown.innerHTML = '';  // Clear current options

        // Add the default option
        const defaultOption = document.createElement('option');
        defaultOption.textContent = "Select item";
        defaultOption.value = "";
        dropdown.appendChild(defaultOption);

        for (const item of items) {
            if (!selectedItems.includes(item) || item === currentValue) {
                const option = document.createElement('option');
                option.value = option.textContent = item;
                dropdown.appendChild(option);
            }
        }

        if (changedDropdown === dropdown) {
            dropdown.value = currentValue;
        }
    }
}

// Initial setup for the first relay dropdown
document.addEventListener('DOMContentLoaded', function() {
    updateAllDropdowns();

    const form = document.getElementById('relaysForm');
    form.onsubmit = function(event) {
        event.preventDefault();

        const formData = new FormData(form);
        const params = new URLSearchParams();

        for (const [key, value] of formData.entries()) {
            params.append(key, value);
        }

        console.log('?' + params.toString());
    };
});
