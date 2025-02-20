import time
import subprocess
import os
import signal
import atexit

# List to track all Popen processes that we start.
child_procs = []

def kill_processes_by_keyword(keyword):
    """
    Kill any processes that match the given keyword.
    Adjust the keyword to match the command string you expect.
    """
    try:
        subprocess.run(f"pkill -f '{keyword}'", shell=True, check=True)
    except subprocess.CalledProcessError:
        # pkill returns nonzero if no process matched, so ignore that.
        pass

def cleanup():
    # Terminate all child processes we started.
    print("Cleaning up child processes...")
    for proc in child_procs:
        try:
            proc.terminate()  # politely ask to terminate
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()
atexit.register(cleanup)

# Optionally handle SIGINT (Ctrl+C) so cleanup is triggered.
def signal_handler(sig, frame):
    print("Signal received, cleaning up...")
    cleanup()
    exit(0)

signal.signal(signal.SIGINT, signal_handler)

# Kill previous instances before starting new ones.
# Adjust the keywords as needed.
kill_processes_by_keyword("npm run dev")
kill_processes_by_keyword("firebase emulators:start")
kill_processes_by_keyword("go run cmd/main.go")
kill_processes_by_keyword("go run main.go")
# Also kill any emulator running on your ports, if needed:
subprocess.run("lsof -t -i:8080 -i:9000 -i:9099 -i:9199 -i:9090 | xargs kill -9", shell=True)

# Start npm dev server (runs in background)
npm_dev = subprocess.Popen("npm run dev -- --open", cwd="frontend", shell=True)
child_procs.append(npm_dev)

# Restart Docker container for RabbitMQ
subprocess.run("docker stop rabbit", shell=True)
time.sleep(2)
subprocess.run("docker rm rabbit", shell=True)
subprocess.run("docker run -d --hostname my-rabbit --name rabbit -p 5672:5672 -p 15672:15672 rabbitmq:3-management", shell=True)
time.sleep(5)

# Start the rabbit consumer (Go process)
rabbit_consumer = subprocess.Popen("go run main.go", cwd="rabbit-consumer", shell=True)
child_procs.append(rabbit_consumer)

# Start Firebase emulators in one process
firebase_proc = subprocess.Popen("firebase emulators:start", cwd="go", shell=True)
child_procs.append(firebase_proc)

# Set FIRESTORE_EMULATOR_HOST and run your main Go process
env = os.environ.copy()
env["FIRESTORE_EMULATOR_HOST"] = "localhost:8080"

go_main = subprocess.Popen("go run cmd/main.go", cwd="go", shell=True, env=env)
child_procs.append(go_main)

# Optionally, wait indefinitely (or until a signal interrupts)
try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    cleanup()