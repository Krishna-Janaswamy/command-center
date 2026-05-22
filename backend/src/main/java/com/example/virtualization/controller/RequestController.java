package com.example.virtualization.controller;

import com.example.virtualization.model.Request;
import com.example.virtualization.model.User;
import com.example.virtualization.service.JwtService;
import com.example.virtualization.util.SecurityHelper;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api")
@CrossOrigin(origins = "*")
public class RequestController {
    private final JdbcTemplate jdbc;
    private final JwtService jwtService;

    public RequestController(JdbcTemplate jdbc, JwtService jwtService) {
        this.jdbc = jdbc;
        this.jwtService = jwtService;
    }

    @GetMapping("/recorded-requests")
    public ResponseEntity<List<Request>> getRecordedRequests(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @RequestParam(name = "limit", defaultValue = "100") int limit) {
        
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();

        String sql = "SELECT * FROM requests";
        if (!SecurityHelper.isAdmin(user)) {
            sql += " WHERE ownerGroup = '" + user.getAdGroup().replace("'", "") + "' OR ownerGroup = 'admin'";
        }
        sql += " ORDER BY timestamp DESC LIMIT " + limit;

        List<Request> requests = jdbc.query(sql, (rs, rowNum) -> {
            Request req = new Request();
            req.setId(rs.getString("id"));
            req.setMethod(rs.getString("method"));
            req.setUrl(rs.getString("url"));
            req.setBaseUrl(rs.getString("baseUrl"));
            req.setEndpoint(rs.getString("endpoint"));
            req.setHeaders(rs.getString("headers"));
            req.setBody(rs.getString("body"));
            req.setTimestamp(rs.getString("timestamp"));
            req.setStatus(rs.getInt("status"));
            req.setResponse(rs.getString("response"));
            req.setResponseHeaders(rs.getString("responseHeaders"));
            req.setRecorded(rs.getInt("isRecorded") == 1);
            req.setCategory(rs.getString("category"));
            req.setOwnerGroup(rs.getString("ownerGroup"));
            return req;
        });

        return ResponseEntity.ok(requests);
    }

    @GetMapping("/requests")
    public ResponseEntity<List<Request>> getRequests(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @RequestParam(name = "useToggle", required = false) String useToggle,
            @RequestParam(name = "limit", defaultValue = "100") int limit) {
        // useToggle logic isn't strictly mapping to DB filters here unless requested by client,
        // often just returns recorded-requests or filters by isRecorded flag.
        return getRecordedRequests(authHeader, limit);
    }

    @DeleteMapping("/requests")
    public ResponseEntity<Void> deleteAllRequests(
            @RequestHeader(value = "Authorization", required = false) String authHeader) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.isAdmin(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        jdbc.update("DELETE FROM requests");
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/requests/{id}")
    public ResponseEntity<Void> deleteRequest(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("id") String id) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.isAdmin(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        jdbc.update("DELETE FROM requests WHERE id = ?", id);
        return ResponseEntity.ok().build();
    }
}
