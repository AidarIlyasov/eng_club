const base = ''

async function request(path, options = {}) {
  const url = `${base}${path}`
  const headers = { ...options.headers }
  let body = options.body
  if (body != null && typeof body === 'object' && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify(body)
  }
  const res = await fetch(url, { ...options, headers, body })
  const text = await res.text()
  let data = null
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = { error: text }
    }
  }
  if (!res.ok) {
    const msg = data?.error || res.statusText || `HTTP ${res.status}`
    const err = new Error(msg)
    err.status = res.status
    err.data = data
    throw err
  }
  return data
}

export const api = {
  // Members
  getMembers:() => request('/api/members'),
  getWithoutPairs: () => request(`/api/pairs/without`),
  addWithoutPairs: (participant) => request(`/api/pairs/without/add`, { method: 'POST', body: participant }),

  // Metro stations
  getMetroList: () => request('/api/metro-list'),
  
  // Places
  listPlaces: () => request('/api/places'),
  getPlace: (id) => request(`/api/places/${id}`),
  createPlace: (body) => request('/api/places', { method: 'POST', body }),
  updatePlace: (id, body) => request(`/api/places/${id}`, { method: 'PUT', body }),
  deletePlace: (id) => request(`/api/places/${id}`, { method: 'DELETE' }),

  // Events
  listEvents: (filters = {}) => {
    const params = new URLSearchParams()
    if (filters.dateFrom) params.append('date_from', filters.dateFrom)
    if (filters.dateTo) params.append('date_to', filters.dateTo)
    if (filters.placeId) params.append('place_id', filters.placeId)
    if (filters.status) params.append('status', filters.status)
    const query = params.toString()
    return request(`/api/events${query ? '?' + query : ''}`)
  },
  getEvent: (id) => request(`/api/events/${id}`),
  createEvent: (body) => request('/api/events', { method: 'POST', body }),
  updateEvent: (id, body) => request(`/api/events/${id}`, { method: 'PUT', body }),
  deleteEvent: (id) => request(`/api/events/${id}`, { method: 'DELETE' }),
  startEvent: (id, body) =>
    request(`/api/events/${id}/start`, { method: 'POST', body }),
  getEventParticipants: (id) => request(`/api/events/${id}/participants`),
  markAttendance: (id, memberID, attended) =>
    request(`/api/events/${id}/attendance`, { method: 'POST', body: { member_id: memberID, attended } }),
  moveToTable: (id, memberID, tableID) =>
    request(`/api/events/${id}/move-to-table`, { method: 'POST', body: { member_id: memberID, table_id: tableID } }),
  addParticipantToEvent: (id, participant) =>
    request(`/api/events/${id}/add-participant`, { method: 'POST', body: participant }),
  removeParticipantFromEvent: (id, memberID) =>
    request(`/api/events/${id}/participants/${memberID}`, { method: 'DELETE' }),
  getSessionActivities: (id) => request(`/api/events/${id}/activities`),
}
