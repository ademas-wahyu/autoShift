# AI Constraints & Alur Kerja

## Constraints AI

| Constraint | Penjelasan |
|------------|------------|
| **Jeda Istirahat** | Minimal 12 jam antar shift berurutan |
| **Role Composition** | Setiap shift terisi role yang dibutuhkan (misal: ≥1 Supervisor) |
| **No Double-Booking** | 1 karyawan = 1 shift per hari |
| **Leave Quota** | Kuota libur per karyawan/bulan tidak terlampaui |
| **Fairness** | Distribusi shift merata (diukur via std dev) |

## Alur Kerja

```
Admin pilih bulan & konfigurasi → AI Generate (batch) → Validasi
    → Draf (Pending Review) → Review & Drag-Drop → Publish → Export / Share
```

- **Batch Processing**: Generate per 20 karyawan per batch, digabung & divalidasi backend
- **Validation Loop**: Validasi jeda 12h, role, double-booking, kuota libur — retry max 3x
- **Holiday API**: [date.nager.at](https://date.nager.at) — gratis, tanpa auth

---

- [Panduan instalasi & konfigurasi](quickstart.md)
- [Detail desktop app & sidecar](desktop.md)
