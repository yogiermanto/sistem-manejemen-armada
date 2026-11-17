# Sistem Manajemen Armada

Backend real-time untuk tracking lokasi kendaraan: ingest lokasi via **MQTT**, simpan ke **PostgreSQL**, sajikan via **REST API (Gin)**, deteksi **geofence** dinamis (Haversine; titik & radius dibuat via API, default 50 m), dan kirim event ke **RabbitMQ** untuk diproses worker. Semua jalan dengan satu perintah **Docker Compose**.

🎥 **Video demo:** [Google Drive](https://drive.google.com/file/d/1JEJTAzxOQXh86mPXJEl8jacGxB5n1Cy6/view?usp=sharing)

## Arsitektur

```
 simulator ──MQTT QoS1──▶ Mosquitto ──▶ api (Go)
 (tiap 2 dtk)            /fleet/vehicle/{id}/location
                                          │ 1. validasi payload
                                          │ 2. diproses satu per satu (urutan terjaga)
                                          │ 3. INSERT vehicle_locations (idempotent)
                                          │ 4. Haversine vs semua geofence → entry/exit?
                                          ▼
                     PostgreSQL ◀── REST API :8080      RabbitMQ
                         ▲          /vehicles/...       exchange fleet.events (topic)
                         │          /geofences
                         │                               │ geofence.entry / geofence.exit
                         │                               ▼
                         └──────── worker ◀──── queue geofence_alerts ──nack──▶ geofence_alerts.dlq
                          INSERT geofence_events
```

| Service | Fungsi |
|---|---|
| `api` | Subscriber MQTT + deteksi geofence + publisher RabbitMQ + REST API (lokasi & geofence) |
| `worker` | Consumer `geofence_alerts`, simpan event ke `geofence_events` |
| `simulator` | Publish lokasi 3 kendaraan tiap 2 detik, bolak-balik lewat geofence default |
| `migrate` | [golang-migrate](https://github.com/golang-migrate/migrate): jalankan migrasi `db/migrations` lalu exit |
| `postgres`, `rabbitmq`, `mosquitto` | Infrastruktur |

## Database

PostgreSQL dengan 3 tabel. Semuanya dibuat oleh migrasi di [`db/migrations`](db/migrations) (golang-migrate), yang juga membuat tabel `schema_migrations` untuk mencatat versi.

| Tabel | Isi | Ditulis oleh | Dibaca oleh | Kunci & index |
|---|---|---|---|---|
| `vehicle_locations` | Semua ping lokasi kendaraan | `api` (dari MQTT) | `GET /vehicles/{vehicle_id}/location`, `GET /vehicles/{vehicle_id}/history` | PK `id`; UNIQUE `(vehicle_id, timestamp)` dipakai untuk lokasi terakhir, histori, dan menolak duplikat |
| `geofences` | Pusat + radius tiap area geofence | Migrasi (seed "Default", 50 m), `POST /geofences` | `api` saat start (detector), `GET /geofences` | PK `id`; `CHECK (radius_m > 0)` |
| `geofence_events` | Riwayat kendaraan masuk/keluar geofence | `worker` (dari RabbitMQ) | Query manual / laporan | PK `id`; FK `geofence_id` → `geofences.id`; UNIQUE `(vehicle_id, geofence_id, event, timestamp)` supaya event tidak dobel |

Relasi: satu `geofences` punya banyak `geofence_events`. `vehicle_locations` berdiri sendiri karena tidak ada tabel master kendaraan; kendaraan "ada" begitu ia mengirim lokasi.

### `vehicle_locations`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | `UUID` PK | UUIDv7, dibuat aplikasi |
| `vehicle_id` | `VARCHAR(32)` | ID kendaraan dari payload MQTT |
| `latitude` | `DOUBLE PRECISION` | Lintang, -90 s/d 90 |
| `longitude` | `DOUBLE PRECISION` | Bujur, -180 s/d 180 |
| `timestamp` | `BIGINT` | Unix detik, waktu dari perangkat |
| `created_at` | `TIMESTAMPTZ`, default `now()` | Waktu server menerima |

### `geofences`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | `UUID` PK | UUIDv7 dari aplikasi (seed "Default": `gen_random_uuid()`) |
| `name` | `VARCHAR(100)` | Nama area |
| `latitude` | `DOUBLE PRECISION` | Lintang pusat lingkaran |
| `longitude` | `DOUBLE PRECISION` | Bujur pusat lingkaran |
| `radius_m` | `DOUBLE PRECISION`, `> 0` | Radius dalam meter (default 50 dari API) |
| `created_at` | `TIMESTAMPTZ`, default `now()` | Waktu dibuat |

### `geofence_events`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | `UUID` PK | UUIDv7, dibuat aplikasi |
| `vehicle_id` | `VARCHAR(32)` | Kendaraan yang masuk/keluar |
| `geofence_id` | `UUID` FK → `geofences.id` | Geofence yang dilewati |
| `event` | `VARCHAR(32)` | `geofence_entry` atau `geofence_exit` |
| `latitude` | `DOUBLE PRECISION` | Lintang ping yang memicu event |
| `longitude` | `DOUBLE PRECISION` | Bujur ping yang memicu event |
| `timestamp` | `BIGINT` | Unix detik dari ping pemicu |
| `received_at` | `TIMESTAMPTZ`, default `now()` | Waktu worker menyimpan |

## Menjalankan (step by step)

Ikuti urut dari langkah 1. Semua perintah dijalankan dari folder project.

### Langkah 1: Prasyarat

| Tool | Wajib? | Cek |
|---|---|---|
| Docker + Docker Compose v2 | wajib | `docker compose version` |
| `curl` | wajib (untuk uji API) | `curl --version` |
| Go 1.27.4 | hanya untuk `go test` / jalan tanpa Docker | `go version` |
| Node.js (`npx`) | hanya untuk uji Postman via CLI | `npx --version` |

Pastikan port **8080, 1883, 5432, 5672, 15672** di laptop belum dipakai aplikasi lain (mis. Postgres lokal). Kalau bentrok, ganti port host di `.env` (lihat [Port](#port)).

### Langkah 2: Ambil source code

```bash
git clone https://github.com/yogiermanto/sistem-manejemen-armada.git sistem-manajemen-armada
```

```bash
cd sistem-manajemen-armada
```

### Langkah 3: Buat `.env`

Semua konfigurasi (kredensial Postgres/RabbitMQ, port, simulator) dibaca dari `.env`:

```bash
cp .env.example .env
```

Nilai default sudah langsung jalan, tidak perlu diubah.

### Langkah 4: Nyalakan semua service

```bash
docker compose up -d --build
```

Build pertama ±1–2 menit. Yang dinyalakan:

| Service | Fungsi |
|---|---|
| `postgres` | Database |
| `mosquitto` | MQTT broker |
| `rabbitmq` | Message broker event geofence |
| `migrate` | Membuat tabel lalu berhenti (`Exited (0)` itu normal) |
| `api` | REST API + consumer MQTT + deteksi geofence |
| `worker` | Consumer RabbitMQ `geofence_alerts` |
| `simulator` | 3 kendaraan palsu, kirim lokasi tiap 2 detik |

### Langkah 5: Pastikan semua sehat

Tunggu ±15 detik, lalu:

```bash
docker compose ps -a
```

Harus terlihat seperti ini:

| Service | Status yang benar |
|---|---|
| `postgres`, `mosquitto`, `rabbitmq`, `api` | `Up ... (healthy)` |
| `worker`, `simulator` | `Up ...` |
| `migrate` | `Exited (0)` |

Kalau ada yang `Exited (1)` atau `Restarting`, lihat lognya, misalnya `docker compose logs api`.

### Langkah 6: Lihat aliran data (opsional)

Ikuti log semua service sekaligus (keluar dengan `Ctrl+C`, service tetap jalan):

```bash
docker compose logs -f api worker simulator
```

Yang terlihat: simulator `published`, api `geofence event published`, worker `geofence event processed`.

Setelah semua sehat, lanjut ke [Pengujian step by step](#pengujian-step-by-step).

### Port

| Port | Service |
|---|---|
| 8080 | REST API |
| 1883 | MQTT (Mosquitto) |
| 5432 | PostgreSQL (`DATABASE_USER` / `DATABASE_PASSWORD` di `.env`, default `armada` / `armada`) |
| 5672 / 15672 | RabbitMQ AMQP / Management UI (`RABBITMQ_USER` / `RABBITMQ_PASSWORD`, default `armada` / `armada`) → http://localhost:15672 |

Port host bisa diubah lewat `DATABASE_PORT`, `MQTT_PORT`, `RABBITMQ_PORT`, `RABBITMQ_MANAGEMENT_PORT` di `.env`.

### Menjalankan tanpa Docker (opsional)

Butuh Go 1.27.4. Nilai di `.env` sudah mengarah ke `localhost`, jadi cukup nyalakan infrastruktur saja lalu jalankan binary dari host (masing-masing di terminal terpisah):

```bash
docker compose up -d postgres rabbitmq mosquitto migrate
```

```bash
go run ./cmd/api
```

```bash
go run ./cmd/worker
```

```bash
go run ./cmd/simulator
```

## Pengujian step by step

Jalankan setelah [Langkah 5](#langkah-5-pastikan-semua-sehat) beres. Urutan uji mengikuti aliran data: MQTT → PostgreSQL → REST API → geofence → RabbitMQ.

### Uji 1: Simulator mengirim lokasi lewat MQTT

```bash
docker compose logs simulator --tail 3
```

Harus ada 3 baris `"msg":"published"` dengan topic `/fleet/vehicle/{id}/location`:

```
{"level":"INFO","msg":"published","topic":"/fleet/vehicle/B1234XYZ/location","latitude":-6.2104,"longitude":106.8456,"distance_from_center_m":180}
```

Kirim satu lokasi manual juga untuk membuktikan broker menerima dari luar simulator:

```bash
docker compose exec mosquitto mosquitto_pub -q 1 -t /fleet/vehicle/TEST001/location \
  -m "{\"vehicle_id\":\"TEST001\",\"latitude\":-6.2088,\"longitude\":106.8456,\"timestamp\":$(date +%s)}"
```

Tidak ada output = terkirim.

### Uji 2: Lokasi tersimpan di PostgreSQL

Jumlah baris per kendaraan (jalankan dua kali berjarak beberapa detik, angkanya harus naik):

```bash
docker compose exec postgres psql -U armada -c "SELECT vehicle_id, count(*), max(timestamp) FROM vehicle_locations GROUP BY vehicle_id ORDER BY vehicle_id;"
```

```
 vehicle_id | count |    max
------------+-------+------------
 B1234XYZ   |    42 | 1794881700
 B5678ABC   |    42 | 1794881700
 B9012DEF   |    42 | 1794881700
 TEST001    |     1 | 1794881650
```

`TEST001` dengan `count = 1` membuktikan pesan manual dari Uji 1 ikut tersimpan (kalau Uji 1 diulang, angkanya bertambah; itu normal).

### Uji 3: API lokasi terakhir

```bash
curl -s http://localhost:8080/vehicles/B1234XYZ/location
```

```json
{"data":{"vehicle_id":"B1234XYZ","latitude":-6.2104,"longitude":106.8456,"timestamp":1794881700}}
```

Panggil lagi beberapa detik kemudian: `timestamp` berubah. Kendaraan yang tidak ada → `404`:

```bash
curl -s -w ' HTTP %{http_code}\n' http://localhost:8080/vehicles/TIDAKADA/location
```

```
{"error":"not found: no location for vehicle TIDAKADA"} HTTP 404
```

### Uji 4: API histori + pagination

10 menit terakhir, 5 data per halaman:

```bash
curl -s "http://localhost:8080/vehicles/B1234XYZ/history?start=$(( $(date +%s) - 600 ))&end=$(date +%s)&limit=5"
```

Cek: `data` berisi 5 item urut naik berdasarkan `timestamp`, `meta` berisi `"page":1,"limit":5,"count":5,"has_more":true`.

Halaman berikutnya (`page=2`): data lanjut dari halaman 1, tidak ada yang dobel:

```bash
curl -s "http://localhost:8080/vehicles/B1234XYZ/history?start=$(( $(date +%s) - 600 ))&end=$(date +%s)&limit=5&page=2"
```

Parameter salah (`end` < `start`) → `400`:

```bash
curl -s -w ' HTTP %{http_code}\n' "http://localhost:8080/vehicles/B1234XYZ/history?start=200&end=100"
```

### Uji 5: Geofence → RabbitMQ → worker

Geofence "Default" (`-6.2088, 106.8456`, radius 50 m) sudah dibuat migrasi:

```bash
curl -s http://localhost:8080/geofences
```

**5a. Masuk geofence (entry).** `TEST001` dari Uji 1 dikirim tepat di titik Default, jadi sudah memicu `geofence_entry`. Cek di log worker (consumer queue `geofence_alerts`):

```bash
docker compose logs worker | grep TEST001
```

```
{"level":"INFO","msg":"geofence event processed","vehicle_id":"TEST001","event":"geofence_entry","latitude":-6.2088,"longitude":106.8456,...}
```

**5b. Keluar geofence (exit).** Tunggu ±2 detik (timestamp harus lebih baru), lalu kirim `TEST001` ±55 m di utara titik Default:

```bash
docker compose exec mosquitto mosquitto_pub -q 1 -t /fleet/vehicle/TEST001/location \
  -m "{\"vehicle_id\":\"TEST001\",\"latitude\":-6.2083,\"longitude\":106.8456,\"timestamp\":$(date +%s)}"
```

```bash
docker compose logs worker | grep TEST001
```

Sekarang ada baris kedua dengan `"event":"geofence_exit"`. Event hanya muncul saat **berpindah** status (luar → dalam = entry, dalam → luar = exit); ping yang tetap di dalam atau tetap di luar tidak memicu event.

**5c. Event tersimpan di database** (ditulis oleh worker):

```bash
docker compose exec postgres psql -U armada -c "SELECT e.vehicle_id, g.name, e.event, e.timestamp FROM geofence_events e JOIN geofences g ON g.id = e.geofence_id WHERE e.vehicle_id = 'TEST001' ORDER BY e.timestamp;"
```

```
 vehicle_id |  name   |     event      | timestamp
------------+---------+----------------+------------
 TEST001    | Default | geofence_entry | 1794881650
 TEST001    | Default | geofence_exit  | 1794881660
```

**5d. Topologi RabbitMQ.** Queue `geofence_alerts` ada dan punya consumer (worker), `messages` 0 karena langsung diproses; DLQ kosong:

```bash
docker compose exec rabbitmq rabbitmqctl list_queues name messages consumers
```

```
name                 messages  consumers
geofence_alerts.dlq  0         0
geofence_alerts      0         1
```

Bisa juga dilihat di UI http://localhost:15672 (login `armada` / `armada`): tab **Exchanges** → `fleet.events`, tab **Queues** → `geofence_alerts`.

**5e. Geofence baru dari API.** Buat titik 200 m di utara pusat simulator (rute `B1234XYZ` lewat situ):

```bash
curl -s -X POST http://localhost:8080/geofences -H 'Content-Type: application/json' \
  -d '{"name":"Pos Utara","latitude":-6.2070034,"longitude":106.8456,"radius_m":50}'
```

Respons `201` dengan `id` UUIDv7. Tunggu ±30 detik tanpa restart apa pun, lalu:

```bash
docker compose exec postgres psql -U armada -c "SELECT e.vehicle_id, g.name, e.event FROM geofence_events e JOIN geofences g ON g.id = e.geofence_id WHERE g.name = 'Pos Utara';"
```

Harus muncul `B1234XYZ | Pos Utara | geofence_entry` (dan `geofence_exit` setelah lewat).

### Uji 6: Payload tidak valid di-drop

```bash
docker compose exec mosquitto mosquitto_pub -q 1 -t /fleet/vehicle/TEST002/location -m 'bukan json'
```

```bash
docker compose exec mosquitto mosquitto_pub -q 1 -t /fleet/vehicle/TEST002/location \
  -m "{\"vehicle_id\":\"TEST002\",\"latitude\":999,\"longitude\":106.8456,\"timestamp\":$(date +%s)}"
```

```bash
docker compose exec mosquitto mosquitto_pub -q 1 -t /fleet/vehicle/TEST002/location \
  -m "{\"vehicle_id\":\"LAIN\",\"latitude\":-6.2088,\"longitude\":106.8456,\"timestamp\":$(date +%s)}"
```

Ketiganya tercatat sebagai `drop location` dengan alasan masing-masing:

```bash
docker compose logs api | grep 'drop location'
```

```
"reason":"invalid json"
"reason":"invalid location"
"reason":"vehicle_id does not match topic"
```

Dan tidak ada yang masuk database (hasil `0`):

```bash
docker compose exec postgres psql -U armada -tc "SELECT count(*) FROM vehicle_locations WHERE vehicle_id IN ('TEST002','LAIN');"
```

### Uji 7: Postman (semua endpoint + assertion)

Import `postman/sistem-manajemen-armada.postman_collection.json` ke Postman lalu **Run collection**, atau via CLI:

```bash
npx newman run postman/sistem-manajemen-armada.postman_collection.json
```

Hasil yang benar: `requests 11 / 0 failed`, `assertions 24 / 0 failed`.

### Uji 8: Unit test

Butuh Go di host, tidak butuh Docker:

```bash
go test ./...
```

Semua paket `ok` (Haversine, entry/exit, banyak geofence, usecase ingest, handler MQTT, config).

## Berhenti & reset

### Berhenti sementara (data tetap ada)

```bash
docker compose down
```

Container dihapus, tapi volume Postgres/RabbitMQ tetap. `docker compose up -d` berikutnya melanjutkan dengan data lama.

### Reset total (kembali seperti baru clone)

Hapus container, network, **semua data** (volume Postgres & RabbitMQ) dan image yang di-build project ini:

```bash
docker compose down -v --rmi local --remove-orphans
```

Pastikan bersih (keduanya tidak menampilkan apa-apa):

```bash
docker compose ps -a
```

```bash
docker volume ls --filter name=sistem-manajemen-armada
```

Opsional, hapus juga `.env` kalau mau mulai dari Langkah 3 lagi:

```bash
rm .env
```

Untuk mulai lagi dari nol, ulangi dari [Langkah 3](#langkah-3-buat-env). Migrasi akan membuat ulang tabel dan geofence Default.

## REST API

Format response: sukses `{"data": ...}` (+ `meta` untuk list), gagal `{"error": "..."}`.

### `GET /vehicles/{vehicle_id}/location`

Lokasi terakhir kendaraan.

```bash
curl http://localhost:8080/vehicles/B1234XYZ/location
```

```json
{ "data": { "vehicle_id": "B1234XYZ", "latitude": -6.2088, "longitude": 106.8456, "timestamp": 1715003456 } }
```

`404` jika kendaraan belum pernah mengirim lokasi.

### `GET /vehicles/{vehicle_id}/history?start=&end=&page=&limit=`

Histori lokasi dalam rentang waktu, urut naik berdasarkan `timestamp`.

| Param | Wajib | Keterangan |
|---|---|---|
| `start` | ya | unix seconds |
| `end` | ya | unix seconds, `>= start` |
| `page` | tidak | default `1` |
| `limit` | tidak | default `100`, maks `1000` |

```bash
curl "http://localhost:8080/vehicles/B1234XYZ/history?start=$(( $(date +%s) - 600 ))&end=$(date +%s)&limit=5"
```

```json
{
  "data": [ { "vehicle_id": "B1234XYZ", "latitude": -6.2105, "longitude": 106.8456, "timestamp": 1715003400 } ],
  "meta": { "page": 1, "limit": 5, "count": 5, "has_more": true }
}
```

`400` jika parameter tidak valid. `has_more` dihitung dengan mengambil `limit + 1` baris (tanpa `COUNT(*)`).

### `POST /geofences`

Buat titik geofence baru. Langsung aktif tanpa restart: ping berikutnya sudah dicek terhadap titik ini, dan event entry/exit-nya membawa `geofence_id`.

| Field | Wajib | Keterangan |
|---|---|---|
| `name` | ya | maks 100 karakter |
| `latitude` | ya | -90 … 90 |
| `longitude` | ya | -180 … 180 |
| `radius_m` | tidak | meter, > 0, maks 100000, default `50` |

```bash
curl -X POST http://localhost:8080/geofences -H 'Content-Type: application/json' \
  -d '{"name":"Pos Utara","latitude":-6.2070034,"longitude":106.8456,"radius_m":50}'
```

```json
{ "data": { "id": "019a8c2e-5b1f-7c3a-9d41-2f6e8b0a7c55", "name": "Pos Utara", "latitude": -6.2070034, "longitude": 106.8456, "radius_m": 50, "created_at": "2026-11-16T10:00:00Z" } }
```

`201` jika berhasil, `400` jika body tidak valid. `id` berupa UUIDv7. Geofence "Default" (`-6.2088, 106.8456`, 50 m) sudah dibuat oleh migrasi.

### `GET /geofences`

Daftar semua geofence.

```bash
curl http://localhost:8080/geofences
```

### `GET /health`

Dipakai healthcheck Docker.

## Postman

Import `postman/sistem-manajemen-armada.postman_collection.json`. Variabel collection hanya `baseUrl` (default `http://localhost:8080`). `vehicleId` diisi di tab **Params → Path Variables**, sedangkan `start`, `end`, `page`, `limit` di tab **Params → Query Params** (`end` default `{{$timestamp}}` = waktu sekarang). Tiap request punya test assertion. Bisa juga dijalankan via CLI:

```bash
npx newman run postman/sistem-manajemen-armada.postman_collection.json
```

## Migrasi database (golang-migrate)

Schema dikelola dengan [golang-migrate](https://github.com/golang-migrate/migrate). File ada di `db/migrations`, tiap perubahan punya pasangan `.up.sql` dan `.down.sql`. Service `migrate` di Compose menjalankan `up` setiap `docker compose up` (yang sudah diterapkan dilewati, versinya dicatat di tabel `schema_migrations`), dan `api`/`worker` baru start setelah migrasi sukses.

Cek versi schema:

```bash
docker compose run --rm migrate version
```

Rollback satu langkah, lalu terapkan lagi:

```bash
docker compose run --rm migrate down 1
```

```bash
docker compose run --rm migrate up
```

Buat migrasi baru (butuh CLI: `brew install golang-migrate`), lalu isi kedua file yang dibuat:

```bash
migrate create -ext sql -dir db/migrations -seq add_column_x
```

Jangan mengubah file migrasi yang sudah pernah dijalankan; buat migrasi baru.

## Konfigurasi (env)

Satu sumber: **`.env`** (template `.env.example`, di-gitignore).

- **Binary Go** membacanya dengan [viper](https://github.com/spf13/viper) (`internal/config`). Prioritas: environment variable > `.env` > default di kode.
- **Docker Compose** memakai `.env` untuk interpolasi `${VAR}` (kredensial container Postgres, RabbitMQ, URL migrate, port host) dan `env_file: .env` untuk service `api`/`worker`/`simulator`. Di dalam container hanya `*_HOST`, `*_PORT`, dan `HTTP_ADDR` yang di-override ke nama service & port internal (`postgres:5432`, `mosquitto:1883`, `rabbitmq:5672`).

Koneksi dikonfigurasi per bagian (host, port, user, password), bukan URL utuh, supaya override itu cukup mengganti host/port. Password di-escape otomatis saat URL dirakit. Nilai yang tidak valid (mis. `SIM_INTERVAL=abc`) membuat service gagal start dengan pesan jelas.

| Env | Default | Dipakai oleh |
|---|---|---|
| `HTTP_ADDR` | `:8080` | api |
| `DATABASE_HOST` / `DATABASE_PORT` | `localhost` / `5432` | api, worker (+ port host container postgres) |
| `DATABASE_USER` / `DATABASE_PASSWORD` / `DATABASE_NAME` | `armada` / `armada` / `armada` | api, worker, container postgres, migrate |
| `DATABASE_SSLMODE` | `disable` | api, worker, migrate |
| `MQTT_HOST` / `MQTT_PORT` | `localhost` / `1883` | api, simulator (+ port host mosquitto) |
| `MQTT_CLIENT_ID` | `armada-api` | api (harus tetap agar sesi persisten berfungsi) |
| `MQTT_TOPIC` | `/fleet/vehicle/+/location` | api |
| `RABBITMQ_HOST` / `RABBITMQ_PORT` | `localhost` / `5672` | api, worker (+ port host rabbitmq) |
| `RABBITMQ_MANAGEMENT_PORT` | `15672` | port host UI RabbitMQ |
| `RABBITMQ_USER` / `RABBITMQ_PASSWORD` | `armada` / `armada` | api, worker, container rabbitmq |
| `SIM_VEHICLES` | `B1234XYZ,B5678ABC,B9012DEF` | simulator |
| `SIM_INTERVAL` | `2s` | simulator (minimal `1s`) |
| `SIM_CENTER_LAT` / `SIM_CENTER_LON` | `-6.2088` / `106.8456` | simulator — pusat rute (= geofence default) |

Kredensial Postgres & RabbitMQ hanya dipakai saat volume pertama dibuat. Setelah mengubahnya, jalankan `docker compose down -v` (menghapus data).

## Struktur (clean architecture)

```
cmd/
  api/main.go            entrypoint backend: config → infra → config.Bootstrap
  worker/main.go         entrypoint consumer RabbitMQ
  simulator/main.go      simulator kendaraan (tool, bukan bagian service)
internal/
  config/                konstruksi infrastruktur + wiring
    app.go                 Bootstrap: repository → usecase → controller/route + MQTT consumer
    config.go              viper (.env + env)
    gin.go, postgres.go, rabbitmq.go (topologi), mqtt.go, validator.go, logger.go
  entity/                tabel DB + aturan domain (Geofence.Contains, Haversine)
  model/                 request/response, payload MQTT, event RabbitMQ, error
    converter/             entity ⇄ model
  repository/            interface + implementasi PostgreSQL (pgx)
  usecase/               business logic
    vehicle_location_usecase.go   ingest, latest, history
    geofence_usecase.go           create/list geofence
    geofence_event_usecase.go     proses event (worker)
    geofence_detector.go          state machine entry/exit
  delivery/              input dari luar → usecase
    http/                  controller Gin, route/, middleware/
    messaging/             consumer MQTT (lokasi) & RabbitMQ (event geofence)
  gateway/               output ke sistem luar
    messaging/             producer RabbitMQ (geofence event)
db/migrations/           migrasi golang-migrate (NNNNNN_nama.up.sql / .down.sql)
postman/                 Postman collection
```

### Layer & library per folder

| Folder | Layer clean architecture | Library yang dipakai | Isi |
|---|---|---|---|
| `internal/entity/` | **Entity** (paling dalam) | Standard library saja (`math`) | Bentuk tabel database + rumus Haversine |
| `internal/model/` | **Entity** (kontrak data) | Standard library saja | Format JSON request/response, pesan MQTT & RabbitMQ, aturan validasi (tag) |
| `internal/usecase/` | **Use Case** (business logic) | `go-playground/validator` (validasi input), `google/uuid` (ID UUIDv7) | Simpan lokasi, deteksi geofence, proses event |
| `internal/delivery/http/` | **Interface Adapter** (pintu masuk) | `gin-gonic/gin` | REST API: route, controller, middleware log |
| `internal/delivery/messaging/` | **Interface Adapter** (pintu masuk) | `eclipse/paho.mqtt.golang`, `rabbitmq/amqp091-go` | Consumer MQTT (lokasi) & consumer RabbitMQ (worker) |
| `internal/repository/` | **Interface Adapter** (pintu keluar) | `jackc/pgx/v5` | Query SQL ke PostgreSQL |
| `internal/gateway/messaging/` | **Interface Adapter** (pintu keluar) | `rabbitmq/amqp091-go` | Producer event ke RabbitMQ |
| `internal/config/` | **Frameworks & Drivers** + perakit | `spf13/viper`, `gin`, `pgx/pgxpool`, `paho`, `amqp091-go`, `validator`, `log/slog` | Baca `.env`, buka koneksi, `Bootstrap` merakit semua layer |
| `cmd/api`, `cmd/worker`, `cmd/simulator` | **Main** (titik mulai) | Lewat `config/` (simulator langsung `paho`) | 3 program, dibangun dari satu Dockerfile |
| `db/migrations/` | Database | `golang-migrate` (container `migrate`) | File SQL `up`/`down` |