package com.example.virtualization.controller;

import com.example.virtualization.model.User;
import com.example.virtualization.service.HttpForwardService;
import com.example.virtualization.service.JwtService;
import com.example.virtualization.util.SecurityHelper;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/health-check")
@CrossOrigin(origins = "*")
public class HealthCheckController {

    private final HttpForwardService httpForwardService;
    private final JwtService jwtService;

    public HealthCheckController(HttpForwardService httpForwardService, JwtService jwtService) {
        this.httpForwardService = httpForwardService;
        this.jwtService = jwtService;
    }

    @PostMapping
    public ResponseEntity<Map<String, Object>> performCheck(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @RequestBody Map<String, Object> requestBody) {
        
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }

        String url = (String) requestBody.get("url");
        String method = (String) requestBody.get("method");
        if (method == null) method = "GET";

        Map<String, String> headers = (Map<String, String>) requestBody.get("headers");
        Map<String, String> params = (Map<String, String>) requestBody.get("params");
        String body = (String) requestBody.get("body");

        HttpForwardService.OutboundResponse res = httpForwardService.send(url, method, headers, body, params, 15);

        Map<String, Object> result = new HashMap<>();
        if (res.error != null) {
            result.put("error", res.error);
        } else {
            result.put("statusCode", res.statusCode);
            result.put("responseTime", res.responseTime);
            result.put("body", res.body);
            result.put("headers", res.headers);
        }

        return ResponseEntity.ok(result);
    }
}
