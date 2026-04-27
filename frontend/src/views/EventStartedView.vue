<script setup>
import SessionInProgress from '../components/SessionInProgress.vue'
import TableAssignments from '../components/TableAssignments.vue'
import SessionManagement from '../components/SessionManagement.vue'

const props = defineProps({
  event: { type: Object, required: true },
  currentSection: { type: String, default: null },
  sectionTimeRemaining: { type: Number, default: 0 },
  sectionStates: { type: Array, required: true },
  participantsByTable: { type: Object, required: true },
  withoutPairs: { type: Array, required: true },
  allMembers: { type: Array, required: true },
  eventId: { type: Number, required: true },
  eventParticipants: { type: Array, required: true }
})

const emit = defineEmits(['late-participant-added', 'member-moved'])

function handleLateParticipantAdded(memberData) {
  emit('late-participant-added', memberData)
}

function handleMemberMoved(moveData) {
  emit('member-moved', moveData)
}

// Show SessionInProgress only for many_activities events
const showSessionInProgress = props.event.type === 'many_activities'
</script>

<template>
  <div>
    <!-- Session in Progress (only for many_activities) -->
    <SessionInProgress
      v-if="showSessionInProgress"
      :current-section="currentSection"
      :section-time-remaining="sectionTimeRemaining"
      :section-states="sectionStates"
    />

    <!-- Single Activity Simple Display -->
    <div v-else class="mt-4">
      <h2 class="h5">Single Activity Session in Progress</h2>
      <div class="alert alert-info">
        <i class="bi bi-info-circle me-2"></i>
        <strong>{{ event.topic }}</strong> is currently running.
        <div class="small text-muted mt-1">
          <span v-if="event.finished_at">
            Scheduled to finish at {{ new Date(event.finished_at).toLocaleString() }}
          </span>
        </div>
      </div>
      
      <!-- Participants List for Single Activity -->
      <div class="card shadow-sm mt-3 mb-3">
        <div class="card-header bg-white">
          <h6 class="mb-0">Participants ({{ eventParticipants.length }})</h6>
        </div>
        <div class="card-body">
          <div v-if="eventParticipants.length === 0" class="text-muted small">
            No participants recorded
          </div>
          <div v-else class="d-flex flex-wrap gap-2">
            <div 
              v-for="participant in eventParticipants" 
              :key="participant.member_id" 
              class="badge bg-light text-dark border"
            >
              <div class="fw-medium">{{ participant.name }}</div>
              <div class="small text-muted">{{ participant.telegram }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Table Assignments (only for many_activities) -->
    <TableAssignments 
      v-if="showSessionInProgress"
      :participants-by-table="participantsByTable"
    />

    <!-- Session Management -->
    <SessionManagement
      :without-pairs="withoutPairs"
      :all-members="allMembers"
      :participants-by-table="participantsByTable"
      :event-id="eventId"
      @late-participant-added="handleLateParticipantAdded"
      @member-moved="handleMemberMoved"
    />
  </div>
</template>