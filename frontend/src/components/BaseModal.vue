<script setup>
const props = defineProps({
  show: { type: Boolean, default: false },
  title: { type: String, required: true },
  size: { type: String, default: 'modal-dialog-centered' }, // can be 'modal-lg', 'modal-sm', etc.
  hideFooter: { type: Boolean, default: false },
  submitText: { type: String, default: 'Save' },
  cancelText: { type: String, default: 'Cancel' },
  submitDisabled: { type: Boolean, default: false },
  submitVariant: { type: String, default: 'primary' } // primary, success, danger, etc.
})

const emit = defineEmits(['close', 'submit'])

function handleSubmit() {
  emit('submit')
}

function handleClose() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div 
      v-if="show" 
      class="modal fade show d-block" 
      tabindex="-1" 
      style="background: rgba(0,0,0,0.5);"
      @click.self="handleClose"
    >
      <div class="modal-dialog" :class="size">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ title }}</h5>
            <button type="button" class="btn-close" @click="handleClose"></button>
          </div>
          
          <form @submit.prevent="handleSubmit">
            <div class="modal-body">
              <slot />
            </div>
            
            <div v-if="!hideFooter" class="modal-footer">
              <slot name="footer">
                <button type="button" class="btn btn-secondary" @click="handleClose">
                  {{ cancelText }}
                </button>
                <button 
                  type="submit" 
                  class="btn" 
                  :class="`btn-${submitVariant}`"
                  :disabled="submitDisabled"
                >
                  {{ submitText }}
                </button>
              </slot>
            </div>
          </form>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal {
  animation: fadeIn 0.2s;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>