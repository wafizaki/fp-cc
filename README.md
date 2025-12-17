# Multi-Tenant FileBrowser Orchestrator

Script Python sederhana untuk men-deploy layanan **FileBrowser** secara otomatis berbasis Docker. Script ini mendukung multi-tenancy dimana setiap user memiliki container dan port terpisah, serta penyimpanan data yang terisolasi.

## Prasyarat (Prerequisites)

Pastikan sistem operasi Anda (USAHAKAN LINUX JANGAN WSL) sudah terinstall:
1.  **Python 3.8+**
2.  **Docker Engine** (Pastikan service Docker sudah berjalan)
3.  User Linux Anda sudah dimasukkan ke grup docker (`sudo usermod -aG docker $USER`).

## Instalasi

1.  **Clone Repository ini**.

2.  **Buat Virtual Environment (Venv)**
    Agar library python tidak mengotori sistem utama.
    ```bash
    python3 -m venv venv
    ```

3.  **Aktifkan Venv**
    Lakukan ini setiap kali ingin menjalankan script.
    ```bash
    source venv/bin/activate
    ```
    *(Tanda berhasil: muncul tulisan `(venv)` di terminal)*
    
    Kemudian lakukan ini untuk mematikan venv
    ```bash
    deactivate
    ```

4.  **Install Dependencies**
    ```bash
    pip install -r requirements.txt
    ```
    *(Atau jika manual: `pip install docker`)*

## Cara Menjalankan

### 1. Deploy Tenant Baru
Gunakan argumen `--names` diikuti daftar nama tenant yang ingin dibuat.

```bash
python deploy_tenant.py --names budi siti tono
```

### 2. Akses Aplikasi
Buka browser dan akses alamat berikut sesuai urutan deploy:

Tenant 1: http://localhost:8000

Tenant 2: http://localhost:8001

..dst
