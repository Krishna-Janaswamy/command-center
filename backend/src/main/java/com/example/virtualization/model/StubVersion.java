package com.example.virtualization.model;

public class StubVersion {
    private String versionId;
    private String stubId;
    private String version;
    private String versionTag;
    private int responseStatus;
    private String responseBody;
    private String responseHeaders;
    private boolean isActive;
    private String createdAt;

    public StubVersion() {}

    // Getters and Setters
    public String getVersionId() { return versionId; }
    public void setVersionId(String versionId) { this.versionId = versionId; }

    public String getStubId() { return stubId; }
    public void setStubId(String stubId) { this.stubId = stubId; }

    public String getVersion() { return version; }
    public void setVersion(String version) { this.version = version; }

    public String getVersionTag() { return versionTag; }
    public void setVersionTag(String versionTag) { this.versionTag = versionTag; }

    public int getResponseStatus() { return responseStatus; }
    public void setResponseStatus(int responseStatus) { this.responseStatus = responseStatus; }

    public String getResponseBody() { return responseBody; }
    public void setResponseBody(String responseBody) { this.responseBody = responseBody; }

    public String getResponseHeaders() { return responseHeaders; }
    public void setResponseHeaders(String responseHeaders) { this.responseHeaders = responseHeaders; }

    public boolean isActive() { return isActive; }
    public void setActive(boolean active) { isActive = active; }

    public String getCreatedAt() { return createdAt; }
    public void setCreatedAt(String createdAt) { this.createdAt = createdAt; }
}
