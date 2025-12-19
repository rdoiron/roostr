# Umbrel Integration Test Checklist

Use this checklist when testing Roostr on the Umbrel platform.

## Test Session Info

| Field | Value |
|-------|-------|
| **Date** | ______________ |
| **Tester** | ______________ |
| **Umbrel Version** | ______________ |
| **Roostr Version** | 0.1.0 |
| **Device** | ______________ |

---

## 1. Installation & Startup

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| App installs successfully from Umbrel UI | [ ] | [ ] | |
| Container starts without errors | [ ] | [ ] | |
| Health check passes (`/health` returns 200) | [ ] | [ ] | |
| nostr-rs-relay process starts (PID visible in status) | [ ] | [ ] | |
| App proxy routes correctly (accessible on port 8880) | [ ] | [ ] | |
| Logs accessible via Umbrel app log viewer | [ ] | [ ] | |

---

## 2. Setup Wizard

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Redirects to setup wizard on first launch | [ ] | [ ] | |
| npub input validates correctly | [ ] | [ ] | |
| NIP-05 resolution works (e.g., `user@domain.com`) | [ ] | [ ] | |
| Relay name/description saves to config | [ ] | [ ] | |
| Access mode selection works (Private/Paid/Public) | [ ] | [ ] | |
| Optional whitelist additions work | [ ] | [ ] | |
| Completion screen shows correct relay URLs | [ ] | [ ] | |
| Setup completion persists across container restart | [ ] | [ ] | |

---

## 3. Dashboard

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Relay status shows Online/Offline correctly | [ ] | [ ] | |
| Uptime displays and updates | [ ] | [ ] | |
| WebSocket URL is correct and copyable | [ ] | [ ] | |
| Tor URL displays (when `TOR_ADDRESS` set) | [ ] | [ ] | |
| QR codes generate for URLs | [ ] | [ ] | |
| Stats cards show accurate counts | [ ] | [ ] | |
| Event type breakdown cards work | [ ] | [ ] | |
| Recent activity feed updates | [ ] | [ ] | |
| Quick action buttons navigate correctly | [ ] | [ ] | |
| Auto-refresh works (stats update ~30s) | [ ] | [ ] | |

---

## 4. Access Control

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Access mode changes persist | [ ] | [ ] | |
| Whitelist add (npub) works | [ ] | [ ] | |
| Whitelist add (NIP-05) resolves and adds | [ ] | [ ] | |
| Whitelist remove with confirmation works | [ ] | [ ] | |
| Blacklist add works | [ ] | [ ] | |
| Blacklist remove works | [ ] | [ ] | |
| Event counts per pubkey display correctly | [ ] | [ ] | |
| Bulk import works | [ ] | [ ] | |
| Bulk export works | [ ] | [ ] | |
| Config reload (SIGHUP) applies changes | [ ] | [ ] | |

---

## 5. Event Browser

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Events load with pagination | [ ] | [ ] | |
| Filter by kind works | [ ] | [ ] | |
| Filter by author works | [ ] | [ ] | |
| Filter by date range works | [ ] | [ ] | |
| Search by content works | [ ] | [ ] | |
| Event detail modal shows raw JSON | [ ] | [ ] | |
| Delete event with confirmation works | [ ] | [ ] | |
| Deep links work (`?id=xxx`) | [ ] | [ ] | |
| "Mentions me" filter works | [ ] | [ ] | |

---

## 6. Configuration

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Current config loads correctly | [ ] | [ ] | |
| Relay identity edits save (name, desc, contact) | [ ] | [ ] | |
| Rate limits configuration works | [ ] | [ ] | |
| Event policies work (kinds, PoW) | [ ] | [ ] | |
| Validation prevents invalid config | [ ] | [ ] | |
| Reset to defaults works | [ ] | [ ] | |
| Config reload applies changes to relay | [ ] | [ ] | |

---

## 7. Storage Management

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Storage status shows correct usage | [ ] | [ ] | |
| Retention policy settings work | [ ] | [ ] | |
| Manual cleanup works | [ ] | [ ] | |
| NIP-09 deletion requests display | [ ] | [ ] | |
| Database vacuum works | [ ] | [ ] | |
| Storage alerts appear when appropriate | [ ] | [ ] | |

---

## 8. Export & Backup

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Export with kind filter works | [ ] | [ ] | |
| Export with date range works | [ ] | [ ] | |
| NDJSON format export works | [ ] | [ ] | |
| JSON format export works | [ ] | [ ] | |
| Large exports show progress | [ ] | [ ] | |
| Downloaded files are valid/parseable | [ ] | [ ] | |

---

## 9. Sync from Public Relays

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Sync modal opens from dashboard | [ ] | [ ] | |
| Pubkey selection from whitelist works | [ ] | [ ] | |
| Custom relay URL input works | [ ] | [ ] | |
| Event type selection works | [ ] | [ ] | |
| Sync starts and shows progress | [ ] | [ ] | |
| Events import correctly (no duplicates) | [ ] | [ ] | |
| Sync history displays | [ ] | [ ] | |
| Sync cancel works | [ ] | [ ] | |

---

## 10. Paid Relay Access (Lightning)

**Prerequisites:** LND node configured on Umbrel

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| LND connection auto-detects on Umbrel | [ ] | [ ] | |
| Lightning status shows connected/disconnected | [ ] | [ ] | |
| Pricing configuration saves | [ ] | [ ] | |
| Public signup page loads (`/signup`) | [ ] | [ ] | |
| Invoice generation works | [ ] | [ ] | |
| QR code displays for invoice | [ ] | [ ] | |
| Payment detection works | [ ] | [ ] | |
| Auto-whitelist on payment works | [ ] | [ ] | |
| Paid users list shows entries | [ ] | [ ] | |
| Revenue summary displays | [ ] | [ ] | |
| Subscription expiry tracking works | [ ] | [ ] | |

---

## 11. Statistics

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Events over time chart renders | [ ] | [ ] | |
| Events by kind chart renders | [ ] | [ ] | |
| Top authors list displays | [ ] | [ ] | |
| Time range selector works (Today/7d/30d/All) | [ ] | [ ] | |
| Data matches actual event counts | [ ] | [ ] | |

---

## 12. Relay Controls

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Status shows PID, memory, uptime | [ ] | [ ] | |
| Reload button sends SIGHUP to relay | [ ] | [ ] | |
| Restart button restarts relay process | [ ] | [ ] | |
| Log viewer shows relay logs | [ ] | [ ] | |
| Real-time log streaming works (SSE) | [ ] | [ ] | |

---

## 13. Support Page

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Lightning address QR displays | [ ] | [ ] | |
| Bitcoin address QR displays | [ ] | [ ] | |
| WebLN tip button works (if available) | [ ] | [ ] | |
| Help links are correct | [ ] | [ ] | |
| Version displays correctly | [ ] | [ ] | |

---

## 14. Umbrel-Specific Integrations

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| `TOR_ADDRESS` env var populates Tor URL | [ ] | [ ] | |
| LND macaroon mounts correctly (`/lnd`) | [ ] | [ ] | |
| LND host/port env vars work | [ ] | [ ] | |
| Data persists in `/data` volume across restarts | [ ] | [ ] | |
| App appears correctly in Umbrel dashboard | [ ] | [ ] | |
| exports.sh variables available to other apps | [ ] | [ ] | |

---

## 15. Nostr Client Connectivity

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| Connect client via local WebSocket URL | [ ] | [ ] | Client: ______ |
| Connect via Tor URL from external network | [ ] | [ ] | |
| Post events from client, appear in relay | [ ] | [ ] | |
| NIP-42 auth works for whitelisted users | [ ] | [ ] | |
| Non-whitelisted users rejected (private mode) | [ ] | [ ] | |

---

## 16. Edge Cases & Error Handling

| Test | Pass | Fail | Notes |
|------|:----:|:----:|-------|
| App handles relay crash gracefully | [ ] | [ ] | |
| App handles database corruption gracefully | [ ] | [ ] | |
| App handles network interruptions | [ ] | [ ] | |
| Large event imports don't OOM | [ ] | [ ] | |
| Config validation prevents breaking changes | [ ] | [ ] | |

---

## Summary

| Category | Passed | Failed | Total |
|----------|--------|--------|-------|
| 1. Installation & Startup | | | 6 |
| 2. Setup Wizard | | | 8 |
| 3. Dashboard | | | 10 |
| 4. Access Control | | | 10 |
| 5. Event Browser | | | 9 |
| 6. Configuration | | | 7 |
| 7. Storage Management | | | 6 |
| 8. Export & Backup | | | 6 |
| 9. Sync | | | 8 |
| 10. Paid Access | | | 11 |
| 11. Statistics | | | 5 |
| 12. Relay Controls | | | 5 |
| 13. Support Page | | | 5 |
| 14. Umbrel Integrations | | | 6 |
| 15. Client Connectivity | | | 5 |
| 16. Edge Cases | | | 5 |
| **TOTAL** | | | **112** |

---

## Issues Found

| # | Category | Description | Severity | Status |
|---|----------|-------------|----------|--------|
| 1 | | | | |
| 2 | | | | |
| 3 | | | | |

---

## Notes

_Additional observations, performance notes, or suggestions:_
