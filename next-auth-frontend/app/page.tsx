"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function AuthPage() {
  const [isLogin, setIsLogin] = useState(true);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setMessage(""); 

    const endpoint = isLogin ? "/login" : "/register";
    
    // ENTERPRISE BEST PRACTICE: Use environment variables, fallback to live URL
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "https://golang-auth-qc7d.onrender.com";

    try {
      const response = await fetch(`${API_URL}${endpoint}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
        credentials: "include", // CRITICAL: This allows cross-site cookies to be saved!
      });

      const text = await response.text();

      if (response.ok) {
        if (isLogin) {
          router.push("/dashboard");
        } else {
          setMessage("✨ Registration successful! Please log in.");
          setIsLogin(true); 
          setPassword(""); 
        }
      } else {
        setMessage(`❌ Error: ${text}`);
      }
    } catch (error) {
      setMessage("❌ Failed to connect to the server.");
    } finally {
      setIsLoading(false); 
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#0a0a0a] px-4 font-sans text-gray-100">
      <div className="w-full max-w-md rounded-2xl border border-gray-800 bg-[#111111] p-8 shadow-2xl transition-all">
        
        {/* Header Section */}
        <div className="mb-8 text-center">
          <h2 className="text-3xl font-extrabold tracking-tight text-white">
            {isLogin ? "Welcome back" : "Create an account"}
          </h2>
          <p className="mt-2 text-sm text-gray-400">
            {isLogin ? "Enter your details to access your dashboard." : "Sign up to get started with our enterprise app."}
          </p>
        </div>

        {/* Form Section */}
        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-gray-300">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={isLoading}
              className="w-full rounded-lg border border-gray-700 bg-gray-900 px-4 py-3 text-white placeholder-gray-500 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 disabled:opacity-50"
              placeholder="Enter your username"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-gray-300">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isLoading}
              className="w-full rounded-lg border border-gray-700 bg-gray-900 px-4 py-3 text-white placeholder-gray-500 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 disabled:opacity-50"
              placeholder="••••••••"
              required
            />
          </div>
          <button
            type="submit"
            disabled={isLoading}
            className="mt-2 w-full rounded-lg bg-indigo-600 px-4 py-3 font-semibold text-white transition-all hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-900 disabled:cursor-not-allowed disabled:opacity-70"
          >
            {isLoading ? "Processing..." : isLogin ? "Sign In" : "Sign Up"}
          </button>
        </form>

        {/* Dynamic Message Box */}
        {message && (
          <div className={`mt-6 rounded-lg p-3 text-center text-sm font-medium ${message.includes("✨") ? "border border-green-800/50 bg-green-900/30 text-green-400" : "border border-red-800/50 bg-red-900/30 text-red-400"}`}>
            {message}
          </div>
        )}

        {/* Toggle Login/Register */}
        <div className="mt-6 text-center text-sm text-gray-400">
          {isLogin ? "Don't have an account? " : "Already have an account? "}
          <button
            type="button"
            onClick={() => {
              setIsLogin(!isLogin);
              setMessage("");
              setPassword(""); 
            }}
            className="font-semibold text-indigo-400 transition-colors hover:text-indigo-300 hover:underline"
          >
            {isLogin ? "Sign up" : "Log in"}
          </button>
        </div>
        
      </div>
    </div>
  );
}