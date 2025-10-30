import requests
import subprocess
import time
import sys
import tkinter as tk
import threading
import os
import uuid
from flask import Flask, request, jsonify
from flask_cors import CORS

# Device ID logic
DEVICE_ID_FILE = ".device_id"

# Reads device ID from .device_id file. If not present, instruct user to copy from browser localStorage.
def get_device_id():
    
    if os.path.exists(DEVICE_ID_FILE):
        with open(DEVICE_ID_FILE, "r") as f:
            return f.read().strip()
    print("[LAUNCHER-INFO] No .device_id file found.")
    print("[LAUNCHER-INFO] Please copy your deviceId from browser localStorage and paste it below.")
    print("[LAUNCHER-INFO] In your browser, run: localStorage.getItem('deviceId')")
    device_id = input("Enter deviceId: ").strip()
    with open(DEVICE_ID_FILE, "w") as f:
        f.write(device_id)
    return device_id

# Flask server to receive device ID from browser
def start_device_id_server():
    app = Flask(__name__)
    CORS(app)

    @app.route('/set-device-id', methods=['POST'])
    def set_device_id():
        data = request.get_json()
        device_id = data.get('deviceId')
        if device_id:
            with open(DEVICE_ID_FILE, "w") as f:
                f.write(device_id)
            print(f"[SERVER] Device ID received and saved: {device_id}")
            return jsonify({"status": "success"}), 200
        else:
            return jsonify({"status": "error", "message": "No deviceId provided"}), 400

    print("[SERVER] Starting device ID server on http://localhost:5000 ...")
    app.run(host="127.0.0.1", port=5000)

# Start server in background thread
def run_server_in_background():
    server_thread = threading.Thread(target=start_device_id_server)
    server_thread.daemon = True
    server_thread.start()

run_server_in_background()

def get_backend_url():
    device_id = get_device_id()
    return f"http://10.0.5.163:8080/api/v1/game-launch-status?userId={device_id}"
GAME_PATH = "/home/vm/Downloads/snake-pygame-master/snake_Game.py"

class TimerWindow(tk.Tk):
    def __init__(self, total_seconds):
        super().__init__()
        self.title("Game Timer")
        self.geometry("220x80+20+20")  # Small window, top-left corner
        self.resizable(False, False)
        self.configure(bg="#222831")
        self.total_seconds = total_seconds
        self.remaining = total_seconds

        self.label = tk.Label(self, text="", font=("Segoe UI", 24, "bold"), fg="#FDD226", bg="#222831")
        self.label.pack(expand=True)

        self.protocol("WM_DELETE_WINDOW", self.on_close)
        self.update_timer()

    def update_timer(self):
        mins, secs = divmod(self.remaining, 60)
        self.label.config(text=f"{mins:02d}:{secs:02d} left")
        if self.remaining > 0:
            self.remaining -= 1
            self.after(1000, self.update_timer)
        else:
            self.label.config(text="Time's up!")
            self.after(2000, self.destroy)

    def on_close(self):
        pass  # Prevent closing the timer window manually

def wait_for_device_id_change():
    initial_id = get_device_id()
    print(f"[LAUNCHER-DEBUG] Initial device ID: {initial_id}")
    # If device ID is valid (not empty or placeholder), start polling immediately
    if initial_id and not initial_id.startswith('dev-'):
        print("[LAUNCHER-DEBUG] Device ID already set, starting backend polling...")
        return initial_id
    print("[LAUNCHER-DEBUG] Waiting for device ID to be updated by frontend...")
    while True:
        current_id = get_device_id()
        if current_id != initial_id and current_id and not current_id.startswith('dev-'):
            print(f"[LAUNCHER-DEBUG] Device ID updated: {current_id}")
            return current_id
        time.sleep(1)

def poll_for_launch():
    # Wait for device ID to be updated before polling backend
    device_id = wait_for_device_id_change()
    while True:
        backend_url = f"http://10.0.5.163:8080/api/v1/game-launch-status?userId={device_id}"
        print(f"[LAUNCHER-DEBUG] Polling backend for launch status: {backend_url}")
        try:
            resp = requests.get(backend_url)
            print(f"[LAUNCHER-DEBUG] Backend response: status={resp.status_code}, body={resp.text}")
            if resp.status_code == 200:
                data = resp.json()
                print(f"[LAUNCHER-DEBUG] Parsed JSON: {data}")
                if data.get("launch"):
                    print(f"[LAUNCHER-DEBUG] Launch flag received! Time limit: {data.get('time_limit', 10)} minutes")
                    return int(data.get("time_limit", 10))
                else:
                    print("[LAUNCHER-DEBUG] Launch flag not set yet. Waiting...")
        except Exception as e:
            print("[LAUNCHER-ERROR] Error polling backend:", e)
        time.sleep(5)

def launch_game_with_timer(game_path, time_limit):
    print(f"[LAUNCHER-DEBUG] Launching game for {time_limit} minutes...")

    # Start timer UI in a separate thread
    timer_thread = threading.Thread(target=run_timer, args=(time_limit,))
    timer_thread.daemon = True
    timer_thread.start()

    # Start the game process
    print(f"[LAUNCHER-DEBUG] Starting game process: {game_path}")
    proc = subprocess.Popen([sys.executable, game_path])
    start_time = time.time()
    while True:
        elapsed = int(time.time() - start_time)
        remaining = time_limit * 60 - elapsed
        print(f"[LAUNCHER-DEBUG] Game running. Elapsed: {elapsed}s, Remaining: {remaining}s")
        if remaining <= 0:
            print("[LAUNCHER-DEBUG] Time's up! Closing the game.")
            proc.terminate()
            break
        time.sleep(1)
    timer_thread.join()

def run_timer(total_minutes):
    total_seconds = total_minutes * 60
    print(f"[LAUNCHER-DEBUG] Starting timer window for {total_minutes} minutes ({total_seconds} seconds)")
    timer = TimerWindow(total_seconds)
    timer.mainloop()

def monitor_and_launch():
    last_userid = None
    while True:
        current_userid = get_device_id()
        if current_userid != last_userid and current_userid and not current_userid.startswith('dev-'):
            print(f"[LAUNCHER-DEBUG] Detected new device ID: {current_userid}")
            print("[LAUNCHER-DEBUG] Waiting for payment success signal from backend...")
            allowed_time = None
            try:
                allowed_time = poll_for_launch_with_id(current_userid)
            except Exception as e:
                print(f"[LAUNCHER-ERROR] Error during polling for launch: {e}")
            if allowed_time:
                print(f"[LAUNCHER-DEBUG] Launching game with allowed time: {allowed_time} minutes")
                launch_game_with_timer(GAME_PATH, allowed_time)
            last_userid = current_userid
        time.sleep(1)

# Poll for launch for a specific userId
def poll_for_launch_with_id(device_id):
    while True:
        backend_url = f"http://10.0.5.163:8080/api/v1/game-launch-status?userId={device_id}"
        print(f"[LAUNCHER-DEBUG] Polling backend for launch status: {backend_url}")
        try:
            resp = requests.get(backend_url)
            print(f"[LAUNCHER-DEBUG] Backend response: status={resp.status_code}, body={resp.text}")
            if resp.status_code == 200:
                data = resp.json()
                print(f"[LAUNCHER-DEBUG] Parsed JSON: {data}")
                if data.get("launch"):
                    print(f"[LAUNCHER-DEBUG] Launch flag received! Time limit: {data.get('time_limit', 10)} minutes")
                    return int(data.get("time_limit", 10))
                else:
                    print("[LAUNCHER-DEBUG] Launch flag not set yet. Waiting...")
        except Exception as e:
            print("[LAUNCHER-ERROR] Error polling backend:", e)
        # Check for userId change before next poll
        new_id = get_device_id()
        if new_id != device_id:
            print(f"[LAUNCHER-DEBUG] Device ID changed from {device_id} to {new_id}, restarting polling...")
            return None
        time.sleep(5)

if __name__ == "__main__":
    monitor_and_launch()
