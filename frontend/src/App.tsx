import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { TimezoneProvider } from "./context/TimezoneContext";
import { ThemeProvider } from "./context/ThemeContext";
import AuthenticatedLayout from "./components/AuthenticatedLayout";
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
        <Route path="/auth/callback" element={<GoogleCallback />} />
        <Route path="/setup" element={<SetupPage />} />

        {/* Parent routes — shared Layout */}
        <Route element={<AuthenticatedLayout userType="parent" />}>
          <Route path="/dashboard/:childName?" element={<ParentDashboard />} />
          <Route path="/growth/:childName?" element={<GrowthPage />} />
          <Route path="/settings" element={<Navigate to="/settings/general" replace />} />
          <Route path="/settings/children/:childName" element={<SettingsPage />} />
          <Route path="/settings/:category" element={<SettingsPage />} />
        </Route>

        {/* Child routes — shared Layout */}
        <Route element={<AuthenticatedLayout userType="child" />}>
          <Route path="/child/dashboard" element={<ChildDashboard />} />
          <Route path="/child/growth" element={<GrowthPage />} />
          <Route path="/child/settings" element={<ChildSettingsPage />} />
        </Route>

        <Route path="/:familySlug" element={<FamilyLogin />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
    </ThemeProvider>
    </TimezoneProvider>
  );
}

export default App;
