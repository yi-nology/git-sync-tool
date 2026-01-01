// repos.js - Repository Management Logic

let allRepos = [];
let clonePollInterval = null;
let currentRepoAuths = {};
let currentParentPath = "";
let currentPath = "";
let fileBrowserTarget = 'add';

document.addEventListener('DOMContentLoaded', () => {
    initToastContainer();
    loadRepos();
});

async function loadRepos() {
    try {
        const data = await request('/repos');
        allRepos = data;
        const tbody = document.getElementById('repo-list');
        tbody.innerHTML = '';

        if (!data || data.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="text-center text-muted">暂无仓库，请点击右上角注册</td></tr>';
            return;
        }

        data.forEach(repo => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${repo.id}</td>
                <td>${repo.name}</td>
                <td><code class="text-muted small">${repo.path}</code></td>
                <td><span class="badge bg-secondary">${repo.config_source === 'database' ? '数据库' : '本地文件'}</span></td>
                <td><span class="text-break small">${repo.remote_url || '-'}</span></td>
                <td>
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-outline-secondary" onclick="openRepoDetailModal(${repo.id})" title="详情"><i class="bi bi-info-circle"></i></button>
                        <button class="btn btn-outline-primary" onclick="openEditRepoModal(${repo.id})" title="编辑"><i class="bi bi-pencil"></i></button>
                        <a class="btn btn-outline-dark" href="repo_sync.html?repo_key=${repo.key}" title="同步配置"><i class="bi bi-arrow-repeat"></i></a>
                        <button class="btn btn-outline-info" onclick="openRepoHistoryModal('${repo.key}')" title="同步历史"><i class="bi bi-clock-history"></i></button>
                        <button class="btn btn-outline-danger" onclick="deleteRepo(${repo.id})" title="删除"><i class="bi bi-trash"></i></button>
                    </div>
                </td>
            `;
            tbody.appendChild(tr);
        });
    } catch (e) {
        // request.js handles logging and toast
    }
}

async function openRepoDetailModal(id) {
    const repo = allRepos.find(r => r.id === id);
    if (!repo) return;
    
    document.getElementById('detailName').innerText = repo.name;
    document.getElementById('detailSource').innerText = repo.config_source;
    document.getElementById('detailPath').innerText = repo.path;
    
    document.getElementById('detailRemotesList').innerHTML = '<tr><td colspan="4" class="text-center text-muted">加载中...</td></tr>';
    document.getElementById('detailBranchesList').innerHTML = '<span class="text-muted">加载中...</span>';
    
    new bootstrap.Modal(document.getElementById('repoDetailModal')).show();
    
    try {
        const config = await request('/repos/scan', {
            method: 'POST',
            body: { path: repo.path }
        });
        
        // Remotes
        const tbody = document.getElementById('detailRemotesList');
        tbody.innerHTML = '';
        if (!config.remotes || config.remotes.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" class="text-center text-muted">无远程仓库配置</td></tr>';
        } else {
            config.remotes.forEach(r => {
                tbody.innerHTML += `
                    <tr>
                        <td>${r.name}</td>
                        <td class="small text-break">${r.fetch_url}</td>
                        <td class="small text-break">${r.push_url || '-'}</td>
                        <td>${r.is_mirror ? '<i class="bi bi-check text-success"></i>' : '-'}</td>
                    </tr>
                `;
            });
        }
        
        // Branches
        const branchContainer = document.getElementById('detailBranchesList');
        branchContainer.innerHTML = '';
        if (!config.branches || config.branches.length === 0) {
            branchContainer.innerText = '无分支';
        } else {
            config.branches.forEach(b => {
                const badge = document.createElement('span');
                badge.className = 'badge bg-light text-dark border';
                badge.innerText = b.name;
                if (b.is_head) badge.classList.add('bg-success', 'text-white');
                branchContainer.appendChild(badge);
            });
        }
    } catch (e) {
        document.getElementById('detailRemotesList').innerHTML = '<tr><td colspan="4" class="text-center text-danger">加载失败</td></tr>';
        document.getElementById('detailBranchesList').innerHTML = '<span class="text-danger">加载失败</span>';
    }
}

async function openRepoHistoryModal(repoKey) {
    new bootstrap.Modal(document.getElementById('repoHistoryModal')).show();
    reloadHistoryTable(repoKey);
}

async function deleteHistory(id, repoKey) {
    if (!confirm("确定删除这条历史记录吗？")) return;
    
    try {
        await request(`/sync/history/${id}`, { method: 'DELETE' });
        showToast('历史记录已删除', 'success');
        reloadHistoryTable(repoKey);
    } catch (e) {
        // Handled by request
    }
}

async function reloadHistoryTable(repoKey) {
    const tbody = document.getElementById('repoHistoryList');
    tbody.innerHTML = '<tr><td colspan="6" class="text-center">加载中...</td></tr>';
    
    try {
        const history = await request(`/sync/history?repo_key=${repoKey}`);
        
        tbody.innerHTML = '';
        if (!history || history.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="text-center text-muted">暂无同步记录</td></tr>';
            return;
        }
        
        history.forEach(h => {
            const duration = h.end_time ? (new Date(h.end_time) - new Date(h.start_time)) + 'ms' : '-';
            let taskInfo = '-';
            if (h.Task) {
                taskInfo = `${h.Task.source_remote}/${h.Task.source_branch} -> ${h.Task.target_remote}`;
            }
            
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${new Date(h.start_time).toLocaleString()}</td>
                <td class="small">${taskInfo}</td>
                <td><span class="badge bg-${getStatusColor(h.status)}">${h.status}</span></td>
                <td>${duration}</td>
                <td><button class="btn btn-xs btn-link p-0" onclick='showLog(${JSON.stringify(h.details || "")})'>日志</button></td>
                <td>
                    <button class="btn btn-xs btn-outline-danger" onclick="deleteHistory(${h.id}, '${repoKey}')" title="删除记录">
                        <i class="bi bi-trash"></i>
                    </button>
                </td>
            `;
            tbody.appendChild(tr);
        });
    } catch (e) {
        tbody.innerHTML = '<tr><td colspan="6" class="text-center text-danger">加载错误</td></tr>';
    }
}

function openAddRepoModal() {
    document.getElementById('addRepoForm').reset();
    setRepoMode('local');
    document.getElementById('mode-local-tab').click();
    document.getElementById('scanResultArea').classList.add('d-none');
    
    toggleCloneAuth('none');
    document.getElementById('cloneFormArea').classList.remove('d-none');
    document.getElementById('cloneProgressArea').classList.add('d-none');
    document.querySelector('#addRepoModal .modal-footer').classList.remove('d-none');
    
    loadSSHKeys('cloneSshKeySelect');
    new bootstrap.Modal(document.getElementById('addRepoModal')).show();
}

function setRepoMode(mode) {
    document.getElementById('repoModeInput').value = mode;
}

function autoFillName(url) {
    if (!url) return;
    const parts = url.split('/');
    let last = parts[parts.length - 1];
    if (last.endsWith('.git')) {
        last = last.substring(0, last.length - 4);
    }
    const nameInput = document.getElementById('cloneNameInput');
    if (!nameInput.value) {
        nameInput.value = last;
    }
}

function toggleCloneAuth(type) {
    document.getElementById('clone-auth-ssh').className = type === 'ssh' ? 'd-block' : 'd-none';
    document.getElementById('clone-auth-http').className = type === 'http' ? 'd-block' : 'd-none';
}

async function testCloneConnection() {
    const url = document.getElementById('cloneUrlInput').value;
    if (!url) { showToast('请输入 URL', 'warning'); return; }
    
    const resLabel = document.getElementById('cloneTestResult');
    resLabel.innerText = '测试中...';
    resLabel.className = 'form-text text-muted';
    
    try {
        const data = await request('/git/test-connection', {
            method: 'POST',
            body: { url }
        });
        
        if (data.status === 'success') {
            resLabel.innerText = '连接成功';
            resLabel.className = 'form-text text-success fw-bold';
        } else {
            resLabel.innerText = '失败: ' + data.error;
            resLabel.className = 'form-text text-danger';
        }
    } catch(e) {
        resLabel.innerText = '请求错误';
        resLabel.className = 'form-text text-danger';
    }
}

async function scanRepo(targetPathId = 'localPathInput', resultAreaId = 'scanResultArea', remotesListId = 'remotesList', trackingId = 'trackingBranches') {
    const path = document.getElementById(targetPathId).value;
    if (!path) return;
    
    if (targetPathId === 'localPathInput') {
        const nameInput = document.getElementById('localNameInput');
        if (!nameInput.value) {
            const parts = path.split(path.includes('/') ? '/' : '\\'); 
            const name = parts[parts.length-1] || parts[parts.length-2];
            if (name) nameInput.value = name;
        }
    }

    try {
        const config = await request('/repos/scan', {
            method: 'POST',
            body: { path }
        });
        
        const resultArea = document.getElementById(resultAreaId);
        renderRemotes(config.remotes, remotesListId, currentRepoAuths);
        renderTracking(config.branches, trackingId);
        if (resultArea) resultArea.classList.remove('d-none');
    } catch(e) {
        // Quiet fail on scan or show partial error?
        const resultArea = document.getElementById(resultAreaId);
        if (resultArea) resultArea.classList.add('d-none');
    }
}

function renderRemotes(remotes, containerId = 'remotesList', auths = {}) {
    const tbody = document.getElementById(containerId);
    tbody.innerHTML = '';
    (remotes || []).forEach(r => {
        addRemoteRow(r, containerId, auths[r.name]);
    });
}

function renderTracking(branches, containerId = 'trackingBranches') {
     const container = document.getElementById(containerId);
     container.innerHTML = '';
     (branches || []).filter(b => b.upstream_ref).forEach(b => {
         const span = document.createElement('span');
         span.className = 'badge bg-light text-dark border me-1 mb-1';
         span.innerText = `${b.name} -> ${b.upstream_ref}`;
         container.appendChild(span);
     });
     if (container.innerHTML === '') {
         container.innerText = '无追踪分支';
     }
}

function addRemoteRow(data = {}, containerId = 'remotesList', auth = null) {
    const tbody = document.getElementById(containerId);
    const tr = document.createElement('tr');
    const rowId = 'remote-row-' + Math.random().toString(36).substr(2, 9);
    tr.id = rowId;
    const authData = auth || {type: 'none', key: '', secret: ''};

    tr.innerHTML = `
        <td><input type="text" class="form-control form-control-sm remote-name" value="${data.name || ''}" placeholder="origin"></td>
        <td>
            <input type="text" class="form-control form-control-sm mb-1 remote-fetch" value="${data.fetch_url || ''}" placeholder="Fetch URL">
            <input type="text" class="form-control form-control-sm remote-push text-muted" value="${data.push_url || ''}" placeholder="Push URL (选填)">
        </td>
        <td class="text-center align-middle">
            <div class="form-check d-flex justify-content-center">
                <input class="form-check-input remote-mirror" type="checkbox" ${data.is_mirror ? 'checked' : ''}>
            </div>
        </td>
        <td class="text-center align-middle">
            <button type="button" class="btn btn-sm btn-outline-primary me-1" onclick="testRemoteConnection('${rowId}')" title="测试连接">
                <i class="bi bi-wifi"></i>
            </button>
            <button type="button" class="btn btn-sm btn-outline-dark me-1" onclick="openExecuteSyncModal('${rowId}')" title="执行同步">
                <i class="bi bi-arrow-repeat"></i>
            </button>
            <button type="button" class="btn btn-sm ${authData.type !== 'none' ? 'btn-success' : 'btn-outline-secondary'} me-1 auth-btn" onclick="openRemoteAuthModal('${rowId}')" title="配置认证">
                <i class="bi bi-shield-lock"></i>
            </button>
            <button type="button" class="btn btn-sm btn-outline-danger" onclick="this.closest('tr').remove()"><i class="bi bi-trash"></i></button>
        </td>
    `;
    tr.dataset.auth = JSON.stringify(authData);
    tbody.appendChild(tr);
}

async function testRemoteConnection(rowId) {
    const tr = document.getElementById(rowId);
    const url = tr.querySelector('.remote-fetch').value;
    if (!url) { showToast('请输入 Fetch URL', 'warning'); return; }
    
    const btn = tr.querySelector('.bi-wifi').parentElement;
    const originalIcon = btn.innerHTML;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm"></span>';
    btn.disabled = true;
    
    try {
        const data = await request('/git/test-connection', {
            method: 'POST',
            body: { url }
        });

        if (data.status === 'success') {
            btn.classList.remove('btn-outline-primary', 'btn-danger');
            btn.classList.add('btn-success');
            btn.title = "连接成功";
            showToast('连接成功', 'success');
        } else {
            btn.classList.remove('btn-outline-primary', 'btn-success');
            btn.classList.add('btn-danger');
            btn.title = "失败: " + data.error;
            showToast("连接失败: " + data.error, 'error');
        }
    } catch (e) {
         // handled by request
    } finally {
        btn.innerHTML = '<i class="bi bi-wifi"></i>';
        btn.disabled = false;
    }
}

function openRemoteAuthModal(rowId) {
    const tr = document.getElementById(rowId);
    const name = tr.querySelector('.remote-name').value || 'New Remote';
    const auth = JSON.parse(tr.dataset.auth || '{}');
    
    document.getElementById('authRemoteName').innerText = name;
    document.getElementById('authRemoteRowId').value = rowId;
    
    document.getElementById('remoteAuthType').value = auth.type || 'none';
    toggleRemoteAuthFields();
    
    if (auth.type === 'ssh') {
        document.getElementById('remoteAuthKey').value = auth.key || '';
        document.getElementById('remoteAuthSecretSsh').value = auth.secret || '';
    } else if (auth.type === 'http') {
        document.getElementById('remoteAuthUser').value = auth.key || '';
        document.getElementById('remoteAuthSecretHttp').value = auth.secret || '';
    }
    
    loadSSHKeys('remoteSshKeySelect');
    new bootstrap.Modal(document.getElementById('remoteAuthModal')).show();
}

function saveRemoteAuth() {
    const rowId = document.getElementById('authRemoteRowId').value;
    const type = document.getElementById('remoteAuthType').value;
    let key = '', secret = '';
    
    if (type === 'ssh') {
        key = document.getElementById('remoteAuthKey').value;
        secret = document.getElementById('remoteAuthSecretSsh').value;
    } else if (type === 'http') {
        key = document.getElementById('remoteAuthUser').value;
        secret = document.getElementById('remoteAuthSecretHttp').value;
    }
    
    const auth = {type, key, secret};
    const tr = document.getElementById(rowId);
    tr.dataset.auth = JSON.stringify(auth);
    
    const btn = tr.querySelector('.auth-btn');
    if (type !== 'none') {
        btn.classList.remove('btn-outline-secondary');
        btn.classList.add('btn-success');
    } else {
        btn.classList.remove('btn-success');
        btn.classList.add('btn-outline-secondary');
    }
    
    bootstrap.Modal.getInstance(document.getElementById('remoteAuthModal')).hide();
}

function toggleRemoteAuthFields() {
     const type = document.getElementById('remoteAuthType').value;
     document.getElementById('remote-auth-ssh').className = type === 'ssh' ? 'd-block' : 'd-none';
     document.getElementById('remote-auth-http').className = type === 'http' ? 'd-block' : 'd-none';
}

function openExecuteSyncModal(rowId) {
    const tr = document.getElementById(rowId);
    const name = tr.querySelector('.remote-name').value;
    if (!name) { showToast('请先填写 Remote 名称', 'warning'); return; }
    
    // In Add/Edit mode, we might not have a saved repo ID accessible easily if it's new
    // But this button is only inside the Edit Modal usually or Add Modal?
    // If Add Modal, we can't sync yet.
    // Check if we are in Edit Modal
    const repoIdInput = document.querySelector('#editRepoForm input[name=id]');
    if (!repoIdInput || !repoIdInput.value) {
        showToast('请先保存仓库后再执行同步操作', 'warning');
        return;
    }

    openSyncModalCommon(repoIdInput.value, name, [name]);
}

function openSyncModalCommon(repoId, defaultTargetRemote, allRemotes) {
    document.getElementById('execSyncRepoId').value = repoId;
    
    const srcSelect = document.getElementById('execSourceSelect');
    const tgtSelect = document.getElementById('execTargetSelect');
    
    const options = `<option value="local">Local (本地)</option>` + 
                    allRemotes.map(r => `<option value="${r}">${r}</option>`).join('');
    
    srcSelect.innerHTML = options;
    tgtSelect.innerHTML = options;
    
    srcSelect.value = 'local';
    tgtSelect.value = defaultTargetRemote || (allRemotes.length > 0 ? allRemotes[0] : 'local');
    
    updateExecSyncUI();
    
    new bootstrap.Modal(document.getElementById('executeSyncModal')).show();
}

function updateExecSyncUI() {
     const src = document.getElementById('execSourceSelect').value;
     const tgt = document.getElementById('execTargetSelect').value;
     const hint = document.getElementById('execSyncHint');
     
     if (src === tgt) {
         hint.className = 'alert alert-warning py-2 small';
         hint.innerHTML = '<i class="bi bi-exclamation-triangle"></i> 源和目标相同，请选择不同的仓库。';
         return;
     }
     
     hint.className = 'alert alert-info py-2 small';
     
     if (src === 'local') {
         hint.innerHTML = `<i class="bi bi-arrow-right-circle"></i> <b>Push (推送)</b>: 将本地更改推送到远程 <b>${tgt}</b>。`;
     } else if (tgt === 'local') {
         hint.innerHTML = `<i class="bi bi-arrow-down-circle"></i> <b>Fetch (拉取)</b>: 从远程 <b>${src}</b> 获取更新到本地 (不会自动合并工作区)。`;
     } else {
         hint.innerHTML = `<i class="bi bi-arrow-repeat"></i> <b>Sync (同步)</b>: 将 <b>${src}</b> 的代码同步到 <b>${tgt}</b>。`;
     }
}

async function submitExecuteSync() {
    const repoId = document.getElementById('execSyncRepoId').value;
    const src = document.getElementById('execSourceSelect').value;
    const tgt = document.getElementById('execTargetSelect').value;
    const force = document.getElementById('execForce').checked;
    
    if (src === tgt) {
        showToast("源和目标不能相同", 'warning');
        return;
    }

    const req = {
        repo_id: parseInt(repoId),
        source_remote: src,
        source_branch: document.getElementById('execSourceBranch').value,
        target_remote: tgt,
        target_branch: document.getElementById('execTargetBranch').value,
        push_options: force ? "--force" : ""
    };

    const btn = document.querySelector('#executeSyncModal .modal-footer .btn-primary');
    btn.disabled = true;
    btn.innerText = "提交中...";

    try {
        const data = await request('/sync/execute', {
            method: 'POST',
            body: req
        });
        
        showToast("同步任务已开始，Task Key: " + data.task_key, 'success');
        bootstrap.Modal.getInstance(document.getElementById('executeSyncModal')).hide();
    } catch (e) {
        // handled by request
    } finally {
        btn.disabled = false;
        btn.innerText = "执行";
    }
}

async function submitRepo() {
    const form = document.getElementById('addRepoForm');
    const mode = document.getElementById('repoModeInput').value;
    const btn = document.querySelector('#addRepoModal .modal-footer .btn-primary');
    
    btn.disabled = true;

    try {
        if (mode === 'local') {
            const remotes = [];
            const remoteAuths = {};
            
            document.querySelectorAll('#remotesList tr').forEach(tr => {
                const name = tr.querySelector('.remote-name').value;
                const fetch = tr.querySelector('.remote-fetch').value;
                const push = tr.querySelector('.remote-push').value;
                const mirror = tr.querySelector('.remote-mirror').checked;
                if (name && fetch) {
                    remotes.push({
                        name: name,
                        fetch_url: fetch,
                        push_url: push || fetch,
                        is_mirror: mirror
                    });
                    
                    const auth = JSON.parse(tr.dataset.auth || '{}');
                    if (auth.type && auth.type !== 'none') {
                        remoteAuths[name] = auth;
                    }
                }
            });

            const data = {
                name: document.getElementById('localNameInput').value,
                path: document.getElementById('localPathInput').value,
                config_source: document.querySelector('select[name=local_config_source]').value,
                auth_type: 'none',
                remotes: remotes,
                remote_auths: remoteAuths
            };
            
            if (!data.name || !data.path) throw new Error("请填写完整信息");

            btn.innerText = '保存中...';
            await request('/repos', {
                method: 'POST',
                body: data
            });
            
            showToast('仓库注册成功', 'success');
            bootstrap.Modal.getInstance(document.getElementById('addRepoModal')).hide();
            loadRepos();
            form.reset();
            btn.disabled = false;
            btn.innerText = '保存';

        } else {
            const data = {
                remote_url: form.clone_url.value,
                local_path: form.clone_path.value,
                name: form.clone_name.value,
                auth_type: form.clone_auth_type.value,
                config_source: 'database'
            };
            
            if (data.auth_type === 'ssh') {
                data.auth_key = form.clone_auth_key.value;
                data.auth_secret = form.clone_auth_secret.value;
            } else if (data.auth_type === 'http') {
                data.auth_key = form.clone_auth_user.value;
                data.auth_secret = form.clone_auth_pass.value;
            }
            
            if (!data.remote_url || !data.local_path) throw new Error("请填写完整克隆信息");

            btn.innerText = '请求中...';
            const result = await request('/repos/clone', {
                method: 'POST',
                body: data
            });
            
            startClonePolling(result.task_id);
        }
    } catch (e) {
        btn.disabled = false;
        btn.innerText = '保存';
    }
}

function startClonePolling(taskId) {
    document.getElementById('cloneFormArea').classList.add('d-none');
    document.getElementById('cloneProgressArea').classList.remove('d-none');
    document.querySelector('#addRepoModal .modal-footer').classList.add('d-none');

    const logBox = document.getElementById('cloneLogs');
    const statusText = document.getElementById('cloneStatusText');
    logBox.innerHTML = '';

    if (clonePollInterval) clearInterval(clonePollInterval);
    
    clonePollInterval = setInterval(async () => {
        try {
            const task = await request(`/tasks/${taskId}`);
            
            if (task.progress) {
                 logBox.innerHTML = task.progress.join('<br>');
                 logBox.scrollTop = logBox.scrollHeight;
            }

            if (task.status === 'success') {
                clearInterval(clonePollInterval);
                statusText.innerText = '克隆成功！正在跳转...';
                statusText.className = 'text-success fw-bold';
                
                setTimeout(() => {
                     document.getElementById('cloneFormArea').classList.remove('d-none');
                     document.getElementById('cloneProgressArea').classList.add('d-none');
                     document.querySelector('#addRepoModal .modal-footer').classList.remove('d-none');
                     const btn = document.querySelector('#addRepoModal .modal-footer .btn-primary');
                     btn.disabled = false;
                     btn.innerText = '保存';
                     
                     setRepoMode('local');
                     document.getElementById('mode-local-tab').click();
                     document.getElementById('localPathInput').value = document.getElementById('clonePathInput').value; 
                     scanRepo();
                     
                }, 1500);
            } else if (task.status === 'failed') {
                clearInterval(clonePollInterval);
                statusText.innerText = '克隆失败: ' + task.error;
                statusText.className = 'text-danger fw-bold';
                 document.querySelector('#addRepoModal .modal-footer').classList.remove('d-none');
                 const btn = document.querySelector('#addRepoModal .modal-footer .btn-primary');
                 btn.disabled = false;
                 btn.innerText = '重试';
                 btn.onclick = () => {
                     document.getElementById('cloneFormArea').classList.remove('d-none');
                     document.getElementById('cloneProgressArea').classList.add('d-none');
                     btn.innerText = '保存';
                     btn.onclick = submitRepo; 
                 };
            }
        } catch(e) {
            // Ignore polling errors
        }
    }, 1000);
}

// File Browser
function openFileBrowser(target = 'add') {
    fileBrowserTarget = target;
    new bootstrap.Modal(document.getElementById('fileBrowserModal')).show();
    loadDirs(""); 
}

function searchDirs() {
    const query = document.getElementById('fileSearchInput').value;
    loadDirs(currentPath, query);
}

async function loadDirs(path, search = "") {
    try {
        const data = await request(`/system/dirs?path=${encodeURIComponent(path)}&search=${encodeURIComponent(search)}`);
        
        currentPath = data.current;
        currentParentPath = data.parent;
        
        document.getElementById('currentPathDisplay').value = data.current;
        
        const list = document.getElementById('dirList');
        list.innerHTML = '';
        
        data.dirs.forEach(d => {
            const item = document.createElement('button');
            item.className = 'list-group-item list-group-item-action';
            item.innerHTML = `<i class="bi bi-folder"></i> ${d.name}`;
            item.onclick = () => {
                if (search) {
                     document.getElementById('fileSearchInput').value = "";
                     loadDirs(d.path, "");
                } else {
                     loadDirs(d.path);
                }
            };
            list.appendChild(item);
        });
    } catch(e) {}
}

function selectCurrentDir() {
    let inputId = 'localPathInput';
    if (fileBrowserTarget === 'edit') inputId = 'editLocalPathInput';
    else if (fileBrowserTarget === 'clone_parent') inputId = 'clonePathInput';
    
    document.getElementById(inputId).value = currentPath;
    bootstrap.Modal.getInstance(document.getElementById('fileBrowserModal')).hide();
    
    if (fileBrowserTarget === 'add') {
        scanRepo();
    }
}

function openEditRepoModal(id) {
    const repo = allRepos.find(r => r.id === id);
    if (!repo) return;
    
    const form = document.getElementById('editRepoForm');
    form.id.value = repo.id;
    form.name.value = repo.name;
    form.path.value = repo.path;
    form.config_source.value = repo.config_source || 'local';

    new bootstrap.Modal(document.getElementById('editRepoModal')).show();
    
    currentRepoAuths = repo.remote_auths || {};
    
    // Trigger scan for edit modal
    scanRepo('editLocalPathInput', 'editScanResultArea', 'editRemotesList', 'editTrackingBranches');
}

async function updateRepo() {
    const form = document.getElementById('editRepoForm');

    // Collect Remotes from Edit Table
    const remotes = [];
    const remoteAuths = {};

    document.querySelectorAll('#editRemotesList tr').forEach(tr => {
        const name = tr.querySelector('.remote-name').value;
        const fetch = tr.querySelector('.remote-fetch').value;
        const push = tr.querySelector('.remote-push').value;
        const mirror = tr.querySelector('.remote-mirror').checked;
        
        if (name && fetch) {
            remotes.push({
                name: name,
                fetch_url: fetch,
                push_url: push || fetch,
                is_mirror: mirror
            });
            
            const auth = JSON.parse(tr.dataset.auth || '{}');
            if (auth.type && auth.type !== 'none') {
                remoteAuths[name] = auth;
            }
        }
    });

    const data = {
        name: form.name.value,
        path: form.path.value,
        config_source: form.config_source.value,
        remotes: remotes,
        remote_auths: remoteAuths
    };

    try {
        await request(`/repos/${form.id.value}`, {
            method: 'PUT',
            body: data
        });
        
        showToast('仓库更新成功', 'success');
        bootstrap.Modal.getInstance(document.getElementById('editRepoModal')).hide();
        loadRepos();
    } catch (e) {
        // handled by request
    }
}

async function deleteRepo(id) {
    if (!confirm('确定要删除这个仓库吗？如果被同步任务使用将无法删除。')) {
        return;
    }
    try {
        await request(`/repos/${id}`, { method: 'DELETE' });
        showToast('仓库已删除', 'success');
        loadRepos();
    } catch (e) {
        // handled by request
    }
}

// Global functions for window access if needed (usually modules handle this better)
window.loadSSHKeys = async function(selectId) {
    try {
        const keys = await request('/system/ssh-keys');
        const select = document.getElementById(selectId);
        const current = select.value;
        select.innerHTML = '<option value="">手动输入路径...</option>' + 
            keys.map(k => `<option value="${k}">${k}</option>`).join('');
        if (current) select.value = current;
    } catch(e) {}
};
