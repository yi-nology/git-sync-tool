document.addEventListener("DOMContentLoaded", function() {
    const path = window.location.pathname;
    const page = path.split("/").pop() || "index.html";

    const navItems = [
        { name: "首页", link: "index.html", active: page === "index.html" || page === "" },
        { name: "仓库管理", link: "repos.html", active: page === "repos.html" },
        { name: "设置", link: "settings.html", active: page === "settings.html" },
        { name: "度量", link: "stats.html", active: page === "stats.html" }
    ];

    let navHtml = `
    <nav class="navbar navbar-expand-lg navbar-dark bg-dark mb-4">
        <div class="container-fluid">
            <a class="navbar-brand" href="index.html">分支管理工具</a>
            <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
                <span class="navbar-toggler-icon"></span>
            </button>
            <div class="collapse navbar-collapse" id="navbarNav">
                <ul class="navbar-nav">
    `;

    navItems.forEach(item => {
        navHtml += `
        <li class="nav-item">
            <a class="nav-link ${item.active ? 'active' : ''}" href="${item.link}">${item.name}</a>
        </li>`;
    });

    navHtml += `
                </ul>
            </div>
        </div>
    </nav>
    `;

    document.body.insertAdjacentHTML("afterbegin", navHtml);
});
