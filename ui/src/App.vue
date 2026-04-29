<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { themes, loadSavedTheme, saveTheme } from './themes'

const services = ref([])
const overall = ref('loading')
const operational = ref(0)
const total = ref(0)
const checkedAt = ref(null)
const loading = ref(true)
const refreshing = ref(false)
const secondsAgo = ref(0)

const currentTheme = ref(loadSavedTheme())
const showThemePicker = ref(false)

let pollTimer = null
let tickTimer = null

const overallLabel = computed(() => {
  const labels = {
    operational: 'All Systems Operational',
    degraded: 'Some Systems Degraded',
    outage: 'Major Outage',
    loading: 'Checking systems…'
  }
  return labels[overall.value] || overall.value
})

const overallColor = computed(() => {
  const colors = { operational: '#34C759', degraded: '#FFD60A', outage: '#FF453A', loading: 'var(--app-muted)' }
  return colors[overall.value] || 'var(--app-muted)'
})

function statusColor(status) {
  return { operational: '#34C759', degraded: '#FFD60A', offline: '#FF453A' }[status] || 'var(--app-muted)'
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
    applyData(await r.json())
  } catch (e) {
    console.error('Fetch failed:', e)
  }
}

async function refresh() {
  if (refreshing.value) return
  refreshing.value = true
  try {
    const r = await fetch('/api/refresh', { method: 'POST' })
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    applyData(await r.json())
  } catch (e) {
    console.error('Refresh failed:', e)
  } finally {
    refreshing.value = false
  }
}

function selectTheme(theme) {
  currentTheme.value = theme
  saveTheme(theme)
  showThemePicker.value = false
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
          <svg
            class="brand-logo"
            viewBox="0 0 533 533"
            width="28"
            height="28"
            fill="currentColor"
            aria-hidden="true"
            focusable="false"
          >
            <path d="M266.5 410.156C230.912 410.156 199.106 402.203 171.081 386.297C143.056 370.39 121.036 348.519 105.022 320.684C89.0072 292.848 81 261.256 81 225.909C81 190.12 89.0072 158.308 105.022 130.472C121.036 102.636 143.056 80.7655 171.081 64.8593C199.106 48.9531 230.912 41 266.5 41C302.087 41 333.671 48.9531 361.252 64.8593C389.277 80.7655 411.297 102.636 427.311 130.472C443.326 158.308 451.555 190.12 452 225.909C452 261.256 443.77 292.848 427.311 320.684C411.297 348.519 389.277 370.39 361.252 386.297C333.671 402.203 302.087 410.156 266.5 410.156ZM266.5 363.763C292.301 363.763 315.433 357.798 335.896 345.868C356.359 333.939 372.373 317.591 383.939 296.824C395.505 276.058 401.288 252.42 401.288 225.909C401.288 199.399 395.505 175.761 383.939 154.994C372.373 133.786 356.359 117.217 335.896 105.287C315.433 93.3579 292.301 87.393 266.5 87.393C240.699 87.393 217.567 93.3579 197.104 105.287C176.641 117.217 160.405 133.786 148.394 154.994C136.828 175.761 131.045 199.399 131.045 225.909C131.045 252.42 136.828 276.058 148.394 296.824C160.405 317.591 176.641 333.939 197.104 345.868C217.567 357.798 240.699 363.763 266.5 363.763Z"/>
            <path d="M378.22 451.578C393.077 451.578 405.121 460.85 405.121 472.289C405.121 483.727 393.077 493 378.22 493H160.945C146.089 493 134.044 483.727 134.044 472.289C134.044 460.85 146.089 451.578 160.945 451.578H378.22Z"/>
          </svg>
          <span class="brand-wordmark">Construct</span>
        </div>
        <h1 class="page-title">System Status</h1>
        <p class="page-sub">Live health of every public service.</p>
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
        <span class="meta-time">
          <template v-if="loading">Checking…</template>
          <template v-else-if="checkedAt">Checked {{ timeAgoText() }}</template>
        </span>
        <button class="refresh-btn" :disabled="refreshing" @click="refresh">
          <svg
            class="refresh-icon"
            :class="{ spinning: refreshing }"
            width="14" height="14" viewBox="0 0 24 24"
            fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round"
            aria-hidden="true"
          >
            <path d="M3 12a9 9 0 0 1 15-6.7L21 8"/>
            <path d="M21 3v5h-5"/>
            <path d="M21 12a9 9 0 0 1-15 6.7L3 16"/>
            <path d="M3 21v-5h5"/>
          </svg>
          <span>Refresh</span>
        </button>
      </div>

      <!-- Service list -->
      <div v-if="!loading" class="service-list">
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
      </div>

      <!-- Footer -->
      <footer class="status-footer">
        <a href="https://construct.space">construct.space</a>
        <span class="footer-sep">·</span>
        <a href="https://my.construct.space">my.construct.space</a>
      </footer>
    </div>

    <!-- Theme picker -->
    <div class="theme-anchor">
      <button class="theme-toggle" @click="showThemePicker = !showThemePicker" :title="`Theme: ${currentTheme.name}`">
        <span class="theme-toggle-dot" :style="{ background: currentTheme.accent }"></span>
      </button>
      <div v-if="showThemePicker" class="theme-picker">
        <button
          v-for="t in themes"
          :key="t.id"
          class="theme-dot"
          :class="{ active: t.id === currentTheme.id }"
          :style="{ background: t.bg }"
          :title="t.name"
          @click="selectTheme(t)"
        >
          <span class="theme-dot-accent" :style="{ background: t.accent }"></span>
        </button>
      </div>
    </div>
  </div>
</template>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: 'Rubik', ui-sans-serif, system-ui, sans-serif;
  background: var(--app-canvas-bg, #1a1a2e);
  color: var(--app-foreground, #e5e5e5);
  -webkit-font-smoothing: antialiased;
  min-height: 100vh;
}

#app {
  min-height: 100vh;
  background:
    radial-gradient(ellipse 80% 60% at 50% -20%, color-mix(in srgb, var(--app-accent) 7%, transparent) 0%, transparent 70%),
    var(--app-canvas-bg, #1a1a2e);
}

.status-page { min-height: 100vh; position: relative; }

.status-container {
  max-width: 720px;
  margin: 0 auto;
  padding: 0 20px;
}

/* Header */
.status-header {
  padding: 72px 0 36px;
  text-align: center;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 28px;
}

.brand-logo {
  color: var(--app-accent);
  display: inline-block;
}

.brand-wordmark {
  font-size: 18px;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: var(--app-foreground);
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  letter-spacing: -0.02em;
  margin-bottom: 8px;
  color: var(--app-foreground);
}

.page-sub {
  font-size: 14px;
  color: var(--app-muted);
}

/* Banner */
.banner {
  padding: 24px 28px;
  border-radius: 12px;
  border: 1px solid var(--app-border);
  background: var(--app-card-bg);
  position: relative;
  overflow: hidden;
  animation: fadeUp 0.5s ease forwards;
}

.banner::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 12px;
  background: linear-gradient(135deg, var(--accent, var(--app-accent)), transparent 60%);
  opacity: 0.06;
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
}

.banner-dot.operational { animation: pulse 2.5s ease infinite; }

.banner-label {
  font-size: 16px;
  font-weight: 500;
  color: var(--app-foreground);
}

.banner-count {
  margin-left: auto;
  font-size: 12px;
  font-family: 'JetBrains Mono', ui-monospace, monospace;
  color: var(--app-muted);
}

/* Meta row */
.meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 4px 18px;
}

.meta-time {
  font-size: 12px;
  color: var(--app-muted);
  font-family: 'JetBrains Mono', ui-monospace, monospace;
  opacity: 0.7;
}

.refresh-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: transparent;
  border: 1px solid var(--app-border);
  border-radius: 6px;
  color: var(--app-muted);
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--app-card-hover);
  color: var(--app-foreground);
  border-color: var(--app-muted);
}

.refresh-btn:disabled { cursor: default; opacity: 0.6; }

.refresh-icon { color: currentColor; }
.refresh-icon.spinning { animation: spin 0.8s linear infinite; }

/* Service list */
.service-list {
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: var(--app-border);
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--app-border);
}

.service-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: var(--app-card-bg);
  transition: background 0.15s;
  animation: fadeUp 0.35s ease both;
}

.service-row:hover { background: var(--app-card-hover); }

.svc-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.svc-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--app-foreground);
}

.svc-domain {
  font-size: 11px;
  font-family: 'JetBrains Mono', ui-monospace, monospace;
  color: var(--app-muted);
  opacity: 0.7;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.svc-latency {
  font-size: 12px;
  font-family: 'JetBrains Mono', ui-monospace, monospace;
  color: var(--app-muted);
  text-align: right;
  white-space: nowrap;
}

.svc-status { display: flex; align-items: center; gap: 8px; }

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
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 48px 0 40px;
  font-size: 12px;
  color: var(--app-muted);
}

.footer-sep { color: var(--app-border); }

.status-footer a {
  color: var(--app-muted);
  text-decoration: none;
  transition: color 0.15s;
}

.status-footer a:hover { color: var(--app-foreground); }

/* Theme picker */
.theme-anchor {
  position: fixed;
  bottom: 24px;
  right: 24px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10px;
  z-index: 50;
}

.theme-toggle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--app-card-bg);
  border: 1px solid var(--app-border);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s, transform 0.15s;
}

.theme-toggle:hover { background: var(--app-card-hover); transform: translateY(-1px); }

.theme-toggle-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
}

.theme-picker {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 6px;
  padding: 10px;
  background: var(--app-card-bg);
  border: 1px solid var(--app-border);
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25);
}

.theme-dot {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid var(--app-border);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 0;
  transition: transform 0.15s;
}

.theme-dot:hover { transform: scale(1.08); }
.theme-dot.active { outline: 2px solid var(--app-accent); outline-offset: 1px; }
.theme-dot-accent { width: 8px; height: 8px; border-radius: 50%; }

/* Animations */
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Responsive */
@media (max-width: 540px) {
  .status-header { padding-top: 48px; }
  .page-title { font-size: 24px; }
  .banner { padding: 20px 18px; }
  .service-row { padding: 14px 16px; gap: 10px; }
  .svc-latency { display: none; }
  .banner-count { display: none; }
  .theme-anchor { bottom: 16px; right: 16px; }
}
</style>
