const urlParams = new URLSearchParams(window.location.search);
const repoId = urlParams.get('repo_id');
let currentDiffView = 'side-by-side';
let currentFile = null;
let currentDiffData = null;

document.addEventListener('DOMContentLoaded', () => {
    initToastContainer();
    if (!repoId) {
        showToast("缺少 repo_id 参数", "error");
        return;
    }
    document.getElementById('repo-link').href = `branches.html?repo_id=${repoId}`;
    loadBranches();
});

async function loadBranches() {
    try {
        const res = await request(`/repos/${repoId}/branches?page_size=1000`);
        const branches = res.list || [];
        
        const srcSelect = document.getElementById('sourceBranch');
        const tgtSelect = document.getElementById('targetBranch');
        
        let options = '<option value="">选择分支...</option>';
        branches.forEach(b => {
            options += `<option value="${b.name}">${b.name}</option>`;
        });
        
        srcSelect.innerHTML = options;
        tgtSelect.innerHTML = options;
        
        // Auto select if params exist
        const pSource = urlParams.get('source'); // Changed from 'base' to 'source' to match intent
        const pTarget = urlParams.get('target');
        
        // My URL param convention: compare.html?source=feature&target=main
        // API param convention: base=main&target=feature (diff from base to target)
        
        if (pSource) srcSelect.value = pSource;
        if (pTarget) tgtSelect.value = pTarget;
        
        // Default: If Main exists, set as Target
        if (!tgtSelect.value) {
            if (branches.find(b => b.name === 'main')) tgtSelect.value = 'main';
            else if (branches.find(b => b.name === 'master')) tgtSelect.value = 'master';
        }
        
        // If we have both, trigger check automatically
        if (srcSelect.value && tgtSelect.value) {
            checkDiff();
        }

    } catch (e) {
        showToast("加载分支失败", "error");
    }
}

async function checkDiff(force = false) {
    const source = document.getElementById('sourceBranch').value;
    const target = document.getElementById('targetBranch').value;
    
    if (!source || !target) return;
    if (source === target) {
        if (force) showToast("源分支和目标分支相同", "warning");
        return;
    }

    const btn = document.getElementById('compareBtn');
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm"></span> 对比中';
    
    document.getElementById('empty-state').classList.add('d-none');
    document.getElementById('diff-stats').classList.add('d-none');
    document.getElementById('diff-content-area').classList.add('d-none');

    try {
        // 1. Get Stats & File List
        // Note: API compare expects base and target. 
        // Git diff base target means diff FROM base TO target. 
        // If we want to see what Source adds to Target, we should do `git diff target source`.
        // My API uses `git diff base target`.
        // So I should pass base=target, target=source to see changes introduced by source.
        
        const res = await request(`/repos/${repoId}/compare?base=${encodeURIComponent(target)}&target=${encodeURIComponent(source)}`);
        
        // Render Stats
        document.getElementById('stat-files').innerText = res.stat.FilesChanged;
        document.getElementById('stat-insert').innerText = res.stat.Insertions;
        document.getElementById('stat-delete').innerText = res.stat.Deletions;
        document.getElementById('diff-stats').classList.remove('d-none');
        
        // Render Files
        const fileList = document.getElementById('file-list');
        fileList.innerHTML = '';
        
        if (res.files.length === 0) {
            fileList.innerHTML = '<div class="text-muted p-3">无差异</div>';
        } else {
            res.files.forEach((f, idx) => {
                const item = document.createElement('button');
                item.className = 'list-group-item list-group-item-action small';
                // Status color
                let color = 'secondary';
                if (f.status.startsWith('A')) color = 'success';
                else if (f.status.startsWith('M')) color = 'primary';
                else if (f.status.startsWith('D')) color = 'danger';
                
                item.innerHTML = `<span class="badge bg-${color} me-2" style="width:20px">${f.status.substring(0,1)}</span> ${f.path}`;
                item.onclick = () => loadFileDiff(f.path);
                fileList.appendChild(item);
                
                // Load first file automatically
                if (idx === 0) loadFileDiff(f.path);
            });
        }
        
        document.getElementById('diff-content-area').classList.remove('d-none');
        document.getElementById('mergeBtn').disabled = false;

    } catch (e) {
        showToast("对比失败: " + e.message, "error");
    } finally {
        btn.disabled = false;
        btn.innerHTML = '<i class="bi bi-arrow-left-right"></i> 对比';
    }
}

async function loadFileDiff(filePath) {
    const source = document.getElementById('sourceBranch').value;
    const target = document.getElementById('targetBranch').value;
    
    currentFile = filePath;
    document.getElementById('current-file-path').innerText = filePath;
    document.getElementById('diff-viewer').innerHTML = '<div class="text-center py-5"><div class="spinner-border text-secondary"></div></div>';

    // Highlight selected file
    document.querySelectorAll('#file-list button').forEach(b => b.classList.remove('active'));
    // Find button with text containing path (simple approximation)
    // Better to store ID, but for now ok.
    
    try {
        const res = await request(`/repos/${repoId}/diff?base=${encodeURIComponent(target)}&target=${encodeURIComponent(source)}&file=${encodeURIComponent(filePath)}`);
        
        const diffString = res.diff;
        if (!diffString) {
             document.getElementById('diff-viewer').innerHTML = '<div class="text-center py-5 text-muted">无内容差异 (可能是二进制文件)</div>';
             return;
        }

        const targetElement = document.getElementById('diff-viewer');
        const configuration = {
            drawFileList: false,
            fileListToggle: false,
            fileContentToggle: false,
            matching: 'lines',
            outputFormat: currentDiffView,
            synchronisedScroll: true,
            highlight: true,
            renderNothingWhenEmpty: false,
        };
        
        const diff2htmlUi = new Diff2HtmlUI(targetElement, diffString, configuration);
        diff2htmlUi.draw();
        
    } catch (e) {
        document.getElementById('diff-viewer').innerHTML = `<div class="text-center py-5 text-danger">加载失败: ${e.message}</div>`;
    }
}

function changeDiffView(view) {
    currentDiffView = view;
    if (currentFile) loadFileDiff(currentFile);
}

function downloadPatch() {
    const source = document.getElementById('sourceBranch').value;
    const target = document.getElementById('targetBranch').value;
    if (!source || !target) return;
    
    window.open(`/api/repos/${repoId}/patch?base=${encodeURIComponent(target)}&target=${encodeURIComponent(source)}`, '_blank');
}

// Merge Logic
async function openMergeModal() {
    const source = document.getElementById('sourceBranch').value;
    const target = document.getElementById('targetBranch').value;
    
    document.getElementById('mergeSource').innerText = source;
    document.getElementById('mergeTarget').innerText = target;
    document.getElementById('mergeForm').message.value = `Merge branch '${source}' into ${target}`;
    
    document.querySelectorAll('.source-name').forEach(e => e.innerText = source);
    document.querySelectorAll('.target-name').forEach(e => e.innerText = target);

    // Reset UI
    document.getElementById('mergeCheckResult').classList.remove('d-none');
    document.getElementById('mergeForm').classList.add('d-none');
    document.getElementById('conflictAlert').classList.add('d-none');
    document.getElementById('confirmMergeBtn').disabled = true;

    new bootstrap.Modal(document.getElementById('mergeModal')).show();
    
    // Perform Dry Run
    try {
        const res = await request(`/repos/${repoId}/merge/check?base=${encodeURIComponent(source)}&target=${encodeURIComponent(target)}`);
        
        document.getElementById('mergeCheckResult').classList.add('d-none');
        
        if (res.success) {
            document.getElementById('mergeForm').classList.remove('d-none');
            document.getElementById('confirmMergeBtn').disabled = false;
        } else {
            // Conflict
            showConflictUI(res.conflicts, null); // No merge ID yet for simple check
        }
    } catch (e) {
        document.getElementById('mergeCheckResult').innerHTML = `<span class="text-danger">检测失败: ${e.message}</span>`;
    }
}

function showConflictUI(conflicts, reportUrl) {
    const alert = document.getElementById('conflictAlert');
    const list = document.getElementById('conflictList');
    list.innerHTML = '';
    conflicts.forEach(c => {
        list.innerHTML += `<li>${c}</li>`;
    });
    
    if (reportUrl) {
        document.getElementById('conflictReportLink').href = reportUrl;
        document.getElementById('conflictReportLink').classList.remove('d-none');
    } else {
        document.getElementById('conflictReportLink').classList.add('d-none');
    }
    
    alert.classList.remove('d-none');
}

async function submitMerge() {
    const source = document.getElementById('sourceBranch').value;
    const target = document.getElementById('targetBranch').value;
    const message = document.getElementById('mergeForm').message.value;
    
    const btn = document.getElementById('confirmMergeBtn');
    btn.disabled = true;
    btn.innerText = "合并中...";
    
    try {
        const res = await request(`/repos/${repoId}/merge`, {
            method: 'POST',
            body: { source, target, message }
        });
        
        // Handled by request logic? Wait, request.js throws on error?
        // If 409 Conflict, request.js might throw or return data depending on implementation.
        // Assuming request.js returns parsed JSON if status is ok-ish or throws.
        // If my API returns 200 with Code 409 (as implemented in handler), then:
        
        if (res.code === 409) {
            // Backend detected conflict during execution (race condition or dry run missed something)
            document.getElementById('mergeForm').classList.add('d-none');
            showConflictUI(res.data.conflicts, res.data.report_url);
            showToast("合并因冲突中断", "error");
        } else {
            showToast("合并成功", "success");
            bootstrap.Modal.getInstance(document.getElementById('mergeModal')).hide();
            // Refresh diff (should be empty now)
            checkDiff();
        }
        
    } catch (e) {
        // request.js might handle non-200
        showToast("合并请求失败", "error");
    } finally {
        btn.innerText = "确认合并";
    }
}
