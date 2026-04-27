<script setup>
const props = defineProps({
  participantsByTable: { type: Object, required: true }
})

const hasAssignments = Object.keys(props.participantsByTable).length > 0
</script>

<template>
  <div class="card shadow-sm mb-3">
    <div class="card-header bg-white">
      <h6 class="mb-0">
        <i class="bi bi-people-fill me-2"></i>
        Table Assignments
      </h6>
    </div>
    <div class="card-body">
      <div v-if="hasAssignments">
        <div class="table-responsive">
          <table class="table table-hover align-middle">
            <tbody>
              <tr v-for="(participants, tableId) in participantsByTable" :key="tableId">
                <td>
                  <div class="d-flex align-items-center">
                    <strong>Table {{ tableId }}</strong>
                  </div>
                </td>
                <td>
                  <div class="d-flex flex-wrap gap-2">
                    <div v-for="participant in participants" :key="participant.member_id" class="badge bg-light text-dark border">
                      <div class="fw-medium">{{ participant.name }}</div>
                      <div class="small text-muted">{{ participant.telegram }}</div>
                    </div>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div v-else class="text-muted text-center py-3">
        <i class="bi bi-people display-4 text-muted mb-2"></i>
        <p>No table assignments yet</p>
      </div>
    </div>
  </div>
</template>