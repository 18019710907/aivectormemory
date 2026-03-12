<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useIssues } from '../composables/useIssues'
import { showError } from '../utils/toast'
import IssueCard from '../components/issues/IssueCard.vue'
import IssueEditModal from '../components/issues/IssueEditModal.vue'
import SearchBox from '../components/common/SearchBox.vue'
import Pager from '../components/common/Pager.vue'
import Modal from '../components/layout/Modal.vue'

const { t } = useI18n()
const route = useRoute()

const {
  issues, total, totalAll, totalToday, page, statusFilter, dateFrom, dateTo, loading,
  load, getDetail, create, updateFull, archive, remove,
  setPage, setStatus, setDateRange, setQuery, PAGE_SIZE,
} = useIssues()

// Edit/Create modal
const editModalShow = ref(false)
const editIssue = ref<any>(null)
const editMode = ref<'create' | 'edit'>('create')

// Archive/Delete confirmation
const confirmModalShow = ref(false)
const confirmAction = ref<'archive' | 'delete'>('archive')
const confirmTarget = ref<any>(null)

// Archived detail modal
const viewModalShow = ref(false)
const viewIssue = ref<any>(null)

onMounted(() => {
  if (route.query.status) setStatus(route.query.status as string)
  else load()
})

function onSearch(q: string) { setQuery(q) }

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function quickFilter(s: string) {
  if (s === 'today') {
    const t = todayStr()
    setDateRange(t, t)
  } else if (s === 'last3') {
    const end = new Date()
    const start = new Date(end)
    start.setDate(start.getDate() - 2)
    const from = `${start.getFullYear()}-${String(start.getMonth() + 1).padStart(2, '0')}-${String(start.getDate()).padStart(2, '0')}`
    const to = todayStr()
    setDateRange(from, to)
  } else if (s === 'all') {
    setDateRange('', '')
  } else {
    setStatus(s)
  }
}

function onStatusChange(e: Event) { setStatus((e.target as HTMLSelectElement).value) }

function onDateFromChange(e: Event) { setDateRange((e.target as HTMLInputElement).value, dateTo.value) }

function onDateToChange(e: Event) { setDateRange(dateFrom.value, (e.target as HTMLInputElement).value) }

// Create
function openCreate() {
  editMode.value = 'create'
  editIssue.value = null
  editModalShow.value = true
}

// Edit
async function openEdit(issue: any) {
  try {
    const detail = await getDetail(issue.id)
    editIssue.value = detail
    editMode.value = 'edit'
    editModalShow.value = true
  } catch (e) { showError(e) }
}

async function onSave(data: any) {
  try {
    if (editMode.value === 'create') {
      const result = await create(data.title, data.content, 'pending', data.tags || [], 0)
      if (result?.duplicate) {
        window.__toast?.show(t('issueDuplicated'), 'warning')
        return
      }
      window.__toast?.show(t('issueCreated'), 'success')
    } else if (editIssue.value) {
      await updateFull(editIssue.value.id, data)
      window.__toast?.show(t('issueUpdated'), 'success')
    }
    editModalShow.value = false
    load()
  } catch (e: any) {
    window.__toast?.show(e?.message || 'Failed', 'error')
  }
}

// Archive
function openArchive(issue: any) {
  confirmTarget.value = issue
  confirmAction.value = 'archive'
  confirmModalShow.value = true
}

// Delete
function openDelete(issue: any) {
  confirmTarget.value = issue
  confirmAction.value = 'delete'
  confirmModalShow.value = true
}

async function doConfirm() {
  if (!confirmTarget.value) return
  try {
    if (confirmAction.value === 'archive') {
      await archive(confirmTarget.value.id)
      window.__toast?.show(t('issueArchived'), 'success')
    } else {
      await remove(confirmTarget.value.id, confirmTarget.value.status === 'archived')
      window.__toast?.show(t('issueDeleted'), 'success')
    }
    confirmModalShow.value = false
    load()
  } catch (e: any) {
    window.__toast?.show(e?.message || 'Failed', 'error')
  }
}

// View archived
async function openView(issue: any) {
  try {
    viewIssue.value = await getDetail(issue.id)
    viewModalShow.value = true
  } catch (e) { showError(e) }
}
</script>

<template>
  <div class="issues-view">
    <div class="page-header">
      <h2 class="page-title">{{ t('issueTracking') }}</h2>
      <div class="page-stats">
        <span class="stat-pill">
          <span class="stat-pill__label">{{ t('issuesTotalAll') }}</span>
          <span class="stat-pill__num">{{ totalAll }}</span>
          <span class="stat-pill__unit">{{ t('items') }}</span>
        </span>
        <span class="stat-pill">
          <span class="stat-pill__label">{{ t('issuesTotalToday') }}</span>
          <span class="stat-pill__num">{{ totalToday }}</span>
          <span class="stat-pill__unit">{{ t('items') }}</span>
        </span>
        <span class="stat-pill stat-pill--highlight">
          <span class="stat-pill__label">{{ t('issuesCurrentQuery') }}</span>
          <span class="stat-pill__num">{{ total }}</span>
          <span class="stat-pill__unit">{{ t('items') }}</span>
        </span>
      </div>
    </div>

    <!-- Toolbar: 时间范围 + 状态 + 范围搜索 -->
    <div class="toolbar toolbar--wrap">
      <span class="filter-label">{{ t('dateRange') }}</span>
      <input type="date" class="filter-input" :value="dateFrom" @change="onDateFromChange" :title="t('dateFrom')" />
      <span class="filter-sep">–</span>
      <input type="date" class="filter-input" :value="dateTo" @change="onDateToChange" :title="t('dateTo')" />
      <button class="btn btn--outline btn--sm" @click="quickFilter('today')">{{ t('today') }}</button>
      <button class="btn btn--outline btn--sm" @click="quickFilter('last3')">{{ t('last3Days') }}</button>
      <button class="btn btn--outline btn--sm" @click="quickFilter('all')">{{ t('viewAll') }}</button>
      <select class="filter-select" :value="statusFilter" @change="onStatusChange">
        <option value="all">{{ t('allIncludeArchived') }}</option>
        <option value="active">{{ t('activeOnly') }}</option>
        <option value="pending">{{ t('status.pending') }}</option>
        <option value="in_progress">{{ t('status.in_progress') }}</option>
        <option value="completed">{{ t('status.completed') }}</option>
        <option value="archived">{{ t('status.archived') }}</option>
      </select>
      <SearchBox :placeholder="t('searchIssueInRange')" @search="onSearch" />
      <div class="toolbar-right">
        <button class="btn btn--primary btn--sm" @click="openCreate">{{ t('addIssue') }}</button>
      </div>
    </div>

    <!-- Issue list -->
    <div class="card-list">
      <IssueCard
        v-for="issue in issues"
        :key="issue.id"
        :issue="issue"
        @edit="openEdit"
        @archive="openArchive"
        @delete="openDelete"
        @view="openView"
      />
      <div v-if="!loading && issues.length === 0" class="empty-state">{{ t('noIssues') }}</div>
    </div>

    <Pager :page="page" :total="total" :page-size="PAGE_SIZE" @page-change="setPage" />

    <!-- Create/Edit modal -->
    <IssueEditModal
      :show="editModalShow"
      :issue="editIssue"
      :mode="editMode"
      @close="editModalShow = false"
      @save="onSave"
    />

    <!-- Archive/Delete confirmation -->
    <Modal
      :show="confirmModalShow"
      :title="confirmAction === 'archive' ? t('archiveIssue') : t('deleteIssue')"
      :confirm-text="t('confirm')"
      :danger="confirmAction === 'delete'"
      @close="confirmModalShow = false"
      @confirm="doConfirm"
    >
      <p>{{ confirmAction === 'archive' ? t('confirmArchiveIssue') : t('confirmDeleteIssue') }}</p>
    </Modal>

    <!-- Archived detail view -->
    <Modal :show="viewModalShow" :title="viewIssue?.title || ''" hide-footer @close="viewModalShow = false" width="700px">
      <template v-if="viewIssue">
        <div v-if="viewIssue.content" class="view-field">
          <div class="view-label">{{ t('issueContent') }}</div>
          <div class="view-value">{{ viewIssue.content }}</div>
        </div>
        <div v-for="field in ['description','investigation','root_cause','solution','test_result','notes']" :key="field">
          <div v-if="viewIssue[field]" class="view-field">
            <div class="view-label">{{ t(`issue${field.charAt(0).toUpperCase() + field.slice(1).replace(/_([a-z])/g, (_, c) => c.toUpperCase())}`) }}</div>
            <div class="view-value">{{ viewIssue[field] }}</div>
          </div>
        </div>
        <div v-if="viewIssue.files_changed?.length" class="view-field">
          <div class="view-label">{{ t('issueFilesChanged') }}</div>
          <div class="view-value"><code v-for="f in viewIssue.files_changed" :key="f" style="display:block">{{ f }}</code></div>
        </div>
        <div v-if="viewIssue.tags?.length" class="view-field">
          <div class="view-label">{{ t('tags') }}</div>
          <div class="view-tags"><span v-for="tag in viewIssue.tags" :key="tag" class="tag">{{ tag }}</span></div>
        </div>
        <div class="view-footer">
          <button class="btn btn--danger" @click="confirmTarget = viewIssue; confirmAction = 'delete'; viewModalShow = false; confirmModalShow = true">{{ t('delete') }}</button>
        </div>
      </template>
    </Modal>
  </div>
</template>

<style scoped>
.issues-view { display: flex; flex-direction: column; flex: 1; }
.page-header { display: flex; align-items: center; gap: 20px; flex-wrap: wrap; }
.page-stats { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.stat-pill {
  display: inline-flex; align-items: baseline; gap: 6px;
  padding: 6px 12px; border-radius: 8px;
  background: rgba(148, 163, 184, 0.08); border: 1px solid var(--border);
  font-size: 13px;
}
.stat-pill__label { color: var(--text-muted); font-weight: 500; }
.stat-pill__num { color: var(--text-heading); font-weight: 600; font-variant-numeric: tabular-nums; }
.stat-pill__unit { color: var(--text-muted); font-size: 12px; }
.stat-pill--highlight { background: rgba(59, 130, 246, 0.08); border-color: rgba(59, 130, 246, 0.25); }
.stat-pill--highlight .stat-pill__num { color: var(--color-primary, #3b82f6); }
.filter-label { font-size: 12px; color: var(--text-muted); margin-right: 6px; }
.filter-sep { margin: 0 4px; color: var(--text-muted); }
.view-field { margin-bottom: 16px; }
.view-label { font-family: var(--font-mono); font-size: 11px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 4px; }
.view-value { font-size: 13px; color: var(--text-heading); white-space: pre-wrap; line-height: 1.6; }
.view-tags { display: flex; gap: 6px; flex-wrap: wrap; }
.view-footer { margin-top: 24px; padding-top: 16px; border-top: 1px solid var(--border); text-align: right; }
</style>
