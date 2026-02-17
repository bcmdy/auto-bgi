<template>
  <div class="kawaii-page">
    <div class="main-container">
      <header class="kawaii-header">
        <div class="header-actions left">
          <button @click="goBack" class="kawaii-btn home-btn icon-btn">◀ 返回</button>
        </div>
        <div class="title-box">
          <h1 class="main-title">🎒 我的背包</h1>
          <span class="sub-title">查看背包分类与物品数量</span>
        </div>
      </header>

      <main class="data-display-section">
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner">🍥</div>
          <p>加载中...</p>
        </div>

        <div v-else>
          <div v-if="Object.keys(bagInfo).length === 0" class="empty-state">
            <div class="empty-img">🎒</div>
            <p class="empty-text">背包空空如也～</p>
          </div>

          <div v-else class="bag-categories">
            <div v-for="(items, category) in bagInfo" :key="category" class="category-card">
              <div class="category-header" @click="toggleCategory(category)">
                <span class="category-title">{{ category }}</span>
                <span class="category-count">{{ items.length }} 项</span>
                <button class="toggle-btn">{{ openedCategories.includes(category) ? '收起' : '展开' }}</button>
              </div>

              <transition name="fade">
                <div v-show="openedCategories.includes(category)" class="category-body">
                  <div class="item-grid">
                    <div v-for="it in items" :key="it.ID" class="item-card">
                      <img :src="it.Icon" class="item-icon" alt="icon"/>
                      <div class="item-meta">
                        <div class="item-name">{{ it.Name }}</div>
                        <div class="item-num">数量: {{ it.Number }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </transition>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script>
import api from '@/utils/api'
import { message } from 'ant-design-vue'

export default {
  name: 'MyBag',
  data() {
    return {
      bagInfo: {},
      loading: true,
      openedCategories: []
    }
  },
  async mounted() {
    await this.loadBagInfo();
  },
  methods: {
    goBack() { this.$router.back(); },
    toggleCategory(category) {
      const idx = this.openedCategories.indexOf(category);
      if (idx === -1) this.openedCategories.push(category);
      else this.openedCategories.splice(idx, 1);
    },
    async loadBagInfo() {
      try {
        this.loading = true;
        const data = await api.get('/api/BagStatistics/getBagInfo');
        // 接口返回的是对象，每个键是分类
        this.bagInfo = data || {};
        // 默认全部展开
        this.openedCategories = Object.keys(this.bagInfo);
      } catch (error) {
        console.error('加载背包信息失败', error);
        message.error('加载背包信息失败');
      } finally {
        this.loading = false;
      }
    }
  }
}
</script>

<style scoped>
/* Page layout */
.kawaii-page { padding: 18px; }
.kawaii-header { display:flex; align-items:center; justify-content:space-between; gap:12px; margin-bottom: 14px; }
.main-title { margin:0; font-size:18px; }
.sub-title { color:#777; font-size:12px; }

.data-display-section { background: #fff; padding: 14px; border-radius: 14px; box-shadow: 0 8px 20px rgba(0,0,0,0.06); }

.category-card { margin-bottom: 12px; border-radius: 12px; overflow: hidden; border: 1px solid rgba(230,230,230,0.9); }
.category-header {
  display:flex;
  align-items:center;
  justify-content:space-between;
  padding: 12px 14px;
  background: linear-gradient(90deg,#fff,#fcfcfd);
  cursor: pointer;
  user-select: none;
}
.category-header .category-title { font-weight:700; color:#333; }
.category-header .category-count { color:#888; font-size:13px; margin-left:8px; }
.category-header .toggle-btn { background:transparent; border:0; color:#666; font-weight:600; }

.category-body { padding: 12px; background: linear-gradient(180deg,#fff,#fffaf7); }

.item-grid { display:flex; flex-wrap:wrap; gap:12px; }
.item-card {
  width: 160px;
  display:flex;
  gap:10px;
  align-items:center;
  padding:8px 10px;
  border-radius:10px;
  background: linear-gradient(180deg,#fff,#fff6f6);
  box-shadow: 0 6px 14px rgba(255,182,193,0.06);
  transition: transform .12s ease, box-shadow .12s ease;
}
.item-card:hover { transform: translateY(-4px); box-shadow: 0 14px 30px rgba(255,182,193,0.12); }
.item-icon { width:56px; height:56px; object-fit:contain; border-radius:8px; background: transparent; }
.item-meta { display:flex; flex-direction:column; min-width:0; }
.item-name { font-weight:700; color:#2b2b2b; font-size:14px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.item-num { color:#666; font-size:12px; margin-top:4px; }

.empty-state { text-align:center; padding:36px; color:#666; }
.loading-state { text-align:center; padding:28px; }

/* Small helper for touch targets */
.category-header, .toggle-btn { -webkit-tap-highlight-color: transparent; }

/* Mobile / narrow screens */
@media (max-width: 700px) {
  .kawaii-page { padding: 12px; }
  .main-title { font-size:16px; }

  /* Make category cards full-width and bolder header */
  .category-card { border-radius: 10px; }
  .category-header { padding: 14px 12px; }
  .category-header .category-title { font-size:16px; }
  .category-header .toggle-btn { font-size:13px; padding:6px 10px; border-radius:8px; background: #fff; box-shadow: 0 6px 12px rgba(0,0,0,0.04); }

  /* Single-column item list for better readability on mobile */
  .item-grid { display:block; }
  .item-card { width:100%; padding:12px; gap:12px; }
  .item-icon { width:64px; height:64px; }
  .item-name { font-size:15px; }
  .item-num { font-size:13px; }

  /* Increase spacing so items are easier to tap */
  .category-body { padding: 10px; }
}

/* Very small phones */
@media (max-width: 420px) {
  .main-title { font-size:15px; }
  .category-header { padding: 12px; }
  .item-icon { width:56px; height:56px; }
  .item-name { font-size:14px; }
}

</style>
