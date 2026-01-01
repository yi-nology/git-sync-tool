const urlParams = new URLSearchParams(window.location.search);
const repoId = urlParams.get('repo_id');

document.addEventListener('DOMContentLoaded', () => {
    initToastContainer();
    if (!repoId) {
        showToast("缺少 repo_id 参数", "error");
        return;
    }
    loadRepoInfo();
    loadBranches();
});

async function loadRepoInfo() {
    try {
        // Since we don't have GET /repos/:id, we fetch all
        const repos = await request('/repos');
        const repo = repos.find(r => r.id == repoId);
        if (repo) {
            currentRepo = repo;
            document.getElementById('repo-title').innerText = `${repo.name} - 分支管理`;
            document.title = `${repo.name} - 分支管理`;
        } else {
            document.getElementById('repo-title').innerText = "仓库未找到";
        }
    } catch (e) {
        console.error(e);
    }
}

async function loadBranches() {
    const tbody = document.getElementById('branch-list');
    tbody.innerHTML = '<tr><td colspan="5" class="text-center py-4">加载中...</td></tr>';

    const keyword = document.getElementById('searchInput').value;
    
    try {
        const res = await request(`/repos/${repoId}/branches?keyword=${encodeURIComponent(keyword)}`);
        // Response structure: { total: N, list: [] }
        const list = res.list || [];
        const total = res.total || 0;

        document.getElementById('total-count').innerText = `共 ${total} 个分支`;
        tbody.innerHTML = '';

        if (list.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" class="text-center py-4 text-muted">无匹配分支</td></tr>';
            return;
        }

        list.forEach(b => {
            const tr = document.createElement('tr');
            if (b.is_current) {
                tr.classList.add('table-success');
            }

            const dateStr = b.date ? new Date(b.date).toLocaleString() : '-';
            const shortHash = b.hash ? b.hash.substring(0, 7) : '-';
            
            tr.innerHTML = `
                <td>
                    <span class="${b.is_current ? 'branch-current' : ''}">
                        ${b.is_current ? '<i class="bi bi-check-circle-fill me-1"></i>' : '<i class="bi bi-git me-1"></i>'}
                        ${b.name}
                    </span>
                </td>
                <td>
                    <div class="commit-hash"><i class="bi bi-hash"></i> ${shortHash}</div>
                    <div class="small text-muted text-truncate" style="max-width: 300px;" title="${b.message}">${b.message}</div>
                </td>
                <td>
                    <div>${b.author}</div>
                    <div class="small text-muted">${b.author_email}</div>
                </td>
                <td class="small">${dateStr}</td>
                <td class="text-end">
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-outline-secondary" onclick="openDetail('${b.name}')" title="详情"><i class="bi bi-info-circle"></i></button>
                        <button class="btn btn-outline-primary" onclick="openRenameModal('${b.name}')" title="重命名/描述"><i class="bi bi-pencil"></i></button>
                        ${b.is_current ? '' : `<button class="btn btn-outline-danger" onclick="deleteBranch('${b.name}')" title="删除"><i class="bi bi-trash"></i></button>`}
                    </div>
                </td>
            `;
            tbody.appendChild(tr);
        });

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="5" class="text-center py-4 text-danger">加载失败: ${e.message}</td></tr>`;
    }
}

function openComparePage() {
    window.location.href = `compare.html?repo_id=${repoId}`;
}

function openCreateModal() {
    document.getElementById('createForm').reset();
    new bootstrap.Modal(document.getElementById('createModal')).show();
}

async function submitCreate() {
    const form = document.getElementById('createForm');
    const name = form.name.value.trim();
    const base = form.base_ref.value.trim();

    if (!name) {
        showToast("请输入分支名称", "warning");
        return;
    }

    const btn = document.querySelector('#createModal .btn-primary');
    btn.disabled = true;

    try {
        await request(`/repos/${repoId}/branches`, {
            method: 'POST',
            body: { name, base_ref: base }
        });
        showToast("分支创建成功", "success");
        bootstrap.Modal.getInstance(document.getElementById('createModal')).hide();
        loadBranches();
    } catch (e) {
        // handled by request
    } finally {
        btn.disabled = false;
    }
}

function openRenameModal(name) {
    const form = document.getElementById('renameForm');
    form.old_name.value = name;
    form.current_name_display.value = name;
    form.new_name.value = name;
    form.desc.value = ""; // We don't fetch desc yet
    new bootstrap.Modal(document.getElementById('renameModal')).show();
}

async function submitRename() {
    const form = document.getElementById('renameForm');
    const oldName = form.old_name.value;
    const newName = form.new_name.value.trim();
    const desc = form.desc.value.trim();

    if (!newName) {
        showToast("请输入新名称", "warning");
        return;
    }

    const btn = document.querySelector('#renameModal .btn-primary');
    btn.disabled = true;

    try {
        await request(`/repos/${repoId}/branches/${encodeURIComponent(oldName)}`, {
            method: 'PUT',
            body: { new_name: newName, desc: desc }
        });
        showToast("更新成功", "success");
        bootstrap.Modal.getInstance(document.getElementById('renameModal')).hide();
        loadBranches();
    } catch (e) {
        // handled
    } finally {
        btn.disabled = false;
    }
}

async function deleteBranch(name) {
    if (!confirm(`确定要删除分支 "${name}" 吗？\n注意：未合并的改动将会丢失！`)) return;

    try {
        await request(`/repos/${repoId}/branches/${encodeURIComponent(name)}?force=true`, {
            method: 'DELETE'
        });
        showToast("删除成功", "success");
        loadBranches();
    } catch (e) {
        // handled
    }
}

function openDetail(branchName) {
    window.location.href = `branch_detail.html?repo_id=${repoId}&branch=${encodeURIComponent(branchName)}`;
}
