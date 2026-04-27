<script setup>
import { computed, onMounted, reactive, ref, watch, onUnmounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { api } from '../api/client'
import { useToast } from '../composables/useToast'
import MetroIcon from '../components/MetroIcon.vue'
import BaseModal from '../components/BaseModal.vue'
import EventPlannedView from './EventPlannedView.vue'
import EventStartedView from './EventStartedView.vue'

const props = defineProps({
  id: { type: String, required: true },
})

const route = useRoute()
const toast = useToast()
const eventId = computed(() => Number(props.id || route.params.id))

// Core state
const event = ref(null)
const place = ref(null)
const metroStations = ref([])
const withoutPairs = ref([])
const loading = ref(true)
const error = ref('')
const eventParticipants = ref([])
const allMembers = ref([])
const startResult = ref(null)

// Edit modal state
const showEditModal = ref(false)
const editForm = reactive({
  topic: '',
  scheduled_at: '',
  notify_at: '',
  notifications_enabled: true
})

// Timer and session state
const currentTime = ref(new Date())
const timerInterval = ref(null)
const sessionActivities = ref([])

// Computed properties for metro line
function getMetroLineForArea(areaName) {
  if (!areaName) return null
  const station = metroStations.value.find(s => 
    s.name.toLowerCase().includes(areaName.toLowerCase()) ||
    areaName.toLowerCase().includes(s.name.toLowerCase())
  )
  return station ? station.line : null
}

// Session timer computations
const sections = computed(() => {
  // If event is started/completed and we have session activities, use those durations
  if ((event.value?.status === 'started' || event.value?.status === 'completed') && sessionActivities.value.length > 0) {
    return sessionActivities.value.map(activity => ({
      name: activity.name,
      duration: activity.duration_minutes * 60
    }))
  }
  
  // Default durations for planned events
  return [
    { name: 'One-to-One Talks', duration: 15 * 60 },
    { name: 'Small Group Discussions', duration: 30 * 60 },
    { name: 'Alias Game', duration: 15 * 60 }
  ]
})

const elapsedSeconds = computed(() => {
  if (!event.value?.status || event.value.status !== 'started' || !event.value?.started_at) {
    return 0
  }
  const startTime = new Date(event.value.started_at)
  const elapsed = Math.floor((currentTime.value - startTime) / 1000)
  return elapsed
})

const sectionStates = computed(() => {
  const elapsed = elapsedSeconds.value
  let totalDuration = 0
  const states = []
  
  for (let i = 0; i < sections.value.length; i++) {
    const sectionStartTime = totalDuration
    const sectionEndTime = totalDuration + sections.value[i].duration
    
    let status = 'pending'
    let timeRemaining = sections.value[i].duration
    
    if (elapsed >= sectionEndTime) {
      status = 'completed'
      timeRemaining = 0
    } else if (elapsed >= sectionStartTime && elapsed < sectionEndTime) {
      status = 'active'
      timeRemaining = sectionEndTime - elapsed
    }
    
    states.push({
      ...sections.value[i],
      status,
      timeRemaining
    })
    
    totalDuration = sectionEndTime
  }
  
  return states
})

const currentSection = computed(() => {
  const active = sectionStates.value.find(s => s.status === 'active')
  return active ? active.name : (elapsedSeconds.value >= sections.value.reduce((sum, s) => sum + s.duration, 0) ? 'Session Complete' : null)
})

const sectionTimeRemaining = computed(() => {
  const active = sectionStates.value.find(s => s.status === 'active')
  return active ? active.timeRemaining : 0
})

const participantsByTable = computed(() => {
  if (!eventParticipants.value.length) return {}
  
  const grouped = {}
  eventParticipants.value
    .filter(participant => participant.attended && participant.table_id)
    .forEach(participant => {
      const tableId = participant.table_id
      if (!grouped[tableId]) {
        grouped[tableId] = []
      }
      grouped[tableId].push(participant)
    })
  
  return grouped
})

// Utility functions
function fmt(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleString()
}

function toLocalDateTime(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  const offset = d.getTimezoneOffset()
  const adjustedDate = new Date(d.getTime() - (offset * 60 * 1000))
  return adjustedDate.toISOString().slice(0, 16)
}

function updateTimer() {
  currentTime.value = new Date()
}

// Modal handlers
function openEditModal() {
  editForm.topic = event.value.topic
  editForm.scheduled_at = toLocalDateTime(event.value.scheduled_at)
  editForm.notify_at = event.value.notify_at ? toLocalDateTime(event.value.notify_at) : ''
  editForm.notifications_enabled = !!event.value.notify_at
  
  // If notifications are enabled but no time is set, calculate default
  if (editForm.notifications_enabled && !editForm.notify_at) {
    const scheduled = new Date(editForm.scheduled_at)
    const defaultNotify = new Date(scheduled.getTime() - 8 * 60 * 60 * 1000)
    editForm.notify_at = toLocalDateTime(defaultNotify.toISOString())
  }
  
  showEditModal.value = true
}

function closeEditModal() {
  showEditModal.value = false
}

function autoFillNotificationTime() {
  // Auto-fill notification time if empty when user focuses on the field
  if (!editForm.notify_at && editForm.scheduled_at) {
    const scheduled = new Date(editForm.scheduled_at)
    const defaultNotify = new Date(scheduled.getTime() - 8 * 60 * 60 * 1000)
    editForm.notify_at = toLocalDateTime(defaultNotify.toISOString())
  }
}

async function saveEdit() {
  try {
    const updateData = {
      topic: editForm.topic,
      scheduled_at: new Date(editForm.scheduled_at).toISOString()
    }
    if (editForm.notifications_enabled) {
      if (editForm.notify_at) {
        updateData.notify_at = new Date(editForm.notify_at).toISOString()
      } else {
        // Auto-calculate 8 hours before event if notifications enabled but no time specified
        const scheduled = new Date(editForm.scheduled_at)
        const defaultNotify = new Date(scheduled.getTime() - 8 * 60 * 60 * 1000)
        updateData.notify_at = defaultNotify.toISOString()
      }
    } else {
      updateData.notify_at = ""  // Send empty string to disable notifications
    }
    await api.updateEvent(eventId.value, updateData)
    toast.success('Event updated successfully')
    closeEditModal()
    await load()
  } catch (e) {
    toast.error('Failed to update event')
  }
}

// Data loading
async function load() {
  loading.value = true
  error.value = ''
  startResult.value = null
  try {
    const [ev, metroList, memberList, late] = await Promise.all([
      api.getEvent(eventId.value),
      api.getMetroList(),
      api.getMembers(),
      api.getWithoutPairs()
    ])
    event.value = ev
    metroStations.value = metroList
    allMembers.value = memberList
    withoutPairs.value = late
    
    if (ev?.place_id) {
      place.value = await api.getPlace(ev.place_id)
    } else {
      place.value = null
    }

    // Load participants for all event statuses
    if (ev?.status === 'planned' || ev?.status === 'started' || ev?.status === 'completed') {
      await loadParticipants()
    }
    
    // Load session activities for started/completed events
    if (ev?.status === 'started' || ev?.status === 'completed') {
      try {
        sessionActivities.value = await api.getSessionActivities(eventId.value)
      } catch (e) {
        console.error('Failed to load session activities:', e)
        sessionActivities.value = []
      }
    }
    
  } catch (e) {
    error.value = e.message
    event.value = null
  } finally {
    loading.value = false
  }
}

async function loadParticipants() {
  try {
    const parts = await api.getEventParticipants(eventId.value)
    eventParticipants.value = parts
  } catch (e) {
    console.error('Failed to load participants:', e)
  }
}

// Event handlers from child components
function handleParticipantAdded(participant) {
  eventParticipants.value.push(participant)
}

function handleParticipantRemoved(memberId) {
  const index = eventParticipants.value.findIndex(p => p.member_id === memberId)
  if (index !== -1) {
    eventParticipants.value.splice(index, 1)
  }
}

function handleAttendanceToggled({ participant, attended }) {
  participant.attended = attended
}

function handleStartEvent(result) {
  startResult.value = result
  load() // Reload to get updated event status
}

function handleLateParticipantAdded(memberData) {
  // Reload without pairs and participants
  load()
}

function handleMemberMoved({ memberId, tableId }) {
  // Remove from withoutPairs
  const memberIndex = withoutPairs.value.findIndex(member => member.member_id === memberId)
  if (memberIndex !== -1) {
    const member = withoutPairs.value[memberIndex]
    withoutPairs.value.splice(memberIndex, 1)
    
    // Update or add to eventParticipants
    const eventParticipantIndex = eventParticipants.value.findIndex(p => p.member_id === memberId)
    if (eventParticipantIndex !== -1) {
      eventParticipants.value[eventParticipantIndex].table_id = parseInt(tableId)
    } else {
      eventParticipants.value.push({
        member_id: member.member_id,
        name: member.name,
        telegram: member.telegram,
        attended: true,
        table_id: parseInt(tableId)
      })
    }
  }
}

// Watchers
watch(currentSection, async (newSection, oldSection) => {
  if (newSection && newSection !== oldSection && event.value?.status === 'started') {
    // Refresh participants when activity changes (new table assignments)
    await loadParticipants()
  }
})

watch(
  () => props.id,
  () => {
    load()
  },
)

// Lifecycle
onMounted(() => {
  load()
  updateTimer()
  timerInterval.value = setInterval(updateTimer, 1000)
})

onUnmounted(() => {
  if (timerInterval.value) {
    clearInterval(timerInterval.value)
  }
})
</script>

<template>
  <div>
    <nav aria-label="breadcrumb" class="mb-3">
      <ol class="breadcrumb mb-0">
        <li class="breadcrumb-item"><RouterLink to="/">Events</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">Event #{{ eventId }}</li>
      </ol>
    </nav>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-if="loading" class="text-muted">Loading…</div>

    <template v-else-if="event">
      <!-- Event Header -->
      <div class="mb-3">
        <div class="d-flex justify-content-between align-items-start gap-2 mb-2">
          <div>
            <h1 class="h3 mb-1">
              {{ event.topic || 'Untitled' }}
              <button v-if="event.status === 'planned'" class="btn btn-sm btn-outline-secondary ms-2" @click="openEditModal">
                <i class="bi bi-pencil"></i> Edit
              </button>
            </h1>
            <p class="text-muted mb-0">{{ fmt(event.scheduled_at) }}</p>
            <p v-if="place" class="mb-0">
              <i class="bi bi-building me-1"></i>
              <span class="fw-medium">{{ place.name }}</span>
              <span v-if="place.metro_area" class="text-muted d-inline-flex align-items-center">
                <span class="mx-1">·</span>
                <MetroIcon :size="20" :line="getMetroLineForArea(place.metro_area)" class="me-1" />
                {{ place.metro_area }}
              </span>
              <a v-if="place.map_url" :href="place.map_url" class="ms-2 small" target="_blank" rel="noopener">
                <i class="bi bi-map"></i> Map
              </a>
            </p>
          </div>
          
          <span
            class="badge fs-6 flex-shrink-0"
            :class="{
              'text-bg-secondary': event.status === 'planned',
              'text-bg-primary': event.status === 'started',
              'text-bg-success': event.status === 'completed',
            }"
          >
            {{ event.status }}
          </span>
        </div>
      </div>

      <!-- Start Result Display -->
      <div v-if="startResult" class="alert alert-success">
        <strong>Session started.</strong>
        <details class="mt-2 small">
          <summary>Response payload</summary>
          <pre class="mb-0 mt-2 p-2 bg-light rounded small overflow-auto" style="max-height: 16rem">{{ JSON.stringify(startResult, null, 2) }}</pre>
        </details>
      </div>

      <!-- Event Status Specific Views -->
      <EventPlannedView
        v-if="event.status === 'planned'"
        :event="event"
        :event-participants="eventParticipants"
        :all-members="allMembers"
        :event-id="eventId"
        @participant-added="handleParticipantAdded"
        @participant-removed="handleParticipantRemoved"
        @attendance-toggled="handleAttendanceToggled"
        @start-event="handleStartEvent"
      />

      <EventStartedView
        v-else-if="event.status === 'started'"
        :event="event"
        :current-section="currentSection"
        :section-time-remaining="sectionTimeRemaining"
        :section-states="sectionStates"
        :participants-by-table="participantsByTable"
        :without-pairs="withoutPairs"
        :all-members="allMembers"
        :event-id="eventId"
        :event-participants="eventParticipants"
        @late-participant-added="handleLateParticipantAdded"
        @member-moved="handleMemberMoved"
      />

      <!-- Completed Events -->
      <template v-else>
        <div class="mt-4">
          <h2 class="h5">Event Completed</h2>
          
          <!-- Participants Attendance -->
          <div class="card shadow-sm">
            <div class="card-header bg-white">
              <h6 class="mb-0">Participants Attendance</h6>
            </div>
            <div class="card-body">
              <div v-if="eventParticipants.length === 0" class="text-muted small">
                No participants recorded
              </div>
              <div v-else class="list-group list-group-flush">
                <div v-for="participant in eventParticipants" :key="participant.member_id" class="list-group-item px-0 d-flex align-items-center">
                  <div class="form-check me-3">
                    <input 
                      class="form-check-input" 
                      type="checkbox" 
                      :checked="participant.attended"
                      disabled
                    />
                  </div>
                  <div class="flex-grow-1">
                    <div class="fw-medium">{{ participant.name }}</div>
                    <div class="small text-muted">{{ participant.telegram }}</div>
                  </div>
                  <div v-if="participant.attended" class="badge bg-success">
                    <i class="bi bi-check-circle me-1"></i>Attended
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </template>

    <p v-else class="text-muted">Event not found.</p>

    <!-- Edit Modal -->
    <BaseModal
      :show="showEditModal"
      title="Edit Event"
      submit-text="Save Changes"
      @close="closeEditModal"
      @submit="saveEdit"
    >
      <div class="mb-3">
        <label class="form-label">Topic</label>
        <input v-model="editForm.topic" class="form-control" required />
      </div>
      <div class="mb-3">
        <label class="form-label">Date and Time</label>
        <input v-model="editForm.scheduled_at" type="datetime-local" class="form-control" required />
      </div>
      <div class="mb-3">
        <div class="form-check mb-2">
          <input
            id="edit-notifications-enabled"
            v-model="editForm.notifications_enabled"
            class="form-check-input"
            type="checkbox"
          />
          <label class="form-check-label" for="edit-notifications-enabled">
            Enable notifications
          </label>
        </div>
        <div v-if="editForm.notifications_enabled">
          <label class="form-label">Notification Time</label>
          <input 
            v-model="editForm.notify_at" 
            type="datetime-local" 
            class="form-control"
            @focus="autoFillNotificationTime" 
          />
          <div class="form-text">When to send notifications to participants (default: 8 hours before event)</div>
        </div>
      </div>
    </BaseModal>
  </div>
</template>