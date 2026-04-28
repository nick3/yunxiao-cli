<!-- Parent: ../AGENTS.md -->

# scripts

## Purpose
Installer scripts for Yunxiao CLI distribution.

## For AI Agents
- Keep installers fail-closed: unsupported OS/arch, missing checksum, checksum mismatch, unsafe uninstall, or unwritable target must exit non-zero.
- Preserve install directory priority: explicit flag, then `YUNXIAO_INSTALL_DIR`, then user-writable default.
- Do not add preview release support to public installers; stable releases only.
- Uninstall must remove only the resolved `yunxiao` / `yunxiao.exe` binary and must not touch config, tokens, PATH entries, shell profiles, caches, logs, or remote data.
- Keep POSIX and PowerShell parameter behavior equivalent.

## Testing Requirements
- Run `sh -n scripts/install.sh` after editing the POSIX installer.
- Run the PowerShell parse check when `pwsh` is available after editing `install.ps1`.
- Run `go test ./test/installer -count=1` after installer behavior changes.
