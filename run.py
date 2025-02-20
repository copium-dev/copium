import time
import subprocess
import os

npm_dev = subprocess.Popen("npm run dev -- --open", cwd="frontend", shell=True)

subprocess.run("docker stop rabbit", shell=True)
time.sleep(2)
subprocess.run("docker rm rabbit", shell=True)
subprocess.run("docker run -d --hostname my-rabbit --name rabbit -p 5672:5672 -p 15672:15672 rabbitmq:3-management", shell=True)
time.sleep(5)

subprocess.Popen("go run main.go", cwd="rabbit-consumer", shell=True)

firebase_proc = subprocess.Popen(
    "firebase emulators:start",
    cwd="go",
    shell=True
)

env = os.environ.copy()
env["FIRESTORE_EMULATOR_HOST"] = "localhost:8080"

subprocess.Popen(
    "go run cmd/main.go",
    cwd="go",
    shell=True,
    env=env
)
