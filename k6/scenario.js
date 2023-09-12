import http from 'k6/http';
import { sleep, check } from 'k6';

export let options = {
  stages: [
    { duration: '10s', target: 10 },  
  ],
};

export default function () {
// Step 1: POST /auth
let authPayload = JSON.stringify({ "user_id": 1 });
let authHeaders = { 'Content-Type': 'application/json' };
let authResponse = http.post('http://localhost:8092/auth', authPayload, { headers: authHeaders });
check(authResponse, { 'Auth Request Successful': (r) => r.status === 200 });

let authToken = JSON.parse(authResponse.body).auth_key;

// Step 2: GET /list
let listHeaders = {
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${authToken}`,
};
let listResponse = http.get('http://localhost:8092/list', { headers: listHeaders });
check(listResponse, { 'List Request Successful': (r) => r.status === 200 });

let itemIds = JSON.parse(listResponse.body);
itemIds = itemIds.items

// Step 3: POST /order
let randomItemId = itemIds[Math.floor(Math.random() * itemIds.length)];
let orderPayload = JSON.stringify({ "item_id": randomItemId });
let orderHeaders = {
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${authToken}`,
};
let orderResponse = http.post('http://localhost:8092/order', orderPayload, { headers: orderHeaders });
check(orderResponse, { 'Order Request Successful': (r) => r.status === 200 });

// Add some sleep time to simulate user pacing
// sleep(1);
}
