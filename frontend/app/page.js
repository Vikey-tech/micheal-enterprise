"use client";
import Link from "next/link";
import { useEffect, useState } from "react";
import {
  ArrowUpRight,
  ShieldCheck,
  MessageCircle,
  KeyRound,
  Car,
} from "lucide-react";
import { VehicleCard, Skeletons, Empty } from "../components/site";
import { api } from "../components/api";
export default function Home() {
  const [cars, setCars] = useState(null),
    [error, setError] = useState("");
  useEffect(() => {
    api("/vehicles")
      .then(setCars)
      .catch((e) => setError(e.message));
  }, []);
  return (
    <>
      <section className="shell py-16 lg:py-24">
        <div className="hero-grid grid grid-cols-2 items-center gap-14">
          <div>
            <p className="eyebrow mb-7">Welcome to Micheal Enterprise</p>
            <h1>
              A better drive.
              <br />
              <em className="text-[var(--gold)]">A new chapter.</em>
            </h1>
            <p className="muted max-w-md text-base leading-7 my-8">
              Finding the right car should feel as good as driving it. Explore a
              carefully selected collection, with a team that puts you first.
            </p>
            <div className="flex gap-4 flex-wrap">
              <Link className="btn" href="/vehicles">
                Explore the collection <ArrowUpRight size={18} />
              </Link>
              <Link className="btn secondary" href="/dealer">
                Meet your dealer
              </Link>
            </div>
            <div className="mt-10 flex gap-3 items-center text-xs muted">
              <span className="w-7 border-t border-[var(--gold)]" /> Personal
              service, from first hello to handover.
            </div>
          </div>
          <div className="relative">
            <div className="aspect-[1.1] rounded-t-[150px] rounded-b-lg bg-[var(--line)] overflow-hidden relative flex items-center justify-center">
              {cars?.find((v) => v.featured && v.images?.length)?.images[0] ||
              cars?.find((v) => v.images?.length)?.images[0] ? (
                <img
                  src={
                    (
                      cars.find((v) => v.featured && v.images?.length) ||
                      cars.find((v) => v.images?.length)
                    ).images[0]
                  }
                  alt="A vehicle from our current collection"
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="text-center px-10">
                  <Car
                    strokeWidth={0.8}
                    size={150}
                    className="mx-auto text-[var(--gold)]"
                  />
                  <p className="eyebrow mt-5">A considered collection</p>
                  <p className="serif text-2xl mt-3">
                    Made for your next journey.
                  </p>
                </div>
              )}
              <span className="absolute top-7 right-7 badge">
                Driven by trust
              </span>
            </div>
            <div className="panel absolute -bottom-7 left-6 right-6 flex gap-4 items-center shadow-sm">
              <KeyRound className="text-[var(--gold)]" />
              <div>
                <p className="serif text-xl">Your journey, our commitment.</p>
                <p className="muted text-xs mt-1">
                  Discover. Connect. Drive away with confidence.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>
      <section className="shell mt-12">
        <div className="grid md:grid-cols-3 gap-7 py-9 border-y border-[var(--line)]">
          {[
            [
              ShieldCheck,
              "Carefully selected",
              "Vehicles chosen with attention and care.",
            ],
            [
              MessageCircle,
              "Real people, real answers",
              "Speak directly with our dealership team.",
            ],
            [
              KeyRound,
              "A personal experience",
              "Guidance from enquiry to your next drive.",
            ],
          ].map(([Icon, title, body]) => (
            <div key={title} className="flex gap-4">
              <Icon
                size={24}
                strokeWidth={1.3}
                className="text-[var(--gold)] shrink-0"
              />
              <div>
                <p className="text-sm font-semibold">{title}</p>
                <p className="text-xs muted mt-2">{body}</p>
              </div>
            </div>
          ))}
        </div>
      </section>
      <section className="shell mt-20">
        <div className="flex justify-between items-end mb-9 gap-5">
          <div>
            <p className="eyebrow mb-4">The collection</p>
            <h2>Find your next drive.</h2>
          </div>
          <Link href="/vehicles" className="text-xs flex gap-2 items-center">
            View all vehicles <ArrowUpRight size={16} />
          </Link>
        </div>
        {error ? (
          <p className="error" role="alert">
            {error}
          </p>
        ) : !cars ? (
          <Skeletons />
        ) : cars.length ? (
          <div className="grid-cars">
            {[...cars]
              .sort((a, b) => Number(b.featured) - Number(a.featured))
              .slice(0, 3)
              .map((v) => (
                <VehicleCard key={v.id} vehicle={v} />
              ))}
          </div>
        ) : (
          <Empty />
        )}
      </section>
      <section className="shell mt-20">
        <div className="panel !p-10 md:!p-16 flex justify-between items-center gap-8 flex-wrap">
          <div>
            <p className="eyebrow mb-4">Let's talk cars</p>
            <h2>
              Great service starts
              <br />
              with a conversation.
            </h2>
            <p className="muted mt-5 text-sm">
              Tell us what matters to you. We’ll help you find the right fit.
            </p>
          </div>
          <Link href="/dealer" className="btn">
            Get to know us <ArrowUpRight size={17} />
          </Link>
        </div>
      </section>
    </>
  );
}
