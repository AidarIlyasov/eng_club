<script setup>
import { ref, computed, reactive } from 'vue'
import BaseModal from './BaseModal.vue'
import MemberSearch from './MemberSearch.vue'
import { api } from '../api/client'
import { useToast } from '../composables/useToast'

const props = defineProps({
  withoutPairs: { type: Array, required: true },
  allMembers: { type: Array, required: true },
  participantsByTable: { type: Object, default: () => ({}) },
  eventId: { type: Number, required: true }
})

const emit = defineEmits(['late-participant-added', 'member-moved'])

const toast = useToast()
const showLateParticipantModal = ref(false)
const showMoveToTableModal = ref(false)
const memberSearchRef = ref(null)
const submitting = ref(false)

const moveToTableForm = reactive({
  memberId: null,
  tableId: null,
  newTableId: null
})

const existingTableIds = computed(() => Object.keys(props.participantsByTable))

const isMoveFormValid = computed(() => {
  return (moveToTableForm.tableId || moveToTableForm.newTableId) && moveToTableForm.memberId
})

function openLateParticipantModal() {
  showLateParticipantModal.value = true
}

function closeLateParticipantModal() {
  if (memberSearchRef.value) {
    memberSearchRef.value.clearSearch()
  }
  showLateParticipantModal.value = false
}

async function handleMemberSelected(member) {
  await addLateParticipant(member, false)
}

async function handleNewMemberAdded(memberData) {
  await addLateParticipant(memberData, true)
}

async function addLateParticipant(memberData, isNew = false) {
  submitting.value = true
  
  try {
    if (isNew) {
      // For new members, use the addWithoutPairs API which creates member and adds to without_pairs
      await api.addWithoutPairs({
        name: memberData.name,
        telegram: memberData.telegram
      })
    } else {
      // For existing members, add them to without_pairs directly
      await api.addWithoutPairs({
        id: memberData.member_id,
        name: memberData.name,
        telegram: memberData.telegram
      })
    }
    
    emit('late-participant-added', memberData)
    toast.success('Late participant added')
    closeLateParticipantModal()
  } catch (e) {
    toast.error('Failed to add late participant')
  } finally {
    submitting.value = false
  }
}

function openMoveToTableModal(memberId) {
  moveToTableForm.memberId = memberId
  moveToTableForm.tableId = null
  moveToTableForm.newTableId = null
  showMoveToTableModal.value = true
}

function closeMoveToTableModal() {
  moveToTableForm.memberId = null
  moveToTableForm.tableId = null
  moveToTableForm.newTableId = null
  showMoveToTableModal.value = false
}

async function handleMoveToTable() {
  const targetTableId = parseInt(moveToTableForm.newTableId || moveToTableForm.tableId)
  
  if (!targetTableId || !moveToTableForm.memberId) {
    toast.error('Please select an existing table or enter a new table number')
    return
  }

  try {
    // Call backend API to move member
    await api.moveToTable(props.eventId, moveToTableForm.memberId, targetTableId)
    
    emit('member-moved', {
      memberId: moveToTableForm.memberId,
      tableId: targetTableId
    })
    
    toast.success(`Member moved to table ${targetTableId} successfully`)
    closeMoveToTableModal()
  } catch (e) {
    toast.error('Failed to move member to table')
    console.error(e)
  }
}
</script>

<template>
  <div class="card shadow-sm">
    <div class="card-header bg-white d-flex justify-content-between align-items-center">
      <h6 class="mb-0">Session Management</h6>
      <button class="btn btn-sm btn-outline-primary" @click="openLateParticipantModal">
        <i class="bi bi-person-plus"></i> Add Late Participant
      </button>
    </div>
    <div class="card-body">
      <div class="row">
        <div class="col-5 col-md-5">
          <label class="form-label small text-muted mb-1">Name</label>
        </div>
        <div class="col-5 col-md-5">
          <label class="form-label small text-muted mb-1">Telegram</label>
        </div>
        <div class="col-2 col-md-2 text-center">
          <label class="form-label small text-muted mb-1">Options</label>
        </div>
      </div>
      
      <!-- Without pairs members -->
      <div v-for="member in withoutPairs" :key="member.member_id" class="row g-2 align-items-center mb-2">
        <div class="col-5 col-md-5">
          <div class="form-control form-control-sm bg-light">{{ member.name }}</div>
        </div>
        <div class="col-5 col-md-5">
          <div class="form-control form-control-sm bg-light">{{ member.telegram }}</div>
        </div>
        <div class="col-2 col-md-2 text-center text-md-end">
          <button 
            type="button" 
            class="btn btn-outline-primary btn-sm" 
            @click="openMoveToTableModal(member.member_id)"
          >
            <i class="bi bi-arrow-right d-md-none"></i>
            <span class="d-none d-md-inline">Move to Table</span>
          </button>
        </div>
      </div>
      
      <div v-if="withoutPairs.length === 0" class="text-muted text-center py-3">
        <i class="bi bi-people display-4 text-muted mb-2"></i>
        <p>No participants without table assignments</p>
      </div>
    </div>
    
    <!-- Add Late Participant Modal -->
    <BaseModal
      :show="showLateParticipantModal"
      title="Add Late Participant"
      hide-footer
      @close="closeLateParticipantModal"
    >
      <MemberSearch
        ref="memberSearchRef"
        :all-members="allMembers"
        :exclude-members="withoutPairs"
        :show-add-new="true"
        placeholder="Search member name or telegram..."
        @member-selected="handleMemberSelected"
        @new-member-added="handleNewMemberAdded"
      />
      
      <div v-if="submitting" class="text-center mt-3">
        <div class="spinner-border spinner-border-sm" role="status">
          <span class="visually-hidden">Adding participant...</span>
        </div>
      </div>
    </BaseModal>
    
    <!-- Move to Table Modal -->
    <BaseModal
      :show="showMoveToTableModal"
      title="Move to Table"
      submit-text="Move to Table"
      :submit-disabled="!isMoveFormValid"
      @close="closeMoveToTableModal"
      @submit="handleMoveToTable"
    >
      <div class="mb-3">
        <label class="form-label">Select Existing Table</label>
        <select v-model="moveToTableForm.tableId" class="form-select">
          <option value="">Choose table...</option>
          <option v-for="tableId in existingTableIds" :key="tableId" :value="tableId">
            Table {{ tableId }}
          </option>
        </select>
      </div>
      
      <div class="d-flex align-items-center">
        <hr class="flex-grow-1">
        <span class="mx-3 text-muted">Or create a new table</span>
        <hr class="flex-grow-1">
      </div>
      
      <div class="mb-3">
        <label class="form-label">New Table ID</label>
        <input 
          v-model="moveToTableForm.newTableId" 
          class="form-control" 
          type="number" 
          placeholder="Enter new table number..."
          min="1"
        />
      </div>
    </BaseModal>
  </div>
</template>