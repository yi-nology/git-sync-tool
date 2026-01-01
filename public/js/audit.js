// audit.js - Audit Log Logic

document.addEventListener('DOMContentLoaded', () => {
    initToastContainer();
    loadAuditLogs();
});

async function loadAuditLogs() {
    const tbody = document.getElementById('auditList');
    tbody.innerHTML = '<tr><td colspan="5" class="text-center">加载中...</td></tr>';
    
    try {
        const logs = await request('/audit/logs');
        tbody.innerHTML = '';
        
        if (!logs || logs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" class="text-center text-muted">暂无审计记录</td></tr>';
            return;
        }

        logs.forEach(log => {
            const tr = document.createElement('tr');
            
            // Format Action Color
            let actionBadge = 'secondary';
            if (log.action === 'CREATE') actionBadge = 'success';
            else if (log.action === 'UPDATE') actionBadge = 'primary';
            else if (log.action === 'DELETE') actionBadge = 'danger';
            else if (log.action.startsWith('SYNC')) actionBadge = 'info';

            // Format Details Button
            let detailsBtn = '';
            if (log.details && log.details !== '{}' && log.details !== 'null') {
                // We need to escape the details string for the onclick handler
                // Or just store it in dataset
                detailsBtn = `<button class="btn btn-sm btn-link" onclick="showAuditDetails(this)">查看</button>`;
            } else {
                detailsBtn = '<span class="text-muted small">-</span>';
            }

            tr.innerHTML = `
                <td>${new Date(log.created_at).toLocaleString()}</td>
                <td><span class="badge bg-${actionBadge}">${log.action}</span></td>
                <td><code class="text-dark">${log.target}</code></td>
                <td>
                    <div>${log.operator}</div>
                    <div class="small text-muted">${log.ip_address || '-'}</div>
                </td>
                <td>${detailsBtn}</td>
            `;
            // Store details in dataset to avoid quoting hell
            tr.dataset.details = log.details;
            tbody.appendChild(tr);
        });
    } catch (e) {
        tbody.innerHTML = '<tr><td colspan="5" class="text-center text-danger">加载失败</td></tr>';
    }
}

function showAuditDetails(btn) {
    const tr = btn.closest('tr');
    const content = tr.dataset.details;
    const modalBody = document.getElementById('auditDetailContent');
    
    try {
        const obj = JSON.parse(content);
        modalBody.innerText = JSON.stringify(obj, null, 2);
    } catch (e) {
        modalBody.innerText = content;
    }
    
    new bootstrap.Modal(document.getElementById('auditDetailModal')).show();
}
