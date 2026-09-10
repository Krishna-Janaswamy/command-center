const axios = require('axios');

async function test() {
  try {
    const res = await axios.get('http://localhost:3001/api/proxy-request', {
      params: {
        url: 'https://jsonplaceholder.typicode.com/todos/1',
        endpoint: '/todos/1',
        allowRealApi: true
      },
      headers: {
        'X-Target-Host': 'https://jsonplaceholder.typicode.com'
      }
    });
    console.log(res.status, res.data);
  } catch(e) {
    console.error(e.response ? e.response.status : e, e.response ? e.response.data : '');
  }
}
test();
