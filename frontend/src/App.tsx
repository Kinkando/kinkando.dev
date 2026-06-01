import { Routes, Route, Navigate } from 'react-router-dom'
import NavBar from './components/NavBar'
import PortfolioPage from './pages/PortfolioPage'
import LoginPage from './pages/LoginPage'
// import RegisterPage from './pages/RegisterPage'
import FinancePage from './pages/FinancePage'
import KanbanPage from './pages/KanbanPage'
import KanbanArchivePage from './pages/KanbanArchivePage'
import ChatPage from './pages/ChatPage'
import ProtectedRoute from './auth/ProtectedRoute'

export default function App() {
  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      <NavBar />
      <Routes>
        <Route path="/" element={<Navigate to="/portfolio" replace />} />
        <Route path="/portfolio" element={<PortfolioPage />} />
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
      </Routes>
    </div>
  )
}
