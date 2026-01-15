<template>
  <div class="kawaii-page">
    <div class="bg-deco">
      <div class="floating-element coin-1">💰</div>
      <div class="floating-element star-1">✨</div>
      <div class="floating-element coin-2">💵</div>
      <div class="floating-element star-2">🌟</div>
      <div class="floating-element sparkle-1">🎀</div>
      <div class="floating-element sparkle-2">💖</div>
      <div class="floating-element sparkle-3">🌸</div>
      <div class="floating-element sparkle-4">⭐</div>
    </div>

    <div class="main-container">
      <header class="kawaii-header">
        <div class="header-top">
          <div class="header-actions left" >
            <button @click="goBack" class="kawaii-btn home-btn icon-btn" style="margin-right: 30px;">
              ← <span class="btn-text">返回</span>
            </button>
          </div>

          <div class="title-box">
            <h1 class="main-title">💰💖 摩拉收益统计 💖💰</h1>
            <span class="sub-title">✨ kawaii 收益分析 ✨</span>
          </div>

          <div class="header-actions right">
            <button @click="updateMoraleRecord" class="kawaii-btn update-btn icon-btn" :disabled="isUpdating">
              <span v-if="!isUpdating">💫 <span class="btn-text">更新记录</span></span>
              <span v-else>⏳ <span class="btn-text">更新中...</span></span>
            </button>
            <button @click="resetFilters" class="kawaii-btn reset-btn icon-btn">
              🔄 <span class="btn-text">重置筛选</span>
            </button>
            <button @click="exportToExcel" class="kawaii-btn export-btn icon-btn" :disabled="!resultData || !resultData.items || resultData.items.length === 0">
              📊 <span class="btn-text">导出Excel</span>
            </button>
          </div>
        </div>
      </header>

      <section class="filter-section">
        <div class="filter-card">
          <h3 class="filter-title">🔍 筛选条件</h3>
          
          <div class="filter-grid">
            <div class="filter-item">
              <label class="filter-label">📅 统计周期</label>
              <select v-model="filters.type" class="kawaii-select" @change="onTypeChange">
                <option value="day">按天统计</option>
                <option value="month">按月统计</option>
                <option value="year">按年统计</option>
              </select>
            </div>

            <div class="filter-item">
              <label class="filter-label">📅 查询日期</label>
              <a-date-picker 
                v-if="filters.type === 'day'"
                v-model:value="dateValue"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                placeholder="选择日期"
                class="kawaii-date-picker"
                :locale="locale"
                @change="onDateChange"
                :get-popup-container="trigger => trigger.parentElement"
              />
              <a-month-picker 
                v-else-if="filters.type === 'month'"
                v-model:value="dateValue"
                format="YYYY-MM"
                value-format="YYYY-MM"
                placeholder="选择月份"
                class="kawaii-date-picker"
                :locale="locale"
                @change="onDateChange"
                :get-popup-container="trigger => trigger.parentElement"
              />
              <a-date-picker 
                v-else
                v-model:value="dateValue"
                picker="year"
                format="YYYY"
                value-format="YYYY"
                placeholder="选择年份"
                class="kawaii-date-picker"
                :locale="locale"
                @change="onDateChange"
                :get-popup-container="trigger => trigger.parentElement"
              />
            </div>

            <div class="filter-item">
              <label class="filter-label">📊 收支类型</label>
              <select v-model="filters.action" class="kawaii-select">
                <option value="">全部</option>
                <option value="击杀怪物奖励">击杀怪物奖励</option>
                <option value="其他">其他</option>
                <option value="地城通关奖励">地城通关奖励</option>
                <option value="探索派遣奖励">探索派遣奖励</option>
                <option value="每日委托奖励">每日委托奖励</option>
                <option value="每日委托完成奖励">每日委托完成奖励</option>
                <option value="宝箱奖励">宝箱奖励</option>
                <option value="地脉之花奖励">地脉之花奖励</option>
                <option value="活动奖励">活动奖励</option>
                <option value="击杀BOSS奖励">击杀BOSS奖励</option>
              </select>
            </div>
          </div>

          <div class="filter-actions">
            <button @click="searchRecords" class="kawaii-btn primary">
              🔍 查询
            </button>
          </div>
        </div>
      </section>

      <main class="data-section">
        <div v-if="isLoading" class="loading-state">
          <div class="loading-spinner">🍥</div>
          <p>少女祈祷中...数据加载ing...</p>
        </div>

        <div v-else-if="!resultData || !resultData.items || resultData.items.length === 0" class="empty-state">
          <div class="empty-img">📦</div>
          <p class="empty-text">暂无摩拉记录数据哦~</p>
        </div>

        <div v-else>
          <div class="summary-card">
            <div class="summary-item">
              <span class="summary-label">📅 查询日期：</span>
              <span class="summary-value">{{ resultData.target_date }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">💰 总收益：</span>
              <span class="summary-value total-morale">+{{ filteredTotalMorale.toLocaleString() }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">📊 记录数：</span>
              <span class="summary-value">{{ filteredItems.length }} 条</span>
            </div>
          </div>

          <!-- 桌面端表格视图 -->
          <a-table
            :columns="columns"
            :data-source="filteredItems"
            :pagination="false"
            :loading="isLoading"
            row-key="id"
            class="morale-table desktop-view"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'time'">
                <span class="time-cell">{{ formatTime(record.Time) }}</span>
              </template>
              <template v-else-if="column.key === 'action'">
                <a-tag color="success">
                  📈 {{ record.action }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'num'">
                <span class="num-cell income-num">
                  +{{ (record.morale || 0).toLocaleString() }}
                </span>
              </template>
            </template>
          </a-table>

          <!-- 移动端卡片视图 -->
          <div class="mobile-card-list mobile-view">
            <div v-for="(record, index) in filteredItems" :key="index" class="mobile-card">
              <div class="mobile-card-row">
                <span class="mobile-label">📅 时间</span>
                <span class="mobile-value time-cell">{{ formatTime(record.Time) }}</span>
              </div>
              <div class="mobile-card-row">
                <span class="mobile-label">📊 类型</span>
                <a-tag color="success" class="mobile-tag">
                  📈 {{ record.action }}
                </a-tag>
              </div>
              <div class="mobile-card-row highlight">
                <span class="mobile-label">💰 数量</span>
                <span class="mobile-value num-cell income-num">
                  +{{ (record.morale || 0).toLocaleString() }}
                </span>
              </div>
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
import locale from 'ant-design-vue/es/date-picker/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import * as XLSX from 'xlsx'

dayjs.locale('zh-cn')

export default {
  name: 'Morale',
  data() {
    return {
      locale,
      columns: [
        {
          title: '📅 时间',
          key: 'time',
          dataIndex: 'Time',
          width: '40%'
        },
        {
          title: '📊 类型',
          key: 'action',
          dataIndex: 'action',
          width: '30%',
          align: 'center'
        },
        {
          title: '💰 数量',
          key: 'num',
          dataIndex: 'morale',
          width: '30%',
          align: 'right'
        }
      ],
      filters: {
        type: 'day',
        date: this.getTodayDate(),
        action: ''
      },
      dateValue: null,
      resultData: null,
      isLoading: false,
      isUpdating: false
    }
  },
  mounted() {
    this.dateValue = this.getTodayDate()
    this.searchRecords()
  },
  computed: {
    filteredItems() {
      if (!this.resultData || !this.resultData.items) {
        return []
      }
      if (!this.filters.action) {
        return this.resultData.items
      }
      return this.resultData.items.filter(item => item.action === this.filters.action)
    },
    
    // 根据筛选条件计算总收益
    filteredTotalMorale() {
      if (!this.filteredItems || this.filteredItems.length === 0) {
        return 0
      }
      return this.filteredItems.reduce((total, item) => {
        return total + (item.morale || 0)
      }, 0)
    }
  },
  methods: {
    goBack() {
      this.$router.go(-1)
    },
    
    getTodayDate() {
      const today = new Date()
      const year = today.getFullYear()
      const month = String(today.getMonth() + 1).padStart(2, '0')
      const day = String(today.getDate()).padStart(2, '0')
      return `${year}-${month}-${day}`
    },
    
    onDateChange(date, dateString) {
      // Ant Design 组件的日期变化回调
      this.filters.date = dateString || ''
    },
    
    onTypeChange() {
      // 当切换统计周期时，自动调整日期格式
      const today = new Date()
      const year = today.getFullYear()
      const month = String(today.getMonth() + 1).padStart(2, '0')
      const day = String(today.getDate()).padStart(2, '0')
      
      if (this.filters.type === 'day') {
        this.dateValue = `${year}-${month}-${day}`
        this.filters.date = `${year}-${month}-${day}`
      } else if (this.filters.type === 'month') {
        this.dateValue = `${year}-${month}`
        this.filters.date = `${year}-${month}`
      } else if (this.filters.type === 'year') {
        this.dateValue = `${year}`
        this.filters.date = `${year}`
      }
    },
    
    async searchRecords() {
      if (!this.filters.date) {
        message.warning('请选择查询日期')
        return
      }
      
      try {
        this.isLoading = true
        
        // 构建查询参数
        const params = {
          type: this.filters.type,
          date: this.filters.date
        }
        
        const response = await api.get('/api/BagStatistics/Morale', { params })
        console.log('API返回数据:', response)
        console.log('response.data:', response.data)
        
        // 后端返回的是 { data: { target_date, total_morale, items } }
        if (response.data && response.data.data) {
          // 如果有嵌套的data字段
          this.resultData = response.data.data
        } else {
          // 如果没有嵌套，直接使用
          this.resultData = response.data
        }
        console.log('resultData:', this.resultData)
        
        if (!this.resultData || !this.resultData.items || this.resultData.items.length === 0) {
          message.info('该日期暂无摩拉记录')
        }
      } catch (error) {
        console.error('查询摩拉记录失败:', error)
        message.error('查询失败，请稍后重试')
        this.resultData = null
      } finally {
        this.isLoading = false
      }
    },
    
    resetFilters() {
      this.filters = {
        type: 'day',
        date: this.getTodayDate(),
        action: ''
      }
      this.searchRecords()
    },
    
    formatTime(timeStr) {
      if (!timeStr) return '-'
      return timeStr.replace('T', ' ').substring(0, 19)
    },
    
    async updateMoraleRecord() {
      try {
        this.isUpdating = true
        
        // 显示加载提示
        const loadingMessage = message.loading('正在更新摩拉记录，请耐心等待...', 0)
        
        const response = await api.post('/api/BagStatistics/updateMorale')
        console.log('更新摩拉记录返回:', response)
        
        // 关闭加载提示
        loadingMessage()
        
        // 获取后端返回的消息
        const messageText = response.message
        
        // 弹框提示，显示时间更长
        message.success({
          content: messageText,
          duration: 10  // 增加到10秒
        })
        
        // 更新成功后自动刷新当前数据
        await this.searchRecords()
      } catch (error) {
        console.error('更新摩拉记录失败:', error)
        message.error({
          content: '更新失败，请稍后重试',
          duration: 5
        })
      } finally {
        this.isUpdating = false
      }
    },
    
    exportToExcel() {
      if (!this.resultData || !this.resultData.items || this.resultData.items.length === 0) {
        message.warning('暂无数据可导出')
        return
      }
      
      try {
        // 准备Excel数据
        const excelData = this.filteredItems.map((item, index) => ({
          '序号': index + 1,
          '时间': this.formatTime(item.Time),
          '类型': item.action,
          '数量': item.morale
        }))
        
        // 添加汇总行
        excelData.push({
          '序号': '',
          '时间': '',
          '类型': '总计',
          '数量': this.resultData.total_morale || 0
        })
        
        // 创建工作簿
        const ws = XLSX.utils.json_to_sheet(excelData)
        
        // 设置列宽
        ws['!cols'] = [
          { wch: 8 },  // 序号
          { wch: 20 }, // 时间
          { wch: 18 }, // 类型
          { wch: 15 }  // 数量
        ]
        
        const wb = XLSX.utils.book_new()
        XLSX.utils.book_append_sheet(wb, ws, '摩拉收益统计')
        
        // 生成文件名
        const fileName = `摩拉收益统计_${this.resultData.target_date || this.filters.date}_${Date.now()}.xlsx`
        
        // 导出文件
        XLSX.writeFile(wb, fileName)
        
        message.success('导出Excel成功！')
      } catch (error) {
        console.error('导出Excel失败:', error)
        message.error('导出失败，请稍后重试')
      }
    }
  }
}
</script>

<style scoped>
/* 二次元粉丝风格主题 */
.kawaii-page {
  --k-pink-primary: #FF6B9D;
  --k-pink-light: #FFB7D5;
  --k-pink-ultra-light: #FFE5F0;
  --k-purple-primary: #C77DFF;
  --k-purple-light: #E0AAFF;
  --k-purple-ultra-light: #F2E5FF;
  --k-blue-primary: #7DD3FC;
  --k-blue-light: #BAE6FD;
  --k-yellow-accent: #FDE68A;
  --k-mint: #A7F3D0;
  --k-mint-light: #D1FAE5;
  --k-gold: #FFC107;
  --k-text-dark: #4A2463;
  --k-text-light: #9D6FB8;
  --k-white: #FFFFFF;
  --k-radius: 28px;
  --k-radius-sm: 16px;
  --k-shadow: 0 10px 30px rgba(255, 107, 157, 0.25), 0 4px 10px rgba(199, 125, 255, 0.15);
  --k-shadow-hover: 0 15px 40px rgba(255, 107, 157, 0.35), 0 6px 15px rgba(199, 125, 255, 0.25);
  --k-border: 3px solid var(--k-pink-primary);

  position: relative;
  min-height: 100vh;
  background: linear-gradient(135deg, #FFE5F0 0%, #F2E5FF 50%, #E0F2FE 100%);
  color: var(--k-text-dark);
  font-family: "Nunito", "PingFang SC", "Microsoft YaHei", cursive, sans-serif;
  overflow-x: hidden;
  animation: gradientShift 10s ease infinite;
}

@keyframes gradientShift {
  0%, 100% { filter: hue-rotate(0deg) brightness(1); }
  50% { filter: hue-rotate(5deg) brightness(1.05); }
}

/* 背景装饰 */
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

.coin-1 { top: 10%; left: 5%; animation-delay: 0s; font-size: 3rem; }
.star-1 { top: 20%; right: 10%; animation-delay: -5s; }
.coin-2 { bottom: 15%; right: 20%; animation-delay: -10s; }
.star-2 { bottom: 10%; left: 15%; animation-delay: -15s; }
.sparkle-1 { top: 30%; left: 20%; animation-delay: -3s; font-size: 2.5rem; }
.sparkle-2 { top: 60%; right: 15%; animation-delay: -8s; font-size: 2rem; }
.sparkle-3 { bottom: 30%; left: 10%; animation-delay: -12s; font-size: 2.5rem; }
.sparkle-4 { top: 50%; right: 30%; animation-delay: -6s; font-size: 2rem; }

@keyframes floatAround {
  0% { transform: translate(0, 0) rotate(0deg); }
  25% { transform: translate(50px, 50px) rotate(90deg); }
  50% { transform: translate(0, 100px) rotate(180deg); }
  75% { transform: translate(-50px, 50px) rotate(270deg); }
  100% { transform: translate(0, 0) rotate(360deg); }
}

/* 主容器 */
.main-container {
  position: relative;
  z-index: 1;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

/* 通用按钮 - 二次元风格 */
.kawaii-btn {
  padding: 12px 24px;
  border-radius: 50px;
  border: 3px solid var(--k-pink-primary);
  background: linear-gradient(135deg, var(--k-pink-ultra-light) 0%, var(--k-purple-ultra-light) 100%);
  color: var(--k-text-dark);
  font-weight: bold;
  cursor: pointer;
  box-shadow: 0 6px 0 var(--k-pink-primary), 0 8px 20px rgba(255, 107, 157, 0.3);
  transition: all 0.3s cubic-bezier(0.68, -0.55, 0.27, 1.55);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 1rem;
  position: relative;
  overflow: hidden;
}

.kawaii-btn::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.5), transparent);
  transform: rotate(45deg);
  animation: buttonShine 3s infinite;
}

@keyframes buttonShine {
  0%, 100% { transform: translateX(-100%) translateY(-100%) rotate(45deg); }
  50% { transform: translateX(100%) translateY(100%) rotate(45deg); }
}

.kawaii-btn:hover {
  transform: translateY(3px) scale(1.02);
  box-shadow: 0 3px 0 var(--k-pink-primary), 0 5px 15px rgba(255, 107, 157, 0.4);
}

.kawaii-btn:active {
  transform: translateY(6px) scale(0.98);
  box-shadow: 0 0 0 var(--k-pink-primary);
}

.kawaii-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.kawaii-btn.primary {
  background: linear-gradient(135deg, var(--k-pink-primary) 0%, var(--k-purple-primary) 100%);
  color: var(--k-white);
  border-color: var(--k-pink-primary);
  box-shadow: 0 6px 0 #E63E7A, 0 8px 20px rgba(255, 107, 157, 0.4);
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.kawaii-btn.small {
  padding: 6px 14px;
  font-size: 0.9rem;
}

/* Header - 二次元风格 */
.kawaii-header {
  background: linear-gradient(135deg, rgba(255, 229, 240, 0.95) 0%, rgba(242, 229, 255, 0.95) 100%);
  backdrop-filter: blur(15px);
  border-radius: var(--k-radius);
  border: 4px solid transparent;
  background-clip: padding-box;
  position: relative;
  margin-bottom: 30px;
  box-shadow: var(--k-shadow);
  padding: 25px 30px;
}

.kawaii-header::before {
  content: '';
  position: absolute;
  inset: -4px;
  border-radius: var(--k-radius);
  padding: 4px;
  background: linear-gradient(135deg, var(--k-pink-primary), var(--k-purple-primary), var(--k-blue-primary));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  z-index: -1;
}

.header-top {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 20px;
}

.title-box {
  text-align: center;
  grid-column: 2;
  white-space: nowrap;
}

.main-title {
  margin: 0;
  font-size: 2rem;
  background: linear-gradient(135deg, var(--k-pink-primary), var(--k-purple-primary), var(--k-blue-primary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 900;
  text-shadow: 3px 3px 6px rgba(255, 107, 157, 0.3);
  animation: titleGlow 2s ease-in-out infinite;
}

@keyframes titleGlow {
  0%, 100% { filter: drop-shadow(0 0 8px rgba(255, 107, 157, 0.6)); }
  50% { filter: drop-shadow(0 0 15px rgba(199, 125, 255, 0.8)); }
}

.sub-title {
  display: block;
  font-size: 0.85rem;
  color: var(--k-text-light);
  letter-spacing: 1px;
  margin-top: 5px;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 30px;
}

.header-actions.left {
  justify-content: flex-start;
  grid-column: 1;
}

.header-actions.right {
  justify-content: flex-end;
  grid-column: 3;
}

.home-btn {
  background: linear-gradient(135deg, var(--k-blue-light) 0%, var(--k-blue-primary) 100%);
  border-color: var(--k-blue-primary);
  color: #0369A1;
  box-shadow: 0 6px 0 #0891B2, 0 8px 20px rgba(125, 211, 252, 0.3);
}

.update-btn {
  background: linear-gradient(135deg, var(--k-purple-ultra-light) 0%, var(--k-purple-light) 100%);
  border-color: var(--k-purple-primary);
  color: #6B21A8;
  box-shadow: 0 6px 0 #A855F7, 0 8px 20px rgba(199, 125, 255, 0.3);
}

.update-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 0.8; }
}

.reset-btn {
  background: linear-gradient(135deg, var(--k-mint-light) 0%, var(--k-mint) 100%);
  border-color: #34D399;
  color: #065F46;
  box-shadow: 0 6px 0 #10B981, 0 8px 20px rgba(167, 243, 208, 0.3);
}

.export-btn {
  background: linear-gradient(135deg, var(--k-yellow-accent) 0%, var(--k-gold) 100%);
  border-color: var(--k-gold);
  color: #92400E;
  box-shadow: 0 6px 0 #F59E0B, 0 8px 20px rgba(253, 230, 138, 0.3);
}

.export-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.icon-btn .btn-text {
  display: none;
}


@media (min-width: 768px) {
  .icon-btn .btn-text {
    display: inline;
  }
}

/* 筛选区域 - 二次元风格 */
.filter-section {
  margin-bottom: 25px;
  position: relative;
  z-index: 10;
}

.filter-card {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 229, 240, 0.9) 100%);
  border-radius: var(--k-radius);
  border: 4px solid transparent;
  background-clip: padding-box;
  position: relative;
  padding: 30px;
  box-shadow: var(--k-shadow);
  overflow: visible;
}

.filter-card::before {
  content: '';
  position: absolute;
  inset: -4px;
  border-radius: var(--k-radius);
  padding: 4px;
  background: linear-gradient(135deg, var(--k-pink-light), var(--k-purple-light));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  z-index: -1;
}

.filter-title {
  margin: 0 0 25px 0;
  font-size: 1.4rem;
  background: linear-gradient(90deg, var(--k-pink-primary), var(--k-purple-primary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 800;
  text-align: center;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  position: relative;
  z-index: 1;
}

.filter-label {
  font-weight: bold;
  color: var(--k-text-dark);
  font-size: 0.95rem;
}

.kawaii-input,
.kawaii-select {
  padding: 12px 18px;
  border-radius: 25px;
  border: 3px solid var(--k-pink-light);
  background: linear-gradient(135deg, var(--k-white) 0%, var(--k-pink-ultra-light) 100%);
  color: var(--k-text-dark);
  font-weight: bold;
  font-size: 1rem;
  transition: all 0.3s;
  box-shadow: 0 4px 10px rgba(255, 107, 157, 0.1);
}

/* Ant Design 日期选择器样式 - 提高优先级覆盖全局样式 */
.kawaii-date-picker {
  width: 100% !important;
}

.kawaii-date-picker :deep(.ant-picker) {
  width: 100% !important;
  border-radius: 25px !important;
  border: 3px solid var(--k-pink-light) !important;
  padding: 10px 18px !important;
  font-weight: bold !important;
  background: linear-gradient(135deg, var(--k-white) 0%, var(--k-pink-ultra-light) 100%) !important;
  transition: all 0.3s !important;
  box-shadow: 0 4px 10px rgba(255, 107, 157, 0.1) !important;
}

.kawaii-date-picker :deep(.ant-picker-input) {
  display: flex !important;
  align-items: center !important;
  flex: 1 !important;
  width: 100% !important;
}

.kawaii-date-picker :deep(.ant-picker-input > input) {
  font-weight: bold !important;
  color: var(--k-text-dark) !important;
  font-size: 1rem !important;
  background: transparent !important;
  flex: 1 !important;
  width: 100% !important;
  min-width: 0 !important;
  text-overflow: clip !important;
}

.kawaii-date-picker :deep(.ant-picker:hover) {
  border-color: var(--k-pink-primary) !important;
  box-shadow: 0 6px 15px rgba(255, 107, 157, 0.2) !important;
}

.kawaii-date-picker :deep(.ant-picker-focused) {
  border-color: var(--k-pink-primary) !important;
  box-shadow: 0 0 0 4px rgba(255, 107, 157, 0.15), 0 6px 15px rgba(199, 125, 255, 0.2) !important;
}

.kawaii-date-picker :deep(.ant-picker-suffix) {
  color: var(--k-pink-primary) !important;
  font-size: 1.3rem !important;
}

.kawaii-date-picker :deep(.ant-picker-clear) {
  background: var(--k-white) !important;
  color: var(--k-text-light) !important;
}

/* 日期选择器弹出面板样式 - 确保日历面板宽度足够 */
.kawaii-date-picker :deep(.ant-picker-dropdown) {
  min-width: 280px !important;
}

.kawaii-date-picker :deep(.ant-picker-panel-container) {
  min-width: 280px !important;
}

.kawaii-date-picker :deep(.ant-picker-date-panel),
.kawaii-date-picker :deep(.ant-picker-month-panel),
.kawaii-date-picker :deep(.ant-picker-year-panel) {
  min-width: 280px !important;
}

.kawaii-date-picker :deep(.ant-picker-content) {
  width: 100% !important;
  min-width: 260px !important;
}

.kawaii-date-picker :deep(.ant-picker-cell) {
  min-width: 32px !important;
}

.kawaii-input:focus,
.kawaii-select:focus {
  outline: none;
  border-color: var(--k-pink-primary);
  box-shadow: 0 0 0 4px rgba(255, 107, 157, 0.15), 0 6px 15px rgba(255, 107, 157, 0.2);
  transform: translateY(-2px);
}

.filter-actions {
  display: flex;
  justify-content: center;
  gap: 15px;
}

/* 数据展示区 - 二次元风格 */
.data-section {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9) 0%, rgba(255, 229, 240, 0.85) 100%);
  border-radius: var(--k-radius);
  border: 4px solid transparent;
  background-clip: padding-box;
  position: relative;
  padding: 25px;
  box-shadow: var(--k-shadow);
  min-height: 400px;
}

.data-section::before {
  content: '';
  position: absolute;
  inset: -4px;
  border-radius: var(--k-radius);
  padding: 4px;
  background: linear-gradient(135deg, var(--k-pink-light), var(--k-purple-light), var(--k-blue-light));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  z-index: -1;
}

/* 加载和空状态 */
.loading-state,
.empty-state {
  text-align: center;
  padding: 60px 0;
  color: var(--k-text-light);
}

.loading-spinner {
  font-size: 4rem;
  animation: spin 2s linear infinite;
  display: inline-block;
  margin-bottom: 20px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-img {
  font-size: 5rem;
  margin-bottom: 20px;
  opacity: 0.7;
}

.empty-text {
  margin-bottom: 25px;
  font-size: 1.1rem;
}

/* 表格样式 - 二次元风格 */
.morale-table :deep(.ant-table) {
  background: transparent;
  border-radius: var(--k-radius-sm);
  overflow: hidden;
}

.morale-table :deep(.ant-table-thead > tr > th) {
  background: linear-gradient(135deg, var(--k-pink-ultra-light) 0%, var(--k-purple-ultra-light) 100%);
  color: var(--k-text-dark);
  font-weight: 800;
  border-bottom: 3px solid var(--k-pink-light);
  font-size: 1.05rem;
  text-shadow: 0 1px 2px rgba(255, 255, 255, 0.8);
}

.morale-table :deep(.ant-table-tbody > tr > td) {
  background: rgba(255, 255, 255, 0.9);
  border-bottom: 2px solid var(--k-pink-ultra-light);
  transition: all 0.3s;
}

.morale-table :deep(.ant-table-tbody > tr:hover > td) {
  background: linear-gradient(135deg, var(--k-pink-ultra-light) 0%, rgba(255, 255, 255, 0.95) 100%);
  transform: scale(1.01);
  box-shadow: 0 4px 12px rgba(255, 107, 157, 0.15);
}

.morale-table :deep(.ant-tag) {
  border-radius: 25px;
  padding: 8px 16px;
  font-weight: bold;
  font-size: 0.95rem;
  background: linear-gradient(135deg, var(--k-mint-light) 0%, var(--k-mint) 100%);
  border: 2px solid #34D399;
  color: #065F46;
  box-shadow: 0 2px 8px rgba(167, 243, 208, 0.3);
}

.time-cell {
  color: var(--k-text-light);
  font-weight: bold;
}

.num-cell {
  font-family: "Comic Sans MS", cursive, sans-serif;
  font-size: 1.2rem;
  font-weight: bold;
}

.income-num {
  background: linear-gradient(135deg, #34D399, #10B981);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 900;
}

.expense-num {
  color: var(--k-red);
}

/* 汇总卡片 - 二次元风格 */
.summary-card {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
  padding: 25px;
  background: linear-gradient(135deg, var(--k-pink-ultra-light) 0%, var(--k-purple-ultra-light) 50%, var(--k-blue-light) 100%);
  border-radius: var(--k-radius);
  border: 4px solid transparent;
  background-clip: padding-box;
  position: relative;
  box-shadow: var(--k-shadow);
}

.summary-card::before {
  content: '';
  position: absolute;
  inset: -4px;
  border-radius: var(--k-radius);
  padding: 4px;
  background: linear-gradient(135deg, var(--k-pink-primary), var(--k-purple-primary), var(--k-blue-primary));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  z-index: -1;
  animation: borderRotate 3s linear infinite;
}

@keyframes borderRotate {
  0% { filter: hue-rotate(0deg); }
  100% { filter: hue-rotate(360deg); }
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: var(--k-radius-sm);
  border: 3px solid var(--k-pink-light);
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
}

.summary-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.6), transparent);
  transition: left 0.5s;
}

.summary-item:hover {
  transform: translateY(-5px) scale(1.02);
  box-shadow: 0 10px 25px rgba(255, 107, 157, 0.4);
  border-color: var(--k-pink-primary);
}

.summary-item:hover::before {
  left: 100%;
}

.summary-label {
  font-size: 0.95rem;
  color: var(--k-text-light);
  font-weight: bold;
}

.summary-value {
  font-size: 1.3rem;
  font-weight: bold;
  color: var(--k-text-dark);
  font-family: "Comic Sans MS", cursive, sans-serif;
}

.summary-value.total-morale {
  font-size: 1.8rem;
  background: linear-gradient(135deg, var(--k-pink-primary), var(--k-purple-primary), var(--k-gold));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 900;
  animation: valueGlow 2s ease-in-out infinite;
}

@keyframes valueGlow {
  0%, 100% { filter: drop-shadow(0 0 6px rgba(255, 107, 157, 0.5)); }
  50% { filter: drop-shadow(0 0 12px rgba(199, 125, 255, 0.7)); }
}

/* 桌面端和移动端视图切换 */
.desktop-view {
  display: block;
}

.mobile-view {
  display: none!important;;
}

/* 移动端卡片布局 */
.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.mobile-card {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(255, 229, 240, 0.95) 100%);
  border-radius: var(--k-radius-sm);
  border: 3px solid var(--k-pink-light);
  padding: 18px;
  box-shadow: 0 4px 12px rgba(255, 107, 157, 0.15);
  transition: all 0.3s;
}

.mobile-card:active {
  transform: scale(0.98);
  box-shadow: 0 2px 8px rgba(255, 107, 157, 0.2);
}

.mobile-card-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 2px solid var(--k-pink-ultra-light);
}

.mobile-card-row:last-child {
  border-bottom: none;
}

.mobile-card-row.highlight {
  background: linear-gradient(135deg, var(--k-pink-ultra-light) 0%, rgba(255, 255, 255, 0.5) 100%);
  margin: 8px -18px -18px;
  padding: 15px 18px;
  border-radius: 0 0 var(--k-radius-sm) var(--k-radius-sm);
  border-bottom: none;
}

.mobile-label {
  font-weight: bold;
  color: var(--k-text-dark);
  font-size: 0.95rem;
}

.mobile-value {
  font-weight: bold;
  color: var(--k-text-dark);
  text-align: right;
}

.mobile-tag {
  margin: 0;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .kawaii-page {
    overflow-x: hidden;
  }

  /* 切换到移动端视图 */
  .desktop-view {
    display: none !important;
  }

  .mobile-view {
    display: block !important;
  }

  .kawaii-header {
    padding: 20px 15px;
  }

  .header-top {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .title-box {
    order: -1;
    margin-bottom: 8px;
  }

  .header-actions.left,
  .header-actions.right {
    width: 100%;
    justify-content: stretch;
  }

  .header-actions.left {
    order: 1;
  }

  .header-actions.right {
    order: 2;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .kawaii-btn {
    width: 100%;
    padding: 12px 12px;
    font-size: 0.9rem;
    justify-content: center;
    box-sizing: border-box;
  }

  .kawaii-btn.icon-btn .btn-text {
    display: inline;
  }

  /* 按钮适配 */
  .home-btn,
  .update-btn,
  .reset-btn {
    width: 100%;
    white-space: nowrap;
    text-align: center;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 44px;
    box-sizing: border-box;
  }

  .main-title {
    font-size: 1.5rem;
  }

  .sub-title {
    font-size: 0.75rem;
  }

  .filter-grid {
    grid-template-columns: 1fr;
    gap: 15px;
  }

  .filter-card {
    padding: 15px;
  }

  .summary-card {
    grid-template-columns: 1fr;
    gap: 15px;
    padding: 15px;
  }

  .main-container {
    padding: 10px;
  }

  /* 移动端日期选择器弹出面板优化 */
  .kawaii-date-picker :deep(.ant-picker-dropdown) {
    max-width: calc(100vw - 20px) !important;
    left: 10px !important;
    right: 10px !important;
  }

  .kawaii-date-picker :deep(.ant-picker-panel-container) {
    max-width: 100% !important;
  }

  .kawaii-date-picker :deep(.ant-picker-date-panel),
  .kawaii-date-picker :deep(.ant-picker-month-panel),
  .kawaii-date-picker :deep(.ant-picker-year-panel) {
    max-width: 100% !important;
  }

  .kawaii-date-picker :deep(.ant-picker-content) {
    max-width: 100% !important;
  }

  .kawaii-date-picker :deep(.ant-picker-cell) {
    padding: 4px 0 !important;
  }

  .kawaii-date-picker :deep(.ant-picker-cell-inner) {
    min-width: 28px !important;
    height: 28px !important;
    line-height: 28px !important;
    font-size: 0.85rem !important;
  }

  /* 移动端卡片样式优化 */
  .mobile-card {
    padding: 15px;
  }

  .mobile-card-row {
    padding: 8px 0;
  }

  .mobile-card-row.highlight {
    margin: 8px -15px -15px;
    padding: 12px 15px;
  }

  .mobile-label {
    font-size: 0.9rem;
  }

  .mobile-value {
    font-size: 0.95rem;
  }

  .num-cell {
    font-size: 1.1rem;
  }
}
</style>
