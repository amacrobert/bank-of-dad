import { BrowserRouter, Routes, Route } from "react-router-dom";
import { TimezoneProvider } from "./context/TimezoneContext";
import { ThemeProvider } from "./context/ThemeContext";
import HomePage from "./pages/HomePage";
import GoogleCallback from "./pages/GoogleCallback";
import SetupPage from "./pages/SetupPage";
import ParentDashboard from "./pages/ParentDashboard";
import FamilyLogin from "./pages/FamilyLogin";
import ChildDashboard from "./pages/ChildDashboard";
import GrowthPage from "./pages/GrowthPage";
import SettingsPage from "./pages/SettingsPage";
import ChildSettingsPage from "./pages/ChildSettingsPage";
import NotFound from "./pages/NotFound";

function App() {
  return (
    <TimezoneProvider>
    <ThemeProvider>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/dashboard" element={<ParentDashboard />} />
        <Route path="/setup" element={<SetupPage />} />
        <Route path="/auth/callback" element={<GoogleCallback />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="/child/dashboard" element={<ChildDashboard />} />
        <Route path="/child/growth" element={<GrowthPage />} />
        <Route path="/child/settings" element={<ChildSettingsPage />} />
        <Route path="/:familySlug" element={<FamilyLogin />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
    </ThemeProvider>
    </TimezoneProvider>
  );
}

export default App;
