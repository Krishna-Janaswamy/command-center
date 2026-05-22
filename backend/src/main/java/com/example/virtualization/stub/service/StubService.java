package com.example.virtualization.stub.service;

import com.example.virtualization.model.Stub;
import com.example.virtualization.model.StubVersion;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;

@Service
public class StubService {
    private final JdbcTemplate jdbc;

    public StubService(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    public List<Stub> getAllStubs() {
        return jdbc.query("SELECT * FROM stubs", (rs, rowNum) -> {
            Stub stub = new Stub();
            stub.setId(rs.getString("id"));
            stub.setName(rs.getString("name"));
            stub.setMethod(rs.getString("method"));
            stub.setUrlPattern(rs.getString("urlPattern"));
            stub.setTargetHost(rs.getString("targetHost"));
            stub.setRequestMatcher(rs.getString("requestMatcher"));
            stub.setResponseStatus(rs.getInt("responseStatus"));
            stub.setResponseBody(rs.getString("responseBody"));
            stub.setResponseHeaders(rs.getString("responseHeaders"));
            stub.setDelay(rs.getInt("delay"));
            stub.setEnabled(rs.getInt("enabled") == 1);
            stub.setCategory(rs.getString("category"));
            stub.setVersion(rs.getString("version"));
            stub.setOwnerGroup(rs.getString("ownerGroup"));
            stub.setCreatedAt(rs.getString("createdAt"));
            return stub;
        });
    }

    public Stub getStub(String id) {
        List<Stub> stubs = jdbc.query("SELECT * FROM stubs WHERE id = ?", (rs, rowNum) -> {
            Stub stub = new Stub();
            stub.setId(rs.getString("id"));
            stub.setName(rs.getString("name"));
            stub.setMethod(rs.getString("method"));
            stub.setUrlPattern(rs.getString("urlPattern"));
            stub.setTargetHost(rs.getString("targetHost"));
            stub.setRequestMatcher(rs.getString("requestMatcher"));
            stub.setResponseStatus(rs.getInt("responseStatus"));
            stub.setResponseBody(rs.getString("responseBody"));
            stub.setResponseHeaders(rs.getString("responseHeaders"));
            stub.setDelay(rs.getInt("delay"));
            stub.setEnabled(rs.getInt("enabled") == 1);
            stub.setCategory(rs.getString("category"));
            stub.setVersion(rs.getString("version"));
            stub.setOwnerGroup(rs.getString("ownerGroup"));
            stub.setCreatedAt(rs.getString("createdAt"));
            return stub;
        }, id);
        return stubs.isEmpty() ? null : stubs.get(0);
    }

    public Stub createOrUpdateStub(Stub stub) {
        if (stub.getId() == null || stub.getId().isEmpty()) {
            stub.setId(UUID.randomUUID().toString());
            jdbc.update("INSERT INTO stubs (id, name, method, urlPattern, targetHost, requestMatcher, responseStatus, responseBody, responseHeaders, delay, enabled, category, version, ownerGroup) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                    stub.getId(), stub.getName(), stub.getMethod(), stub.getUrlPattern(), stub.getTargetHost(), stub.getRequestMatcher(), stub.getResponseStatus(), stub.getResponseBody(), stub.getResponseHeaders(), stub.getDelay(), stub.isEnabled() ? 1 : 0, stub.getCategory(), stub.getVersion(), stub.getOwnerGroup());
        } else {
            jdbc.update("UPDATE stubs SET name=?, method=?, urlPattern=?, targetHost=?, requestMatcher=?, responseStatus=?, responseBody=?, responseHeaders=?, delay=?, enabled=?, category=?, version=? WHERE id=?",
                    stub.getName(), stub.getMethod(), stub.getUrlPattern(), stub.getTargetHost(), stub.getRequestMatcher(), stub.getResponseStatus(), stub.getResponseBody(), stub.getResponseHeaders(), stub.getDelay(), stub.isEnabled() ? 1 : 0, stub.getCategory(), stub.getVersion(), stub.getId());
        }
        return getStub(stub.getId());
    }

    public void deleteStub(String id) {
        jdbc.update("DELETE FROM stubs WHERE id = ?", id);
    }

    public void clearAllStubs() {
        jdbc.update("DELETE FROM stubs");
    }

    public Stub toggleStub(String id) {
        jdbc.update("UPDATE stubs SET enabled = CASE WHEN enabled = 1 THEN 0 ELSE 1 END WHERE id = ?", id);
        return getStub(id);
    }

    // Version management
    public List<StubVersion> getVersions(String stubId) {
        return jdbc.query("SELECT * FROM stub_versions WHERE stubId = ? ORDER BY version ASC", (rs, rowNum) -> {
            StubVersion v = new StubVersion();
            v.setVersionId(rs.getString("versionId"));
            v.setStubId(rs.getString("stubId"));
            v.setVersion(rs.getString("version"));
            v.setVersionTag(rs.getString("versionTag"));
            v.setResponseStatus(rs.getInt("responseStatus"));
            v.setResponseBody(rs.getString("responseBody"));
            v.setResponseHeaders(rs.getString("responseHeaders"));
            v.setActive(rs.getInt("isActive") == 1);
            v.setCreatedAt(rs.getString("createdAt"));
            return v;
        }, stubId);
    }

    public StubVersion getVersion(String versionId) {
        List<StubVersion> versions = jdbc.query("SELECT * FROM stub_versions WHERE versionId = ?", (rs, rowNum) -> {
            StubVersion v = new StubVersion();
            v.setVersionId(rs.getString("versionId"));
            v.setStubId(rs.getString("stubId"));
            v.setVersion(rs.getString("version"));
            v.setVersionTag(rs.getString("versionTag"));
            v.setResponseStatus(rs.getInt("responseStatus"));
            v.setResponseBody(rs.getString("responseBody"));
            v.setResponseHeaders(rs.getString("responseHeaders"));
            v.setActive(rs.getInt("isActive") == 1);
            v.setCreatedAt(rs.getString("createdAt"));
            return v;
        }, versionId);
        return versions.isEmpty() ? null : versions.get(0);
    }

    public StubVersion createVersion(String stubId, StubVersion version) {
        version.setVersionId(UUID.randomUUID().toString());
        version.setStubId(stubId);
        if (version.getVersion() == null || version.getVersion().isEmpty()) {
            version.setVersion("v1");
        }
        jdbc.update("INSERT INTO stub_versions (versionId, stubId, version, versionTag, responseStatus, responseBody, responseHeaders, isActive) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
                version.getVersionId(), version.getStubId(), version.getVersion(), version.getVersionTag(), version.getResponseStatus(), version.getResponseBody(), version.getResponseHeaders(), version.isActive() ? 1 : 0);
        return getVersion(version.getVersionId());
    }

    public void updateVersion(String versionId, StubVersion version) {
        jdbc.update("UPDATE stub_versions SET versionTag=?, responseStatus=?, responseBody=?, responseHeaders=? WHERE versionId=?",
                version.getVersionTag(), version.getResponseStatus(), version.getResponseBody(), version.getResponseHeaders(), versionId);
    }

    public void deleteVersion(String versionId) {
        jdbc.update("DELETE FROM stub_versions WHERE versionId = ?", versionId);
    }

    public void activateVersion(String stubId, String versionId) {
        jdbc.update("UPDATE stub_versions SET isActive = 0 WHERE stubId = ?", stubId);
        jdbc.update("UPDATE stub_versions SET isActive = 1 WHERE versionId = ?", versionId);

        StubVersion active = getVersion(versionId);
        if (active != null) {
            jdbc.update("UPDATE stubs SET responseStatus = ?, responseBody = ?, responseHeaders = ? WHERE id = ?",
                    active.getResponseStatus(), active.getResponseBody(), active.getResponseHeaders(), stubId);
        }
    }
}
