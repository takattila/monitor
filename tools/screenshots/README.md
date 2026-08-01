# Screenshots

Headless Chrome screenshot generator for the monitor dashboard. It captures the
dashboard for every available skin (in dark and light mode) plus a few special
shots (default theme and the Network Traffic section) from the locally running
monitor web service.

The generated PNGs are meant to be copied into the repository's `assets/`
folder, which the README references.

## Requirements

- The monitor stack must be running locally (web on port `8383`, api on `7070`).
- A Chrome/Chromium binary. Default: `/opt/google/chrome/chrome`.
- Node.js with npm.

## Install

```sh
cd tools/screenshots
npm install
```

## Usage

```sh
node capture.js [options]
```

### Options

| Option        | Description                                                    |
|---------------|----------------------------------------------------------------|
| `--themes`    | Capture `desktop-<skin>-<mode>.png` for every skin (default).  |
| `--defaults`  | Capture `desktop-dark.png`, `desktop-light.png` and `desktop-full-light.png`. |
| `--network`   | Capture `network.png`: the expanded Network Traffic section with generated loopback traffic. |
| `--skin <n>`  | Limit the theme capture to a single skin (with `--themes`).    |
| `--all`       | Shortcut for `--themes --defaults --network`.                  |
| `-h` / `--help` | Show help.                                                   |

### Examples

```sh
# All theme screenshots (dark + light for all 33 skins)
node capture.js --themes

# One skin, for quick iteration
node capture.js --themes --skin mint

# Everything
node capture.js --all
```

## Environment variables

| Variable         | Default                   | Description                         |
|------------------|---------------------------|-------------------------------------|
| `MONITOR_BASE`   | `http://127.0.0.1:8383`   | Web service base URL.               |
| `MONITOR_OUT`    | `./out` (this directory)  | Output directory for the PNGs.      |
| `MONITOR_USER`   | `admin`                   | Login user name.                    |
| `MONITOR_PASS`   | `admin`                   | Login password.                     |
| `MONITOR_CHROME` | `/opt/google/chrome/chrome` | Chrome/Chromium executable.       |

## Deploying to assets/

```sh
node capture.js --all
cp out/*.png ../../assets/
```

The assets are referenced from `../../README.md`.

## Notes

- The network capture generates loopback traffic with `curl` so the charts show
  real movement; `curl` must be available for that shot to look meaningful.
- The script logs in with the default `admin`/`admin` credentials (see
  `configs/auth.db` in the repository root). Override with
  `MONITOR_USER`/`MONITOR_PASS` if your instance uses different credentials.
