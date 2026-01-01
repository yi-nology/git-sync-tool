const urlParams = new URLSearchParams(window.location.search);
const repoKey = urlParams.get('repo_key');
const source = urlParams.get('source');
const target = urlParams.get('target');
const mergeId = urlParams.get('merge_id');

document.addEventListener('DOMContentLoaded', () => {
    initToastContainer();
    if (!repoKey || !source || !target) {
        showToast("缺少必要参数", "error");
        return;
    }

    document.getElementById('merge-id').innerText = mergeId || 'Unknown';
    document.querySelectorAll('.source-branch').forEach(el => el.innerText = source);
    document.querySelectorAll('.target-branch').forEach(el => el.innerText = target);
    
    loadRepoInfo();
    loadConflicts();
});

async function loadRepoInfo() {
    try {
        const repos = await request('/repos');
        const repo = repos.find(r => r.key === repoKey);
        if (repo) {
            document.getElementById('repo-path').innerText = `cd ${repo.path}`;
        }
    } catch (e) {
        // ignore
    }
}

async function loadConflicts() {
    try {
        // Re-run the dry run check to get conflict list
        const res = await request(`/repos/${repoKey}/merge/check?base=${encodeURIComponent(source)}&target=${encodeURIComponent(target)}`);
        
        const list = document.getElementById('conflict-list');
        list.innerHTML = '';
        
        if (res.success) {
            list.innerHTML = '<li class="list-group-item text-center text-success">未检测到冲突 (可能是偶发错误或已解决)</li>';
            return;
        }

        if (res.conflicts && res.conflicts.length > 0) {
            res.conflicts.forEach(file => {
                const li = document.createElement('li');
                li.className = 'list-group-item d-flex justify-content-between align-items-center';
                li.innerHTML = `
                    <span class="font-monospace"><i class="bi bi-file-earmark-code"></i> ${file}</span>
                    <span class="badge bg-danger rounded-pill">Conflict</span>
                `;
                list.appendChild(li);
            });
        } else {
            list.innerHTML = '<li class="list-group-item text-center">未知冲突错误</li>';
        }

    } catch (e) {
        document.getElementById('conflict-list').innerHTML = `<li class="list-group-item text-center text-danger">加载失败: ${e.message}</li>`;
    }
}
