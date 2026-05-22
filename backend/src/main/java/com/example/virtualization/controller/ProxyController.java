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
    public ResponseEntity<Map<String, Object>> handleProxy(
            HttpServletRequest request,
            @RequestHeader HttpHeaders headers,
            @RequestBody(required = false) String body) {

        String authHeader = request.getHeader("Authorization");
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }

        String toggleHeader = SecurityHelper.firstNonBlank(request.getHeader("X-Use-Toggle"), request.getHeader("X-Mock-Toggle"));
        boolean useMock = "off".equalsIgnoreCase(SecurityHelper.resolveToggleValue(toggleHeader));

        String targetHost = request.getHeader("X-Target-Host");
        String path = request.getParameter("endpoint");
        if (path == null) path = "/";

        String url = request.getRequestURI().endsWith("/proxy-external") ? request.getParameter("url") : (targetHost + path);
        
        if (url == null || url.isEmpty()) {
            return ResponseEntity.badRequest().body(Map.of("error", "URL or Target Host is required"));
        }

        try {
            URI uri = new URI(url);
            InetAddress address = InetAddress.getByName(uri.getHost());
            if (address.isAnyLocalAddress() || address.isLoopbackAddress() || address.isLinkLocalAddress() || address.isSiteLocalAddress()) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN).body(Map.of("error", "SSRF protection: Internal addresses are not allowed"));
            }
        } catch (Exception e) {
            return ResponseEntity.badRequest().body(Map.of("error", "Invalid URL"));
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
                return ResponseEntity.status(stub.getResponseStatus())
                        .header("X-Response-Source", "stub")
                        .body(Map.of("status", stub.getResponseStatus(), "data", stub.getResponseBody(), "source", "stub"));
            }
            // If no stub is found, fall through to live API to allow auto-recording
        }

        HttpForwardService.OutboundResponse res = httpForwardService.send(url, request.getMethod(), requestHeaders, body, null, 15);

        // Record request
        String reqId = UUID.randomUUID().toString();
        jdbc.update("INSERT INTO requests (id, method, url, baseUrl, endpoint, headers, body, status, response, responseHeaders, isRecorded, category, ownerGroup) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                reqId, request.getMethod(), url, targetHost, path, requestHeaders.toString(), body, res.statusCode, res.body, res.headers != null ? res.headers.toString() : "", 1, "other", user.getAdGroup());

        if (res.error != null) {
            // Fallback to stub if upstream fails
            Stub stub = stubMatchingService.findMatchingStub(request.getMethod(), url, body, targetHost);
            if (stub != null) {
                return ResponseEntity.status(stub.getResponseStatus())
                        .header("X-Response-Source", "stub (fallback)")
                        .body(Map.of("status", stub.getResponseStatus(), "data", stub.getResponseBody(), "source", "stub (fallback)"));
            }
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE).body(Map.of("error", res.error));
        }

        // Auto-create stub or version on 2xx
        if (res.statusCode >= 200 && res.statusCode < 300) {
            Stub existingStub = stubMatchingService.findMatchingStub(request.getMethod(), url, body, targetHost, true);
            if (existingStub != null) {
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
                // Create new stub
                Stub stub = new Stub();
                stub.setName("Auto-recorded: " + path);
                stub.setMethod(request.getMethod());
                stub.setUrlPattern(path);
                stub.setTargetHost(targetHost);
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

        return ResponseEntity.status(res.statusCode)
                .header("X-Response-Source", "live-api")
                .body(Map.of("status", res.statusCode, "data", res.body, "source", "live-api", "recordedId", reqId));
    }
}
