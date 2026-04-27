<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="min-vh-100 d-flex flex-column bg-light">
    <nav class="navbar navbar-expand-md navbar-dark bg-primary shadow-sm">
      <div class="container">
        <RouterLink class="navbar-brand fw-semibold" to="/">Eng Club Admin</RouterLink>
        <button
          class="navbar-toggler"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#nav"
          aria-controls="nav"
          aria-expanded="false"
          aria-label="Toggle navigation"
        >
          <span class="navbar-toggler-icon" />
        </button>
        <div id="nav" class="collapse navbar-collapse">
          <ul class="navbar-nav ms-auto gap-md-2">
            <li class="nav-item">
              <RouterLink class="nav-link" to="/places" active-class="fw-bold text-white">Places</RouterLink>
            </li>
            <li class="nav-item">
              <RouterLink class="nav-link" to="/statistics" active-class="fw-bold text-white">Statistics</RouterLink>
            </li>
          </ul>
        </div>
      </div>
    </nav>
    <main class="container py-4 flex-grow-1">
      <RouterView />
    </main>
    <footer class="border-top py-3 bg-white text-muted small">
      <div class="container">have fun</div>
    </footer>
  </div>
</template>

<style>
.container {
  max-width: 1200px !important;
}
</style>
