import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import PlacesView from '../views/PlacesView.vue'
import EventCreateView from '../views/EventCreateView.vue'
import EventDetailView from '../views/EventDetailView.vue'
import StatisticsView from '../views/StatisticsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/places', name: 'places', component: PlacesView },
    { path: '/events/new', name: 'event-new', component: EventCreateView },
    { path: '/events/:id', name: 'event-detail', component: EventDetailView, props: true },
    { path: '/statistics', name: 'statistics', component: StatisticsView },
  ],
})

export default router
