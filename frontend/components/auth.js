"use client";
import Link from "next/link";
import { useState } from "react";
import { api } from "./api";
import { useApp } from "./site";
export function Auth({ register = false }) {
  const { setUser } = useApp();
  const [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  async function submit(e) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const u = await api("/auth/" + (register ? "register" : "login"), {
        method: "POST",
        body: Object.fromEntries(new FormData(e.currentTarget)),
      });
      setUser(u);
      const chat = new URLSearchParams(location.search).has("chat");
      location.href =
        u.role === "customer"
          ? chat
            ? "/?chat=1"
            : "/vehicles"
          : "/dashboard";
    } catch (e) {
      setError(e.message);
      setBusy(false);
    }
  }
  return (
    <div className="shell py-16">
      <div className="panel max-w-md mx-auto !p-9">
        <p className="eyebrow mb-4">MICHEAL ENTERPRISE</p>
        <h2 className="!text-3xl">
          {register ? "Start a conversation." : "Welcome back."}
        </h2>
        <p className="muted text-sm mt-4 mb-7">
          {register
            ? "Create a customer account to message our team."
            : "Sign in to continue your journey."}
        </p>
        <form onSubmit={submit} className="grid gap-5">
          {register && (
            <label>
              <span className="label">Your name</span>
              <input
                className="field"
                name="name"
                required
                minLength={2}
                maxLength={100}
                autoComplete="name"
              />
            </label>
          )}
          <label>
            <span className="label">Email address</span>
            <input
              className="field"
              name="email"
              type="email"
              required
              autoComplete="email"
            />
          </label>
          <label>
            <span className="label">Password</span>
            <input
              className="field"
              name="password"
              type="password"
              required
              minLength={register ? 12 : 1}
              maxLength={72}
              autoComplete={register ? "new-password" : "current-password"}
            />
            {register && (
              <span className="text-xs muted mt-2 block">
                Use 12–72 characters. Choose a unique password.
              </span>
            )}
          </label>
          {error && (
            <p role="alert" className="error text-sm">
              {error}
            </p>
          )}
          <button disabled={busy} className="btn">
            {busy ? "Please wait…" : register ? "Create account" : "Sign in"}
          </button>
        </form>
        <p className="text-sm muted mt-6">
          {register ? "Already have an account?" : "New here?"}{" "}
          <Link href={register ? "/login" : "/register"} className="underline">
            {register ? "Sign in" : "Create an account"}
          </Link>
        </p>
        <p className="text-xs muted mt-7">
          For account access help, contact the dealership using the details on
          the dealer page.
        </p>
      </div>
    </div>
  );
}
