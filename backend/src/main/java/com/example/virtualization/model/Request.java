package com.example.virtualization.model;

public class Request {
    private String id;
    private String method;
    private String url;
    private String baseUrl;
    private String endpoint;
    private String headers;
    private String body;
    private String timestamp;
    private Integer status;
    private String response;
    private String responseHeaders;
    private boolean isRecorded;
    private String category;
    private String ownerGroup;

    public Request() {}

    // Getters and Setters
    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getMethod() { return method; }
    public void setMethod(String method) { this.method = method; }

    public String getUrl() { return url; }
    public void setUrl(String url) { this.url = url; }

    public String getBaseUrl() { return baseUrl; }
    public void setBaseUrl(String baseUrl) { this.baseUrl = baseUrl; }

    public String getEndpoint() { return endpoint; }
    public void setEndpoint(String endpoint) { this.endpoint = endpoint; }

    public String getHeaders() { return headers; }
    public void setHeaders(String headers) { this.headers = headers; }

    public String getBody() { return body; }
    public void setBody(String body) { this.body = body; }

    public String getTimestamp() { return timestamp; }
    public void setTimestamp(String timestamp) { this.timestamp = timestamp; }

    public Integer getStatus() { return status; }
    public void setStatus(Integer status) { this.status = status; }

    public String getResponse() { return response; }
    public void setResponse(String response) { this.response = response; }

    public String getResponseHeaders() { return responseHeaders; }
    public void setResponseHeaders(String responseHeaders) { this.responseHeaders = responseHeaders; }

    public boolean isRecorded() { return isRecorded; }
    public void setRecorded(boolean recorded) { isRecorded = recorded; }

    public String getCategory() { return category; }
    public void setCategory(String category) { this.category = category; }

    public String getOwnerGroup() { return ownerGroup; }
    public void setOwnerGroup(String ownerGroup) { this.ownerGroup = ownerGroup; }
}
