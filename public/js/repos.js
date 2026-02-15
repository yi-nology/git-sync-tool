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
        const tbody = document.getElementById('repo-list');
        if (!tbody) return; // Exit if not on repos page

        const data = await request('/repo/list');
        allRepos = data;
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
                        <a class="btn btn-outline-success" href="project.html?repo_key=${repo.key}" title="项目"><i class="bi bi-kanban"></i></a>
                        <button class="btn btn-outline-danger" onclick="deleteRepo('${repo.key}')" title="删除"><i class="bi bi-trash"></i></button>
                    </div>
                </td>
            `;
            tbody.appendChild(tr);
        });
    } catch (e) {
        // request.js handles logging and toast
    }
}

// Removed openRepoDetailModal

// Removed openRepoHistoryModal, deleteHistory, reloadHistoryTable

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
        const data = await request('/system/test-connection', {
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
        const config = await request('/repo/scan', {
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
        const data = await request('/system/test-connection', {
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
    
    // Handle SSH source (local vs database)
    const source = auth.source || 'local';
    document.getElementById('remoteSshSource').value = source;
    toggleSshSource();
    
    if (auth.type === 'ssh') {
        if (source === 'database' && auth.ssh_key_id) {
            document.getElementById('remoteDbSshKeySelect').value = auth.ssh_key_id;
        } else {
            document.getElementById('remoteAuthKey').value = auth.key || '';
            document.getElementById('remoteAuthSecretSsh').value = auth.secret || '';
        }
    } else if (auth.type === 'http') {
        document.getElementById('remoteAuthUser').value = auth.key || '';
        document.getElementById('remoteAuthSecretHttp').value = auth.secret || '';
    }
    
    loadSSHKeys('remoteSshKeySelect');
    loadDbSSHKeys();
    new bootstrap.Modal(document.getElementById('remoteAuthModal')).show();
}

function saveRemoteAuth() {
    const rowId = document.getElementById('authRemoteRowId').value;
    const type = document.getElementById('remoteAuthType').value;
    let key = '', secret = '', source = 'local', ssh_key_id = 0;
    
    if (type === 'ssh') {
        source = document.getElementById('remoteSshSource').value;
        if (source === 'database') {
            const dbSelect = document.getElementById('remoteDbSshKeySelect');
            ssh_key_id = parseInt(dbSelect.value) || 0;
            if (!ssh_key_id) {
                showToast('请选择数据库密钥', 'warning');
                return;
            }
            // Store key name for display
            const selectedOpt = dbSelect.options[dbSelect.selectedIndex];
            key = selectedOpt.dataset.name || '';
        } else {
            key = document.getElementById('remoteAuthKey').value;
            secret = document.getElementById('remoteAuthSecretSsh').value;
        }
    } else if (type === 'http') {
        key = document.getElementById('remoteAuthUser').value;
        secret = document.getElementById('remoteAuthSecretHttp').value;
    }
    
    const auth = {type, key, secret, source, ssh_key_id};
    const tr = document.getElementById(rowId);
    tr.dataset.auth = JSON.stringify(auth);
    
    const btn = tr.querySelector('.auth-btn');
    if (type !== 'none') {
        btn.classList.remove('btn-outline-secondary');
        btn.classList.add('btn-success');
        // Show key info on button
        if (type === 'ssh' && source === 'database' && key) {
            btn.title = `数据库密钥: ${key}`;
        } else if (type === 'ssh' && key) {
            btn.title = `本地密钥: ${key}`;
        } else if (type === 'http' && key) {
            btn.title = `HTTP: ${key}`;
        }
    } else {
        btn.classList.remove('btn-success');
        btn.classList.add('btn-outline-secondary');
        btn.title = '配置认证';
    }
    
    bootstrap.Modal.getInstance(document.getElementById('remoteAuthModal')).hide();
}

function toggleRemoteAuthFields() {
     const type = document.getElementById('remoteAuthType').value;
     document.getElementById('remote-auth-ssh').className = type === 'ssh' ? 'd-block' : 'd-none';
     document.getElementById('remote-auth-http').className = type === 'http' ? 'd-block' : 'd-none';
}

function toggleSshSource() {
    const source = document.getElementById('remoteSshSource').value;
    document.getElementById('ssh-source-local').className = source === 'local' ? 'd-block' : 'd-none';
    document.getElementById('ssh-source-database').className = source === 'database' ? 'd-block' : 'd-none';
}

let dbSSHKeys = [];

async function loadDbSSHKeys() {
    try {
        const data = await request('/system/ssh-keys');
        dbSSHKeys = data || [];
        const select = document.getElementById('remoteDbSshKeySelect');
        select.innerHTML = '<option value="">-- 请选择 --</option>';
        dbSSHKeys.forEach(key => {
            const opt = document.createElement('option');
            opt.value = key.ID;
            opt.textContent = `${key.name} (${key.key_type || 'unknown'})`;
            opt.dataset.name = key.name;
            select.appendChild(opt);
        });
    } catch(e) {
        console.error('加载数据库SSH密钥失败', e);
    }
}

function openExecuteSyncModal(rowId) {
    const tr = document.getElementById(rowId);
    const name = tr.querySelector('.remote-name').value;
    if (!name) { showToast('请先填写 Remote 名称', 'warning'); return; }
    
    // In Add/Edit mode, we might not have a saved repo ID accessible easily if it's new
    // But this button is only inside the Edit Modal usually or Add Modal?
    // If Add Modal, we can't sync yet.
    // Check if we are in Edit Modal
    // Note: We need repo KEY now, not ID.
    // The editRepoForm input[name=id] actually stores ID. We need to find Key from allRepos.
    const repoIdInput = document.querySelector('#editRepoForm input[name=id]');
    if (!repoIdInput || !repoIdInput.value) {
        showToast('请先保存仓库后再执行同步操作', 'warning');
        return;
    }
    
    const repo = allRepos.find(r => r.id == repoIdInput.value);
    if (!repo) {
        showToast('仓库信息丢失，请刷新页面', 'error');
        return;
    }

    openSyncModalCommon(repo.key, name, [name]);
}

function openSyncModalCommon(repoKey, defaultTargetRemote, allRemotes) {
    document.getElementById('execSyncRepoId').value = repoKey;
    
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
    const repoKey = document.getElementById('execSyncRepoId').value;
    const src = document.getElementById('execSourceSelect').value;
    const tgt = document.getElementById('execTargetSelect').value;
    const force = document.getElementById('execForce').checked;
    
    if (src === tgt) {
        showToast("源和目标不能相同", 'warning');
        return;
    }

    const req = {
        repo_key: repoKey,
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
            await request('/repo/create', {
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
            const result = await request('/repo/clone', {
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
            const task = await request(`/repo/task?task_id=${taskId}`);
            
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
    // Using repo ID here but backend endpoint should be updated to use key? 
    // Wait, the API endpoint is PUT /repos/:key now.
    // We need repo key. But edit form only has ID.
    // Let's find repo from allRepos to get key.
    const repo = allRepos.find(r => r.id == form.id.value);
    if (!repo) return;

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
        await request('/repo/update', {
            method: 'POST',
            body: { ...data, key: repo.key }
        });
        
        showToast('仓库更新成功', 'success');
        bootstrap.Modal.getInstance(document.getElementById('editRepoModal')).hide();
        loadRepos();
    } catch (e) {
        // handled by request
    }
}

async function deleteRepo(key) {
    if (!confirm('确定要删除这个仓库吗？如果被同步任务使用将无法删除。')) {
        return;
    }
    try {
        await request('/repo/delete', { method: 'POST', body: { key } });
        showToast('仓库已删除', 'success');
        loadRepos();
    } catch (e) {
        // handled by request
    }
}

function openRepo(key) {
    window.location.href = `project.html?repo_key=${key}`;
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
