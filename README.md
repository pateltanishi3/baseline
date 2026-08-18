# Amit Patel Dental Software

## Master baseline

`Dental_Patient_Management_Software.html` is the approved master HTML baseline. It is embedded directly into the Windows application, so the application UI and existing functionality remain based on this file.

## Windows desktop build

`main.go` is a small Windows desktop wrapper around the master HTML. It provides the existing `/api/db` interface through a localhost service and launches the application in Edge/Chrome app mode, with no normal browser tabs or address bar when Edge or Chrome is installed.

Patient data is stored locally in the Windows user configuration directory under `Amit Patel Dental Software\patient-data.json`.

## Professional Windows installer

GitHub Actions builds a Windows installer using Go and Inno Setup.

- Push to `main` or manually run the workflow to build a Windows installer artifact.
- Push a version tag such as `v1.0.0` to automatically create a GitHub Release and attach the installer.
- The installer creates Desktop and Start Menu shortcuts.
- The installer can launch the software immediately after installation.

## Versioning

Use semantic version tags (`v1.0.0`, `v1.1.0`, `v2.0.0`) for public releases. Future feature changes should start from the approved master baseline and preserve existing functionality unless explicitly requested otherwise.
