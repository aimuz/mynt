document.addEventListener('DOMContentLoaded', () => {
    fetchDisks();
    fetchPools();
    fetchDatasets();
    setupSSE();
    setupModal();
});

let availableDisks = [];

async function fetchDisks() {
    try {
        const response = await fetch('/api/v1/disks');
        availableDisks = await response.json();
        renderDisks(availableDisks);
    } catch (error) {
        console.error('Error fetching disks:', error);
    }
}

async function fetchPools() {
    try {
        const response = await fetch('/api/v1/pools');
        const pools = await response.json();
        renderPools(pools);
    } catch (error) {
        console.error('Error fetching pools:', error);
    }
}

async function fetchDatasets() {
    try {
        const response = await fetch('/api/v1/datasets');
        const datasets = await response.json();
        renderDatasets(datasets);
    } catch (error) {
        console.error('Error fetching datasets:', error);
    }
}

function renderDisks(disks) {
    const container = document.getElementById('disk-list');
    container.innerHTML = '';

    if (!disks || disks.length === 0) {
        container.innerHTML = '<p>No disks found.</p>';
        return;
    }

    disks.forEach(disk => {
        const card = document.createElement('div');
        card.className = 'card';
        card.innerHTML = `
            <h3>${disk.name}</h3>
            <p><strong>Model:</strong> ${disk.model}</p>
            <p><strong>Size:</strong> ${formatBytes(disk.size)}</p>
            <p><strong>Type:</strong> ${disk.type}</p>
            <p><strong>Path:</strong> ${disk.path}</p>
        `;
        container.appendChild(card);
    });
}

function renderPools(pools) {
    const container = document.getElementById('pool-list');
    container.innerHTML = '';

    if (!pools || pools.length === 0) {
        container.innerHTML = '<p>No pools found.</p>';
        return;
    }

    pools.forEach(pool => {
        const card = document.createElement('div');
        card.className = 'card';
        card.innerHTML = `
            <h3>${pool.name}</h3>
            <p><strong>Health:</strong> <span class="${pool.health === 'ONLINE' ? 'status-online' : 'status-offline'}">${pool.health}</span></p>
            <p><strong>Size:</strong> ${formatBytes(pool.size)}</p>
            <p><strong>Free:</strong> ${formatBytes(pool.free)}</p>
        `;
        container.appendChild(card);
    });
}

function renderDatasets(datasets) {
    const tbody = document.querySelector('#dataset-table tbody');
    tbody.innerHTML = '';

    if (!datasets || datasets.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5">No datasets found.</td></tr>';
        return;
    }

    datasets.forEach(ds => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${ds.name}</td>
            <td>${ds.type}</td>
            <td>${formatBytes(ds.used)}</td>
            <td>${formatBytes(ds.available)}</td>
            <td>${ds.mountpoint}</td>
        `;
        tbody.appendChild(tr);
    });
}

function setupModal() {
    const modal = document.getElementById('modal-create-pool');
    const btn = document.getElementById('btn-create-pool');
    const span = document.getElementsByClassName('close')[0];
    const form = document.getElementById('form-create-pool');

    btn.onclick = () => {
        renderDiskSelection();
        modal.classList.remove('hidden');
    }

    span.onclick = () => {
        modal.classList.add('hidden');
    }

    window.onclick = (event) => {
        if (event.target == modal) {
            modal.classList.add('hidden');
        }
    }

    form.onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const devices = [];
        document.querySelectorAll('input[name="devices"]:checked').forEach(cb => {
            devices.push(cb.value);
        });

        const data = {
            name: formData.get('name'),
            type: formData.get('type'),
            devices: devices
        };

        try {
            const resp = await fetch('/api/v1/pools', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            if (resp.ok) {
                modal.classList.add('hidden');
                form.reset();
                // Refresh pools after a delay to allow ZFS to settle
                setTimeout(fetchPools, 1000);
            } else {
                const err = await resp.text();
                alert('Failed to create pool: ' + err);
            }
        } catch (error) {
            console.error('Error creating pool:', error);
            alert('Error creating pool');
        }
    };
}

function renderDiskSelection() {
    const container = document.getElementById('disk-selection');
    container.innerHTML = '';

    if (!availableDisks || availableDisks.length === 0) {
        container.innerHTML = '<p>No disks available.</p>';
        return;
    }

    availableDisks.forEach(disk => {
        const label = document.createElement('label');
        label.innerHTML = `
            <input type="checkbox" name="devices" value="${disk.path}">
            ${disk.name} (${disk.model} - ${formatBytes(disk.size)})
        `;
        container.appendChild(label);
    });
}

function setupSSE() {
    const evtSource = new EventSource("/api/v1/events");

    evtSource.onmessage = (e) => {
        const data = JSON.parse(e.data);
        showNotification(data);
    };

    evtSource.onerror = (e) => {
        console.error("SSE Error:", e);
    };
}

function showNotification(evt) {
    const container = document.getElementById('notifications');
    const list = document.getElementById('notification-list');

    container.classList.remove('hidden');

    const li = document.createElement('li');
    li.textContent = `[${evt.type}] ${evt.message}`;
    list.appendChild(li);
}

function formatBytes(bytes, decimals = 2) {
    if (!+bytes) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}
