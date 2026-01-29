import { BrowserRouter, Routes, Route } from "react-router-dom";
import HomePage from "./pages/HomePage";
import NotFound from "./pages/NotFound";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        {/* Placeholder routes - implemented in later phases */}
        <Route path="/dashboard" element={<div>Parent Dashboard (coming in US1)</div>} />
        <Route path="/setup" element={<div>Setup Page (coming in US1)</div>} />
        <Route path="/auth/callback" element={<div>OAuth Callback (coming in US1)</div>} />
        <Route path="/child/dashboard" element={<div>Child Dashboard (coming in US4)</div>} />
        <Route path="/:familySlug" element={<div>Family Login (coming in US4)</div>} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
