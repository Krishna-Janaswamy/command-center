package com.example.virtualization.controller;

import com.example.virtualization.model.ApiRegistryEntry;
import com.example.virtualization.model.User;
import com.example.virtualization.service.ApiRegistryService;
import com.example.virtualization.service.JwtService;
import com.example.virtualization.util.SecurityHelper;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/registry")
@CrossOrigin(origins = "*")
public class ApiRegistryController {
    private final ApiRegistryService apiRegistryService;
    private final JwtService jwtService;

    public ApiRegistryController(ApiRegistryService apiRegistryService, JwtService jwtService) {
        this.apiRegistryService = apiRegistryService;
        this.jwtService = jwtService;
    }

    @GetMapping
    public ResponseEntity<List<ApiRegistryEntry>> getAllApis(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @RequestParam(name = "env", required = false) String env,
            @RequestParam(name = "category", required = false) String category) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();

        boolean isAdmin = SecurityHelper.isAdmin(user);
        return ResponseEntity.ok(apiRegistryService.listApis(env, category, user.getAdGroup(), isAdmin));
    }

    @GetMapping("/{id}")
    public ResponseEntity<ApiRegistryEntry> getApiById(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("id") String id) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();

        ApiRegistryEntry api = apiRegistryService.getApi(id);
        if (api == null) return ResponseEntity.notFound().build();

        // Ownership check
        if (!SecurityHelper.isAdmin(user) && !user.getAdGroup().equals(api.getOwnerGroup()) && !"admin".equals(api.getOwnerGroup())) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN).build();
        }

        return ResponseEntity.ok(api);
    }

    @PostMapping
    public ResponseEntity<ApiRegistryEntry> createApi(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @RequestBody ApiRegistryEntry api) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        api.setOwnerGroup(user.getAdGroup());
        ApiRegistryEntry created = apiRegistryService.createApi(api);
        return ResponseEntity.ok(created);
    }

    @PatchMapping("/{id}")
    public ResponseEntity<Void> updateApi(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("id") String id,
            @RequestBody ApiRegistryEntry api) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canWrite(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        ApiRegistryEntry existing = apiRegistryService.getApi(id);
        if (existing == null) return ResponseEntity.notFound().build();
        if (!SecurityHelper.isAdmin(user) && !user.getAdGroup().equals(existing.getOwnerGroup())) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN).build();
        }

        apiRegistryService.updateApi(id, api);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteApi(
            @RequestHeader(value = "Authorization", required = false) String authHeader,
            @PathVariable("id") String id) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canDelete(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        apiRegistryService.deleteApi(id);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping
    public ResponseEntity<Void> deleteAllApis(
            @RequestHeader(value = "Authorization", required = false) String authHeader) {
        User user = SecurityHelper.extractUser(authHeader, jwtService);
        if (user == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        if (!SecurityHelper.canDelete(user)) return ResponseEntity.status(HttpStatus.FORBIDDEN).build();

        apiRegistryService.deleteAllApis();
        return ResponseEntity.ok().build();
    }
}
