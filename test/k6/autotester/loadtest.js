import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

export const options = {
  vus: Number(__ENV.K6_VUS || '10'),
  duration: __ENV.K6_DURATION || '1m',
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<40000'], // LLM interactions can be slow
  },
  summaryTrendStats: ['avg', 'p(50)', 'p(90)', 'p(95)', 'p(99)'],
};

const payloads = new SharedArray('payloads', () => [
  JSON.parse(open('./payloads/prompt1.json')),
]);

const baseUrl = __ENV.AUTOTESTER_URL || 'http://localhost:8081/api/v1/chat';

const responseTime = new Trend('autotester_response_time', true);
const responses = new Counter('autotester_responses');

export default function () {
  const payload = { ...payloads[Math.floor(Math.random() * payloads.length)] };

  // Randomize IDs if needed to simulate different chats, but for now fixed is fine
  
  const headers = { 'Content-Type': 'application/json' };
  const res = http.post(baseUrl, JSON.stringify(payload), {
    headers,
  });

  responseTime.add(res.timings.duration);
  responses.add(1);

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(Number(__ENV.K6_SLEEP || '1'));
}

