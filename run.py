import time
import subprocess
import os
import signal
import atexit

child_procs = []

def kill_processes_by_keyword(keyword, cwd=None):
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
subprocess.Popen("gcloud beta emulators pubsub env-init", shell=True, env=env)
pubsub_emulator_proc = subprocess.Popen("gcloud beta emulators pubsub start --project=jtrackerkimpark", shell=True, env=env)
child_procs.append(pubsub_emulator_proc)
time.sleep(3)

# 3. Start main API server in go/ 
#    with FIRESTORE_EMULATOR_HOST and PUBSUB_EMULATOR_HOST set.
env_go = os.environ.copy()
env_go["FIRESTORE_EMULATOR_HOST"] = "localhost:8080"
env_go["PUBSUB_EMULATOR_HOST"] = "localhost:8085"
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