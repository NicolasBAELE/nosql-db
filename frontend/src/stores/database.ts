import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Document {
  _id: string
  title?: string
  iban?: string
  description?: string
  [key: string]: any
}

export interface Collection {
  name: string
  documents: Document[]
}

export const useDatabaseStore = defineStore('database', () => {
  const collections = ref<Collection[]>([])
  const loading = ref(false)

  // Load all collections
  const loadCollections = async () => {
    try {
      loading.value = true
      const response = await fetch('/api/collections')
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      collections.value = await response.json()
    } catch (error) {
      console.error('Error loading collections:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // Transaction operations
  const beginTransaction = async () => {
    const response = await fetch('/api/transaction/begin', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    })
    
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to begin transaction')
    }
    
    return await response.json()
  }

  const commitTransaction = async (transactionId: string) => {
    const response = await fetch('/api/transaction/commit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ transaction_id: transactionId })
    })
    
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to commit transaction')
    }
    
    return await response.json()
  }

  const rollbackTransaction = async (transactionId: string) => {
    const response = await fetch('/api/transaction/rollback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ transaction_id: transactionId })
    })
    
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to rollback transaction')
    }
    
    return await response.json()
  }

  // Document operations
  const insertDocument = async (transactionId: string, data: { collection: string; document: any }) => {
    const response = await fetch(`/api/transaction/${transactionId}/insert`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
    
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to insert document')
    }
    
    return await response.json()
  }

  const updateDocument = async (transactionId: string, data: { collection: string; document_id: string; updates: any }) => {
    const response = await fetch(`/api/transaction/${transactionId}/update`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
    
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to update document')
    }
    
    return await response.json()
  }

  const deleteDocument = async (transactionId: string, data: { collection: string; document_id: string }) => {
    const response = await fetch(`/api/transaction/${transactionId}/delete`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
    
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to delete document')
    }
  }

  return {
    collections,
    loading,
    loadCollections,
    beginTransaction,
    commitTransaction,
    rollbackTransaction,
    insertDocument,
    updateDocument,
    deleteDocument
  }
}) 