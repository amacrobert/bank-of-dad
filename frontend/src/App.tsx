import { BrowserRouter, Routes, Route } from "react-router-dom";
import HomePage from "./pages/HomePage";
import GoogleCallback from "./pages/GoogleCallback";
import SetupPage from "./pages/SetupPage";
import ParentDashboard from "./pages/ParentDashboard";
import FamilyLogin from "./pages/FamilyLogin";
import ChildDashboard from "./pages/ChildDashboard";
import NotFound from "./pages/NotFound";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/dashboard" element={<ParentDashboard />} />
        <Route path="/setup" element={<SetupPage />} />
        <Route path="/auth/callback" element={<GoogleCallback />} />
        <Route path="/child/dashboard" element={<ChildDashboard />} />
        <Route path="/:familySlug" element={<FamilyLogin />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
