# Multi-Tenant FileBrowser Orchestrator

Script **Bash** sederhana untuk men-deploy layanan **FileBrowser** secara otomatis berbasis Docker. Script ini mendukung multi-tenancy dimana setiap user memiliki container dan port terpisah, serta penyimpanan data yang terisolasi.

## Prasyarat (Prerequisites)

Pastikan sistem operasi Anda (USAHAKAN LINUX JANGAN WSL) sudah terinstall:
1.  **Docker Engine** (Pastikan service Docker sudah berjalan)
2.  User Linux Anda sudah dimasukkan ke grup docker (`sudo usermod -aG docker $USER`).
3.  **Bash** (umumnya sudah ada di Linux)

## Instalasi

1.  **Clone Repository ini**.

2.  **(Opsional) Tambahkan eksekusi pada script**
    ```bash
    chmod +x deploy_tenant.sh remove_tenant.sh
    ```

## Cara Menjalankan

### 1. Deploy Tenant Baru
Gunakan script bash berikut untuk membuat tenant baru:

```bash
./deploy_tenant.sh [--start-port PORT] tenant_a tenant_b ...
```

Contoh:
```bash
./deploy_tenant.sh budi siti tono
```
Atau dengan port custom:
```bash
./deploy_tenant.sh --start-port 9000 budi siti tono
```

### 2. Akses Aplikasi
Buka browser dan akses alamat berikut sesuai urutan deploy:

Tenant 1: http://localhost:8000

Tenant 2: http://localhost:8001

..dst

Password admin akan muncul di log terminal setelah deploy.

### 3. Menghapus Tenant
Untuk menghapus container, volume, dan folder tenant:

```bash
./remove_tenant.sh tenant_a tenant_b ...
```

Contoh:
```bash
./remove_tenant.sh budi siti tono
```

## Catatan Penting
- Folder `tenants/` dan seluruh isinya **jangan di-commit ke git** (sudah ada di `.gitignore`).
- Jika ingin password admin baru, hapus volume docker tenant terkait sebelum deploy ulang (otomatis dilakukan oleh `remove_tenant.sh`).
- Script ini tidak lagi membutuhkan Python sama sekali.
