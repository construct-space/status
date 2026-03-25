<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { CButton, CSkeleton, CToast, useToast } from '@construct-space/ui'

const { add: addToast } = useToast()

const services = ref([])
const overall = ref('loading')
const operational = ref(0)
const total = ref(0)
const checkedAt = ref(null)
const loading = ref(true)
const refreshing = ref(false)
const secondsAgo = ref(0)

let pollTimer = null
let tickTimer = null

const overallLabel = computed(() => {
  const labels = {
    operational: 'All Systems Operational',
    degraded: 'Some Systems Degraded',
    outage: 'Major Outage',
    loading: 'Checking systems...'
  }
  return labels[overall.value] || overall.value
})

const overallColor = computed(() => {
  const colors = { operational: '#34C759', degraded: '#FFD60A', outage: '#FF453A', loading: '#666' }
  return colors[overall.value] || '#666'
})

function statusColor(status) {
  return { operational: '#34C759', degraded: '#FFD60A', offline: '#FF453A' }[status] || '#666'
}

function statusLabel(status) {
  return { operational: 'Operational', degraded: 'Degraded', offline: 'Offline' }[status] || status
}

function timeAgoText() {
  if (!checkedAt.value) return ''
  const s = secondsAgo.value
  if (s < 5) return 'just now'
  if (s < 60) return `${s}s ago`
  return `${Math.floor(s / 60)}m ago`
}

function applyData(data) {
  services.value = data.services || []
  overall.value = data.overall || 'outage'
  operational.value = data.operational || 0
  total.value = data.total || 0
  checkedAt.value = new Date()
  secondsAgo.value = 0
  loading.value = false
}

async function fetchStatus() {
  try {
    const r = await fetch('/api/status')
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const d = await r.json()
    applyData(d)
  } catch (e) {
    console.error('Fetch failed:', e)
  }
}

async function refresh() {
  refreshing.value = true
  try {
    const r = await fetch('/api/refresh', { method: 'POST' })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    const d = await r.json()
    applyData(d)
    addToast({ title: 'Status refreshed', color: 'success', duration: 2000 })
  } catch (e) {
    addToast({ title: 'Refresh failed', description: e.message, color: 'error' })
  } finally {
    refreshing.value = false
  }
}

onMounted(() => {
  fetchStatus()
  pollTimer = setInterval(fetchStatus, 30000)
  tickTimer = setInterval(() => {
    if (checkedAt.value) {
      secondsAgo.value = Math.floor((Date.now() - checkedAt.value.getTime()) / 1000)
    }
  }, 1000)
})

onUnmounted(() => {
  clearInterval(pollTimer)
  clearInterval(tickTimer)
})
</script>

<template>
  <div class="status-page">
    <div class="status-container">

      <!-- Header -->
      <header class="status-header">
        <div class="brand">
          <div class="brand-mark">C</div>
          <span class="brand-text">System Status</span>
        </div>
      </header>

      <!-- Overall banner -->
      <div class="banner" :style="{ '--accent': overallColor }">
        <div class="banner-inner">
          <span class="banner-dot" :class="overall" :style="{ background: overallColor, boxShadow: `0 0 12px ${overallColor}` }"></span>
          <span class="banner-label">{{ overallLabel }}</span>
          <span v-if="!loading" class="banner-count">{{ operational }}/{{ total }}</span>
        </div>
      </div>

      <!-- Meta row -->
      <div class="meta-row">
        <span class="meta-time">{{ checkedAt ? `Checked ${timeAgoText()}` : '' }}</span>
        <CButton
          label="Refresh"
          icon="i-lucide-refresh-cw"
          variant="ghost"
          size="xs"
          :loading="refreshing"
          @click="refresh"
        />
      </div>

      <!-- Service list -->
      <div class="service-list">
        <!-- Loading skeletons -->
        <template v-if="loading">
          <div v-for="i in 9" :key="i" class="service-row">
            <CSkeleton width="120px" height="14px" />
            <CSkeleton width="60px" height="12px" />
            <CSkeleton width="80px" height="12px" />
          </div>
        </template>

        <!-- Services -->
        <template v-else>
          <div
            v-for="(svc, i) in services"
            :key="svc.name"
            class="service-row"
            :style="{ animationDelay: `${i * 0.04}s` }"
          >
            <div class="svc-info">
              <span class="svc-name">{{ svc.name }}</span>
              <span class="svc-domain">{{ svc.domain }}</span>
            </div>
            <span class="svc-latency">{{ svc.latency_ms != null ? `${svc.latency_ms} ms` : '—' }}</span>
            <div class="svc-status">
              <span class="svc-status-label" :style="{ color: statusColor(svc.status) }">
                {{ statusLabel(svc.status) }}
              </span>
              <span class="svc-dot" :style="{ background: statusColor(svc.status) }"></span>
            </div>
          </div>
        </template>
      </div>

      <!-- Footer -->
      <footer class="status-footer">
        <a href="https://construct.space">construct.space</a>
      </footer>
    </div>

    <CToast />
  </div>
</template>

<style>
:root {
  --app-background: #0a0a0a;
  --app-foreground: #e8e8e8;
  --app-muted: #666;
  --app-accent: #34C759;
  --app-accent-hover: #2db84e;
  --app-border: #1e1e1e;
  --app-card-bg: #111111;
  --app-card-hover: #181818;
  --app-input-bg: #111111;
  --app-canvas-bg: #0a0a0a;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: 'Rubik', ui-sans-serif, system-ui, sans-serif;
  background: #0a0a0a;
  color: #e8e8e8;
  -webkit-font-smoothing: antialiased;
  min-height: 100vh;
}

#app {
  min-height: 100vh;
  background:
    radial-gradient(ellipse 80% 60% at 50% -20%, rgba(52, 199, 89, 0.05) 0%, transparent 70%),
    #0a0a0a;
}

.status-page {
  min-height: 100vh;
}

.status-container {
  max-width: 720px;
  margin: 0 auto;
  padding: 0 20px;
}

/* Header */
.status-header {
  padding: 60px 0 32px;
  text-align: center;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.brand-mark {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  background: #34C759;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 14px;
  color: #000;
}

.brand-text {
  font-size: 14px;
  font-weight: 500;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #666;
}

/* Banner */
.banner {
  padding: 28px 32px;
  border-radius: 14px;
  border: 1px solid #1e1e1e;
  background: #111;
  position: relative;
  overflow: hidden;
  animation: fadeUp 0.5s ease forwards;
}

.banner::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 14px;
  background: linear-gradient(135deg, var(--accent, #34C759), transparent 60%);
  opacity: 0.04;
  pointer-events: none;
}

.banner-inner {
  display: flex;
  align-items: center;
  gap: 14px;
  position: relative;
}

.banner-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
  position: relative;
}

.banner-dot.operational {
  animation: pulse 2.5s ease infinite;
}

.banner-label {
  font-size: 16px;
  font-weight: 500;
}

.banner-count {
  margin-left: auto;
  font-size: 12px;
  font-family: 'JetBrains Mono', monospace;
  color: #666;
}

/* Meta row */
.meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0 20px;
}

.meta-time {
  font-size: 12px;
  color: #444;
  font-family: 'JetBrains Mono', monospace;
}

/* Service list */
.service-list {
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: #1e1e1e;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #1e1e1e;
}

.service-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: #111;
  transition: background 0.15s;
  animation: fadeUp 0.35s ease both;
}

.service-row:hover {
  background: #181818;
}

.svc-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.svc-name {
  font-size: 14px;
  font-weight: 500;
}

.svc-domain {
  font-size: 11px;
  font-family: 'JetBrains Mono', monospace;
  color: #444;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.svc-latency {
  font-size: 12px;
  font-family: 'JetBrains Mono', monospace;
  color: #666;
  text-align: right;
  white-space: nowrap;
}

.svc-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.svc-status-label {
  font-size: 12px;
  font-weight: 500;
  min-width: 80px;
  text-align: right;
}

.svc-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* Footer */
.status-footer {
  text-align: center;
  padding: 48px 0 40px;
  font-size: 12px;
  color: #444;
}

.status-footer a {
  color: #666;
  text-decoration: none;
  transition: color 0.15s;
}

.status-footer a:hover { color: #e8e8e8; }

/* Animations */
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Responsive */
@media (max-width: 540px) {
  .status-header { padding-top: 40px; }
  .banner { padding: 20px 18px; }
  .service-row { padding: 14px 16px; gap: 10px; }
  .svc-latency { display: none; }
  .banner-count { display: none; }
}
</style>
