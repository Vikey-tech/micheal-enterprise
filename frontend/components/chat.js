"use client";
import { useState, useEffect, useRef } from "react";
import Link from "next/link";
import { MessageCircle, X, Send } from "lucide-react";
import { api } from "./api";
import { useApp } from "./site";
export function Conversation({ id }) {
  const { user } = useApp();
  const [messages, setMessages] = useState([]),
    [body, setBody] = useState(""),
    [error, setError] = useState(""),
    [sending, setSending] = useState(false),
    [live, setLive] = useState(false);
  const end = useRef(null);
  useEffect(() => {
    if (!id) return;
    let stopped = false,
      socket,
      reconnect,
      timer;
    async function load() {
      try {
        const v = await api("/conversations/" + id + "/messages");
        if (!stopped) {
          setMessages(v);
          setError("");
        }
      } catch (e) {
        if (!stopped) setError(e.message);
      }
    }
    load();
    timer = setInterval(load, 10000);
    function connect() {
      if (stopped) return;
      socket = new WebSocket(
        `${location.protocol === "https:" ? "wss:" : "ws:"}//${location.host}/api/ws`,
      );
      socket.onopen = () => {
        setLive(true);
        load();
      };
      socket.onmessage = (event) => {
        try {
          const m = JSON.parse(event.data);
          if (m.conversation_id === id)
            setMessages((old) =>
              old.some((x) => x.id === m.id) ? old : [...old, m],
            );
        } catch {}
      };
      socket.onclose = () => {
        setLive(false);
        if (!stopped) reconnect = setTimeout(connect, 4000);
      };
      socket.onerror = () => socket.close();
    }
    connect();
    return () => {
      stopped = true;
      clearInterval(timer);
      clearTimeout(reconnect);
      socket?.close();
    };
  }, [id]);
  useEffect(
    () => end.current?.scrollIntoView({ block: "nearest" }),
    [messages],
  );
  async function send(event) {
    event.preventDefault();
    if (sending || !body.trim()) return;
    setSending(true);
    try {
      const m = await api("/conversations/" + id + "/messages", {
        method: "POST",
        body: { body },
      });
      setMessages((old) =>
        old.some((x) => x.id === m.id) ? old : [...old, m],
      );
      setBody("");
      setError("");
    } catch (e) {
      setError(e.message);
    } finally {
      setSending(false);
    }
  }
  return (
    <>
      <div className="px-4 py-2 text-[11px] muted border-b border-[var(--line)]">
        {live ? "Live connection" : "Checking for replies"} · Messages are saved
        when the team is offline
      </div>
      <div className="messages" role="log" aria-label="Conversation messages">
        {!messages.length && (
          <p className="muted text-sm text-center my-auto">
            Tell us about the car you have in mind. Our team will reply here.
          </p>
        )}
        {messages.map((m) => (
          <div
            key={m.id}
            className={"bubble " + (m.sender_id === user?.id ? "mine" : "")}
          >
            <p className="text-[10px] opacity-70 mb-1">
              {m.role === "customer"
                ? m.sender_name
                : "MICHEAL ENTERPRISE team · " + m.sender_name}
            </p>
            {m.body}
            <p className="text-[9px] opacity-60 mt-2">
              {new Date(m.created_at).toLocaleString()}
            </p>
          </div>
        ))}
        <div ref={end} />
      </div>
      {error && (
        <p role="alert" className="error mx-3 text-xs">
          {error}
        </p>
      )}
      <form
        onSubmit={send}
        className="p-3 flex gap-2 border-t border-[var(--line)]"
      >
        <label className="sr-only" htmlFor={"message-" + id}>
          Your message
        </label>
        <textarea
          id={"message-" + id}
          className="field resize-none"
          rows={2}
          placeholder="Write your message…"
          value={body}
          maxLength={4000}
          required
          onChange={(e) => setBody(e.target.value)}
        />
        <button
          className="btn !px-3"
          disabled={sending || !body.trim()}
          aria-label="Send message"
        >
          <Send size={18} />
        </button>
      </form>
    </>
  );
}
export function ChatWidget() {
  const { user, ready } = useApp();
  const [open, setOpen] = useState(false),
    [id, setID] = useState(""),
    [error, setError] = useState("");
  useEffect(() => {
    if (new URLSearchParams(location.search).has("chat")) setOpen(true);
  }, []);
  useEffect(() => {
    setID("");
    setError("");
    if (open && user?.role === "customer") {
      api("/conversations", { method: "POST", body: {} })
        .then((v) => setID(v.id))
        .catch((e) => setError(e.message));
    }
  }, [open, user?.id]);
  return (
    <>
      {open && (
        <section
          className="chat-window"
          aria-label="Chat with MICHEAL ENTERPRISE"
        >
          <div className="p-5 flex items-center justify-between bg-[var(--button)] text-[var(--button-text)]">
            <div>
              <p className="serif text-lg">Chat with us</p>
              <p className="text-[10px] opacity-70 mt-1">
                MICHEAL ENTERPRISE team
              </p>
            </div>
            <button onClick={() => setOpen(false)} aria-label="Close chat">
              <X size={18} />
            </button>
          </div>
          {!ready ? (
            <p className="p-6">Loading…</p>
          ) : !user ? (
            <div className="p-7 my-auto">
              <h3>A conversation starts here.</h3>
              <p className="muted text-sm my-5">
                Sign in to send messages and keep your conversation history.
              </p>
              <Link className="btn" href="/login?chat=1">
                Sign in to chat
              </Link>
              <Link
                className="block mt-4 underline text-sm"
                href="/register?chat=1"
              >
                Create a customer account
              </Link>
            </div>
          ) : user.role !== "customer" ? (
            <div className="p-7 my-auto">
              <p className="muted mb-4">
                Read and reply to customer messages in your dashboard.
              </p>
              <Link href="/dashboard?tab=chat" className="btn">
                Open inbox
              </Link>
            </div>
          ) : id ? (
            <Conversation id={id} />
          ) : (
            <p className="p-6" role={error ? "alert" : "status"}>
              {error || "Opening your conversation…"}
            </p>
          )}
        </section>
      )}
      <button
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        aria-label={open ? "Close chat" : "Chat with us"}
        className="fixed bottom-6 right-6 z-50 btn !rounded-full shadow-xl !px-5 !py-4"
      >
        <MessageCircle size={20} />
        <span className="text-sm">Chat with us</span>
      </button>
    </>
  );
}
