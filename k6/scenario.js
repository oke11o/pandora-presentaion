import http from 'k6/http';
import {check, sleep} from 'k6';
import {SharedArray} from 'k6/data';

const users = new SharedArray('users', function () {
    return JSON.parse(open('./users.json'));
});

export let options = {
    scenarios: {
        scenario_name: {
            executor: 'constant-vus',
            vus: 100,
            duration: '5s',
        },
    },
};

export default function () {
    // Step 1: POST /auth
    const user = users[Math.floor(Math.random() * users.length)];
    let authPayload = JSON.stringify({"user_id": user.user_id});
    let authHeaders = {'Content-Type': 'application/json'};
    let authResponse = http.post(
        'http://localhost:8092/auth?sleep=100',
        authPayload, {headers: authHeaders}
    );
    check(authResponse,
        {'Auth Request Successful': (r) => r.status === 200}
    );
    sleep(0.1)

    let authToken = JSON.parse(authResponse.body).auth_key;

    // Step 2: GET /list
    let listHeaders = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}`,
    };
    let listResponse = http.get('http://localhost:8092/list?sleep=100',
        {headers: listHeaders}
    );
    check(listResponse,
        {'List Request Successful': (r) => r.status === 200}
    );

    let itemIds = JSON.parse(listResponse.body);
    itemIds = itemIds.items
    sleep(0.1)

    // Step 3: POST /order
    for (let n = 0; n < 3; n++) {
        let randomItemId = itemIds[Math.floor(Math.random() * itemIds.length)];
        let orderPayload = JSON.stringify(
            {"item_id": randomItemId}
        );
        let orderHeaders = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authToken}`,
        };
        let orderResponse = http.post(
            'http://localhost:8092/order?sleep=100',
            orderPayload,
            {headers: orderHeaders}
        );
        check(orderResponse,
            {'Order Request Successful': (r) => r.status === 200}
        );
        sleep(0.1)
    }
}
