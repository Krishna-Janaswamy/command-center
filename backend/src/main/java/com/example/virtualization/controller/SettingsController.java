package com.example.virtualization.controller;

import com.example.virtualization.model.User;
import com.example.virtualization.service.JwtService;
import com.example.virtualization.util.SecurityHelper;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/settings")
@CrossOrigin(origins = "*")
public class SettingsController {
    private final JdbcTemplate jdbc;
    private final JwtService jwtService;

    public SettingsController(JdbcTemplate jdbc, JwtService jwtService) {
        this.jdbc = jdbc;
        this.jwtService = jwtService;
    }

    @GetMapping
    public ResponseEntity<Map<String, String>> getAllSettings(
            @RequestHeader(value = "Authorization", required = false) String authHeader) {
        
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();

        List<Map<String, Object>> rows = jdbc.queryForList("SELECT * FROM settings");
        Map<String, String> settings = new HashMap<>();
        // Default settings
        settings.put("recordingMode", "true");
        settings.put("playbackMode", "true");
        settings.put("targetUrl", "");
        settings.put("useToggle", "true");

        for (Map<String, Object> row : rows) {
            settings.put((String) row.get("key"), (String) row.get("value"));
        }

        return ResponseEntity.ok(settings);
    }

    @PutMapping("/{key}")
    public ResponseEntity<Void> updateSetting(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("key") String key,
            @RequestBody Map<String, String> body) {
        
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        String value = body.get("value");
        jdbc.update("INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value);
        
        return ResponseEntity.ok().build();
    }
}
