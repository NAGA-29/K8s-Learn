import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 50 },  // ramp up to 50 VUs
    { duration: '1m',  target: 50 },  // stay at 50 VUs
    { duration: '15s', target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests under 500ms
    http_req_failed:   ['rate<0.01'],  // error rate under 1%
  },
};

export default function () {
  // Health check
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health: status 200': (r) => r.status === 200,
    'health: body contains ok': (r) => r.body.includes('ok'),
  });

  sleep(0.5);

  // Metrics endpoint
  const metricsRes = http.get(`${BASE_URL}/metrics`);
  check(metricsRes, {
    'metrics: status 200': (r) => r.status === 200,
    'metrics: has connections field': (r) => {
      const body = JSON.parse(r.body);
      return body.connections !== undefined;
    },
  });

  sleep(0.5);
}
