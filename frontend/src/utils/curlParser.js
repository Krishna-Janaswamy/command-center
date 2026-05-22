export const parseCurl = (curlStr) => {
  if (!curlStr || !curlStr.trim().startsWith('curl')) {
    throw new Error('Not a valid cURL command');
  }

  const result = {
    method: 'GET',
    url: '',
    path: '',
    host: '',
    headers: {},
    body: ''
  };

  const extractQuoted = (str, regex) => {
    const matches = [];
    let match;
    while ((match = regex.exec(str)) !== null) {
      matches.push(match[1] || match[2] || match[3] || match[4]);
    }
    return matches.filter(Boolean);
  };

  // Find URL anywhere in string
  const urlMatch = curlStr.match(/https?:\/\/[^\s'"]+/);
  if (urlMatch) {
    result.url = urlMatch[0];
    try {
      const parsedUrl = new URL(result.url);
      result.path = parsedUrl.pathname + parsedUrl.search;
      result.host = parsedUrl.host;
    } catch (e) {}
  }

  const methodMatch = curlStr.match(/-X\s+([A-Z]+)/);
  if (methodMatch) {
    result.method = methodMatch[1];
  }

  const headerRegex = /-H\s+(?:'([^']+)'|"([^"]+)")/g;
  const headers = extractQuoted(curlStr, headerRegex);
  headers.forEach(h => {
    const splitIdx = h.indexOf(':');
    if (splitIdx > 0) {
      const key = h.slice(0, splitIdx).trim();
      const value = h.slice(splitIdx + 1).trim();
      result.headers[key] = value;
    }
  });

  const dataRegex = /(?:-d|--data|--data-raw|--data-binary)\s+(?:'([^']+)'|"([^"]+)")/g;
  const dataMatches = extractQuoted(curlStr, dataRegex);
  if (dataMatches.length > 0) {
    result.body = dataMatches.join('&');
    if (!methodMatch) {
      result.method = 'POST';
    }
  }

  return result;
};
