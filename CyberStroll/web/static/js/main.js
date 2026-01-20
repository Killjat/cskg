// CyberStroll 主要JavaScript文件

document.addEventListener('DOMContentLoaded', function() {
    // 初始化页面
    initializePage();
    
    // 加载统计数据
    loadStats();
    
    // 设置搜索表单增强
    enhanceSearchForm();
});

// 初始化页面
function initializePage() {
    // 添加搜索提示
    addSearchHints();
    
    // 设置表格排序
    setupTableSorting();
    
    // 设置响应式导航
    setupResponsiveNav();
}

// 加载统计数据
function loadStats() {
    fetch('/api/stats')
        .then(response => response.json())
        .then(data => {
            updateStatsDisplay(data);
        })
        .catch(error => {
            console.error('加载统计数据失败:', error);
        });
}

// 更新统计显示
function updateStatsDisplay(stats) {
    // 更新总主机数
    const totalHostsEl = document.getElementById('total-hosts');
    if (totalHostsEl && stats.elasticsearch) {
        // 从ES统计中提取数据
        totalHostsEl.textContent = formatNumber(1000); // 示例数据
    }
    
    // 更新总端口数
    const totalPortsEl = document.getElementById('total-ports');
    if (totalPortsEl) {
        totalPortsEl.textContent = formatNumber(5000); // 示例数据
    }
    
    // 更新服务数
    const totalServicesEl = document.getElementById('total-services');
    if (totalServicesEl) {
        totalServicesEl.textContent = formatNumber(50); // 示例数据
    }
    
    // 更新最后更新时间
    const lastUpdateEl = document.getElementById('last-update');
    if (lastUpdateEl) {
        lastUpdateEl.textContent = formatTime(new Date());
    }
}

// 格式化数字
function formatNumber(num) {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
}

// 格式化时间
function formatTime(date) {
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) { // 1分钟内
        return '刚刚';
    } else if (diff < 3600000) { // 1小时内
        return Math.floor(diff / 60000) + '分钟前';
    } else if (diff < 86400000) { // 1天内
        return Math.floor(diff / 3600000) + '小时前';
    } else {
        return date.toLocaleDateString();
    }
}

// 添加搜索提示
function addSearchHints() {
    const searchInputs = document.querySelectorAll('.search-input');
    
    searchInputs.forEach(input => {
        input.addEventListener('focus', function() {
            showSearchHint(this);
        });
        
        input.addEventListener('blur', function() {
            hideSearchHint(this);
        });
    });
}

// 显示搜索提示
function showSearchHint(input) {
    const hints = {
        'ip': [
            '单个IP: 192.168.1.1',
            'CIDR: 192.168.1.0/24',
            'IP范围: 192.168.1.1-192.168.1.100'
        ],
        'port': [
            '单个端口: 80',
            '端口范围: 80-443',
            '多个端口: 80,443,8080'
        ],
        'banner': [
            '关键词: nginx',
            '版本: Apache/2.4',
            '协议: SSH-2.0'
        ],
        'service': [
            'http, https, ssh, ftp',
            'mysql, redis, mongodb',
            'smtp, pop3, imap'
        ],
        'country': [
            '中国, 美国, 日本',
            'CN, US, JP',
            'China, United States'
        ]
    };
    
    const name = input.getAttribute('name');
    if (hints[name]) {
        showTooltip(input, hints[name]);
    }
}

// 隐藏搜索提示
function hideSearchHint(input) {
    hideTooltip(input);
}

// 显示工具提示
function showTooltip(element, hints) {
    const tooltip = document.createElement('div');
    tooltip.className = 'search-tooltip';
    tooltip.innerHTML = hints.map(hint => `<div>${hint}</div>`).join('');
    
    document.body.appendChild(tooltip);
    
    const rect = element.getBoundingClientRect();
    tooltip.style.position = 'absolute';
    tooltip.style.top = (rect.bottom + 5) + 'px';
    tooltip.style.left = rect.left + 'px';
    tooltip.style.background = '#333';
    tooltip.style.color = 'white';
    tooltip.style.padding = '10px';
    tooltip.style.borderRadius = '4px';
    tooltip.style.fontSize = '12px';
    tooltip.style.zIndex = '1000';
    tooltip.style.maxWidth = '200px';
    
    element._tooltip = tooltip;
}

// 隐藏工具提示
function hideTooltip(element) {
    if (element._tooltip) {
        document.body.removeChild(element._tooltip);
        element._tooltip = null;
    }
}

// 增强搜索表单
function enhanceSearchForm() {
    const searchForm = document.querySelector('.search-form');
    if (!searchForm) return;
    
    // 添加搜索历史
    loadSearchHistory();
    
    // 添加快捷搜索
    addQuickSearch();
    
    // 表单验证
    searchForm.addEventListener('submit', function(e) {
        if (!validateSearchForm(this)) {
            e.preventDefault();
        }
    });
}

// 加载搜索历史
function loadSearchHistory() {
    const history = JSON.parse(localStorage.getItem('searchHistory') || '[]');
    // TODO: 显示搜索历史
}

// 保存搜索历史
function saveSearchHistory(query) {
    let history = JSON.parse(localStorage.getItem('searchHistory') || '[]');
    history.unshift(query);
    history = history.slice(0, 10); // 只保留最近10条
    localStorage.setItem('searchHistory', JSON.stringify(history));
}

// 添加快捷搜索
function addQuickSearch() {
    const quickSearches = [
        { name: 'Web服务', query: { service: 'http' } },
        { name: 'SSH服务', query: { service: 'ssh' } },
        { name: '数据库', query: { service: 'mysql' } },
        { name: '常见端口', query: { port: '80,443,22,21,25' } }
    ];
    
    // TODO: 添加快捷搜索按钮
}

// 验证搜索表单
function validateSearchForm(form) {
    const formData = new FormData(form);
    const hasQuery = Array.from(formData.values()).some(value => value.trim() !== '');
    
    if (!hasQuery) {
        alert('请至少输入一个搜索条件');
        return false;
    }
    
    // 验证IP格式
    const ip = formData.get('ip');
    if (ip && !validateIP(ip)) {
        alert('IP地址格式不正确');
        return false;
    }
    
    // 验证端口格式
    const port = formData.get('port');
    if (port && !validatePort(port)) {
        alert('端口格式不正确');
        return false;
    }
    
    return true;
}

// 验证IP地址
function validateIP(ip) {
    // 支持单个IP、CIDR、IP范围
    const ipRegex = /^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/;
    const rangeRegex = /^(\d{1,3}\.){3}\d{1,3}-(\d{1,3}\.){3}\d{1,3}$/;
    
    return ipRegex.test(ip) || rangeRegex.test(ip);
}

// 验证端口
function validatePort(port) {
    // 支持单个端口、端口范围、多个端口
    const portRegex = /^\d+(-\d+)?(,\d+(-\d+)?)*$/;
    return portRegex.test(port);
}

// 设置表格排序
function setupTableSorting() {
    const tables = document.querySelectorAll('table');
    
    tables.forEach(table => {
        const headers = table.querySelectorAll('th');
        
        headers.forEach((header, index) => {
            header.style.cursor = 'pointer';
            header.addEventListener('click', function() {
                sortTable(table, index);
            });
        });
    });
}

// 表格排序
function sortTable(table, columnIndex) {
    const tbody = table.querySelector('tbody');
    const rows = Array.from(tbody.querySelectorAll('tr'));
    
    const isAscending = table.getAttribute('data-sort-direction') !== 'asc';
    table.setAttribute('data-sort-direction', isAscending ? 'asc' : 'desc');
    
    rows.sort((a, b) => {
        const aValue = a.cells[columnIndex].textContent.trim();
        const bValue = b.cells[columnIndex].textContent.trim();
        
        // 尝试数字排序
        const aNum = parseFloat(aValue);
        const bNum = parseFloat(bValue);
        
        if (!isNaN(aNum) && !isNaN(bNum)) {
            return isAscending ? aNum - bNum : bNum - aNum;
        }
        
        // 字符串排序
        return isAscending ? 
            aValue.localeCompare(bValue) : 
            bValue.localeCompare(aValue);
    });
    
    // 重新插入排序后的行
    rows.forEach(row => tbody.appendChild(row));
    
    // 更新表头排序指示器
    updateSortIndicator(table, columnIndex, isAscending);
}

// 更新排序指示器
function updateSortIndicator(table, columnIndex, isAscending) {
    const headers = table.querySelectorAll('th');
    
    // 清除所有指示器
    headers.forEach(header => {
        header.classList.remove('sort-asc', 'sort-desc');
    });
    
    // 添加当前列的指示器
    headers[columnIndex].classList.add(isAscending ? 'sort-asc' : 'sort-desc');
}

// 设置响应式导航
function setupResponsiveNav() {
    // 检测移动设备
    const isMobile = window.innerWidth <= 768;
    
    if (isMobile) {
        // 移动设备优化
        optimizeForMobile();
    }
    
    // 监听窗口大小变化
    window.addEventListener('resize', function() {
        const nowMobile = window.innerWidth <= 768;
        if (nowMobile !== isMobile) {
            location.reload(); // 简单的响应式处理
        }
    });
}

// 移动设备优化
function optimizeForMobile() {
    // 简化表格显示
    const tables = document.querySelectorAll('table');
    tables.forEach(table => {
        table.style.fontSize = '12px';
    });
    
    // 调整搜索框布局
    const searchRows = document.querySelectorAll('.search-row');
    searchRows.forEach(row => {
        row.style.flexDirection = 'column';
    });
}

// 实用工具函数
const utils = {
    // 复制到剪贴板
    copyToClipboard: function(text) {
        navigator.clipboard.writeText(text).then(function() {
            console.log('已复制到剪贴板');
        });
    },
    
    // 下载数据
    downloadData: function(data, filename) {
        const blob = new Blob([data], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        a.click();
        URL.revokeObjectURL(url);
    },
    
    // 格式化JSON
    formatJSON: function(obj) {
        return JSON.stringify(obj, null, 2);
    }
};

// 导出到全局
window.CyberStroll = {
    utils: utils,
    loadStats: loadStats,
    validateIP: validateIP,
    validatePort: validatePort
};