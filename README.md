
# GoQRShare

**GoQRShare** is a small command-line (CLI) tool to **share files between your laptop and phone over the same Wi-Fi network**, without cables, accounts, or an internet connection. You run a single command on your laptop, scan the **QR code** with your phone, and then download or upload files through your phone's browser.

> Written in Go. Runs on Windows, macOS, and Linux. No additional installation is required if you use the pre-built binaries from the *Releases* page.

---

## Table of Contents

1. [Features](#features)
2. [How it works](#how-it-works)
3. [Disclaimer and limitations](#disclaimer-and-limitations)
4. [Downloads](#downloads)
5. [Installation](#installation)
6. [Build from source](#build-from-source)
7. [Usage](#usage)
8. [Flag list](#flag-list)
9. [Flag rules and common mistakes](#flag-rules-and-common-mistakes)
10. [Security](#security)
11. [Troubleshooting](#troubleshooting)
12. [Known limitations](#known-limitations)
13. [Verification status](#verification-status)
14. [License](#license)

---

## Features

- **Share a single file** from your laptop to your phone: The phone opens the URL (or scans the QR) and downloads the file.
- **Receive files from your phone** (`--upload`): Select one or multiple files in your phone's browser, and they are saved to your specified folder on the laptop.
- **Mixed mode**: Share a file and receive uploads in a single session.
- **Terminal QR code**: Open the URL without typing.
- **Random token** in the URL (12 characters, 72 bits of entropy) or a custom fixed token.
- **One-time use** (`--once`) and **auto-expiry** (`--ttl`).
- **Upload size limit** (`--max-size`).
- **Bound to a single network interface**: Automatically selects the Wi-Fi/LAN interface (can be forced with `--iface`).
- **Encrypted connection (HTTPS)** with a *self-signed* certificate generated in memory every time the program runs.
- The web page on the phone **does not load external internet resources** (no CDN), so it works without an internet connection.

---

## How it works

GoQRShare runs a small web server on your laptop, **bound only to a single local IPv4 address** (not `0.0.0.0`), and prepares a URL with a random token and a QR code for that URL. The server handles connections over **HTTPS** using a self-signed certificate generated at startup.

```text
Laptop (goqrshare)                              Phone (browser)
  |  1. pick LAN IP, generate token               |
  |  2. print URL + QR in terminal                |
  |<---------- 3. Phone scans QR / types URL -----|
  |----------- 4. page / file is sent ----------->|
  |<---------- 5. (optional) file is uploaded ----|
```

Only devices **on the same network** and **knowing the URL (token)** can access the server. The server stops when you press `Ctrl+C`, when `--ttl` expires, or after one successful transfer if using `--once`.

---

## Disclaimer and limitations

Read this section before using GoQRShare.

1. **Encryption uses a self-signed certificate.** File contents and tokens are encrypted while traversing the Wi-Fi, preventing passive eavesdroppers from reading them. However, this certificate is **not issued by a trusted authority**, so browsers will show a warning, and GoQRShare **cannot prevent active man-in-the-middle (MITM) attacks** by someone intentionally impersonating the server. See the *Encryption (HTTPS)* section for details.
2. **Anyone who knows the URL (token) can download or upload** while the server is running. Do not share the URL or QR code photo with others. Use `--once` and `--ttl` when possible.
3. **No rate limiting.** Security relies on the random token being hard to guess (72 bits for automatic tokens). A short or easily guessable custom token (`--token`) weakens this protection.
4. **One shared file per session.** Upload mode accepts multiple files, but only **one file** can be shared from the laptop. **Directories are not supported** as shared files.
5. **IPv4 only.**
6. **Large uploads are written twice:** to the system's temporary folder, then to the destination folder. Ensure the temporary drive has enough space for the uploaded file, or lower `--max-size`.
7. **Windows device names** (e.g., `NUL.txt`, `CON.txt`) are not filtered: files with these names will not save properly. Names over 255 bytes are rejected by the system (HTTP 500).
8. **`--once` on downloads** waits for a **full** download. Clients downloading with multiple parallel Range requests might not trigger auto-shutdown; use `--ttl` as a safeguard.
9. **No code signing.** Windows (SmartScreen) and macOS (Gatekeeper) will show warnings for downloaded files. See the [Installation](#installation) section.
10. This project has **not been tested on physical hardware** for all combinations (see [Verification status](#verification-status)). Use reasonably and at your own risk.

---

## Downloads

Get the appropriate file from the **Releases** page: `https://github.com/argnta/goqrshare/releases`

| System | File |
|---|---|
| Windows 64-bit | `goqrshare-windows-amd64.exe` |
| Mac Apple Silicon (M1/M2/M3/M4) | `goqrshare-darwin-arm64` |
| Mac Intel | `goqrshare-darwin-amd64` |
| Linux 64-bit | `goqrshare-linux-amd64` |
| Verification | `checksums.txt` (SHA-256) |

**Don't know your Mac chip?** Open Terminal and run `uname -m`:

- `arm64` → Apple Silicon → `goqrshare-darwin-arm64`
- `x86_64` → Intel → `goqrshare-darwin-amd64`

### Verifying file integrity (recommended)

Compare the SHA-256 hash of the downloaded file with the contents of `checksums.txt`:

```powershell
# Windows (PowerShell)
Get-FileHash .\goqrshare-windows-amd64.exe -Algorithm SHA256
```

```bash
# macOS
shasum -a 256 goqrshare-darwin-arm64

# Linux
sha256sum goqrshare-linux-amd64
```

The hash value must **match exactly** with the line for that file in `checksums.txt`.

---

## Installation

GoQRShare is a **standalone single file**. No installer: download, grant permissions, and run.

### Windows

1. Download `goqrshare-windows-amd64.exe` to a folder, for example `C:\Tools\goqrshare\`.
2. Open **PowerShell** in that folder (In File Explorer: click the address bar, type `powershell`, press Enter).
3. Test: `.\goqrshare-windows-amd64.exe --version`
4. If **"Windows protected your PC"** (SmartScreen) appears: click **More info → Run anyway**.
5. On the first run, a **Windows Defender Firewall** dialog may appear: check **Private networks** and click **Allow access**. Without this, your phone cannot connect.
6. If an antivirus flags the file (false positive): verify the SHA-256 with `checksums.txt`; the source code is open, and you can build it yourself.

> Double-clicking the `.exe` **is not enough**: GoQRShare needs arguments (a file name or `--upload`). Run it from PowerShell or Command Prompt.

Optional: rename it to `goqrshare.exe` and add the folder to your `PATH` to call it as `goqrshare` from anywhere.

### macOS

1. Download the file matching your chip (`uname -m`, see the Downloads section).
2. Open Terminal in the download folder, grant execution permission, and remove the quarantine flag:

```bash
chmod +x goqrshare-darwin-arm64
xattr -d com.apple.quarantine goqrshare-darwin-arm64
./goqrshare-darwin-arm64 --version
```

   (Replace `goqrshare-darwin-arm64` with `goqrshare-darwin-amd64` for Intel Macs.)

3. If macOS still refuses ("developer cannot be verified"): open **System Settings → Privacy & Security**, scroll down, and click **Open Anyway** after the first failed attempt.
4. When the **"Do you want the application to accept incoming network connections?"** dialog appears: select **Allow**.
5. If the downloaded file is a `.zip` archive and cannot be executed (`permission denied`): run `chmod +x filename`.

Optional: move it to `/usr/local/bin` to call it as `goqrshare`:

```bash
sudo mv goqrshare-darwin-arm64 /usr/local/bin/goqrshare
```

### Linux

```bash
chmod +x goqrshare-linux-amd64
./goqrshare-linux-amd64 --version
```

---

## Build from source

This section is for those who want to build it themselves. **Cross-compilation** is supported: you **do not need a MacBook** to build the macOS version, and you don't need Windows to build the `.exe`.

### Prerequisites

- **Go 1.22 or newer**. Check with `go version`. Download from <https://go.dev/dl/>.
- Internet access during the first build (to download the `github.com/mdp/qrterminal/v4` dependency).
- Project structure: `main.go` in the root and packages in `internal/` (`apperror`, `appserver`, `config`, `netutil`, `qr`). The module name used in the code is: **`goqrshare`**.

### Build for Windows (`.exe`)

**On Windows (PowerShell), from the project folder:**

```powershell
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o goqrshare-windows-amd64.exe .
Remove-Item Env:CGO_ENABLED
.\goqrshare-windows-amd64.exe --version
```

**On Windows (Command Prompt / cmd):**

```bat
set CGO_ENABLED=0
go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o goqrshare-windows-amd64.exe .
```

**From macOS or Linux (cross-compile to Windows):**

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o goqrshare-windows-amd64.exe .
```

### Build for macOS

**On a Mac (native build), from the project folder:**

```bash
# Apple Silicon (M1/M2/M3/M4)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o goqrshare-darwin-arm64 .

# Intel
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o goqrshare-darwin-amd64 .
```

**From Windows (PowerShell, cross-compile to macOS):**

```powershell
$env:CGO_ENABLED = "0"
$env:GOOS = "darwin"; $env:GOARCH = "arm64"
go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o goqrshare-darwin-arm64 .
$env:GOARCH = "amd64"
go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o goqrshare-darwin-amd64 .
Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
```

> After cross-compiling from Windows, the macOS file **does not carry execution permissions**. Mac users need to run `chmod +x filename` (already included in the macOS installation steps).

### Building all targets at once and creating `checksums.txt`

```bash
VERSION=1.0.0
mkdir -p dist
for t in windows/amd64 darwin/arm64 darwin/amd64 linux/amd64; do
  os=${t%/*}; arch=${t#*/}; ext=""; [ $os = windows ] && ext=".exe"
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o dist/goqrshare-$os-$arch$ext .
done
cd dist && sha256sum goqrshare-* > checksums.txt   # on Mac: shasum -a 256 goqrshare-* > checksums.txt
```

### Pre-build quality checks

```bash
gofmt -l .
go vet ./...
go build ./...
```

---

## Usage

Below, `goqrshare` refers to the program file for your system (e.g., `.\goqrshare-windows-amd64.exe` on Windows or `./goqrshare-darwin-arm64` on Mac). Ensure your laptop and phone **are connected to the same Wi-Fi**.

### A. Sharing a file from laptop to phone

1. Run:

   ```bash
   goqrshare file.zip
   ```

2. The terminal displays the file name, network used, URL, and QR code. Example output (your values will differ):

   ```text
   Sharing  : file.zip (10485760 bytes)
   Network  : Wi-Fi (192.168.1.23)

   Open on phone (same Wi-Fi):
     https://192.168.1.23:52114/Xy7-Qk2mPz9A/

   [QR code]

   Press Ctrl+C to stop. Connection is encrypted (HTTPS, self-signed certificate): your browser will show a warning, choose Proceed.
   ```

3. On your phone, **scan the QR code** with your camera or type the URL exactly as shown in the terminal.
   > **Certificate warning.** Your browser will display a warning because of the self-signed certificate. On Chrome Android: **Advanced → Proceed to ... (unsafe)**. On Safari iOS: **Show Details → visit this website**. Proceed only if the IP address and port in the browser match those in the terminal.

4. A page with a **Download** button opens (or the download starts immediately). Save the file.
5. Press **Ctrl+C** in the terminal to stop.

### B. Receiving files from phone to laptop

1. Run (the destination folder must already exist):

   ```bash
   goqrshare --upload --dir ./receive
   ```

   Without `--dir`, files are saved to the user's `Downloads` folder, or to the current folder if `Downloads` does not exist.

2. Scan the QR on your phone. The **"Send to laptop"** page opens.
3. Press **Choose file**, select one or multiple files, and press **Send**. A progress bar and result message will appear on the page.
4. Files appear in the destination folder. If a name already exists, GoQRShare appends `name (1).ext`, `name (2).ext`, etc. (existing files are never overwritten).

### C. Mixed mode (download and upload)

```bash
goqrshare --upload --dir ./receive file.zip
```

The page on the phone displays a **Download** button for `file.zip` as well as the upload form.

### D. Other command examples

```bash
goqrshare file.zip                                   # share file
goqrshare --once --ttl 10m file.zip                  # one-time use, expires in 10 minutes
goqrshare file.zip --no-qr --verbose                 # flags can go after the file name
goqrshare --upload --dir=./receive --verbose         # "=" syntax is also allowed (safest)
goqrshare --upload --max-size 200 --dir ./receive    # limit to 200 MB per upload
goqrshare --iface "Wi-Fi" file.zip                   # choose interface (Windows example)
goqrshare --iface en0 file.zip                       # choose interface (Mac example)
goqrshare --port 8080 --token MySecretToken file.zip # fixed port and token
```

### E. Stopping

- Press **Ctrl+C** (normal shutdown, exit code 0).
- Or automatically: `--ttl` expires, or one successful transfer if using `--once`.
- Final message: `Server stopped: <reason>`.

---

## Flag list

There are 11 flags. Prefixes `-` and `--` are equivalent. `goqrshare --help` (or `-h`) prints a summary.

| Flag | Requires value? | Default | Function |
|---|---|---|---|
| `--port N` | Yes | `0` | Server port. `0` = automatically chosen by the system. Valid range: 1-65535 |
| `--upload` | No | off | Enable receive-upload mode from phone (web page + upload endpoint) |
| `--dir FOLDER` | Yes | User `Downloads`, or `.` | Upload destination folder. **Only applies with `--upload`**; must exist and be writable |
| `--max-size MB` | Yes | `1024` | **Total** size limit for one upload transfer, in MB. Minimum `1`. Exceeding limit → HTTP 413 |
| `--iface NAME` | Yes | auto | Force a specific network interface (case-insensitive), e.g., `"Wi-Fi"`, `en0`, `wlan0` |
| `--once` | No | off | Stop the server after **one successful transfer** (one full download or one successful upload) |
| `--ttl DURATION` | Yes | `0` (unlimited) | Auto-expiry. **Requires a unit**: `30s`, `10m`, `1h`. Cannot be negative |
| `--token T` | Yes | random | Fixed token in URL. 8-64 characters from `A-Z a-z 0-9 - _`. Empty = 12-character random token |
| `--no-qr` | No | off | Print only the URL, without a QR code |
| `--verbose` | No | off | Show detailed access logs. Tokens in logs are masked as `***` |
| `--version` | No | off | Print version and exit |

Standalone arguments (without a flag before them) are considered **the shared file**. Only **one** is allowed.

---

## Flag rules and common mistakes

**Four rules:**

1. **Flags with values** (`--dir`, `--port`, `--ttl`, `--max-size`, `--iface`, `--token`): the value must come **immediately** after the flag. Do not insert other flags in between. Values are taken literally, even if they start with `-`.
2. **Flags without values** (`--upload`, `--once`, `--no-qr`, `--verbose`, `--version`): write them alone, without a separate `true` or `false`. To explicitly disable, use the `=` syntax, e.g., `--once=false`.
3. **Standalone words = the shared file.** The upload folder is only specified via `--dir`.
4. **Flag order is free**, either before or after the file name.

**Common mistakes:**

| Incorrect | Result | Correct |
|---|---|---|
| `goqrshare --upload --dir -verbose ./receive` | `-verbose` read as folder name → `upload folder not found or not a directory: -verbose` | `goqrshare --upload --dir ./receive --verbose` |
| `--once false` | `false` read as file name → "only one file" error | `--once` or `--once=false` |
| `--ttl 5` | `invalid value "5" for flag -ttl` (requires a unit) | `--ttl 5s`, `--ttl 10m`, `--ttl 1h` |
| `--token abc` | Token must be 8-64 characters | `--token abc12345` |
| `--dir` at the very end | `flag needs an argument: -dir` | `--dir ./receive` |
| `--dir ./receive` without `--upload` | `--dir only applies with --upload` | add `--upload` |
| Unquoted path with spaces | Read as two arguments | `--dir "D:\My Folder\receive"` |
| Folder passed as a file | `directories are not supported` | Folders only via `--dir` |
| No file and no `--upload` | `no file to share and upload mode is not active` | Provide a file, or add `--upload` |

---

## Security

What GoQRShare does:

- **Random token** in the URL path (`/<token>/`). Wrong paths or tokens return 404; wrong methods return 405.
- **Bound to a single LAN IP**, not all interfaces. A warning is printed if the chosen address is not private.
- **Security headers** on every response (`X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`, strict `Content-Security-Policy` without external sources).
- **Upload filename sanitization:** only base names are used (`../` paths are stripped), illegal characters are replaced with `_`, trailing dots/spaces are removed. Example: `../../etc/passwd` → `passwd`, `a:b.txt` → `a_b.txt`, `..` → `upload.bin`.
- **Uploaded files never overwrite existing files** (created exclusively; duplicate names get `(1)`, `(2)`, etc.) and are created with `600` permissions on Unix-like systems.
- **Total size limit** via `--max-size`; partial files are deleted if saving fails.
- **Token masking** in access logs when using `--verbose`.

### Encryption (HTTPS)

- All traffic (pages, downloads, uploads, and tokens in URLs) is sent over **TLS 1.2 or higher**.
- The certificate is generated **in memory** at program startup (ECDSA P-256, valid for 12 hours, containing the IP address of the network used). **No keys or certificates are written to disk**, and a new certificate is generated every time the program runs.
- Because it is a self-signed certificate, the phone's browser will show a warning like *"Your connection is not private"*. This is **normal for GoQRShare**: proceed only if the IP address and port in the browser **match those shown in the terminal**.
- **What is protected:** passive eavesdropping (other devices merely recording Wi-Fi traffic cannot read the files or tokens).
- **What is not protected:** an active attacker on the same network impersonating your laptop. Because you click "proceed" on the certificate warning, the browser cannot distinguish the real server from an impersonator. Practical mitigations: use a trusted network, `--once`, and `--ttl`.
- GoQRShare does not provide an option to revert to HTTP.

What GoQRShare **does not** do: user authentication, rate limiting, virus scanning of uploaded files. Treat files from your phone like any other downloaded file.

---

## Troubleshooting

| Symptom | Possible cause | Solution |
|---|---|---|
| `Error: no matching private IP` or `no interfaces with IPv4` | Laptop is not connected to a network, or only has a non-IPv4 address | Connect to Wi-Fi/LAN, then run again |
| `Error: more than one eligible interface: ...; specify with --iface` | Multiple active network adapters | Choose one from the list: `--iface "Wi-Fi"` (Windows) or `--iface en0` (Mac, usually Wi-Fi) |
| `port already in use by another program (try --port 0)` | The requested port is taken | Use `--port 0` or another port |
| Phone cannot open the URL | Different Wi-Fi networks; firewall blocked; "AP isolation" enabled on router; wrong interface | Ensure same Wi-Fi; allow **Private networks** in Windows Firewall / **Allow** on macOS; try phone/laptop hotspot; try `--iface` |
| Browser shows *"Your connection is not private"* / *"Not Secure"* | Self-signed certificate (normal) | Proceed after ensuring the IP and port match the terminal (see *Encryption (HTTPS)*) |
| Terminal shows `http: TLS handshake error from ...: remote error: tls: bad certificate` | Browser or OS rejected the self-signed certificate on the first attempt, before you clicked "proceed" | Normal; ignore if the page opens afterward |
| `Client sent an HTTP request to an HTTPS server.` | URL was typed with `http://` | Use `https://` exactly as shown in the terminal, or scan the QR |
| Page opens but downloads fail on phone | Some mobile browsers restrict downloads from sites with untrusted certificates | Try another browser; report the results so this section can be updated |
| QR not readable | Terminal too small or font does not support half-blocks | Maximize the terminal window; or use `--no-qr` and type the URL |
| QR characters broken on Windows | Console code page is not UTF-8 | Use Windows Terminal, or run `chcp 65001` before GoQRShare |
| Upload fails with "File too large or failed to process" (413) | Exceeded `--max-size`, or temporary drive is full | Increase `--max-size`, free up space on the temporary drive |
| `upload folder is not writable` | Folder permissions | Choose another folder or change permissions |
| `--once` doesn't stop the server on partial download | Intentional: only full downloads are counted | Complete the download, or use `--ttl` |

---

## Known limitations

Intentionally left unfixed to keep v1 simple and safe:

1. One file per share session; folders not supported.
2. IPv4 only; no rate limiting.
3. Large uploads written twice (temp folder then destination folder). Streaming version planned for v1.1.
4. Windows device names (`NUL`, `CON`, etc.) are not filtered.
5. `--once` with clients using multiple parallel Range requests might not auto-stop.
6. Error messages for flag typos are raw (e.g., `-verbose` read as a folder name); friendlier messages planned for v1.1.
7. No Windows code signing or Apple notarization, so system warnings will appear.
8. `go install` is not supported because the module name `goqrshare` is not a GitHub address.
9. **Certificate cannot be verified** (self-signed, regenerated on every run). The certificate fingerprint **is not printed** in the terminal in this version, so users have no easy way to match it.

---

## Verification status

Honest about what has and hasn't been checked:

| Item | Status |
|---|---|
| Core logic, flags, routes, upload limits, file naming, `--once`/`--ttl`/Ctrl+C | Tested via `curl` and race detector on Linux |
| HTTPS, self-signed certificate generation, TLS handshake, `--once` on downloads | Tested successfully via `curl` on Windows (v1.0.0) |
| Cross-compilation 4 targets via GitHub Actions (`CGO_ENABLED=0`) | Successful (v1.0.0 release) |
| Build with native `qrterminal/v4` and QR readable by phone camera | Successful (v1.0.0 release) |
| Web page on Chrome Android and Safari iOS (1 file, many files, >100 MB) | Successful (v1.0.0 release) |
| macOS warnings and permissions (quarantine, Open Anyway, network dialog) | **Not yet verified** on actual Mac hardware |
| Windows Firewall and SmartScreen dialogs | Successful (v1.0.0 release) |
| `.exe` and macOS binaries | **Built and available** via GitHub Releases |

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
