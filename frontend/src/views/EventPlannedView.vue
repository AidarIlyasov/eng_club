<script setup>
import { ref } from 'vue'
import MemberSearch from '../components/MemberSearch.vue'
import { api } from '../api/client'
import { useToast } from '../composables/useToast'

const props = defineProps({
  event: { type: Object, required: true },
  eventParticipants: { type: Array, required: true },
  allMembers: { type: Array, required: true },
  eventId: { type: Number, required: true }
})

const emit = defineEmits(['participant-added', 'participant-removed', 'attendance-toggled', 'start-event'])

const toast = useToast()
const starting = ref(false)

async function handleMemberSelected(member) {
  try {
    const response = await api.addParticipantToEvent(props.eventId, { 
      name: member.name, 
      telegram: member.telegram 
    })
    
    const participant = {
      member_id: member.member_id,
      name: member.name,
      telegram: member.telegram,
      attended: false,
      table_id: null
    }
    
    emit('participant-added', participant)
    toast.success(`${member.name} added to participants`)
  } catch (e) {
    toast.error('Failed to add participant')
  }
}

async function handleNewMemberAdded(memberData) {
  try {
    const response = await api.addParticipantToEvent(props.eventId, { 
      name: memberData.name, 
      telegram: memberData.telegram 
    })
    
    const participant = {
      member_id: response.member_id,
      name: memberData.name,
      telegram: memberData.telegram,
      attended: false,
      table_id: null
    }
    
    emit('participant-added', participant)
    toast.success('Participant added')
  } catch (e) {
    toast.error('Failed to add participant')
  }
}

async function toggleAttendance(participant) {
  try {
    await api.markAttendance(props.eventId, participant.member_id, !participant.attended)
    emit('attendance-toggled', { participant, attended: !participant.attended })
    toast.success(`Marked ${participant.name} as ${!participant.attended ? 'attended' : 'not attended'}`)
  } catch (e) {
    toast.error('Failed to update attendance')
  }
}

async function removeParticipant(memberId) {
  if (!confirm('Are you sure you want to remove this participant?')) {
    return
  }
  
  try {
    await api.removeParticipantFromEvent(props.eventId, memberId)
    emit('participant-removed', memberId)
    toast.success('Participant removed')
  } catch (e) {
    toast.error('Failed to remove participant')
  }
}

async function startEvent() {
  // Validate before starting
  if (!props.event || props.event.status !== 'planned') {
    toast.error('Event must be in planned status to start')
    return
  }

  if (props.event.type === 'single_activity') {
    if (props.eventParticipants.length === 0) {
      toast.error('At least 1 participant is required for single activity events')
      return
    }
  } else {
    if (props.eventParticipants.length < 3) {
      toast.error('At least 3 participants are required for many activities events to enable table shuffling')
      return
    }
  }

  starting.value = true
  
  const body = {
    participants: props.eventParticipants.map(p => ({
      name: p.name,
      telegram: p.telegram,
      member_id: p.member_id
    }))
  }
  
  try {
    const res = await api.startEvent(props.eventId, body)
    emit('start-event', res)
    toast.success('Session started successfully!')
  } catch (e) {
    toast.error('Failed to start session')
  } finally {
    starting.value = false
  }
}
</script>

<template>
  <div>
    <h2 class="h5 mt-4">Participants</h2>
    <p class="text-muted small">Add one row per person. Telegram handles are sent to the server as you type them (with or without <code>@</code>).</p>

    <div class="card shadow-sm">
      <div class="card-body">
        <!-- Existing participants -->
        <div v-for="(participant, i) in eventParticipants" :key="participant.member_id" class="row g-2 align-items-center mb-2">
          <div class="col-5 col-md-5">
            <label v-if="i === 0" class="form-label small text-muted mb-1">Name</label>
            <div class="form-control form-control-sm bg-light">{{ participant.name }}</div>
          </div>
          <div class="col-5 col-md-5">
            <label v-if="i === 0" class="form-label small text-muted mb-1">Telegram</label>
            <div class="form-control form-control-sm bg-light">{{ participant.telegram }}</div>
          </div>
          <div class="col-1 col-md-1 text-center">
            <label v-if="i === 0" class="form-label small text-muted mb-1">Attended</label>
            <div class="form-check d-flex justify-content-center">
              <input 
                :checked="participant.attended"
                class="form-check-input" 
                type="checkbox" 
                :id="`participant-${participant.member_id}`"
                @change="toggleAttendance(participant)"
              />
            </div>
          </div>
          <div class="col-1 col-md-1 text-center text-md-end">
            <button 
              type="button" 
              class="btn btn-outline-danger btn-sm" 
              @click="removeParticipant(participant.member_id)"
            >
              <i class="bi bi-trash d-md-none"></i>
              <span class="d-none d-md-inline">Remove</span>
            </button>
          </div>
        </div>
        
        <!-- Member Search Component -->
        <MemberSearch
          :all-members="allMembers"
          :exclude-members="eventParticipants"
          :show-add-new="true"
          @member-selected="handleMemberSelected"
          @new-member-added="handleNewMemberAdded"
        />
      </div>
    </div>

    <button 
      type="button" 
      class="btn btn-success mt-4" 
      :disabled="starting" 
      @click="startEvent"
    >
      {{ starting ? 'Starting…' : 'Start session' }}
    </button>
  </div>
</template>