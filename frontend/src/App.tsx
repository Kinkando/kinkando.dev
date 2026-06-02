import { Routes, Route, Navigate } from 'react-router-dom'
import NavBar from './components/NavBar'
import PortfolioPage from './pages/PortfolioPage'
import LoginPage from './pages/LoginPage'
// import RegisterPage from './pages/RegisterPage'
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
import ProtectedRoute from './auth/ProtectedRoute'

export default function App() {
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
