import { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import NavBar from './components/NavBar'
import PortfolioPage from './pages/PortfolioPage'
import LoginPage from './pages/LoginPage'
// import RegisterPage from './pages/RegisterPage'
import AccountPage from './pages/AccountPage'
import NotificationsPage from './pages/NotificationsPage'
import FinancePage from './pages/FinancePage'
import KanbanPage from './pages/KanbanPage'
import KanbanArchivePage from './pages/KanbanArchivePage'
import ChatPage from './pages/ChatPage'
import HealthPage from './pages/HealthPage'
import WorkoutPage from './pages/WorkoutPage'
import MedicinePage from './pages/MedicinePage'
import FoodPage from './pages/FoodPage'
import SleepPage from './pages/SleepPage'
import NewsPage from './pages/NewsPage'
import QuestPage from './pages/QuestPage'
import ProtectedRoute from './auth/ProtectedRoute'
import {
  getCurrentToken,
  isPushSupported,
  onForegroundMessage,
  showLocalNotification,
} from './lib/messaging'
import { registerPushToken } from './lib/api/notifications'

export default function App() {
  // Wire a single global foreground message listener so push notifications
  // display on every page when the tab is in the foreground.
  // Also silently re-registers the current FCM token so the backend stays in
  // sync after browser data clears or FCM rotates the token.
  useEffect(() => {
    if (!isPushSupported() || Notification.permission !== 'granted') return
    const cleanup = onForegroundMessage(showLocalNotification)
    getCurrentToken().then((token) => {
      if (token) registerPushToken(token).catch(() => undefined)
    })
    return cleanup
  }, [])

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      <NavBar />
      <Routes>
        <Route path="/" element={<Navigate to="/portfolio" replace />} />
        <Route path="/portfolio" element={<PortfolioPage />} />
        <Route path="/news" element={<NewsPage />} />
        <Route path="/login" element={<LoginPage />} />
        {/* <Route path="/register" element={<RegisterPage />} /> */}
        <Route
          path="/account"
          element={
            <ProtectedRoute>
              <AccountPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/notifications"
          element={
            <ProtectedRoute>
              <NotificationsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/kanban"
          element={
            <ProtectedRoute>
              <KanbanPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/kanban/archive"
          element={
            <ProtectedRoute>
              <KanbanArchivePage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/finance"
          element={
            <ProtectedRoute>
              <FinancePage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/chat"
          element={
            <ProtectedRoute>
              <ChatPage />
            </ProtectedRoute>
          }
        />
        {/* Health group */}
        <Route
          path="/health"
          element={
            <ProtectedRoute>
              <HealthPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/health/workout"
          element={
            <ProtectedRoute>
              <WorkoutPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/health/medicine"
          element={
            <ProtectedRoute>
              <MedicinePage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/health/food"
          element={
            <ProtectedRoute>
              <FoodPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/health/sleep"
          element={
            <ProtectedRoute>
              <SleepPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/quest"
          element={
            <ProtectedRoute>
              <QuestPage />
            </ProtectedRoute>
          }
        />
        {/* Legacy redirects for moved routes */}
        <Route
          path="/workout"
          element={<Navigate to="/health/workout" replace />}
        />
        <Route
          path="/medicine"
          element={<Navigate to="/health/medicine" replace />}
        />
      </Routes>
    </div>
  )
}
