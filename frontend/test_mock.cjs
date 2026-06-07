const axios = require('axios');
async function test() {
  try {
    const res = await axios.post('http://localhost:3001/api/health-check', {
      url: 'http://localhost:3001/api/mock/123',
      method: 'GET',
      headers: {},
      params: {}
    });
    console.log("Success:", res.data);
  } catch(e) {
    console.error("Error:", e.response ? e.response.status + " " + JSON.stringify(e.response.data) : e.message);
  }
}
test();
