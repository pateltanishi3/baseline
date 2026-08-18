# Amit Patel Dental Software — Master Baseline

The approved master baseline is the uploaded dental patient-management HTML from this ChatGPT project. A Windows desktop executable has been built from that exact baseline.

## Local Windows build

The executable runs a small localhost database service and opens the baseline in an app-style Edge/Chrome window. Patient data is stored locally in the Windows user configuration directory under `Amit Patel Dental Software/patient-data.json`.

## Source

`main.go` is the Windows desktop launcher/database wrapper. The master `index.html` is preserved in the downloadable project package produced with the build.

Future software changes must start from the approved master baseline and preserve existing functionality unless explicitly requested otherwise.
