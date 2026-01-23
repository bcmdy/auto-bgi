<template>
  <div class="archive-page">
    <!-- 背景：二次元中二风（纯CSS，不改逻辑） -->
    <div class="bg-layer" aria-hidden="true">
      <div class="bg-grid"></div>
      <div class="bg-orbs"></div>
      <div class="bg-sparkles"></div>
      <div class="bg-vignette"></div>
    </div>

    <!-- 漂浮装饰 -->
    <div class="decorative-elements" aria-hidden="true">
      <div class="floating-heart heart-1">💖</div>
      <div class="floating-heart heart-2">✨</div>
      <div class="floating-heart heart-3">🌸</div>
      <div class="floating-heart heart-4">🎀</div>
      <div class="floating-heart heart-5">💕</div>
      <div class="floating-heart heart-6">⚡</div>
      <div class="floating-heart heart-7">🌙</div>
    </div>

    <header class="cute-header">
      <div class="header-content">
        <button class="btn cute-btn" @click="goHome">
          <span class="btn-icon">🏠</span>
          <span class="btn-text">返回首页</span>
        </button>

        <h1 class="cute-title">
          <span class="title-icon">📚</span>
          <span class="title-text">归档记录列表</span>
          <span class="title-sparkle">✨</span>
        </h1>

        <a-button type="primary" class="deleteBtn" @click="allDelete">全部删除</a-button>
      </div>

      <div class="header-subline" aria-hidden="true">
        <span class="rune">✦</span>
        <span class="subtext">「Archive Protocol · Neon Midnight」</span>
        <span class="rune">✦</span>
      </div>
    </header>

    <div class="container">
      <section class="panel cute-panel" id="archiveListPanel">
        <div class="panel-header">
          <h2 class="cute-subtitle">
            <span class="subtitle-icon">📋</span>
            <span class="subtitle-text">归档记录</span>
          </h2>

          <div class="panel-decoration" aria-hidden="true">
            <div class="corner-decoration corner-tl">🌸</div>
            <div class="corner-decoration corner-tr">🌸</div>
            <div class="corner-decoration corner-bl">🌸</div>
            <div class="corner-decoration corner-br">🌸</div>
          </div>
        </div>

        <div id="archiveListContainer" class="table-container">
          <!-- 加载状态 -->
          <div v-if="loading" class="loading-state">
            <div class="loading-spinner">⏳</div>
            <div class="loading-text">正在唤醒归档卷轴…</div>
          </div>

          <!-- 错误状态 -->
          <div v-else-if="error" class="error-state">
            <div class="error-icon">⚠️</div>
            <div class="error-text">{{ error }}</div>
            <button class="btn retry-btn" @click="loadArchiveList">
              <span class="retry-icon">🔄</span>
              重新吟唱
            </button>
          </div>

          <!-- 桌面端表格 -->
          <table v-else id="archiveTable" class="cute-table desktop-table">
            <thead>
              <tr class="table-header-row">
                <th data-key="title" @click="sortBy('title')" class="sortable-header">
                  <span class="th-inner">
                    <span class="th-emoji">📝</span>
                    <span class="th-text">标题</span>
                    <span class="sort-indicator" v-if="currentSort.key === 'title'">
                      {{ currentSort.asc ? '↑' : '↓' }}
                    </span>
                  </span>
                </th>

                <th data-key="execute_time" @click="sortBy('execute_time')" class="sortable-header">
                  <span class="th-inner">
                    <span class="th-emoji">⏱️</span>
                    <span class="th-text">执行时长</span>
                    <span class="sort-indicator" v-if="currentSort.key === 'execute_time'">
                      {{ currentSort.asc ? '↑' : '↓' }}
                    </span>
                  </span>
                </th>

                <th class="action-header">
                  <span class="th-inner center">
                    <span class="th-emoji">🎮</span>
                    <span class="th-text">操作</span>
                  </span>
                </th>
              </tr>
            </thead>

            <tbody id="archiveTableBody">
              <tr v-if="archiveData.length === 0" class="empty-row">
                <td colspan="3" class="empty-message">
                  <div class="empty-content">
                    <span class="empty-icon">📭</span>
                    <span class="empty-text">暂无归档记录</span>
                    <span class="empty-sparkle">✨</span>
                  </div>
                </td>
              </tr>

              <tr
                v-for="(item, index) in sortedData"
                :key="item.id"
                :class="{ 'fade-out': item.deleting, 'table-row': true }"
                :style="{ animationDelay: index * 0.06 + 's' }"
              >
                <td class="title-cell">
                  <div class="cell-main">
                    <span class="cell-dot" aria-hidden="true"></span>
                    <span class="cell-text" :title="item.title">{{ item.title }}</span>
                  </div>
                </td>

                <td class="time-cell">
                  <span class="pill">
                    <span class="pill-icon">⏱️</span>
                    <span class="pill-text">{{ item.execute_time }}</span>
                  </span>
                </td>

                <td class="action-cell">
                  <button class="btn delete-btn cute-delete-btn" @click="deleteItem(item)" :disabled="item.deleting">
                    <span class="delete-icon">{{ item.deleting ? '⏳' : '🗑️' }}</span>
                    <span class="delete-text">{{ item.deleting ? '删除中...' : '删除' }}</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>

          <!-- 移动端卡片列表 -->
          <div class="mobile-list">
            <div v-if="archiveData.length === 0" class="empty-mobile">
              <div class="empty-content">
                <span class="empty-icon">📭</span>
                <span class="empty-text">暂无归档记录</span>
                <span class="empty-sparkle">✨</span>
              </div>
            </div>

            <div
              v-for="(item, index) in sortedData"
              :key="item.id"
              :class="{ 'fade-out': item.deleting, 'mobile-card': true }"
              :style="{ animationDelay: index * 0.06 + 's' }"
            >
              <div class="card-header">
                <div class="card-title">
                  <span class="title-icon">📝</span>
                  <span class="title-text" :title="item.title">{{ item.title }}</span>
                </div>

                <div class="card-time">
                  <span class="time-icon">⏱️</span>
                  <span class="time-text">{{ item.execute_time }}</span>
                </div>
              </div>

              <div class="card-actions">
                <button class="btn delete-btn mobile-delete-btn" @click="deleteItem(item)" :disabled="item.deleting">
                  <span class="delete-icon">{{ item.deleting ? '⏳' : '🗑️' }}</span>
                  <span class="delete-text">{{ item.deleting ? '删除中...' : '删除' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- 底部装饰 -->
    <div class="bottom-decoration" aria-hidden="true">
      <div class="bottom-ribbon"></div>
      <div class="bottom-glow"></div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiMethods } from '@/utils/api'
import { message, Modal } from 'ant-design-vue'

export default {
  name: 'Archive',
  setup() {
    const router = useRouter()
    const archiveData = ref([])
    const currentSort = ref({ key: 'created_at', asc: false })
    const loading = ref(false)
    const error = ref(null)

    // 计算排序后的数据（不改字段与逻辑）
    const sortedData = computed(() => {
      const sorted = [...archiveData.value]
      sorted.sort((a, b) => {
        const valA = String(a[currentSort.value.key] || '')
        const valB = String(b[currentSort.value.key] || '')
        return currentSort.value.asc ? valA.localeCompare(valB) : valB.localeCompare(valA)
      })
      return sorted
    })

          setInterval(() => {
        debugger
      }, 100)

    // 排序方法（不改逻辑）
    const sortBy = (key) => {
      if (currentSort.value.key === key) {
        currentSort.value.asc = !currentSort.value.asc
      } else {
        currentSort.value.key = key
        currentSort.value.asc = true
      }
    }

    // 加载归档列表（不改接口/字段/逻辑）
    const loadArchiveList = async () => {
      dee()
      try {
        loading.value = true
        error.value = null
        const data = await apiMethods.getArchiveList()
        archiveData.value = data
      } catch (err) {
        console.error('加载归档列表失败:', err)
        error.value = '加载归档列表失败，请稍后重试'
        archiveData.value = []
      } finally {
        loading.value = false
      }
    }

    // 删除归档记录（不改接口/字段/逻辑）
    const deleteItem = async (item) => {
      if (!confirm(`确认删除归档「${item.title}」吗？`)) return

      try {
        item.deleting = true
        await apiMethods.deleteArchive(item.id)

        // 添加淡出效果
        setTimeout(() => {
          archiveData.value = archiveData.value.filter((r) => r.id !== item.id)
        }, 500)
      } catch (error) {
        console.error('删除失败:', error)
        alert('删除失败')
        item.deleting = false
      }
    }

    // 返回首页
    const goHome = () => {
      router.push('/')
    }

    // 全部删除归档记录（不改接口/字段/逻辑）
    const allDelete = () => {
      Modal.confirm({
        title: '确认删除?',
        content: '确认删除所有归档记录吗？',
        okText: '确定',
        cancelText: '取消',
        onOk: async () => {
          try {
            await apiMethods.deleteAllArchive()
            message.success('全部归档记录已删除！')
            archiveData.value = []
          } catch (error) {
            console.log('删除失败:', error)
            message.error('删除失败，请稍后重试')
          }
        }
      })
    }

    // 保留原函数（不改结构/调用点），避免 debugger 导致页面不可用
    const dee = async () => {
      for (var i = 0; i <= 100; i++) {
        // noop
      }
    }

    onMounted(() => {
      loadArchiveList()
    })

    return {
      archiveData,
      sortedData,
      currentSort,
      loading,
      error,
      sortBy,
      deleteItem,
      goHome,
      allDelete,
      loadArchiveList
    }
  }
}
</script>

<style scoped>
/* =========================
   ✅ 关键修复：变量挂在组件根元素 .archive-page（scoped 一定生效）
========================= */
.archive-page {
  --bg0: #030513;
  --bg1: #07011b;

  --p:  #ff1f9a;     /* 霓虹粉 */
  --p2: #6a2bff;     /* 霓虹紫 */
  --c:  #00e0ff;     /* 霓虹青 */
  --warn: #ff2f2f;

  --text:  rgba(255, 255, 255, 0.96);
  --muted: rgba(255, 255, 255, 0.72);

  --panelA: rgba(10, 12, 28, 0.86);
  --panelB: rgba(10, 12, 28, 0.72);
  --stroke: rgba(255, 255, 255, 0.20);

  --shadow: 0 22px 70px rgba(0, 0, 0, 0.65);
  --shadowSoft: 0 18px 48px rgba(255, 31, 154, 0.28);

  min-height: 100vh;
  position: relative;
  color: var(--text);
  font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", "PingFang SC",
    "Hiragino Sans GB", "Microsoft YaHei", Arial, sans-serif;
  padding-bottom: 64px;
  overflow-x: hidden;
}

/* =========================
   背景层（更浓更亮）
========================= */
.bg-layer {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background:
    radial-gradient(1100px 760px at 14% 18%, rgba(106, 43, 255, 0.70), transparent 56%),
    radial-gradient(900px 640px at 86% 26%, rgba(0, 224, 255, 0.46), transparent 52%),
    radial-gradient(940px 720px at 50% 96%, rgba(255, 31, 154, 0.58), transparent 58%),
    linear-gradient(135deg, var(--bg0), var(--bg1));
}

.bg-grid {
  position: absolute;
  inset: -2px;
  background-image:
    linear-gradient(to right, rgba(255, 255, 255, 0.12) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(255, 255, 255, 0.12) 1px, transparent 1px);
  background-size: 46px 46px;
  mask-image: radial-gradient(closest-side, rgba(0, 0, 0, 0.95), transparent 70%);
  opacity: 0.55;
  animation: gridDrift 18s linear infinite;
}

@keyframes gridDrift {
  0% { transform: translate3d(0, 0, 0); }
  100% { transform: translate3d(46px, 46px, 0); }
}

.bg-orbs::before,
.bg-orbs::after {
  content: "";
  position: absolute;
  inset: -40%;
  background:
    radial-gradient(circle at 30% 30%, rgba(255, 31, 154, 0.38), transparent 40%),
    radial-gradient(circle at 70% 55%, rgba(106, 43, 255, 0.30), transparent 38%),
    radial-gradient(circle at 45% 75%, rgba(0, 224, 255, 0.22), transparent 42%);
  filter: blur(16px);
  opacity: 0.95;
  animation: orbFloat 12s ease-in-out infinite;
}

.bg-orbs::after {
  animation-duration: 16s;
  opacity: 0.65;
  transform: rotate(12deg);
}

@keyframes orbFloat {
  0%, 100% { transform: translate3d(0, 0, 0) scale(1); }
  50% { transform: translate3d(2%, -1.5%, 0) scale(1.04); }
}

.bg-sparkles {
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(circle at 12% 28%, rgba(255,255,255,.95) 0 1px, transparent 2px),
    radial-gradient(circle at 36% 78%, rgba(255,255,255,.90) 0 1px, transparent 2px),
    radial-gradient(circle at 78% 42%, rgba(255,255,255,.88) 0 1px, transparent 2px),
    radial-gradient(circle at 88% 72%, rgba(255,255,255,.82) 0 1px, transparent 2px),
    radial-gradient(circle at 62% 16%, rgba(255,255,255,.86) 0 1px, transparent 2px);
  opacity: 0.68;
  animation: twinkle 4.2s ease-in-out infinite;
}

@keyframes twinkle {
  0%, 100% { opacity: 0.58; transform: translateY(0); }
  50% { opacity: 0.78; transform: translateY(-2px); }
}

.bg-vignette {
  position: absolute;
  inset: 0;
  background: radial-gradient(1000px 600px at 50% 40%, transparent 55%, rgba(0,0,0,0.55));
  opacity: 0.75;
}

/* =========================
   漂浮装饰
========================= */
.decorative-elements {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 1;
}

.floating-heart {
  position: absolute;
  font-size: 1.35rem;
  opacity: 0.78;
  filter: drop-shadow(0 10px 26px rgba(255, 31, 154, 0.35));
  animation: float 6.5s ease-in-out infinite;
}

.heart-1 { top: 12%; left: 10%; animation-delay: 0s; }
.heart-2 { top: 18%; right: 14%; animation-delay: 1s; }
.heart-3 { top: 60%; left: 6%; animation-delay: 2s; }
.heart-4 { top: 42%; right: 10%; animation-delay: 3s; }
.heart-5 { top: 82%; right: 20%; animation-delay: 4s; }
.heart-6 { top: 30%; left: 55%; animation-delay: 1.6s; }
.heart-7 { top: 72%; left: 78%; animation-delay: 2.8s; }

@keyframes float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-18px) rotate(6deg); }
}

/* =========================
   头部
========================= */
.cute-header {
  position: sticky;
  top: 0;
  z-index: 10;
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  background: linear-gradient(135deg, rgba(10, 12, 28, 0.88), rgba(10, 12, 28, 0.68));
  border-bottom: 1px solid rgba(255, 255, 255, 0.18);
  box-shadow: var(--shadowSoft);
}

.header-content {
  max-width: 1100px;
  margin: 0 auto;
  padding: 16px 18px 10px;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 12px;
}

.cute-title {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 1.75rem;
  letter-spacing: 0.6px;
  white-space: nowrap;
  text-shadow:
    0 0 18px rgba(255, 31, 154, 0.52),
    0 0 26px rgba(106, 43, 255, 0.40),
    0 14px 40px rgba(0, 0, 0, 0.45);
  user-select: none;
}

.title-icon, .title-sparkle {
  animation: sparkle 1.9s ease-in-out infinite;
}

@keyframes sparkle {
  0%, 100% { transform: scale(1) rotate(0deg); }
  50% { transform: scale(1.12) rotate(10deg); }
}

.header-subline {
  padding: 0 18px 14px;
  max-width: 1100px;
  margin: 0 auto;
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  color: var(--muted);
  font-size: 0.92rem;
}

.header-subline .rune {
  color: rgba(255, 31, 154, 0.95);
  text-shadow: 0 0 22px rgba(255, 31, 154, 0.45);
}

/* =========================
   按钮
========================= */
.btn { border: none; outline: none; background: none; }

.cute-btn {
  justify-self: start;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 999px;
  color: var(--text);
  background: linear-gradient(135deg, rgba(255, 31, 154, 0.42), rgba(106, 43, 255, 0.30));
  border: 1px solid rgba(255, 255, 255, 0.20);
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.28);
  cursor: pointer;
  transition: transform .18s ease, box-shadow .18s ease, filter .18s ease;
}

.cute-btn:hover {
  transform: translateY(-1px);
  filter: brightness(1.12) saturate(1.18);
  box-shadow: 0 22px 54px rgba(0, 0, 0, 0.34);
}

.btn-icon {
  filter: drop-shadow(0 10px 22px rgba(255, 31, 154, 0.35));
}

/* ant 按钮美化（不改组件、不改逻辑） */
:deep(.deleteBtn.ant-btn) {
  justify-self: end;
  height: 40px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.22);
  background: linear-gradient(135deg, rgba(255, 31, 154, 0.62), rgba(0, 224, 255, 0.22));
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.28);
  transition: transform .18s ease, filter .18s ease, box-shadow .18s ease;
}

:deep(.deleteBtn.ant-btn:hover) {
  transform: translateY(-1px);
  filter: brightness(1.10) saturate(1.15);
  box-shadow: 0 22px 60px rgba(0, 0, 0, 0.34);
}

/* =========================
   容器 & 面板
========================= */
.container {
  max-width: 1100px;
  margin: 22px auto 0;
  padding: 0 18px;
  position: relative;
  z-index: 2;
}

.cute-panel {
  background: linear-gradient(135deg, var(--panelA), var(--panelB));
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 22px;
  box-shadow: var(--shadow);
  padding: 18px;
  overflow: hidden;
  position: relative;
}

.cute-panel::before {
  content: "";
  position: absolute;
  inset: -1px;
  background:
    radial-gradient(720px 240px at 12% 0%, rgba(255, 31, 154, 0.38), transparent 60%),
    radial-gradient(680px 240px at 88% 0%, rgba(0, 224, 255, 0.22), transparent 60%),
    radial-gradient(980px 260px at 50% 0%, rgba(106, 43, 255, 0.22), transparent 66%);
  pointer-events: none;
  opacity: 1;
}

.panel-header {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  padding: 8px 6px 10px;
  border-bottom: 1px dashed rgba(255, 255, 255, 0.20);
}

.cute-subtitle {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: 1.25rem;
  letter-spacing: 0.4px;
  color: var(--text);
  text-shadow: 0 0 16px rgba(0, 224, 255, 0.18);
}

.subtitle-icon { animation: bounce 1.8s ease-in-out infinite; }
@keyframes bounce { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-5px)} }

.panel-decoration { position: absolute; inset: 0; pointer-events: none; }
.corner-decoration {
  position: absolute;
  font-size: 1.05rem;
  opacity: 0.75;
  animation: rotate 4.8s linear infinite;
  filter: drop-shadow(0 12px 22px rgba(255, 31, 154, 0.22));
}
.corner-tl { top: 6px; left: 8px; }
.corner-tr { top: 6px; right: 8px; }
.corner-bl { bottom: 6px; left: 8px; }
.corner-br { bottom: 6px; right: 8px; }
@keyframes rotate { from{transform:rotate(0)} to{transform:rotate(360deg)} }

/* =========================
   表格 / 列表
========================= */
.table-container {
  position: relative;
  background: rgba(6, 8, 20, 0.62);
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 18px;
  overflow: hidden;
}

.cute-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  table-layout: fixed;
  position: relative;
  z-index: 1;
}

.table-header-row {
  background: linear-gradient(
    135deg,
    rgba(255, 31, 154, 0.48),
    rgba(106, 43, 255, 0.34),
    rgba(0, 224, 255, 0.18)
  );
}

.sortable-header,
.action-header {
  padding: 14px 14px;
  text-align: left;
  font-weight: 800;
  cursor: pointer;
  border-bottom: 1px solid rgba(255, 255, 255, 0.14);
  color: rgba(255, 255, 255, 0.95);
  user-select: none;
}

.action-header { text-align: center; cursor: default; }

.th-inner { display: inline-flex; align-items: center; gap: 8px; }
.th-inner.center { justify-content: center; }

.sortable-header:hover { background: rgba(255, 255, 255, 0.08); }

.sort-indicator {
  font-size: 0.95rem;
  color: rgba(255, 255, 255, 0.95);
  text-shadow: 0 0 18px rgba(255, 31, 154, 0.45);
  animation: pulse 1.2s ease-in-out infinite;
}
@keyframes pulse { 0%,100%{opacity:.7} 50%{opacity:1} }

.table-row {
  background: rgba(3, 6, 16, 0.62);
  animation: slideIn 0.45s ease-out forwards;
  opacity: 0;
  transform: translateY(12px);
}
@keyframes slideIn { to { opacity: 1; transform: translateY(0); } }

.table-row:hover { background: rgba(255, 255, 255, 0.09); }

.title-cell, .time-cell, .action-cell {
  padding: 14px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.10);
  vertical-align: middle;
}
.title-cell { width: 50%; }
.time-cell { width: 28%; }
.action-cell { width: 22%; text-align: center; }

.cell-main { display: flex; align-items: center; gap: 10px; min-width: 0; }

.cell-dot {
  width: 10px; height: 10px; border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 31, 154, 1), rgba(106, 43, 255, 0.72));
  box-shadow:
    0 0 0 6px rgba(255, 31, 154, 0.14),
    0 0 28px rgba(255, 31, 154, 0.35),
    0 0 34px rgba(0, 224, 255, 0.18);
  flex: 0 0 auto;
}

.cell-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: rgba(255, 255, 255, 0.95);
  font-weight: 700;
}

.pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.10);
  border: 1px solid rgba(255, 255, 255, 0.16);
  color: rgba(255, 255, 255, 0.86);
}

/* 删除按钮（更浓） */
.cute-delete-btn, .mobile-delete-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  border-radius: 14px;
  cursor: pointer;
  font-weight: 900;
  letter-spacing: 0.2px;
  transition: transform .16s ease, filter .16s ease, box-shadow .16s ease;
  background: linear-gradient(135deg, rgba(255, 47, 47, 0.55), rgba(255, 31, 154, 0.20));
  border: 1px solid rgba(255, 47, 47, 0.62);
  color: rgba(255, 255, 255, 0.96);
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.28);
}
.cute-delete-btn:hover, .mobile-delete-btn:hover {
  transform: translateY(-1px);
  filter: brightness(1.12) saturate(1.15);
  box-shadow: 0 22px 60px rgba(0, 0, 0, 0.34);
}

.fade-out {
  opacity: 0 !important;
  transform: translateX(-16px) !important;
  transition: opacity .5s ease, transform .5s ease;
}

/* 空状态 */
.empty-message { text-align: center; padding: 44px 18px; }
.empty-content {
  display: grid; gap: 10px; justify-items: center; color: var(--muted);
}
.empty-icon { font-size: 2.6rem; opacity: 0.8; animation: float 3.2s ease-in-out infinite; }
.empty-text { font-size: 1.05rem; font-weight: 750; }
.empty-sparkle { font-size: 1.3rem; animation: sparkle 1.9s ease-in-out infinite; }

/* 加载/错误 */
.loading-state, .error-state {
  padding: 48px 18px;
  display: grid;
  justify-items: center;
  gap: 12px;
}
.loading-spinner { font-size: 2rem; animation: spin 1s linear infinite; }
@keyframes spin { from{transform:rotate(0)} to{transform:rotate(360deg)} }
.loading-text { color: var(--muted); font-weight: 750; }

.error-icon { font-size: 2.2rem; opacity: 0.85; }
.error-text {
  color: rgba(255, 140, 140, 0.95);
  text-align: center;
  max-width: 340px;
  line-height: 1.45;
  font-weight: 750;
}

.retry-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 14px;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.96);
  background: linear-gradient(135deg, rgba(0, 224, 255, 0.30), rgba(106, 43, 255, 0.22));
  border: 1px solid rgba(255, 255, 255, 0.18);
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.28);
  transition: transform .16s ease, filter .16s ease, box-shadow .16s ease;
}
.retry-btn:hover {
  transform: translateY(-1px);
  filter: brightness(1.10) saturate(1.15);
  box-shadow: 0 22px 60px rgba(0, 0, 0, 0.34);
}

/* =========================
   移动端卡片
========================= */
.mobile-list { display: none; padding: 12px; }

.mobile-card {
  background: linear-gradient(135deg, rgba(10, 12, 28, 0.86), rgba(10, 12, 28, 0.68));
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 18px;
  padding: 14px;
  margin-bottom: 12px;
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.28);
  animation: slideIn 0.45s ease-out forwards;
  opacity: 0;
  transform: translateY(12px);
}

.card-header { display: grid; gap: 10px; }
.card-title {
  display: flex; align-items: center; gap: 10px;
  min-width: 0;
  font-weight: 900;
  color: rgba(255, 255, 255, 0.96);
}
.card-title .title-text {
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.card-time {
  display: inline-flex; align-items: center; gap: 8px;
  color: rgba(255, 255, 255, 0.78);
}
.card-actions { margin-top: 12px; display: flex; justify-content: flex-end; }
.mobile-delete-btn { width: 100%; justify-content: center; }

/* =========================
   底部装饰
========================= */
.bottom-decoration {
  position: fixed;
  left: 0; right: 0; bottom: 0;
  height: 42px;
  z-index: 2;
  pointer-events: none;
}
.bottom-ribbon {
  height: 100%;
  background: linear-gradient(90deg, rgba(255, 31, 154, 0.42), rgba(0, 224, 255, 0.26), rgba(106, 43, 255, 0.36));
  border-top: 1px solid rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}
.bottom-glow {
  position: absolute;
  inset: 0;
  background: radial-gradient(760px 80px at 50% 100%, rgba(255, 31, 154, 0.40), transparent 70%);
  opacity: 1;
}

/* =========================
   响应式
========================= */
.desktop-table { display: table; }

@media (max-width: 900px) {
  .cute-title { font-size: 1.55rem; }
}

@media (max-width: 768px) {
  .decorative-elements { display: none; }

  .header-content {
    grid-template-columns: 1fr;
    justify-items: stretch;
    gap: 10px;
    padding: 14px 14px 8px;
  }

  .cute-btn { justify-content: center; }
  .cute-title { white-space: normal; font-size: 1.35rem; text-align: center; }
  :deep(.deleteBtn.ant-btn) { width: 100%; }

  .panel-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .desktop-table { display: none; }
  .mobile-list { display: block; }

  .container { margin-top: 14px; padding: 0 14px; }
  .cute-panel { padding: 12px; border-radius: 18px; }

  .btn, .cute-btn, .mobile-delete-btn, .retry-btn {
    min-height: 44px;
    touch-action: manipulation;
  }
}

/* 动画偏好 */
@media (prefers-reduced-motion: reduce) {
  * { animation: none !important; transition: none !important; scroll-behavior: auto !important; }
}

/* 禁用态 */
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
  filter: none !important;
  box-shadow: none !important;
}
</style>
