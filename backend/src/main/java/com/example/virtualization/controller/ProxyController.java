package com.example.virtualization.controller;

import com.example.virtualization.model.Stub;
import com.example.virtualization.model.User;
import com.example.virtualization.service.HttpForwardService;
import com.example.virtualization.service.JwtService;
import com.example.virtualization.stub.service.StubMatchingService;
import com.example.virtualization.stub.service.StubService;
import com.example.virtualization.util.SecurityHelper;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.net.InetAddress;
import java.net.URI;
import java.net.UnknownHostException;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("/api")
@CrossOrigin(origins = "*")
public class ProxyController {

    private final HttpForwardService httpForwardService;
    private final StubMatchingService stubMatchingService;
    private final StubService stubService;
    private final JwtService jwtService;
    private final JdbcTemplate jdbc;

    public ProxyController(HttpForwardService httpForwardService, StubMatchingService stubMatchingService, StubService stubService, JwtService jwtService, JdbcTemplate jdbc) {
        this.httpForwardService = httpForwardService;
        this.stubMatchingService = stubMatchingService;
        this.stubService = stubService;
        this.jwtService = jwtService;
        this.jdbc = jdbc;
    }

    @RequestMapping(value = {"/proxy-external", "/proxy-request"}, method = {RequestMethod.GET, RequestMethod.POST, RequestMethod.PUT, RequestMethod.DELETE, RequestMethod.PATCH})
    public ResponseEntity<?> handleProxy(
            HttpServletRequest request,
            @RequestHeader HttpHeaders headers,
            @RequestBody(required = false) String body) {

        String authHeader = request.getHeader("Authorization");
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }

        try {

        String globalUseToggle = "off";
        String globalRecordingMode = "true";
        try {
            java.util.List<Map<String, Object>> settings = jdbc.queryForList("SELECT key, value FROM settings WHERE key IN ('useToggle', 'recordingMode')");
            for (Map<String, Object> row : settings) {
                String k = (String) row.get("key");
                String v = (String) row.get("value");
                if ("useToggle".equals(k)) globalUseToggle = v;
                if ("recordingMode".equals(k)) globalRecordingMode = v;
            }
        } catch (Exception e) {}

        String toggleHeader = SecurityHelper.firstNonBlank(request.getHeader("X-Use-Toggle"), request.getHeader("X-Mock-Toggle"), globalUseToggle);
        boolean useMock = "off".equalsIgnoreCase(SecurityHelper.resolveToggleValue(toggleHeader));
        boolean autoRecord = "on".equalsIgnoreCase(SecurityHelper.resolveToggleValue(globalRecordingMode));

        String targetHost = request.getHeader("X-Target-Host");
        String path = request.getParameter("endpoint");
        if (path == null) path = "/";

        String url = request.getRequestURI().endsWith("/proxy-external") ? request.getParameter("url") : (targetHost + path);
        
        if (url == null || url.isEmpty()) {
            return ResponseEntity.badRequest().body("{\"error\": \"URL or Target Host is required\"}");
        }

        try {
            URI uri = new URI(url);
            InetAddress address = InetAddress.getByName(uri.getHost());
            if (address.isAnyLocalAddress() || address.isLoopbackAddress() || address.isLinkLocalAddress() || address.isSiteLocalAddress()) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN).body("{\"error\": \"SSRF protection: Internal addresses are not allowed\"}");
            }
        } catch (Exception e) {
            return ResponseEntity.badRequest().body("{\"error\": \"Invalid URL\"}");
        }

        Map<String, String> requestHeaders = new HashMap<>();
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            requestHeaders.put(headerName, request.getHeader(headerName));
        }

        if (useMock) {
            Stub stub = stubMatchingService.findMatchingStub(request.getMethod(), url, body, targetHost);
            if (stub != null) {
                if (stub.getDelay() > 0) {
                    try {
                        Thread.sleep(stub.getDelay());
                    } catch (InterruptedException ignored) {}
                }
                String reqId = UUID.randomUUID().toString();
                jdbc.update("INSERT INTO requests (id, method, url, baseUrl, endpoint, headers, body, status, response, responseHeaders, isRecorded, category, ownerGroup, source) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                        reqId, request.getMethod(), url, targetHost, path, requestHeaders.toString(), body, stub.getResponseStatus(), stub.getResponseBody(), stub.getResponseHeaders() != null ? stub.getResponseHeaders() : "", 1, stub.getCategory() != null ? stub.getCategory() : "other", user.getAdGroup(), "stub");

                ResponseEntity.BodyBuilder builder = ResponseEntity.status(stub.getResponseStatus())
                        .header("X-Response-Source", "stub")
                        .header("X-Recorded-Id", reqId);
                
                if (stub.getResponseHeaders() != null && !stub.getResponseHeaders().isEmpty() && !stub.getResponseHeaders().equals("{}")) {
                    try {
                        com.fasterxml.jackson.databind.ObjectMapper mapper = new com.fasterxml.jackson.databind.ObjectMapper();
                        Map<String, String> hdrs = mapper.readValue(stub.getResponseHeaders(), new com.fasterxml.jackson.core.type.TypeReference<Map<String, String>>(){});
                        for (Map.Entry<String, String> entry : hdrs.entrySet()) {
                            builder.header(entry.getKey(), entry.getValue());
                        }
                    } catch (Exception e) {}
                }
                
                return builder.body(stub.getResponseBody());
            }
            // If no stub is found, fall through to live API to allow auto-recording
        }

        HttpForwardService.OutboundResponse res = httpForwardService.send(url, request.getMethod(), requestHeaders, body, null, 15);

        // Record request
        String reqId = UUID.randomUUID().toString();
        jdbc.update("INSERT INTO requests (id, method, url, baseUrl, endpoint, headers, body, status, response, responseHeaders, isRecorded, category, ownerGroup, source) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                reqId, request.getMethod(), url, targetHost, path, requestHeaders.toString(), body, res.statusCode, res.body, res.headers != null ? res.headers.toString() : "", 1, "other", user.getAdGroup(), "live-api");

        if (res.error != null) {
            if (useMock) {
                // Fallback to stub if upstream fails
                Stub stub = stubMatchingService.findMatchingStub(request.getMethod(), url, body, targetHost);
                if (stub != null) {
                    String fallbackReqId = UUID.randomUUID().toString();
                    jdbc.update("INSERT INTO requests (id, method, url, baseUrl, endpoint, headers, body, status, response, responseHeaders, isRecorded, category, ownerGroup, source) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                            fallbackReqId, request.getMethod(), url, targetHost, path, requestHeaders.toString(), body, stub.getResponseStatus(), stub.getResponseBody(), stub.getResponseHeaders() != null ? stub.getResponseHeaders() : "", 1, stub.getCategory() != null ? stub.getCategory() : "other", user.getAdGroup(), "stub (fallback)");

                    ResponseEntity.BodyBuilder builder = ResponseEntity.status(stub.getResponseStatus())
                            .header("X-Response-Source", "stub (fallback)")
                            .header("X-Recorded-Id", fallbackReqId);
                    if (stub.getResponseHeaders() != null && !stub.getResponseHeaders().isEmpty() && !stub.getResponseHeaders().equals("{}")) {
                        try {
                            com.fasterxml.jackson.databind.ObjectMapper mapper = new com.fasterxml.jackson.databind.ObjectMapper();
                            Map<String, String> hdrs = mapper.readValue(stub.getResponseHeaders(), new com.fasterxml.jackson.core.type.TypeReference<Map<String, String>>(){});
                            for (Map.Entry<String, String> entry : hdrs.entrySet()) {
                                String key = entry.getKey();
                                String val = entry.getValue();
                                if (key == null || val == null) continue;
                                if (!key.equalsIgnoreCase("content-length") && !key.equalsIgnoreCase("transfer-encoding") && !key.equalsIgnoreCase("connection") && !key.equalsIgnoreCase("content-encoding")) {
                                    val = val.replaceAll("[\\r\\n]+", " ");
                                    if (key.matches("^[a-zA-Z0-9!#$%&'*+.^_`|~-]+$")) {
                                        builder.header(key, val);
                                    }
                                }
                            }
                        } catch (Exception e) {}
                    }
                    return builder.body(stub.getResponseBody());
                }
            }
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE).body("{\"error\": \"" + res.error + "\"}");
        }

        try {
            // Auto-create stub or version for all status codes
            if (autoRecord) {
                Stub existingStub = stubMatchingService.findMatchingStub(request.getMethod(), url, body, targetHost, true);
                if (existingStub != null) {
                    java.util.List<com.example.virtualization.model.StubVersion> existingVersions = stubService.getVersions(existingStub.getId());
                    boolean statusExists = existingVersions.stream().anyMatch(v -> v.getResponseStatus() == res.statusCode);
                    
                    if (!statusExists) {
                        // Create a new version for the existing stub
                        com.example.virtualization.model.StubVersion newVersion = new com.example.virtualization.model.StubVersion();
                        newVersion.setVersion("auto-" + System.currentTimeMillis());
                        newVersion.setVersionTag("Auto-recorded live traffic");
                        newVersion.setResponseStatus(res.statusCode);
                        newVersion.setResponseBody(res.body);
                        newVersion.setResponseHeaders(res.headers != null ? res.headers.toString() : "");
                        newVersion.setActive(false); // Save silently as inactive
                        stubService.createVersion(existingStub.getId(), newVersion);
                    } else {
                        // Update the existing version with the latest live response
                        existingVersions.stream()
                                .filter(v -> v.getResponseStatus() == res.statusCode)
                                .findFirst()
                                .ifPresent(v -> {
                                    v.setResponseBody(res.body);
                                    v.setResponseHeaders(res.headers != null ? res.headers.toString() : "");
                                    stubService.updateVersion(v.getVersionId(), v);
                                    
                                    // If this version is the currently active version, we must ALSO update the main stub
                                    if (v.isActive()) {
                                        existingStub.setResponseBody(res.body);
                                        existingStub.setResponseHeaders(res.headers != null ? res.headers.toString() : "");
                                        stubService.createOrUpdateStub(existingStub);
                                    }
                                });
                    }
                } else {
                    Stub stub = new Stub();
                    stub.setName("Auto-recorded: " + path);
                    stub.setMethod(request.getMethod());
                    stub.setEndpoint(path);
                    stub.setEnvironment("Dev");
                    stub.setDescription("Auto-recorded from live traffic");
                    stub.setBaseUrl(targetHost);
                    stub.setResponseStatus(res.statusCode);
                    stub.setResponseBody(res.body);
                    stub.setResponseHeaders(res.headers != null ? res.headers.toString() : "");
                    stub.setEnabled(true);
                    stub.setCategory("other");
                    stub.setVersion("v1");
                    stub.setOwnerGroup(user.getAdGroup());
                    Stub createdStub = stubService.createOrUpdateStub(stub);

                    // Create initial v1 version
                    com.example.virtualization.model.StubVersion v1 = new com.example.virtualization.model.StubVersion();
                    v1.setVersion("v1");
                    v1.setVersionTag("Initial recording");
                    v1.setResponseStatus(res.statusCode);
                    v1.setResponseBody(res.body);
                    v1.setResponseHeaders(res.headers != null ? res.headers.toString() : "");
                    v1.setActive(true);
                    stubService.createVersion(createdStub.getId(), v1);
                }
            }
        } catch (Exception e) {
            System.err.println("Auto-record failed: " + e.getMessage());
        }

        ResponseEntity.BodyBuilder builder = ResponseEntity.status(res.statusCode)
                .header("X-Response-Source", "live-api")
                .header("X-Recorded-Id", reqId);

        if (res.headers != null) {
            for (Map.Entry<String, String> entry : res.headers.entrySet()) {
                String key = entry.getKey();
                String val = entry.getValue();
                if (key == null || val == null) continue;
                if (!key.equalsIgnoreCase("content-length") && !key.equalsIgnoreCase("transfer-encoding") && !key.equalsIgnoreCase("connection") && !key.equalsIgnoreCase("content-encoding")) {
                    try {
                        val = val.replaceAll("[\\r\\n]+", " ");
                        if (key.matches("^[a-zA-Z0-9!#$%&'*+.^_`|~-]+$")) {
                            builder.header(key, val);
                        }
                    } catch (Exception e) {}
                }
            }
        }
        
        return builder.body(res.body);
        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body("{\"error\": \"Backend Proxy Exception: " + e.getClass().getName() + " - " + e.getMessage() + "\"}");
        }
    }
}
