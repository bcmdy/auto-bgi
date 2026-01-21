<template>
    <div class="log-detail-page">
        <!-- 页面头部 - 已隐藏
        <header class="page-header enhanced-header">
            <div class="header-bg"></div>
            <div class="header-content">
                <div class="header-logo">
                    <span class="logo-icon">📄</span>
                </div>
                <div class="header-title-group">
                    <h1>日志详情</h1>
                    <p class="current-file" v-if="currentFileName">{{ currentFileName }}</p>
                </div>
                <button class="btn header-btn" @click="goBack">返回</button>
            </div>
            <div class="header-divider"></div>
        </header>
        -->

        <div class="container">
            <!-- 加载状态 -->
            <section v-if="loading" class="panel">
                <p class="loading-text">正在加载日志内容...</p>
            </section>

            <!-- 日志内容展示区域 -->
            <section v-else-if="logContent" class="panel log-content-panel">
                <iframe 
                    ref="logIframe"
                    class="log-iframe"
                    :srcdoc="iframeContent"
                    @load="onIframeLoad"
                ></iframe>
            </section>

            <!-- 无数据状态 -->
            <section v-else class="panel">
                <p class="no-data-text">无日志数据</p>
            </section>
        </div>
    </div>
</template>

<script>
import api from '@/utils/api'
import indexHtmlContent from '@/assets/index.html?raw'

export default {
    name: 'LogDetail',
    data() {
        return {
            currentFileName: '',
            logContent: '',
            loading: false,
            iframeContent: ''
        }
    },
    async mounted() {
        // 从路由参数获取文件名
        this.currentFileName = this.$route.query.file || ''
        if (this.currentFileName) {
            await this.loadLogContent()
        }
    },
    methods: {
        // 加载日志内容
        async loadLogContent() {
            if (!this.currentFileName) return

            this.loading = true
            try {
                // 调用接口获取日志内容
                const response = await api.get(`/api/logInfo?fileName=${encodeURIComponent(this.currentFileName)}`)
                this.logContent = response || ''
                
                // 加载 index.html 模板并注入日志内容
                await this.loadIndexHtml()
            } catch (error) {
                console.error('加载日志内容失败:', error)
                this.$message?.error('加载日志内容失败')
                this.logContent = ''
            } finally {
                this.loading = false
            }
        },

        // 加载 index.html 模板
        async loadIndexHtml() {
            try {
                // 使用导入的 HTML 内容
                let htmlContent = indexHtmlContent
                
                // 转义日志内容以便安全地注入到 JavaScript 中
                const escapedLogContent = this.logContent
                    .replace(/\\/g, '\\\\')
                    .replace(/`/g, '\\`')
                    .replace(/\$/g, '\\$')
                
                // 将日志内容注入到 HTML 中,直接调用 parseLog 函数处理
                this.iframeContent = htmlContent.replace(
                    '</body>',
                    `<script>
                        // 等待页面完全加载后直接处理日志内容
                        window.addEventListener('load', function() {
                            try {
                                const logContent = \`${escapedLogContent}\`;
                                const fileName = '${this.currentFileName}';
                                
                                // 重置解析上下文
                                if (typeof parsingContext !== 'undefined') {
                                    parsingContext.activeGroups.clear();
                                    parsingContext.activeTasks.clear();
                                    parsingContext.allGroups = [];
                                }
                                
                                // 重置时间追踪
                                if (typeof resetTimeTracker === 'function') {
                                    resetTimeTracker();
                                }
                                
                                const dropZone = document.getElementById('dropZone');
                                if (dropZone && typeof parseLog === 'function' && typeof finalizeParsing === 'function' && typeof generateHTML === 'function') {
                                    // 显示加载状态
                                    dropZone.innerHTML = '<div class="loading">解析中...</div>';
                                    dropZone.className = 'has-content';
                                    
                                    // 从文件名解析日期
                                    let fileDate = null;
                                    if (typeof parseDateFromFileName === 'function') {
                                        fileDate = parseDateFromFileName(fileName);
                                    }
                                    
                                    // 直接解析日志内容
                                    parseLog(logContent, fileDate);
                                    
                                    // 完成解析并生成HTML
                                    const result = finalizeParsing();
                                    dropZone.innerHTML = generateHTML(result);
                                    
                                    // 设置弹窗和其他功能
                                    if (typeof setupCoordPopups === 'function') {
                                        setupCoordPopups();
                                    }
                                    
                                    // 设置折叠功能
                                    document.querySelectorAll('.group-header').forEach((el, i) => {
                                        el.onclick = (e) => {
                                            const arrow = el.querySelector('.arrow');
                                            const content = document.getElementById(\`group-\${i}\`);
                                            const container = el.closest('.group-container');
                                            const headerRect = el.getBoundingClientRect();
                                            const isStickyAtTop = headerRect.top <= 15 && headerRect.top >= -10;
                                            
                                            if (isStickyAtTop) {
                                                const targetPosition = container.offsetTop - 40;
                                                const currentPosition = dropZone.scrollTop;
                                                const scrollDistance = Math.abs(targetPosition - currentPosition);
                                                const scrollDuration = Math.min(2000, Math.max(200, scrollDistance)) + 50;
                                                
                                                dropZone.scrollTo({
                                                    top: targetPosition,
                                                    behavior: 'smooth'
                                                });
                                                
                                                setTimeout(() => {
                                                    if (typeof toggleCollapseState === 'function') {
                                                        toggleCollapseState(el, content, arrow);
                                                    }
                                                }, scrollDuration);
                                            } else {
                                                if (typeof toggleCollapseState === 'function') {
                                                    toggleCollapseState(el, content, arrow);
                                                }
                                            }
                                        };
                                    });
                                    
                                    // 应用初始显示状态
                                    if (typeof updateTimeColumnsVisibility === 'function') {
                                        updateTimeColumnsVisibility();
                                    }
                                    if (typeof updateStatsColumnsVisibility === 'function') {
                                        updateStatsColumnsVisibility();
                                    }
                                    if (typeof updateTableFormat === 'function') {
                                        updateTableFormat();
                                    }
                                    if (typeof setupRowHighlight === 'function') {
                                        setupRowHighlight();
                                    }
                                    if (typeof showQuickNavToggle === 'function') {
                                        showQuickNavToggle();
                                    }
                                    
                                    // 标记已有日志数据
                                    if (typeof hasLogData !== 'undefined') {
                                        hasLogData = true;
                                    }
                                } else {
                                    dropZone.innerHTML = '<div class="error">日志解析函数未找到</div>';
                                }
                            } catch (error) {
                                console.error('处理日志内容失败:', error);
                                const dropZone = document.getElementById('dropZone');
                                if (dropZone) {
                                    dropZone.innerHTML = \`<div class="error">解析失败：\${error.message}</div>\`;
                                }
                            }
                        });
                    <\/script>
                    </body>`
                )
            } catch (error) {
                console.error('加载 index.html 失败:', error)
                this.$message?.error('加载页面模板失败')
            }
        },

        // iframe 加载完成回调
        onIframeLoad() {
            try {
                const iframe = this.$refs.logIframe
                if (iframe && iframe.contentWindow) {
                    // 向 iframe 传递日志数据
                    iframe.contentWindow.postMessage({
                        type: 'LOG_DATA',
                        fileName: this.currentFileName,
                        content: this.logContent
                    }, '*')
                }
            } catch (error) {
                console.error('向 iframe 发送数据失败:', error)
            }
        },

        // 返回上一页
        goBack() {
            this.$router.back()
        }
    }
}
</script>

<style scoped>
:root {
    --primary-color: #ff6eb4;
    --background-light: #fff6fb;
    --text-color: #ff6eb4;
    --border-color: #ffc0da;
    --hover-color: rgba(255, 192, 218, 0.3);
}

.log-detail-page {
    min-height: 100vh;
    background-color: var(--background-light);
    color: var(--text-color);
    background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><circle cx="20" cy="20" r="5" fill="%23ffd6eb" opacity="0.6"/><circle cx="70" cy="70" r="7" fill="%23ffc0da" opacity="0.5"/></svg>');
    padding-bottom: 0;
    overflow: hidden; /* 防止滚动条 */
}

.page-header {
    background-color: rgba(255, 255, 255, 0.8);
    padding: 30px 0 10px;
    text-align: center;
    height: 100px;
    box-shadow: 0 0 20px var(--primary-color);
    border-radius: 0 0 30px 30px;
    position: sticky;
    top: 0;
    z-index: 10;
}

.page-header.enhanced-header {
    position: relative;
    background: linear-gradient(90deg, #fff6fb 60%, #ff9ecf 100%);
    box-shadow: 0 8px 32px rgba(255, 110, 180, 0.15), 0 2px 8px rgba(255, 110, 180, 0.08);
    border-radius: 0 0 24px 24px;
    padding: 0;
    margin-bottom: 8px;
    overflow: hidden;
    z-index: 10;
}

.header-bg {
    position: absolute;
    
    top: 0; left: 0; right: 0; bottom: 0;
    background: radial-gradient(circle at 20% 40%, #e9a0d1 0%, #ecccde 60%, transparent 100%);
    opacity: 0.7;
    z-index: 0;
}

.header-content {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 20px;
    padding: 20px 20px 12px 20px;
    z-index: 1;
}

.header-logo {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 10px;
}

.logo-icon {
    font-size: 1.8rem;
    background: linear-gradient(45deg, #ff6eb4, #ff9ecf);
    border-radius: 12px;
    box-shadow: 0 2px 12px rgba(255, 110, 180, 0.18);
    padding: 6px 10px;
    color: #fff;
    border: 2px solid #ffc0da;
}

.header-title-group {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    flex: 1;
}

.page-header.enhanced-header h1 {
    color: #ff6eb4;
    font-size: 1.4rem;
    font-weight: bold;
    margin: 0;
    text-shadow: 0 2px 12px #ffc0da;
    letter-spacing: 1px;
}

.current-file {
    color: #666;
    font-size: 0.75rem;
    margin: 0;
    font-weight: normal;
}

.header-btn {
    margin-top: 0;
    font-size: 0.85rem;
    padding: 6px 16px;
    border-radius: 20px;
    box-shadow: 0 2px 8px #ffc0da;
    background: linear-gradient(45deg, #fff, #ffe3f3);
    color: #ff6eb4;
    border: 2px solid #ff6eb4;
    font-weight: bold;
    transition: all 0.3s;
}

.header-btn:hover {
    background: linear-gradient(45deg, #ff6eb4, #ff9ecf);
    color: #fff;
    box-shadow: 0 4px 16px #ff9ecf;
    transform: scale(1.07);
}

.header-divider {
    width: 80%;
    height: 4px;
    margin: 0 auto 0 auto;
    background: linear-gradient(90deg, #ff6eb4 0%, #ff9ecf 100%);
    border-radius: 2px;
    box-shadow: 0 2px 8px #ffc0da;
    opacity: 0.25;
    margin-bottom: 2px;
}

.container {
    max-width: 100%;
    margin: 0;
    padding: 0;
    height: 100vh;
}

.panel {
    background: rgba(255, 255, 255, 0.8);
    box-shadow: 0 0 15px #ffcce6;
    border-radius: 20px;
    padding: 20px 25px;
    margin-bottom: 30px;
}

.log-content-panel {
    padding: 0;
    overflow: hidden;
    height: 100vh;
    border-radius: 0;
    margin-bottom: 0;
}

.log-iframe {
    width: 100%;
    height: 100%;
    border: none;
    border-radius: 0;
}

.loading-text, .no-data-text {
    text-align: center;
    color: var(--primary-color);
    font-size: 1.2rem;
    padding: 40px 20px;
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(255, 245, 252, 0.8));
    border-radius: 15px;
    border: 2px dashed var(--border-color);
}

/* 移动端适配 */
@media (max-width: 600px) {
    .header-content {
        flex-direction: column;
        gap: 15px;
        padding: 20px;
    }

    .header-title-group {
        align-items: center;
        text-align: center;
    }

    .page-header.enhanced-header h1 {
        font-size: 1.2rem;
    }

    .current-file {
        font-size: 0.7rem;
        text-align: center;
        word-break: break-all;
    }

    .logo-icon {
        font-size: 1.4rem;
        padding: 5px 8px;
    }

    .header-btn {
        font-size: 0.75rem;
        padding: 5px 12px;
    }

    .log-content-panel {
        height: calc(100vh - 180px);
        min-height: 400px;
    }
}
</style>
