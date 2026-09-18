# Changelog

<!-- next version -->

## v0.45.0

### 🛑 Breaking changes 🛑

- `receiver/foo`: Some change relevant to [user,api] (#11)

### 🚩 Deprecations 🚩

- `receiver/foo`:
  - Some change relevant to [user] (#123)
  - Some change relevant to [user,api] (#234)

### 💡 Enhancements 💡

- `receiver/foo`:
  - Some change relevant to [user] (#21)
  - Some change relevant to [api,user] (#333)

### 🧰 Bug fixes 🧰

- `receiver/foo`:
  - Some change relevant to [user] (#32)
  - Some change relevant to [user] (#222)

## v0.44.0

### 🛑 Breaking changes 🛑

- `prometheusexporter`: Automatically rename metrics with units to follow Prometheus naming convention (#8950)

### 💡 Enhancements 💡

- `filterprocessor`: Ability to filter `Spans` (#6341)
- `flinkmetricsreceiver`: add attribute values to metadata #11520

### 🧰 Bug fixes 🧰

- `redactionprocessor`: respect allow_all_keys configuration (#11542)
