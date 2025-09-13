<template>
  <div class="container">
    <!-- 标题区域 -->
    <header class="header">
      <h1 class="title">🌸 一条龙配置 🌸</h1>
    </header>

    <!-- 配置选择区域 -->
    <section class="config-section">
      <div class="input-group">
        <label for="configSelect" class="label">选择配置</label>
        <div class="select-button-group">
          <select 
            id="configSelect" 
            v-model="selectedConfig" 
            @change="loadConfig"
            class="select-input"
          >
            <option v-for="config in configList" :key="config" :value="config">
              {{ config }}
            </option>
          </select>
          <button @click="loadConfig" class="action-btn secondary">
            📁 加载
          </button>
        </div>
      </div>
    </section>

    <!-- 搜索区域 -->
    <section class="search-section">
      <h2 class="section-title">🔍 资源搜索</h2>
      
      <!-- 天赋书搜索 -->
      <div class="search-item">
        <div class="input-group">
          <label for="talentSelect" class="label">天赋书搜索</label>
          <div class="select-button-group">
            <select 
              id="talentSelect" 
              v-model="selectedTalent"
              class="select-input"
            >
              <option value="">请选择天赋书</option>
              <option v-for="talent in talentList" :key="talent" :value="talent">
                {{ talent }}
              </option>
            </select>
            <button 
              @click="searchModule('talent')" 
              class="action-btn secondary"
              :disabled="!selectedTalent"
            >
              🔍 搜索
            </button>
          </div>
          <div v-if="talentResult" class="search-result">
            {{ talentResult }}
          </div>
        </div>
      </div>

      <!-- 武器材料搜索 -->
      <div class="search-item">
        <div class="input-group">
          <label for="weaponSelect" class="label">武器材料搜索</label>
          <div class="select-button-group">
            <select 
              id="weaponSelect" 
              v-model="selectedWeapon"
              class="select-input"
            >
              <option value="">请选择武器材料</option>
              <option v-for="weapon in weaponList" :key="weapon" :value="weapon">
                {{ weapon }}
              </option>
            </select>
            <button 
              @click="searchModule('weapon')" 
              class="action-btn secondary"
              :disabled="!selectedWeapon"
            >
              🔍 搜索
            </button>
          </div>
          <div v-if="weaponResult" class="search-result">
            {{ weaponResult }}
          </div>
        </div>
      </div>
    </section>

    <!-- 每日副本配置 -->
    <section class="domain-section">
      <h2 class="section-title">📅 每日副本</h2>
      <div class="domain-grid">
        <div v-for="(day, index) in dayKeys" :key="day" class="domain-item">
          <label class="domain-label">{{ dayNames[index] }}</label>
          <select 
            v-model="currentConfig[day + 'DomainName']"
            class="domain-select"
          >
            <option value="">请选择副本</option>
            <option v-for="domain in domainOptions" :key="domain" :value="domain">
              {{ domain }}
            </option>
          </select>
        </div>
      </div>
    </section>

    <!-- 完成后动作 -->
    <section class="action-section">
      <h2 class="section-title">⚙️ 完成后动作</h2>
      <div class="input-group">
        <select 
          v-model="currentConfig.CompletionAction"
          class="completion-select"
        >
          <option v-for="action in actionOptions" :key="action" :value="action">
            {{ action }}
          </option>
        </select>
      </div>
    </section>

    <!-- 任务开关 -->
    <section class="task-section">
      <h2 class="section-title">🎯 任务开关</h2>
      <div class="task-grid">
        <div 
          v-for="(enabled, taskName) in currentConfig.TaskEnabledList" 
          :key="taskName" 
          class="task-item"
        >
          <span class="task-name">{{ taskName }}</span>
          <label class="switch">
            <input 
              type="checkbox" 
              v-model="currentConfig.TaskEnabledList[taskName]"
            >
            <span class="slider"></span>
          </label>
        </div>
      </div>
    </section>

    <!-- 保存按钮 -->
    <footer class="footer">
      <button @click="saveConfig" class="save-btn">
        💾 保存配置
      </button>
    </footer>
  </div>
</template>

<script>
import { ref, reactive, onMounted, nextTick } from 'vue'

export default {
  name: 'OneLong',
  setup() {
    // ============ 常量配置 ============
    const dayKeys = [
      "Sunday", "Monday", "Tuesday", "Wednesday", 
      "Thursday", "Friday", "Saturday"
    ]
    
    const dayNames = [
      "星期天", "星期一", "星期二", "星期三", 
      "星期四", "星期五", "星期六"
    ]
    
    const domainOptions = [
      "无妄引咎密宫", "孤云凌霄之处", "华池岩岫", "仲夏庭院", 
      "铭记之谷", "芬德尼尔之顶", "山脊守望", "沉眠之庭", 
      "椛染之庭", "缘觉塔", "熔铁的孤塞", "罪祸的终末",
      "岩中幽谷", "朽废的集所", "临瀑之城", "荒废砌造坞", 
      "忘却之峡", "太山府", "堇色之庭", "昏识塔", 
      "蕴火的幽墟", "塞西莉亚苗圃", "震雷连山密宫", "砂流之庭",
      "有顶塔", "深潮的余响","深古瞭望所","霜凝的机枢","失落的月庭","无光的深都"
    ]
    
    const actionOptions = ["无", "关闭游戏和软件", "关机"]

    // ============ 响应式数据 ============
    const configList = ref([])
    const selectedConfig = ref('')
    const talentList = ref([])
    const selectedTalent = ref('')
    const weaponList = ref([])
    const selectedWeapon = ref('')
    const talentResult = ref('')
    const weaponResult = ref('')
    
    // 当前配置对象
    const currentConfig = reactive({
      Name: '',
      CompletionAction: '',
      TaskEnabledList: {},
      SecretTreasureObjects: [],
      SundayDomainName: '',
      MondayDomainName: '',
      TuesdayDomainName: '',
      WednesdayDomainName: '',
      ThursdayDomainName: '',
      FridayDomainName: '',
      SaturdayDomainName: ''
    })

    // ============ API 相关方法 ============
    
    /**
     * 通用API请求方法
     * @param {string} url - 请求URL
     * @param {object} options - 请求选项
     */
    const apiRequest = async (url, options = {}) => {
      try {
        const response = await fetch(url, {
          headers: {
            'Content-Type': 'application/json',
            ...options.headers
          },
          ...options
        })
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        return await response.json()
      } catch (error) {
        console.error('API request failed:', error)
        throw error
      }
    }

    /**
     * 初始化搜索模块数据
     * @param {string} type - 模块类型 ('talent' | 'weapon')
     * @param {string} apiPrefix - API前缀
     */
    const initSearchModule = async (type, apiPrefix) => {
      try {
        const data = await apiRequest(apiPrefix)
        
        if (type === 'talent') {
          talentList.value = data.data || []
        } else if (type === 'weapon') {
          weaponList.value = data.data || []
        }
      } catch (error) {
        console.error(`Failed to load ${type} data:`, error)
        // 可以添加用户友好的错误提示
      }
    }

    /**
     * 搜索模块
     * @param {string} type - 搜索类型 ('talent' | 'weapon')
     */
    const searchModule = async (type) => {
      const name = type === 'talent' ? selectedTalent.value : selectedWeapon.value
      
      if (!name) {
        const resultRef = type === 'talent' ? talentResult : weaponResult
        resultRef.value = "请先选择要搜索的内容"
        return
      }

      const apiMap = {
        talent: '/api/talentBooks/search',
        weapon: '/api/WeaponDomain/search'
      }
      
      try {
        const data = await apiRequest(`${apiMap[type]}?name=${encodeURIComponent(name)}`)
        
        if (data.status !== "success" || !data.data || data.data.length === 0) {
          const message = "没有找到相关秘境信息"
          if (type === 'talent') {
            talentResult.value = message
          } else {
            weaponResult.value = message
          }
          return
        }
        
        const resultText = data.data
          .map(item => `秘境：${item.DomainName} ｜ 时间：${dayNames[item.Weekday]}`)
          .join("\n")
        
        if (type === 'talent') {
          talentResult.value = resultText
        } else {
          weaponResult.value = resultText
        }
      } catch (error) {
        const errorMessage = "搜索请求失败，请稍后重试"
        if (type === 'talent') {
          talentResult.value = errorMessage
        } else {
          weaponResult.value = errorMessage
        }
      }
    }

    /**
     * 加载指定配置
     */
    const loadConfig = async () => {
      if (!selectedConfig.value) {
        console.warn('No config selected')
        return
      }
      
      try {
        const cfg = await apiRequest(
          `/api/onelong/config?name=${encodeURIComponent(selectedConfig.value)}`
        )
        
        // 重置并更新当前配置
        Object.assign(currentConfig, {
          Name: '',
          CompletionAction: '',
          TaskEnabledList: {},
          SecretTreasureObjects: [],
          SundayDomainName: '',
          MondayDomainName: '',
          TuesdayDomainName: '',
          WednesdayDomainName: '',
          ThursdayDomainName: '',
          FridayDomainName: '',
          SaturdayDomainName: '',
          ...cfg
        })
        
        // 确保 TaskEnabledList 存在
        if (!currentConfig.TaskEnabledList) {
          currentConfig.TaskEnabledList = {}
        }
      } catch (error) {
        console.error('Failed to load config:', error)
        alert('❌ 加载配置失败，请稍后重试')
      }
    }

    /**
     * 保存当前配置
     */
    const saveConfig = async () => {
      if (!currentConfig.Name) {
        alert('⚠️ 请先选择一个配置')
        return
      }

      const payload = {
        ...currentConfig,
        TaskEnabledList: { ...currentConfig.TaskEnabledList },
        SecretTreasureObjects: currentConfig.SecretTreasureObjects || []
      }

      try {
        const result = await apiRequest('/api/onelong/saveConfig', {
          method: 'POST',
          body: JSON.stringify(payload)
        })
        
        const message = result.status === "success" ? "✅ 保存成功" : "❌ 保存失败"
        alert(message)
      } catch (error) {
        console.error('Save config failed:', error)
        alert("❌ 保存失败，请稍后重试")
      }
    }

    /**
     * 加载配置列表
     */
    const loadConfigList = async () => {
      try {
        const data = await apiRequest('/api/oneLongAllName')
        configList.value = data.data || []
        
        // 自动选择第一个配置
        if (configList.value.length > 0) {
          selectedConfig.value = configList.value[0]
          await nextTick()
          await loadConfig()
        }
      } catch (error) {
        console.error('Failed to load config list:', error)
        alert('❌ 加载配置列表失败')
      }
    }

    // ============ 生命周期钩子 ============
    onMounted(async () => {
      try {
        // 并行加载所有初始数据
        await Promise.all([
          loadConfigList(),
          initSearchModule("talent", "/api/talentBooks"),
          initSearchModule("weapon", "/api/WeaponDomain")
        ])
      } catch (error) {
        console.error('Initialization failed:', error)
      }
    })

    // ============ 返回暴露的属性和方法 ============
    return {
      // 常量
      dayKeys,
      dayNames,
      domainOptions,
      actionOptions,
      
      // 响应式数据
      configList,
      selectedConfig,
      talentList,
      selectedTalent,
      weaponList,
      selectedWeapon,
      talentResult,
      weaponResult,
      currentConfig,
      
      // 方法
      searchModule,
      loadConfig,
      saveConfig
    }
  }
}
</script>

<style scoped>
/* ============ 基础样式 ============ */
* {
  box-sizing: border-box;
}

.container {
  max-width: 500px;
  margin: 0 auto;
  padding: 16px;
  font-family: "微软雅黑", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  background: linear-gradient(135deg, #ffd6e8 0%, #fff0f6 50%, #f8f0ff 100%);
  min-height: 100vh;
  color: #2d3748;
  line-height: 1.6;
}

/* ============ 标题区域 ============ */
.header {
  text-align: center;
  margin-bottom: 32px;
}

.title {
  font-size: 28px;
  font-weight: 700;
  color: #e53e3e;
  background: linear-gradient(45deg, #ff4fa2, #ff69b4, #da70d6);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  text-shadow: 0 2px 4px rgba(255, 105, 180, 0.3);
  margin: 0;
  letter-spacing: 0.5px;
}

/* ============ 区块样式 ============ */
section {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 4px 12px rgba(255, 105, 180, 0.1);
  border: 1px solid rgba(255, 157, 207, 0.2);
  backdrop-filter: blur(10px);
}

.section-title {
  font-size: 20px;
  font-weight: 600;
  color: #ff4fa2;
  margin: 0 0 16px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* ============ 表单元素样式 ============ */
.input-group {
  margin-bottom: 16px;
}

.label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 8px;
}

.select-button-group {
  display: flex;
  gap: 8px;
  align-items: stretch;
}

.select-input,
.domain-select,
.completion-select {
  flex: 1;
  padding: 12px 16px;
  font-size: 16px;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  background: white;
  color: #2d3748;
  transition: all 0.3s ease;
  appearance: none;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 12px center;
  background-repeat: no-repeat;
  background-size: 16px;
  padding-right: 40px;
}

.select-input:focus,
.domain-select:focus,
.completion-select:focus {
  outline: none;
  border-color: #ff69b4;
  box-shadow: 0 0 0 3px rgba(255, 105, 180, 0.1);
  transform: translateY(-1px);
}

/* ============ 按钮样式 ============ */
.action-btn {
  padding: 12px 20px;
  font-size: 14px;
  font-weight: 600;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  white-space: nowrap;
  min-width: 80px;
}

.action-btn.secondary {
  background: linear-gradient(135deg, #ff69b4, #ff4fa2);
  color: white;
  box-shadow: 0 2px 8px rgba(255, 105, 180, 0.3);
}

.action-btn.secondary:hover:not(:disabled) {
  background: linear-gradient(135deg, #ff4fa2, #e91e63);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 105, 180, 0.4);
}

.action-btn:disabled {
  background: #cbd5e0;
  color: #a0aec0;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.save-btn {
  width: 100%;
  padding: 16px;
  font-size: 18px;
  font-weight: 700;
  background: linear-gradient(135deg, #48bb78, #38a169);
  color: white;
  border: none;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(72, 187, 120, 0.3);
}

.save-btn:hover {
  background: linear-gradient(135deg, #38a169, #2f855a);
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(72, 187, 120, 0.4);
}

/* ============ 搜索结果样式 ============ */
.search-result {
  margin-top: 12px;
  padding: 12px;
  background: linear-gradient(135deg, #fff5f5, #fed7e2);
  border-radius: 8px;
  border-left: 4px solid #ff69b4;
  color: #c53030;
  font-weight: 500;
  white-space: pre-line;
  font-size: 14px;
  line-height: 1.5;
}

.search-item {
  margin-bottom: 20px;
}

.search-item:last-child {
  margin-bottom: 0;
}

/* ============ 每日副本网格 ============ */
.domain-grid {
  display: grid;
  gap: 12px;
}

.domain-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.domain-label {
  min-width: 80px;
  font-size: 14px;
  font-weight: 600;
  color: #4a5568;
}

.domain-select {
  flex: 1;
}

/* ============ 任务开关网格 ============ */
.task-grid {
  display: grid;
  gap: 12px;
}

.task-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #f7fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

.task-item:hover {
  background: #edf2f7;
  border-color: #ff9dcf;
}

.task-name {
  font-size: 14px;
  font-weight: 500;
  color: #2d3748;
  flex: 1;
  margin-right: 12px;
}

/* ============ 开关样式 ============ */
.switch {
  position: relative;
  display: inline-block;
  width: 52px;
  height: 28px;
  flex-shrink: 0;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #cbd5e0;
  border-radius: 28px;
  transition: all 0.3s ease;
}

.slider:before {
  position: absolute;
  content: "";
  height: 22px;
  width: 22px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

input:checked + .slider {
  background: linear-gradient(135deg, #ff69b4, #ff4fa2);
}

input:checked + .slider:before {
  transform: translateX(24px);
}

/* ============ 响应式设计 ============ */
@media (max-width: 600px) {
  .container {
    padding: 12px;
  }
  
  .title {
    font-size: 24px;
  }
  
  section {
    padding: 16px;
    margin-bottom: 16px;
  }
  
  .section-title {
    font-size: 18px;
  }
  
  .select-button-group {
    flex-direction: column;
    gap: 8px;
  }
  
  .action-btn {
    width: 100%;
  }
  
  .domain-item {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .domain-label {
    min-width: auto;
  }
  
  .task-item {
    padding: 16px;
  }
  
  .task-name {
    margin-right: 0;
    margin-bottom: 8px;
  }
  
  .domain-grid {
    gap: 8px;
  }
  
  .task-grid {
    gap: 8px;
  }
}

@media (max-width: 400px) {
  .container {
    padding: 8px;
  }
  
  .title {
    font-size: 20px;
  }
  
  section {
    padding: 12px;
  }
  
  .task-item {
    flex-direction: column;
    align-items: stretch;
    text-align: center;
  }
  
  .switch {
    align-self: center;
    margin-top: 8px;
  }
}

/* ============ 动画效果 ============ */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

section {
  animation: fadeIn 0.3s ease-out;
}

/* ============ 深色模式支持 ============ */
@media (prefers-color-scheme: dark) {
  .container {
    background: linear-gradient(135deg, #2d1b3d 0%, #1a1625 50%, #0f0a1a 100%);
    color: #e2e8f0;
  }
  
  section {
    background: rgba(45, 55, 72, 0.9);
    border-color: rgba(255, 157, 207, 0.1);
  }
  
  .select-input,
  .domain-select,
  .completion-select {
    background: #2d3748;
    border-color: #4a5568;
    color: #e2e8f0;
  }
  
  .task-item {
    background: #2d3748;
    border-color: #4a5568;
  }
  
  .task-item:hover {
    background: #4a5568;
  }
  
  .task-name {
    color: #e2e8f0;
  }
  
  .domain-label,
  .label {
    color: #cbd5e0;
  }
}
</style>
