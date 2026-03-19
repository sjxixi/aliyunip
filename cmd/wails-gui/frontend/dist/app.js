let app;
let currentStep = 1;
let selectedResources = [];
let resourcesData = null;
let ipAddress = '';
let portNumber = 0;
let description = '';
let sectionStates = {
    alb: false,
    ecs: false,
    rds: false,
    polardb: false,
    redis: false,
    cloudfw: false
};

window.onload = function() {
    app = window.go?.main?.App;
    
    initButtons();
    loadSavedConfig();
};

function initButtons() {
    document.getElementById('validate-btn').addEventListener('click', validateCredentials);
    document.getElementById('step2-back').addEventListener('click', () => goToStep(1));
    document.getElementById('step2-next').addEventListener('click', goToStep3);
    document.getElementById('step3-back').addEventListener('click', () => goToStep(2));
    document.getElementById('step3-next').addEventListener('click', goToStep4);
    document.getElementById('step4-back').addEventListener('click', () => goToStep(3));
    document.getElementById('execute-btn').addEventListener('click', executeConfig);
    document.getElementById('restart-btn').addEventListener('click', restart);
}

function loadSavedConfig() {
    if (!app) return;
    
    try {
        const savedConfig = app.LoadSavedConfig();
        if (savedConfig) {
            document.getElementById('access-key-id').value = savedConfig.accessKeyId || '';
            document.getElementById('access-key-secret').value = savedConfig.accessKeySecret || '';
            document.getElementById('region').value = savedConfig.region || 'cn-shanghai-finance-1';
        }
    } catch (e) {
        console.log('No saved config found');
    }
}

function goToStep(step) {
    document.querySelectorAll('.step').forEach(s => s.classList.remove('active'));
    document.querySelector(`.step-${step}`).classList.add('active');
    
    document.querySelectorAll('.step-content').forEach(c => c.classList.remove('active'));
    document.getElementById(`step${step}`).classList.add('active');
    
    currentStep = step;
}

async function validateCredentials() {
    if (!app) {
        showMessage('step1-message', '无法连接到后端，请确保使用 Wails 运行', 'error');
        return;
    }
    
    const accessKeyId = document.getElementById('access-key-id').value.trim();
    const accessKeySecret = document.getElementById('access-key-secret').value.trim();
    const region = document.getElementById('region').value.trim();
    
    if (!accessKeyId || !accessKeySecret || !region) {
        showMessage('step1-message', '请填写所有必填字段', 'error');
        return;
    }
    
    document.getElementById('validate-btn').disabled = true;
    showMessage('step1-message', '正在验证...', '');
    
    try {
        const result = await app.ValidateCredentials(accessKeyId, accessKeySecret, region);
        
        if (result.success) {
            showMessage('step1-message', result.message, 'success');
            await loadResources();
            goToStep(2);
        } else {
            showMessage('step1-message', result.message + (result.error ? ': ' + result.error : ''), 'error');
        }
    } catch (e) {
        showMessage('step1-message', '验证失败: ' + (e.message || e), 'error');
    } finally {
        document.getElementById('validate-btn').disabled = false;
    }
}

async function loadResources() {
    if (!app) return;
    
    document.getElementById('resources-loading').style.display = 'flex';
    document.getElementById('resources-container').style.display = 'none';
    
    try {
        resourcesData = await app.LoadResources();
        renderResources();
        document.getElementById('resources-loading').style.display = 'none';
        document.getElementById('resources-container').style.display = 'block';
    } catch (e) {
        showMessage('step2-message', '加载资源失败: ' + (e.message || e), 'error');
        document.getElementById('resources-loading').style.display = 'none';
    }
}

function toggleSection(section) {
    const content = document.getElementById(`content-${section}`);
    const toggleBtn = document.getElementById(`toggle-${section}`);
    
    sectionStates[section] = !sectionStates[section];
    
    if (sectionStates[section]) {
        content.classList.remove('collapsed');
        toggleBtn.textContent = '▼';
    } else {
        content.classList.add('collapsed');
        toggleBtn.textContent = '▶';
    }
}

function renderResources() {
    if (!resourcesData) return;
    
    renderResourceList('alb-list', resourcesData.albPolicies || [], 'alb', 'aclId', 'aclName');
    renderResourceList('sg-list', resourcesData.securityGroups || [], 'ecs', 'securityGroupId', 'securityGroupName');
    renderResourceList('rds-list', resourcesData.rdsInstances || [], 'rds', 'dbInstanceId', 'dbInstanceId');
    renderResourceList('polardb-list', resourcesData.polarDBClusters || [], 'polardb', 'dbClusterId', 'dbClusterId');
    renderResourceList('redis-list', resourcesData.redisInstances || [], 'redis', 'instanceId', 'instanceId');
    renderResourceList('cloudfw-list', resourcesData.addressBooks || [], 'cloudfw', 'addressBookId', 'addressBookName');
    
    ['alb', 'ecs', 'rds', 'polardb', 'redis', 'cloudfw'].forEach(section => {
        sectionStates[section] = false;
        document.getElementById(`content-${section}`).classList.add('collapsed');
        document.getElementById(`toggle-${section}`).textContent = '▶';
    });
}

function renderResourceList(containerId, items, type, idField, nameField) {
    const container = document.getElementById(containerId);
    if (!container) return;
    
    container.innerHTML = '';
    
    if (items.length === 0) {
        container.innerHTML = '<div style="padding: 10px; color: #888; font-size: 14px;">暂无资源</div>';
        return;
    }
    
    items.forEach(item => {
        const id = item[idField];
        const name = item[nameField] || id;
        
        let description = '';
        if (type === 'ecs') {
            description = item.description || '';
        } else if (type === 'rds') {
            description = item.dbInstanceDescription || '';
        } else if (type === 'polardb') {
            description = item.dbClusterDescription || '';
        } else if (type === 'redis') {
            description = item.instanceName || '';
        }
        
        const hasGroups = type === 'rds' || type === 'polardb' || type === 'redis';
        const groups = item.securityIpGroups || [];
        const defaultGroup = groups.length > 0 ? groups[0] : 'default';
        
        const div = document.createElement('div');
        div.className = 'resource-item';
        div.innerHTML = `
            <input type="checkbox" data-type="${type}" data-id="${id}" data-name="${name}" data-default-group="${defaultGroup}">
            <div class="resource-info">
                <div class="resource-name">${escapeHtml(name)}</div>
                <div class="resource-id">${escapeHtml(id)}</div>
                ${description ? `
                    <div class="resource-description" style="margin-top: 4px; font-size: 12px; color: #666;">
                        ${escapeHtml(description)}
                    </div>
                ` : ''}
                ${hasGroups && groups.length > 0 ? `
                    <div class="group-select" style="margin-top: 8px;">
                        <label style="font-size: 12px; color: #888; margin-right: 8px;">分组:</label>
                        <select class="security-group-select" data-type="${type}" data-id="${id}" style="padding: 4px 8px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px;">
                            ${groups.map(g => `<option value="${g}">${g}</option>`).join('')}
                        </select>
                    </div>
                ` : ''}
            </div>
        `;
        
        div.addEventListener('click', (e) => {
            if (e.target.type !== 'checkbox' && e.target.tagName !== 'SELECT') {
                const checkbox = div.querySelector('input[type="checkbox"]');
                checkbox.checked = !checkbox.checked;
                toggleResource(checkbox);
            }
        });
        
        const checkbox = div.querySelector('input[type="checkbox"]');
        checkbox.addEventListener('change', () => toggleResource(checkbox));
        
        const select = div.querySelector('select');
        if (select) {
            select.addEventListener('change', (e) => {
                const id = e.target.dataset.id;
                const type = e.target.dataset.type;
                const group = e.target.value;
                
                const resource = selectedResources.find(r => r.id === id && r.type === type);
                if (resource) {
                    resource.securityIpGroup = group;
                    updateSelectedSummary();
                }
            });
        }
        
        container.appendChild(div);
    });
}

function toggleResource(checkbox) {
    const type = checkbox.dataset.type;
    const id = checkbox.dataset.id;
    const name = checkbox.dataset.name;
    const defaultGroup = checkbox.dataset.defaultGroup;
    
    const parent = checkbox.closest('.resource-item');
    const select = parent.querySelector('select');
    
    let securityIpGroup = defaultGroup;
    if (select) {
        securityIpGroup = select.value;
    }
    
    if (checkbox.checked) {
        parent.classList.add('selected');
        const resource = { type, id, name, securityIpGroup: securityIpGroup };
        selectedResources.push(resource);
    } else {
        parent.classList.remove('selected');
        selectedResources = selectedResources.filter(r => r.id !== id || r.type !== type);
    }
    
    updateSelectedSummary();
}

function updateSelectedSummary() {
    const summary = document.getElementById('selected-summary');
    const list = document.getElementById('selected-list');
    
    if (selectedResources.length === 0) {
        summary.style.display = 'none';
        return;
    }
    
    summary.style.display = 'block';
    list.innerHTML = selectedResources.map(r => `
        <div class="selected-tag">
            <span class="tag-type">${r.type.toUpperCase()}</span>
            ${escapeHtml(r.name)}
            ${r.securityIpGroup ? `<span style="font-size: 11px; color: #888; margin-left: 4px;">[${escapeHtml(r.securityIpGroup)}]</span>` : ''}
        </div>
    `).join('');
}

async function goToStep3() {
    if (selectedResources.length === 0) {
        showMessage('step2-message', '请至少选择一个资源', 'error');
        return;
    }
    
    renderSelectedPreview();
    goToStep(3);
    
    const ipHint = document.getElementById('ip-hint');
    const ipInput = document.getElementById('ip-input');
    ipHint.style.display = 'none';
    ipInput.value = '';
    
    if (app) {
        try {
            const result = await app.GetPublicIP();
            if (result.success && result.message) {
                ipInput.value = result.message;
                ipHint.style.display = 'block';
            }
        } catch (e) {
            console.log('Failed to get public IP:', e);
        }
    }
}

function renderSelectedPreview() {
    const preview = document.getElementById('selected-preview');
    preview.innerHTML = `
        <h3>已选择的资源 (${selectedResources.length})</h3>
        <div class="preview-list">
            ${selectedResources.map(r => `
                <div class="preview-item">
                    <div class="item-type">${r.type.toUpperCase()}</div>
                    <div class="item-name">
                        ${escapeHtml(r.name)}
                        ${r.securityIpGroup ? `<span style="font-size: 12px; color: #888; margin-left: 8px;">分组: ${escapeHtml(r.securityIpGroup)}</span>` : ''}
                    </div>
                </div>
            `).join('')}
        </div>
    `;
}

async function goToStep4() {
    const ipInput = document.getElementById('ip-input').value.trim();
    const portInput = document.getElementById('port-input').value.trim();
    const descInput = document.getElementById('description-input').value.trim();
    
    document.getElementById('ip-error').textContent = '';
    document.getElementById('port-error').textContent = '';
    
    if (!app) {
        showMessage('step3-message', '无法连接到后端', 'error');
        return;
    }
    
    let ipValid, portValid;
    try {
        ipValid = await app.ValidateIP(ipInput);
        portValid = await app.ValidatePort(portInput);
    } catch (e) {
        showMessage('step3-message', '验证失败: ' + (e.message || e), 'error');
        return;
    }
    
    if (!ipValid.success) {
        document.getElementById('ip-error').textContent = ipValid.message;
        return;
    }
    
    if (!portValid.success) {
        document.getElementById('port-error').textContent = portValid.message;
        return;
    }
    
    ipAddress = ipInput;
    portNumber = portInput ? parseInt(portInput) : 0;
    description = descInput;
    
    renderExecutionPreview();
    goToStep(4);
}

function renderExecutionPreview() {
    const preview = document.getElementById('execution-preview');
    preview.innerHTML = `
        <h3>即将执行的配置</h3>
        <div class="preview-list">
            <div class="preview-item">
                <div class="item-type">IP</div>
                <div class="item-name">${escapeHtml(ipAddress)}</div>
            </div>
            ${portNumber > 0 ? `
                <div class="preview-item">
                    <div class="item-type">端口</div>
                    <div class="item-name">${portNumber}</div>
                </div>
            ` : ''}
            ${description ? `
                <div class="preview-item">
                    <div class="item-type">备注</div>
                    <div class="item-name">${escapeHtml(description)}</div>
                </div>
            ` : ''}
        </div>
        <h3 style="margin-top: 16px;">目标资源 (${selectedResources.length})</h3>
        <div class="preview-list">
            ${selectedResources.map(r => `
                <div class="preview-item">
                    <div class="item-type">${r.type.toUpperCase()}</div>
                    <div class="item-name">
                        ${escapeHtml(r.name)}
                        ${r.securityIpGroup ? `<span style="font-size: 12px; color: #888; margin-left: 8px;">分组: ${escapeHtml(r.securityIpGroup)}</span>` : ''}
                    </div>
                </div>
            `).join('')}
        </div>
    `;
}

async function executeConfig() {
    if (!app) return;
    
    document.getElementById('execute-btn').disabled = true;
    document.getElementById('step4-back').disabled = true;
    document.getElementById('execution-progress').style.display = 'block';
    document.getElementById('execution-results').style.display = 'none';
    
    const progressBar = document.getElementById('progress');
    const progressText = document.getElementById('progress-text');
    
    try {
        const results = await app.ExecuteConfig(ipAddress, portNumber, description, selectedResources);
        
        let completed = 0;
        const total = results.length;
        
        const resultsList = document.getElementById('results-list');
        resultsList.innerHTML = '';
        
        for (let i = 0; i < results.length; i++) {
            const result = results[i];
            const progress = ((i + 1) / total) * 100;
            progressBar.style.width = progress + '%';
            progressText.textContent = result.message;
            
            const resultDiv = document.createElement('div');
            resultDiv.className = 'result-item ' + (result.success ? 'success' : 'error');
            resultDiv.innerHTML = `
                <div class="result-message">${escapeHtml(result.message)}</div>
                ${result.error ? `<div class="result-error">${escapeHtml(result.error)}</div>` : ''}
            `;
            resultsList.appendChild(resultDiv);
            
            await new Promise(r => setTimeout(r, 300));
        }
        
        progressText.textContent = '执行完成！';
        document.getElementById('execution-results').style.display = 'block';
        document.getElementById('restart-group').style.display = 'flex';
        
    } catch (e) {
        document.getElementById('execution-results').style.display = 'block';
        document.getElementById('results-list').innerHTML = `
            <div class="result-item error">
                <div class="result-message">执行失败</div>
                <div class="result-error">${escapeHtml(e.message || e)}</div>
            </div>
        `;
        document.getElementById('restart-group').style.display = 'flex';
    } finally {
        document.getElementById('execute-btn').style.display = 'none';
        document.getElementById('step4-back').style.display = 'none';
    }
}

function restart() {
    currentStep = 1;
    selectedResources = [];
    resourcesData = null;
    ipAddress = '';
    portNumber = 0;
    description = '';
    
    document.getElementById('step1-message').textContent = '';
    document.getElementById('step2-message').textContent = '';
    document.getElementById('step3-message').textContent = '';
    document.getElementById('ip-error').textContent = '';
    document.getElementById('port-error').textContent = '';
    document.getElementById('ip-input').value = '';
    document.getElementById('port-input').value = '';
    document.getElementById('description-input').value = '';
    document.getElementById('selected-summary').style.display = 'none';
    
    document.getElementById('execution-progress').style.display = 'none';
    document.getElementById('execution-results').style.display = 'none';
    document.getElementById('restart-group').style.display = 'none';
    document.getElementById('execute-btn').style.display = 'inline-block';
    document.getElementById('execute-btn').disabled = false;
    document.getElementById('step4-back').style.display = 'inline-block';
    document.getElementById('step4-back').disabled = false;
    
    goToStep(1);
}

function showMessage(elementId, message, type) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    element.textContent = message;
    element.className = 'message';
    if (type) {
        element.classList.add(type);
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
