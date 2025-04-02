import time
import subprocess
import os
import signal
import atexit
import platform
import re

child_procs = []
is_windows = platform.system() == "Windows"

def kill_processes_by_keyword(keyword, cwd=None):
    if is_windows:
        try:
            subprocess.run(f'taskkill /F /FI "WINDOWTITLE eq *{keyword}*" /T', shell=True, check=False, cwd=cwd)
            subprocess.run(f'taskkill /F /FI "IMAGENAME eq {keyword}*" /T', shell=True, check=False, cwd=cwd)
        except subprocess.CalledProcessError:
            pass
    else:
        try:
            subprocess.run(f"pkill -f '{keyword}'", shell=True, check=True, cwd=cwd)
        except subprocess.CalledProcessError:
            pass

def cleanup():
    print("Cleaning up child processes...")
    
    # Save Supabase schema changes
    try:
        print("Checking for Supabase schema changes...")
        schema_diff = subprocess.run("supabase db diff -f auto_migration", 
                                    shell=True, capture_output=True, text=True)
        if "No schema changes detected" not in schema_diff.stdout:
            print("Schema changes saved to migration file")
        else:
            print("No schema changes detected")
            # Save the schema changes
            subprocess.run("supabase db push", shell=True)
            print("Schema changes pushed to database")
    except Exception as e:
        print(f"Error saving schema changes: {e}")
    
    # Now stop Supabase
    try:
        subprocess.run("supabase stop", shell=True)
        print("Supabase stopped")
    except Exception as e:
        print(f"Error stopping Supabase: {e}")
    
    # Terminate other processes
    for proc in child_procs:
        try:
            proc.terminate()
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()

atexit.register(cleanup)

def signal_handler(sig, frame):
    print("Signal received, cleaning up...")
    cleanup()
    exit(0)

signal.signal(signal.SIGINT, signal_handler)

# Kill any lingering processes
kill_processes_by_keyword("npm run dev", cwd="frontend")
kill_processes_by_keyword("go run cmd/main.go", cwd="go")

# Free up ports
if is_windows:
    ports = ["8080", "3000"]
    for port in ports:
        try:
            result = subprocess.run(f"netstat -ano | findstr :{port}", shell=True, capture_output=True, text=True)
            if result.stdout:
                for line in result.stdout.splitlines():
                    parts = line.strip().split()
                    if len(parts) >= 5:
                        pid = parts[-1]
                        if pid.isdigit():
                            subprocess.run(f"taskkill /F /PID {pid}", shell=True, check=False)
                            print(f"Killed process with PID {pid} on port {port}")
        except Exception as e:
            print(f"Error killing process on port {port}: {e}")
else:
    subprocess.run("lsof -t -i:8080 -i:3000 | xargs kill -9 2>/dev/null || true", shell=True)
    
# 1. Start Supabase
print("Starting Supabase local development...")
supabase_env = os.environ.copy()
env_go = os.environ.copy()  
env_go["FRONTEND_URL"] = "http://localhost:5173"
try:
    # Check if Supabase is already running
    check_proc = subprocess.run("supabase status", shell=True, capture_output=True, text=True)
    if "not running" in check_proc.stdout:
        # Start Supabase if not running
        supabase_proc = subprocess.Popen("supabase start", shell=True, env=supabase_env)
        child_procs.append(supabase_proc)
        print("Waiting for Supabase to start...")
        time.sleep(10)  # Give it time to start
    else:
        print("Supabase already running")
    
    # Get Supabase connection info
    supabase_status = subprocess.run("supabase status", shell=True, capture_output=True, text=True).stdout
    
    # Try getting the connection string with proper error handling
    try:
        # First try the connection-string command
        conn_result = subprocess.run("supabase db connection-string", 
                                    shell=True, capture_output=True, text=True)
        
        # Check if we got an error or help text instead of a connection string
        if "Usage:" in conn_result.stdout or "Available Commands:" in conn_result.stdout:
            print("Could not retrieve connection string with command. Using default local development settings.")
            # Default local Supabase connection string
            env_go["DATABASE_URL"] = "postgresql://postgres:postgres@localhost:54322/postgres"
        else:
            env_go["DATABASE_URL"] = conn_result.stdout.strip()
    except Exception as conn_err:
        print(f"Error getting connection string: {conn_err}")
        
        # Fallback: Try to parse from status output
        try:
            for line in supabase_status.splitlines():
                if "API URL:" in line and "http://localhost:54321" in line:
                    # If we can confirm this is a standard local setup, use standard connection
                    env_go["DATABASE_URL"] = "postgresql://postgres:postgres@localhost:54322/postgres"
                    break
                if "DB URL:" in line:
                    url_part = line.split("DB URL:")[1].strip()
                    if url_part.startswith("postgresql://"):
                        env_go["DATABASE_URL"] = url_part
                        break
        except Exception:
            # Last resort fallback
            env_go["DATABASE_URL"] = "postgresql://postgres:postgres@localhost:54322/postgres"
    
    print(f"Database URL: {env_go['DATABASE_URL']}")
    print("Supabase environment variables set")
        
except Exception as e:
    print(f"Error starting Supabase: {e}")
    # Set default connection in case of error
    env_go["DATABASE_URL"] = "postgresql://postgres:postgres@localhost:54322/postgres"

# 2. Start Go API server
print("Starting Go API server...")
try:
    # pass in the database url 
    go_proc = subprocess.Popen("go run cmd/main.go", shell=True, env=env_go, cwd="go")
    child_procs.append(go_proc)
    print("Go API server started")
except Exception as e:
    print(f"Error starting Go API server: {e}")

# 3. Start Frontend
print("Starting frontend...")
frontend_env = os.environ.copy()
try:
    frontend_proc = subprocess.Popen("npm run dev", shell=True, env=frontend_env, cwd="frontend")
    child_procs.append(frontend_proc)
    print("Frontend started")
except Exception as e:
    print(f"Error starting frontend: {e}")

print("All services started. Press Ctrl+C to stop.")

# Keep script running
try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    print("Interrupted by user, shutting down...")
    cleanup()

    