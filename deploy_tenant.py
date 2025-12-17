import os
import docker
import argparse
import sys
import time
import re

# Initialize Docker Client
try:
    client = docker.from_env()
except docker.errors.DockerException:
    print("[ERROR] Docker is not running.")
    sys.exit(1)

def create_directories(tenant_name):
    base_path = os.path.abspath(os.path.join(os.getcwd(), "tenants", tenant_name))
    folders = ["files", "config"] 
    created_paths = {}

    try:
        for folder in folders:
            path = os.path.join(base_path, folder)
            os.makedirs(path, exist_ok=True)
            created_paths[folder] = path
        return created_paths
    except OSError as e:
        print(f"[ERROR] Failed to create directories: {e}")
        return None

def run_container(tenant_name, port, folder_paths):
    container_name = f"files_{tenant_name}"
    db_volume_name = f"{tenant_name}_settings_vol"

    volumes = {
        folder_paths['files']: {'bind': '/srv', 'mode': 'rw'},
        db_volume_name: {'bind': '/database', 'mode': 'rw'},
        folder_paths['config']: {'bind': '/config', 'mode': 'rw'}
    }

    try:
        # Cleanup old container
        try:
            old = client.containers.get(container_name)
            print(f"[INFO] Removing old container {container_name}...")
            old.stop()
            old.remove()
        except docker.errors.NotFound:
            pass

        print(f"[INFO] Starting {tenant_name} on port {port}...")
        
        container = client.containers.run(
            image="filebrowser/filebrowser",
            name=container_name,
            ports={'80/tcp': port},
            volumes=volumes,
            detach=True
        )
        
        # --- LOG SCRAPING LOGIC ---
        print("[INFO] Waiting 5s for logs...")
        time.sleep(5) 
        
        # Ambil logs container
        logs = container.logs().decode('utf-8')
        
        # Cari pola password di log
        # Pattern: "User 'admin' initialized with randomly generated password: [passwordnya]"
        match = re.search(r"password: (\S+)", logs)
        
        if match:
            password = match.group(1)
            print(f"✅ [SUCCESS] Tenant: {tenant_name}")
            print(f"🔗 URL: http://localhost:{port}")
            print(f"👤 User: admin")
            print(f"🔑 Pass: {password}")
        else:
            print(f"⚠️ [WARN] Password not found in logs.")
            print("   (Reason: Volume 'settings_vol' might already exist from previous run.)")
            print("   Action: Run 'docker volume rm' if you want a fresh password.")

        return True

    except Exception as e:
        print(f"[ERROR] Critical failure for {tenant_name}: {e}")
        return False

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--names', nargs='+', required=True)
    parser.add_argument('--start-port', type=int, default=8000)
    args = parser.parse_args()
    
    current_port = args.start_port
    print("=== DEPLOYMENT START (LOG SCRAPE ONLY) ===\n")
    
    for name in args.names:
        print(f"--- Processing Tenant: {name} ---")
        paths = create_directories(name)
        
        if paths:
            run_container(name, current_port, paths)
            current_port += 1 
        
        print("")

if __name__ == "__main__":
    main()
