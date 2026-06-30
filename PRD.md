# PRD: autoShift

## 1. Ringkasan

autoShift adalah web app untuk HR / scheduler dalam mengatur shift karyawan secara bulanan. Sistem menggunakan AI untuk meng-generate jadwal shift yang **adil dan seimbang** dengan tetap memberikan kendali penuh ke admin (human-in-the-loop).

Pain point utama: **ketidak-adilan distribusi shift** — misal karyawan masuk 3 hari, libur 1 hari, lalu masuk lagi 3 hari (mepet dan tidak seimbang).

---

## 2. Target User

| Role | Kebutuhan |
|---|---|
| HR / Scheduler | Mengatur shift karyawan 20-200+ orang secara adil |
| Admin / Manajer | Review & override hasil generate AI |
| Karyawan (MVP) | Cukup dapat jadwal via Export PDF/Excel atau Public Share Link (read-only) |

---

## 3. Alur Utama

1. Admin pilih **bulan** yang akan dibuat jadwal shift-nya (per bulan kalender)
2. Admin set **konfigurasi shift** (jumlah shift, jenis shift, jam kerja, dll)
3. Admin set **mode libur** (tetap / random) dan detailnya
4. Admin set **role requirement** per shift (misal: minimal 1 supervisor)
5. AI **meng-generate** jadwal shift berdasarkan semua constraints
6. Hasil AI → **Draf** (Status: *Pending Review*)
7. Admin **review** hasil generate
8. Jika ada yang **ngawur** → admin **geser manual** (drag-drop)
9. Jika OK → **Publish** → jadwal final

---

## 4. Fitur Detail

### 4.1 Manajemen Shift

- Konfigurasi jumlah shift per hari (misal: 3 shift: Pagi, Siang, Malam)
- Konfigurasi nama shift dan jam kerja masing-masing
- Jumlah karyawan per shift (bisa berbeda tiap shift)
- **Setiap shift bisa punya komposisi role wajib** (lihat 4.7)

### 4.2 Mode Libur

#### Mode Tetap (Fixed)
- Hari tertentu libur tiap minggu (misal: Minggu)
- Bisa lebih dari 1 hari (misal: Sabtu & Minggu)
- **Tanggal merah** mempengaruhi — otomatis libur
- AI mendistribusikan shift secara merata di hari kerja yang tersisa

#### Mode Random
- Jumlah hari libur per minggu (misal: 2 hari acak)
- Tanggal merah **tidak mempengaruhi** — libur random tetap dihitung terpisah
- AI mendistribusikan hari libur secara acak namun **seimbang** (tidak mepet-mepet)

### 4.3 AI Auto-Generator

Input AI:
- Jumlah & jenis shift + jam kerja per shift
- Mode libur (tetap/random + detail)
- Daftar karyawan, role, & kuota libur masing-masing
- Tanggal merah (dari public holiday API) — hanya untuk mode tetap
- **Minimum jeda istirahat antar shift** (default: 12 jam)
- **Komposisi role wajib** per shift (misal: tiap shift minimal 1 supervisor)

Output AI:
- Jadwal shift per karyawan untuk durasi yang diminta

Constraint AI:
- **Jeda Istirahat**: Minimal 12 jam antar shift — misal shift Malam (00:00-08:00) dan Pagi (08:00-16:00) di hari berurutan TIDAK valid karena melanggar jeda
- **Pemerataan**: Hindari pola tidak seimbang (clustering)
- **Komposisi Role**: Setiap shift terisi role yang dibutuhkan (misal: minimal 1 supervisor)
- **Ketersediaan**: Sesuai jumlah shift & kapasitas per shift
- **Kuota libur**: Maksimal hari libur per karyawan tidak terlampaui
- **Fairness**: Beban shift merata antar karyawan

### 4.4 Human-in-the-Loop (Manual Override)

- Admin bisa menggeser jadwal shift karyawan secara **drag-drop** di kalender
- Jika AI menghasilkan jadwal yang **ngawur**, admin bisa edit manual
- Edit per hari, per karyawan, atau per shift
- Validasi: sistem tetap periksa apakah hasil edit menyebabkan bentrok/kegagalan constraint

### 4.5 Kalender Libur Nasional (Public Holiday API)

- **Tidak pakai Google Calendar API** — ribet OAuth untuk self-hosted
- Gunakan **public API** seperti date.nager.at atau API libur nasional Indonesia yang tersedia di GitHub
- Cukup tembak endpoint, tanpa setup credentials
- Update otomatis tiap tahun
- Tampilkan visual tanggal merah di kalender

### 4.6 Kuota Libur

- Set kuota libur per karyawan (jumlah hari libur dalam 1 bulan)
- Untuk mode random: kuota menentukan berapa hari libur acak
- Untuk mode tetap: kuota tidak relevan (libur sudah fix di hari tertentu)
- AI memastikan kuota tidak terlampaui

### 4.7 Role & Komposisi Shift

- Profil karyawan punya **role** (misal: Supervisor, Staff, Intern)
- Admin bisa set **role requirement** per shift (misal: shift Pagi minimal 1 Supervisor dan 2 Staff)
- AI memastikan komposisi role terpenuhi di setiap shift
- Jika tidak cukup Supervisor untuk semua shift → sistem beri warning ke admin

### 4.8 Skala & Batch Processing

- Generate 200 karyawan × 30 hari = 6.000 titik data — terlalu besar untuk sekali kirim ke AI
- **Backend melakukan batch processing**:
  - Per minggu, atau
  - Per 20-30 karyawan per batch
- Backend merangkai hasil batch sebelum disimpan ke database & ditampilkan ke admin

### 4.9 Export & Distribusi

- **Export to PDF** — jadwal siap cetak
- **Export to Excel/CSV** — editable di spreadsheet
- **Public Share Link (read-only)** — admin dapat link dibagikan ke grup WhatsApp, tanpa login karyawan

---

## 5. Non-Functional Requirements

| Aspect | Detail |
|---|---|
| Platform | Web app (responsive) |
| AI | Constraint-based scheduling + fairness optimization |
| AI Skalabilitas | Batch processing per minggu / per 20-30 karyawan |
| Data | Tersimpan di database (relational) |
| Kalender | Public holiday API (bukan Google Calendar) |
| Bahasa | Indonesia (default) |
| Akses | Multi-tenant (setiap perusahaan punya data sendiri) |

---

## 6. User Stories

### S1: Konfigurasi Shift
> Sebagai admin, saya ingin mengatur jumlah shift (misal 3 shift: Pagi/Siang/Malam) beserta jam kerja masing-masing.

### S2: Mode Libur Tetap
> Sebagai admin, saya ingin menetapkan hari libur tetap (misal setiap Minggu) dan sistem mendeteksi tanggal merah dari API libur nasional.

### S3: Mode Libur Random
> Sebagai admin, saya ingin menetapkan jumlah hari libur acak per minggu (misal 2 hari) dan sistem mendistribusikannya secara seimbang tanpa terpengaruh tanggal merah.

### S4: Auto-Generate
> Sebagai admin, saya ingin AI meng-generate jadwal shift dengan distribusi yang adil.

### S5: Jeda Istirahat
> Sebagai admin, saya ingin sistem memastikan tidak ada karyawan yang dapat shift berurutan tanpa jeda minimal 12 jam (misal shift Malam lalu Pagi di hari berikutnya).

### S6: Komposisi Role
> Sebagai admin, saya ingin memastikan setiap shift memiliki minimal 1 Supervisor, dan AI tidak menempatkan staf saja tanpa Supervisor.

### S7: Manual Override
> Sebagai admin, saya ingin menggeser jadwal karyawan secara drag-drop jika hasil generate AI tidak sesuai.

### S8: Kuota Libur
> Sebagai admin, saya ingin menetapkan kuota libur per karyawan dan sistem memastikan tidak terlampaui.

### S9: Publish & Export
> Sebagai admin, setelah jadwal final saya ingin publish dan membagikannya ke karyawan via export PDF/Excel atau link publik.

---

## 7. Alur Decision Logic (Libur)

```
Apakah mode libur tetap?
  ├── Ya → Hari tertentu libur tiap minggu
  │        → Tanggal merah ikut libur
  │        → AI sebar shift di hari kerja
  │
  └── Tidak (random) → Admin set jumlah hari libur per minggu
                       → Tanggal merah TIDAK jadi libur
                       → AI distribusi X hari libur secara acak & seimbang
```

---

## 8. Flow Approval (MVP)

```
AI Generate
    │
    ▼
Draf (Status: Pending Review)
    │
    ▼
Manajer Review & Edit (drag-drop jika perlu)
    │
    ▼
Tekan "Publish"
    │
    ▼
Jadwal Final → Export / Share Link
```

---

## 9. Cara Mengukur "Keadilan" (Kuantitatif)

| Metrik | Cara Hitung |
|---|---|
| Standar deviasi total jam kerja | Hitung total jam kerja tiap karyawan per bulan → SD. Makin mendekati 0, makin adil |
| Standar deviasi shift akhir pekan | Hitung jumlah shift Sabtu/Minggu tiap karyawan → SD. Idealnya semua sama |
| Distribusi shift pagi/siang/malam | Pastikan rotasi merata, tidak ada yang terus-terusan dapat shift malam |

---

## 10. Batasan (Scope) — MVP V1

- **Tidak** ada fitur pengajuan cuti karyawan (di luar kuota libur)
- **Tidak** ada kalkulasi gaji / lembur
- **Tidak** ada real-time clock-in / absensi
- **Tidak** ada swap shift antar karyawan (cukup admin geser manual)
- **Tidak** ada login karyawan (cukup export / share link)
- Fokus: **generate shift yang adil + human override + export**
