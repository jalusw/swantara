"use client";

import { Search } from "lucide-react";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import type { Item } from "@/lib/services/swantara";
import { formatMoney } from "@/lib/utils";

export function PosProductGrid({
  products,
  onAddItem,
  searchPlaceholder,
}: {
  products: Item[];
  onAddItem: (item: Item) => void;
  searchPlaceholder: string;
}) {
  const t = useTranslations("Pos");
  const [search, setSearch] = useState("");

  const filtered = useMemo(() => {
    if (!search) return products;
    const q = search.toLowerCase();
    return products.filter(
      (p) => p.name.toLowerCase().includes(q) || p.type.toLowerCase().includes(q),
    );
  }, [products, search]);

  return (
    <div className="flex flex-col gap-3">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={searchPlaceholder}
          className="h-10 w-full rounded-md border bg-background pl-9 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring"
        />
      </div>

      {filtered.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">{t("noProductsFound")}</p>
      ) : (
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
          {filtered.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => onAddItem(item)}
              className="flex flex-col items-start rounded-lg border bg-card p-3 text-left transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <span className="line-clamp-2 text-sm">{item.name}</span>
              <span className="mt-1 text-xs text-muted-foreground">{item.type}</span>
              <span className="mt-auto pt-2 text-sm font-bold tabular-nums">
                {formatMoney(item.listPrice)}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
