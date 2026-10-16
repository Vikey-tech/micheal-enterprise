"use client";
import { useEffect, useState } from "react";
import { UserRound, MapPin, Phone, Mail, MessageCircle } from "lucide-react";
import { api } from "../../components/api";
export default function Dealer() {
  const [p, setP] = useState(null),
    [error, setError] = useState("");
  useEffect(() => {
    api("/profile")
      .then(setP)
      .catch((e) => setError(e.message));
  }, []);
  return (
    <div className="shell py-16">
      <p className="eyebrow mb-5">People behind the journey</p>
      <h1 className="!text-5xl mb-12">Meet your dealer.</h1>
      {error ? (
        <p className="error" role="alert">
          {error}
        </p>
      ) : !p ? (
        <div role="status" className="skeleton h-80" />
      ) : (
        <div className="detail-grid grid grid-cols-[1fr_1.5fr] gap-12">
          <div className="car-photo !aspect-square rounded-t-[120px]">
            {p.image ? (
              <img src={p.image} alt={p.name} />
            ) : (
              <UserRound size={140} strokeWidth={0.8} />
            )}
          </div>
          <div className="py-6">
            <p className="eyebrow mb-4">A personal commitment</p>
            <h2>{p.name}</h2>
            <p className="muted leading-8 whitespace-pre-wrap my-7">{p.bio}</p>
            <div className="grid gap-5 text-sm">
              {p.address && (
                <p className="flex gap-3">
                  <MapPin size={18} />
                  {p.address}
                </p>
              )}
              {p.phone && (
                <a href={"tel:" + p.phone} className="flex gap-3">
                  <Phone size={18} />
                  {p.phone}
                </a>
              )}
              {p.email && (
                <a href={"mailto:" + p.email} className="flex gap-3">
                  <Mail size={18} />
                  {p.email}
                </a>
              )}
            </div>
            {p.whatsapp && (
              <a
                href={"https://wa.me/" + p.whatsapp}
                target="_blank"
                rel="noopener noreferrer"
                className="btn mt-9"
              >
                <MessageCircle size={18} /> Speak with us on WhatsApp
              </a>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
