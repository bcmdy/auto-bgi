<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: 'Comic Sans MS', 'Segoe UI', sans-serif;
  background-color: #ffecf5;
  color: #ff6699;
  overflow-x: hidden;
  overflow-y: auto;
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><circle cx="50" cy="50" r="5" fill="%23ffcce6" opacity="0.5"/></svg>');
}

#app {
  height: 100vh;
}

/* =========================== */
/* 日期选择器全局修复 - 最高优先级 */
/* =========================== */
.ant-picker-dropdown {
  z-index: 99999 !important;
}

.ant-picker-dropdown .ant-picker-panel-container {
  width: 280px !important;
  min-width: 280px !important;
}

.ant-picker-dropdown table {
  width: 100% !important;
  min-width: 100% !important;
  table-layout: fixed !important;
  border-collapse: separate !important;
  border-spacing: 0 !important;
}

.ant-picker-dropdown thead,
.ant-picker-dropdown tbody {
  width: 100% !important;
  display: table-row-group !important;
}

.ant-picker-dropdown tr {
  width: 100% !important;
  display: table-row !important;
}

.ant-picker-dropdown th,
.ant-picker-dropdown td {
  width: 14.285714% !important;
  min-width: 14.285714% !important;
  max-width: 14.285714% !important;
  display: table-cell !important;
  border: none !important;
  padding: 4px !important;
  margin: 0 !important;
  text-align: center !important;
  box-sizing: border-box !important;
}

.ant-picker-dropdown th {
  background-color: transparent !important;
}

.ant-picker-dropdown tr:nth-child(2n) {
  background-color: transparent !important;
}

/* =========================== */
/* Ant Design 表格修复 - 覆盖 markdown.css */
/* =========================== */
.ant-table-wrapper table {
  border-collapse: separate !important;
  border-spacing: 0 !important;
  border: 1px solid #f5a5a5 !important;
  border-radius: 8px !important;
  overflow: hidden !important;
}

.ant-table-wrapper th {
  border: none !important;
  border-bottom: 2px solid #74a7eb !important;
  background: #7cb4ec !important;
  padding: 12px 16px !important;
  font-weight: 600 !important;
}

.ant-table-wrapper td {
  border: none !important;
  border-bottom: 1px solid #d9fc3c !important;
  background: rgb(232, 243, 217) !important;
  padding: 12px 16px !important;
}

.ant-table-wrapper tr {
  border: none !important;
}

.ant-table-wrapper tbody tr:nth-child(2n) {
  background: #d9e6cb !important;
}

.ant-table-wrapper tbody tr:last-child td {
  border-bottom: none !important;
}
</style> 