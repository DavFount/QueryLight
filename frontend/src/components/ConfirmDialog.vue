<script setup lang="ts">
import { onMounted, ref } from 'vue'
defineProps<{ title: string; body: string }>()
const emit = defineEmits<{ (event: 'cancel'): void; (event: 'accept'): void }>()
const dialog = ref<HTMLDialogElement>()
onMounted(() => dialog.value?.showModal())
</script>
<template>
  <dialog
    ref="dialog"
    class="confirm-dialog"
    aria-labelledby="confirm-title"
    @cancel.prevent="emit('cancel')"
  >
    <h2 id="confirm-title">{{ title }}</h2>
    <p>{{ body }}</p>
    <div class="form-actions">
      <button autofocus @click="emit('cancel')">Keep working</button
      ><button class="danger" @click="emit('accept')">Continue</button>
    </div>
  </dialog>
</template>
<style scoped>
dialog {
  color: inherit;
}
dialog::backdrop {
  background: #080e16ba;
  backdrop-filter: blur(3px);
}
</style>
