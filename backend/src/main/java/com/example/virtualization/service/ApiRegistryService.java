package com.example.virtualization.service;

import com.example.virtualization.model.ApiRegistryEntry;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;

@Service
public class ApiRegistryService {
    private final JdbcTemplate jdbc;

    public ApiRegistryService(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    public List<ApiRegistryEntry> listApis(String env, String category, String ownerGroup, boolean isAdmin) {
        StringBuilder sql = new StringBuilder("SELECT * FROM api_registry WHERE 1=1");
        if (env != null && !env.isEmpty()) sql.append(" AND environment = '").append(env.replace("'", "")).append("'");
        if (category != null && !category.isEmpty()) sql.append(" AND category = '").append(category.replace("'", "")).append("'");
        if (!isAdmin && ownerGroup != null && !ownerGroup.isEmpty()) {
            sql.append(" AND (ownerGroup = '").append(ownerGroup.replace("'", "")).append("' OR ownerGroup = 'admin')");
        }

        return jdbc.query(sql.toString(), (rs, rowNum) -> {
            ApiRegistryEntry api = new ApiRegistryEntry();
            api.setId(rs.getString("id"));
            api.setFunctionName(rs.getString("functionName"));
            api.setMethod(rs.getString("method"));
            api.setEndpoint(rs.getString("endpoint"));
            api.setCategory(rs.getString("category"));
            api.setEnvironment(rs.getString("environment"));
            api.setDescription(rs.getString("description"));
            api.setHealthCheckUrl(rs.getString("healthCheckUrl"));
            api.setHealthCheckHeaders(rs.getString("healthCheckHeaders"));
            api.setHealthCheckBody(rs.getString("healthCheckBody"));
            api.setHealthCheckParams(rs.getString("healthCheckParams"));
            api.setRetryOn500(rs.getInt("retryOn500"));
            api.setCustom(rs.getInt("isCustom") == 1);
            api.setOwnerGroup(rs.getString("ownerGroup"));
            api.setCreatedAt(rs.getString("createdAt"));
            return api;
        });
    }

    public ApiRegistryEntry getApi(String id) {
        List<ApiRegistryEntry> list = jdbc.query("SELECT * FROM api_registry WHERE id = ?", (rs, rowNum) -> {
            ApiRegistryEntry api = new ApiRegistryEntry();
            api.setId(rs.getString("id"));
            // (populate fields)
            api.setFunctionName(rs.getString("functionName"));
            api.setMethod(rs.getString("method"));
            api.setEndpoint(rs.getString("endpoint"));
            api.setCategory(rs.getString("category"));
            api.setEnvironment(rs.getString("environment"));
            api.setDescription(rs.getString("description"));
            api.setHealthCheckUrl(rs.getString("healthCheckUrl"));
            api.setHealthCheckHeaders(rs.getString("healthCheckHeaders"));
            api.setHealthCheckBody(rs.getString("healthCheckBody"));
            api.setHealthCheckParams(rs.getString("healthCheckParams"));
            api.setRetryOn500(rs.getInt("retryOn500"));
            api.setCustom(rs.getInt("isCustom") == 1);
            api.setOwnerGroup(rs.getString("ownerGroup"));
            api.setCreatedAt(rs.getString("createdAt"));
            return api;
        }, id);
        return list.isEmpty() ? null : list.get(0);
    }

    public ApiRegistryEntry createApi(ApiRegistryEntry api) {
        if (api.getId() == null || api.getId().isEmpty()) {
            api.setId(UUID.randomUUID().toString());
        }
        jdbc.update("INSERT INTO api_registry (id, functionName, method, endpoint, category, environment, description, healthCheckUrl, healthCheckHeaders, healthCheckBody, healthCheckParams, retryOn500, isCustom, ownerGroup) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                api.getId(), api.getFunctionName(), api.getMethod(), api.getEndpoint(), api.getCategory(), api.getEnvironment(), api.getDescription(), api.getHealthCheckUrl(), api.getHealthCheckHeaders(), api.getHealthCheckBody(), api.getHealthCheckParams(), api.getRetryOn500(), api.isCustom() ? 1 : 0, api.getOwnerGroup());
        return getApi(api.getId());
    }

    public void updateApi(String id, ApiRegistryEntry api) {
        jdbc.update("UPDATE api_registry SET functionName=?, method=?, endpoint=?, category=?, environment=?, description=?, healthCheckUrl=?, healthCheckHeaders=?, healthCheckBody=?, healthCheckParams=?, retryOn500=?, isCustom=? WHERE id=?",
                api.getFunctionName(), api.getMethod(), api.getEndpoint(), api.getCategory(), api.getEnvironment(), api.getDescription(), api.getHealthCheckUrl(), api.getHealthCheckHeaders(), api.getHealthCheckBody(), api.getHealthCheckParams(), api.getRetryOn500(), api.isCustom() ? 1 : 0, id);
    }

    public void deleteApi(String id) {
        jdbc.update("DELETE FROM api_registry WHERE id = ?", id);
    }

    public void deleteAllApis() {
        jdbc.update("DELETE FROM api_registry");
    }
}
