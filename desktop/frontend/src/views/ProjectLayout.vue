<script lang="ts" setup>
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useProjectStore } from '../stores/project'
import { useI18n } from 'vue-i18n'
import Sidebar from '../components/layout/Sidebar.vue'

const { t } = useI18n()
const projectStore = useProjectStore()
const route = useRoute()
const router = useRouter()

onMounted(() => {
  if (!projectStore.current) router.replace('/')
})

const pageTitle = computed(() => {
  const name = route.name as string
  const keyMap: Record<string, string> = {
    stats: 'stats',
    status: 'sessionStatus',
    issues: 'issueTracking',
    tasks: 'taskManagement',
    'project-memories': 'projectMemories',
    'user-memories': 'globalMemories',
    tags: 'tagManagement',
    maintenance: 'maintenance',
    settings: 'settings',
  }
  return t(keyMap[name] || name || 'stats')
})

const projectName = computed(() => {
  const dir = projectStore.current
  return dir ? dir.split('/').filter(Boolean).pop() || dir : ''
})
</script>

<template>
  <Sidebar />
  <div class="main">
    <header class="main-header">
      <div class="main-header-left">
        <span class="main-header-title">{{ pageTitle }}</span>
        <span class="main-header-project">{{ projectName }}</span>
      </div>
      <div class="main-header-actions">
        <slot name="header-actions" />
      </div>
    </header>
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.main-header {
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  background: hsl(240 5% 12% / 0.8);
  backdrop-filter: blur(12px);
}
[data-theme="light"] .main-header {
  background: hsl(0 0% 98% / 0.85);
}
.main-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.main-header-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
.main-header-project {
  font-size: 13px;
  color: var(--text-muted);
  padding: 2px 10px;
  background: var(--bg-muted);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}
.main-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}
.main-content {
  flex: 1;
  padding: 20px 24px;
  overflow-y: auto;
}
</style>
