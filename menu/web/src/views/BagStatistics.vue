<template>
  <div class="kawaii-page">
    <div class="bg-deco">
      <div class="floating-element heart-1">💖</div>
      <div class="floating-element star-1">✨</div>
      <div class="floating-element heart-2">💕</div>
      <div class="floating-element star-2">🌟</div>
    </div>

    <div class="main-container">
      <header class="kawaii-header">
        <div class="header-actions left">
          <button @click="goHome" class="kawaii-btn home-btn icon-btn">
            🏠 <span class="btn-text">主页</span>
          </button>
        </div>

        <div class="title-box">
          <h1 class="main-title">🌸 {{ title }} 🌸</h1>
          <span class="sub-title">✨ Chuunibyou Inventory Archive ✨</span>
        </div>

        <div class="header-actions right">
          <button @click="checkBag()" class="kawaii-btn overflow-btn icon-btn">
            🔍 <span class="btn-text">溢出检查</span>
          </button>
          
          <button @click="deleteBag" class="kawaii-btn clean-btn icon-btn">
            🧹 <span class="btn-text">清理统计</span>
          </button>
          <button @click="goBagStatisticsTrend" class="kawaii-btn trend-btn icon-btn">
            📈 <span class="btn-text">变化图</span>
          </button>
          <button @click="openEatStatisticsModal" class="kawaii-btn eat-btn icon-btn">
            💊 <span class="btn-text">吃药查看</span>
          </button>
          <button @click="goMoralePage" class="kawaii-btn morale-btn icon-btn">
            💰 <span class="btn-text">摩拉收益</span>
          </button>
        </div>
      </header>

      <section class="filter-panel-wrapper">
        <button class="filter-toggle-btn" :class="{ 'is-collapsed': filterCollapsed }" @click="toggleFilter">
          <span>{{ filterCollapsed ? '🎀' : '🪄' }} {{ filterCollapsed ? '展开材料筛选' : '收起筛选面板' }}</span>
          <span class="toggle-deco">{{ filterCollapsed ? '⟡' : '✧' }}</span>
        </button>

        <transition name="slide-down">
          <div v-show="!filterCollapsed" class="filter-content-box">
            <div class="filter-tools">
              <span class="tool-label">🧸 快速操作:</span>
              <button @click="cancelSelection" class="kawaii-btn small outline">✨ 取消选择</button>
              <button @click="selectAllOre" class="kawaii-btn small outline">💎 全选矿石</button>
            </div>
            
            <div class="material-checkbox-grid">
              <label 
                v-for="material in uniqueMaterials" 
                :key="material" 
                class="kawaii-checkbox"
                :class="{ checked: selectedMaterials.includes(material) }"
              >
                <input type="checkbox" :value="material" v-model="selectedMaterials" class="hidden-input">
                <span class="checkbox-deco">🌸</span>
                <span class="material-name">{{ material }}</span>
                <button @click.stop="deleteMaterial(material)" class="material-delete-btn" title="删除此材料">
                  ✖
                </button>
              </label>
            </div>
          </div>
        </transition>
      </section>

      <section class="action-bar">
        <div class="bar-info">
          🎁 共统计 <span class="highlight-num">{{ sortedItems.length }}</span> 条记录
        </div>
        <div class="action-bar-buttons">
          <button @click="openAddMaterialModal" class="kawaii-btn small" style="background: #E1F5FE; border-color: #0288D1; color: #01579B; box-shadow: 0 3px 0 #0288D1;">
            ➕ 新增关注材料
          </button>
          <button @click="clearAllStatistics" class="kawaii-btn small" style="background: #FFEBEE; border-color: #EF5350; color: #C62828; box-shadow: 0 3px 0 #EF5350;">
            🗑️ 清空所有
          </button>
          <button @click="openBlackListModal" class="kawaii-btn danger-btn small">
            🚫 黑名单管理
          </button>
        </div>
      </section>

      <main class="data-display-section">
        <div v-if="isLoading" class="loading-state">
          <div class="loading-spinner">🍥</div>
          <p>少女祈祷中...数据加载ing...</p>
        </div>

        <div v-else-if="filteredItems.length === 0" class="empty-state">
          <div class="empty-img">📦</div>
          <p class="empty-text">
            {{ selectedMaterials.length > 0 ? '呜呜，在这个筛选条件下没有找到数据呢~' : '背包空空如也，还没有任何统计数据哦~' }}
          </p>
          <button v-if="selectedMaterials.length > 0" @click="cancelSelection" class="kawaii-btn primary">
            ✨ 清除筛选条件
          </button>
        </div>

        <div v-else>
          <div class="desktop-table-view">
            <table class="kawaii-table">
              <thead>
                <tr>
                  <th>📅 统计日期</th>
                  <th>🎀 材料名称</th>
                  <th>🔢 数量 (变化)</th>
                </tr>
              </thead>
              <tbody>
                <tr 
                  v-for="(item, index) in filteredItems" 
                  :key="index" 
                  :class="item.type === 'spacer' ? 'spacer-row' : 'data-row'"
                >
                  <template v-if="item.type !== 'spacer'">
                    <td class="date-cell">{{ item.date }}</td>
                    <td>
                      <span class="material-pill" :class="{ special: ['原石', '摩拉数值'].includes(item.cl) }">
                        {{ item.materialDisplay }}
                      </span>
                    </td>
                    <td class="num-cell">
                      {{ item.numDisplay }}
                    </td>
                  </template>
                  <template v-else>
                    <td colspan="3" class="spacer-td">
                      <div class="spacer-line"></div>
                    </td>
                  </template>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="mobile-card-view">
            <div v-for="group in groupedMobileMaterials" :key="group.cl" class="kawaii-card">
              <div class="card-header">
                <span class="card-icon">🎁</span>
                <h3 class="card-title">{{ group.cl }}</h3>
                <span class="card-count-badge">{{ group.items.length }}条</span>
              </div>
              <div class="card-body">
                <div v-for="(item, idx) in group.items" :key="idx" class="card-list-item">
                  <span class="item-date">{{ item.date }}</span>
                  <span class="item-num">
                    {{ item.numDisplay }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <div v-if="showDetailModal" class="kawaii-modal-mask" @click.self="closeDetailModal">
      <div class="kawaii-modal">
        <div class="modal-header modal-header-purple">
          <h3>💎 溢出材料详情 (>8000)</h3>
          <button class="close-btn" @click="closeDetailModal">✖</button>
        </div>
        <div class="modal-body">
          <ul class="detail-list">
            <li v-for="(value, key) in checkBagData" :key="key" class="detail-list-item">
              <span class="detail-key">{{ key }}</span>
              <div class="detail-right">
                <span class="detail-value">{{ value }}</span>
                <span v-if="blackList.includes(key)" class="status-tag blocked">🚫 已屏蔽</span>
                <button v-else @click="addToBlackList(key)" class="kawaii-btn small outline danger-btn">加入黑名单</button>
              </div>
            </li>
          </ul>
          <div v-if="Object.keys(checkBagData).length === 0" class="empty-mini-state">
            ✨ 太棒了！没有超过8000的材料溢出哦~
          </div>
        </div>
      </div>
    </div>

    <div v-if="showBlackListModal" class="kawaii-modal-mask" @click.self="closeBlackListModal">
      <div class="kawaii-modal">
        <div class="modal-header modal-header-pink">
          <h3>🚫 黑名单管理</h3>
          <button class="close-btn" @click="closeBlackListModal">✖</button>
        </div>
        <div class="modal-body">
          <div class="modal-tip">✦ 提示：不想看见的材料，可以在上一个窗口直接“加入黑名单”哦。</div>
          
          <div v-if="blackList.length === 0" class="empty-mini-state">
            (｡•́︿•̀｡) 暂时还没有黑名单材料呢
          </div>

          <div v-else class="blacklist-tags">
            <div v-for="item in blackList" :key="item" class="blacklist-tag">
              {{ item }}
              <button class="tag-remove-btn" @click="removeFromBlackList(item)">✖</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showAddMaterialModal" class="kawaii-modal-mask" @click.self="closeAddMaterialModal">
      <div class="kawaii-modal">
        <div class="modal-header" style="background: #0288D1;">
          <h3>➕ 新增关注材料</h3>
          <button class="close-btn" @click="closeAddMaterialModal">✖</button>
        </div>
        <div class="modal-body">
          <div class="add-material-form">
            <label class="form-label">📝 材料名称：</label>
            <input 
              v-model="newMaterialName" 
              type="text" 
              class="kawaii-input" 
              placeholder="请输入材料名称"
              @keyup.enter="addMaterial"
            >
            <button @click="addMaterial" class="kawaii-btn primary" style="margin-top: 15px; width: 100%;">
              ✨ 确认添加
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showEatStatisticsModal" class="kawaii-modal-mask" @click.self="closeEatStatisticsModal">
      <div class="kawaii-modal kawaii-modal-large">
        <div class="modal-header modal-header-green">
          <h3>💊 营养袋吃药统计</h3>
          <button class="close-btn" @click="closeEatStatisticsModal">✖</button>
        </div>
        <div class="modal-body">
          <div class="date-selector">
            <label class="selector-label">📅 选择日期：</label>
            <select v-model="selectedDate" class="kawaii-select">
              <option v-for="date in availableDates" :key="date" :value="date">
                {{ date }}
              </option>
            </select>
          </div>

          <div v-if="selectedDate && eatStatisticsData[selectedDate]" class="eat-statistics-content">
            <div class="consumption-summary">
              <h4 class="summary-title">📊 {{ selectedDate }} 消耗汇总</h4>
              <div class="summary-cards">
                <div v-for="(count, name) in dailyConsumptionSummary" :key="name" class="summary-card" :class="{ negative: count > 0 }">
                  <div class="card-name">{{ name }}</div>
                  <div class="card-count" :class="{ negative: count > 0 }">
                    {{ count > 0 ? '+' : '' }}{{ count }}
                  </div>
                </div>
              </div>
            </div>

            <div class="detail-records">
              <h4 class="detail-title">📝 详细记录</h4>
              <div class="records-list">
                <div v-for="(item, idx) in getDetailRecordsWithDiff(selectedDate)" :key="idx" class="record-item">
                  <span class="record-time">{{ item.Time }}</span>
                  <span class="record-name">{{ item.Name }}</span>
                  <span class="record-count">
                    {{ item.Count }}
                    <span v-if="item.diff !== null" class="diff-badge" :class="{ positive: item.diff > 0, negative: item.diff < 0 }">
                      {{ item.diff > 0 ? '+' : '' }}{{ item.diff }}
                    </span>
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="empty-mini-state">
            ✨ 请选择日期查看吃药统计数据~
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script>
import { apiMethods } from '@/utils/api'
import api from '@/utils/api'
import { Modal, message } from 'ant-design-vue'

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
      blackList: [],
      showEatStatisticsModal: false,
      eatStatisticsData: {},
      selectedDate: '',
      showAddMaterialModal: false,
      newMaterialName: ''
    }
  },
  computed: {
    // 基础数据处理与排序
    sortedItems() {
      const processed = this.items.map(item => ({
        date: item.Data || item.date,
        cl: item.Cl || item.cl,
        num: parseInt(item.Num || item.num || 0)
      }));
      
      return processed.sort((a, b) => {
        if (a.cl === '原石' && b.cl !== '原石') return -1;
        if (a.cl !== '原石' && b.cl === '原石') return 1;
        if (a.cl === '摩拉数值' && b.cl !== '摩拉数值') return -1;
        if (a.cl !== '摩拉数值' && b.cl === '摩拉数值') return 1;
        return a.cl.localeCompare(b.cl);
      });
    },

    

    uniqueMaterials() {
      return [...new Set(this.sortedItems.map(item => item.cl))].sort();
    },

    

    filteredDataRaw() {
      return this.selectedMaterials.length === 0
        ? this.sortedItems
        : this.sortedItems.filter(item => this.selectedMaterials.includes(item.cl));
    },

    // 处理显示逻辑：插入间隔行 (Spacer) 以区分不同材料
    filteredItems() {
      const result = [];
      let lastCl = null;
      let materialMap = {}; // 用于计算变化量

      const rawData = this.filteredDataRaw;

      for (let i = 0; i < rawData.length; i++) {
        const { date, cl, num } = rawData[i];

        // 如果不是第一行，且材料名变了，插入间隔行
        if (lastCl !== null && cl !== lastCl) {
          result.push({ type: 'spacer' });
        }
        lastCl = cl;

        // 显示文本处理
        let materialDisplay = cl;
        let numDisplay = num.toString();

        if (cl === "原石") {
          const pulls = Math.floor(num / 160);
          if (pulls > 0) materialDisplay = `${cl} (${pulls}抽)`;
        }

        // 计算差值
        if (materialMap[cl] !== undefined) {
          const prev = materialMap[cl];
          const diff = num - prev.num;
          if (diff !== 0) {
            const sign = diff > 0 ? '+' : '';
            numDisplay = `${num} (${sign}${diff})`;
          }
        }
        materialMap[cl] = { date, num }; // 记录上一条数据

        result.push({
          date,
          cl,
          num,
          materialDisplay,
          numDisplay,
          type: 'data'
        });
      }

      return result;
    },

    

    // 移动端分组数据
    groupedMobileMaterials() {
      const groups = {};
      const materialMap = {}; // 用于计算变化量
      
      // 使用 raw 数据避免包含 spacer
      this.filteredDataRaw.forEach(item => {
        if (!groups[item.cl]) groups[item.cl] = [];
        
        // 重新计算移动端的显示文本
        let numDisplay = item.num.toString();
        
        // 计算差值
        if (materialMap[item.cl] !== undefined) {
          const prev = materialMap[item.cl];
          const diff = item.num - prev.num;
          if (diff !== 0) {
            const sign = diff > 0 ? '+' : '';
            numDisplay = `${item.num} (${sign}${diff})`;
          }
        }
        materialMap[item.cl] = { date: item.date, num: item.num };
        
        // 原石特殊显示（追加抽数信息）
        if (item.cl === '原石') {
          const pulls = Math.floor(item.num / 160);
          if (pulls > 0) {
            // 如果已有差值显示，则在差值后追加抽数
            if (numDisplay.includes('(') && !numDisplay.includes('抽')) {
              numDisplay = numDisplay.replace(')', ` | ${pulls}抽)`);
            } else if (!numDisplay.includes('(')) {
              numDisplay = `${item.num} (${pulls}抽)`;
            }
          }
        }
        
        groups[item.cl].push({
            ...item,
            numDisplay
        });
      });

      return Object.keys(groups).map(cl => ({
        cl,
        items: groups[cl]
      }));
    },

    // 吃药统计相关计算属性
    availableDates() {
      return Object.keys(this.eatStatisticsData).sort().reverse();
    },

    dailyConsumptionSummary() {
      return this.getDailyConsumption();
    }
  },

  async mounted() {
    await this.loadData();
    await this.loadBlackList();
  },

  methods: {
    async loadData() {
      try {
        this.isLoading = true;
        this.items = await apiMethods.getBagStatistics();
      } catch (error) {
        console.error('加载数据失败:', error);
        message.error('加载背包统计数据失败，请稍后重试');
      } finally {
        this.isLoading = false;
      }
    },

    goHome() { this.$router.push('/'); },
    goBagStatisticsTrend() { this.$router.push('/MaterialTrend'); },
    goMoralePage() { this.$router.push('/Morale'); },

    // 修改：item 变为可选参数，支持按钮直接点击
    async checkBag(item) {
      this.showDetailModal = true;
      try {
          // 这里原逻辑是获取所有 overflow 数据，不需要 item 参数也能查
          const data = await api.get('/api/checkBag');
          this.checkBagData = data;
      } catch (e) {
          console.error(e);
          message.error('获取溢出数据失败，请稍后重试');
      }
    },

    closeDetailModal() { this.showDetailModal = false; },

    async loadBlackList() {
      try {
        const response = await apiMethods.getBlackList();
        this.blackList = response.data.BlackLists || [];
      } catch (error) {
        console.error('加载黑名单失败:', error);
      }
    },

    async addToBlackList(materialName) {
      if (this.blackList.includes(materialName)) return;
      try {
        await apiMethods.addBlackList([materialName]);
        this.blackList.push(materialName);
        message.success('已添加到黑名单');
      } catch (error) {
        message.error('添加黑名单失败: ' + (error.message || error));
      }
    },

    async removeFromBlackList(materialName) {
      Modal.confirm({
        title: '确认移除',
        content: `确定要从黑名单中移除 ${materialName} 吗？`,
        okText: '确定',
        cancelText: '取消',
        onOk: async () => {
          try {
            await apiMethods.deleteBlackList(materialName);
            this.blackList = this.blackList.filter(item => item !== materialName);
            message.success('已从黑名单中移除');
          } catch (error) {
            message.error('移除黑名单失败: ' + (error.message || error));
          }
        }
      });
    },

    openBlackListModal() { this.showBlackListModal = true; },
    closeBlackListModal() { this.showBlackListModal = false; },

    async deleteBag() {
      Modal.confirm({
        title: '确认清理',
        content: '确定要清理统计数据吗？只保留最近一天。',
        okText: '确定',
        cancelText: '取消',
        okType: 'danger',
        onOk: async () => {
          try {
            const data = await api.post('/api/deleteBag');
            message.success(data.message || '操作成功！');
            await this.loadData();
          } catch (error) {
            message.error("请求出错：" + (error.message || error));
          }
        }
      });
    },

    cancelSelection() { this.selectedMaterials = []; },
    selectAllOre() { this.selectedMaterials = [...this.allOre]; },
    toggleFilter() { this.filterCollapsed = !this.filterCollapsed; },

    async openEatStatisticsModal() {
      this.showEatStatisticsModal = true;
      await this.loadEatStatistics();
    },

    async loadEatStatistics() {
      try {
        const data = await api.get('/api/EatStatistics');
        this.eatStatisticsData = data;
        // 默认选择最新日期
        const dates = Object.keys(data).sort().reverse();
        if (dates.length > 0) {
          this.selectedDate = dates[0];
        }
      } catch (error) {
        console.error('加载吃药统计失败:', error);
        message.error('加载吃药统计数据失败，请稍后重试');
      }
    },

    closeEatStatisticsModal() {
      this.showEatStatisticsModal = false;
      this.selectedDate = '';
    },

    // 计算选中日期的消耗统计（通过差值计算真实消耗）
    getDailyConsumption() {
      if (!this.selectedDate || !this.eatStatisticsData[this.selectedDate]) {
        return {};
      }
      
      const records = [...this.eatStatisticsData[this.selectedDate]];
      // 按时间排序（从旧到新）
      records.sort((a, b) => {
        const timeA = a.Time.replace('时间:', '');
        const timeB = b.Time.replace('时间:', '');
        return new Date(timeA) - new Date(timeB);
      });

      // 按物品名称分组
      const groupedByName = {};
      records.forEach(item => {
        if (!groupedByName[item.Name]) {
          groupedByName[item.Name] = [];
        }
        groupedByName[item.Name].push(item);
      });

      // 计算每种物品的总消耗（累加所有差值）
      const consumption = {};
      Object.keys(groupedByName).forEach(name => {
        const group = groupedByName[name];
        let totalConsumption = 0;
        let previousCount = null;
        
        group.forEach(item => {
          if (previousCount !== null) {
            // 差值 = 当前数量 - 上一次数量
            const diff = item.Count - previousCount;
            totalConsumption += diff;
          }
          previousCount = item.Count;
        });
        
        consumption[name] = totalConsumption;
      });
      
      return consumption;
    },

    // 获取带差值的详细记录（按物品名称分组，每组内按时间排序）
    getDetailRecordsWithDiff(date) {
      if (!date || !this.eatStatisticsData[date]) {
        return [];
      }

      const records = [...this.eatStatisticsData[date]];
      // 先按时间排序（从旧到新）
      records.sort((a, b) => {
        const timeA = a.Time.replace('时间:', '');
        const timeB = b.Time.replace('时间:', '');
        return new Date(timeA) - new Date(timeB);
      });

      // 按物品名称分组
      const groupedByName = {};
      records.forEach(item => {
        if (!groupedByName[item.Name]) {
          groupedByName[item.Name] = [];
        }
        groupedByName[item.Name].push(item);
      });

      // 为每组计算差值，并合并所有组
      const result = [];
      Object.keys(groupedByName).sort().forEach(name => {
        const group = groupedByName[name];
        let previousCount = null;
        
        group.forEach(item => {
          let diff = null;
          
          if (previousCount !== null) {
            // 计算变化量：当前数量 - 上一次数量
            diff = item.Count - previousCount;
          }
          
          previousCount = item.Count;
          
          result.push({
            ...item,
            diff
          });
        });
      });
      
      return result;
    },

    async deleteMaterial(materialName) {
      Modal.confirm({
        title: '确认删除',
        content: `确定要删除材料 "${materialName}" 的所有统计记录吗？此操作无法撤销！`,
        okText: '确定删除',
        cancelText: '取消',
        okType: 'danger',
        onOk: async () => {
          try {
            await api.delete(`/api/BagStatistics/DELETE?name=${encodeURIComponent(materialName)}`);
            message.success('材料删除成功！');
            await this.loadData();
          } catch (error) {
            console.error('删除材料失败:', error);
            message.error('删除材料失败: ' + (error.message || error));
          }
        }
      });
    },

    openAddMaterialModal() {
      this.showAddMaterialModal = true;
      this.newMaterialName = '';
    },

    closeAddMaterialModal() {
      this.showAddMaterialModal = false;
      this.newMaterialName = '';
    },

    async addMaterial() {
      if (!this.newMaterialName.trim()) {
        message.warning('请输入材料名称');
        return;
      }
      
      try {
        await api.post(`/api/BagStatistics/ADD?name=${encodeURIComponent(this.newMaterialName.trim())}`);
        message.success('材料添加成功！');
        this.closeAddMaterialModal();
        await this.loadData();
      } catch (error) {
        console.error('添加材料失败:', error);
        message.error('添加材料失败: ' + (error.message || error));
      }
    },

    async clearAllStatistics() {
      Modal.confirm({
        title: '⚠️ 危险操作',
        content: '确定要清空所有背包统计数据吗？此操作将删除所有材料的统计记录，且无法撤销！',
        okText: '确定清空',
        cancelText: '取消',
        okType: 'danger',
        onOk: async () => {
          try {
            await api.post('/api/BagStatistics/CLEAR');
            message.success('所有统计数据已清空！');
            await this.loadData();
          } catch (error) {
            console.error('清空数据失败:', error);
            message.error('清空数据失败: ' + (error.message || error));
          }
        }
      });
    }
  }
}
</script>

<style scoped>
/* ==================================
   二次元粉色 Kawaii 主题变量
   ================================== */
.kawaii-page {
  /* 核心配色 */
  --k-pink-light: #FFF0F5; /* 浅粉背景 */
  --k-pink-main: #FFB6C1;  /* 主粉色 */
  --k-pink-dark: #FF69B4;  /* 深粉色强调 */
  --k-blue-light: #E0FFFF; /* 浅蓝点缀 */
  --k-blue-main: #87CEFA;  /* 主蓝色 */
  --k-purple-light: #E1BEE7; /* 浅紫 */
  --k-purple-main: #CE93D8;  /* 主紫 */
  --k-yellow: #FFFACD;     /* 柠檬黄 */
  --k-text-dark: #5F4B58;  /* 深褐文字，比纯黑柔和 */
  --k-text-light: #8E7D88; /* 浅褐文字 */
  --k-white: #FFFFFF;

  /* 样式变量 */
  --k-radius: 24px;
  --k-radius-sm: 12px;
  --k-shadow: 0 8px 24px rgba(255, 182, 193, 0.3), 0 2px 6px rgba(255, 182, 193, 0.1);
  --k-shadow-hover: 0 12px 32px rgba(255, 182, 193, 0.5), 0 4px 12px rgba(255, 182, 193, 0.2);
  --k-border: 2px solid var(--k-pink-main);

  position: relative;
  min-height: 100vh;
  background: linear-gradient(135deg, var(--k-pink-light) 0%, #fff 100%);
  color: var(--k-text-dark);
  font-family: "Nunito", "PingFang SC", "Microsoft YaHei", cursive, sans-serif;
  overflow-x: hidden;
}

/* --- 背景装饰 --- */
.bg-deco {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}

.floating-element {
  position: absolute;
  font-size: 2rem;
  opacity: 0.6;
  filter: blur(2px);
  animation: floatAround 20s infinite linear;
}

.heart-1 { top: 10%; left: 5%; animation-delay: 0s; font-size: 3rem; color: var(--k-pink-dark); }
.star-1 { top: 20%; right: 10%; animation-delay: -5s; color: var(--k-blue-main); }
.heart-2 { bottom: 15%; right: 20%; animation-delay: -10s; color: var(--k-pink-main); }
.star-2 { bottom: 10%; left: 15%; animation-delay: -15s; color:gold; }

@keyframes floatAround {
  0% { transform: translate(0, 0) rotate(0deg); }
  25% { transform: translate(50px, 50px) rotate(90deg); }
  50% { transform: translate(0, 100px) rotate(180deg); }
  75% { transform: translate(-50px, 50px) rotate(270deg); }
  100% { transform: translate(0, 0) rotate(360deg); }
}

/* --- 主容器 --- */
.main-container {
  position: relative;
  z-index: 1;
  max-width: 1100px;
  margin: 0 auto;
  padding: 20px;
}

/* --- 通用按钮样式 --- */
.kawaii-btn {
  padding: 10px 20px;
  border-radius: 50px; /* 超圆角 */
  border: 2px solid var(--k-pink-main);
  background: var(--k-white);
  color: var(--k-text-dark);
  font-weight: bold;
  cursor: pointer;
  box-shadow: 0 4px 0 var(--k-pink-main); /* 立体感 */
  transition: all 0.2s cubic-bezier(0.68, -0.55, 0.27, 1.55); /* Q弹动画 */
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 1rem;
}
.kawaii-btn:hover {
  transform: translateY(2px);
  box-shadow: 0 2px 0 var(--k-pink-main);
}
.kawaii-btn:active {
  transform: translateY(4px);
  box-shadow: none;
}
.kawaii-btn.primary {
  background: var(--k-pink-main);
  color: var(--k-white);
  border-color: var(--k-pink-dark);
  box-shadow: 0 4px 0 var(--k-pink-dark);
}
.kawaii-btn.primary:hover { box-shadow: 0 2px 0 var(--k-pink-dark); }

.kawaii-btn.danger-btn {
  background: var(--k-white);
  color: #FF4B5E;
  border-color: #FF4B5E;
  box-shadow: 0 4px 0 #FF4B5E;
}
.kawaii-btn.danger-btn:hover { box-shadow: 0 2px 0 #FF4B5E; background: #FFF0F0; }

.kawaii-btn.small {
  padding: 6px 14px;
  font-size: 0.9rem;
  border-width: 1.5px;
  box-shadow: 0 3px 0 currentColor;
}
.kawaii-btn.small:hover { box-shadow: 0 1px 0 currentColor; }
.kawaii-btn.outline { border-style: dashed; }

/* --- Header --- */
.kawaii-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 25px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  border-radius: var(--k-radius);
  border: 3px dotted var(--k-pink-main);
  margin-bottom: 30px;
  box-shadow: var(--k-shadow);
  flex-wrap: wrap;
  gap: 15px;
}
.title-box { text-align: center; flex: 1; min-width: 250px; }
.main-title {
  margin: 0;
  font-size: 1.8rem;
  color: var(--k-pink-dark);
  text-shadow: 2px 2px 0 var(--k-pink-light);
}
.sub-title {
  display: block;
  font-size: 0.85rem;
  color: var(--k-text-light);
  letter-spacing: 1px;
  margin-top: 5px;
}
.header-actions { display: flex; gap: 10px; }
.icon-btn .btn-text { display: none; }
@media (min-width: 768px) { .icon-btn .btn-text { display: inline; } }
.home-btn { border-color: var(--k-blue-main); color: var(--k-blue-main); box-shadow: 0 4px 0 var(--k-blue-main); }
.clean-btn { background: var(--k-yellow); border-color: orange; color: #d97706; box-shadow: 0 4px 0 orange; }
.trend-btn { background: var(--k-blue-light); border-color: var(--k-blue-main); color: var(--k-blue-main); box-shadow: 0 4px 0 var(--k-blue-main); }
/* 溢出检查按钮样式 */
.overflow-btn { background: var(--k-purple-light); border-color: var(--k-purple-main); color: #8E24AA; box-shadow: 0 4px 0 var(--k-purple-main); }
/* 吃药查看按钮样式 */
.eat-btn { background: #C8E6C9; border-color: #66BB6A; color: #2E7D32; box-shadow: 0 4px 0 #66BB6A; }
/* 摩拉收益按钮样式 */
.morale-btn { background: #FFF9C4; border-color: #FBC02D; color: #F57F17; box-shadow: 0 4px 0 #FBC02D; }

/* --- 筛选面板 --- */
.filter-panel-wrapper {
  background: var(--k-white);
  border-radius: var(--k-radius);
  border: var(--k-border);
  overflow: hidden;
  box-shadow: var(--k-shadow);
  margin-bottom: 25px;
}
.filter-toggle-btn {
  width: 100%;
  padding: 15px 20px;
  background: var(--k-pink-light);
  border: none;
  border-bottom: 3px dotted var(--k-pink-main);
  color: var(--k-pink-dark);
  font-size: 1.1rem;
  font-weight: bold;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: background 0.3s;
}
.filter-toggle-btn:hover { background: #FFE4E1; }
.filter-toggle-btn.is-collapsed { border-bottom: none; background: var(--k-white); }
.toggle-icon { font-size: 1.4rem; margin-right: 10px; }
.toggle-deco { font-size: 1.2rem; }

.filter-content-box { padding: 20px; }
.filter-tools { display: flex; align-items: center; gap: 10px; margin-bottom: 20px; flex-wrap: wrap; }
.tool-label { font-weight: bold; color: var(--k-pink-dark); }

.material-checkbox-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 12px;
}
.kawaii-checkbox {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: var(--k-pink-light);
  border: 2px solid transparent;
  border-radius: var(--k-radius-sm);
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  overflow: hidden;
}
.kawaii-checkbox:hover {
  border-color: var(--k-pink-main);
  transform: scale(1.05);
}
.kawaii-checkbox.checked {
  background: var(--k-pink-main);
  color: var(--k-white);
  border-color: var(--k-pink-dark);
  box-shadow: 0 4px 10px rgba(255, 105, 180, 0.3);
}
.hidden-input { display: none; }
.checkbox-deco { margin-right: 8px; font-size: 1.1rem; opacity: 0.5; transition: 0.2s; }
.kawaii-checkbox.checked .checkbox-deco { opacity: 1; transform: rotate(20deg); }
.material-name { font-weight: bold; font-size: 0.95rem; flex: 1; }
.material-delete-btn {
  background: transparent;
  border: none;
  color: #FF4B5E;
  font-size: 0.9rem;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 50%;
  opacity: 0;
  transition: all 0.2s;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  min-height: 20px;
}
.kawaii-checkbox:hover .material-delete-btn {
  opacity: 1;
}
.material-delete-btn:hover {
  background: rgba(255, 75, 94, 0.15);
  transform: scale(1.15);
}
.material-delete-btn:active {
  transform: scale(0.95);
}

/* --- 功能入口栏 --- */
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: var(--k-blue-light);
  border-radius: 50px;
  border: 2px dashed var(--k-blue-main);
  margin-bottom: 25px;
  flex-wrap: wrap;
  gap: 10px;
}
.action-bar-buttons {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.bar-info { font-weight: bold; color: var(--k-blue-main); display: flex; align-items: center; gap: 5px; }
.highlight-num { 
  background: var(--k-white); color: var(--k-pink-dark); padding: 2px 10px; border-radius: 20px; font-size: 1.1rem; 
  box-shadow: 0 2px 5px rgba(0,0,0,0.1);
}

/* --- 数据展示区 --- */
.data-display-section {
  min-height: 400px;
  position: relative;
  background: rgba(255,255,255,0.6);
  border-radius: var(--k-radius);
  padding: 20px;
  border: var(--k-border);
  box-shadow: var(--k-shadow);
}

/* 加载和空状态 */
.loading-state, .empty-state {
  text-align: center; 
  padding: 60px 0; 
  color: var(--k-text-light);
}
.loading-spinner { font-size: 4rem; animation: spin 2s linear infinite; display: inline-block; margin-bottom: 20px; }
@keyframes spin { to { transform: rotate(360deg); } }
.empty-img { font-size: 5rem; margin-bottom: 20px; opacity: 0.7; }
.empty-text { margin-bottom: 25px; font-size: 1.1rem; }

/* PC 表格视图 */
.desktop-table-view { 
  display: block; 
  overflow-x: auto; 
}
.mobile-card-view { display: none; }

.kawaii-table {
  width: 100%;
  border-collapse: collapse; /* 改为 collapse 以便控制内部边框 */
  border-spacing: 0;
}
.kawaii-table th {
  text-align: left;
  padding: 15px 20px;
  color: var(--k-pink-dark);
  font-weight: bold;
  background: var(--k-pink-light);
  border-bottom: 3px dotted var(--k-pink-main);
}
.kawaii-table th:first-child { border-top-left-radius: var(--k-radius-sm); }
.kawaii-table th:last-child { border-top-right-radius: var(--k-radius-sm); }

/* 数据行样式 */
.data-row td {
  padding: 12px 20px;
  background: var(--k-white);
  /* 移除边框，实现相同材料一体化 */
  border: none; 
}

.data-row:hover td {
  background: #FFFBFD;
}

.date-cell { color: var(--k-text-light); font-weight: bold; }
.material-pill {
  display: inline-block;
  padding: 6px 14px;
  background: var(--k-blue-light);
  color: var(--k-blue-main);
  border-radius: 20px;
  font-weight: bold;
  border: 2px solid transparent;
  box-shadow: 0 2px 5px rgba(135, 206, 250, 0.3);
}
.material-pill.special {
  background: var(--k-pink-main); color: var(--k-white);
  border-color: var(--k-pink-dark);
  box-shadow: 0 2px 5px rgba(255, 105, 180, 0.4);
}
.num-cell { 
    font-family: "Comic Sans MS", cursive, sans-serif; 
    font-size: 1.1rem; font-weight: bold; 
    color: var(--k-pink-dark); 
    display: flex; align-items: center; gap: 10px;
}
/* Removed alert-badge-btn styles as they are no longer used */

/* 间隔行样式 (横杠) */
.spacer-row td {
  padding: 0;
  height: 20px;
  background: transparent !important;
  vertical-align: middle;
}
.spacer-line {
  height: 3px;
  background-image: repeating-linear-gradient(to right, var(--k-pink-main) 0, var(--k-pink-main) 10px, transparent 10px, transparent 15px);
  opacity: 0.3;
  border-radius: 10px;
  margin: 0 10px;
}

/* 移动端卡片视图 */
.kawaii-card {
  background: var(--k-white);
  border-radius: var(--k-radius);
  border: var(--k-border);
  box-shadow: var(--k-shadow);
  margin-bottom: 20px;
  overflow: hidden;
}
.card-header {
  background: var(--k-pink-light);
  padding: 12px 15px;
  display: flex; align-items: center; gap: 10px;
  border-bottom: 3px dotted var(--k-pink-main);
}
.card-icon { font-size: 1.4rem; }
.card-title { margin: 0; flex: 1; color: var(--k-pink-dark); }
.card-count-badge { background: var(--k-pink-main); color: white; padding: 4px 10px; border-radius: 20px; font-size: 0.85rem; }
.card-body { padding: 5px 0; }
.card-list-item {
  display: flex; justify-content: space-between; padding: 12px 15px;
  border-bottom: 1px dashed var(--k-pink-light);
}
.card-list-item:last-child { border-bottom: none; }
.item-date { color: var(--k-text-light); }
.item-num { font-weight: bold; color: var(--k-pink-dark); display: flex; align-items: center; }
/* Removed alert-dot styles */

/* --- 模态框 --- */
.kawaii-modal-mask {
  position: fixed; inset: 0; background: rgba(255, 240, 245, 0.8); /* 半透明粉色遮罩 */
  backdrop-filter: blur(5px); z-index: 999;
  display: flex; align-items: center; justify-content: center; padding: 20px;
}
.kawaii-modal {
  width: 100%; max-width: 480px; background: var(--k-white);
  border-radius: var(--k-radius); border: 4px solid var(--k-pink-main);
  box-shadow: var(--k-shadow-hover); overflow: hidden;
  animation: popIn 0.3s cubic-bezier(0.68, -0.55, 0.27, 1.55);
}
.kawaii-modal-large { max-width: 700px; }
.modal-header {
  padding: 15px 20px; display: flex; justify-content: space-between; align-items: center;
  color: white;
}
.modal-header-pink { background: var(--k-pink-main); }
.modal-header-purple { background: #CE93D8; }
.modal-header-green { background: #66BB6A; }
.modal-header h3 { margin: 0; font-size: 1.2rem; }
.close-btn {
  background: rgba(255,255,255,0.3); border: none; width: 32px; height: 32px; border-radius: 50%;
  color: white; font-weight: bold; cursor: pointer; transition: 0.2s;
}
.close-btn:hover { background: rgba(255,255,255,0.5); transform: rotate(90deg); }
.modal-body { padding: 25px; max-height: 70vh; overflow-y: auto; }

/* 隐藏滚动条但保持滚动功能 */
.modal-body::-webkit-scrollbar { width: 0; height: 0; }
.modal-body { scrollbar-width: none; -ms-overflow-style: none; }

/* 详情列表 */
.detail-list { list-style: none; padding: 0; margin: 0; }
.detail-list-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 0; border-bottom: 2px dashed var(--k-pink-light);
}
.detail-key { font-weight: bold; color: var(--k-text-dark); }
.detail-right { display: flex; align-items: center; gap: 10px; }
.detail-value { font-family: "Comic Sans MS", cursive, sans-serif; font-weight: bold; color: var(--k-pink-dark); font-size: 1.1rem; }
.status-tag.blocked { background: #ffcdd2; color: #c62828; padding: 4px 10px; border-radius: 20px; font-size: 0.85rem; }

/* 黑名单 Tags */
.modal-tip { background: var(--k-yellow); padding: 10px; border-radius: var(--k-radius-sm); border: 2px dashed orange; color: #d97706; font-size: 0.9rem; margin-bottom: 20px;}
.empty-mini-state { text-align: center; color: var(--k-text-light); padding: 20px; border: 2px dashed var(--k-pink-light); border-radius: var(--k-radius-sm); }
.blacklist-tags { display: flex; flex-wrap: wrap; gap: 10px; }
.blacklist-tag { 
  background: var(--k-pink-light); border: 2px solid var(--k-pink-main);
  padding: 8px 15px; border-radius: 30px; display: flex; align-items: center; gap: 8px;
  font-weight: bold; color: var(--k-text-dark); box-shadow: 0 3px 0 var(--k-pink-main);
}
.tag-remove-btn { 
  background: transparent; color: #FF4B5E; border: none; font-weight: bold; cursor: pointer;
}

/* 吃药统计模态框样式 */
.date-selector {
  display: flex; align-items: center; gap: 15px; margin-bottom: 25px;
  padding: 15px; background: var(--k-pink-light); border-radius: var(--k-radius-sm);
  flex-wrap: wrap;
}
.selector-label { font-weight: bold; color: var(--k-pink-dark); white-space: nowrap; }
.kawaii-select {
  padding: 8px 15px; border-radius: 20px; border: 2px solid var(--k-pink-main);
  background: var(--k-white); color: var(--k-text-dark); font-weight: bold;
  cursor: pointer; transition: all 0.2s; flex: 1; min-width: 150px;
}
.kawaii-select:hover { border-color: var(--k-pink-dark); box-shadow: 0 2px 8px rgba(255, 105, 180, 0.2); }

.add-material-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.form-label {
  font-weight: bold;
  color: var(--k-text-dark);
  font-size: 1rem;
}
.kawaii-input {
  padding: 10px 15px;
  border-radius: 20px;
  border: 2px solid var(--k-pink-main);
  background: var(--k-white);
  color: var(--k-text-dark);
  font-weight: bold;
  font-size: 1rem;
  transition: all 0.2s;
}
.kawaii-input:focus {
  outline: none;
  border-color: var(--k-pink-dark);
  box-shadow: 0 0 0 3px rgba(255, 105, 180, 0.1);
}
.kawaii-input::placeholder {
  color: var(--k-text-light);
  font-weight: normal;
}

.eat-statistics-content { margin-top: 20px; }

.consumption-summary { margin-bottom: 30px; }
.summary-title {
  margin: 0 0 15px 0; font-size: 1.1rem; color: var(--k-pink-dark);
  padding-bottom: 10px; border-bottom: 3px dotted var(--k-pink-main);
}
.summary-cards {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 15px;
}
.summary-card {
  background: linear-gradient(135deg, #C8E6C9 0%, #A5D6A7 100%);
  border: 2px solid #66BB6A; border-radius: var(--k-radius-sm);
  padding: 15px; text-align: center; box-shadow: 0 4px 10px rgba(102, 187, 106, 0.3);
  transition: transform 0.2s;
}
.summary-card:hover { transform: translateY(-3px); }
.summary-card.negative { background: linear-gradient(135deg, #FFE0B2 0%, #FFCC80 100%); border-color: #FF9800; }
.card-name { font-size: 0.9rem; color: #2E7D32; margin-bottom: 8px; }
.card-count {
  font-size: 1.8rem; font-weight: bold; color: #1B5E20;
  font-family: "Comic Sans MS", cursive, sans-serif;
}
.card-count.negative { color: #E65100; }

.detail-records { }
.detail-title {
  margin: 0 0 15px 0; font-size: 1.1rem; color: var(--k-pink-dark);
  padding-bottom: 10px; border-bottom: 3px dotted var(--k-pink-main);
}
.records-list { max-height: 300px; overflow-y: auto; }

/* 隐藏记录列表滚动条 */
.records-list::-webkit-scrollbar { width: 0; height: 0; }
.records-list { scrollbar-width: none; -ms-overflow-style: none; }
.record-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 15px; background: var(--k-pink-light); border-radius: var(--k-radius-sm);
  margin-bottom: 10px; border: 2px solid transparent; transition: all 0.2s;
  flex-wrap: wrap; gap: 8px;
}
.record-item:hover { border-color: var(--k-pink-main); background: #FFE4E1; }
.record-time { font-size: 0.85rem; color: var(--k-text-light); flex: 0 0 180px; }
.record-name { flex: 0 0 150px; font-weight: bold; color: var(--k-text-dark); padding: 0 10px; }
.record-count {
  font-weight: bold; color: #66BB6A; font-size: 1.1rem;
  background: var(--k-white); padding: 4px 12px; border-radius: 20px;
  display: flex; align-items: center; gap: 6px; flex-shrink: 0;
}
.diff-badge {
  font-size: 0.9rem; padding: 2px 8px; border-radius: 12px;
  font-weight: bold;
}
.diff-badge.positive {
  background: #FFE0B2; color: #E65100; /* 橙色表示增加（消耗） */
}
.diff-badge.negative {
  background: #C8E6C9; color: #2E7D32; /* 绿色表示减少（补充） */
}

/* --- 响应式适配 --- */
@media (max-width: 768px) {
  .kawaii-header { flex-direction: column; text-align: center; padding: 15px; }
  .header-actions, .title-box { width: 100%; }
  .header-actions { 
    flex-wrap: wrap; 
    justify-content: center;
  }
  .header-actions.left { order: 1; }
  .header-actions.right { order: 2; }
  .title-box { order: -1; margin-bottom: 10px; }
  
  .kawaii-btn.icon-btn {
    padding: 10px 14px;
    min-width: 44px;
    min-height: 44px;
    box-sizing: border-box;
  }

  .desktop-table-view { display: none; }
  .mobile-card-view { display: block; }

  .material-checkbox-grid { grid-template-columns: repeat(2, 1fr); }
  .main-container { padding: 10px; }
  .data-display-section { padding: 10px; background: transparent; box-shadow: none; border: none; }

  /* 吃药统计移动端适配 */
  .kawaii-modal-large { max-width: 100%; margin: 10px; }
  .modal-body { padding: 15px; }
  
  .date-selector { 
    flex-direction: column; align-items: stretch; gap: 10px; padding: 12px;
  }
  .kawaii-select { width: 100%; min-width: auto; }
  
  .summary-cards { 
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr)); 
    gap: 10px;
  }
  .summary-card { padding: 12px; }
  .card-count { font-size: 1.5rem; }
  
  .summary-title, .detail-title { font-size: 1rem; }
  
  .record-item { 
    flex-direction: column; align-items: flex-start; padding: 10px; gap: 6px;
  }
  .record-time { 
    flex: none; font-size: 0.75rem; width: 100%;
    padding-bottom: 4px; border-bottom: 1px dashed var(--k-pink-main);
  }
  .record-name { 
    flex: none; font-size: 0.95rem; padding: 0; width: 100%;
  }
  .record-count { 
    font-size: 1rem; align-self: flex-end;
  }
  .diff-badge { font-size: 0.85rem; }
}

/* 动画 */
@keyframes pulse { 0% { transform: scale(1); } 50% { transform: scale(1.15); } 100% { transform: scale(1); } }
@keyframes popIn { from { opacity: 0; transform: scale(0.8); } to { opacity: 1; transform: scale(1); } }
.slide-down-enter-active, .slide-down-leave-active { transition: all 0.3s ease-in-out; max-height: 500px; opacity: 1; }
.slide-down-enter, .slide-down-leave-to { max-height: 0; opacity: 0; padding-top: 0; padding-bottom: 0; overflow: hidden; }
</style>