package com.example.virtualization.model;

public class Stub {
    private String id;
    private String name;
    private String method;
    private String endpoint;
    private String baseUrl;
    private String environment;
    private String description;
    private String requestMatcher;
    private int responseStatus;
    private String responseBody;
    private String responseHeaders;
    private int delay;
    private boolean enabled;
    private String category;
    private String version;
    private String ownerGroup;
    private String createdAt;

    public Stub() {}

    // Getters and Setters
    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }

    public String getMethod() { return method; }
    public void setMethod(String method) { this.method = method; }

    public String getEndpoint() { return endpoint; }
    public void setEndpoint(String endpoint) { this.endpoint = endpoint; }

    public String getEnvironment() { return environment; }
    public void setEnvironment(String environment) { this.environment = environment; }

    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }

    public String getBaseUrl() { return baseUrl; }
    public void setBaseUrl(String baseUrl) { this.baseUrl = baseUrl; }

    public String getRequestMatcher() { return requestMatcher; }
    public void setRequestMatcher(String requestMatcher) { this.requestMatcher = requestMatcher; }

    public int getResponseStatus() { return responseStatus; }
    public void setResponseStatus(int responseStatus) { this.responseStatus = responseStatus; }

    public String getResponseBody() { return responseBody; }
    public void setResponseBody(String responseBody) { this.responseBody = responseBody; }

    public String getResponseHeaders() { return responseHeaders; }
    public void setResponseHeaders(String responseHeaders) { this.responseHeaders = responseHeaders; }

    public int getDelay() { return delay; }
    public void setDelay(int delay) { this.delay = delay; }

    public boolean isEnabled() { return enabled; }
    public void setEnabled(boolean enabled) { this.enabled = enabled; }

    public String getCategory() { return category; }
    public void setCategory(String category) { this.category = category; }

    public String getVersion() { return version; }
    public void setVersion(String version) { this.version = version; }

    public String getOwnerGroup() { return ownerGroup; }
    public void setOwnerGroup(String ownerGroup) { this.ownerGroup = ownerGroup; }

    public String getCreatedAt() { return createdAt; }
    public void setCreatedAt(String createdAt) { this.createdAt = createdAt; }
}
