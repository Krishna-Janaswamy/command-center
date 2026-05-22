package com.example.virtualization.service;

import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import jakarta.annotation.PostConstruct;

@Service
public class DbService {
    private final JdbcTemplate jdbc;

    public DbService(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    @PostConstruct
    public void initializeDatabase() {
        // api_registry
        jdbc.execute("CREATE TABLE IF NOT EXISTS api_registry (id TEXT PRIMARY KEY, functionName TEXT, method TEXT, endpoint TEXT, category TEXT, environment TEXT, description TEXT, healthCheckUrl TEXT, healthCheckHeaders TEXT, healthCheckBody TEXT, healthCheckParams TEXT, retryOn500 INTEGER DEFAULT 0, isCustom INTEGER DEFAULT 0, ownerGroup TEXT DEFAULT 'admin', createdAt DATETIME DEFAULT CURRENT_TIMESTAMP)");

        // stubs
        jdbc.execute("CREATE TABLE IF NOT EXISTS stubs (id TEXT PRIMARY KEY, name TEXT NOT NULL, method TEXT NOT NULL, urlPattern TEXT NOT NULL, targetHost TEXT DEFAULT '', requestMatcher TEXT, responseStatus INTEGER DEFAULT 200, responseBody TEXT, responseHeaders TEXT, delay INTEGER DEFAULT 0, enabled INTEGER DEFAULT 1, category TEXT DEFAULT 'other', version TEXT DEFAULT 'v1', ownerGroup TEXT DEFAULT 'admin', createdAt DATETIME DEFAULT CURRENT_TIMESTAMP)");

        // stub_versions
        jdbc.execute("CREATE TABLE IF NOT EXISTS stub_versions (versionId TEXT PRIMARY KEY, stubId TEXT NOT NULL, version TEXT NOT NULL, versionTag TEXT, responseStatus INTEGER DEFAULT 200, responseBody TEXT, responseHeaders TEXT, isActive INTEGER DEFAULT 0, createdAt DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (stubId) REFERENCES stubs(id) ON DELETE CASCADE)");

        // requests
        jdbc.execute("CREATE TABLE IF NOT EXISTS requests (id TEXT PRIMARY KEY, method TEXT NOT NULL, url TEXT NOT NULL, baseUrl TEXT, endpoint TEXT, headers TEXT NOT NULL, body TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP, status INTEGER, response TEXT, responseHeaders TEXT, isRecorded INTEGER DEFAULT 0, category TEXT DEFAULT 'other', ownerGroup TEXT DEFAULT 'admin')");

        // settings
        jdbc.execute("CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT)");

        // users
        jdbc.execute("CREATE TABLE IF NOT EXISTS users (username TEXT PRIMARY KEY, password TEXT NOT NULL, email TEXT, role TEXT, adGroup TEXT, createdAt DATETIME DEFAULT CURRENT_TIMESTAMP)");
    }
}
