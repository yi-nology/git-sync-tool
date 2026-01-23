const urlParams = new URLSearchParams(window.location.search);
const repoKey = urlParams.get('repo_key');
const branchName = urlParams.get('branch');

document.addEventListener('DOMContentLoaded', () => {
    initToastContainer();
    if (!repoKey || !branchName) {
        showToast("缺少参数", "error");
        return;
    }
    
    document.getElementById('branch-title').innerHTML = `<i class="bi bi-git"></i> ${branchName}`;
    // Breadcrumbs removed
    
    loadData();
});

async function loadData() {
    loadStats();
    loadCommits();
    checkUncommittedChanges();
}

async function checkUncommittedChanges() {
    try {
        const res = await request(`/system/repo/status?repo_key=${repoKey}`);
        const status = res.status;
        const badge = document.getElementById('submit-badge');
        
        // Parse current branch from status (Support English and Chinese)
        let currentBranch = null;
        const matchEn = status.match(/On branch (.+)/);
        const matchZh = status.match(/位于分支 (.+)/);
        
        if (matchEn) {
            currentBranch = matchEn[1].trim();
        } else if (matchZh) {
            currentBranch = matchZh[1].trim();
        }

        // Check for clean status (English or Chinese)
        const isClean = status.includes('nothing to commit') || status.includes('无文件要提交') || status.includes('working tree clean');

        // Only show badge if viewing the current branch AND there are changes
        if (currentBranch === branchName && !isClean) {
            badge.style.display = 'block';
            showToast('检测到未提交的变更，请及时提交', 'info');
        } else {
            badge.style.display = 'none';
        }
    } catch (e) {
        console.error("Failed to check status:", e);
    }
}

async function loadStats() {
    try {
        const data = await request(`/stats/analyze?repo_key=${repoKey}&branch=${encodeURIComponent(branchName)}`);
        
        document.getElementById('stat-lines').innerText = data.total_lines.toLocaleString();
        document.getElementById('stat-authors').innerText = data.authors.length;
        
        // File types count
        const types = new Set();
        data.authors.forEach(a => {
            Object.keys(a.file_types).forEach(t => types.add(t));
        });
        document.getElementById('stat-files').innerText = types.size;

        // Render Authors
        const authorList = document.getElementById('author-list');
        authorList.innerHTML = '';
        
        const sortedAuthors = data.authors.sort((a, b) => b.total_lines - a.total_lines);
        
        sortedAuthors.forEach((a, index) => {
            const li = document.createElement('li');
            li.className = 'list-group-item d-flex justify-content-between align-items-center';
            li.innerHTML = `
                <div>
                    <div class="fw-bold">${a.name}</div>
                    <div class="small text-muted">${a.email}</div>
                </div>
                <span class="badge bg-primary rounded-pill">${a.total_lines.toLocaleString()} 行</span>
            `;
            authorList.appendChild(li);
        });

    } catch (e) {
        document.getElementById('stat-lines').innerText = '-';
        console.error(e);
    }
}

async function loadCommits() {
    const list = document.getElementById('commit-list');
    try {
        // Fetch last 100 commits? API doesn't support limit yet, fetches all. 
        // This might be slow for huge repos.
        const commits = await request(`/stats/commits?repo_key=${repoKey}&branch=${encodeURIComponent(branchName)}`);
        
        document.getElementById('stat-commits').innerText = commits.length.toLocaleString();

        list.innerHTML = '';
        // Show top 20
        commits.slice(0, 20).forEach(c => {
            const date = new Date(c.date).toLocaleString();
            const item = document.createElement('div');
            item.className = 'list-group-item';
            item.innerHTML = `
                <div class="d-flex w-100 justify-content-between">
                    <h6 class="mb-1 text-truncate" style="max-width: 600px;">${c.message}</h6>
                    <small class="text-muted text-nowrap">${date}</small>
                </div>
                <div class="d-flex justify-content-between align-items-center mt-1">
                    <small class="text-muted">
                        <i class="bi bi-person-circle"></i> ${c.author}
                        <span class="ms-2 font-monospace bg-light border px-1 rounded">${c.hash.substring(0, 7)}</span>
                    </small>
                </div>
            `;
            list.appendChild(item);
        });
        
        if (commits.length > 20) {
            const more = document.createElement('div');
            more.className = 'list-group-item text-center text-muted small';
            more.innerText = `... 还有 ${commits.length - 20} 条提交 ...`;
            list.appendChild(more);
        }

    } catch (e) {
        list.innerHTML = `<div class="text-center py-3 text-danger">加载失败: ${e.message}</div>`;
    }
}

function refreshData() {
    loadData();
}

function openCompare() {
    // Navigate to compare page, pre-setting the source branch as current branch
    // Target defaults to main/master handled by compare.js logic
    window.location.href = `compare.html?repo_key=${repoKey}&source=${encodeURIComponent(branchName)}`;
}

async function deleteCurrentBranch() {
    if (!confirm(`确定要删除分支 "${branchName}" 吗？此操作不可撤销！`)) return;
    
    try {
        await request('/branch/delete', {
            method: 'POST',
            body: { repo_key: repoKey, name: branchName, force: true }
        });
        showToast("删除成功", "success");
        setTimeout(() => {
            window.location.href = `branches.html?repo_key=${repoKey}`;
        }, 1000);
    } catch (e) {
        // handled
    }
}

async function pushRemote() {
    if (!confirm(`确定要将分支 "${branchName}" 推送到远程 origin 吗？`)) return;

    try {
        const res = await request('/branch/push', {
            method: 'POST',
            body: {
                repo_key: repoKey,
                name: branchName,
                remotes: ['origin']
            }
        });
        showToast(res.message, "success");
    } catch (e) {
        showToast("推送失败: " + e.message, "error");
    }
}

// Submit Changes Logic
let submitModal = null;

function openSubmitModal() {
    if (!submitModal) {
        submitModal = new bootstrap.Modal(document.getElementById('submitModal'));
    }
    submitModal.show();
    checkRepoStatus();
    loadGitConfig();
}

async function loadGitConfig() {
    try {
        const res = await request(`/system/repo/git-config?repo_key=${repoKey}`);
        document.getElementById('author-name').value = res.name || '';
        document.getElementById('author-email').value = res.email || '';
    } catch (e) {
        console.error("Failed to load git config", e);
    }
}

async function checkRepoStatus() {
    const loading = document.getElementById('submit-loading');
    const content = document.getElementById('submit-content');
    const statusDisplay = document.getElementById('git-status-display');
    const formArea = document.getElementById('submit-form-area');
    const noChanges = document.getElementById('no-changes-alert');
    const btnSubmit = document.getElementById('btn-do-submit');

    loading.style.display = 'block';
    content.style.display = 'none';

    try {
        const res = await request(`/system/repo/status?repo_key=${repoKey}`);
        const status = res.status;
        
        statusDisplay.textContent = status;
        
        if (status.includes('nothing to commit, working tree clean')) {
            formArea.style.display = 'none';
            noChanges.style.display = 'block';
            btnSubmit.disabled = true;
        } else {
            formArea.style.display = 'block';
            noChanges.style.display = 'none';
            btnSubmit.disabled = false;
        }

        loading.style.display = 'none';
        content.style.display = 'block';
    } catch (e) {
        loading.innerHTML = `<span class="text-danger">检查状态失败: ${e.message}</span>`;
    }
}

async function doSubmit() {
    const msg = document.getElementById('commit-message').value;
    const push = document.getElementById('push-after-commit').checked;
    const authorName = document.getElementById('author-name').value;
    const authorEmail = document.getElementById('author-email').value;
    
    if (!msg.trim()) {
        showToast('请输入提交信息', 'error');
        return;
    }

    const btn = document.getElementById('btn-do-submit');
    const originalText = btn.innerHTML;
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm"></span> 提交中...';

    try {
        const res = await request('/system/repo/submit', {
            method: 'POST',
            body: {
                repo_key: repoKey,
                message: msg,
                push: push,
                author_name: authorName,
                author_email: authorEmail
            }
        });

        showToast(res.message, res.warning ? 'warning' : 'success');
        
        // Reset form
        document.getElementById('commit-message').value = '';
        submitModal.hide();
        
        // Refresh data to show new commits
        refreshData();
        checkUncommittedChanges(); // Refresh badge
    } catch (e) {
        showToast(e.message || '提交失败', 'error');
    } finally {
        btn.disabled = false;
        btn.innerHTML = originalText;
    }
}
