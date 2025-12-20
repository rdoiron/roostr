# STARTOS-007: Test on StartOS - Plan

## Task Overview
Test the Roostr .s9pk package on StartOS to verify it works correctly before marketplace submission.

## Approach
Set up StartOS in virt-manager (KVM/QEMU) for full platform testing.

---

## Phase 1: Environment Setup

### 1.1 Install virt-manager (if needed)
```bash
sudo apt install virt-manager qemu-kvm libvirt-daemon-system
sudo usermod -aG libvirt,kvm $USER
# Log out and back in for group changes
```

### 1.2 Download StartOS ISO
- Go to: https://github.com/Start9Labs/start-os/releases
- Download latest `startos-x86_64.iso` (or similar naming)

### 1.3 Create VM in virt-manager
1. Launch virt-manager
2. Click "Create a new virtual machine"
3. Select "Local install media (ISO image)"
4. Browse to downloaded ISO
5. **Important:** Uncheck "Automatically detect", select "Ubuntu 22.04 LTS" or "Generic Linux"
6. Allocate resources:
   - **Minimum:** 4GB RAM, 2 vCPUs, 64GB disk
   - **Recommended:** 8GB RAM, 4 vCPUs, 128GB disk
7. Name the VM (e.g., "startos-test")
8. Click "Finish"

### 1.4 Install StartOS
1. Start the VM
2. Select "Install StartOS" from boot menu
3. Select the virtual disk when prompted
4. Confirm disk erasure
5. Wait for installation to complete
6. Reboot when prompted (may need to remove ISO from VM settings)

### 1.5 Initial StartOS Setup
1. Access StartOS via browser (check VM console for IP/URL)
2. Select "Start Fresh"
3. Choose storage drive
4. Set a secure password
5. Complete setup wizard

---

## Phase 2: Build Package

Packages are built via GitHub Actions to avoid local SDK installation issues.

### 2.1 Trigger Build (Option A: Release Tag)
```bash
git tag v0.1.0
git push origin v0.1.0
```
This creates a GitHub Release with `roostr.s9pk` attached.

### 2.2 Trigger Build (Option B: Manual)
1. Go to GitHub > Actions > Release workflow
2. Click "Run workflow"
3. Enter version (e.g., `0.1.0`)
4. Wait for build to complete

### 2.3 Download Package
1. Go to [Releases](https://github.com/rdoiron/roostr/releases) or Actions artifacts
2. Download `roostr.s9pk`
3. Verify checksum (optional):
   ```bash
   sha256sum -c roostr.s9pk.sha256
   ```

---

## Phase 3: Test on StartOS

### 3.1 Sideload Package
1. Open StartOS dashboard in browser
2. Navigate to **System > Sideload Service**
3. Upload `roostr.s9pk` file
4. Monitor installation progress
5. Verify no errors in install logs

### 3.2 Startup Testing
- [ ] Start Roostr service from StartOS dashboard
- [ ] Verify health check passes (green status indicator)
- [ ] Check container logs for errors (StartOS > Roostr > Logs)
- [ ] Verify web interface accessible (click through from StartOS)
- [ ] Note Tor URL and LAN URL provided by StartOS

### 3.3 Functional Testing
- [ ] Complete Roostr setup wizard (enter npub, relay name, access mode)
- [ ] Verify dashboard loads with correct data
- [ ] Test relay connectivity with a Nostr client (use Tor or LAN URL)
- [ ] Send a test event to the relay
- [ ] Verify event appears in event browser
- [ ] Test whitelist add/remove
- [ ] Test sync functionality (if public relays accessible)
- [ ] Test configuration changes and relay reload
- [ ] Test storage management page

### 3.4 StartOS Integration Testing
- [ ] Verify Tor .onion URL works (may need Tor browser)
- [ ] Test LAN access with SSL
- [ ] Test both interfaces:
  - Port 80 (Tor) → 8080 (web UI)
  - Port 7000 → 7000 (relay WebSocket)

### 3.5 Backup/Restore Testing
1. Create events and configuration changes
2. Navigate to **StartOS > Roostr > Backup > Create Backup**
3. Verify backup completes successfully
4. Stop Roostr service
5. Delete/reset data (or note current state)
6. Restore from backup
7. Verify all data restored correctly:
   - Events present
   - Configuration preserved
   - Whitelist entries intact

### 3.6 Stability Testing
- [ ] Restart Roostr via StartOS UI
- [ ] Verify data persists after restart
- [ ] Reboot entire StartOS VM
- [ ] Verify Roostr auto-starts
- [ ] Verify data persists after system reboot

---

## Phase 4: Document Results

### 4.1 Record Any Issues
- Note bugs, errors, or unexpected behavior
- Capture logs for debugging
- Document workarounds if needed

### 4.2 Update TASKS.md
- Mark STARTOS-007 as complete if testing passes
- Create follow-up tasks for any issues found

---

## Files Involved
- `platforms/startos/Dockerfile` - Build configuration
- `platforms/startos/manifest.yaml` - StartOS manifest
- `platforms/startos/entrypoint.sh` - Container startup
- `platforms/startos/scripts/health-check.sh` - Health probe
- `platforms/startos/scripts/backup.sh` - Backup script
- `platforms/startos/scripts/restore.sh` - Restore script

---

## Success Criteria
1. Package builds without errors
2. Installs successfully via sideload
3. Health check passes (green status)
4. Web UI accessible and functional
5. Relay accepts connections and events
6. Backup/restore works correctly
7. Data persists across restarts

---

## Prerequisites for CI/CD

Before running release builds, configure these GitHub repository secrets:
- `DOCKERHUB_USERNAME` - Docker Hub username
- `DOCKERHUB_TOKEN` - Docker Hub access token

---

## Sources
- [StartOS VM Installation Guide](https://community.start9.com/t/installing-startos-on-virtual-machine-manager-virt-manager/77)
- [Start9 Developer Docs](https://docs.start9.com/)
- [Start9 SDK GitHub Action](https://github.com/Start9Labs/sdk)
- [StartOS GitHub Releases](https://github.com/Start9Labs/start-os/releases)
