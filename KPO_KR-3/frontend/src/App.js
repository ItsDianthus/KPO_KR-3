const API = 'http://localhost:8080';

// Хелперы
async function api(path, opts) {
    const res = await fetch(API + path, opts);
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
    return res.json().catch(() => null);
}

// UI
document.getElementById('root').innerHTML = `
  <h1>Shop Microservices</h1>
  <section id="accounts">
    <h2>Accounts</h2>
    <input id="acct-id" placeholder="User ID" />
    <button id="create-acct">Create Account</button>
    <button id="get-balance">Get Balance</button>
    <div id="balance"></div>
    <input id="topup-amt" placeholder="Top-up Amount" />
    <button id="topup-btn">Top Up</button>
  </section>
  <hr/>
  <section id="orders">
    <h2>Orders</h2>
    <input id="order-user" placeholder="User ID" />
    <input id="order-amt" placeholder="Amount" />
    <button id="create-order">Create Order</button>
    <button id="list-orders">List Orders</button>
    <pre id="orders-list"></pre>
  </section>
`;

// Привязываем события
document.getElementById('create-acct').onclick = async () => {
    const uid = document.getElementById('acct-id').value;
    await api(`/payments/accounts?user_id=${uid}`, { method: 'POST' });
    alert('Account created');
};
document.getElementById('get-balance').onclick = async () => {
    const uid = document.getElementById('acct-id').value;
    const data = await api(`/payments/accounts/${uid}/balance`);
    document.getElementById('balance').textContent = `Balance: ${data.balance}`;
};
document.getElementById('topup-btn').onclick = async () => {
    const uid = document.getElementById('acct-id').value;
    const amt = document.getElementById('topup-amt').value;
    await api(`/payments/accounts/${uid}/topup?amount=${amt}`, { method: 'POST' });
    alert('Topped up');
};

document.getElementById('create-order').onclick = async () => {
    const uid = document.getElementById('order-user').value;
    const amt = document.getElementById('order-amt').value;
    const res = await fetch(`${API}/orders?user_id=${uid}&amount=${amt}`, { method: 'POST' });
    const id = await res.text();
    alert(`Order created: ${id}`);
};
document.getElementById('list-orders').onclick = async () => {
    const list = await api('/orders');
    document.getElementById('orders-list').textContent = JSON.stringify(list, null, 2);
};
