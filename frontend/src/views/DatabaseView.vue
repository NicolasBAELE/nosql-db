<template>
  <div class="database-view">
    <div class="container">
      <div class="header">
        <h1>NoSQL Database Management</h1>
        <p>BDD NoSQL qui permet l'indexation des documents par champs et la gestion des transactions</p>
      </div>

      <!-- Search Bar -->
      <div class="search-panel">
        <div class="search-controls">
          <div class="search-input-group">
            <input 
              type="text" 
              v-model="searchQuery" 
              placeholder="Search documents..." 
              class="search-input"
              @input="performSearch"
            >
            <button 
              class="btn btn-primary search-btn" 
              @click="performSearch"
            >
              Search
            </button>
          </div>
          <div class="search-filters">
            <select v-model="searchField" class="search-select">
              <option value="">All fields</option>
              <option value="title">Title</option>
              <option value="iban">IBAN</option>
              <option value="description">Description</option>
              <option value="name">Name</option>
              <option value="email">Email</option>
              <option value="age">Age</option>
            </select>
            <select v-model="searchCollection" class="search-select">
              <option value="">All collections</option>
              <option v-for="collection in collections" :key="collection.name" :value="collection.name">
                {{ collection.name }}
              </option>
            </select>
          </div>
        </div>
        <div v-if="searchResults.length > 0" class="search-results">
          <h3>Search Results ({{ searchResults.length }} found)</h3>
          <div v-for="result in searchResults" :key="`${result.collection}-${result.document._id}`" class="search-result-item">
            <div class="result-header">
              <span class="result-collection">{{ result.collection }}</span>
              <span class="result-id">ID: {{ result.document._id }}</span>
            </div>
            <pre>{{ JSON.stringify(result.document, null, 2) }}</pre>
            <div class="result-actions">
              <button 
                class="btn btn-warning" 
                @click="showEditForm(result.collection, result.document)" 
                :disabled="!isTransactionActive"
              >
                Edit
              </button>
              <button 
                class="btn btn-danger" 
                @click="showDeleteConfirm(result.collection, result.document._id)" 
                :disabled="!isTransactionActive"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Transaction Panel -->
      <div class="transaction-panel">
        <div class="transaction-status" style="color:black">
          {{ transactionStatus }}
        </div>
        <div class="transaction-actions">
          <button 
            class="btn btn-primary" 
            @click="beginTransaction" 
            :disabled="isTransactionActive"
          >
            Begin Transaction
          </button>
          <button 
            class="btn btn-success" 
            @click="commitTransaction" 
            :disabled="!isTransactionActive"
          >
            Commit Transaction
          </button>
          <button 
            class="btn btn-danger" 
            @click="rollbackTransaction" 
            :disabled="!isTransactionActive"
          >
            Rollback Transaction
          </button>
        </div>
      </div>

      <!-- Alert Messages -->
      <div v-for="alert in alerts" :key="alert.id" :class="['alert', `alert-${alert.type}`]">
        {{ alert.message }}
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading">
        Loading collections...
      </div>

      <!-- Collections -->
      <div v-else>
        <div v-for="collection in collections" :key="collection.name" class="collection">
          <h2>Collection: {{ collection.name }}</h2>
          
          <!-- Add Document Button -->
          <div class="form-group">
            <button 
              class="btn btn-primary" 
              @click="showAddForm(collection.name)" 
              :disabled="!isTransactionActive"
            >
              Add New Document
            </button>
          </div>

          <!-- Documents -->
          <div v-if="collection.documents && collection.documents.length > 0">
            <h3>Documents:</h3>
            <div v-for="document in collection.documents" :key="document._id" class="document">
              <pre>{{ JSON.stringify(document, null, 2) }}</pre>
              <div class="document-actions">
                <button 
                  class="btn btn-warning" 
                  @click="showEditForm(collection.name, document)" 
                  :disabled="!isTransactionActive"
                >
                  Edit
                </button>
                <button 
                  class="btn btn-danger" 
                  @click="showDeleteConfirm(collection.name, document._id)" 
                  :disabled="!isTransactionActive"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
          <div v-else>
            <p style="color:black">No documents in this collection</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Document Modal -->
    <Modal v-model:show="showAddModal" title="Add New Document">
      <form @submit.prevent="addDocument">
        <div class="form-group">
          <label for="addCollection">Collection:</label>
          <select id="addCollection" v-model="addForm.collection" required>
            <option value="">Select a collection</option>
            <option v-for="collection in collections" :key="collection.name" :value="collection.name">
              {{ collection.name }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label for="addTitle">Title:</label>
          <input type="text" id="addTitle" v-model="addForm.title" required>
        </div>
        <div class="form-group">
          <label for="addIban">IBAN:</label>
          <input type="text" id="addIban" v-model="addForm.iban" required>
        </div>
        <div class="form-group">
          <label for="addDescription">Description:</label>
          <textarea id="addDescription" v-model="addForm.description"></textarea>
        </div>
        <div class="form-group">
          <button type="submit" class="btn btn-success">Add Document</button>
          <button type="button" class="btn btn-secondary" @click="closeAddModal">Cancel</button>
        </div>
      </form>
    </Modal>

    <!-- Edit Document Modal -->
    <Modal v-model:show="showEditModal" title="Edit Document">
      <form @submit.prevent="updateDocument">
        <div class="form-group">
          <label for="editCollection">Collection:</label>
          <input type="text" id="editCollection" v-model="editForm.collection" readonly>
        </div>
        <div class="form-group">
          <label for="editId">Document ID:</label>
          <input type="text" id="editId" v-model="editForm.id" readonly>
        </div>
        <div class="form-group">
          <label for="editTitle">Title:</label>
          <input type="text" id="editTitle" v-model="editForm.title" required>
        </div>
        <div class="form-group">
          <label for="editIban">IBAN:</label>
          <input type="text" id="editIban" v-model="editForm.iban" required>
        </div>
        <div class="form-group">
          <label for="editDescription">Description:</label>
          <textarea id="editDescription" v-model="editForm.description"></textarea>
        </div>
        <div class="form-group">
          <button type="submit" class="btn btn-success">Update Document</button>
          <button type="button" class="btn btn-secondary" @click="closeEditModal">Cancel</button>
        </div>
      </form>
    </Modal>

    <!-- Delete Confirmation Modal -->
    <Modal v-model:show="showDeleteModal" title="Confirm Delete">
      <p>Are you sure you want to delete this document?</p>
      <p><strong>Collection:</strong> {{ deleteForm.collection }}</p>
      <p><strong>Document ID:</strong> {{ deleteForm.id }}</p>
      <div class="form-group">
        <button class="btn btn-danger" @click="deleteDocument">Delete</button>
        <button class="btn btn-secondary" @click="closeDeleteModal">Cancel</button>
      </div>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import Modal from '@/components/Modal.vue'
import { useDatabaseStore } from '@/stores/database'

// Store
const databaseStore = useDatabaseStore()

// Reactive data
const loading = ref(true)
const currentTransactionId = ref<string | null>(null)
const pendingOperations = ref<any[]>([])

// Search functionality
const searchQuery = ref('')
const searchField = ref('')
const searchCollection = ref('')
const searchResults = ref<Array<{ collection: string; document: any }>>([])

// Modal states
const showAddModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)

// Form data
const addForm = ref({
  collection: '',
  title: '',
  iban: '',
  description: ''
})

const editForm = ref({
  collection: '',
  id: '',
  title: '',
  iban: '',
  description: ''
})

const deleteForm = ref({
  collection: '',
  id: ''
})

// Alerts
const alerts = ref<Array<{ id: number; message: string; type: string }>>([])
const alertCounter = ref(0)

// Computed
const isTransactionActive = computed(() => currentTransactionId.value !== null)

const transactionStatus = computed(() => {
  return isTransactionActive.value 
    ? `Transaction active: ${currentTransactionId.value}`
    : 'No active transaction'
})

const collections = computed(() => databaseStore.collections)

// Methods
const loadCollections = async () => {
  try {
    loading.value = true
    await databaseStore.loadCollections()
  } catch (error) {
    showAlert('Error loading collections: ' + (error as Error).message, 'error')
  } finally {
    loading.value = false
  }
}

// Transaction Management
const beginTransaction = async () => {
  try {
    const result = await databaseStore.beginTransaction()
    currentTransactionId.value = result.transaction_id
    showAlert('Transaction started successfully', 'success')
  } catch (error) {
    showAlert('Error starting transaction: ' + (error as Error).message, 'error')
  }
}

const commitTransaction = async () => {
  if (!currentTransactionId.value) {
    showAlert('No active transaction', 'error')
    return
  }

  try {
    await databaseStore.commitTransaction(currentTransactionId.value)
    showAlert('Transaction committed successfully', 'success')
    resetTransaction()
    await loadCollections() // Refresh data
  } catch (error) {
    showAlert('Error committing transaction: ' + (error as Error).message, 'error')
  }
}

const rollbackTransaction = async () => {
  if (!currentTransactionId.value) {
    showAlert('No active transaction', 'error')
    return
  }

  try {
    await databaseStore.rollbackTransaction(currentTransactionId.value)
    showAlert('Transaction rolled back successfully', 'success')
    resetTransaction()
  } catch (error) {
    showAlert('Error rolling back transaction: ' + (error as Error).message, 'error')
  }
}

const resetTransaction = () => {
  currentTransactionId.value = null
  pendingOperations.value = []
}

// Form Management
const showAddForm = (collectionName: string) => {
  if (!isTransactionActive.value) {
    showAlert('Please start a transaction first', 'info')
    return
  }
  
  addForm.value = {
    collection: collectionName, // Pre-select the collection that was clicked
    title: '',
    iban: '',
    description: ''
  }
  showAddModal.value = true
}

const showEditForm = (collectionName: string, document: any) => {
  if (!isTransactionActive.value) {
    showAlert('Please start a transaction first', 'info')
    return
  }
  
  editForm.value = {
    collection: collectionName,
    id: document._id,
    title: document.title || '',
    iban: document.iban || '',
    description: document.description || ''
  }
  showEditModal.value = true
}

const showDeleteConfirm = (collectionName: string, documentId: string) => {
  if (!isTransactionActive.value) {
    showAlert('Please start a transaction first', 'info')
    return
  }
  
  deleteForm.value = {
    collection: collectionName,
    id: documentId
  }
  showDeleteModal.value = true
}

const closeAddModal = () => {
  showAddModal.value = false
}

const closeEditModal = () => {
  showEditModal.value = false
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
}

// Form Submissions
const addDocument = async () => {
  const formData = {
    collection: addForm.value.collection,
    document: {
      title: addForm.value.title,
      iban: addForm.value.iban,
      description: addForm.value.description
    }
  }

  try {
    const result = await databaseStore.insertDocument(currentTransactionId.value!, formData)
    showAlert('Document added to transaction successfully', 'success')
    closeAddModal()
    pendingOperations.value.push({ type: 'add', data: result })
  } catch (error) {
    showAlert('Error adding document: ' + (error as Error).message, 'error')
  }
}

const updateDocument = async () => {
  const formData = {
    collection: editForm.value.collection,
    document_id: editForm.value.id,
    updates: {
      title: editForm.value.title,
      iban: editForm.value.iban,
      description: editForm.value.description
    }
  }

  try {
    const result = await databaseStore.updateDocument(currentTransactionId.value!, formData)
    showAlert('Document updated in transaction successfully', 'success')
    closeEditModal()
    pendingOperations.value.push({ type: 'update', data: result })
  } catch (error) {
    showAlert('Error updating document: ' + (error as Error).message, 'error')
  }
}

const deleteDocument = async () => {
  const formData = {
    collection: deleteForm.value.collection,
    document_id: deleteForm.value.id
  }

  try {
    await databaseStore.deleteDocument(currentTransactionId.value!, formData)
    showAlert('Document deleted from transaction successfully', 'success')
    closeDeleteModal()
    pendingOperations.value.push({ type: 'delete', data: formData })
  } catch (error) {
    showAlert('Error deleting document: ' + (error as Error).message, 'error')
  }
}

// Search Functions
const performSearch = () => {
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    return
  }

  const query = searchQuery.value.toLowerCase().trim()
  const results: Array<{ collection: string; document: any }> = []

  collections.value.forEach(collection => {
    // Filter by collection if specified
    if (searchCollection.value && collection.name !== searchCollection.value) {
      return
    }

    collection.documents.forEach(document => {
      let matchFound = false

      if (searchField.value) {
        // Search in specific field
        const fieldValue = document[searchField.value]
        if (fieldValue && String(fieldValue).toLowerCase().includes(query)) {
          matchFound = true
        }
      } else {
        // Search in all fields
        for (const [key, value] of Object.entries(document)) {
          if (key === '_id') continue // Skip document ID
          if (value && String(value).toLowerCase().includes(query)) {
            matchFound = true
            break
          }
        }
      }

      if (matchFound) {
        results.push({
          collection: collection.name,
          document: document
        })
      }
    })
  })

  searchResults.value = results
}

// Utility Functions
const showAlert = (message: string, type: string) => {
  const alert = {
    id: alertCounter.value++,
    message: message,
    type: type
  }
  
  alerts.value.push(alert)
  
  setTimeout(() => {
    alerts.value = alerts.value.filter(a => a.id !== alert.id)
  }, 5000)
}

// Lifecycle
onMounted(() => {
  loadCollections()
})
</script>

<style scoped>
.database-view {
  width: 100%;
  padding: 20px;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  background: #2c3e50;
  color: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.collection {
  background: white;
  border: 1px solid #ddd;
  padding: 20px;
  margin-bottom: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.collection h2 {
  margin-top: 0;
  color: #2c3e50;
  border-bottom: 2px solid #3498db;
  padding-bottom: 10px;
}

.document {
  background: #f8f9fa;
  padding: 15px;
  margin: 10px 0;
  border-radius: 5px;
  border-left: 4px solid #3498db;
}

.document-actions {
  margin-top: 10px;
}

.btn {
  padding: 8px 16px;
  margin: 2px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary { background: #3498db; color: white; }
.btn-success { background: #27ae60; color: white; }
.btn-warning { background: #f39c12; color: white; }
.btn-danger { background: #e74c3c; color: white; }
.btn-secondary { background: #95a5a6; color: white; }
.btn:hover { opacity: 0.8; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
  color: #2c3e50;
}

.form-group input, .form-group textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-sizing: border-box;
}

.form-group textarea {
  height: 80px;
  resize: vertical;
}

.transaction-panel {
  background: #ecf0f1;
  padding: 15px;
  border-radius: 5px;
  margin-bottom: 20px;
  border-left: 4px solid #e67e22;
}

.transaction-status {
  font-weight: bold;
  margin-bottom: 10px;
}

.transaction-actions {
  display: flex;
  gap: 10px;
}

.alert {
  padding: 10px;
  margin: 10px 0;
  border-radius: 4px;
}

.alert-success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
.alert-error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
.alert-info { background: #d1ecf1; color: #0c5460; border: 1px solid #bee5eb; }

.loading {
  text-align: center;
  padding: 20px;
  color: #666;
}

pre {
  background: #2c3e50;
  color: #ecf0f1;
  padding: 15px;
  border-radius: 5px;
  overflow-x: auto;
  font-size: 12px;
}

/* Search Panel Styles */
.search-panel {
  background: white;
  border: 1px solid #ddd;
  padding: 20px;
  margin-bottom: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.search-controls {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.search-input-group {
  display: flex;
  gap: 10px;
}

.search-input {
  flex: 1;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.search-btn {
  padding: 10px 20px;
  white-space: nowrap;
}

.search-filters {
  display: flex;
  gap: 10px;
}

.search-select {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  font-size: 14px;
}

.search-results {
  margin-top: 20px;
  border-top: 1px solid #eee;
  padding-top: 20px;
}

.search-results h3 {
  color: #2c3e50;
  margin-bottom: 15px;
}

.search-result-item {
  background: #f8f9fa;
  padding: 15px;
  margin: 10px 0;
  border-radius: 5px;
  border-left: 4px solid #3498db;
}

.result-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  font-weight: bold;
}

.result-collection {
  color: #3498db;
  background: #e3f2fd;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 12px;
}

.result-id {
  color: #666;
  font-size: 12px;
}

.result-actions {
  margin-top: 10px;
}

@media (max-width: 768px) {
  .search-input-group {
    flex-direction: column;
  }
  
  .search-filters {
    flex-direction: column;
  }
  
  .result-header {
    flex-direction: column;
    gap: 5px;
  }
}
</style> 