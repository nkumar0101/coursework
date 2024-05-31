#define CAMERA_MODEL_AI_THINKER // Define the camera model

#include "esp_camera.h"
#include <WiFi.h>
#include <WiFiClient.h>

// Replace with your network credentials
const char* ssid = "challenger";
const char* password = "784765776800510559988564";

// Server IP and port
const char* serverIP = "192.168.101.21"; // Change to your server's IP
const int serverPort = 12345;

// Camera Pin Definition for ESP32-CAM AI-Thinker Model
#define PWDN_GPIO_NUM     32
#define RESET_GPIO_NUM    -1
#define XCLK_GPIO_NUM      0
#define SIOD_GPIO_NUM     26
#define SIOC_GPIO_NUM     27

#define Y9_GPIO_NUM       35
#define Y8_GPIO_NUM       34
#define Y7_GPIO_NUM       39
#define Y6_GPIO_NUM       36
#define Y5_GPIO_NUM       21
#define Y4_GPIO_NUM       19
#define Y3_GPIO_NUM       18
#define Y2_GPIO_NUM        5
#define VSYNC_GPIO_NUM    25
#define HREF_GPIO_NUM     23
#define PCLK_GPIO_NUM     22

WiFiClient client;

void setup() {
  Serial.begin(115200);

  // Connect to Wi-Fi
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(100);
    Serial.println("Connecting to WiFi...");
  }
  Serial.println("Connected to WiFi");

  // Configure camera settings
  camera_config_t config;
  config.ledc_channel = LEDC_CHANNEL_0;
  config.ledc_timer = LEDC_TIMER_0;
  config.pin_d0 = Y2_GPIO_NUM;
  config.pin_d1 = Y3_GPIO_NUM;
  config.pin_d2 = Y4_GPIO_NUM;
  config.pin_d3 = Y5_GPIO_NUM;
  config.pin_d4 = Y6_GPIO_NUM;
  config.pin_d5 = Y7_GPIO_NUM;
  config.pin_d6 = Y8_GPIO_NUM;
  config.pin_d7 = Y9_GPIO_NUM;
  config.pin_xclk = XCLK_GPIO_NUM;
  config.pin_pclk = PCLK_GPIO_NUM;
  config.pin_vsync = VSYNC_GPIO_NUM;
  config.pin_href = HREF_GPIO_NUM;
  config.pin_sccb_sda = SIOD_GPIO_NUM;
  config.pin_sccb_scl = SIOC_GPIO_NUM;
  config.pin_pwdn = PWDN_GPIO_NUM;
  config.pin_reset = RESET_GPIO_NUM;
  config.xclk_freq_hz = 20000000;
  config.pixel_format = PIXFORMAT_JPEG;

  // Set frame size and quality
  config.frame_size = FRAMESIZE_HVGA;  // Reduce the frame size for faster transmission
  config.jpeg_quality = 15;  // Increase compression for smaller size
  config.fb_count = 1;

  // Initialize the camera
  esp_err_t err = esp_camera_init(&config);
  if (err != ESP_OK) {
    Serial.printf("Camera init failed with error 0x%x\n", err);
    return;
  }
}

void loop() {
  if (WiFi.status() == WL_CONNECTED) {
    // Capture a picture
    camera_fb_t * fb = esp_camera_fb_get();
    if (!fb) {
      Serial.println("Camera capture failed");
      delay(100); // Wait before retrying
      return;
    }

    // Print image size
    Serial.printf("Image size: %d bytes\n", fb->len);

    // Connect to server
    if (!client.connected()) {
      if (!client.connect(serverIP, serverPort)) {
        Serial.println("Connection to server failed");
        esp_camera_fb_return(fb);
        delay(100); // Wait before retrying
        return;
      }
    }

    // Send image size as 4 bytes
    uint32_t img_size = fb->len;
    uint8_t size_bytes[4];
    size_bytes[0] = (img_size >> 24) & 0xFF;
    size_bytes[1] = (img_size >> 16) & 0xFF;
    size_bytes[2] = (img_size >> 8) & 0xFF;
    size_bytes[3] = img_size & 0xFF;
    client.write(size_bytes, sizeof(size_bytes));

    // Send image data
    client.write(fb->buf, fb->len);

    esp_camera_fb_return(fb);
    Serial.println("Frame sent to server");

    // No delay for higher frame rate
  } else {
    Serial.println("WiFi disconnected");
    WiFi.begin(ssid, password); // Attempt to reconnect
    delay(100);
  }
}
