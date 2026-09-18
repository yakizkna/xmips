> **Language / 语言**: [中文](IO_SYSCALL_SPEC.md) · [English](IO_SYSCALL_SPEC.en.md)

# Xmips File and Network Socket Feature Specification

> Goal: add **file I/O** (real host directory mapping) and **real network client sockets** (TCP client, client-only, no bind/listen) to xmips.
> Finalized direction: **all four syscalls use INT (15~18) and are natively executed by the dispatcher** (no `.scp`, no channel registers); **networking uses real TCP with optional TLS; blocking/non-blocking is determined by single-process mode; both are protected by a timeout**.
> Status: implemented and verified (HTML fetch, file read/write, deadlock prevention all pass).

---

## 1. Background and Fixed Constraints

### 1.1 Current State
- Syscall: user `MOV #10 call_number; INT`, dispatcher handles it via `r-10`. Existing `INT 13` (output string) / `INT 14` (input buffer) are also **natively executed by the dispatcher** (originally `.scp`, refactored, see note in §2). Only `INT 10` (bubble sort) remains a `.scp` system function.
- Precedent for native syscalls: `SYSR/SYSW/WAKE/SET` are executed directly by the dispatcher in `swap2`'s `case 5/6/7` (no system-function process created).
- Key constraint: `INT` pushes the caller's `#0~#14`, PC, and flag onto the stack; they are popped from the stack on resume. **Therefore the return value of a dispatcher-native syscall must be written into the `#13` slot of the caller's stack** (reusing `case 3`'s `up()` write-back: `S.write(sp-4, #13)`); writing to a GM register would be overwritten.
- `.scp` (ABC assembly) can only address `#0~#14`, and cannot reach `#15~#19`. This is the main reason for dropping `.scp` in this feature.

### 1.2 Finalized Decisions
| Decision | Content |
|------|------|
| Interface form | All unified as INT (15=open,16=read,17=write,18=close), consistent for the user |
| Implementation locus | All native in dispatcher/fsdev (SYSR style), no `.scp`, no `#17/#18` channel registers |
| File storage | Real host directory `disk/` mapping |
| Networking | Real TCP client; **optional TLS** (`tlssock:` forced / `sock:` auto on port 443); **blocking/non-blocking determined by single-process mode, both protected by a `sockTimeout` timeout** |
| Scope | Only read/write/open/close; bind/listen/server are not done |
| | |

---

## 2. Interface (ABI)

Unified: `#10`=call number, `#11/#12`=arguments, `#9`=fd/return value. (All within `#0~#14`, reliable across INT.)

| Call number | Function | Inputs | Return `#9`=fd/result |
|--------|------|------|-----------|
| 15 | open | `#11`=path string address (data segment, 0-terminated), `#12`=mode | fd (≥0); <0 error code |
| 16 | read | `#11`=buffer address, `#12`=max length, `#9`=fd | number of bytes actually read (0=no data/EOF) |
| 17 | write | `#11`=buffer address, `#12`=length to write, `#9`=fd | number of bytes actually written |
| 18 | close | `#9`=fd | 0 success; <0 error code |

> **fd register is `#9`**: `#13` is not used because `#13` is the output register — user `MOV #13 x` would trigger a stdout side effect. `#9` is a general-purpose register with no side effects, and is within the INT push/pop range (`#0~#14`), reliable across syscalls.

- **mode `#12`**: 0=read-only, 1=write (truncate), 2=read/write append (only valid for files).
- **path prefix**:
  - `"sock:host:port"` → TCP client; **TLS is automatically enabled on port 443** (`ServerName`=host, handshake constrained by `sockTimeout`). cupa need not care about TLS.
  - `"tlssock:host:port"` → TCP client, **TLS forced**.
  - otherwise → file (confined to the virtual disk).
- Error codes (negative): -1 open failed/escape, -2 invalid fd, -3 socket connect failed, -4 write buffer full, -5 socket closed/read error, the rest see §7.
- `read` 0 semantics for sockets: **0 = no data currently (non-blocking returns immediately)**, distinct from the negative error code for "connection closed". For files, 0=EOF.

User program example (write a file then read it back; fd stored in #1, moved into #9 before use):

```
MOV #11 PATH        ~ data-segment path string
MOV #12 1           ~ mode 1=write
MOV #10 15
INT                 ~ open → #9=fd
MOV #1 #9           ~ stash fd
MOV #11 BUF
MOV #12 LEN
MOV #9 #1           ~ restore fd (writing #9 has no side effect)
MOV #10 17
INT                 ~ write → #9=written count
MOV #9 #1
MOV #10 18
INT                 ~ close (#9 is fd)
```

---

## 3. Implementation Mechanism (dispatcher-native)

### 3.1 New `case 15/16/17/18` in `swap2`
Explicitly intercept 15~18 in the dispatcher main loop's switch, before `default` (r>=10 goes to `.scp`). Execution flow (modeled after `case 5` SYSR):

```
case 15..18:
    resp := fsdev.do(r-10, /#11 #12 #9 args, caller MData)
    // write the result into the caller stack's fdReg(#9) slot (reuse case3's stack-slot write-back mechanism)
    sp := d.runb.pptr.S.SP
    val := resp; if resp == fdNoData { val = 0 }   // socket no data → user reads 0
    d.runb.pptr.S.write(sp-(17-fdReg), val)        // fdReg=9 → sp-8
    if resp == fdNoData && !singleProc {
        d.readyb.enQueue(d.runb)                   // multi-process: put at ready tail, yield time slice
    } else {
        d.readyb.insertToHead(d.runb)              // has data / single-process: return to queue head and continue
    }
    RES print; r = 1
```

- **Yield time slice when no data**: under multi-process, when a socket read has no data (`fdNoData`), the dispatcher puts the process **back at the ready tail** and then schedules; other processes run ahead, and this process retries later — avoiding busy-spin without blocking round-robin. In single-process mode it goes through `insertToHead` (no other processes, yielding is meaningless).

- No system-function process is created; no PCB consumption, no scheduling round-trip, no `.scp`.
- open/read/write/close may be called by any user process (consistent with `INT 13/14`, **without** SYSR's priority guard).

### 3.2 Why the Return Value Is Written to a Stack Slot
`INT` has already pushed `#0~#14` onto the caller's stack; on resume, exer pops them completely and overwrites GM. So writing GM is ineffective. Writing to the corresponding stack slot is what actually takes effect after resume: the slot within frame for GM[n] = `sp-(17-n)`, fdReg(9)=`sp-8`. This mechanism is fully homologous with the existing `case 3` INT 14 write-back (#13=`sp-4`).

### 3.3 Alternative (recorded, not implemented)
Changing read/write to `.scp` byte-by-byte (following 13/14) to enhance teaching, but this requires freeing byte-channel registers within `#0~#14` and open's fd return also needs a register, adding complexity and error surface. Finalized as dropped, keeping unified native.

---

## 4. fsdev and fd Table (new `fsdev.go`)

```
type fdEntry struct {
    kind   int          // FILE or SOCK(TCP)
    path   string       // path or sock:host:port
    host   *os.File     // file
    mode   int
    offset int64        // file read/write offset (maintained independently)
    conn   net.Conn     // socket connection
}
```
- fd starts at 3 (0/1/2 reserved for stdio semantics).
- File bytes ↔ data-segment bytes correspond one-to-one (one int stores the low 8 bits of one byte).
- Concurrency safety: the simulator is single-threaded, no locking needed.

### fsdev Operations
- **open**: prefix `sock:`/`tlssock:` → `net.DialTimeout("tcp", host:port, sockTimeout)`; with `tlssock:` or port 443, wrap with `crypto/tls` and handshake (`ServerName`=host, handshake constrained by `sockTimeout`); otherwise file → `os.OpenFile` after "path cleanup + `diskRoot` confinement".
- **read**: file → `ReadAt` by `offset`; socket → blocking wait/non-blocking determined by single-process mode, returning `fdNoData` on timeout (see §5).
- **write**: file → `WriteAt` updates `offset`; socket → subject to `sockTimeout` write timeout, returns actual bytes.
- **close**: close host/conn, free fd (mark slot reusable).

---

## 5. Real Networking: Blocking/Non-blocking Semantics (Determined by Single-Process Mode)

Whether read blocks depends on **single-process mode**:

Single-process mode = running `./xmips xxx.cupa` directly, or `run.list` having exactly **1** file (computed at startup, written to global `singleProc`).

| Scenario | `read`(socket) | write | connect/handshake(open) |
|------|--------------|-------|--------------------|
| **Single-process** | **Blocking** wait for data, protected by `sockTimeout` timeout; on **timeout returns 0** (no more permanent hang) | `sockTimeout` write timeout | timeout `sockTimeout` |
| **Multi-process** | **Non-blocking**: `SetReadDeadline(now)`; returns N if data present, **returns 0 on no data and yields the time slice** (put at ready tail) | `sockTimeout` write timeout | timeout `sockTimeout` |

- **Multi-process yields time slice**: when a socket read has no data, the sentinel `fdNoData(-6)` is returned; once recognized, the dispatcher puts the process back at the **ready tail** (`enQueue`), then keeps scheduling other processes; this process is later rescheduled to retry — neither busy-spin nor blocking round-robin.
- **Single-process timeout**: `read` sets `SetReadDeadline(now+sockTimeout)`; on timeout it likewise returns `fdNoData`, which the dispatcher converts to 0 and writes back (no other process to yield to, goes through `insertToHead`). Prevents permanent blocking when a server's keep-alive doesn't close the connection.
- No fixed short window (e.g. 0.1ms): multi-process uses `deadline=now` (**true immediate return**), single-process uses `sockTimeout` absolute timeout.
- Connection closed/IO error → returns `-5`, distinct from "no data returns 0"; **reset the deadline** after reading.
- Results depend on the host's real network, no longer fully reproducible (accepted in the finalization).

---

## 6. Virtual Disk (Configuration Items)

New additions to config.ini:

| Parameter | Value | Description |
|------|---|------|
| diskRoot | `disk` | Virtual disk directory (relative to run directory), auto-creates `dist/disk/` |
| sockTimeout | `2000` | Socket connect/handshake/read-write timeout (ms) |

`open` file paths are forcibly confined within `diskRoot`: `filepath.Clean` + reject absolute paths/`..` escape (§7 error -1).

---

## 7. Error Codes and RES (New)

`#9` returns error codes (negative): -1 open failed/escape, -2 invalid fd, -3 socket connect/handshake failed, -4 write buffer full, -5 socket closed/read-write error. The sentinel `fdNoData=-6` is for internal use only (socket read has no data); it is uniformly converted to 0 when written back to the user.

New system status message table entries:

| RES | Level | Message |
|-----|-------|---------|
| 200 | 3 | fsdev::open: path escape / open failed |
| 201 | 3 | fsdev::read: invalid fd |
| 202 | 3 | fsdev::write: invalid fd / write failed |
| 203 | 3 | fsdev::close: invalid fd |
| 204 | 3 | fsdev::net: socket connect / io error |

---

## 8. Go Files to Modify

| File | Change |
|------|------|
| `src/fsdev.go` (new) | fd table, file/Socket implementation, path cleanup, TLS (`crypto/tls`), blocking/non-blocking/timeout read/write, sentinel `fdNoData` |
| `src/dispatcher.go` | add `case 15/16/17/18` in `swap2`; `#9` stack-slot write-back; `fdNoData` multi-process tail-yield of time slice |
| `src/main.go` | parse `diskRoot/sockTimeout`; compute `singleProc` and inject into fsdev, global constants |
| `src/global.go` | global `diskRoot/sockTimeout/singleProc` |
| `dist/config.ini` | add §6 configuration items |
| `README.md` / `使用说明.txt` | add syscall table (15~18), TLS, virtual disk, blocking/timeout/yield explanations |

`assembler.go`, `interpreter.go` need **no additional opcode/system-function requirements**; among `sysfun/INT_xx.scp`, **only `INT_10.scp`** (bubble sort) is kept, `INT_13.scp`/`INT_14.scp` have been deleted (replaced by dispatcher-native execution).

---

## 9. Verification Results

1. File write → read back: write `disk/test.txt`, then open the same file and echo-compare a read ✅ (`userfile/FILEIO.cupa`)
2. Escape protection: `open "../etc/passwd"` rejected (error -1).
3. **HTTPS fetch + write file** ✅ (`userfile/NETTOUR.cupa`): `sock:ace.yakidev.top:443` auto TLS → send HTTP GET → chunked read → write `disk/tour.html`, 22KB complete HTML.
4. **Single-process deadlock prevention** ✅ (`userfile/HTIMEOUT.cupa` + a local silent server that accepts but sends no data and doesn't close for 30s): the program exits within ~4s, proving the `sockTimeout` read timeout works.
5. Regression: `SUM`, `BUBBLE_INT`, `FILEIO` normal; `INT 13/14` unaffected ✅
6. Word length / multi-process time-slice yield: dedicated verification pending.

---

## 10. Open Items / Review Points

- [x] Whether the `sockTimeout` default of 2000ms is appropriate (used; both single-process timeout and multi-process non-blocking are protected by it).
- [ ] Non-blocking read's "no data" returns the sentinel via `fdNoData`; the dispatcher **yields by putting at the tail under multi-process** — relying only on `enQueue` spins idly when the ready queue has only 1 process, which is acceptable (multi-process mode usually has multiple processes).
- [x] Convention of negative error codes and ≥0 for success.
- [x] fd starting at 3 (0/1/2 reserved for stdio).
- [ ] `read` 0 semantics: file=EOF, socket=no data. If files need to also distinguish "read 0 bytes" from "EOF", currently both are handled uniformly as 0.