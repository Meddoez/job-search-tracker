import { StrictMode, useState, useEffect } from "react"
import { createRoot } from "react-dom/client"
import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
  useNavigate,
} from "react-router"

import "./styles.css"
import { Header } from "./components/header"
import { Sidebar } from "./components/sidebar"
import { Home } from "./pages/Home"
import { Jobs } from "./pages/Jobs"
import { CreateProfilePage } from "./pages/CreateProfile"
import { getProfiles } from "./services/profile"
import type { Profile } from "./services/profile"
import { ProfileDetailsPage } from "./pages/ProfileDetails"

function App() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(
    () => window.matchMedia("(min-width: 768px)").matches,
  )
  const [profiles, setProfiles] = useState<Profile[]>([])
  const navigate = useNavigate()

  useEffect(() => {
    void getProfiles().then(setProfiles)
  }, [])

  const handleProfileCreated = (profile: Profile) => {
    setProfiles((current) => [profile, ...current])
    navigate(`/profiles/${profile.id}`)
  }

  return (
    <>
      <Header
        isSidebarOpen={isSidebarOpen}
        onMenuClick={() => setIsSidebarOpen((open) => !open)}
      />
      <Sidebar isOpen={isSidebarOpen} profiles={profiles} />
      <main
        className={`min-h-screen bg-slate-900 px-5 py-8 text-slate-100 transition-[margin] duration-200 md:px-10 md:py-12 ${
          isSidebarOpen ? "md:ml-64" : "md:ml-0"
        }`}
      >
        <div className="mx-auto max-w-6xl">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/jobs" element={<Jobs />} />
            <Route
              path="/profile"
              element={
                <CreateProfilePage onProfileCreated={handleProfileCreated} />
              }
            />
            <Route path="*" element={<Navigate to="/" replace />} />
            <Route path="/profiles/:id" element={<ProfileDetailsPage />} />
          </Routes>
        </div>
      </main>
    </>
  )
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
