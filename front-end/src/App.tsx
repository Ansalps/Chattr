import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import AdminDashboard from './pages/AdminDashboard';
import AdminLogin from './pages/AdminLogin';
import LandingPage from './pages/LandingPage';
import UserLogin from './pages/UserLogin';
import UserSignup from './pages/UserSignup';
import VerifyOtp from './pages/VerifyOtp';
import NewsFeedPage from './pages/NewsFeedPage';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<UserLogin />} />
        <Route path="/signup" element={<UserSignup />} />
        <Route path="/verify-otp" element={<VerifyOtp />} />
        <Route path="/newsfeed" element={<NewsFeedPage />} />
        <Route path="/admin/login" element={<AdminLogin />} />
        <Route path="/admin/dashboard" element={<AdminDashboard />} />

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Router>
  );
}

export default App;