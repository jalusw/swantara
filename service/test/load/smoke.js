import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080/api/v1';
const EMAIL = __ENV.EMAIL || 'admin@swantara.id';
const PASSWORD = __ENV.PASSWORD || 'password123';

export const options = {
  stages: [
    { duration: '10s', target: 5 },
    { duration: '30s', target: 10 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate < 0.01'],
    http_req_duration: ['p(95) < 500'],
  },
};

export default function () {
  const login = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: EMAIL,
    password: PASSWORD,
  }), { headers: { 'Content-Type': 'application/json' } });

  check(login, { 'login 200': (r) => r.status === 200 });

  const token = login.json('data.access_token');

  if (token) {
    const headers = { Authorization: `Bearer ${token}` };
    check(http.get(`${BASE_URL}/users?size=10`, { headers }), { 'users 200': (r) => r.status === 200 });
  }

  sleep(1);
}