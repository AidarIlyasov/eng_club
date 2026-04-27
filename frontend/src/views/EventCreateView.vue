<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { api } from '../api/client'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()
const places = ref([])
const loading = ref(true)
const error = ref('')
const saving = ref(false)

const selectedPlaceId = ref(null)
const scheduledLocal = ref('')
const topic = ref('')
const eventType = ref('many_activities')
const notifyLocal = ref('')
const finishedLocal = ref('')
const notificationsEnabled = ref(true)
const activities = ref([
  { name: '', duration: 15, groupSize: 3 }
])

function addActivity() {
  activities.value.push({ name: '', duration: 15, groupSize: 3 })
}

function removeActivity(index) {
  if (activities.value.length > 1) {
    activities.value.splice(index, 1)
  }
}

// Auto-calculate notify_at when scheduled_at changes
function updateNotifyAt() {
  if (scheduledLocal.value && notificationsEnabled.value) {
    const scheduled = new Date(scheduledLocal.value)
    const notify = new Date(scheduled.getTime() - 8 * 60 * 60 * 1000) // 8 hours before
    const offset = notify.getTimezoneOffset()
    const adjustedNotify = new Date(notify.getTime() - (offset * 60 * 1000))
    notifyLocal.value = adjustedNotify.toISOString().slice(0, 16)
  } else if (!notificationsEnabled.value) {
    notifyLocal.value = ''
  }
}

// Helper function to get full image URL
function getImageUrl(filename) {
  if (!filename) return ''
  if (filename.startsWith('http://') || filename.startsWith('https://')) return filename
  return `/api/uploads/${filename}`
}

function toRFC3339(localValue) {
  if (!localValue) return ''
  const d = new Date(localValue)
  if (Number.isNaN(d.getTime())) return ''
  return d.toISOString()
}

const canSubmit = computed(() => {
  const basicRequirements = selectedPlaceId.value && scheduledLocal.value && topic.value.trim() && 
    toRFC3339(scheduledLocal.value)
  
  // For many_activities events, validate activities
  if (eventType.value === 'many_activities') {
    const hasValidActivities = activities.value.length > 0 && 
      activities.value.every(a => a.name.trim() && a.duration > 0 && a.groupSize > 0)
    return basicRequirements && hasValidActivities
  }
  
  // For single_activity events, finished_at is required
  if (eventType.value === 'single_activity') {
    return basicRequirements && finishedLocal.value && toRFC3339(finishedLocal.value)
  }
  
  return basicRequirements
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    places.value = await api.listPlaces()
    if (places.value.length && selectedPlaceId.value == null) {
      selectedPlaceId.value = places.value[0].id
    }
  } catch (e) {
    error.value = e.message
    places.value = []
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!canSubmit.value) return
  saving.value = true
  error.value = ''
  try {
    const eventData = {
      place_id: selectedPlaceId.value,
      scheduled_at: toRFC3339(scheduledLocal.value),
      topic: topic.value.trim(),
      type: eventType.value
    }
    
    // For single_activity events, add finished_at
    if (eventType.value === 'single_activity') {
      eventData.finished_at = toRFC3339(finishedLocal.value)
    }
    
    // Only add activities for many_activities events
    if (eventType.value === 'many_activities') {
      eventData.activities = activities.value.map(a => ({
        name: a.name.trim(),
        duration: a.duration,
        groupSize: a.groupSize
      }))
    }
    if (notificationsEnabled.value) {
      if (notifyLocal.value) {
        eventData.notify_at = toRFC3339(notifyLocal.value)
      } else {
        // Auto-calculate 8 hours before event if notifications enabled but no time specified
        const scheduled = new Date(scheduledLocal.value)
        const defaultNotify = new Date(scheduled.getTime() - 8 * 60 * 60 * 1000)
        eventData.notify_at = defaultNotify.toISOString()
      }
    }
    // If notifications disabled, notify_at is not included (remains undefined)
    const ev = await api.createEvent(eventData)
    toast.success('Event created successfully!')
    router.push({ name: 'event-detail', params: { id: String(ev.id) } })
  } catch (e) {
    error.value = e.message
    toast.error('Failed to create event')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <nav aria-label="breadcrumb" class="mb-3">
      <ol class="breadcrumb mb-0">
        <li class="breadcrumb-item"><RouterLink to="/">Events</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">New event</li>
      </ol>
    </nav>

    <h1 class="h3 mb-4">Create event</h1>
    <p class="text-muted">Step 1: choose a place, date and time, and topic. Then open the event to add participants and start the session.</p>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-if="loading" class="text-muted">Loading places…</div>

    <template v-else>
      <div v-if="!places.length" class="alert alert-warning">
        No places defined.
        <RouterLink to="/places">Add places</RouterLink>
        first.
      </div>

      <form v-else @submit.prevent="submit">
        <h2 class="h5 mb-3">1. Place</h2>
        <div class="row row-cols-1 row-cols-md-2 g-3 mb-4">
          <div v-for="p in places" :key="p.id" class="col">
            <div
              class="card h-100 shadow-sm place-card"
              :class="{ 'border-primary border-2': selectedPlaceId === p.id }"
              role="button"
              tabindex="0"
              @click="selectedPlaceId = p.id"
              @keydown.enter.prevent="selectedPlaceId = p.id"
            >
              <div class="row g-0">
                <div class="col-5 col-md-4">
                  <div class="ratio ratio-4x3 rounded-start overflow-hidden bg-body-secondary">
                    <img
                      v-if="p.image_url"
                      :src="getImageUrl(p.image_url)"
                      class="object-fit-cover"
                      alt=""
                      @error="($event.target.style.display = 'none')"
                    />
                  </div>
                </div>
                <div class="col">
                  <div class="card-body py-2 px-3">
                    <div class="form-check">
                      <input
                        :id="'place-' + p.id"
                        v-model="selectedPlaceId"
                        class="form-check-input"
                        type="radio"
                        :value="p.id"
                        name="place"
                        @click.stop
                      />
                      <label class="form-check-label fw-semibold" :for="'place-' + p.id">{{ p.name }}</label>
                    </div>
                    <p v-if="p.metro_area" class="small text-muted mb-1">📍 {{ p.metro_area }}</p>
                    <a v-if="p.map_url" :href="p.map_url" class="small" target="_blank" rel="noopener" @click.stop>Map</a>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <h2 class="h5 mb-3">2. Date and topic</h2>
        <div class="row g-3 mb-4">
          <div class="col-md-4">
            <label class="form-label">Date and time</label>
            <input v-model="scheduledLocal" class="form-control" type="datetime-local" required @change="updateNotifyAt" />
          </div>
          <div class="col-md-4">
            <label class="form-label">Topic</label>
            <input v-model="topic" class="form-control" placeholder="e.g. Theft, Addiction, Movie night…" required />
          </div>
          <div class="col-md-4">
            <div class="mb-3">
              <div class="form-check">
                <input
                  id="notifications-enabled"
                  v-model="notificationsEnabled"
                  class="form-check-input"
                  type="checkbox"
                  @change="updateNotifyAt"
                />
                <label class="form-check-label" for="notifications-enabled">
                  Enable notifications
                </label>
              </div>
              <div class="form-text small">Send reminder to participants before the event</div>
            </div>
            <div v-if="notificationsEnabled">
              <label class="form-label">Notification time</label>
              <input v-model="notifyLocal" class="form-control" type="datetime-local" />
              <div class="form-text small">Default: 8 hours before event</div>
            </div>
          </div>
        </div>

        <h2 class="h5 mb-3">3. Event Type</h2>
        <div class="mb-4">
          <div class="row g-3">
            <div class="col-md-6">
              <div class="form-check">
                <input 
                  id="type-many-activities" 
                  v-model="eventType" 
                  class="form-check-input" 
                  type="radio" 
                  value="many_activities" 
                />
                <label class="form-check-label" for="type-many-activities">
                  <strong>Many Activities</strong>
                </label>
                <div class="form-text">
                  Multiple activities with table shuffling and group assignments. Participants will be divided into small groups and rotated between tables.
                </div>
              </div>
            </div>
            <div class="col-md-6">
              <div class="form-check">
                <input 
                  id="type-single-activity" 
                  v-model="eventType" 
                  class="form-check-input" 
                  type="radio" 
                  value="single_activity" 
                />
                <label class="form-check-label" for="type-single-activity">
                  <strong>Single Activity</strong>
                </label>
                <div class="form-text">
                  One activity for the whole group (e.g., movie night, hiking). No table assignments or shuffling - everyone participates together.
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Finished At input for single activity events -->
        <div v-if="eventType === 'single_activity'" class="mb-4">
          <h3 class="h6 mb-2">Activity End Time</h3>
          <div class="row g-3">
            <div class="col-md-4">
              <label class="form-label">When does the activity finish?</label>
              <input 
                v-model="finishedLocal" 
                class="form-control" 
                type="datetime-local" 
                required 
              />
              <div class="form-text small">Set the time when the single activity will end</div>
            </div>
          </div>
        </div>

        <h2 v-if="eventType === 'many_activities'" class="h5 mb-3">4. Activities</h2>
        <div v-if="eventType === 'many_activities'" class="mb-4">
          <div v-for="(activity, index) in activities" :key="index" class="row g-2 mb-2">
            <div class="col-md-5">
              <input 
                v-model="activity.name" 
                class="form-control" 
                placeholder="Activity name (e.g. One-to-One Talks, Group Discussion)" 
                required 
              />
            </div>
            <div class="col-md-3">
              <div class="input-group">
                <input 
                  v-model.number="activity.duration" 
                  class="form-control" 
                  type="number" 
                  min="1" 
                  placeholder="Duration"
                  required 
                />
                <span class="input-group-text">minutes</span>
              </div>
            </div>
            <div class="col-md-2">
              <div class="input-group">
                <input 
                  v-model.number="activity.groupSize" 
                  class="form-control" 
                  type="number" 
                  min="1" 
                  placeholder="Size"
                  required 
                />
                <span class="input-group-text">people</span>
              </div>
            </div>
            <div class="col-md-2">
              <button 
                v-if="activities.length > 1" 
                type="button" 
                class="btn btn-outline-danger w-100"
                @click="removeActivity(index)"
              >
                <i class="bi bi-trash"></i> Remove
              </button>
            </div>
          </div>
          <button 
            type="button" 
            class="btn btn-outline-primary"
            @click="addActivity"
          >
            <i class="bi bi-plus-circle me-1"></i> Add Activity
          </button>
        </div>

        <button type="submit" class="btn btn-primary btn-lg" :disabled="!canSubmit || saving">
          {{ saving ? 'Saving…' : 'Create event' }}
        </button>
      </form>
    </template>
  </div>
</template>

<style scoped>
.place-card {
  cursor: pointer;
  transition: box-shadow 0.15s ease;
}
.place-card:hover {
  box-shadow: 0 0.25rem 0.75rem rgba(0, 0, 0, 0.08) !important;
}
.object-fit-cover {
  object-fit: cover;
  width: 100%;
  height: 100%;
}
</style>
