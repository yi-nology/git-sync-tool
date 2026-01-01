const API_BASE = '/api';

function getStatusColor(status) {
    if (status === 'success') return 'success';
    if (status === 'failed') return 'danger';
    if (status === 'conflict') return 'warning';
    return 'secondary';
}

function showLog(details) {
    document.getElementById('logContent').textContent = details;
    new bootstrap.Modal(document.getElementById('logModal')).show();
}

async function loadSSHKeys(selectId = 'sshKeySelect') {
    const res = await fetch(`${API_BASE}/system/ssh-keys`);
    if (!res.ok) return;
    const keys = await res.json();
    const select = document.getElementById(selectId);
    select.innerHTML = '<option value="">手动输入路径...</option>';
    keys.forEach(key => {
        select.innerHTML += `<option value="${key.path}">${key.name}</option>`;
    });
}
