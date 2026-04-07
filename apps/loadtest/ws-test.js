import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

const WS_URL = __ENV.WS_URL || 'ws://localhost:8080/ws';

const wsMessages = new Counter('ws_messages_received');
const wsSendErrors = new Counter('ws_send_errors');
const wsConnectTime = new Trend('ws_connect_time', true);

export const options = {
  stages: [
    { duration: '30s', target: 100 },  // ramp up to 100 connections
    { duration: '2m',  target: 100 },  // hold 100 connections
    { duration: '15s', target: 0 },    // ramp down
  ],
  thresholds: {
    ws_connect_time: ['p(95)<2000'],   // 95% connect under 2s
    ws_send_errors:  ['count<10'],     // fewer than 10 send errors
  },
};

export default function () {
  const connectStart = Date.now();

  const res = ws.connect(WS_URL, {}, function (socket) {
    const connectDuration = Date.now() - connectStart;
    wsConnectTime.add(connectDuration);

    socket.on('open', function () {
      // Send a message every 2 seconds
      socket.setInterval(function () {
        const msg = JSON.stringify({
          type: 'message',
          data: `Load test message from VU ${__VU} iter ${__ITER}`,
        });
        try {
          socket.send(msg);
        } catch (e) {
          wsSendErrors.add(1);
        }
      }, 2000);
    });

    socket.on('message', function (data) {
      wsMessages.add(1);
      // Validate that the message is valid JSON
      check(data, {
        'message is valid JSON': (d) => {
          try { JSON.parse(d); return true; } catch(e) { return false; }
        },
      });
    });

    socket.on('error', function (e) {
      console.error('WebSocket error:', e.error());
    });

    // Keep connection open for the duration of the iteration
    socket.setTimeout(function () {
      socket.close();
    }, 30000);
  });

  check(res, {
    'WebSocket status is 101': (r) => r && r.status === 101,
  });

  sleep(1);
}
