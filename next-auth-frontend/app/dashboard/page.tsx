"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

export default function DashboardPage() {
  const [username, setUsername] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await fetch("http://localhost:8080/protected", {
          method: "GET",
          credentials: "include", 
        });

        if (response.ok) {
          const text = await response.text();
          // Hacky parsing (We will fix this in Go soon!)
          const extractedName = text.replace("Welcome to the secret dashboard, ", "").replace("!\n", "");
          setUsername(extractedName);
        } else {
          router.push("/");
        }
      } catch (error) {
        console.error("Auth check failed", error);
        router.push("/");
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, [router]);

  const handleLogout = async () => {
    try {
      await fetch("http://localhost:8080/logout", {
        method: "POST",
        credentials: "include",
      });
      router.push("/");
    } catch (error) {
      console.error("Logout failed", error);
    }
  };

  // Premium Loading Spinner
  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#0a0a0a]">
        <div className="flex flex-col items-center space-y-4">
          <div className="h-10 w-10 animate-spin rounded-full border-4 border-gray-800 border-t-indigo-500"></div>
          <p className="text-sm font-medium text-gray-400">Authenticating session...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#0a0a0a] font-sans text-gray-100">
      
      {/* Glassmorphism Navbar */}
      <nav className="sticky top-0 z-10 border-b border-gray-800 bg-[#111111]/80 px-6 py-4 backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl items-center justify-between">
          <h1 className="text-xl font-extrabold tracking-tight text-white flex items-center gap-2">
            <span className="h-3 w-3 rounded-full bg-indigo-500"></span>
            Enterprise<span className="text-gray-500">App</span>
          </h1>
          
          <div className="flex items-center space-x-6">
            {/* User Icon & Name */}
            <div className="flex items-center space-x-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 text-lg font-bold text-white shadow-lg">
                {username ? username.charAt(0).toUpperCase() : "U"}
              </div>
              <span className="hidden text-sm font-medium text-gray-300 sm:block">
                {username}
              </span>
            </div>

            {/* Premium Logout Button */}
            <button
              onClick={handleLogout}
              className="rounded-lg px-4 py-2 text-sm font-semibold text-red-400 transition-colors hover:bg-red-500/10 hover:text-red-300 focus:outline-none focus:ring-2 focus:ring-red-500/50"
            >
              Log out
            </button>
          </div>
        </div>
      </nav>

      {/* Main Content Area */}
      <main className="mx-auto mt-10 max-w-6xl px-6 pb-12">
        <div className="grid gap-6 md:grid-cols-3">
          
          {/* Welcome Card (Spans 2 columns) */}
          <div className="col-span-1 md:col-span-2 rounded-2xl border border-gray-800 bg-[#111111] p-8 shadow-2xl transition-all hover:border-gray-700">
            <h2 className="mb-2 text-2xl font-bold text-white">
              Welcome back, <span className="text-indigo-400">{username}</span>! 👋
            </h2>
            <p className="leading-relaxed text-gray-400">
              You have successfully authenticated through your Go backend using an HTTP-Only cookie. 
              Because the cookie is HTTP-Only, JavaScript cannot steal it, making this system incredibly secure against XSS attacks.
            </p>
            
            <div className="mt-8 inline-flex items-center space-x-2 rounded-lg bg-green-900/20 px-4 py-2 text-sm font-medium text-green-400 border border-green-900/50">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"></span>
                <span className="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
              </span>
              <span>Secure Session Active</span>
            </div>
          </div>

          {/* Side Info Card */}
          <div className="col-span-1 rounded-2xl border border-gray-800 bg-gradient-to-br from-[#111111] to-[#1a1a1a] p-8 shadow-2xl">
            <h3 className="mb-4 text-lg font-semibold text-white">System Status</h3>
            <ul className="space-y-4 text-sm text-gray-400">
              <li className="flex justify-between border-b border-gray-800 pb-2">
                <span>Database</span>
                <span className="font-medium text-white">Neon PostgreSQL</span>
              </li>
              <li className="flex justify-between border-b border-gray-800 pb-2">
                <span>Backend</span>
                <span className="font-medium text-white">Go (Golang)</span>
              </li>
              <li className="flex justify-between border-b border-gray-800 pb-2">
                <span>Frontend</span>
                <span className="font-medium text-white">Next.js</span>
              </li>
            </ul>
          </div>
          
        </div>
      </main>
    </div>
  );
}