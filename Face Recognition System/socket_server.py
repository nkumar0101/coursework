import socket
import select
import cv2
import numpy as np
from ultralytics import YOLO

# Load YOLOv5 model
model = YOLO('yolov5s.pt')  # Using YOLOv5 small model

# Socket server setup
SERVER_IP = '0.0.0.0'
SERVER_PORT = 12345

server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
server_socket.bind((SERVER_IP, SERVER_PORT))
server_socket.listen(5)
server_socket.setblocking(False)
print(f"Server listening on {SERVER_IP}:{SERVER_PORT}")

sockets_list = [server_socket]
clients = {}

def receive_image(client_socket):
    try:
        # Receive image size
        data = client_socket.recv(4)
        if not data:
            print("No data received for image size.")
            return None, None
        img_size = int.from_bytes(data, byteorder='big')
        print(f"Image size received: {img_size}")
        if img_size <= 0 or img_size > 1000000:  # Added condition to filter invalid sizes
            print(f"Invalid image size received: {img_size}")
            return None, None

        # Receive image data
        data = b""
        while len(data) < img_size:
            ready_to_read, _, _ = select.select([client_socket], [], [], 0.1)
            if ready_to_read:
                packet = client_socket.recv(min(img_size - len(data), 4096))
                if not packet:
                    print(f"No packet received. Current data length: {len(data)}")
                    return None, None
                data += packet
            else:
                print(f"Waiting for data. Current data length: {len(data)}")

        if len(data) != img_size:
            print(f"Data size mismatch: expected {img_size}, got {len(data)}")
            return None, None

        return img_size, data

    except Exception as e:
        print(f"Error receiving image: {e}")
        return None, None

def process_frame(image_data):
    try:
        # Convert image data to a numpy array
        nparr = np.frombuffer(image_data, np.uint8)
        img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
        
        if img is None:
            print("Failed to decode image")
            return None

        # Perform YOLO detection
        results = model(img)

        # Draw bounding boxes for detected objects
        for result in results[0].boxes:
            x1, y1, x2, y2 = map(int, result.xyxy[0])
            conf = result.conf[0]
            cls = result.cls[0]
            label = model.names[int(cls)]
            if label == 'person':  # Only draw bounding boxes for 'person'
                cv2.rectangle(img, (x1, y1), (x2, y2), (255, 0, 0), 2)
                cv2.putText(img, f"{label} {conf:.2f}", (x1, y1 - 10), cv2.FONT_HERSHEY_SIMPLEX, 0.5, (255, 0, 0), 2)
                print(f"Detected {label} with confidence {conf:.2f} at [{x1}, {y1}, {x2}, {y2}]")

        return img

    except Exception as e:
        print(f"Error processing frame: {e}")
        return None

while True:
    read_sockets, _, exception_sockets = select.select(sockets_list, [], sockets_list)

    for notified_socket in read_sockets:
        if notified_socket == server_socket:
            client_socket, client_address = server_socket.accept()
            client_socket.setblocking(False)
            sockets_list.append(client_socket)
            clients[client_socket] = (client_address, f"Camera {client_address}", b"")
            print(f"Accepted new connection from {client_address}")

        else:
            image_size, image_data = receive_image(notified_socket)
            if image_size is None:
                print(f"Closed connection from {clients[notified_socket][0]}")
                sockets_list.remove(notified_socket)
                del clients[notified_socket]
            else:
                clients[notified_socket] = (clients[notified_socket][0], clients[notified_socket][1], image_data)

    for client_socket, (addr, window_name, data) in clients.items():
        if data:
            img = process_frame(data)
            if img is not None:
                cv2.imshow(window_name, img)
                cv2.waitKey(1)

    for notified_socket in exception_sockets:
        sockets_list.remove(notified_socket)
        del clients[notified_socket]

cv2.destroyAllWindows()
