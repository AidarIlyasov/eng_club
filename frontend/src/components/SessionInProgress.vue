<script setup>
const props = defineProps({
  currentSection: { type: String, default: null },
  sectionTimeRemaining: { type: Number, default: 0 },
  sectionStates: { type: Array, required: true }
})

function formatTime(seconds) {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}
</script>

<template>
  <div class="mt-4">
    <h2 class="h5">Session in Progress</h2>
    
    <!-- Current Activity Timer -->
    <div v-if="currentSection" class="d-flex justify-content-end mb-3">
      <div class="card bg-primary text-white">
        <div class="card-body py-2 px-3">
          <div class="small">Current Section:</div>
          <div class="fw-bold">{{ currentSection }}</div>
          <div v-if="sectionTimeRemaining > 0" class="h4 mb-0 font-monospace">
            {{ formatTime(sectionTimeRemaining) }}
          </div>
        </div>
      </div>
    </div>
    
    <!-- Activity Timeline -->
    <div class="card shadow-sm mb-3">
      <div class="card-body">
        <h6 class="card-subtitle mb-3">Activity Timeline</h6>
        <div class="list-group list-group-flush">
          <div v-for="(section, index) in sectionStates" :key="index" class="list-group-item px-0">
            <div class="d-flex justify-content-between align-items-center">
              <div class="d-flex align-items-center">
                <i v-if="section.status === 'completed'" class="bi bi-check-circle-fill text-success me-2 fs-5"></i>
                <i v-else-if="section.status === 'active'" class="bi bi-play-circle-fill text-primary me-2 fs-5"></i>
                <i v-else class="bi bi-circle me-2 text-muted"></i>
                <span :class="{ 'text-muted text-decoration-line-through': section.status === 'completed' }">
                  {{ section.name }}
                </span>
              </div>
              <div class="d-flex align-items-center gap-3">
                <span v-if="section.status !== 'completed'" 
                      class="font-monospace fw-bold fs-5"
                      :class="{ 'text-primary': section.status === 'active', 'text-muted': section.status === 'pending' }">
                  {{ formatTime(section.timeRemaining) }}
                </span>
                <span v-else class="text-success fw-bold">
                  Completed
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.font-monospace {
  font-family: 'Courier New', monospace;
}

.text-decoration-line-through {
  text-decoration: line-through;
}
</style>