// API Hunter Web应用JavaScript

class APIHunter {
    constructor() {
        this.baseURL = '/api/v1';
        this.currentSession = null;
        this.init();
    }

    init() {
        this.loadSessions();
        this.setupEventListeners();
        this.setupSearch();
    }

    // 设置事件监听器
    setupEventListeners() {
        // 会话选择
        document.addEventListener('change', (e) => {
            if (e.target.id === 'sessionSelect') {
                this.currentSession = e.target.value;
                this.loadSessionData();
            }
        });

        // 导出按钮
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('export-btn')) {
                this.exportData(e.target.dataset.format);
            }
        });

        // 删除会话
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('delete-session-btn')) {
                this.deleteSession(e.target.dataset.sessionId);
            }
        });

        // 分析JS文件
        document.addEventListener('click', (e) => {
            if (e.target.id === 'analyzeJSBtn') {
                this.analyzeJSFiles();
            }
        });

        // 刷新按钮
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('refresh-btn')) {
                this.refreshCurrentView();
            }
        });
    }

    // 设置搜索功能
    setupSearch() {
        const searchInput = document.getElementById('searchInput');
        if (searchInput) {
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                searchTimeout = setTimeout(() => {
                    this.searchAPIs(e.target.value);
                }, 300);
            });
        }
    }

    // 加载会话列表
    async loadSessions() {
        try {
            const response = await fetch(`${this.baseURL}/sessions`);
            const data = await response.json();
            this.renderSessions(data.sessions);
        } catch (error) {
            console.error('加载会话失败:', error);
            this.showError('加载会话失败');
        }
    }

    // 渲染会话列表
    renderSessions(sessions) {
        const sessionSelect = document.getElementById('sessionSelect');
        const sessionsList = document.getElementById('sessionsList');

        if (sessionSelect) {
            sessionSelect.innerHTML = '<option value="">选择会话</option>';
            sessions.forEach(session => {
                const option = document.createElement('option');
                option.value = session.session_id;
                option.textContent = `${session.session_id} - ${session.target_url}`;
                sessionSelect.appendChild(option);
            });
        }

        if (sessionsList) {
            sessionsList.innerHTML = '';
            sessions.forEach(session => {
                const sessionCard = this.createSessionCard(session);
                sessionsList.appendChild(sessionCard);
            });
        }
    }

    // 创建会话卡片
    createSessionCard(session) {
        const card = document.createElement('div');
        card.className = 'card';
        
        const statusClass = session.status === 'completed' ? 'status-2xx' : 
                           session.status === 'running' ? 'status-3xx' : 'status-4xx';

        card.innerHTML = `
            <div class="card-title">
                ${session.session_id}
                <span class="status-code ${statusClass}">${session.status}</span>
            </div>
            <p><strong>目标URL:</strong> ${session.target_url}</p>
            <p><strong>开始时间:</strong> ${new Date(session.start_time).toLocaleString()}</p>
            <p><strong>页面数:</strong> ${session.pages_found} | <strong>API数:</strong> ${session.apis_found}</p>
            <div style="margin-top: 1rem;">
                <button class="btn btn-primary btn-sm" onclick="app.selectSession('${session.session_id}')">查看详情</button>
                <button class="btn btn-secondary btn-sm export-btn" data-format="json">导出JSON</button>
                <button class="btn btn-danger btn-sm delete-session-btn" data-session-id="${session.session_id}">删除</button>
            </div>
        `;

        return card;
    }

    // 选择会话
    selectSession(sessionId) {
        this.currentSession = sessionId;
        this.loadSessionData();
        
        // 更新选择框
        const sessionSelect = document.getElementById('sessionSelect');
        if (sessionSelect) {
            sessionSelect.value = sessionId;
        }
    }

    // 加载会话数据
    async loadSessionData() {
        if (!this.currentSession) return;

        try {
            // 加载统计信息
            await this.loadSessionStats();
            
            // 加载API列表
            await this.loadAPIs();
            
            // 加载页面列表
            await this.loadPages();
            
            // 加载JS文件
            await this.loadJSFiles();
            
        } catch (error) {
            console.error('加载会话数据失败:', error);
            this.showError('加载会话数据失败');
        }
    }

    // 加载会话统计
    async loadSessionStats() {
        const response = await fetch(`${this.baseURL}/sessions/${this.currentSession}/stats`);
        const stats = await response.json();
        this.renderStats(stats);
    }

    // 渲染统计信息
    renderStats(stats) {
        const statsContainer = document.getElementById('statsContainer');
        if (!statsContainer) return;

        statsContainer.innerHTML = `
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-number">${stats.total_pages}</div>
                    <div class="stat-label">总页面数</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${stats.total_apis}</div>
                    <div class="stat-label">总API数</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${stats.rest_apis}</div>
                    <div class="stat-label">REST APIs</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${stats.graphql_apis}</div>
                    <div class="stat-label">GraphQL APIs</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${stats.websocket_apis}</div>
                    <div class="stat-label">WebSocket APIs</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${stats.js_files}</div>
                    <div class="stat-label">JS文件数</div>
                </div>
            </div>
        `;
    }

    // 加载API列表
    async loadAPIs(page = 1, limit = 50) {
        const response = await fetch(`${this.baseURL}/apis?session_id=${this.currentSession}&limit=${limit}&offset=${(page-1)*limit}`);
        const data = await response.json();
        this.renderAPIs(data.apis);
    }

    // 渲染API列表
    renderAPIs(apis) {
        const apisContainer = document.getElementById('apisContainer');
        if (!apisContainer) return;

        if (apis.length === 0) {
            apisContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">🔍</div>
                    <div class="empty-state-text">暂无API数据</div>
                    <div class="empty-state-subtext">请先运行扫描或选择其他会话</div>
                </div>
            `;
            return;
        }

        const table = document.createElement('div');
        table.className = 'table-container';
        table.innerHTML = `
            <table>
                <thead>
                    <tr>
                        <th>方法</th>
                        <th>路径</th>
                        <th>域名</th>
                        <th>类型</th>
                        <th>状态码</th>
                        <th>来源</th>
                        <th>发现时间</th>
                    </tr>
                </thead>
                <tbody>
                    ${apis.map(api => `
                        <tr>
                            <td><span class="method-tag method-${api.method}">${api.method}</span></td>
                            <td><code>${api.path}</code></td>
                            <td>${api.domain}</td>
                            <td><span class="type-tag type-${api.type}">${api.type}</span></td>
                            <td><span class="status-code status-${Math.floor(api.status/100)}xx">${api.status || '-'}</span></td>
                            <td>${api.source}</td>
                            <td>${new Date(api.created_at).toLocaleString()}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;

        apisContainer.innerHTML = '';
        apisContainer.appendChild(table);
    }

    // 搜索API
    async searchAPIs(keyword) {
        if (!keyword.trim()) {
            this.loadAPIs();
            return;
        }

        try {
            const response = await fetch(`${this.baseURL}/apis/search?q=${encodeURIComponent(keyword)}`);
            const data = await response.json();
            this.renderAPIs(data.apis);
        } catch (error) {
            console.error('搜索失败:', error);
            this.showError('搜索失败');
        }
    }

    // 加载页面列表
    async loadPages() {
        if (!this.currentSession) return;

        try {
            const response = await fetch(`${this.baseURL}/pages?session_id=${this.currentSession}`);
            const data = await response.json();
            this.renderPages(data.pages);
        } catch (error) {
            console.error('加载页面失败:', error);
        }
    }

    // 渲染页面列表
    renderPages(pages) {
        const pagesContainer = document.getElementById('pagesContainer');
        if (!pagesContainer) return;

        if (pages.length === 0) {
            pagesContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📄</div>
                    <div class="empty-state-text">暂无页面数据</div>
                </div>
            `;
            return;
        }

        const table = document.createElement('div');
        table.className = 'table-container';
        table.innerHTML = `
            <table>
                <thead>
                    <tr>
                        <th>URL</th>
                        <th>标题</th>
                        <th>深度</th>
                        <th>大小</th>
                        <th>链接数</th>
                        <th>API数</th>
                        <th>爬取时间</th>
                    </tr>
                </thead>
                <tbody>
                    ${pages.map(page => `
                        <tr>
                            <td><a href="${page.url}" target="_blank">${this.truncateURL(page.url)}</a></td>
                            <td>${page.title || '-'}</td>
                            <td>${page.depth}</td>
                            <td>${this.formatFileSize(page.size)}</td>
                            <td>${page.links}</td>
                            <td>${page.apis}</td>
                            <td>${new Date(page.created_at).toLocaleString()}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;

        pagesContainer.innerHTML = '';
        pagesContainer.appendChild(table);
    }

    // 加载JS文件
    async loadJSFiles() {
        if (!this.currentSession) return;

        try {
            const response = await fetch(`${this.baseURL}/jsfiles?session_id=${this.currentSession}`);
            const data = await response.json();
            this.renderJSFiles(data.js_files);
        } catch (error) {
            console.error('加载JS文件失败:', error);
        }
    }

    // 渲染JS文件列表
    renderJSFiles(jsFiles) {
        const jsContainer = document.getElementById('jsFilesContainer');
        if (!jsContainer) return;

        if (jsFiles.length === 0) {
            jsContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📜</div>
                    <div class="empty-state-text">暂无JavaScript文件</div>
                </div>
            `;
            return;
        }

        const table = document.createElement('div');
        table.className = 'table-container';
        table.innerHTML = `
            <table>
                <thead>
                    <tr>
                        <th>URL</th>
                        <th>大小</th>
                        <th>API数</th>
                        <th>已分析</th>
                        <th>发现时间</th>
                    </tr>
                </thead>
                <tbody>
                    ${jsFiles.map(jsFile => `
                        <tr>
                            <td><a href="${jsFile.url}" target="_blank">${this.truncateURL(jsFile.url)}</a></td>
                            <td>${this.formatFileSize(jsFile.size)}</td>
                            <td>${jsFile.apis}</td>
                            <td>${jsFile.analyzed ? '✅' : '❌'}</td>
                            <td>${new Date(jsFile.created_at).toLocaleString()}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;

        jsContainer.innerHTML = '';
        jsContainer.appendChild(table);
    }

    // 分析JS文件
    async analyzeJSFiles() {
        if (!this.currentSession) {
            this.showError('请先选择会话');
            return;
        }

        try {
            const response = await fetch(`${this.baseURL}/jsfiles/analyze`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    session_id: this.currentSession
                })
            });

            const data = await response.json();
            this.showSuccess('JavaScript文件分析已启动');
            
            // 刷新数据
            setTimeout(() => {
                this.loadSessionData();
            }, 2000);
            
        } catch (error) {
            console.error('分析JS文件失败:', error);
            this.showError('分析JS文件失败');
        }
    }

    // 导出数据
    async exportData(format) {
        if (!this.currentSession) {
            this.showError('请先选择会话');
            return;
        }

        try {
            const response = await fetch(`${this.baseURL}/export`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    session_id: this.currentSession,
                    format: format,
                    include_details: true
                })
            });

            const data = await response.json();
            this.showSuccess(`导出成功: ${data.result.file_path}`);
            
        } catch (error) {
            console.error('导出失败:', error);
            this.showError('导出失败');
        }
    }

    // 删除会话
    async deleteSession(sessionId) {
        if (!confirm('确定要删除这个会话吗？此操作不可恢复。')) {
            return;
        }

        try {
            const response = await fetch(`${this.baseURL}/sessions/${sessionId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.showSuccess('会话删除成功');
                this.loadSessions();
                
                if (this.currentSession === sessionId) {
                    this.currentSession = null;
                    this.clearCurrentView();
                }
            } else {
                throw new Error('删除失败');
            }
            
        } catch (error) {
            console.error('删除会话失败:', error);
            this.showError('删除会话失败');
        }
    }

    // 刷新当前视图
    refreshCurrentView() {
        if (this.currentSession) {
            this.loadSessionData();
        } else {
            this.loadSessions();
        }
    }

    // 清空当前视图
    clearCurrentView() {
        const containers = ['statsContainer', 'apisContainer', 'pagesContainer', 'jsFilesContainer'];
        containers.forEach(id => {
            const container = document.getElementById(id);
            if (container) {
                container.innerHTML = '';
            }
        });
    }

    // 工具函数
    truncateURL(url, maxLength = 50) {
        if (url.length <= maxLength) return url;
        return url.substring(0, maxLength) + '...';
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    // 显示成功消息
    showSuccess(message) {
        this.showMessage(message, 'success');
    }

    // 显示错误消息
    showError(message) {
        this.showMessage(message, 'error');
    }

    // 显示消息
    showMessage(message, type = 'info') {
        // 创建消息元素
        const messageEl = document.createElement('div');
        messageEl.className = `message message-${type}`;
        messageEl.textContent = message;
        
        // 添加样式
        messageEl.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 1rem 1.5rem;
            border-radius: 5px;
            color: white;
            font-weight: 500;
            z-index: 1000;
            animation: slideIn 0.3s ease;
        `;

        if (type === 'success') {
            messageEl.style.background = '#28a745';
        } else if (type === 'error') {
            messageEl.style.background = '#dc3545';
        } else {
            messageEl.style.background = '#17a2b8';
        }

        document.body.appendChild(messageEl);

        // 3秒后自动移除
        setTimeout(() => {
            messageEl.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => {
                document.body.removeChild(messageEl);
            }, 300);
        }, 3000);
    }
}

// 添加动画样式
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }
`;
document.head.appendChild(style);

// 初始化应用
const app = new APIHunter();