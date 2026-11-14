import { ref } from 'vue'

const notifications = ref([])
let nextId = 1

export function useNotification() {
  const showNotification = (message, type = 'info', timeout = 4000) => {
    const id = nextId++
    const notification = { id, message, type }
    notifications.value.push(notification)

    if (timeout > 0) {
      setTimeout(() => {
        removeNotification(id)
      }, timeout)
    }
  }

  const removeNotification = (id) => {
    notifications.value = notifications.value.filter((n) => n.id !== id)
  }

  return {
    notifications,
    success: (msg, timeout) => showNotification(msg, 'success', timeout),
    error: (msg, timeout) => showNotification(msg, 'error', timeout),
    info: (msg, timeout) => showNotification(msg, 'info', timeout),
    removeNotification
  }
}
