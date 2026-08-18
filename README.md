# Amit Patel Dental Software

## Master baseline

`Dental_Patient_Management_Software.html` is the approved master HTML baseline. It remains the source of the existing patient-management functionality.

## Windows desktop build

`main.go` packages the master HTML inside a native Windows WebView2 desktop window. The application has its own branded Windows icon and does not depend on Chrome or Edge as the visible application shell.

Patient data is stored locally in the Windows user configuration directory under `Amit Patel Dental Software\patient-data.json`.

## Doctor / clinic branding

The Settings screen provides a **Doctor & Clinic Profile** section. The doctor name and dental clinic/practice name are stored with the local database and are used in the dashboard greeting, WhatsApp messages, prescription WhatsApp messages, receipt branding, and other clinic identity areas. The professional dental logo is built in by default, while the existing custom-logo option remains available.

## Professional Windows installer

GitHub Actions builds a Windows installer using Go, WebView2, Windows resources, and Inno Setup.

- Push to `main` or manually run the workflow to build a Windows installer artifact.
- Push a version tag such as `v1.0.0` to automatically create a GitHub Release and attach the installer.
- The installer creates Desktop and Start Menu shortcuts using the branded application icon.
- The installer can launch the software immediately after installation.
- The workflow performs an installation and application-launch smoke test before publishing the installer artifact.

## Versioning

Use semantic version tags (`v1.0.0`, `v1.1.0`, `v2.0.0`) for public releases. Future feature changes should start from the approved master baseline and preserve existing functionality unless explicitly requested otherwise.

Build pipeline validation: 2026-08-18
