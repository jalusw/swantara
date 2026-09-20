import axios, { type CreateAxiosDefaults } from "axios";
import { getCsrfToken } from "@/lib/constants/cookies";

declare module "axios" {
  interface InternalAxiosRequestConfig {
    __refreshed?: boolean;
  }
}

const config: CreateAxiosDefaults = {
  baseURL: "/api/v1",
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 1000 * 45,
  withCredentials: true,
};

const axiosInstance = axios.create(config);

axiosInstance.interceptors.request.use((cfg) => {
  const csrfToken = getCsrfToken();
  if (csrfToken) {
    cfg.headers.set("x-csrf-token", csrfToken);
  }
  return cfg;
});

axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => Promise.reject(error),
);

export { axiosInstance };
