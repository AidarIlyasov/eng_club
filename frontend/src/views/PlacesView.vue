<script setup>
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '../api/client'
import { useToast } from '../composables/useToast'
import MetroIcon from '../components/MetroIcon.vue'

const places = ref([])
const metroStations = ref([])
const loading = ref(true)
const error = ref('')
const saving = ref(false)
const showModal = ref(false)
const uploadingImage = ref(false)

// Helper function to get full image URL
function getImageUrl(filename) {
  if (!filename) return ''
  if (filename.startsWith('http://') || filename.startsWith('https://')) return filename
  return `/api/uploads/${filename}`
}

// Metro search functionality
const metroSearch = ref('')
const showMetroDropdown = ref(false)
const filteredStations = ref([])

const toast = useToast()

const form = reactive({
  id: null,
  name: '',
  metro_area: '',
  map_url: '',
  image_url: '',
})

function resetForm() {
  form.id = null
  form.name = ''
  form.metro_area = ''
  form.map_url = ''
  form.image_url = ''
  metroSearch.value = ''
  showMetroDropdown.value = false
  filteredStations.value = []
}

function openModal() {
  resetForm()
  showModal.value = true
  // Initialize filtered stations when opening modal
  filterMetroStations()
}

function closeModal() {
  showModal.value = false
  resetForm()
}

function edit(p) {
  form.id = p.id
  form.name = p.name
  form.metro_area = p.metro_area || ''
  form.map_url = p.map_url || ''
  form.image_url = p.image_url || ''
  
  // Set metro search to current metro area value
  metroSearch.value = form.metro_area
  
  showModal.value = true
  // Initialize filtered stations for editing
  filterMetroStations()
}

function filterMetroStations() {
  const search = metroSearch.value.toLowerCase().trim()
  
  if (!search) {
    filteredStations.value = metroStations.value.slice(0, 20) // Show first 20 when empty
  } else {
    filteredStations.value = metroStations.value
      .filter(station => 
        station.name.toLowerCase().includes(search) || 
        station.value.toLowerCase().includes(search)
      )
      .slice(0, 10) // Limit to 10 results for performance
  }
}

function selectStation(station) {
  form.metro_area = station.name
  metroSearch.value = station.name
  showMetroDropdown.value = false
}

function hideDropdown() {
  // Delay hiding to allow click events to fire
  setTimeout(() => {
    showMetroDropdown.value = false
  }, 200)
}

function getMetroLineForStation(stationName) {
  if (!stationName) return null
  
  // Find station by name (case-insensitive partial match)
  const station = metroStations.value.find(s => 
    s.name.toLowerCase().includes(stationName.toLowerCase()) ||
    stationName.toLowerCase().includes(s.name.toLowerCase())
  )
  
  return station ? station.line : null
}

async function handleImageUpload(event) {
  const file = event.target.files[0]
  if (!file) return

  // Check file size (5MB limit)
  if (file.size > 5 * 1024 * 1024) {
    toast.error('Image size should be less than 5MB')
    return
  }

  // Check file type
  if (!file.type.startsWith('image/')) {
    toast.error('Please select an image file')
    return
  }

  uploadingImage.value = true
  const formData = new FormData()
  formData.append('image', file)

  try {
    const response = await fetch('/api/images/upload', {
      method: 'POST',
      body: formData,
    })
    
    if (!response.ok) throw new Error('Upload failed')
    
    const data = await response.json()
    form.image_url = data.filename
    toast.success('Image uploaded successfully')
  } catch (e) {
    toast.error('Failed to upload image')
  } finally {
    uploadingImage.value = false
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    // Load both places and metro stations in parallel
    const [placesList, metroList] = await Promise.all([
      api.listPlaces(),
      api.getMetroList()
    ])
    places.value = placesList
    metroStations.value = metroList
  } catch (e) {
    error.value = e.message
    places.value = []
    metroStations.value = []
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!form.name.trim()) {
    toast.error('Name is required')
    return
  }
  saving.value = true
  error.value = ''
  const body = {
    name: form.name.trim(),
    metro_area: form.metro_area.trim(),
    map_url: form.map_url.trim(),
    image_url: form.image_url.trim(),
  }
  try {
    if (form.id) {
      await api.updatePlace(form.id, body)
      toast.success('Place updated successfully')
    } else {
      await api.createPlace(body)
      toast.success('Place created successfully')
    }
    closeModal()
    await load()
  } catch (e) {
    toast.error(e.message)
  } finally {
    saving.value = false
  }
}

async function remove(p) {
  if (!confirm(`Delete place "${p.name}"? Events referencing it may fail.`)) return
  try {
    await api.deletePlace(p.id)
    toast.success('Place deleted successfully')
    if (form.id === p.id) resetForm()
    await load()
  } catch (e) {
    toast.error(e.message)
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h1 class="h3 mb-0">Places</h1>
      <button class="btn btn-primary" @click="openModal">
        <i class="bi bi-plus-circle me-2"></i>Add Place
      </button>
    </div>
    
    <p class="text-muted small mb-4">
      Venues appear as cards when <RouterLink to="/events/new">creating an event</RouterLink>. Use map and image URLs
      your members can open (e.g. Yandex Maps, hosted photos).
    </p>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <div v-if="loading" class="text-muted text-center py-4">Loading…</div>
    <div v-else-if="!places.length" class="alert alert-info">No places yet. Click "Add Place" to create one.</div>
    <div v-else class="row g-4">
      <div v-for="p in places" :key="p.id" class="col-md-6 col-lg-4">
        <div class="card h-100 shadow-sm">
          <div v-if="p.image_url" class="card-img-top" style="height: 200px; overflow: hidden;">
            <img :src="getImageUrl(p.image_url)" :alt="p.name" class="w-100 h-100 object-fit-cover" />
          </div>
          <div v-else class="card-img-top bg-light d-flex align-items-center justify-content-center" style="height: 200px;">
            <i class="bi bi-building text-muted" style="font-size: 3rem;"></i>
          </div>
          <div class="card-body">
            <h5 class="card-title">{{ p.name }}</h5>
            <p v-if="p.metro_area" class="card-text text-muted d-flex align-items-center">
              <MetroIcon :size="20" :line="getMetroLineForStation(p.metro_area)" class="me-1" />{{ p.metro_area }}
            </p>
            <div class="d-flex gap-2 mt-3">
              <a v-if="p.map_url" :href="p.map_url" target="_blank" class="btn btn-sm btn-outline-primary" rel="noopener">
                <i class="bi bi-map me-1"></i>Map
              </a>
              <button type="button" class="btn btn-sm btn-outline-secondary" @click="edit(p)">
                <i class="bi bi-pencil me-1"></i>Edit
              </button>
              <button type="button" class="btn btn-sm btn-outline-danger" @click="remove(p)">
                <i class="bi bi-trash me-1"></i>Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal for Create/Edit Place -->
    <Teleport to="body">
      <div v-if="showModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ form.id ? 'Edit Place' : 'Add New Place' }}</h5>
              <button type="button" class="btn-close" @click="closeModal" :disabled="saving"></button>
            </div>
            <form @submit.prevent="submit">
              <div class="modal-body">
                <div class="mb-3">
                  <label class="form-label">Name *</label>
                  <input v-model="form.name" class="form-control" required />
                </div>
                <div class="mb-3">
                  <label class="form-label">Metro Station</label>
                  <div class="position-relative">
                    <input 
                      v-model="metroSearch" 
                      @focus="showMetroDropdown = true"
                      @blur="hideDropdown"
                      @input="filterMetroStations"
                      class="form-control" 
                      type="text" 
                      placeholder="Search metro station..."
                      autocomplete="off"
                    />
                    <div v-if="showMetroDropdown && filteredStations.length" class="dropdown-menu show position-absolute w-100" style="max-height: 250px; overflow-y: auto;">
                      <div v-for="station in filteredStations" :key="station.value" 
                           @mousedown="selectStation(station)"
                           class="dropdown-item d-flex justify-content-between align-items-center">
                        <span>{{ station.name }}</span>
                        <small class="text-muted">Line {{ station.line }}</small>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="mb-3">
                  <label class="form-label">Map URL</label>
                  <input v-model="form.map_url" class="form-control" type="url" placeholder="https://…" />
                </div>
                <div class="mb-3">
                  <label class="form-label">Image</label>
                  <div v-if="form.image_url" class="mb-2">
                    <img :src="getImageUrl(form.image_url)" class="img-thumbnail" style="max-height: 150px;" />
                  </div>
                  <div>
                    <label class="btn btn-outline-primary" :class="{ 'disabled': uploadingImage }">
                      <i class="bi bi-upload me-1"></i>{{ uploadingImage ? 'Uploading...' : 'Upload Image' }}
                      <input type="file" accept="image/*" @change="handleImageUpload" class="d-none" :disabled="uploadingImage" />
                    </label>
                  </div>
                </div>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary" @click="closeModal" :disabled="saving">Cancel</button>
                <button type="submit" class="btn btn-primary" :disabled="saving || uploadingImage">
                  {{ saving ? 'Saving...' : (form.id ? 'Save Changes' : 'Create Place') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.object-fit-cover {
  object-fit: cover;
}

.modal {
  animation: fadeIn 0.2s;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.dropdown-menu {
  z-index: 1050;
  border: 1px solid #dee2e6;
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
}

.dropdown-item {
  cursor: pointer;
}

.dropdown-item:hover {
  background-color: #e9ecef;
}
</style>