<template>
  <el-switch
    v-model="isDark"
    inline-prompt
    :active-icon="Moon"
    :inactive-icon="Sunny"
    @change="toggleDark"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Moon, Sunny } from '@element-plus/icons-vue'
import { useUIStore } from '@/stores/useUIStore'

const uiStore = useUIStore()

const isDark = computed({
  get: () => uiStore.isDarkMode,
  set: () => uiStore.toggleDarkMode(),
})

const toggleDark = () => {
  uiStore.toggleDarkMode()
}
</script>

<style scoped>
/* 样式由 Element Plus 内置 */
</style>
