# SleepGuard — Electronics and Wiring

Parts list and wiring reference. Use this to verify you have everything before each phase.

> **Production (June 2026):** PIR on **GPIO17 (pin 11)**. Pi Camera v2 on CSI — use **`rpicam-still`** on current Pi OS (`rpicam-apps`). Capture is triggered manually from the PWA, not on every motion. See [checklist.md](checklist.md) Phase F.

---

## Your kit (confirmed)

| Component | Your spec | MVP status |
|-----------|-----------|------------|
| Raspberry Pi 4 | 4 GB RAM | Ready |
| MicroSD card | 32 GB Class 10 | Ready |
| Power supply | 5 V / 3 A USB-C | Ready |
| PIR sensor | HC-SR501 | Ready |
| Breadboard + jumper wires | M-F (male–female) | Ready — use breadboard between Pi and PIR |
| Pi case | Yes | Ready |
| Cooling | No heatsink / fan | OK for dev; monitor temp if running 24/7 |
| Resistors | 300 Ω | Ready — fine for LED current limiting (330 Ω ideal; 300 Ω is close) |
| LEDs | Yes | Ready — optional visual alert / debug on GPIO |
| Camera | Pi Camera Module **v2** + ribbon cable | Ready — manual capture (Phase F) |
| Network | Stable Wi‑Fi | Ready — no Ethernet required |

### Gaps (optional, not blocking MVP)

| Item | Impact | Workaround for SleepGuard |
|------|--------|---------------------------|
| USB speaker / headphones | No dedicated audible alert hardware | Phase 2: browser dashboard refresh on phone; optional LED blink on GPIO |
| Heatsink / fan | Pi may run warmer under load | Case ventilation + short dev sessions; add cooling later if 24/7 |
| SD card reader (on dev PC) | Only needed to flash OS from laptop | Use if imaging card off-Pi |
| Buzzer + transistor | Not in your kit | Skip; use LED or LAN dashboard alert first |

**Alert plan with your parts:** GPIO **LED blink** on motion (Phase 2 optional) + **phone dashboard** on Wi‑Fi as primary alert. Camera Module v2 in Phase 4.

---

## Bill of materials

### Required (Phase 1+)

| Item | Qty | Purpose | Notes |
|------|-----|---------|-------|
| Raspberry Pi 4 | 1 | Main controller | 2 GB+ RAM fine; 4 GB comfortable for build |
| MicroSD card | 1 | OS + app + data | 16 GB+ Class 10 |
| Power supply | 1 | Pi power | Official 5 V / 3 A USB-C recommended |
| PIR motion sensor | 1 | Motion detection | HC-SR501 or AM312 (see below) |
| Jumper wires | 3+ | PIR → Pi GPIO | M-F + breadboard works (your setup) |
| Breadboard (optional) | 1 | Prototyping | Helpful for bench testing |

### Recommended

| Item | Qty | Purpose | Notes |
|------|-----|---------|-------|
| Pi case with ventilation | 1 | Mounting + protection | |
| Heat sinks or fan | 1 set | Thermal headroom | Optional for 24/7 run |
| Ethernet cable | 1 | Stable network | Or reliable Wi‑Fi |

### Phase 4 — Camera (pick one)

| Item | Qty | Purpose | Notes |
|------|-----|---------|-------|
| Raspberry Pi Camera Module v2/v3 | 1 | Snapshot on motion | **You have v2** + cable → CSI port |
| **OR** USB webcam | 1 | Snapshot via V4L2/`fswebcam` | Easier on desk; bulkier |

### Optional — Local alert hardware

| Item | Qty | Purpose | Notes |
|------|-----|---------|-------|
| Passive buzzer + transistor | 1 | Audible alert | GPIO-driven; needs resistor |
| **OR** USB speaker | 1 | Alert via `aplay` | Software-only wiring |
| Small LED + 300–330 Ω resistor | 1 | Visual debug / alert | You have 300 Ω + LEDs — use one on GPIO27 |

---

## Sensor options

### HC-SR501 (common PIR)

- **Logic:** Digital OUT — HIGH when motion detected
- **Voltage:** 5 V module; OUT often 3.3 V compatible on Pi (verify your module)
- **Adjustments:** Sensitivity pot, time delay pot (start with ~10–30 s delay on module)
- **Warm-up:** ~60 s after power-on before reliable detection

### AM312 (mini PIR)

- **Logic:** Digital OUT, 3.3 V friendly
- **Range:** Shorter than HC-SR501; good for crib/near field
- **No pots:** Fixed sensitivity/delay on board

**SleepGuard default assumption:** HC-SR501 or equivalent active-HIGH on motion.

---

## Wiring — PIR to Raspberry Pi 4

### Pin plan (default in architecture)

| PIR pin | Connect to | Pi physical pin |
|---------|------------|-----------------|
| VCC | 5 V | Pin 2 or 4 (5 V) |
| GND | Ground | Pin 6, 9, 14, 20, 25, 30, 34, or 39 |
| OUT | GPIO input | **GPIO17** — physical pin **11** |

```text
    PIR (HC-SR501)              Raspberry Pi 4
    ┌─────────────┐             ┌─────────────┐
    │ VCC         ├─────────────┤ 5V    (pin 2)│
    │ OUT         ├─────────────┤ GPIO17(pin11)│
    │ GND         ├─────────────┤ GND   (pin 6)│
    └─────────────┘             └─────────────┘
```

### GPIO numbering

- Use **BCM / GPIO numbering** in software (GPIO17), not physical pin number alone.
- Config flag: `-gpio-pin 17` (default)

### Before powering on

- [ ] Double-check VCC is 5 V pin, not 3.3 V (PIR usually needs 5 V)
- [ ] OUT goes to a GPIO input only, not 5 V or GND
- [ ] No short between VCC and GND
- [ ] Jumper wires firmly seated
- [ ] Breadboard: pins on **opposite sides of the center gap are not connected**
- [ ] **Do not use GPIO2 (physical pin 3)** — on this Pi it reads HIGH due to hardware pull-up

### Verified wiring (SleepGuard Pi — Phase 1)

| Wire color | PIR / function | Pi connection |
|------------|----------------|---------------|
| Orange | VCC (+5 V) | Pin 4 (5 V) |
| Gray | GND | Pin 6 (GND) |
| White | OUT | Pin 11 (GPIO17) |

HC-SR501 module settings that worked:

- Jumper on **H** (repeat trigger)
- **Time Delay** pot fully **counter-clockwise** (otherwise OUT can stay HIGH)
- Allow **~60 s** warm-up after power-on

Diagnostic on Pi:

```bash
watch -n 0.2 pinctrl get 17
```

---

## Wiring — Optional LED alert (you have parts)

Visual alert on motion — good for bench testing and portfolio demo.

| Part | Connect to |
|------|------------|
| LED anode (+) | **GPIO27** (pin 13) via **300 Ω** resistor |
| LED cathode (−) | GND (e.g. pin 14) |

- Use a **second GPIO** from the PIR (GPIO17) so sensor and LED do not share a pin.
- In software: set GPIO27 HIGH briefly when alert fires.
- Long leg of LED = anode (+); short leg = cathode (−).

No buzzer in your kit — skip hardware audio unless you add a USB speaker later.

---

## Wiring — Pi Camera Module v2 (your module)

1. **Power off** the Pi.
2. Lift CSI connector tab, insert ribbon cable (contacts face **HDMI side** on Pi 4).
3. Secure tab, route cable so it is not strained.
4. Enable camera: `sudo raspi-config` → Interface Options → Camera (enable legacy stack on older OS if `libcamera` fails).
5. Test:
   - **Bookworm:** `libcamera-still -o test.jpg`
   - **Legacy / Module v2:** `raspistill -o test.jpg`

Camera Module **v2** is fully supported; no USB webcam needed.

---

## Wiring — USB webcam

- Plug into Pi USB port.
- Test: `fswebcam -r 1280x720 test.jpg` or `v4l2-ctl --list-devices`
- No GPIO wiring required.

---

## Power and placement

| Topic | Guidance |
|-------|----------|
| PIR placement | 2–4 m from crib, angled to avoid false triggers from HVAC |
| Pi placement | Outside crib, cable-managed, not reachable by child |
| PIR false triggers | Keep away from direct sunlight, heaters, fans |
| Night use | HC-SR501 time pot — shorter delay reduces repeat triggers |

---

## Inventory checklist

| # | Component | Have it? | Notes |
|---|-----------|----------|-------|
| 1 | Raspberry Pi 4 (4 GB) | [x] | |
| 2 | Pi power supply (5 V / 3 A USB-C) | [x] | |
| 3 | MicroSD 32 GB Class 10 | [x] | Reader only if flashing from PC |
| 4 | PIR sensor HC-SR501 | [x] | |
| 5 | Jumper wires (M-F) | [x] | |
| 6 | Breadboard | [x] | |
| 7 | Pi case | [x] | |
| 8 | Pi Camera Module v2 | [x] | |
| 9 | Camera ribbon cable | [x] | |
| 10 | 300 Ω resistors + LEDs | [x] | Optional GPIO alert |
| 11 | Stable Wi‑Fi | [x] | Replaces Ethernet for LAN dashboard |
| 12 | Heatsink / fan | [ ] | Optional — not required for MVP |
| 13 | USB speaker (audible alert) | [ ] | Optional — use dashboard + LED first |
| 14 | Buzzer + transistor | [ ] | Not needed |

---

## Shopping reference (if something is missing)

Generic search terms (no vendor lock-in):

- `Raspberry Pi 4 Model B`
- `HC-SR501 PIR sensor module`
- `Raspberry Pi Camera Module 3` or `USB webcam UVC`
- `Dupont jumper wires female-female`
- `Raspberry Pi 4 power supply 5V 3A USB-C`

---

## Safety notes

- This is a **hobby / monitoring** project, not a medical or safety-critical device.
- Do not place loose wiring or the Pi inside the crib.
- Use a stable power supply; avoid powering the Pi from an under-rated USB port.
- If using 5 V on a GPIO by mistake, you can damage the Pi — verify wiring twice.

---

## Related documents

- [architecture.md](architecture.md) — GPIO pin in config, sensor interface
- [checklist.md](checklist.md) — Pi bring-up steps per phase
- [implementation-plan.md](implementation-plan.md) — when camera is needed
