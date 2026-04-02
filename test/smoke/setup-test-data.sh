#!/bin/bash
# Setup realistic test data on a warewulf server for w9s verification.
# Run ON the warewulf server: sudo ./setup-test-data.sh
set -euo pipefail

echo "=== Setting up w9s test data ==="

# Step 1: Create profile hierarchy
echo "[1/6] Creating profiles..."
wwctl profile add compute 2>/dev/null || echo "  compute already exists"
wwctl profile set compute --profile default --cluster production

wwctl profile add gpu 2>/dev/null || echo "  gpu already exists"
wwctl profile set gpu --profile compute --comment "GPU compute nodes with CUDA"
wwctl profile set gpu --kernelargs "quiet crashkernel=no net.ifnames=1 nvidia.modeset=1"

wwctl profile add storage 2>/dev/null || echo "  storage already exists"
wwctl profile set storage --profile default --cluster production --comment "Storage nodes"

echo "  Profiles: default, compute, gpu, storage"

# Step 2: Import OS image
echo "[2/6] Importing Rocky 9 image (this may take a few minutes)..."
if wwctl image list 2>/dev/null | grep -q rocky9; then
  echo "  rocky9 already imported"
else
  wwctl image import docker://ghcr.io/warewulf/warewulf-rockylinux:9 rocky9
  echo "  Building image..."
  wwctl image build rocky9
fi

# Set image on compute profile
wwctl profile set compute --image rocky9

# Step 3: Clean old test nodes, create new ones
echo "[3/6] Creating nodes..."
for old in testnode-01 testnode-02; do
  wwctl node delete "$old" --yes 2>/dev/null || true
done

for i in $(seq 1 8); do
  IP=$((49 + i))
  NAME="compute-$(printf '%02d' $i)"
  wwctl node delete "$NAME" --yes 2>/dev/null || true
  wwctl node add "$NAME" \
    --profile compute \
    --ipaddr "192.168.1.$IP" \
    --hwaddr "00:11:22:33:44:$(printf '%02x' $i)"
done

for g in 1 2; do
  NAME="gpu-$(printf '%02d' $g)"
  IP=$((69 + g))
  wwctl node delete "$NAME" --yes 2>/dev/null || true
  wwctl node add "$NAME" \
    --profile gpu \
    --ipaddr "192.168.1.$IP" \
    --hwaddr "00:11:22:33:55:$(printf '%02x' $g)"
done
echo "  10 nodes created (compute-01..08, gpu-01..02)"

# Step 4: IPMI config on GPU nodes
echo "[4/6] Setting IPMI on GPU nodes..."
wwctl node set gpu-01 --ipmiaddr 10.0.1.70 --ipmiuser admin --ipmipass secret
wwctl node set gpu-02 --ipmiaddr 10.0.1.71 --ipmiuser admin --ipmipass secret

# Step 5: Custom site overlay
echo "[5/6] Creating site-config overlay..."
wwctl overlay create site-config 2>/dev/null || echo "  site-config already exists"
wwctl overlay mkdir site-config /etc/cluster 2>/dev/null || true
echo "cluster_name=production" > /tmp/w9s-config.conf
wwctl overlay import site-config /tmp/w9s-config.conf /etc/cluster/config.conf 2>/dev/null || true

cat > /tmp/motd.ww <<'TMPL'
Welcome to {{.Id}} in cluster {{.ClusterName}}
Image: {{.ImageName}}
Profiles: {{range .Profiles}}{{.}} {{end}}
TMPL
wwctl overlay import site-config /tmp/motd.ww /etc/motd.ww 2>/dev/null || true
rm -f /tmp/w9s-config.conf /tmp/motd.ww

# Step 6: Build overlays
echo "[6/6] Building overlays..."
wwctl overlay build 2>/dev/null || echo "  overlay build skipped (may need image)"

echo ""
echo "=== Setup complete ==="
wwctl node list | head -15
echo "---"
wwctl profile list
echo "---"
wwctl image list
