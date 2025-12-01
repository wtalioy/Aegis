<script setup lang="ts">
import Sidebar from './components/layout/Sidebar.vue'
import TopBar from './components/layout/TopBar.vue'
import StatusFooter from './components/layout/StatusFooter.vue'
</script>

<template>
  <div class="app-layout">
    <Sidebar />
    <div class="main-wrapper">
      <TopBar />
      <main class="main-content">
        <router-view />
      </main>
      <StatusFooter />
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  background: var(--bg-void);
  overflow: hidden;
}

.main-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  height: 100vh;
  overflow: hidden;
}

.main-content {
  flex: 1;
  padding: 24px;
  overflow: auto;
}
</style>
