<template>
  <div v-if="show" class="modal" @click="handleBackdropClick">
    <div class="modal-content" @click.stop>
      <span class="close" @click="$emit('update:show', false)">&times;</span>
      <h3>{{ title }}</h3>
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  show: boolean
  title: string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
}>()

const handleBackdropClick = () => {
  emit('update:show', false)
}
</script>

<style scoped>
.modal {
  display: block;
  position: fixed;
  z-index: 1000;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0,0,0,0.5);
}

.modal-content {
  background-color: white;
  margin: 5% auto;
  padding: 20px;
  border-radius: 8px;
  width: 80%;
  max-width: 600px;
}

.close {
  color: #aaa;
  float: right;
  font-size: 28px;
  font-weight: bold;
  cursor: pointer;
}

.close:hover { 
  color: #000; 
}
</style> 