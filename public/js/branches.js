const urlParams = new URLSearchParams(window.location.search);
const repoKey = urlParams.get('repo_key');

document.addEventListener('DOMContentLoaded', () => {
    initToastContainer();
    if (!repoKey) {
        showToast("缺少 repo_key 参数", "error");
        return;
    }
    loadRepoInfo();
    loadBranches();
});

async function loadRepoInfo() {
    try {
        // Use GET /repos/:key
        const repo = await request(`/repos/${repoKey}`);
        if (repo) {
            currentRepo = repo;
            document.getElementById('repo-title').innerText = `${repo.name} - 分支管理`;
            document.title = `${repo.name} - 分支管理`;
        } else {
            document.getElementById('repo-title').innerText = "仓库未找到";
        }
    } catch (e) {
        console.error(e);
        document.getElementById('repo-title').innerText = "仓库加载失败";
    }
}

async function loadBranches() {
    const tbody = document.getElementById('branch-list');
    tbody.innerHTML = '<tr><td colspan="5" class="text-center py-4">加载中...</td></tr>';

    const keyword = document.getElementById('searchInput').value;
    
    try {
        const res = await request(`/repos/${repoKey}/branches?keyword=${encodeURIComponent(keyword)}`);
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
            
            // Sync Status UI
            let syncStatus = '';
            let syncBtn = '';
            if (b.upstream) {
                if (b.behind > 0) {
                    syncStatus += `<span class="badge bg-danger me-1" title="落后 ${b.behind} 个提交"><i class="bi bi-arrow-down"></i> ${b.behind}</span>`;
                    if (b.is_current) {
                        syncBtn = `<button class="btn btn-warning btn-sm" onclick="syncBranch('${b.name}')" title="同步代码 (Pull --rebase)"><i class="bi bi-cloud-download"></i></button>`;
                    }
                }
                if (b.ahead > 0) {
                    syncStatus += `<span class="badge bg-success me-1" title="领先 ${b.ahead} 个提交"><i class="bi bi-arrow-up"></i> ${b.ahead}</span>`;
                }
                if (b.behind === 0 && b.ahead === 0) {
                    syncStatus = `<span class="text-success small"><i class="bi bi-check-all"></i> 已同步</span>`;
                }
            } else {
                syncStatus = `<span class="text-muted small">无上游</span>`;
            }

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
                <td>${syncStatus}</td>
                <td class="text-end">
                    <div class="btn-group btn-group-sm">
                        ${syncBtn}
                        <button class="btn btn-outline-dark" onclick="openPushModal('${b.name}')" title="推送至远端"><i class="bi bi-cloud-upload"></i></button>
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
    window.location.href = `compare.html?repo_key=${repoKey}`;
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
        await request(`/repos/${repoKey}/branches`, {
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
        await request(`/repos/${repoKey}/branches/${encodeURIComponent(oldName)}`, {
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
        await request(`/repos/${repoKey}/branches/${encodeURIComponent(name)}?force=true`, {
            method: 'DELETE'
        });
        showToast("删除成功", "success");
        loadBranches();
    } catch (e) {
        // handled
    }
}

function openDetail(branchName) {
    window.location.href = `branch_detail.html?repo_key=${repoKey}&branch=${encodeURIComponent(branchName)}`;
}

async function syncBranch(branch) {
    if (!confirm(`确定要同步分支 ${branch} 吗？\n这将执行 git pull --rebase，可能会产生冲突。`)) return;
    
    try {
        await request(`/repos/${repoKey}/branches/${encodeURIComponent(branch)}/pull`, { method: 'POST' });
        showToast("同步成功", "success");
        loadBranches();
    } catch (e) {
        // handled
    }
}

async function openPushModal(branch) {
    document.getElementById('pushBranchName').innerText = branch;
    document.getElementById('pushBranchName').dataset.branch = branch;
    
    const list = document.getElementById('pushRemoteList');
    list.innerHTML = '<div class="spinner-border spinner-border-sm"></div>';
    new bootstrap.Modal(document.getElementById('pushModal')).show();
    
    try {
        // We need to fetch repo config to get remotes. 
        // We can reuse /repos/scan or just assume we have list from somewhere.
        // Let's use the repo info we loaded in currentRepo global or fetch again.
        
        // Actually currentRepo only has basic info. Let's fetch detail.
        const config = await request('/repos/scan', {
            method: 'POST',
            body: { path: currentRepo.path }
        });
        
        list.innerHTML = '';
        if (!config.remotes || config.remotes.length === 0) {
            list.innerHTML = '<span class="text-muted">无可用远端</span>';
            return;
        }
        
        config.remotes.forEach(r => {
            const div = document.createElement('div');
            div.className = 'form-check';
            div.innerHTML = `
                <input class="form-check-input remote-checkbox" type="checkbox" value="${r.name}" id="remote-${r.name}" checked>
                <label class="form-check-label" for="remote-${r.name}">
                    ${r.name} <span class="text-muted small">(${r.push_url || r.fetch_url})</span>
                </label>
            `;
            list.appendChild(div);
        });
        
    } catch (e) {
        list.innerHTML = '<span class="text-danger">加载远端失败</span>';
    }
}

async function submitPush() {
    const branch = document.getElementById('pushBranchName').dataset.branch;
    const remotes = [];
    document.querySelectorAll('.remote-checkbox:checked').forEach(cb => remotes.push(cb.value));
    
    if (remotes.length === 0) {
        showToast("请至少选择一个远端", "warning");
        return;
    }
    
    const btn = document.querySelector('#pushModal .btn-primary');
    btn.disabled = true;
    btn.innerText = "推送中...";
    
    try {
        await request(`/repos/${repoKey}/branches/${encodeURIComponent(branch)}/push`, {
            method: 'POST',
            body: { remotes }
        });
        showToast("推送成功", "success");
        bootstrap.Modal.getInstance(document.getElementById('pushModal')).hide();
        loadBranches();
    } catch (e) {
        // handled
    } finally {
        btn.disabled = false;
        btn.innerText = "确认推送";
    }
}
