// audit.js - Audit Log Logic

let currentPage = 1;
const pageSize = 20;
let totalItems = 0;

document.addEventListener('DOMContentLoaded', () => {
    initToastContainer();
    loadAuditLogs();
});

async function loadAuditLogs(page = 1) {
    currentPage = page;
    const tbody = document.getElementById('auditList');
    tbody.innerHTML = '<tr><td colspan="5" class="text-center">加载中...</td></tr>';
    
    try {
        const response = await request(`/audit/logs?page=${page}&page_size=${pageSize}`);
        tbody.innerHTML = '';
        
        // Handle new response format { items: [], total: ... }
        // Note: request helper usually returns data field directly if success
        // But my backend handler returns { data: { items: [], total: ... } }
        // Need to check what request() returns. Assuming it returns response.data
        
        let logs = [];
        if (response && response.items) {
            logs = response.items;
            totalItems = response.total;
        } else if (Array.isArray(response)) {
             // Fallback for old API if something goes wrong
            logs = response;
            totalItems = logs.length;
        }
        
        if (!logs || logs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" class="text-center text-muted">暂无审计记录</td></tr>';
            updatePagination();
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

            // Always show View button because we load details on demand
            const detailsBtn = `<button class="btn btn-sm btn-link" onclick="showAuditDetails(${log.id})">查看</button>`;

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
            tbody.appendChild(tr);
        });
        
        updatePagination();
    } catch (e) {
        console.error(e);
        tbody.innerHTML = '<tr><td colspan="5" class="text-center text-danger">加载失败</td></tr>';
    }
}

function updatePagination() {
    const info = document.getElementById('paginationInfo');
    const btnPrev = document.getElementById('btnPrev');
    const btnNext = document.getElementById('btnNext');
    
    const start = (currentPage - 1) * pageSize + 1;
    const end = Math.min(currentPage * pageSize, totalItems);
    
    info.innerText = `显示 ${totalItems > 0 ? start : 0} - ${end} 共 ${totalItems} 条`;
    
    btnPrev.classList.toggle('disabled', currentPage <= 1);
    btnNext.classList.toggle('disabled', end >= totalItems);
}

function changePage(delta) {
    const newPage = currentPage + delta;
    if (newPage >= 1 && (newPage - 1) * pageSize < totalItems) {
        loadAuditLogs(newPage);
    }
}

async function showAuditDetails(id) {
    const modalBody = document.getElementById('auditDetailContent');
    modalBody.innerText = '加载中...';
    new bootstrap.Modal(document.getElementById('auditDetailModal')).show();
    
    try {
        const log = await request(`/audit/log?id=${id}`);
        
        if (log.details && log.details !== '{}' && log.details !== 'null') {
            try {
                const obj = JSON.parse(log.details);
                modalBody.innerText = JSON.stringify(obj, null, 2);
            } catch (e) {
                modalBody.innerText = log.details;
            }
        } else {
            modalBody.innerText = '无详细信息';
        }
    } catch (e) {
        modalBody.innerText = '加载详情失败: ' + e.message;
    }
}
