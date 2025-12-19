<template>
  <div class="page">
    <!-- 背景装饰层（不影响逻辑） -->
    <div class="bg-layer" aria-hidden="true">
      <div class="bg-orb orb-a"></div>
      <div class="bg-orb orb-b"></div>
      <div class="bg-orb orb-c"></div>
      <div class="bg-grid"></div>
      <div class="bg-sparkles">
        <i class="sp s1"></i><i class="sp s2"></i><i class="sp s3"></i><i class="sp s4"></i><i class="sp s5"></i>
        <i class="sp s6"></i><i class="sp s7"></i><i class="sp s8"></i><i class="sp s9"></i><i class="sp s10"></i>
      </div>
    </div>

    <header class="topbar">
      <div class="topbar__inner">
        <div class="btn-container">
          <button class="btn home-btn" @click="goHome">
            <span class="btn-icon">🏠</span>
            <span class="btn-text">返回主页</span>
            <span class="btn-sparkle">✨</span>
          </button>

          <button class="btn clean-btn" @click="deleteBag">
            <span class="btn-icon">🧹</span>
            <span class="btn-text">清理统计，只保留一天</span>
            <span class="btn-sparkle">💫</span>
          </button>

          <!-- 转为变化图 -->
          <button class="btn trend-btn" @click="goBagStatisticsTrend">
            <span class="btn-icon">📈</span>
            <span class="btn-text">转为变化图</span>
            <span class="btn-sparkle">⚡</span>
          </button>
        </div>

        <div class="title-wrap">
          <div class="title-badge">✦</div>
          <h1 class="page-title">{{ title }}</h1>
          <div class="title-sub">「Chuunibyou Inventory Archive」</div>
        </div>
      </div>
    </header>

    <main class="main">
      <!-- 筛选区 -->
      <div class="container filter-section">
        <button class="btn btn-ghost" @click="toggleFilter" style="margin-bottom:10px;">
          <span class="btn-icon">{{ filterCollapsed ? '🧩' : '🪄' }}</span>
          {{ filterCollapsed ? '展开材料筛选' : '收起材料筛选' }}
          <span class="btn-sparkle">{{ filterCollapsed ? '⟡' : '✧' }}</span>
        </button>

        <div v-show="!filterCollapsed" class="filter-container">
          <div class="filter-header">
            <h3 class="filter-title">
              <span class="filter-icon dancing">🎀</span>
              <span class="title-text">
                <span class="title-main">材料筛选</span>
                <span class="title-sub">Material Filter</span>
              </span>
              <span class="filter-icon dancing">🎀</span>
            </h3>

            <div class="filter-buttons">
              <button class="filter-btn cancel-btn" @click="cancelSelection">
                <span class="btn-icon">✨</span>
                <span class="btn-text">取消选择</span>
                <span class="btn-wave">〜</span>
              </button>
              <button class="filter-btn ore-btn" @click="selectAllOre">
                <span class="btn-icon">💎</span>
                <span class="btn-text">选择矿石</span>
                <span class="btn-wave">〜</span>
              </button>
            </div>
          </div>

          <div class="checkboxes-container" v-if="!isLoading">
            <div
              v-for="material in uniqueMaterials"
              :key="material"
              class="checkbox-item"
              :class="{ selected: selectedMaterials.includes(material) }"
            >
              <input
                type="checkbox"
                :id="'material-' + material"
                :value="material"
                v-model="selectedMaterials"
                @change="filterTable"
                class="cute-checkbox"
              />
              <label :for="'material-' + material" class="checkbox-label" :title="material">
                <span class="checkbox-custom"></span>
                <span class="material-name">{{ material }}</span>
              </label>
            </div>
          </div>

          <div class="loading-container" v-else>
            <div class="loading-animation">
              <div class="loading-dots">
                <span class="dot"></span>
                <span class="dot"></span>
                <span class="dot"></span>
              </div>
              <p class="loading-text">正在加载材料列表...</p>
            </div>
          </div>
        </div>

        <!-- 超过8000材料 + 黑名单 -->
        <div class="tools-row">
          <button class="filter-btn ore-btn" @click="checkBag">
            <span class="btn-icon">💎</span>
            <span class="btn-text">超过8000材料</span>
            <span class="btn-wave">〜</span>
          </button>

          <button
            class="filter-btn danger-btn"
            @click="openBlackListModal"
            style="margin-left: 10px;"
          >
            <span class="btn-icon">🚫</span>
            <span class="btn-text">黑名单管理</span>
            <span class="btn-wave">〜</span>
          </button>
        </div>
      </div>

      <!-- 数据区 -->
      <div class="container data-section">
        <!-- 桌面端表格 -->
        <table
          id="materialTable"
          class="desktop-table"
          v-if="!isLoading && filteredItems.length > 0"
        >
          <thead>
            <tr>
              <th>统计日期</th>
              <th>材料</th>
              <th>数量</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="(item, index) in filteredItems" :key="index">
              <tr v-if="item.type === 'spacer'" class="spacer-row">
                <td colspan="3"></td>
              </tr>
              <tr v-else class="data-row">
                <td class="mono">{{ item.date }}</td>
                <td class="mat-cell">
                  <span class="mat-pill">{{ item.materialDisplay }}</span>
                </td>
                <td class="num-cell mono">{{ item.numDisplay }}</td>
              </tr>
            </template>
          </tbody>
        </table>

        <!-- 移动端卡片列表 -->
        <div class="mobile-list" v-if="!isLoading && filteredItems.length > 0">
          <div v-for="group in groupedMobileMaterials" :key="group.cl" class="mobile-card">
            <div class="material-card">
              <div class="card-header">
                <span class="date-icon">📦</span>
                <span class="date-text">{{ group.materialDisplay }}</span>
                <span class="card-glow" aria-hidden="true"></span>
              </div>
              <div class="card-content">
                <table class="mobile-inner-table">
                  <thead>
                    <tr>
                      <th>日期</th>
                      <th>数量</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in group.items" :key="item.date">
                      <td class="mono">{{ item.date }}</td>
                      <td class="mono">{{ item.numDisplay }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        <!-- 空状态显示 -->
        <div class="empty-state" v-else-if="!isLoading && filteredItems.length === 0">
          <div class="empty-content">
            <div class="empty-icon">📦</div>
            <h3 class="empty-title">暂无数据</h3>
            <p class="empty-description">
              {{ selectedMaterials.length > 0 ? '当前筛选条件下没有找到相关材料数据' : '还没有任何背包统计数据' }}
            </p>
            <button
              v-if="selectedMaterials.length > 0"
              class="btn empty-btn"
              @click="cancelSelection"
            >
              清除筛选条件
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>超过8000材料</h3>
          <button class="modal-close-btn" @click="closeDetailModal">✕</button>
        </div>

        <div class="modal-body">
          <ul class="detail-list">
            <li v-for="(value, key) in checkBagData" :key="key" class="detail-item">
              <div class="detail-line">
                <span class="detail-text"><b>{{ key }}</b>：{{ value }}</span>

                <span v-if="value > 8000 && blackList.includes(key)" class="blacklist-badge">
                  🚫 已在黑名单
                </span>

                <button
                  v-if="value > 8000 && !blackList.includes(key)"
                  class="add-blacklist-btn"
                  @click="addToBlackList(key)"
                >
                  加入黑名单
                </button>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- 黑名单管理模态框 -->
    <div v-if="showBlackListModal" class="modal-overlay">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>黑名单管理</h3>
          <button class="modal-close-btn" @click="closeBlackListModal">✕</button>
        </div>

        <div class="modal-body">
          <div class="blacklist-container">
            <div class="blacklist-list">
              <h4>当前黑名单：</h4>

              <div v-if="blackList.length === 0" class="empty-mini">
                暂无黑名单材料
              </div>

              <div v-else class="blacklist-grid">
                <div v-for="item in blackList" :key="item" class="blacklist-item">
                  <span class="material-name" :title="item">{{ item }}</span>
                  <button class="remove-btn" @click="removeFromBlackList(item)">移除</button>
                </div>
              </div>

              <div class="hint">
                <span class="hint-dot">✦</span>
                提示：超过 8000 且不想看见的材料，直接从上一个窗口“加入黑名单”即可。
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { apiMethods } from '@/utils/api'
import api from '@/utils/api'

export default {
  name: 'BagStatistics',
  data() {
    return {
      title: '背包统计',
      items: [],
      selectedMaterials: [],
      allOre: ["萃凝晶", "水晶块", "星银矿石", "紫晶块", "白铁块", "铁块", "魔晶块", "石珀", "虹滴晶"],
      isLoading: true,
      filterCollapsed: true,
      showDetailModal: false,
      checkBagData: {},
      showBlackListModal: false,
      blackList: []
    }
  },
  computed: {
    // 处理并排序原始数据
    sortedItems() {
      const processed = this.items.map(item => ({
        date: item.Data || item.date,
        cl: item.Cl || item.cl,
        num: parseInt(item.Num || item.num || 0)
      }));

      // 排序逻辑（原石优先，摩拉第二）
      return processed.sort((a, b) => {
        // 原石第一
        if (a.cl === '原石' && b.cl !== '原石') return -1;
        if (a.cl !== '原石' && b.cl === '原石') return 1;

        // 摩拉第二
        if (a.cl === '摩拉数值' && b.cl !== '摩拉数值') return -1;
        if (a.cl !== '摩拉数值' && b.cl === '摩拉数值') return 1;

        // 其他按名称升序
        return a.cl.localeCompare(b.cl);
      });
    },

    // 获取唯一材料列表
    uniqueMaterials() {
      return [...new Set(this.sortedItems.map(item => item.cl))].sort();
    },

    // 筛选后的数据
    filteredData() {
      return this.selectedMaterials.length === 0
        ? this.sortedItems
        : this.sortedItems.filter(item => this.selectedMaterials.includes(item.cl));
    },

    // 处理显示数据（包括间隔行和数据格式）
    filteredItems() {
      const result = [];
      let lastCl = null;
      let materialMap = {};

      for (let i = 0; i < this.filteredData.length; i++) {
        const { date, cl, num } = this.filteredData[i];

        // 添加间隔行
        if (lastCl !== null && cl !== lastCl) {
          result.push({ type: 'spacer' });
        }
        lastCl = cl;

        // 处理显示文本
        let materialDisplay = cl;
        let numDisplay = num.toString();

        // 原石显示抽数
        if (cl === "原石") {
          const pulls = Math.floor(num / 160);
          if (pulls > 0) {
            materialDisplay = `${cl} (${pulls}抽)`;
          }
        }

        // 显示变化量
        if (materialMap[cl] !== undefined) {
          const prev = materialMap[cl];
          const diff = num - prev.num;
          if (diff !== 0) {
            const sign = diff > 0 ? '+' : '';
            numDisplay = `${num} (${sign}${diff})`;
          }
        }

        materialMap[cl] = { date, num };

        result.push({
          date,
          cl,
          num,
          materialDisplay,
          numDisplay
        });
      }

      return result;
    },

    mindsetMaterials() {
      // 心态相关材料
      return this.filteredItems.filter(item => item.cl === '原石' || item.cl === '摩拉数值');
    },
    otherMaterials() {
      // 其他材料
      return this.filteredItems.filter(item => item.cl !== '原石' && item.cl !== '摩拉数值');
    },

    groupedMobileMaterials() {
      // 手机端下将材料按名称分组，每组为一个卡片，卡片内按日期排序
      const groups = {};
      this.filteredItems.forEach(item => {
        if (item.type === 'spacer') return;
        if (!groups[item.cl]) groups[item.cl] = [];
        groups[item.cl].push(item);
      });
      // 返回分组后的数组，每组包含材料名和数据列表
      return Object.keys(groups).map(cl => {
        const items = groups[cl];
        let materialDisplay;
        if (cl === '原石') {
          // 原石取最后一个
          materialDisplay = items[items.length - 1].materialDisplay;
        } else {
          materialDisplay = items[0].materialDisplay;
        }
        return {
          cl,
          materialDisplay,
          items
        };
      });
    },
  },
  async mounted() {
    await this.loadData();
    await this.loadBlackList();
  },

  methods: {
    // 加载数据
    async loadData() {
      try {
        this.isLoading = true;
        this.items = await apiMethods.getBagStatistics();
      } catch (error) {
        console.error('加载数据失败:', error);
        alert('加载背包统计数据失败，请稍后重试');
      } finally {
        this.isLoading = false;
      }
    },

    // 返回主页
    goHome() {
      this.$router.push('/');
    },

    goBagStatisticsTrend() {
      this.$router.push('/MaterialTrend');
    },

    async checkBag() {
      this.showDetailModal = true
      const data = await api.get('/api/checkBag');
      console.log(data)
      this.checkBagData = data
    },

    closeDetailModal() {
      this.showDetailModal = false
    },

    // 黑名单相关方法
    async loadBlackList() {
      try {
        const response = await apiMethods.getBlackList();
        this.blackList = response.data.BlackLists || [];
      } catch (error) {
        console.error('加载黑名单失败:', error);
      }
    },

    async addToBlackList(materialName) {
      if (this.blackList.includes(materialName)) {
        alert('该材料已在黑名单中');
        return;
      }

      try {
        await apiMethods.addBlackList([materialName]);
        this.blackList.push(materialName);
        alert('已添加到黑名单');
      } catch (error) {
        console.error('添加黑名单失败:', error);
        alert('添加黑名单失败: ' + (error.message || error));
      }
    },

    async removeFromBlackList(materialName) {
      if (!confirm(`确定要从黑名单中移除 ${materialName} 吗？`)) {
        return;
      }

      try {
        await apiMethods.deleteBlackList(materialName);
        this.blackList = this.blackList.filter(item => item !== materialName);
        alert('已从黑名单中移除');
      } catch (error) {
        console.error('移除黑名单失败:', error);
        alert('移除黑名单失败: ' + (error.message || error));
      }
    },

    openBlackListModal() {
      this.showBlackListModal = true;
    },

    closeBlackListModal() {
      this.showBlackListModal = false;
    },

    // 删除背包数据
    async deleteBag() {
      if (!confirm('确定要清理统计数据吗？这将只保留最近一天的数据，其他数据将被删除。')) {
        return;
      }

      try {
        const data = await api.post('/api/deleteBag');
        alert(data.message || '操作成功！已清理统计数据');
        await this.loadData();
      } catch (error) {
        alert("请求出错：" + (error.message || error));
      }
    },

    // 取消选择
    cancelSelection() {
      this.selectedMaterials = [];
    },

    // 选择所有矿石
    selectAllOre() {
      this.selectedMaterials = [...this.allOre];
    },

    // 筛选表格（computed自动触发）
    filterTable() {},

    toggleFilter() {
      this.filterCollapsed = !this.filterCollapsed;
    },
  }
}
</script>

<style scoped>
/* =========================
   高对比「二次元中二」主题
   scoped 下别用 :root / body，统一挂在 .page
========================= */
.page{
  --bg0:#070A12;
  --bg1:#0B1030;
  --panel: rgba(14, 18, 40, .68);
  --panel2: rgba(255,255,255,.06);
  --stroke: rgba(255, 255, 255, .14);

  --txt:#EAF0FF;
  --muted: rgba(234,240,255,.75);

  --primary:#FF4FD8;  /* 霓虹粉 */
  --primary2:#7C4DFF; /* 霓虹紫 */
  --cyan:#27F5FF;     /* 霓虹青 */
  --danger:#FF4B5E;

  --shadow: 0 18px 50px rgba(0,0,0,.45);
  --glow: 0 0 18px rgba(255,79,216,.40), 0 0 36px rgba(39,245,255,.18);
  --radius: 22px;

  min-height: 100vh;
  color: var(--txt);
  font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "PingFang SC","Microsoft YaHei", sans-serif;
  position: relative;
  overflow-x: hidden;
}

/* 背景层 */
.bg-layer{
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background: radial-gradient(1200px 700px at 15% 10%, rgba(255,79,216,.18), transparent 55%),
              radial-gradient(900px 600px at 85% 20%, rgba(39,245,255,.14), transparent 52%),
              radial-gradient(900px 700px at 50% 90%, rgba(124,77,255,.16), transparent 55%),
              linear-gradient(180deg, var(--bg0), var(--bg1));
}
.bg-orb{
  position:absolute;
  width: 520px;
  height: 520px;
  border-radius: 50%;
  filter: blur(22px);
  opacity: .85;
  mix-blend-mode: screen;
  animation: floaty 10s ease-in-out infinite;
}
.orb-a{ left:-180px; top:-160px; background: radial-gradient(circle at 30% 30%, rgba(255,79,216,.9), transparent 60%); }
.orb-b{ right:-220px; top:-120px; background: radial-gradient(circle at 40% 40%, rgba(39,245,255,.8), transparent 62%); animation-delay: -3s;}
.orb-c{ left: 18%; bottom:-260px; background: radial-gradient(circle at 45% 45%, rgba(124,77,255,.85), transparent 65%); animation-delay: -6s;}
@keyframes floaty{
  0%,100%{ transform: translate3d(0,0,0) scale(1); }
  50%{ transform: translate3d(22px,-18px,0) scale(1.04); }
}
.bg-grid{
  position:absolute;
  inset:-2px;
  background-image:
    linear-gradient(to right, rgba(255,255,255,.06) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(255,255,255,.05) 1px, transparent 1px);
  background-size: 64px 64px;
  opacity: .25;
  transform: perspective(900px) rotateX(60deg);
  transform-origin: top center;
  top: 45%;
  height: 120%;
}
.bg-sparkles .sp{
  position:absolute;
  width: 6px; height: 6px;
  border-radius: 50%;
  background: rgba(255,255,255,.9);
  box-shadow: 0 0 14px rgba(255,79,216,.55), 0 0 22px rgba(39,245,255,.28);
  opacity:.6;
  animation: tw 2.8s ease-in-out infinite;
}
@keyframes tw{ 0%,100%{ transform:scale(.8); opacity:.35 } 50%{ transform:scale(1.2); opacity:.9 } }
.s1{ left:10%; top:18% } .s2{ left:18%; top:34% } .s3{ left:32%; top:22% } .s4{ left:48%; top:14% } .s5{ left:62%; top:28% }
.s6{ left:76%; top:18% } .s7{ left:86%; top:34% } .s8{ left:24%; top:58% } .s9{ left:58%; top:62% } .s10{ left:82%; top:70% }

/* 结构层 */
.topbar{
  position: sticky;
  top: 0;
  z-index: 20;
  padding: calc(env(safe-area-inset-top) + 14px) 12px 14px;
  backdrop-filter: blur(14px);
  background: linear-gradient(180deg, rgba(7,10,18,.85), rgba(7,10,18,.35));
  border-bottom: 1px solid rgba(255,255,255,.10);
}
.topbar__inner{
  max-width: 1100px;
  margin: 0 auto;
  border-radius: var(--radius);
  background: linear-gradient(135deg, rgba(255,255,255,.08), rgba(255,255,255,.04));
  border: 1px solid rgba(255,255,255,.12);
  box-shadow: var(--shadow);
  padding: 14px 14px 16px;
  position: relative;
  overflow: hidden;
}
.topbar__inner::before{
  content:"";
  position:absolute;
  inset:-2px;
  background: radial-gradient(600px 180px at 30% 0%, rgba(255,79,216,.22), transparent 60%),
              radial-gradient(520px 200px at 80% 10%, rgba(39,245,255,.16), transparent 62%);
  pointer-events:none;
}

.title-wrap{
  margin-top: 10px;
  text-align: center;
  position: relative;
  z-index: 2;
}
.title-badge{
  width: 44px;
  height: 44px;
  margin: 0 auto 8px;
  border-radius: 14px;
  display:flex;
  align-items:center;
  justify-content:center;
  background: linear-gradient(135deg, rgba(255,79,216,.35), rgba(39,245,255,.18));
  border: 1px solid rgba(255,255,255,.18);
  box-shadow: var(--glow);
  font-weight: 900;
}
.page-title{
  margin: 0;
  font-size: 28px;
  letter-spacing: .5px;
  text-shadow: 0 0 14px rgba(255,79,216,.25);
}
.title-sub{
  margin-top: 6px;
  font-size: 12px;
  color: var(--muted);
  letter-spacing: 1.6px;
  text-transform: uppercase;
}

.main{
  position: relative;
  z-index: 1;
  padding: 18px 12px 36px;
}
.container{
  max-width: 1100px;
  margin: 16px auto;
  padding: 0;
}

/* 按钮 */
.btn-container{
  margin: 0 auto;
  display: flex;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
  position: relative;
  z-index: 2;
}
.btn{
  position: relative;
  border: 1px solid rgba(255,255,255,.14);
  border-radius: 999px;
  padding: 12px 16px;
  font-size: 14px;
  cursor: pointer;
  transition: transform .18s ease, box-shadow .18s ease, filter .18s ease;
  font-weight: 800;
  overflow: hidden;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 170px;
  justify-content: center;
  color: var(--txt);
  background: linear-gradient(135deg, rgba(255,255,255,.10), rgba(255,255,255,.06));
  box-shadow: 0 10px 24px rgba(0,0,0,.28);
}
.btn::before{
  content:"";
  position:absolute;
  top:0; left:-120%;
  width:120%;
  height:100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,.25), transparent);
  transition: left .6s ease;
}
.btn:hover::before{ left:120%; }
.btn:hover{ transform: translateY(-2px); box-shadow: 0 16px 34px rgba(0,0,0,.38), var(--glow); }
.btn:active{ transform: translateY(-1px) scale(.99); }
.btn:focus-visible{
  outline: 2px solid rgba(39,245,255,.65);
  outline-offset: 3px;
}
.btn-icon{ font-size: 16px; }
.btn-sparkle{ opacity: .9; filter: drop-shadow(0 0 10px rgba(255,79,216,.35)); }

.home-btn{
  background: linear-gradient(135deg, rgba(124,77,255,.55), rgba(255,79,216,.35));
  border-color: rgba(124,77,255,.45);
}
.clean-btn{
  background: linear-gradient(135deg, rgba(255,79,216,.60), rgba(255,75,94,.35));
  border-color: rgba(255,79,216,.48);
}
.trend-btn{
  background: linear-gradient(135deg, rgba(39,245,255,.35), rgba(124,77,255,.45));
  border-color: rgba(39,245,255,.35);
}
.btn-ghost{
  background: linear-gradient(135deg, rgba(255,255,255,.10), rgba(255,255,255,.06));
}

/* 筛选区 */
.filter-section{
  border-radius: var(--radius);
  background: linear-gradient(135deg, rgba(255,255,255,.08), rgba(255,255,255,.04));
  border: 1px solid rgba(255,255,255,.12);
  box-shadow: var(--shadow);
  padding: 14px;
}
.filter-container{
  margin-top: 10px;
  border-radius: var(--radius);
  background: linear-gradient(135deg, rgba(14,18,40,.72), rgba(14,18,40,.46));
  border: 1px solid rgba(255,255,255,.12);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.08);
  overflow: hidden;
  position: relative;
}
.filter-container::before{
  content:"";
  position:absolute;
  inset:0;
  background: radial-gradient(520px 180px at 20% 0%, rgba(255,79,216,.20), transparent 60%),
              radial-gradient(520px 200px at 85% 10%, rgba(39,245,255,.14), transparent 62%);
  pointer-events:none;
}
.filter-header{
  display:flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding: 16px 16px 10px;
  position: relative;
  z-index: 1;
}
.filter-title{
  margin: 0;
  display:flex;
  align-items:center;
  gap: 10px;
}
.title-text{
  display:flex;
  flex-direction: column;
  align-items:flex-start;
  gap:2px;
}
.title-main{
  font-size: 18px;
  font-weight: 900;
  background: linear-gradient(90deg, var(--primary), var(--cyan), var(--primary2));
  -webkit-background-clip: text;
  background-clip:text;
  color: transparent;
  text-shadow: 0 0 18px rgba(255,79,216,.18);
}
.title-sub{
  font-size: 12px;
  color: var(--muted);
  letter-spacing: 1.2px;
}
.filter-icon{ font-size: 18px; }
.filter-icon.dancing{ animation: dance 2.8s ease-in-out infinite; }
@keyframes dance{
  0%,100%{ transform: translateY(0) rotate(0deg) }
  50%{ transform: translateY(-3px) rotate(-6deg) }
}

.filter-buttons{ display:flex; gap:10px; flex-wrap: wrap; }

.filter-btn{
  position: relative;
  border-radius: 999px;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 900;
  cursor: pointer;
  display:inline-flex;
  align-items:center;
  gap: 8px;
  border: 1px solid rgba(255,255,255,.14);
  color: var(--txt);
  background: linear-gradient(135deg, rgba(255,255,255,.10), rgba(255,255,255,.06));
  transition: transform .18s ease, box-shadow .18s ease, filter .18s ease;
  box-shadow: 0 10px 20px rgba(0,0,0,.26);
}
.filter-btn:hover{
  transform: translateY(-2px);
  box-shadow: 0 16px 30px rgba(0,0,0,.38), var(--glow);
}
.filter-btn:active{ transform: translateY(-1px); }
.filter-btn .btn-wave{ opacity: .8; }
.cancel-btn{ border-color: rgba(255,79,216,.45); }
.ore-btn{ border-color: rgba(39,245,255,.35); }
.danger-btn{
  background: linear-gradient(135deg, rgba(255,75,94,.62), rgba(255,79,216,.32));
  border-color: rgba(255,75,94,.55);
}
.tools-row{
  display:flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 12px;
}

/* 复选区：更“清晰”+对比更强 */
.checkboxes-container{
  position: relative;
  z-index: 1;
  display:grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 12px;
  padding: 14px 16px 18px;
}
.checkbox-item{ position: relative; }
.cute-checkbox{
  position:absolute;
  opacity:0;
  pointer-events:none;
}
.checkbox-label{
  display:flex;
  align-items:center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 16px;
  background: linear-gradient(135deg, rgba(255,255,255,.10), rgba(255,255,255,.06));
  border: 1px solid rgba(255,255,255,.14);
  box-shadow: 0 10px 20px rgba(0,0,0,.22);
  cursor:pointer;
  transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease;
}
.checkbox-label:hover{
  transform: translateY(-2px);
  border-color: rgba(39,245,255,.30);
  box-shadow: 0 16px 30px rgba(0,0,0,.34), 0 0 18px rgba(39,245,255,.16);
}
.checkbox-custom{
  width: 22px;
  height: 22px;
  border-radius: 10px;
  border: 2px solid rgba(255,255,255,.22);
  background: rgba(0,0,0,.16);
  position: relative;
  flex-shrink: 0;
}
.checkbox-custom::after{
  content:"";
  position:absolute;
  inset: 4px;
  border-radius: 7px;
  background: linear-gradient(135deg, rgba(255,79,216,.85), rgba(39,245,255,.70));
  transform: scale(0);
  transition: transform .16s ease;
}
.cute-checkbox:checked + .checkbox-label{
  border-color: rgba(255,79,216,.55);
  box-shadow: 0 16px 30px rgba(0,0,0,.35), 0 0 22px rgba(255,79,216,.22);
}
.cute-checkbox:checked + .checkbox-label .checkbox-custom::after{
  transform: scale(1);
}
.checkbox-item.selected{ transform: translateY(-1px); }

.material-name{
  max-width: 100%;
  white-space: nowrap;
  overflow:hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  font-weight: 800;
  color: rgba(234,240,255,.92);
}

/* 数据区 */
.data-section{
  border-radius: var(--radius);
  background: linear-gradient(135deg, rgba(255,255,255,.08), rgba(255,255,255,.04));
  border: 1px solid rgba(255,255,255,.12);
  box-shadow: var(--shadow);
  padding: 12px;
}

/* 表格：强对比、可读性优先 */
table{
  width: 100%;
  border-collapse: separate;
  border-spacing: 0 10px;
}
thead th{
  text-align:left;
  padding: 12px 14px;
  font-size: 13px;
  letter-spacing: .6px;
  color: rgba(234,240,255,.92);
  background: linear-gradient(135deg, rgba(255,79,216,.35), rgba(39,245,255,.18));
  border: 1px solid rgba(255,255,255,.12);
}
thead th:first-child{ border-top-left-radius: 16px; border-bottom-left-radius: 16px; }
thead th:last-child{ border-top-right-radius: 16px; border-bottom-right-radius: 16px; }

.data-row td{
  padding: 12px 14px;
  background: linear-gradient(135deg, rgba(14,18,40,.78), rgba(14,18,40,.52));
  border-top: 1px solid rgba(255,255,255,.10);
  border-bottom: 1px solid rgba(255,255,255,.08);
  color: rgba(234,240,255,.90);
}
.data-row td:first-child{
  border-left: 1px solid rgba(255,255,255,.10);
  border-top-left-radius: 16px;
  border-bottom-left-radius: 16px;
}
.data-row td:last-child{
  border-right: 1px solid rgba(255,255,255,.10);
  border-top-right-radius: 16px;
  border-bottom-right-radius: 16px;
}
.data-row:hover td{
  box-shadow: 0 0 0 1px rgba(39,245,255,.18), 0 0 20px rgba(255,79,216,.14);
  transform: translateY(-1px);
}
.mono{
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-variant-numeric: tabular-nums;
}
.mat-pill{
  display:inline-flex;
  align-items:center;
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(255,255,255,.06);
  border: 1px solid rgba(255,255,255,.12);
  box-shadow: 0 0 12px rgba(39,245,255,.10);
}
.num-cell{
  font-weight: 900;
  color: rgba(234,240,255,.95);
}
.spacer-row td{
  height: 10px;
  padding: 0;
  background: transparent;
}

/* 空状态 */
.empty-state{
  display:flex;
  justify-content:center;
  align-items:center;
  min-height: 280px;
  padding: 24px 12px;
}
.empty-content{
  text-align:center;
  max-width: 520px;
  padding: 18px 16px;
  border-radius: var(--radius);
  border: 1px dashed rgba(255,255,255,.18);
  background: linear-gradient(135deg, rgba(14,18,40,.62), rgba(14,18,40,.38));
  box-shadow: var(--shadow);
}
.empty-icon{ font-size: 56px; opacity: .9; filter: drop-shadow(0 0 18px rgba(255,79,216,.25)); }
.empty-title{ margin: 10px 0 6px; font-size: 18px; }
.empty-description{ margin: 0; color: var(--muted); line-height: 1.6; }
.empty-btn{
  margin-top: 14px;
  background: linear-gradient(135deg, rgba(255,79,216,.72), rgba(124,77,255,.45));
  border-color: rgba(255,79,216,.55);
}

/* 移动端卡片 */
.mobile-list{ display:none; gap: 12px; }
.mobile-card{
  border-radius: var(--radius);
  background: linear-gradient(135deg, rgba(14,18,40,.74), rgba(14,18,40,.46));
  border: 1px solid rgba(255,255,255,.12);
  box-shadow: var(--shadow);
  overflow: hidden;
}
.card-header{
  position: relative;
  display:flex;
  align-items:center;
  gap: 10px;
  padding: 12px 14px;
  font-weight: 900;
  background: linear-gradient(135deg, rgba(255,79,216,.30), rgba(39,245,255,.14));
  border-bottom: 1px solid rgba(255,255,255,.10);
}
.card-header .date-text{ flex: 1; }
.card-glow{
  position:absolute;
  inset:-20px;
  background: radial-gradient(260px 120px at 30% 40%, rgba(255,79,216,.22), transparent 60%);
  pointer-events:none;
  filter: blur(10px);
}
.card-content{ padding: 10px 12px 14px; }

.mobile-inner-table{
  width:100%;
  border-collapse: collapse;
  background: transparent;
}
.mobile-inner-table th,
.mobile-inner-table td{
  border: 1px solid rgba(255,255,255,.12);
  padding: 10px 10px;
  text-align: left;
  background: rgba(255,255,255,.04);
  color: rgba(255, 234, 248, 0.9);
  font-size: 13px;
}
.mobile-inner-table th{
  background: rgba(255,255,255,.08);
  color: rgba(20, 1, 20, 0.95);
  font-weight: 900;
}

/* 模态框 */
.modal-overlay{
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,.55);
  display:flex;
  justify-content:center;
  align-items:center;
  z-index: 1000;
  padding: 16px;
  backdrop-filter: blur(10px);
}
.modal-content{
  width: min(920px, 96vw);
  max-height: 88vh;
  overflow: hidden;
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(14,18,40,.92), rgba(14,18,40,.72));
  border: 1px solid rgba(255,255,255,.14);
  box-shadow: 0 30px 90px rgba(0,0,0,.65), var(--glow);
}
.modal-header{
  padding: 16px 18px;
  display:flex;
  justify-content: space-between;
  align-items:center;
  background: linear-gradient(135deg, rgba(255,79,216,.38), rgba(39,245,255,.16));
  border-bottom: 1px solid rgba(255,255,255,.12);
}
.modal-header h3{
  margin: 0;
  font-size: 16px;
  font-weight: 1000;
}
.modal-close-btn{
  width: 38px;
  height: 38px;
  border-radius: 14px;
  border: 1px solid rgba(255,255,255,.18);
  background: rgba(255,255,255,.06);
  color: rgba(234,240,255,.92);
  cursor: pointer;
  transition: transform .16s ease, box-shadow .16s ease;
}
.modal-close-btn:hover{ transform: translateY(-1px); box-shadow: 0 14px 26px rgba(0,0,0,.35); }
.modal-body{
  padding: 16px 18px;
  max-height: 70vh;
  overflow: auto;
}
.modal-body::-webkit-scrollbar{ width: 10px; }
.modal-body::-webkit-scrollbar-thumb{
  background: linear-gradient(180deg, rgba(255,79,216,.85), rgba(39,245,255,.55));
  border-radius: 10px;
}
.modal-body::-webkit-scrollbar-track{ background: rgba(255,255,255,.06); border-radius: 10px; }

/* 详情列表 */
.detail-list{ list-style: none; padding: 0; margin: 0; display:flex; flex-direction: column; gap: 10px; }
.detail-item{
  border-radius: 16px;
  background: rgba(255,255,255,.05);
  border: 1px solid rgba(255,255,255,.12);
  padding: 10px 12px;
}
.detail-line{
  display:flex;
  align-items:center;
  justify-content: center;
  gap: 10px;
  flex-wrap: wrap;
  text-align:center;
}
.detail-text{ color: rgba(234,240,255,.92); }

.add-blacklist-btn{
  background: linear-gradient(135deg, rgba(255,75,94,.85), rgba(255,79,216,.45));
  color: white;
  border: 1px solid rgba(255,75,94,.55);
  border-radius: 999px;
  padding: 8px 12px;
  font-size: 12px;
  cursor: pointer;
  box-shadow: 0 10px 22px rgba(0,0,0,.30);
  transition: transform .16s ease, box-shadow .16s ease;
}
.add-blacklist-btn:hover{ transform: translateY(-2px); box-shadow: 0 16px 30px rgba(0,0,0,.40); }

.blacklist-badge{
  background: linear-gradient(135deg, rgba(255,75,94,.95), rgba(255,79,216,.55));
  color: white;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 900;
  box-shadow: 0 10px 22px rgba(0,0,0,.30);
}

/* 黑名单 */
.blacklist-container{ max-height: 66vh; overflow:auto; }
.blacklist-list h4{
  margin: 0 0 12px 0;
  font-size: 14px;
  color: rgba(234,240,255,.90);
}
.blacklist-grid{
  display:grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 10px;
}
.blacklist-item{
  display:flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 16px;
  background: rgba(255,255,255,.05);
  border: 1px solid rgba(255,255,255,.12);
}
.remove-btn{
  background: linear-gradient(135deg, rgba(255,75,94,.92), rgba(255,75,94,.42));
  color:white;
  border: 1px solid rgba(255,75,94,.55);
  border-radius: 999px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 900;
  cursor:pointer;
  transition: transform .16s ease, box-shadow .16s ease;
  box-shadow: 0 10px 22px rgba(0,0,0,.30);
}
.remove-btn:hover{ transform: translateY(-2px); box-shadow: 0 16px 30px rgba(0,0,0,.40); }

.empty-mini{
  text-align:center;
  color: var(--muted);
  padding: 18px 12px;
  border-radius: 16px;
  border: 1px dashed rgba(255,255,255,.16);
  background: rgba(255,255,255,.04);
}
.hint{
  margin-top: 14px;
  font-size: 12px;
  color: var(--muted);
  line-height: 1.6;
  display:flex;
  gap: 8px;
  align-items:flex-start;
}
.hint-dot{ color: rgba(39,245,255,.8); }

/* 响应式 */
@media (max-width: 860px){
  .page-title{ font-size: 24px; }
  .btn{ min-width: 150px; padding: 11px 14px; }
  .checkboxes-container{ grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); }
}
@media (max-width: 768px){
  .desktop-table{ display: none; }
  .mobile-list{ display: grid; }
  .filter-section{ padding: 12px; }
  .filter-header{ padding: 14px 12px 10px; }
  .checkboxes-container{ padding: 12px; gap: 10px; }
  .blacklist-grid{ grid-template-columns: 1fr; }
}
@media (max-width: 420px){
  .btn-container{ gap: 10px; }
  .btn{ width: 100%; min-width: unset; }
  .filter-buttons .filter-btn{ width: 100%; }
  .tools-row .filter-btn{ width: 100%; }
}

/* 减少动画（系统设置） */
@media (prefers-reduced-motion: reduce){
  .bg-orb, .filter-icon.dancing, .bg-sparkles .sp{ animation: none !important; }
  .btn, .filter-btn, .checkbox-label{ transition: none !important; }
}
</style>
