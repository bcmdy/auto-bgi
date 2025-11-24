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
</style> 