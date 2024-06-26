﻿# cat-social
# Kucing Matcher

Selamat datang di Kucing Matcher, sebuah aplikasi penjodohan kucing yang memungkinkan Anda untuk menemukan kucing yang cocok dengan preferensi Anda!

## Fitur Utama

- **Pencarian Kucing:** Temukan kucing berdasarkan jenis, usia, jenis kelamin, dan preferensi lainnya.
- **Proses Penjodohan:** Masukkan preferensi Anda, dan biarkan aplikasi mencocokkan Anda dengan kucing yang sesuai.
- **Profil Kucing:** Lihat detail lengkap tentang kucing, termasuk foto, deskripsi, dan informasi lainnya.
- **Interaksi Mudah:** Navigasi yang intuitif dan antarmuka pengguna yang ramah.
- **Update Real-time:** Dapatkan pembaruan langsung saat kucing yang cocok ditemukan.

## Demo

Anda dapat melihat demo aplikasi Kucing Matcher [di sini](link-demo).

## Cara Menggunakan

1. **Pencarian Kucing:** Buka aplikasi dan mulai pencarian kucing berdasarkan preferensi Anda.
2. **Proses Penjodohan:** Masukkan preferensi Anda dalam proses penjodohan untuk menemukan kucing yang sesuai.
3. **Lihat Hasil:** Lihat hasil penjodohan dan temukan kucing yang cocok dengan Anda.
4. **Jodohkan Kucing:** Setelah menemukan kucing yang cocok, Anda dapat mulai proses adopsi atau penjodohan.

## Kontribusi

Kami sangat menghargai kontribusi dari para pengembang. Silakan ikuti langkah-langkah di bawah ini untuk berkontribusi:

1. Lakukan *fork* repositori ini.
2. Buat *branch* baru (`git checkout -b fitur-anda`).
3. Lakukan perubahan yang diperlukan.
4. *Commit* perubahan Anda (`git commit -am 'Menambahkan fitur baru'`).
5. *Push* ke *branch* yang dibuat sebelumnya (`git push origin fitur-anda`).
6. Buat permintaan tarik (*pull request*).

## Catatan Penting
1. Build APP `GOARCH=amd64 GOOS=linux go build -o main_bagasseptyonoo`
2. Run K6 `$env:BASE_URL = "http://localhost:8080"` >  `make run`
3. Mode Debug `$env:DEBUG_ALL = "true"` > `make run`
4. SCP `scp -i w1key main_namauserlead ubuntu@128.x.x.x`
   Tutor :
   1. Download ap-southeast-1-bundle.pem dan w1key ke folder project golang
   2. Compile golangnya sesuai format main_namauserlead

# Migration
1. Execute `migrate -database "postgres://postgres:nama@localhost:5432/cat?sslmode=disable" -path db/migrations up`
2. Rollback `migrate -database "postgres://postgres:nama@localhost:5432/cat?sslmode=disable" -path db/migrations down`

