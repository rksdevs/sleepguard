# SleepGuard: Motion Sensor in Go

> Local-first baby sleep motion monitor.  
> Time budget: 2 hours per day.  
> Current Go level assumed: variables, functions, loops, if/else, errors, structs, interfaces.  
> Rule: no AI-generated code. Learn the concept, then write it by hand.

## How to use this plan

Each day has:
- `Learn` = what concept to read or understand
- `Build` = what to implement by hand
- `Packages / modules` = what to use that day
- `Done when` = the green light for that day

Keep it simple:
- local first
- motion alert first
- camera second
- telemetry third
- cloud/Kubernetes later, only if the local product is solid

## Suggested repo layout

```text
cmd/sleepguard/main.go
internal/config/
internal/sensor/
internal/alert/
internal/store/
internal/web/
internal/camera/
internal/telemetry/
```

You do not need all folders on day 1. Create them as the project grows.

## Week 1: Go basics + hardware bring-up

### Day 1
- Learn: Go modules, project layout, `go mod init`, package names, `main` package
- Build: create the project folder, initialize the module, print a startup message, make a tiny config file path plan
- Packages / modules: `go mod`, `os`, `fmt`, `log/slog`
- Done when: the project runs from `main.go` and you know where each future package will live

### Day 2
- Learn: `time`, `flag`, `os`, basic config loading, durations, timestamps
- Build: add command-line flags for device name, alert cooldown, and debug mode
- Packages / modules: `flag`, `time`, `os`, `strconv`
- Done when: you can start the app with different values without editing code

### Day 3
- Learn: structs and methods for modeling sensor events
- Build: create an `Event` struct with fields like timestamp, type, source, and state
- Packages / modules: `time`, `encoding/json`
- Done when: one event can be created, printed, and serialized cleanly

### Day 4
- Learn: interfaces as boundaries between hardware and logic
- Build: define a sensor interface so your app does not depend on one concrete device
- Packages / modules: `context`, `time`
- Done when: your app can talk to a sensor through an interface, not direct hardware calls everywhere

### Day 5
- Learn: install and read external modules, GPIO basics
- Build: wire the PIR sensor to the Pi and read motion states
- Packages / modules: `periph.io/x/host/v3`, `periph.io/x/conn/v3/gpio`
- Done when: motion detection can be read from the Pi in a loop

### Day 6
- Learn: loops around polling, debouncing, and state changes
- Build: add a motion loop that only emits an event when the state changes or cooldown expires
- Packages / modules: `time`, `sync/atomic` if needed
- Done when: one motion does not spam many repeated events

### Day 7
- Learn: basic logging and troubleshooting
- Build: add structured logs for startup, motion detected, cooldown skipped, and sensor error
- Packages / modules: `log/slog`
- Done when: you can read the logs and understand the full sensor flow

## Week 2: Local alerting + HTTP dashboard

### Day 8
- Learn: `net/http` server basics and request handlers
- Build: create a local web server with `/health` and `/status`
- Packages / modules: `net/http`
- Done when: you can open the Pi IP on your phone browser and see a response

### Day 9
- Learn: routing with `http.ServeMux`
- Build: add `/events` and `/config` endpoints
- Packages / modules: `net/http`, `encoding/json`
- Done when: the app can return recent events as JSON

### Day 10
- Learn: HTML templates for simple dashboards
- Build: create a basic local UI that lists recent motion events
- Packages / modules: `html/template`, `net/http`
- Done when: your phone can show a readable local dashboard

### Day 11
- Learn: alerting options and tradeoffs
- Build: choose the first alert channel for phase 1, such as local sound, browser refresh, or LAN-only notification
- Packages / modules: `net/http`, `os/exec` if you trigger a local command
- Done when: one motion event creates a visible or audible alert

### Day 12
- Learn: simple state machines
- Build: add states like `idle`, `motion_detected`, `cooldown`, and `alert_sent`
- Packages / modules: `time`
- Done when: your app behaves predictably through the alert flow

### Day 13
- Learn: keeping mutable state safe at a simple level
- Build: create an in-memory event store for recent history
- Packages / modules: `sync`
- Done when: the dashboard can read the same recent events the sensor loop writes

### Day 14
- Learn: review and cleanup day
- Build: refactor file names, package names, and config values
- Packages / modules: whichever you used in week 2
- Done when: the local-only MVP feels stable enough to demo at home

## Week 3: Concurrency + storage + telemetry

### Day 15
- Learn: goroutines and why they help for sensor/event handling
- Build: split sensor reading, alerting, and UI updates into separate goroutines
- Packages / modules: `sync`, `time`
- Done when: one slow task does not block everything else

### Day 16
- Learn: channels as event pipes
- Build: send motion events through a channel from sensor loop to alert/store workers
- Packages / modules: `chan`, `sync`, `context`
- Done when: your architecture is event-driven instead of one giant loop

### Day 17
- Learn: `context` for shutdown and cancellation
- Build: clean shutdown for the HTTP server and worker goroutines
- Packages / modules: `context`, `net/http`
- Done when: Ctrl+C stops the app without leaving things messy

### Day 18
- Learn: basic persistence choices
- Build: store event history in a simple local format first, such as JSONL or a small file
- Packages / modules: `encoding/json`, `os`, `bufio`
- Done when: events survive a restart

### Day 19
- Learn: when to move from file storage to SQL
- Build: decide whether to stay with JSONL for now or move to SQLite for structured history
- Packages / modules: `database/sql` plus an SQLite driver only if needed
- Done when: you have one clear storage choice and a reason for it

### Day 20
- Learn: telemetry basics and what to measure
- Build: track motion count, alert count, false alerts, uptime, and last event time
- Packages / modules: `sync/atomic`, `time`
- Done when: the dashboard can show useful operational numbers

### Day 21
- Learn: review and consolidation
- Build: clean up the sensor-event-alert pipeline and write notes about what each goroutine does
- Packages / modules: all Week 3 packages
- Done when: you can explain the flow without opening the code

## Week 4: Camera, polish, and optional deployment

### Day 22
- Learn: Pi camera workflow and command-driven capture
- Build: capture a snapshot on motion, not live stream yet
- Packages / modules: `os/exec`
- Done when: a motion event can save or display a still image

### Day 23
- Learn: file handling for images
- Build: attach snapshot metadata to the motion event history
- Packages / modules: `os`, `path/filepath`, `time`
- Done when: each motion event can point to a captured image

### Day 24
- Learn: local preview options
- Build: show the latest snapshot on the dashboard
- Packages / modules: `net/http`, `html/template`
- Done when: you can see the latest camera image from your phone on the LAN

### Day 25
- Learn: reliability and alert spam prevention
- Build: tune cooldowns, rate limits, and error handling
- Packages / modules: `time`, `sync`
- Done when: the system feels calm and trustworthy instead of noisy

### Day 26
- Learn: optional metrics export
- Build: expose a small `/metrics`-style endpoint or a telemetry page
- Packages / modules: `net/http`, optional `github.com/prometheus/client_golang/prometheus`
- Done when: you can inspect the system health from one place

### Day 27
- Learn: Docker basics
- Build: containerize the app so the same binary can run repeatably
- Packages / modules: no new Go package required; focus on runtime packaging
- Done when: the app can start the same way on another machine

### Day 28
- Learn: roadmap review
- Build: document the project, its limits, and next steps: cloud sync, Kubernetes, AI vision
- Packages / modules: none new
- Done when: you have a finished MVP, a clear story, and a next-phase backlog

## Suggested next phases after 4 weeks

- Cloud sync only if other parents actually use it
- Kubernetes only after you have multiple services that justify it
- AI vision only after the motion + camera pipeline is stable and trusted

## What to practice while building

- Go every day: variables, functions, loops, if/else, errors, structs, interfaces, then add concurrency
- DSA every day separately from this project
- Keep project notes short: what changed, what broke, what you learned

## End-of-project resume story

By the end of week 4, you should be able to say:
- I built a local-first baby sleep monitoring system in Go
- It detects motion, triggers alerts, stores event history, and shows telemetry on a dashboard
- I used GPIO hardware, HTTP services, concurrency, and structured logging
- I designed it to grow into cloud sync and smarter vision features later
