package com.example.virtualization.service;

import org.springframework.stereotype.Service;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.zip.GZIPInputStream;
import java.util.zip.InflaterInputStream;

@Service
public class HttpForwardService {
    private final HttpClient httpClient;

    public HttpForwardService() {
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(5))
                .build();
    }

    public OutboundResponse send(String url, String method, Map<String, String> headers, String body, Map<String, String> params, int timeoutSeconds) {
        try {
            // Append params to URL
            if (params != null && !params.isEmpty()) {
                StringBuilder query = new StringBuilder(url.contains("?") ? "&" : "?");
                for (Map.Entry<String, String> entry : params.entrySet()) {
                    query.append(entry.getKey()).append("=").append(entry.getValue()).append("&");
                }
                url += query.substring(0, query.length() - 1);
            }

            HttpRequest.Builder requestBuilder = HttpRequest.newBuilder()
                    .uri(URI.create(url))
                    .timeout(Duration.ofSeconds(timeoutSeconds));

            if (headers != null) {
                for (Map.Entry<String, String> entry : headers.entrySet()) {
                    String key = entry.getKey().toLowerCase();
                    // Strip restricted headers
                    if (!key.equals("host") && !key.equals("connection") && !key.equals("content-length")) {
                        requestBuilder.header(entry.getKey(), entry.getValue());
                    }
                }
            }

            HttpRequest.BodyPublisher bodyPublisher = (body != null && !body.isEmpty()) 
                    ? HttpRequest.BodyPublishers.ofString(body) 
                    : HttpRequest.BodyPublishers.noBody();

            requestBuilder.method(method.toUpperCase(), bodyPublisher);

            long startTime = System.currentTimeMillis();
            HttpResponse<byte[]> response = httpClient.send(requestBuilder.build(), HttpResponse.BodyHandlers.ofByteArray());
            long responseTime = System.currentTimeMillis() - startTime;

            Map<String, String> responseHeaders = singleValueHeaders(response.headers().map());
            String responseBody = decompress(response.body(), responseHeaders.get("content-encoding"));

            return new OutboundResponse(response.statusCode(), responseTime, responseBody, responseHeaders, null);

        } catch (Exception e) {
            return new OutboundResponse(-1, -1, null, null, e.getMessage());
        }
    }

    private Map<String, String> singleValueHeaders(Map<String, List<String>> multiValueMap) {
        Map<String, String> map = new HashMap<>();
        for (Map.Entry<String, List<String>> entry : multiValueMap.entrySet()) {
            map.put(entry.getKey().toLowerCase(), String.join(", ", entry.getValue()));
        }
        return map;
    }

    private String decompress(byte[] compressed, String encoding) throws IOException {
        if (compressed == null || compressed.length == 0) return "";
        if (encoding == null) return new String(compressed);

        InputStream is = new ByteArrayInputStream(compressed);
        if (encoding.contains("gzip")) {
            is = new GZIPInputStream(is);
        } else if (encoding.contains("deflate")) {
            is = new InflaterInputStream(is);
        } else {
            return new String(compressed);
        }

        ByteArrayOutputStream baos = new ByteArrayOutputStream();
        byte[] buffer = new byte[1024];
        int len;
        while ((len = is.read(buffer)) > 0) {
            baos.write(buffer, 0, len);
        }
        return baos.toString();
    }

    public static class OutboundResponse {
        public final int statusCode;
        public final long responseTime;
        public final String body;
        public final Map<String, String> headers;
        public final String error;

        public OutboundResponse(int statusCode, long responseTime, String body, Map<String, String> headers, String error) {
            this.statusCode = statusCode;
            this.responseTime = responseTime;
            this.body = body;
            this.headers = headers;
            this.error = error;
        }
    }
}
