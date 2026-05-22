const axios = require('axios');
async function test() {
  try {
    const res = await axios.post('http://localhost:8080/api/health-check', {
      url: 'https://jsonplaceholder.typicode.com/posts',
      method: 'GET',
      headers: {},
      params: {}
    }, {
      headers: {
        'Authorization': `Bearer ${process.env.TOKEN}`
      }
    });
    console.log(res.data);
  } catch(e) {
    console.error(e.response ? e.response.data : e.message);
  }
}
test();
