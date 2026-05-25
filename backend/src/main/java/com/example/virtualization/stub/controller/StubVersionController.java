package com.example.virtualization.stub.controller;

import com.example.virtualization.model.StubVersion;
import com.example.virtualization.model.User;
import com.example.virtualization.service.JwtService;
import com.example.virtualization.stub.service.StubService;
import com.example.virtualization.util.SecurityHelper;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/stubs/{stubId}/versions")
@CrossOrigin(origins = "*")
public class StubVersionController {

    private final StubService stubService;
    private final JwtService jwtService;

    public StubVersionController(StubService stubService, JwtService jwtService) {
        this.stubService = stubService;
        this.jwtService = jwtService;
    }

    @GetMapping
    public ResponseEntity<List<StubVersion>> getVersions(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("stubId") String stubId) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();

        return ResponseEntity.ok(stubService.getVersions(stubId));
    }

    @GetMapping("/{versionId}")
    public ResponseEntity<StubVersion> getVersion(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("stubId") String stubId,
            @PathVariable("versionId") String versionId) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();

        StubVersion version = stubService.getVersion(versionId);
        return version != null ? ResponseEntity.ok(version) : ResponseEntity.notFound().build();
    }

    @PostMapping
    public ResponseEntity<?> createVersion(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("stubId") String stubId,
            @RequestBody StubVersion version) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        try {
            return ResponseEntity.ok(stubService.createVersion(stubId, version));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.status(HttpStatus.CONFLICT).body(e.getMessage());
        }
    }

    @PutMapping("/{versionId}")
    public ResponseEntity<?> updateVersion(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("stubId") String stubId,
            @PathVariable("versionId") String versionId,
            @RequestBody StubVersion version) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        try {
            stubService.updateVersion(versionId, version);
            return ResponseEntity.ok().build();
        } catch (IllegalArgumentException e) {
            return ResponseEntity.status(HttpStatus.CONFLICT).body(e.getMessage());
        }
    }

    @DeleteMapping("/{versionId}")
    public ResponseEntity<Void> deleteVersion(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("stubId") String stubId,
            @PathVariable("versionId") String versionId) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canDelete(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        StubVersion version = stubService.getVersion(versionId);
        if (version != null && version.isActive()) {
            return ResponseEntity.status(HttpStatus.CONFLICT).build(); // Cannot delete active version
        }

        stubService.deleteVersion(versionId);
        return ResponseEntity.ok().build();
    }

    @PostMapping("/{versionId}/activate")
    public ResponseEntity<Void> activateVersion(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("stubId") String stubId,
            @PathVariable("versionId") String versionId) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        stubService.activateVersion(stubId, versionId);
        return ResponseEntity.ok().build();
    }
}
