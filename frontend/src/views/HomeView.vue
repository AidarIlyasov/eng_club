<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '../api/client'
import { useToast } from '../composables/useToast'
import MetroIcon from '../components/MetroIcon.vue'

const events = ref([])
const places = ref([])
const metroStations = ref([])
const loading = ref(true)
const error = ref('')
const toast = useToast()

// Filters
const filters = ref({
  dateFrom: '',
  dateTo: '',
  placeId: '',
  status: ''
})

function clearFilters() {
  filters.value.dateFrom = ''
  filters.value.dateTo = ''
  filters.value.placeId = ''
  filters.value.status = ''
  load() // Reload events without filters
}

const hasActiveFilters = computed(() => {
  return filters.value.dateFrom || filters.value.dateTo || filters.value.placeId || filters.value.status
})

const showFilters = ref(false)

function applyFilters() {
  load() // Reload events with current filters
}

function fmt(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleString(undefined, {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getMetroLineForArea(areaName) {
  if (!areaName) return null
  
  // Find station by name (case-insensitive partial match)
  const station = metroStations.value.find(s => 
    s.name.toLowerCase().includes(areaName.toLowerCase()) ||
    areaName.toLowerCase().includes(s.name.toLowerCase())
  )
  
  return station ? station.line : null
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    // Load events with filters
    const eventsList = await api.listEvents(filters.value)
    events.value = eventsList
    
    // Extract unique metro stations from events
    const stationsMap = new Map()
    eventsList.forEach(ev => {
      if (ev.place?.metro_area) {
        // Parse metro area to extract line info if needed
        stationsMap.set(ev.place.metro_area, { name: ev.place.metro_area, line: null })
      }
    })
    metroStations.value = Array.from(stationsMap.values())
  } catch (e) {
    error.value = e.message
    events.value = []
    metroStations.value = []
  } finally {
    loading.value = false
  }
}

async function loadPlaces() {
  try {
    // Load places only for the filter dropdown
    places.value = await api.listPlaces()
  } catch (e) {
    console.error('Failed to load places:', e)
    places.value = []
  }
}

async function remove(ev) {
  if (ev.status !== 'planned') return
  if (!confirm(`Delete planned event #${ev.id}?`)) return
  try {
    await api.deleteEvent(ev.id)
    toast.success('Event deleted successfully')
    await load()
  } catch (e) {
    error.value = e.message
    toast.error('Failed to delete event')
  }
}

onMounted(() => {
  loadPlaces() // Load places once for filter dropdown
  load() // Load events
})
</script>

<template>
  <div>
    <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-4">
      <h1 class="h3 mb-0">Events</h1>
      <RouterLink to="/events/new" class="btn btn-primary">
        <i class="bi bi-plus-circle me-1"></i>Create event
      </RouterLink>
    </div>

    <div v-if="error" class="alert alert-danger" role="alert">{{ error }}</div>
    <div v-if="loading" class="text-muted text-center py-4">Loading…</div>

    <div v-else-if="!events.length" class="card card-body text-muted text-center">
      <div class="py-4">
        <i class="bi bi-calendar-event display-1 text-muted mb-3"></i>
        <h5>No events yet</h5>
        <p>Add <RouterLink to="/places">places</RouterLink> first, then
        <RouterLink to="/events/new">create an event</RouterLink>.</p>
      </div>
    </div>

    <template v-else>
      <!-- Filters -->
      <div class="card shadow-sm mb-3">
        <div class="card-header bg-white d-md-none">
          <button 
            class="btn btn-sm btn-link text-decoration-none p-0 w-100 text-start d-flex justify-content-between align-items-center"
            @click="showFilters = !showFilters"
          >
            <span>
              <i class="bi bi-funnel me-2"></i>Filters
              <span v-if="hasActiveFilters" class="badge bg-primary ms-2">Active</span>
            </span>
            <i class="bi" :class="showFilters ? 'bi-chevron-up' : 'bi-chevron-down'"></i>
          </button>
        </div>
        <div class="card-body" :class="{ 'd-none d-md-block': !showFilters }">
          <div class="row g-2">
            <div class="col-md-3">
              <label class="form-label small text-muted">Date From</label>
              <input v-model="filters.dateFrom" type="date" class="form-control form-control-sm" @change="applyFilters" />
            </div>
            <div class="col-md-3">
              <label class="form-label small text-muted">Date To</label>
              <input v-model="filters.dateTo" type="date" class="form-control form-control-sm" @change="applyFilters" />
            </div>
            <div class="col-md-3">
              <label class="form-label small text-muted">Place</label>
              <select v-model="filters.placeId" class="form-select form-select-sm" @change="applyFilters">
                <option value="">All Places</option>
                <option v-for="place in places" :key="place.id" :value="place.id">{{ place.name }}</option>
              </select>
            </div>
            <div class="col-md-2">
              <label class="form-label small text-muted">Status</label>
              <select v-model="filters.status" class="form-select form-select-sm" @change="applyFilters">
                <option value="">All Statuses</option>
                <option value="planned">Planned</option>
                <option value="started">Started</option>
                <option value="completed">Completed</option>
              </select>
            </div>
            <div class="col-md-1 d-flex align-items-end">
              <button 
                v-if="hasActiveFilters" 
                class="btn btn-sm btn-outline-secondary w-100" 
                @click="clearFilters"
                title="Clear filters"
              >
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
          </div>
          <div v-if="hasActiveFilters" class="mt-2 small text-muted">
            <i class="bi bi-funnel"></i> Showing {{ events.length }} filtered events
          </div>
        </div>
      </div>

      <!-- Desktop table view -->
      <div class="d-none d-md-block table-responsive shadow-sm rounded border bg-white">
      <table class="table table-hover align-middle mb-0">
        <thead class="table-light">
          <tr>
            <th>When</th>
            <th>Topic</th>
            <th>Place</th>
            <th>Status</th>
            <th class="text-end">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="ev in events" 
            :key="ev.id"
            :class="{
              'table-success': ev.status === 'completed',
              'table-primary': ev.status === 'started'
            }"
          >
            <td>
              <div class="fw-medium">{{ fmt(ev.scheduled_at) }}</div>
            </td>
            <td>
              <div class="fw-medium">{{ ev.topic || '—' }}</div>
            </td>
            <td>
              <div>
                <i class="bi bi-building me-1 text-muted"></i>
                <span class="fw-medium">{{ ev.place?.name || '—' }}</span>
              </div>
              <div v-if="ev.place?.metro_area" class="small text-muted d-flex align-items-center">
                <MetroIcon :size="20" :line="getMetroLineForArea(ev.place.metro_area)" class="me-1" />{{ ev.place.metro_area }}
              </div>
            </td>
            <td>
              <span
                class="badge rounded-pill"
                :class="{
                  'text-bg-secondary': ev.status === 'planned',
                  'text-bg-primary': ev.status === 'started',
                  'text-bg-success': ev.status === 'completed',
                }"
              >
                {{ ev.status }}
              </span>
            </td>
            <td class="text-end text-nowrap">
              <RouterLink :to="`/events/${ev.id}`" class="btn btn-sm btn-primary">
                <i class="bi bi-eye me-1"></i>Open
              </RouterLink>
              <button
                v-if="ev.status === 'planned'"
                type="button"
                class="btn btn-sm btn-outline-danger ms-1"
                @click="remove(ev)"
              >
                <i class="bi bi-trash me-1"></i>Delete
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Mobile card view -->
    <div class="d-md-none">
      <div 
        v-for="ev in events" 
        :key="ev.id"
        class="card mb-3 shadow-sm"
        :class="{
          'border-success bg-success bg-opacity-10': ev.status === 'completed',
          'border-primary bg-primary bg-opacity-10': ev.status === 'started'
        }"
      >
        <div class="card-body">
          <div class="d-flex justify-content-between align-items-start mb-2">
            <h5 class="card-title mb-0">{{ ev.topic || 'Untitled' }}</h5>
            <span
              class="badge rounded-pill flex-shrink-0 ms-2"
              :class="{
                'text-bg-secondary': ev.status === 'planned',
                'text-bg-primary': ev.status === 'started',
                'text-bg-success': ev.status === 'completed',
              }"
            >
              {{ ev.status }}
            </span>
          </div>
          
          <div class="text-muted small mb-2">
            <i class="bi bi-calendar me-1"></i>{{ fmt(ev.scheduled_at) }}
          </div>
          
          <div class="mb-2">
            <div>
              <i class="bi bi-building me-1 text-muted"></i>
              <span class="fw-medium">{{ ev.place?.name || '—' }}</span>
            </div>
            <div v-if="ev.place?.metro_area" class="small text-muted d-flex align-items-center">
              <MetroIcon :size="20" :line="getMetroLineForArea(ev.place.metro_area)" class="me-1" />{{ ev.place.metro_area }}
            </div>
          </div>
          
          <div class="d-flex gap-2">
            <RouterLink :to="`/events/${ev.id}`" class="btn btn-sm btn-primary flex-grow-1">
              <i class="bi bi-eye me-1"></i>Open
            </RouterLink>
            <button
              v-if="ev.status === 'planned'"
              type="button"
              class="btn btn-sm btn-outline-danger"
              @click="remove(ev)"
            >
              <i class="bi bi-trash"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
    </template>
  </div>
</template>