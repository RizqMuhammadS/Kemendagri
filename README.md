# Meeting Minutes AI - Sistem Notulensi Otomatis

Sistem notulensi rapat otomatis berbasis AI yang mengubah rekaman audio rapat menjadi notulensi terstruktur lengkap dengan ringkasan, poin pembahasan, keputusan, dan action items.

## Alur Sistem

```
Peserta Rapat
      │
      ▼
Perekaman Audio (Microphone/Meeting)
      │
      ▼
Speech-to-Text (Whisper/Google/Azure)
      │
      ▼
Pembersihan Teks (Hapus kata pengisi seperti "eh", "anu", dll.)
      │
      ▼
AI Summarization (LLM - OpenAI/Gemini/Claude)
      │
      ├── Ringkasan Rapat
      ├── Poin Pembahasan
      ├── Keputusan
      ├── Action Items + Penanggung Jawab + Deadline
      │
      ▼
Dashboard Notulensi
      │
      ├── Export PDF
      ├── Export Word
      ├── Email
      └── Riwayat Rapat
```

## Fitur

- **Speech-to-Text**: Dukungan OpenAI Whisper, Google STT, Azure STT
- **Text Cleaning**: Otomatis membersihkan kata pengisi (filler words), pengulangan kata, dan tanda baca berlebih
- **AI Summarization**: Meringkas rapat menggunakan LLM (OpenAI, Gemini, Claude, atau LLM lokal)
- **Structured Output**: Ringkasan, poin pembahasan, keputusan, action items dengan penanggung jawab dan deadline
- **Export**: PDF dan Word (DOCX)
- **Email**: Kirim notulensi via email sebagai lampiran
- **Dashboard**: Statistik rapat, action items, dan progress
- **Autentikasi**: JWT-based authentication dengan role user/admin
- **REST API**: Full RESTful API dengan dokumentasi Swagger

## Tech Stack

- **Backend**: Go (Golang) + Gin Framework
- **Database**: In-Memory (development) / PostgreSQL/MySQL/SQLite (production via GORM)
- **AI/ML**: OpenAI Whisper API, OpenAI/Gemini/Claude API
- **Authentication**: JWT (golang-jwt)
- **Export**: go-pdf (FPDF), go-docx
- **Dokumentasi**: Swagger/OpenAPI

## Struktur Project

```
meeting-minutes-ai/
├── cmd/
│   └── main.go                 # Entry point aplikasi
├── internal/
│   ├── config/
│   │   └── config.go           # Konfigurasi aplikasi
│   ├── controllers/
│   │   ├── auth_controller.go  # Handler autentikasi
│   │   └── meeting_controller.go # Handler meeting
│   ├── services/
│   │   ├── stt_service.go      # Speech-to-Text service
│   │   ├── text_cleaner.go     # Pembersihan teks
│   │   ├── llm_service.go      # AI Summarization service
│   │   ├── meeting_service.go  # Orchestrator bisnis logic
│   │   ├── export_service.go   # Export PDF/Word
│   │   ├── email_service.go    # Kirim email
│   │   └── auth_service.go     # Autentikasi & JWT
│   ├── repositories/
│   │   ├── user_repository.go  # Repository user
│   │   └── meeting_repository.go # Repository meeting
│   ├── models/
│   │   ├── meeting.go          # Model meeting, participant, dll
│   │   └── user.go             # Model user
│   ├── middleware/
│   │   └── auth.go             # Middleware JWT
│   ├── dto/
│   │   ├── request.go          # Request DTOs
│   │   └── response.go         # Response DTOs
│   ├── utils/
│   │   └── response.go         # Helper response
│   └── routes/
│       └── routes.go           # Route definitions
├── uploads/                    # Folder upload audio
├── exports/                    # Folder hasil export
├── docs/                       # Dokumentasi API (Swagger)
├── .env                        # Environment variables
├── go.mod
└── README.md
```

## Instalasi

### Prerequisites

- Go 1.21+
- API Key untuk layanan yang digunakan (OpenAI, dll)

### Langkah Instalasi

1. Clone repository:
```bash
git clone https://github.com/yourusername/meeting-minutes-ai.git
cd meeting-minutes-ai
```

2. Copy dan konfigurasi .env:
```bash
cp .env.example .env
# Edit .env sesuai kebutuhan
```

3. Install dependencies:
```bash
go mod tidy
```

4. Jalankan aplikasi:
```bash
go run cmd/main.go
```

Server akan berjalan di `http://localhost:8080`

## Konfigurasi (.env)

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `SERVER_PORT` | Port server | `8080` |
| `JWT_SECRET` | Secret key JWT | `your-secret-key` |
| `JWT_EXPIRATION` | Durasi token JWT | `24h` |
| `STT_ENGINE` | Engine STT (whisper/google/azure) | `whisper` |
| `LLM_API_KEY` | API Key LLM | - |
| `LLM_API_URL` | URL API LLM | `https://api.openai.com/v1/chat/completions` |
| `LLM_MODEL` | Model LLM | `gpt-3.5-turbo` |
| `UPLOAD_DIR` | Direktori upload | `./uploads` |
| `EXPORT_DIR` | Direktori export | `./exports` |
| `SMTP_HOST` | Host SMTP | - |
| `SMTP_PORT` | Port SMTP | `587` |
| `SMTP_USER` | User SMTP | - |
| `SMTP_PASS` | Password SMTP | - |
| `SMTP_FROM` | Email pengirim | - |

## API Endpoints

### Autentikasi (Public)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/auth/register` | Registrasi user baru |
| POST | `/api/auth/login` | Login user |

### Meeting (Protected - Perlu Bearer Token)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/meetings` | Buat meeting baru |
| GET | `/api/meetings` | List semua meeting |
| GET | `/api/meetings/:id` | Detail meeting |
| POST | `/api/meetings/upload-audio` | Upload audio meeting |
| POST | `/api/meetings/process-transcript` | Proses transkrip notulensi |
| POST | `/api/meetings/export` | Export notulensi (PDF/Word) |
| POST | `/api/meetings/send-email` | Kirim notulensi via email |

### Dashboard (Protected)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/dashboard` | Statistik dashboard |

### Health Check

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/health` | Cek status server |

## Contoh Penggunaan API

### Register User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "role": "user"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Buat Meeting
```bash
curl -X POST http://localhost:8080/api/meetings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "title": "Rapat Evaluasi Proyek",
    "date": "2024-03-15",
    "location": "Ruang Rapat Utama",
    "participants": [
      {"name": "Alice", "email": "alice@example.com", "role": "host"},
      {"name": "Bob", "email": "bob@example.com", "role": "speaker"}
    ]
  }'
```

### Proses Transkrip
```bash
curl -X POST http://localhost:8080/api/meetings/process-transcript \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "meeting_id": 1,
    "transcript": "Selamat pagi semua... hari ini kita akan membahas..."
  }'
```

### Export Notulensi
```bash
curl -X POST http://localhost:8080/api/meetings/export \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "meeting_id": 1,
    "format": "pdf"
  }'
```

## Pengembangan

### Menambahkan STT Engine Baru

1. Tambahkan case baru di `stt_service.go` method `Transcribe()`
2. Implementasi method transcribe baru
3. Tambahkan konfigurasi di `config.go`

### Menambahkan LLM Provider Baru

1. Tambahkan konfigurasi endpoint/model di `.env`
2. Sesuaikan `callLLM()` di `llm_service.go` untuk provider baru

### Menggunakan Database Relational

Untuk production, ganti in-memory repository dengan GORM:

```go
import "gorm.io/driver/postgres"

dsn := "host=localhost user=postgres password=pass dbname=meeting_minutes port=5432"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// Auto migrate
db.AutoMigrate(&models.User{}, &models.Meeting{}, &models.Participant{},
    &models.DiscussionPoint{}, &models.Decision{}, &models.ActionItem{})

// Gunakan GORM repository
userRepo := repositories.NewUserRepository(db)
meetingRepo := repositories.NewMeetingRepository(db)
```

## Lisensi

MIT License