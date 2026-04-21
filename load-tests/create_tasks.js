import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 20,          // 20 виртуальных пользователей
  duration: '30s',  // 30 секунд
};

export default function () {
  const payload = JSON.stringify({
    title: `Load test task ${__VU}-${__ITER}`,
    description: "k6 load test",
    status: "new",
    recurrence: {
      type: "daily",
      start_date: "2026-04-21T00:00:00Z",
      every_n_days: 1,
      is_active: true
    }
  });

  http.post('http://localhost:8089/api/v1/tasks', payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  sleep(0.1);
}