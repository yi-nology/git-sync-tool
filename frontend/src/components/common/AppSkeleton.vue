<template>
  <div class="app-skeleton">
    <el-skeleton :loading="loading" :animated="animated" :rows="rows" :throttle="throttle">
      <template #default>
        <slot></slot>
      </template>
      <template v-if="$slots.template" #template>
        <slot name="template"></slot>
      </template>
    </el-skeleton>
  </div>
</template>

<script setup lang="ts">
interface Props {
  loading?: boolean
  rows?: number
  animated?: boolean
  throttle?: number
}

const props = withDefaults(defineProps<Props>(), {
  loading: true,
  rows: 5,
  animated: true,
  throttle: 0,
})
</script>

<style scoped>
.app-skeleton {
  width: 100%;
}
</style>
