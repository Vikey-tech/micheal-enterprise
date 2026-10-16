"use client";
import { use, useEffect, useState } from "react";
import Link from "next/link";
import { Car, ArrowLeft, MessageCircle } from "lucide-react";
import { api, money } from "../../../components/api";
import { useApp } from "../../../components/site";
export default function Detail({ params }) {
  const { id } = use(params);
  const { notify } = useApp();
  const [car, setCar] = useState(null),
    [profile, setProfile] = useState(null),
    [active, setActive] = useState(0),
    [error, setError] = useState(""),
    [sending, setSending] = useState(false);
  useEffect(() => {
    Promise.all([api("/vehicles/" + id), api("/profile")])
      .then(([v, p]) => {
        setCar(v);
        setProfile(p);
      })
      .catch((e) => setError(e.message));
  }, [id]);
  async function enquire(e) {
    e.preventDefault();
    setSending(true);
    const form = e.currentTarget;
    try {
      await api("/enquiries", {
        method: "POST",
        body: { ...Object.fromEntries(new FormData(form)), vehicle_id: id },
      });
      form.reset();
      notify("Enquiry received. Our team will contact you.");
    } catch (e) {
      notify(e.message);
    } finally {
      setSending(false);
    }
  }
  if (error)
    return (
      <div className="shell py-16">
        <p role="alert" className="error">
          {error}
        </p>
        <Link href="/vehicles" className="btn mt-6">
          Back to collection
        </Link>
      </div>
    );
  if (!car)
    return (
      <div className="shell py-16" role="status">
        <div className="skeleton h-96" />
      </div>
    );
  return (
    <div className="shell py-12">
      <Link
        href="/vehicles"
        className="inline-flex gap-2 text-xs mb-9 items-center"
      >
        <ArrowLeft size={15} /> Back to collection
      </Link>
      <div className="detail-grid grid grid-cols-[1.4fr_1fr] gap-10">
        <div>
          <div className="car-photo !aspect-[1.4] rounded-lg">
            {car.images[active] ? (
              <img
                src={car.images[active]}
                alt={car.title + " — photo " + (active + 1)}
              />
            ) : (
              <Car size={120} strokeWidth={1} />
            )}
          </div>
          <div className="flex gap-3 mt-4 overflow-auto">
            {car.images.map((src, i) => (
              <button
                key={src}
                onClick={() => setActive(i)}
                aria-label={"Show vehicle photo " + (i + 1)}
                aria-pressed={active === i}
                className={
                  "w-24 h-16 shrink-0 rounded border-2 " +
                  (active === i ? "border-[var(--gold)]" : "border-transparent")
                }
              >
                <img
                  src={src}
                  alt=""
                  className="w-full h-full object-cover rounded"
                />
              </button>
            ))}
          </div>
          <section className="mt-10">
            <h3>About this vehicle</h3>
            <p className="muted text-sm leading-7 whitespace-pre-wrap mt-4">
              {car.description ||
                "Contact our team for more information about this vehicle."}
            </p>
          </section>
        </div>
        <div>
          <span className="badge mb-5">{car.status}</span>
          <p className="eyebrow mb-3">
            {car.make} · {car.model}
          </p>
          <h1 className="!text-4xl">{car.title}</h1>
          <p className="text-3xl my-7">{money(car.price)}</p>
          <div className="grid grid-cols-2 gap-5 border-y border-[var(--line)] py-6">
            {[
              ["Year", car.year],
              ["Mileage", car.mileage.toLocaleString() + " km"],
              ["Transmission", car.transmission],
              ["Fuel type", car.fuel],
            ].map(([k, v]) => (
              <div key={k}>
                <p className="eyebrow !text-[9px]">{k}</p>
                <p className="mt-2 text-sm">{v}</p>
              </div>
            ))}
          </div>
          {profile?.whatsapp && (
            <a
              href={`https://wa.me/${profile.whatsapp}?text=${encodeURIComponent("Hello MICHEAL ENTERPRISE, I am interested in " + car.title + " " + location.href)}`}
              target="_blank"
              rel="noopener noreferrer"
              className="btn w-full mt-6"
            >
              <MessageCircle size={18} /> Enquire on WhatsApp
            </a>
          )}
          <div className="panel mt-6">
            <h3>Interested? Let’s talk.</h3>
            <p className="text-xs muted my-4">
              Send an enquiry and our team will get back to you. Availability is
              confirmed by the dealership.
            </p>
            <form onSubmit={enquire} className="grid gap-4">
              {[
                ["name", "Your name", "text"],
                ["email", "Email address", "email"],
                ["phone", "Phone number", "tel"],
              ].map(([name, label, type]) => (
                <label key={name}>
                  <span className="label">{label}</span>
                  <input
                    className="field"
                    name={name}
                    type={type}
                    required={name !== "phone"}
                    maxLength={name === "phone" ? 30 : 100}
                  />
                </label>
              ))}
              <label>
                <span className="label">Your enquiry</span>
                <textarea
                  className="field"
                  name="body"
                  required
                  minLength={3}
                  maxLength={4000}
                  rows={3}
                  defaultValue={
                    "I would like to know more about " + car.title + "."
                  }
                />
              </label>
              <button className="btn" disabled={sending}>
                {sending ? "Sending…" : "Send enquiry"}
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
