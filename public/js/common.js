// API_BASE is defined in request.js
// const API_BASE = '/api';

function getStatusColor(status) {
    if (status === 'success') return 'success';
    if (status === 'failed') return 'danger';
    if (status === 'conflict') return 'warning';
    return 'secondary';
}

function showLog(log) {
    const modal = new bootstrap.Modal(document.getElementById('logModal'));
    const logContent = document.getElementById('logContent');
    const cmdContainer = document.getElementById('commandContainer');
    const cmdContent = document.getElementById('commandContent');

    logContent.innerText = log || "无日志";
    
    // Parse Command
    const cmdMatch = (log || "").match(/\[.*\] Command: (.+)/);
    if (cmdMatch && cmdMatch[1]) {
        cmdContent.innerText = cmdMatch[1].trim();
        cmdContainer.classList.remove('d-none');
    } else {
        cmdContainer.classList.add('d-none');
    }
    
    modal.show();
}

function copyCommand() {
    const text = document.getElementById('commandContent').innerText;
    navigator.clipboard.writeText(text).then(() => {
        showToast("命令已复制", "success");
    });
}

async function loadSSHKeys(selectId = 'sshKeySelect') {
    try {
        const keys = await request('/system/ssh-keys');
        const select = document.getElementById(selectId);
        select.innerHTML = '<option value="">手动输入路径...</option>';
        (keys || []).forEach(key => {
            select.innerHTML += `<option value="${key.path}">${key.name}</option>`;
        });
    } catch (e) {
        console.error("Failed to load SSH keys", e);
    }
}
