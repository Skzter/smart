import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 50 },
    { duration: '2m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '1m', target: 0 }, // ramp down
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1000'],
  },
  summaryTrendStats: ['avg', 'p(50)', 'p(90)', 'p(95)', 'p(99)'],
};

const payloads = new SharedArray('payloads', () => [
  JSON.parse(open('./payloads/happy.json')),
]);

const baseUrl = __ENV.SUPROXY_URL || 'http://localhost:8080/api/v1/Offerlist';
const destinationOverride = __ENV.SUPROXY_DESTINATION;

const responseTime = new Trend('suproxy_response_time', true);
const responses = new Counter('suproxy_responses');

export default function () {
  const payload = { ...payloads[Math.floor(Math.random() * payloads.length)] };

  if (destinationOverride) {
    payload.destination = destinationOverride;
  }

  const headers = { 'Content-Type': 'application/json', ...payload.header };
  const res = http.post(baseUrl, JSON.stringify(payload), {
    headers,
    tags: { scenario: payload.tags || 'unspecified' },
  });

  responseTime.add(res.timings.duration);
  responses.add(1);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'has body': (r) => r.body && r.body.length > 0,
  });

  sleep(Number(__ENV.K6_SLEEP || '1'));
}

