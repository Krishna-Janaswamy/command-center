package com.example.virtualization.config;

import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.jdbc.core.JdbcTemplate;

import javax.sql.DataSource;
import java.io.File;

@Configuration
public class DataSourceConfig {

    @Value("${spring.datasource.url:}")
    private String datasourceUrl;

    @Value("${spring.datasource.username:postgres}")
    private String datasourceUsername;

    @Value("${spring.datasource.password:postgres}")
    private String datasourcePassword;

    @Value("${db.path:./data/data.db}")
    private String dbPath;

    @Bean
    public DataSource dataSource() {
        HikariConfig cfg = new HikariConfig();
        if (datasourceUrl != null && !datasourceUrl.isBlank()) {
            cfg.setJdbcUrl(datasourceUrl);
            cfg.setUsername(datasourceUsername);
            cfg.setPassword(datasourcePassword);
            cfg.setMaximumPoolSize(5);
            cfg.setPoolName("postgres-pool");
        } else {
            try {
                File file = new File(dbPath);
                File parent = file.getParentFile();
                if (parent != null && !parent.exists()) {
                    parent.mkdirs();
                }
            } catch (Exception ignored) {
            }
            cfg.setJdbcUrl("jdbc:sqlite:" + dbPath);
            cfg.setMaximumPoolSize(1);
            cfg.setPoolName("sqlite-pool");
            cfg.addDataSourceProperty("journal_mode", "WAL");
        }
        return new HikariDataSource(cfg);
    }

    @Bean
    public JdbcTemplate jdbcTemplate(DataSource ds) {
        return new JdbcTemplate(ds);
    }
}
