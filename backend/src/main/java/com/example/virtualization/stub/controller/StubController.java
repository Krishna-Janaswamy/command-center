package com.example.virtualization.stub.controller;

import com.example.virtualization.model.Stub;
import com.example.virtualization.model.User;
import com.example.virtualization.service.JwtService;
import com.example.virtualization.stub.service.StubService;
import com.example.virtualization.util.SecurityHelper;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/stubs")
@CrossOrigin(origins = "*")
public class StubController {

    private final StubService stubService;
    private final JwtService jwtService;

    public StubController(StubService stubService, JwtService jwtService) {
        this.stubService = stubService;
        this.jwtService = jwtService;
    }

    @GetMapping
    public ResponseEntity<List<Stub>> getAllStubs(
            @RequestHeader(value = "Authorization", required = false) String authHeader) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();

        return ResponseEntity.ok(stubService.getAllStubs());
    }

    @PostMapping
    public ResponseEntity<Stub> createStub(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @RequestBody Stub stub) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        stub.setOwnerGroup(user.getAdGroup());
        return ResponseEntity.ok(stubService.createOrUpdateStub(stub));
    }

    @PutMapping("/{id}")
    public ResponseEntity<Stub> updateStub(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("id") String id,
            @RequestBody Stub stub) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        stub.setId(id);
        return ResponseEntity.ok(stubService.createOrUpdateStub(stub));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteStub(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("id") String id) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canDelete(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        stubService.deleteStub(id);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping
    public ResponseEntity<Void> clearAllStubs(
            @RequestHeader(value = "Authorization", required = false) String authHeader) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canDelete(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        stubService.clearAllStubs();
        return ResponseEntity.ok().build();
    }

    @PostMapping("/{id}/toggle")
    public ResponseEntity<Stub> toggleStub(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("id") String id) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        return ResponseEntity.ok(stubService.toggleStub(id));
    }
}
