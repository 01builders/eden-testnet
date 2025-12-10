# UPGRADES.md

## ev-reth v0.1.0 => v0.2.0

---

### **Step 1: Remove the Deprecated CLI Flag**

```bash
--ev-reth.enable
```

---

### **Step 2: Backup the Current genesis.json**

---

### **Step 3: Edit genesis.json**

#### Add the following **keys** under `"config"`:

**a. Osaka activation time**
```json
"osakaTime": 1765555200
```

**b. Evolve configuration**
```json
"evolve": {
  "baseFeeSink": "0x2Be4490805BE0D0500b38A6687d94738a26fFC22",
  "baseFeeRedirectActivationHeight": 25000000,
  "mintPrecompileAdmin": "0xc259e540167B7487A89b45343F4044d5951cf871",
  "mintPrecompileActivationHeight": 25000000,
  "contractSizeLimit": 131072,
  "contractSizeLimitActivationHeight": 25000000
}
```

Ensure valid JSON syntax (no trailing commas).

---

### **Step 4: Restart ev-reth**

---

### **Step 5: Verify Upgrade**
Check ev-reth and ev-node logs

Ensure:
- No `--ev-reth.enable` errors
- Blocks are produced
