#!/usr/bin/python

# Script reads config json file then checks status and lights status led to appropriate color.

# from pyautogui import hotkey
import RPi.GPIO as GPIO
from time import sleep
import json

pathToConfig = '/home/pi/living-frame-go/config.json'

GPIO.setmode(GPIO.BOARD)

GPIO.setwarnings(False)

GPIO.setup(11, GPIO.OUT)
GPIO.setup(13, GPIO.OUT)
GPIO.setup(15, GPIO.OUT)

red = GPIO.PWM(11, 100)
green = GPIO.PWM(13, 100)
blue = GPIO.PWM(15, 100)

red.start(0)
green.start(0)
blue.start(0)


shown = False

try:
  while True:
      print("checking")
      f = open(pathToConfig)
      data = json.load(f)
      status = data["status"]

      # No status (off)
      if status == 0:
            red.ChangeDutyCycle(0)
            green.ChangeDutyCycle(0)
            blue.ChangeDutyCycle(0)
      # Pocessing/waiting (blue)
      elif status == 1:
            shown = False
            red.ChangeDutyCycle(0)
            green.ChangeDutyCycle(0)
            blue.ChangeDutyCycle(100)
      # All good status (green)
      elif status == 2:
            # Keeps green on for 12 seconds then turns it off
            if shown >= 12:
                  red.ChangeDutyCycle(0)
                  green.ChangeDutyCycle(10)
                  blue.ChangeDutyCycle(0)
            elif shown:
                  shown = shown + 3
                  red.ChangeDutyCycle(0)
                  green.ChangeDutyCycle(100)
                  blue.ChangeDutyCycle(0)
            else:
                  red.ChangeDutyCycle(0)
                  green.ChangeDutyCycle(100)
                  blue.ChangeDutyCycle(0)
                  shown = 3
      # Error status (red)
      else:
            shown = False
            red.ChangeDutyCycle(100)
            green.ChangeDutyCycle(0)
            blue.ChangeDutyCycle(0)
      
      # Loop every three seconds
      sleep(3)

except KeyboardInterrupt:
      GPIO.cleanup()
      pass