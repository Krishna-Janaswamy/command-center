package com.example.virtualization.controller;

import com.example.virtualization.model.Stub;
import com.example.virtualization.stub.service.StubService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/mock")
@CrossOrigin(origins = "*")
public class MockEndpointController {

    private final StubService stubService;

    public MockEndpointController(StubService stubService) {
        this.stubService = stubService;
    }

    @RequestMapping(value = {"/{id}", "/{id}/**"}, method = {RequestMethod.GET, RequestMethod.POST, RequestMethod.PUT, RequestMethod.DELETE, RequestMethod.PATCH})
    public ResponseEntity<String> serveMock(@PathVariable("id") String id) {
        Stub stub = stubService.getStub(id);

        if (stub == null || !stub.isEnabled()) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body("{\"error\": \"Mock endpoint not found or disabled.\"}");
        }

        // Apply delay if configured
        if (stub.getDelay() > 0) {
            try {
                Thread.sleep(stub.getDelay());
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }

        ResponseEntity.BodyBuilder builder = ResponseEntity.status(stub.getResponseStatus());

        // Apply headers if present
        if (stub.getResponseHeaders() != null && !stub.getResponseHeaders().isEmpty() && !stub.getResponseHeaders().equals("{}")) {
            try {
                com.fasterxml.jackson.databind.ObjectMapper mapper = new com.fasterxml.jackson.databind.ObjectMapper();
                Map<String, String> hdrs = mapper.readValue(stub.getResponseHeaders(), new com.fasterxml.jackson.core.type.TypeReference<Map<String, String>>(){});
                for (Map.Entry<String, String> entry : hdrs.entrySet()) {
                    String key = entry.getKey();
                    String val = entry.getValue();
                    if (key == null || val == null) continue;
                    
                    // Exclude headers that can mess with the Spring response directly
                    if (!key.equalsIgnoreCase("content-length") && !key.equalsIgnoreCase("transfer-encoding") && !key.equalsIgnoreCase("connection") && !key.equalsIgnoreCase("content-encoding")) {
                        val = val.replaceAll("[\\r\\n]+", " ");
                        if (key.matches("^[a-zA-Z0-9!#$%&'*+.^_`|~-]+$")) {
                            builder.header(key, val);
                        }
                    }
                }
            } catch (Exception e) {
                // Ignore parsing errors for headers
            }
        }

        return builder.body(stub.getResponseBody());
    }
}
