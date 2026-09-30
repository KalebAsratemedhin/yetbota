"use client";

import Image from "next/image";
import Link from "next/link";
import { Hash, Heart, MapPin } from "lucide-react";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";

type MockPlace = {
  id: string;
  title: string;
  address: string;
  likes: number;
  photo: string;
  tag: string;
};

const MOCK_PLACES: MockPlace[] = [
  {
    id: "1",
    title: "Tomoca Coffee – Piazza",
    address: "Piazza, Addis Ababa",
    likes: 128,
    photo: "/images/places/galani.jpg",
    tag: "coffee",
  },
  {
    id: "2",
    title: "Galani Coffee",
    address: "Bole, Addis Ababa",
    likes: 96,
    photo: "/images/places/galani.jpg",
    tag: "coffee",
  },
  {
    id: "3",
    title: "Yeshi Bunna",
    address: "Kazanchis, Addis Ababa",
    likes: 84,
    photo: "/images/places/yeshi.jpg",
    tag: "coffee",
  },
  {
    id: "4",
    title: "Mercato Spice Market",
    address: "Addis Ketema",
    likes: 210,
    photo: "/images/places/mercato.jpg",
    tag: "markets",
  },
  {
    id: "5",
    title: "Shola Traditional Market",
    address: "Yeka, Addis Ababa",
    likes: 74,
    photo: "/images/places/shola.jpg",
    tag: "markets",
  },
  {
    id: "6",
    title: "Bole Day Market",
    address: "Bole, Addis Ababa",
    likes: 61,
    photo: "/images/places/bole.jpg",
    tag: "markets",
  },
  {
    id: "7",
    title: "Friendship Park",
    address: "Addis Ababa",
    likes: 155,
    photo: "/images/places/friendship.jpg",
    tag: "parks",
  },
  {
    id: "8",
    title: "Unity Park",
    address: "Arat Kilo, Addis Ababa",
    likes: 142,
    photo: "/images/places/unity.jpg",
    tag: "parks",
  },
];

function displayTag(tag: string): string {
  return tag.replace(/[-_]+/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function TrendingPlaceCard({ place }: { place: MockPlace }) {
  return (
    <div className="group block h-full cursor-pointer overflow-hidden rounded-2xl border border-border-subtle bg-bg transition-all duration-300 hover:-translate-y-1.5 hover:border-brand/40 hover:shadow-xl hover:shadow-brand/10">
      <div className="relative h-36 w-full overflow-hidden sm:h-40 md:h-44">
        <Image
          src={place.photo}
          alt={place.title}
          fill
          className="object-cover transition-transform duration-500 group-hover:scale-110"
          sizes="(max-width: 640px) 50vw, (max-width: 768px) 50vw, 33vw"
        />
        <span className="absolute top-2 right-2 z-10 flex items-center gap-1 rounded-full bg-black/60 px-1.5 py-0.5 text-[9px] leading-tight font-medium whitespace-nowrap text-white backdrop-blur-sm">
          <Heart className="h-2.5 w-2.5 fill-current" />
          {place.likes}
        </span>
      </div>
      <div className="p-3">
        <p className="mb-1 truncate text-xs font-semibold text-fg md:text-sm">{place.title}</p>
        <div className="flex items-center gap-1">
          <MapPin className="h-3 w-3 shrink-0 text-fg-faint" />
          <span className="truncate text-[10px] text-fg-faint">{place.address}</span>
        </div>
      </div>
    </div>
  );
}

export default function DiscoverySection() {
  const t = useContent();
  const tags = ["coffee", "markets", "parks"];

  return (
    <section className="bg-brand/5 py-12 dark:bg-surface">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-12">
        <Reveal className="mb-10 text-center">
          <h2 className="mb-1 text-xl font-bold text-fg md:text-2xl">{t.discovery.title}</h2>
          <p className="text-sm text-fg-faint">{t.discovery.subtitle}</p>
        </Reveal>

        {tags.map((tag) => {
          const places = MOCK_PLACES.filter((p) => p.tag === tag);
          return (
            <div key={tag} className="mb-10">
              <div className="mb-4 flex items-center">
                <Hash className="mr-2 h-4 w-4 shrink-0 text-brand" />
                <h3 className="text-sm font-semibold text-fg md:text-base">{displayTag(tag)}</h3>
              </div>
              <div className="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-4">
                {places.map((place, i) => (
                  <Reveal key={place.id} direction="scale" delay={i * 90} className="h-full">
                    <TrendingPlaceCard place={place} />
                  </Reveal>
                ))}
              </div>
            </div>
          );
        })}

        <Reveal direction="scale" className="mt-8 flex justify-center">
          <Link
            href="/places"
            className="inline-flex rounded-full bg-brand px-8 py-2.5 text-sm font-semibold text-black transition-transform hover:scale-105 hover:bg-brand-dark"
          >
            {t.discovery.viewFeed} →
          </Link>
        </Reveal>
      </div>
    </section>
  );
}
