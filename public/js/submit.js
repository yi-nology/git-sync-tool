let currentRepoKey = null;

document.addEventListener('DOMContentLoaded', () => {
    loadRepos();
    
    // Check URL params for repo_key
    const urlParams = new URLSearchParams(window.location.search);
    const repoKey = urlParams.get('repo_key');
    if (repoKey) {
        selectRepo(repoKey);
    }
});

async function loadRepos() {
    try {
        const repos = await Request.get('/api/repos');
        const list = document.getElementById('repoList');
        list.innerHTML = '';
        
        repos.data.forEach(repo => {
            const li = document.createElement('li');
            li.innerHTML = `<a class="dropdown-item" href="#" onclick="selectRepo('${repo.key}', '${repo.name}')">${repo.name}</a>`;
            list.appendChild(li);
        });
    } catch (err) {
        Toast.error('Failed to load repositories');
    }
}

function selectRepo(key, name) {
    currentRepoKey = key;
    document.getElementById('repoDropdown').textContent = name || 'Repository Selected';
    document.getElementById('select-repo-msg').style.display = 'none';
    document.getElementById('workspace-area').style.display = 'block';
    
    // Update URL without reload
    const url = new URL(window.location);
    url.searchParams.set('repo_key', key);
    window.history.pushState({}, '', url);

    loadStatus();
}

async function loadStatus() {
    if (!currentRepoKey) return;
    
    const display = document.getElementById('status-display');
    const noChanges = document.getElementById('no-changes-msg');
    const commitCard = document.getElementById('commit-card');
    
    display.textContent = 'Loading status...';
    
    try {
        const res = await Request.get(`/api/repos/${currentRepoKey}/status`);
        const status = res.data.status;
        
        display.textContent = status;
        
        if (status.includes('nothing to commit, working tree clean')) {
            noChanges.style.display = 'block';
            commitCard.style.opacity = '0.5';
            document.getElementById('submit-btn').disabled = true;
        } else {
            noChanges.style.display = 'none';
            commitCard.style.opacity = '1';
            document.getElementById('submit-btn').disabled = false;
        }
    } catch (err) {
        display.textContent = 'Error loading status: ' + err.message;
        Toast.error('Failed to load status');
    }
}

document.getElementById('submit-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    if (!currentRepoKey) return;

    const msg = document.getElementById('commit-msg').value;
    const push = document.getElementById('push-check').checked;
    
    if (!msg.trim()) {
        Toast.error('Please enter a commit message');
        return;
    }

    if (!confirm('Are you sure you want to stage all changes and commit?')) {
        return;
    }

    const btn = document.getElementById('submit-btn');
    const originalText = btn.innerHTML;
    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Processing...';

    try {
        const res = await Request.post(`/api/repos/${currentRepoKey}/submit`, {
            message: msg,
            push: push
        });
        
        Toast.success(res.data.message);
        document.getElementById('commit-msg').value = '';
        loadStatus(); // Refresh status
    } catch (err) {
        Toast.error(err.message || 'Submit failed');
    } finally {
        btn.disabled = false;
        btn.innerHTML = originalText;
    }
});
