import time
import subprocess
import os
import signal
import atexit
import platform

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
kill_processes_by_keyword("firebase emulators:start", cwd="go")
kill_processes_by_keyword("go run cmd/main.go", cwd="go")
kill_processes_by_keyword("go run main.go", cwd="algolia-consumer")
kill_processes_by_keyword("go run main.go", cwd="bigquery-consumer")
kill_processes_by_keyword("gcloud beta emulators pubsub start")

if is_windows:
    ports = ["8080", "8085", "9000", "9099", "9199"]
    for port in ports:
        # Fixed command with proper % escaping and error handling
        try:
            # Find PID using port
            result = subprocess.run(f"netstat -ano | findstr :{port}", shell=True, capture_output=True, text=True)
            if result.stdout:
                # Parse each line of output to get PIDs
                for line in result.stdout.splitlines():
                    parts = line.strip().split()
                    if len(parts) >= 5:  # Make sure we have enough parts
                        pid = parts[-1]  # Last element should be the PID
                        if pid.isdigit():
                            # Kill the process
                            subprocess.run(f"taskkill /F /PID {pid}", shell=True, check=False)
                            print(f"Killed process with PID {pid} on port {port}")
        except Exception as e:
            print(f"Error killing process on port {port}: {e}")
else:
    subprocess.run("lsof -t -i:8080 -i:8085 -i:9000 -i:9099 -i:9199 | xargs kill -9", shell=True)
    
# 1. Start Firebase emulator in go/ directory.
firebase_proc = subprocess.Popen("firebase emulators:start", cwd="go", shell=True)
child_procs.append(firebase_proc)
time.sleep(3)

# 2. Start Pub/Sub emulator externally.
#    NOTE: You must have exported:
#          PUBSUB_EMULATOR_HOST=localhost:8085
#          PUBSUB_PROJECT_ID=jtrackerkimpark
#    and run: gcloud beta emulators pubsub env-init before starting.
env = os.environ.copy()
env["PUBSUB_EMULATOR_HOST"] = "localhost:8085"
env["PUBSUB_PROJECT_ID"] = "jtrackerkimpark"
if is_windows:
    # for some reason if windows we don't run env-init. but this obviously means you should have run it before manually
    pass
else:
    subprocess.Popen("gcloud beta emulators pubsub env-init", shell=True, env=env)
pubsub_emulator_proc = subprocess.Popen("gcloud beta emulators pubsub start --project=jtrackerkimpark", shell=True, env=env)
child_procs.append(pubsub_emulator_proc)
time.sleep(3)

# 3. Start main API server in go/ 
#    with FIRESTORE_EMULATOR_HOST and PUBSUB_EMULATOR_HOST set.
env_go = os.environ.copy()
env_go["FIRESTORE_EMULATOR_HOST"] = "localhost:8080"
env_go["PUBSUB_EMULATOR_HOST"] = "localhost:8085"
env_go["PUBSUB_ORDERING_KEY"] = "insecure_local_key"
env_go["FRONTEND_URL"] = "http://localhost:5173"
go_main = subprocess.Popen("go run cmd/main.go", cwd="go", shell=True, env=env_go)
child_procs.append(go_main)
time.sleep(3)

# 4. Start algolia-consumer
algolia_consumer = subprocess.Popen("go run main.go", cwd="algolia-consumer", shell=True, env=env_go)
child_procs.append(algolia_consumer)

# 5. Start bigquery-consumer
bigquery_consumer = subprocess.Popen("go run main.go", cwd="bigquery-consumer", shell=True, env=env_go)
child_procs.append(bigquery_consumer)

# 6. Start frontend dev server
npm_dev = subprocess.Popen("npm run dev -- --open", cwd="frontend", shell=True)
child_procs.append(npm_dev)

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    cleanup()