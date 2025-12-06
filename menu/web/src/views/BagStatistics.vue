<template>
  <div>
    <header>
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
          <button class="btn clean-btn" @click="goBagStatisticsTrend">
  
          <span class="btn-text">转为变化图</span>
   
        </button>
      </div>
      <h1>{{ title }}</h1>
    </header>
    
    <div class="container filter-section">
      <button class="btn" @click="toggleFilter" style="margin-bottom:10px;border: 4px solid #ffcce6;">
        {{ filterCollapsed ? '展开材料筛选' : '收起材料筛选' }}
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
            :class="{ 'selected': selectedMaterials.includes(material) }"
          >
            <input 
              type="checkbox" 
              :id="'material-' + material"
              :value="material" 
              v-model="selectedMaterials"
              @change="filterTable"
              class="cute-checkbox"
            />
            <label 
              :for="'material-' + material" 
              class="checkbox-label"
            >
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
          <!-- 超过8000材料 -->
    <div>
         <button class="filter-btn ore-btn" @click="checkBag">
              <span class="btn-icon">💎</span>
              <span class="btn-text">超过8000材料</span>
              <span class="btn-wave">〜</span>
            </button>
         <button class="filter-btn ore-btn" @click="openBlackListModal" style="margin-left: 10px; background: linear-gradient(135deg, #ff6b6b, #ee5a52);">
              <span class="btn-icon">🚫</span>
              <span class="btn-text">黑名单管理</span>
              <span class="btn-wave">〜</span>
            </button>
            
    </div>
    </div>




    <div class="container">
      <!-- 桌面端表格 -->
      <table id="materialTable" class="desktop-table" v-if="!isLoading && filteredItems.length > 0">
        <thead>
          <tr>
            <th>统计日期</th>
            <th>材料</th>
            <th>数量</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="(item, index) in filteredItems" :key="index">
            <tr v-if="item.type === 'spacer'">
              <td colspan="3" style="height: 12px; background-color: #ffcce6;"></td>
            </tr>
            <tr v-else>
              <td>{{ item.date }}</td>
              <td>{{ item.materialDisplay }}</td>
              <td>{{ item.numDisplay }}</td>
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
            </div>
            <div class="card-content">
              <table class="mobile-inner-table">
                <thead>
                  <tr>
                    <th>日期</th>
                    <th>数量</th>
                  </tr>
                </thead>
                <tbody >
                  <tr v-for="item in group.items" :key="item.date">
                    <td>{{ item.date }}</td>
                    <td>{{ item.numDisplay }}</td>
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
  </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" >
       <div class="modal-content" @click.stop>
       <div class="modal-header">
          <h3>超过8000材料</h3>
          <button class="modal-close-btn" @click="closeDetailModal">✕</button>
        </div>
         <div class="modal-body">
        
            <ul>
              <li v-for="(value, key) in checkBagData" :key="key" style="text-align: center; margin-bottom: 10px; padding: 8px; background: rgba(255, 255, 255, 0.5); border-radius: 10px;border: 1px solid #ccc;">
                <div style="display: flex; align-items: center; justify-content: center; gap: 10px; flex-wrap: wrap;">
                  <span>{{ key }}：{{ value }}</span>
                  <span v-if="value > 8000 && blackList.includes(key)" class="blacklist-badge">
                    🚫 已在黑名单
                  </span>
                  <button v-if="value > 8000 && !blackList.includes(key)" class="add-blacklist-btn" @click="addToBlackList(key)">
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
              <div v-if="blackList.length === 0" class="empty-state">
                暂无黑名单材料
              </div>
              <div v-else>
                <div v-for="item in blackList" :key="item" class="blacklist-item">
                  <span class="material-name">{{ item }}</span>
                  <button class="remove-btn" @click="removeFromBlackList(item)">移除</button>
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
      allOre: ["萃凝晶", "水晶块", "星银矿石", "紫晶块", "白铁块", "铁块", "魔晶块", "石珀","虹滴晶"],
      isLoading: true,
      filterCollapsed: true, // 新增：筛选区折叠状态
      showDetailModal:false,
      checkBagData:{},
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
        // 如果API调用失败，可以显示错误信息给用户
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

    //
    async checkBag(){
      this.showDetailModal=true
      const data = await api.get('/api/checkBag');
      console.log(data)
      this.checkBagData=data
      
    },

    closeDetailModal(){
        this.showDetailModal=false
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
        await this.loadData(); // 重新加载数据
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

    // 筛选表格（由于使用了computed，这个方法可能不需要）
    filterTable() {
      // 由于使用了响应式数据和computed，筛选会自动触发
    },
    toggleFilter() {
      this.filterCollapsed = !this.filterCollapsed;
    },
  }
}
</script>

<style scoped>
:root {
  --primary-color: #ff85c2;
  --background-light: #ffeef5;
  --text-color: #ff6699;
  --border-color: #ffbcd9;
  --row-hover: rgba(255, 188, 217, 0.3);
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: "Comic Sans MS", "Segoe UI", sans-serif;
  background-color: var(--background-light);
  color: var(--text-color);
  min-height: 100vh;
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><path d="M25,50 Q50,25 75,50 Q50,75 25,50 Z" fill="%23ffbcd9" opacity="0.3"/></svg>');
}

header {
  background-color: rgba(255, 255, 255, 0.8);
  padding: 30px 0 10px;
  text-align: center;
  box-shadow: 0 0 20px var(--primary-color);
  border-radius: 0 0 30px 30px;
}

h1 {
  color: var(--primary-color);
  font-size: 2rem;
  text-shadow: 0 0 10px var(--primary-color);
  margin-top: 15px;
}

.btn-container {
  margin: 10px auto;
  display: flex;
  justify-content: center;
  gap: 15px;
  flex-wrap: wrap;
}

/* 模态框样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.28); /* 降低遮罩层暗度 */
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(5px);
}

.btn {
  position: relative;
  border: none;
  border-radius: 50px;
  padding: 12px 24px;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  margin: 5px;
  font-weight: bold;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 160px;
  justify-content: center;
}

.btn::before {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
  transition: left 0.6s ease;
}

.btn:hover::before {
  left: 100%;
}

.btn-icon {
  font-size: 1.2rem;
  transition: transform 0.3s ease;
}

.btn-text {
  position: relative;
  z-index: 2;
}

.btn-sparkle {
  font-size: 0.9rem;
  animation: sparkle-rotate 3s ease-in-out infinite;
  transition: transform 0.3s ease;
}

@keyframes sparkle-rotate {
  0%, 100% { transform: rotate(0deg) scale(1); }
  25% { transform: rotate(10deg) scale(1.1); }
  50% { transform: rotate(-5deg) scale(0.9); }
  75% { transform: rotate(8deg) scale(1.05); }
}

/* 返回主页按钮样式 */
.home-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
}

.home-btn:hover {
  background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
  box-shadow: 0 12px 35px rgba(102, 126, 234, 0.6);
  transform: translateY(-3px) scale(1.02);
}

.home-btn:hover .btn-icon {
  transform: scale(1.2) rotate(-5deg);
}

.home-btn:hover .btn-sparkle {
  transform: scale(1.3) rotate(180deg);
}

.home-btn:active {
  transform: translateY(-1px) scale(0.98);
}

/* 清理统计按钮样式 */
.clean-btn {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
  box-shadow: 0 8px 25px rgba(240, 147, 251, 0.4);
}

.clean-btn:hover {
  background: linear-gradient(135deg, #f5576c 0%, #f093fb 100%);
  box-shadow: 0 12px 35px rgba(240, 147, 251, 0.6);
  transform: translateY(-3px) scale(1.02);
}

.clean-btn:hover .btn-icon {
  transform: scale(1.2) rotate(15deg);
}

.clean-btn:hover .btn-sparkle {
  transform: scale(1.3) rotate(-180deg);
}

.clean-btn:active {
  transform: translateY(-1px) scale(0.98);
}

.container {
  max-width: 900px;
  margin: 30px auto;
  padding: 20px;
  position: relative;
}

.modal-content {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 246, 251, 0.95));
  border-radius: 20px;
  box-shadow: 0 20px 60px rgba(255, 110, 180, 0.3);
  max-width: 90%;
  max-height: 90%;
  width: 800px;
  overflow: hidden;
  border: 2px solid rgba(255, 110, 180, 0.2);
  backdrop-filter: blur(10px);
}

.modal-header {
  background: linear-gradient(135deg, #ff69b4, #ff1493);
  color: white;
  padding: 20px 25px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.3rem;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}


.modal-close-btn {
  background: rgba(255, 105, 180, 0.3);
  border: none;
  color: white;
  font-size: 1.2rem;
  cursor: pointer;
  border-radius: 50%;
  width: 35px;
  height: 35px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.modal-close-btn:hover {
  background: rgba(255, 20, 147, 0.5);
  transform: scale(1.1);
}

.modal-body {
  padding: 25px;
  max-height: 60vh;
  overflow-y: auto;
}

/* 黑名单相关样式 */
.add-blacklist-btn {
  background: linear-gradient(135deg, #ff6b6b, #ee5a52);
  color: white;
  border: none;
  border-radius: 15px;
  padding: 6px 12px;
  font-size: 0.8rem;
  cursor: pointer;
  margin-left: 10px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(255, 107, 107, 0.3);
}

.add-blacklist-btn:hover {
  background: linear-gradient(135deg, #ee5a52, #ff6b6b);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 107, 107, 0.5);
}

.blacklist-container {
  max-height: 400px;
  overflow-y: auto;
}

.blacklist-list h4 {
  color: var(--primary-color);
  margin-bottom: 15px;
  font-size: 1.1rem;
  text-shadow: 0 2px 4px rgba(255, 133, 194, 0.3);
}

.blacklist-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.8), rgba(255, 246, 251, 0.8));
  border: 2px solid rgba(255, 133, 194, 0.3);
  border-radius: 15px;
  padding: 12px 16px;
  margin-bottom: 10px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(255, 133, 194, 0.2);
}

.blacklist-item:hover {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(255, 246, 251, 0.9));
  border-color: rgba(255, 133, 194, 0.5);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 133, 194, 0.3);
}

.material-name {
  font-weight: bold;
  color: var(--text-color);
  font-size: 1rem;
}

.remove-btn {
  background: linear-gradient(135deg, #ff6b6b, #ee5a52);
  color: white;
  border: none;
  border-radius: 10px;
  padding: 6px 12px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 6px rgba(255, 107, 107, 0.3);
}

.remove-btn:hover {
  background: linear-gradient(135deg, #ee5a52, #ff6b6b);
  transform: translateY(-1px);
  box-shadow: 0 3px 8px rgba(255, 107, 107, 0.4);
}

.empty-state {
  text-align: center;
  color: #999;
  font-style: italic;
  padding: 20px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 15px;
  border: 2px dashed rgba(255, 133, 194, 0.3);
}

.blacklist-badge {
  background: linear-gradient(135deg, #ff6b6b, #ee5a52);
  color: white;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: bold;
  box-shadow: 0 2px 6px rgba(255, 107, 107, 0.3);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.05); }
  100% { transform: scale(1); }
}


/* 强化滚动条外观（更清晰） */
.modal-body::-webkit-scrollbar { width: 10px; }
.modal-body::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, #ff69b4, #ff1493);
  border-radius: 8px;
}
.modal-body::-webkit-scrollbar-track {
  background: rgba(255, 105, 180, 0.2);
  border-radius: 8px;
}


table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0 8px;
  background-color: rgba(255, 255, 255, 0.8);
  box-shadow: 0 0 15px #ffbcd9;
  border-radius: 20px;
  overflow: hidden;
  padding: 15px;
}

th, td {
  padding: 12px 16px;
  text-align: left;
  color: var(--text-color);
}

th {
  background-color: #ffbcd9;
  color: #fff;
  font-weight: bold;
  text-shadow: none;
  border-radius: 10px;
}

td {
  background-color: rgba(255, 255, 255, 0.5);
  border-radius: 10px;
}

tr:hover td {
  background-color: var(--row-hover);
}

/* 二次元装饰元素 */
.container:before {
  content: "";
  position: absolute;
  top: -20px;
  left: -20px;
  width: 60px;
  height: 60px;
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path d="M50,5 C55,25 75,30 95,50 C75,70 55,75 50,95 C45,75 25,70 5,50 C25,30 45,25 50,5 Z" fill="%23ffbcd9"/></svg>');
  background-size: contain;
  background-repeat: no-repeat;
  opacity: 0.5;
}

.container:after {
  content: "";
  position: absolute;
  bottom: -20px;
  right: -20px;
  width: 60px;
  height: 60px;
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path d="M50,5 C55,25 75,30 95,50 C75,70 55,75 50,95 C45,75 25,70 5,50 C25,30 45,25 50,5 Z" fill="%23ffbcd9"/></svg>');
  background-size: contain;
  background-repeat: no-repeat;
  opacity: 0.5;
  transform: rotate(45deg);
}

/* 筛选区域美化样式 */
.filter-section {
  margin-top: -20px;
  margin-bottom: 20px;
}

.filter-container {
  background: linear-gradient(135deg, #fff0f5 0%, #ffe8f0 50%, #fff5fa 100%);
  border: 3px solid transparent;
  background-clip: padding-box;
  border-radius: 25px;
  padding: 25px;
  box-shadow: 
    0 10px 40px rgba(255, 133, 194, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  position: relative;
  overflow: hidden;
}

.filter-container::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #ff85c2, #ff6b9d, #ff85c2);
  z-index: -1;
  margin: -3px;
  border-radius: inherit;
}

.filter-container::after {
  content: "";
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: 
    radial-gradient(circle at 25% 25%, rgba(255, 133, 194, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 75% 75%, rgba(255, 107, 157, 0.1) 0%, transparent 50%);
  animation: float-background 8s ease-in-out infinite;
  pointer-events: none;
}

@keyframes float-background {
  0%, 100% { transform: rotate(0deg) translate(0, 0); }
  25% { transform: rotate(1deg) translate(2px, -2px); }
  50% { transform: rotate(-1deg) translate(-2px, 2px); }
  75% { transform: rotate(0.5deg) translate(1px, 1px); }
}

@keyframes twinkle {
  0%, 100% { opacity: 0.3; transform: scale(1); }
  50% { opacity: 0.8; transform: scale(1.2); }
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  flex-wrap: wrap;
  gap: 10px;
}

.filter-title {
  color: var(--primary-color);
  font-size: 1.4rem;
  font-weight: bold;
  text-shadow: 0 3px 6px rgba(255, 133, 194, 0.4);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
}

.title-text {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.title-main {
  font-size: 1.4rem;
  background: linear-gradient(135deg, #ff85c2, #ff6b9d, #ff85c2);
  background-size: 200% 200%;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: gradient-shift 3s ease-in-out infinite;
}

.title-sub {
  font-size: 0.7rem;
  color: rgba(255, 133, 194, 0.7);
  font-weight: normal;
  letter-spacing: 1px;
  text-transform: uppercase;
}

@keyframes gradient-shift {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.filter-icon {
  font-size: 1.3rem;
  display: inline-block;
}

.filter-icon.dancing {
  animation: dance 3s ease-in-out infinite;
}

@keyframes dance {
  0%, 100% { transform: scale(1) rotate(0deg) translateY(0px); }
  25% { transform: scale(1.1) rotate(-5deg) translateY(-2px); }
  50% { transform: scale(1.05) rotate(5deg) translateY(-1px); }
  75% { transform: scale(1.1) rotate(-3deg) translateY(-2px); }
}

.filter-buttons {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.filter-btn {
  position: relative;
  background: linear-gradient(135deg, #fff 0%, #ffe8f0 50%, #fff 100%);
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  border-radius: 30px;
  padding: 10px 18px;
  font-size: 0.9rem;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  display: flex;
  align-items: center;
  gap: 6px;
  box-shadow: 
    0 6px 20px rgba(255, 133, 194, 0.25),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  overflow: hidden;
  min-width: 120px;
  justify-content: center;
}

.filter-btn::before {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.6), transparent);
  transition: left 0.6s ease;
}

.filter-btn:hover::before {
  left: 100%;
}

.filter-btn:hover {
  background: linear-gradient(135deg, var(--primary-color) 0%, #ff6b9d 50%, var(--primary-color) 100%);
  color: white;
  transform: translateY(-3px) scale(1.02);
  box-shadow: 
    0 8px 25px rgba(255, 133, 194, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  border-color: #ff6b9d;
}

.filter-btn:active {
  transform: translateY(-1px) scale(0.98);
}

.filter-btn .btn-icon {
  font-size: 1rem;
  transition: transform 0.3s ease;
}

.filter-btn .btn-text {
  position: relative;
  z-index: 2;
}

.filter-btn .btn-wave {
  font-size: 0.8rem;
  opacity: 0.7;
  animation: wave 2s ease-in-out infinite;
}

@keyframes wave {
  0%, 100% { transform: translateX(0px) rotate(0deg); }
  25% { transform: translateX(2px) rotate(5deg); }
  50% { transform: translateX(-1px) rotate(-3deg); }
  75% { transform: translateX(1px) rotate(2deg); }
}

/* 取消选择按钮特殊样式 */
.cancel-btn:hover .btn-icon {
  transform: rotate(180deg) scale(1.2);
}

.cancel-btn:hover .btn-wave {
  animation: wave-fast 0.5s ease-in-out infinite;
}

/* 选择矿石按钮特殊样式 */
.ore-btn:hover .btn-icon {
  transform: scale(1.3) rotate(10deg);
  text-shadow: 0 0 10px rgba(255, 255, 255, 0.8);
}

.ore-btn:hover .btn-wave {
  animation: wave-glow 1s ease-in-out infinite;
}

@keyframes wave-fast {
  0%, 100% { transform: translateX(0px) scale(1); }
  50% { transform: translateX(-3px) scale(1.1); }
}

@keyframes wave-glow {
  0%, 100% { 
    transform: translateX(0px); 
    text-shadow: 0 0 5px rgba(255, 133, 194, 0.5);
  }
  50% { 
    transform: translateX(2px); 
    text-shadow: 0 0 15px rgba(255, 133, 194, 0.8);
  }
}

.checkboxes-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 15px;
  margin-top: 20px;
  padding: 15px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.3) 0%, rgba(255, 233, 240, 0.3) 100%);
  border-radius: 20px;
  border: 1px solid rgba(255, 188, 217, 0.3);
}

.checkbox-item {
  position: relative;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.checkbox-item.selected {
  transform: scale(1.08) translateY(-2px);
  z-index: 10;
}

.checkbox-item::before {
  content: "";
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(135deg, transparent, rgba(255, 133, 194, 0.2), transparent);
  border-radius: 18px;
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: -1;
}

.checkbox-item.selected::before {
  opacity: 1;
}

.cute-checkbox {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 10px 14px;
  border-radius: 16px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9) 0%, rgba(255, 240, 245, 0.9) 100%);
  border: 2px solid transparent;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  font-size: 0.85rem;
  font-weight: 500;
  position: relative;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(255, 133, 194, 0.1);
}

.checkbox-label::before {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.5), transparent);
  transition: left 0.6s ease;
}

.checkbox-label:hover::before {
  left: 100%;
}

.checkbox-label:hover {
  background: linear-gradient(135deg, rgba(255, 188, 217, 0.3) 0%, rgba(255, 218, 236, 0.4) 100%);
  border-color: var(--border-color);
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(255, 133, 194, 0.2);
}

.checkbox-custom {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-radius: 10px;
  background: linear-gradient(135deg, #fff 0%, #ffe8f0 100%);
  position: relative;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  flex-shrink: 0;
  box-shadow: inset 0 1px 3px rgba(255, 133, 194, 0.1);
}

.checkbox-custom::before {
  content: "";
  position: absolute;
  top: 50%;
  left: 50%;
  width: 8px;
  height: 8px;
  background: linear-gradient(135deg, var(--primary-color), #ff6b9d);
  border-radius: 50%;
  transform: translate(-50%, -50%) scale(0);
  transition: transform 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.cute-checkbox:checked + .checkbox-label .checkbox-custom {
  background: linear-gradient(135deg, var(--primary-color) 0%, #ff6b9d 100%);
  border-color: var(--primary-color);
  box-shadow: 
    0 0 15px rgba(255, 133, 194, 0.6),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  transform: scale(1.1);
}

.cute-checkbox:checked + .checkbox-label .checkbox-custom::before {
  transform: translate(-50%, -50%) scale(0);
}

.cute-checkbox:checked + .checkbox-label .checkbox-custom::after {
  content: "💖";
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(1);
  font-size: 10px;
  animation: heart-bounce 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

@keyframes heart-bounce {
  0% { transform: translate(-50%, -50%) scale(0); }
  50% { transform: translate(-50%, -50%) scale(1.3); }
  100% { transform: translate(-50%, -50%) scale(1); }
}

.cute-checkbox:checked + .checkbox-label {
  background: linear-gradient(135deg, 
    rgba(255, 133, 194, 0.25) 0%, 
    rgba(255, 188, 217, 0.35) 50%, 
    rgba(255, 133, 194, 0.25) 100%);
  border-color: var(--primary-color);
  color: var(--primary-color);
  font-weight: bold;
  box-shadow: 0 4px 20px rgba(255, 133, 194, 0.3);
  transform: translateY(-1px);
}

.material-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  position: relative;
  z-index: 2;
}

/* 加载动画样式 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 20px;
  min-height: 100px;
}

.loading-animation {
  text-align: center;
}

.loading-dots {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  margin-bottom: 15px;
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary-color) 0%, #ff6b9d 100%);
  animation: bounce 1.5s ease-in-out infinite;
}

.dot:nth-child(2) {
  animation-delay: 0.3s;
}

.dot:nth-child(3) {
  animation-delay: 0.6s;
}

@keyframes bounce {
  0%, 60%, 100% {
    transform: translateY(0);
    opacity: 0.7;
  }
  30% {
    transform: translateY(-15px);
    opacity: 1;
  }
}

.loading-text {
  color: var(--primary-color);
  font-size: 0.9rem;
  font-weight: 500;
  margin: 0;
  opacity: 0.8;
}

/* 空状态样式 */
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
  padding: 40px 20px;
}

.empty-content {
  text-align: center;
  max-width: 400px;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 20px;
  opacity: 0.6;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-10px); }
}

.empty-title {
  color: var(--primary-color);
  font-size: 1.5rem;
  font-weight: bold;
  margin: 0 0 10px 0;
  text-shadow: 0 2px 4px rgba(255, 133, 194, 0.2);
}

.empty-description {
  color: var(--text-color);
  font-size: 1rem;
  margin: 0 0 20px 0;
  opacity: 0.8;
  line-height: 1.5;
}

.empty-btn {
  background: linear-gradient(135deg, var(--primary-color) 0%, #ff6b9d 100%);
  color: white;
  border: none;
  margin-top: 10px;
}

.empty-btn:hover {
  background: linear-gradient(135deg, #ff6b9d 0%, var(--primary-color) 100%);
  transform: translateY(-1px);
}

/* 移动端卡片样式 */
.mobile-list {
  display: none;
  grid-template-columns: 1fr; /* 单列布局 */
  gap: 15px;
  margin-top: 20px;
  padding: 15px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.3) 0%, rgba(255, 233, 240, 0.3) 100%);
  border-radius: 20px;
  border: 1px solid rgba(255, 188, 217, 0.3);
}

.mobile-card {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 15px;
  padding: 15px;
  box-shadow: 0 4px 15px rgba(255, 133, 194, 0.1);
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.mobile-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 25px rgba(255, 133, 194, 0.2);
}

.spacer-card {
  height: 12px;
  background-color: #ffcce6;
  border-radius: 8px;
  border: 1px solid rgba(226, 7, 102, 0.5);
}

.material-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 10px;
  border-bottom: 3px dashed rgba(255, 133, 194, 0.3);
  font-weight: bold;
}

.material-card .card-date {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-color);
  font-size: 0.9rem;
}

.material-card .date-icon {
  font-size: 1.1rem;
}

.material-card .card-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 15px;
  margin-top: 15px;
}

.material-card .material-info {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--primary-color);
  font-size: 1rem;
  font-weight: bold;
}

.material-card .material-icon {
  font-size: 1.2rem;
}

.material-card .quantity-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-color);
  font-size: 1rem;
  font-weight: 500;
}

.material-card .quantity-icon {
  font-size: 1.1rem;
}

.material-card .quantity-value {
  color: var(--primary-color);
  font-size: 1.1rem;
  font-weight: bold;
}

@media (max-width: 768px) {
  .container {
    margin: 20px auto;
    padding: 15px;
  }

  .modal-body {
    padding: 20px;
    max-height: 70vh;
  }
  
  /* 桌面端表格隐藏 */
  .desktop-table {
    display: none;
  }
  
  /* 移动端卡片显示 */
  .mobile-list {
    display: grid;
  }
  
  /* 移动端卡片优化 */
  .mobile-card {
    padding: 12px;
  }
  
  .material-card .card-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .material-card .material-info {
    font-size: 0.9rem;
  }
  
  .material-card .quantity-info {
    font-size: 0.9rem;
  }
  
  .material-card .quantity-value {
    font-size: 1rem;
  }
  
  /* 筛选区域移动端优化 */
  .filter-container {
    padding: 20px;
    margin: 0 10px 20px;
  }
  
  .filter-title {
    font-size: 1.2rem;
  }
  
  .filter-description {
    font-size: 0.9rem;
  }
  
  .material-tags {
    gap: 8px;
  }
  
  .material-tag {
    padding: 6px 12px;
    font-size: 0.8rem;
  }
  
  .filter-actions {
    flex-direction: column;
    gap: 10px;
  }
  
  .filter-btn {
    width: 100%;
    padding: 10px;
    font-size: 0.9rem;
  }
  
  /* 空状态移动端优化 */
  .empty-state {
    padding: 30px 20px;
  }
  
  .empty-icon {
    font-size: 3rem;
  }
  
  .empty-title {
    font-size: 1.4rem;
  }
  
  .empty-description {
    font-size: 0.9rem;
  }
  
  .empty-btn {
    padding: 10px 20px;
    font-size: 0.9rem;
  }
}

@media (max-width: 480px) {
  .container {
    padding: 10px;
    margin: 15px auto;
  }
  
  /* 小屏幕卡片进一步优化 */
  .mobile-card {
    padding: 10px;
  }
  
  .material-card .card-header {
    padding-bottom: 8px;
  }
  
  .material-card .card-date {
    font-size: 0.85rem;
  }
  
  .material-card .date-icon {
    font-size: 1rem;
  }
  
  .material-card .material-info {
    font-size: 0.85rem;
  }
  
  .material-card .material-icon {
    font-size: 1.1rem;
  }
  
  .material-card .quantity-info {
    font-size: 0.85rem;
  }
  
  .material-card .quantity-icon {
    font-size: 1rem;
  }
  
  .material-card .quantity-value {
    font-size: 0.9rem;
  }
  
  /* 筛选区域小屏幕优化 */
  .filter-container {
    padding: 15px;
    margin: 0 5px 15px;
  }
  
  .filter-title {
    font-size: 1.1rem;
  }
  
  .filter-description {
    font-size: 0.85rem;
  }
  
  .material-tags {
    gap: 6px;
  }
  
  .material-tag {
    padding: 5px 10px;
    font-size: 0.75rem;
  }
  
  .filter-btn {
    padding: 8px 16px;
    font-size: 0.85rem;
  }
  
  /* 空状态小屏幕优化 */
  .empty-state {
    padding: 25px 15px;
  }
  
  .empty-icon {
    font-size: 2.5rem;
  }
  
  .empty-title {
    font-size: 1.2rem;
  }
  
  .empty-description {
    font-size: 0.85rem;
  }
  
  .empty-btn {
    padding: 8px 16px;
    font-size: 0.85rem;
  }
}

/* 移动端内嵌表格格子样式 */
.mobile-inner-table {
  width: 100%;
  border-collapse: collapse;
  background: transparent;
}

.mobile-inner-table th,
.mobile-inner-table td {
  border: 1px solid #ffbcd9;
  padding: 8px 10px;
  text-align: left;
  background: #fff7fa;
  font-size: 0.95em;
}

.mobile-inner-table th {
  background: #ffbcd9;
  color: #fff;
  font-weight: bold;
}

.mobile-inner-table tr:not(:last-child) td {
  border-bottom: 1px solid #ffbcd9;
}
</style>
