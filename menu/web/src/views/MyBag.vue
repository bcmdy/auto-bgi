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
.kawaii-page { padding: 20px; }
.kawaii-header { display:flex; align-items:center; justify-content:space-between; margin-bottom: 16px; }
.main-title { margin:0; }
.data-display-section { background: #fff; padding: 18px; border-radius: 12px; box-shadow: 0 6px 18px rgba(0,0,0,0.06); }
.category-card { margin-bottom: 12px; border: 1px dashed #eee; border-radius: 10px; overflow: hidden; }
.category-header { display:flex; align-items:center; justify-content:space-between; padding: 10px 14px; background: linear-gradient(90deg,#fff,#fafafa); cursor: pointer; }
.category-body { padding: 12px; }
.item-grid { display:flex; flex-wrap:wrap; gap:12px; }
.item-card { width: 160px; display:flex; gap:10px; align-items:center; padding:8px; border-radius:8px; background:#fff8f8; box-shadow: 0 4px 10px rgba(255,182,193,0.08); }
.item-icon { width:48px; height:48px; object-fit:contain; border-radius:6px; }
.item-meta { display:flex; flex-direction:column; }
.item-name { font-weight:600; }
.item-num { color:#888; font-size:12px; }
.empty-state { text-align:center; padding:40px; }
.loading-state { text-align:center; padding:30px; }
.toggle-btn { background:transparent; border:0; color:#666; }
</style>
