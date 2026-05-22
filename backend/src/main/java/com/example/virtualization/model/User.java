package com.example.virtualization.model;

public class User {
    private String username;
    private String password;
    private String email;
    private String role;
    private String adGroup;
    private String createdAt;

    public User() {}

    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }

    public String getPassword() { return password; }
    public void setPassword(String password) { this.password = password; }

    public String getEmail() { return email; }
    public void setEmail(String email) { this.email = email; }

    public String getRole() { return role; }
    public void setRole(String role) { this.role = role; }

    public String getAdGroup() { return adGroup; }
    public void setAdGroup(String adGroup) { this.adGroup = adGroup; }

    public String getCreatedAt() { return createdAt; }
    public void setCreatedAt(String createdAt) { this.createdAt = createdAt; }
}
