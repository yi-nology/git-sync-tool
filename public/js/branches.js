const urlParams = new URLSearchParams(window.location.search);
const repoKey = urlParams.get('repo_key');
let currentRepo = null;
let currentTab = 'local';
let allBranches = [];
let repoRemotes = [];

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
        const repo = await request(`/repo/detail?key=${repoKey}`);
        if (repo) {
            currentRepo = repo;
            // document.getElementById('repo-title').innerText = `${repo.name} - 分支管理`;
            document.title = `${repo.name} - 分支管理`;
            
            // Load Remotes for filtering
            try {
                const config = await request('/repo/scan', {
                    method: 'POST',
                    body: { path: repo.path }
                });
                repoRemotes = (config.remotes || []).map(r => r.name);
                // Refresh branches after loading remotes (if branches loaded first, render might be wrong)
                if (allBranches.length > 0) renderBranches();
            } catch(e) { console.error("Failed to load remotes", e); }
        } else {
            // document.getElementById('repo-title').innerText = "仓库未找到";
        }
    } catch (e) {
        console.error(e);
        // document.getElementById('repo-title').innerText = "仓库加载失败";
    }
}

async function loadBranches() {
    const tbody = document.getElementById('branch-list');
    tbody.innerHTML = '<tr><td colspan="6" class="text-center py-4">加载中...</td></tr>';
    
    const keyword = document.getElementById('searchInput').value;
    
    try {
        const res = await request(`/branch/list?repo_key=${repoKey}&keyword=${encodeURIComponent(keyword)}&page_size=1000`);
        allBranches = res.list || [];
        console.log("Branches loaded:", allBranches.length);

        renderBranches();

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="6" class="text-center py-4 text-danger">加载失败: ${e.message}</td></tr>`;
    }
}

async function fetchAll() {
    const btn = document.querySelector('button[onclick="fetchAll()"]');
    if (btn) {
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>';
    }

    try {
        await request('/repo/fetch', { method: 'POST', body: { repo_key: repoKey } });
        showToast("远端已刷新", "success");
        loadBranches();
    } catch (e) {
        // handled
    } finally {
        if (btn) {
            btn.disabled = false;
            btn.innerHTML = '<i class="bi bi-cloud-arrow-down"></i> 刷新远端 (Fetch)';
        }
    }
}

function switchTab(tab) {
    currentTab = tab;
    document.querySelectorAll('.nav-link').forEach(btn => btn.classList.remove('active'));
    document.getElementById(`tab-${tab}`).classList.add('active');
    renderBranches();
}

function renderBranches() {
    const tbody = document.getElementById('branch-list');
    tbody.innerHTML = '';
    
    let filtered = [];
    // Prepare map for remote -> local branch lookup
    const upstreamMap = {};

    // First pass: identify local branches and their upstreams
    const localBranches = allBranches.filter(b => !repoRemotes.some(r => b.name.startsWith(r + '/')));
    localBranches.forEach(b => {
        if (b.upstream) {
            upstreamMap[b.upstream] = b;
        }
    });

    if (currentTab === 'local') {
        filtered = localBranches;
    } else {
        // Remote: Starting with remote/
        filtered = allBranches.filter(b => repoRemotes.some(r => b.name.startsWith(r + '/')));
    }
    
    document.getElementById('total-count').innerText = `共 ${filtered.length} 个分支`;

    if (filtered.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="text-center py-4 text-muted">无匹配分支</td></tr>';
        return;
    }

    filtered.forEach(b => {
        const tr = document.createElement('tr');
        if (b.is_current) {
            tr.classList.add('table-success');
        }

        const dateStr = b.date ? new Date(b.date).toLocaleString() : '-';
        const shortHash = b.hash ? b.hash.substring(0, 7) : '-';
        
        // Sync Status UI
        let syncStatus = '';
        let syncBtn = '';
        
        if (currentTab === 'local') {
            if (b.upstream) {
                if (b.is_current) {
                    syncBtn = `<button class="btn btn-outline-warning btn-sm" onclick="syncBranch('${b.name}', true)" title="同步代码 (Pull --rebase)"><i class="bi bi-cloud-download"></i></button>`;
                } else {
                    syncBtn = `<button class="btn btn-outline-warning btn-sm" onclick="syncBranch('${b.name}', false)" title="更新分支 (Fast-forward Only)"><i class="bi bi-cloud-download"></i></button>`;
                }

                if (b.behind > 0) {
                    syncStatus += `<span class="badge bg-danger me-1" title="落后 ${b.behind} 个提交"><i class="bi bi-arrow-down"></i> ${b.behind}</span>`;
                    syncBtn = syncBtn.replace('btn-outline-warning', 'btn-warning');
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
        } else {
            // Remote Tab Logic
            const linkedLocal = upstreamMap[b.name];
            if (linkedLocal) {
                syncStatus = `<span class="badge bg-info text-dark"><i class="bi bi-link-45deg"></i> 已关联本地: ${linkedLocal.name}</span>`;
                // Show sync button for the LINKED LOCAL branch
                if (linkedLocal.is_current) {
                    syncBtn = `<button class="btn btn-outline-warning btn-sm" onclick="syncBranch('${linkedLocal.name}', true)" title="同步本地 ${linkedLocal.name} (Pull --rebase)"><i class="bi bi-cloud-download"></i> 同步本地</button>`;
                } else {
                    syncBtn = `<button class="btn btn-outline-warning btn-sm" onclick="syncBranch('${linkedLocal.name}', false)" title="更新本地 ${linkedLocal.name} (Fast-forward)"><i class="bi bi-cloud-download"></i> 更新本地</button>`;
                }
            } else {
                syncStatus = `<span class="text-muted small">无本地关联</span>`;
                syncBtn = `<button class="btn btn-outline-success btn-sm" onclick="checkoutRemoteBranch('${b.name}')" title="检出为本地分支"><i class="bi bi-download"></i> 新建本地分支</button>`;
            }
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
                    ${!b.is_current && currentTab === 'local' ? `<button class="btn btn-outline-success" onclick="checkoutBranch('${b.name}')" title="切换到此分支"><i class="bi bi-check2-circle"></i></button>` : ''}
                    ${currentTab === 'local' ? `<button class="btn btn-outline-dark" onclick="openPushModal('${b.name}')" title="推送至远端"><i class="bi bi-cloud-upload"></i></button>` : ''}
                    ${currentTab === 'local' ? `<button class="btn btn-outline-warning" onclick="openTagModal('${b.name}')" title="打标签"><i class="bi bi-tag"></i></button>` : ''}
                    ${currentTab === 'local' ? `<button class="btn btn-outline-secondary" onclick="openDetail('${b.name}')" title="详情"><i class="bi bi-info-circle"></i></button>` : ''}
                    ${currentTab === 'local' ? `<button class="btn btn-outline-primary" onclick="openRenameModal('${b.name}')" title="重命名/描述"><i class="bi bi-pencil"></i></button>` : ''}
                    ${currentTab === 'local' && !b.is_current ? `<button class="btn btn-outline-danger" onclick="deleteBranch('${b.name}')" title="删除"><i class="bi bi-trash"></i></button>` : ''}
                </div>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

function openComparePage() {
    window.location.href = `compare.html?repo_key=${repoKey}`;
}

function openCreateModal(baseRef) {
    document.getElementById('createForm').reset();
    if (baseRef) {
        document.querySelector('#createForm input[name="base_ref"]').value = baseRef;
        // Suggest name: strip remote prefix
        const parts = baseRef.split('/');
        if (parts.length > 1) {
             document.querySelector('#createForm input[name="name"]').value = parts.slice(1).join('/');
        }
    }
    new bootstrap.Modal(document.getElementById('createModal')).show();
}

function checkoutRemoteBranch(name) {
    openCreateModal(name);
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
        await request('/branch/create', {
            method: 'POST',
            body: { repo_key: repoKey, name, base_ref: base }
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
        await request('/branch/update', {
            method: 'POST',
            body: { repo_key: repoKey, name: oldName, new_name: newName, desc: desc }
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
        await request('/branch/delete', {
            method: 'POST',
            body: { repo_key: repoKey, name, force: true }
        });
        showToast("删除成功", "success");
        loadBranches();
    } catch (e) {
        // handled
    }
}

async function checkoutBranch(name) {
    if (!confirm(`确定要切换到分支 "${name}" 吗？\n请确保当前工作区已提交，否则可能会失败。`)) return;

    try {
        await request('/branch/checkout', {
            method: 'POST',
            body: { repo_key: repoKey, name }
        });
        showToast(`已切换到 ${name}`, "success");
        loadBranches();
    } catch (e) {
        // Show specific error message from backend (e.g., dirty worktree)
        showToast("切换失败: " + e.message, "error");
    }
}

function openDetail(branchName) {
    window.location.href = `branch_detail.html?repo_key=${repoKey}&branch=${encodeURIComponent(branchName)}`;
}

async function syncBranch(branch, isCurrent = true) {
    const msg = isCurrent 
        ? `确定要同步分支 ${branch} 吗？\n这将执行 git pull --rebase，可能会产生冲突。`
        : `确定要更新分支 ${branch} 吗？\n仅支持 Fast-forward 更新 (本地无额外提交)。`;

    if (!confirm(msg)) return;
    
    try {
        await request('/branch/pull', { method: 'POST', body: { repo_key: repoKey, name: branch } });
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
        const config = await request('/repo/scan', {
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
        await request('/branch/push', {
            method: 'POST',
            body: { repo_key: repoKey, name: branch, remotes }
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

let nextVersions = {};

async function openTagModal(ref) {
    document.getElementById('tagForm').reset();
    document.querySelector('#tagForm input[name="ref"]').value = ref;
    
    // Reset remote select
    const select = document.getElementById('pushTagRemoteSelect');
    select.innerHTML = '<option value="">加载中...</option>';
    document.getElementById('pushTagRemoteDiv').classList.add('d-none');
    
    // Reset Version UI
    document.getElementById('badge-patch').innerText = '...';
    document.getElementById('badge-minor').innerText = '...';
    document.getElementById('badge-major').innerText = '...';
    document.querySelector('input[name="version_type"][value="patch"]').checked = true;
    updateTagName({value: 'patch'}); // Reset to patch

    new bootstrap.Modal(document.getElementById('tagModal')).show();
    
    // Fetch suggested versions
    try {
        const info = await request(`/version/next?repo_key=${repoKey}`);
        if (info) {
            nextVersions = info;
            document.getElementById('badge-patch').innerText = info.next_patch;
            document.getElementById('badge-minor').innerText = info.next_minor;
            document.getElementById('badge-major').innerText = info.next_major;
            
            // Refresh current selection
            const currentType = document.querySelector('input[name="version_type"]:checked');
            if (currentType) updateTagName(currentType);
        }
    } catch(e) {
        console.warn("Failed to fetch suggestions", e);
    }

    // Load remotes in background
    try {
        const config = await request('/repo/scan', {
            method: 'POST',
            body: { path: currentRepo.path }
        });
        select.innerHTML = '';
        if (config.remotes && config.remotes.length > 0) {
            config.remotes.forEach(r => {
                const opt = document.createElement('option');
                opt.value = r.name;
                opt.innerText = r.name;
                select.appendChild(opt);
            });
            // Default to origin if exists
            if (config.remotes.some(r => r.name === 'origin')) {
                select.value = 'origin';
            }
        } else {
            select.innerHTML = '<option value="">无可用远端</option>';
        }
    } catch(e) {}
}

function updateTagName(radio) {
    const input = document.querySelector('input[name="tag_name"]');
    if (radio.value === 'custom') {
        input.readOnly = false;
        input.value = '';
        input.placeholder = '例如 v1.2.3';
        input.focus();
    } else {
        input.readOnly = true;
        if (nextVersions) {
            switch(radio.value) {
                case 'patch': input.value = nextVersions.next_patch || ''; break;
                case 'minor': input.value = nextVersions.next_minor || ''; break;
                case 'major': input.value = nextVersions.next_major || ''; break;
            }
        }
    }
}

function toggleTagPush(checked) {
    const div = document.getElementById('pushTagRemoteDiv');
    if (checked) div.classList.remove('d-none');
    else div.classList.add('d-none');
}

async function submitCreateTag() {
    const form = document.getElementById('tagForm');
    const ref = form.ref.value;
    let tagName = form.tag_name.value.trim();
    const message = form.message.value;
    const push = form.pushTagCheck.checked;
    const remote = form.push_remote.value;
    
    if (!tagName) {
        showToast("请输入标签名", "warning");
        return;
    }
    
    if (push && !remote) {
        showToast("请选择推送的远端", "warning");
        return;
    }
    
    const btn = document.querySelector('#tagModal .btn-primary');
    btn.disabled = true;
    btn.innerText = "创建中...";
    
    try {
        const body = {
            repo_key: repoKey,
            tag_name: tagName,
            ref: ref,
            message: message
        };
        if (push) {
            body.push_remote = remote;
        }
        
        await request('/tag/create', {
            method: 'POST',
            body: body
        });
        
        showToast(push ? "标签已创建并推送" : "标签已创建", "success");
        bootstrap.Modal.getInstance(document.getElementById('tagModal')).hide();
        
    } catch (e) {
        // handled
    } finally {
        btn.disabled = false;
        btn.innerText = "创建";
    }
}
