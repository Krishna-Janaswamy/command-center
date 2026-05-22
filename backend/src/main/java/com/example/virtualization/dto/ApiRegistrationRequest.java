package com.example.virtualization.dto;

import jakarta.validation.constraints.NotBlank;

public class ApiRegistrationRequest {
    @NotBlank
    private String name;
    @NotBlank
    private String matchPath;
    @NotBlank
    private String targetUrl;
    private String httpMethod = "ANY";

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getMatchPath() {
        return matchPath;
    }

    public void setMatchPath(String matchPath) {
        this.matchPath = matchPath;
    }

    public String getTargetUrl() {
        return targetUrl;
    }

    public void setTargetUrl(String targetUrl) {
        this.targetUrl = targetUrl;
    }

    public String getHttpMethod() {
        return httpMethod;
    }

    public void setHttpMethod(String httpMethod) {
        this.httpMethod = httpMethod;
    }
}
