import { Routes, Route } from 'react-router'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import NotFoundPage from './pages/NotFoundPage'
import RedirectToRooms from './pages/RedirectToRooms'

const App = () => {
    return (
        <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="*" element={<NotFoundPage />} />
            <Route path="/" element={<RedirectToRooms />} />
        </Routes>
    )
}

export default App