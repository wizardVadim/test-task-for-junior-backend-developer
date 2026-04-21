import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 50,
  duration: '30s',
};

export default function () {
  http.get('http://localhost:8089/api/v1/occurrences?date=2026-04-22');
  sleep(0.05);
}