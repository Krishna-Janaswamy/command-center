package com.example.virtualization.model;

public class ApiRegistryEntry {
    private String id;
    private String functionName;
    private String method;
    private String endpoint;
    private String category;
    private String environment;
    private String description;
    private String healthCheckUrl;
    private String healthCheckHeaders;
    private String healthCheckBody;
    private String healthCheckParams;
    private int retryOn500;
    private boolean isCustom;
    private String ownerGroup;
    private String createdAt;

    public ApiRegistryEntry() {}

    // Getters and Setters
    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getFunctionName() { return functionName; }
    public void setFunctionName(String functionName) { this.functionName = functionName; }

    public String getMethod() { return method; }
    public void setMethod(String method) { this.method = method; }

    public String getEndpoint() { return endpoint; }
    public void setEndpoint(String endpoint) { this.endpoint = endpoint; }

    public String getCategory() { return category; }
    public void setCategory(String category) { this.category = category; }

    public String getEnvironment() { return environment; }
    public void setEnvironment(String environment) { this.environment = environment; }

    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }

    public String getHealthCheckUrl() { return healthCheckUrl; }
    public void setHealthCheckUrl(String healthCheckUrl) { this.healthCheckUrl = healthCheckUrl; }

    public String getHealthCheckHeaders() { return healthCheckHeaders; }
    public void setHealthCheckHeaders(String healthCheckHeaders) { this.healthCheckHeaders = healthCheckHeaders; }

    public String getHealthCheckBody() { return healthCheckBody; }
    public void setHealthCheckBody(String healthCheckBody) { this.healthCheckBody = healthCheckBody; }

    public String getHealthCheckParams() { return healthCheckParams; }
    public void setHealthCheckParams(String healthCheckParams) { this.healthCheckParams = healthCheckParams; }

    public int getRetryOn500() { return retryOn500; }
    public void setRetryOn500(int retryOn500) { this.retryOn500 = retryOn500; }

    public boolean isCustom() { return isCustom; }
    public void setCustom(boolean custom) { isCustom = custom; }

    public String getOwnerGroup() { return ownerGroup; }
    public void setOwnerGroup(String ownerGroup) { this.ownerGroup = ownerGroup; }

    public String getCreatedAt() { return createdAt; }
    public void setCreatedAt(String createdAt) { this.createdAt = createdAt; }
}
