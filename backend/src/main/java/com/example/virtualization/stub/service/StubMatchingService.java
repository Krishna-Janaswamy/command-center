package com.example.virtualization.stub.service;

import com.example.virtualization.model.Stub;
import org.springframework.stereotype.Service;
import org.springframework.util.AntPathMatcher;

import java.util.List;
import java.util.regex.Pattern;

@Service
public class StubMatchingService {
    private final StubService stubService;
    private final AntPathMatcher matcher = new AntPathMatcher();

    public StubMatchingService(StubService stubService) {
        this.stubService = stubService;
    }

    public Stub findMatchingStub(String method, String url, String body, String targetHost) {
        return findMatchingStub(method, url, body, targetHost, false);
    }

    public Stub findMatchingStub(String method, String url, String body, String targetHost, boolean includeDisabled) {
        List<Stub> stubs = stubService.getAllStubs();
        for (Stub stub : stubs) {
            if (!includeDisabled && !stub.isEnabled()) continue;
            
            boolean methodMatches = "ANY".equalsIgnoreCase(stub.getMethod()) || stub.getMethod().equalsIgnoreCase(method);
            if (!methodMatches) continue;

            boolean pathMatches = false;
            String pattern = stub.getEndpoint();
            if (pattern != null) {
                if (pattern.startsWith("^") || pattern.contains(".*")) {
                    pathMatches = Pattern.compile(pattern).matcher(url).find();
                } else {
                    pathMatches = matcher.match(pattern, url) || url.contains(pattern);
                }
            } else {
                pathMatches = true;
            }

            if (!pathMatches) continue;

            if (targetHost != null && !targetHost.isEmpty() && stub.getBaseUrl() != null && !stub.getBaseUrl().isEmpty()) {
                if (!stub.getBaseUrl().equalsIgnoreCase(targetHost)) continue;
            }

            return stub;
        }
        return null;
    }
}
