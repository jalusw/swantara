"use client";

import { createAsyncStoragePersister } from "@tanstack/query-async-storage-persister";
import { QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { del, get, set } from "idb-keyval";
import { useState } from "react";

const PERSIST_KEY = "swantara:query-cache";
const MAX_AGE_MS = 1000 * 60 * 60 * 24;

export function TanstackQueryProvider({ children }: { children: React.ReactNode }) {
  const staleTimeInMinutes = 1000 * 60 * 2;
  const gcTimeInMinutes = 1000 * 60 * 30;
  const retriesMaximum = 3;

  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: staleTimeInMinutes,
            gcTime: gcTimeInMinutes,
            retry: retriesMaximum,
            refetchOnWindowFocus: false,
            refetchOnReconnect: true,
            networkMode: "offlineFirst",
          },
          mutations: {
            retry: false,
            networkMode: "online",
          },
        },
      }),
  );

  const [persister] = useState(() =>
    createAsyncStoragePersister({
      storage: {
        getItem: (key) => get(key),
        setItem: (key, value) => set(key, value),
        removeItem: (key) => del(key),
      },
      key: PERSIST_KEY,
      throttleTime: 1000,
    }),
  );

  return (
    <PersistQueryClientProvider
      client={queryClient}
      persistOptions={{
        persister,
        maxAge: MAX_AGE_MS,
        buster: "1",
        dehydrateOptions: {
          shouldDehydrateQuery: (query) => query.state.status === "success",
        },
      }}
    >
      {children}
      {process.env.NODE_ENV === "development" ? (
        <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-left" />
      ) : null}
    </PersistQueryClientProvider>
  );
}
