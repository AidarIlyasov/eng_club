<script setup>
import { computed, ref } from 'vue'
import { api } from '../api/client'
import { useToast } from '../composables/useToast'

const props = defineProps({
  allMembers: { type: Array, required: true },
  excludeMembers: { type: Array, default: () => [] }, // members to exclude from search
  showAddNew: { type: Boolean, default: true },
  placeholder: { type: String, default: 'Type name or telegram to search...' },
  maxResults: { type: Number, default: 5 }
})

const emit = defineEmits(['member-selected', 'new-member-added'])

const toast = useToast()
const searchQuery = ref('')
const newMember = ref({ name: '', telegram: '' })

// Filter members for search
const filteredMembers = computed(() => {
  if (!searchQuery.value.trim()) {
    return []
  }
  
  const search = searchQuery.value.toLowerCase()
  const excludeIds = new Set(props.excludeMembers.map(m => m.member_id || m.id))
  
  return props.allMembers
    .filter(member => 
      !excludeIds.has(member.member_id || member.id) &&
      (member.name.toLowerCase().includes(search) || 
       member.telegram?.toLowerCase().includes(search))
    )
    .slice(0, props.maxResults)
})

function selectMember(member) {
  emit('member-selected', member)
  searchQuery.value = ''
}

function addNewMember() {
  if (!newMember.value.name.trim() || !newMember.value.telegram.trim()) {
    toast.error('Please fill in both name and telegram')
    return
  }
  
  emit('new-member-added', {
    name: newMember.value.name.trim(),
    telegram: newMember.value.telegram.trim()
  })
  
  // Clear the form
  newMember.value.name = ''
  newMember.value.telegram = ''
}

// Expose method to clear search
defineExpose({
  clearSearch: () => {
    searchQuery.value = ''
    newMember.value.name = ''
    newMember.value.telegram = ''
  }
})
</script>

<template>
  <div>
    <!-- Search existing members -->
    <div class="mb-3">
      <label class="form-label small text-muted mb-1">Search Existing Members</label>
      <div class="position-relative">
        <input 
          v-model="searchQuery" 
          class="form-control form-control-sm" 
          :placeholder="placeholder"
        />
        <div 
          v-if="filteredMembers.length > 0" 
          class="position-absolute w-100 bg-white border rounded shadow-sm mt-1" 
          style="z-index: 1000; max-height: 200px; overflow-y: auto;"
        >
          <div 
            v-for="member in filteredMembers" 
            :key="member.member_id || member.id"
            class="px-3 py-2 border-bottom cursor-pointer hover-bg-light"
            @click="selectMember(member)"
            style="cursor: pointer;"
          >
            <div class="fw-medium">{{ member.name }}</div>
            <div class="small text-muted">{{ member.telegram }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Or separator and new member form -->
    <template v-if="showAddNew">
      <div class="d-flex align-items-center">
        <hr class="flex-grow-1">
        <span class="mx-3 text-muted small">Or add a new one</span>
        <hr class="flex-grow-1">
      </div>

      <div class="row g-2 align-items-end">
        <div class="col-5">
          <label class="form-label small text-muted mb-1">Name</label>
          <input v-model="newMember.name" class="form-control form-control-sm" placeholder="Alice" />
        </div>
        <div class="col-5">
          <label class="form-label small text-muted mb-1">Telegram</label>
          <input v-model="newMember.telegram" class="form-control form-control-sm" placeholder="@alice" />
        </div>
        <div class="col-2">
          <button 
            type="button" 
            class="btn btn-outline-success btn-sm w-100" 
            @click="addNewMember"
            :disabled="!newMember.name.trim() || !newMember.telegram.trim()"
          >
            <i class="bi bi-plus d-md-none"></i>
            <span class="d-none d-md-inline">Add</span>
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.hover-bg-light:hover {
  background-color: #f8f9fa !important;
}
</style>